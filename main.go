package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	_ "net/http/pprof"

	"github.com/PuerkitoBio/goquery"
)

type SeenURLs struct {
	mu   sync.Mutex
	urls map[string]struct{}
}

type Result struct {
	Url    string
	Status int
	Err    error
}

type Report struct {
	Results      []Result
	Working      int
	Redirects    int
	Broken       int
	ServerErrors int
	Errors       int
}

type PageResult struct {
	URL    string
	Status int
	Links  []string
	Err    error
}

var client = &http.Client{
	Timeout: 10 * time.Second,
}

func (r *Report) Add(result Result) {
	r.Results = append(r.Results, result)

	if result.Err != nil {
		r.Errors++
		return
	}

	switch {
	case result.Status >= 200 && result.Status < 300:
		r.Working++

	case result.Status >= 300 && result.Status < 400:
		r.Redirects++

	case result.Status >= 400 && result.Status < 500:
		r.Broken++

	case result.Status >= 500:
		r.ServerErrors++
	}
}
func normalizeURL(href string, baseURL string) (string, bool) {
	if href == "" {
		return "", false
	}

	base, err := url.Parse(baseURL)
	if err != nil {
		return "", false
	}

	ref, err := url.Parse(href)
	if err != nil {
		return "", false
	}

	if ref.Scheme == "mailto" ||
		ref.Scheme == "tel" ||
		ref.Scheme == "javascript" {
		return "", false
	}

	normalized := base.ResolveReference(ref)

	normalized.Fragment = ""

	return normalized.String(), true
}

func extractLinks(r io.Reader) ([]string, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, err

	}
	var links []string
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if exists {
			links = append(links, href)
		}

	})
	return links, nil
}

func isInternal(baseURL string, link string) bool {
	linkURL, err := url.Parse(link)
	if err != nil {
		fmt.Println(err)
		return false
	}
	root, err := url.Parse(baseURL)
	if err != nil {
		fmt.Println(err)
		return false
	}
	return root.Hostname() == linkURL.Hostname()
}

func fetchPage(urls chan string, jobs chan PageResult, wg *sync.WaitGroup) {
	defer wg.Done()
	for url := range urls {
		start := time.Now()
		resp, err := client.Get(url)
		fmt.Println("CRAWL:", url, time.Since(start))
		if err != nil {
			jobs <- PageResult{URL: url, Status: 0, Links: nil, Err: err}
			continue
		}

		contentType := resp.Header.Get("Content-Type")

		if !strings.HasPrefix(contentType, "text/html") {
			jobs <- PageResult{URL: url, Status: resp.StatusCode, Links: nil, Err: err}
			continue
		}
		links, err := extractLinks(resp.Body)
		if err != nil {
			jobs <- PageResult{URL: url, Status: resp.StatusCode, Links: nil, Err: err}
			continue
		}
		jobs <- PageResult{URL: url, Status: resp.StatusCode, Links: links, Err: nil}
		resp.Body.Close()
	}
}

// func checkStatus(jobs <-chan PageResult, results chan<- Result, wg *sync.WaitGroup) {
// 	defer wg.Done()
// 	for job := range jobs {
// 		start := time.Now()

// 		url, status, err := job.URL, job.Status, job.Err

// 		fmt.Println(job, "took", time.Since(start))

// 		results <- Result{
// 			Url:    url,
// 			Status: status,
// 			Err:    err,
// 		}
// 	}

// }
func (s *SeenURLs) AlreadySeenOrMark(link string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exits := s.urls[link]; exits {
		return true
	}
	s.urls[link] = struct{}{}
	return false
}

func validateURL(raw string) (string, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", err
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return "", fmt.Errorf("URL must start with http:// or https://")
	}

	if u.Hostname() == "" {
		return "", fmt.Errorf("invalid URL")
	}

	return u.String(), nil
}

func Scan(ctx context.Context, baseURL string) (Report, error) {
	r := Report{}
	resultDone := make(chan struct{})

	wg := sync.WaitGroup{}
	wg.Add(50)

	seen := SeenURLs{
		urls: make(map[string]struct{}),
	}
	seen.AlreadySeenOrMark(baseURL)

	urls := make(chan string)
	pageResultChan := make(chan PageResult)
	results := make(chan Result)
	for i := 0; i < 50; i++ {
		go fetchPage(urls, pageResultChan, &wg)
	}

	// for i := 0; i < 20; i++ {
	// 	go checkStatus(pageResultChan, results, &wg)
	// }

	go func() {
		defer close(resultDone)
		for res := range results {
			r.Add(res)
		}
	}()

	queue := []string{baseURL}
	outstanding := 0

	for len(queue) > 0 || outstanding > 0 {

		var nextURL string
		var sendChan chan string

		if len(queue) > 0 {
			nextURL = queue[0]
			sendChan = urls
		}

		select {

		case sendChan <- nextURL:
			queue = queue[1:]
			outstanding++

		case pageRes := <-pageResultChan:
			outstanding--

			results <- Result{
				Url:    pageRes.URL,
				Status: pageRes.Status,
				Err:    pageRes.Err,
			}

			if pageRes.Err != nil {
				continue
			}

			// External URLs are checked,
			// but we don't crawl their links.
			if !isInternal(baseURL, pageRes.URL) {
				continue
			}

			for _, href := range pageRes.Links {
				normalized, ok := normalizeURL(href, pageRes.URL)
				if !ok {
					continue
				}

				if seen.AlreadySeenOrMark(normalized) {
					continue
				}

				queue = append(queue, normalized)
			}
		}
	}

	close(urls)
	wg.Wait()

	close(pageResultChan)
	close(results)
	<-resultDone
	return r, nil

}

func scanHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	baseURL := r.FormValue("url")

	baseURL, err := validateURL(baseURL)
	if err != nil {
		http.Error(
			w,
			err.Error(),
			http.StatusBadRequest,
		)
		return
	}
	report, err := Scan(r.Context(), baseURL)
	if err != nil {
		http.Error(
			w,
			err.Error(),
			http.StatusInternalServerError,
		)
		return
	}

	err = reportTemplate.Execute(w, report)
	if err != nil {
		http.Error(
			w,
			"failed to render report",
			http.StatusInternalServerError,
		)
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	err := homeTemplate.Execute(w, nil)
	if err != nil {
		http.Error(
			w,
			"failed to render page",
			http.StatusInternalServerError,
		)
	}
}
func main() {
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/scan", scanHandler)

	fmt.Println("SiteCheck running on http://localhost:8080")

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println(err)
	}
}
