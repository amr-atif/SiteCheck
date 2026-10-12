package main

import "html/template"

const sharedStyles = `
<style>
	:root { --paper:#f1eee5; --ink:#11110f; --signal:#ff4d18; --muted:#6b6a62; --rule:rgba(17,17,15,.24); }
	* { box-sizing:border-box; }
	html,body { margin:0; height:100%; }
	body { background:var(--paper); color:var(--ink); font-family:Arial,Helvetica,sans-serif; overflow:hidden; }
	button,input { font:inherit; }
	.topbar { display:flex; align-items:center; justify-content:space-between; }
	.brand { display:flex; align-items:center; gap:11px; font-family:"Arial Narrow",Arial,sans-serif; font-size:15px; font-weight:900; letter-spacing:.12em; text-transform:uppercase; }
	.brand-mark { position:relative; width:25px; height:25px; border:2px solid var(--ink); }
	.brand-mark::before,.brand-mark::after { content:""; position:absolute; background:var(--ink); }
	.brand-mark::before { width:11px; height:2px; top:6px; left:5px; }
	.brand-mark::after { width:2px; height:11px; top:5px; left:10px; }
	.mono { font-family:"Courier New",monospace; font-size:10px; font-weight:700; letter-spacing:.13em; text-transform:uppercase; }
</style>`

var homeTemplate = template.Must(template.New("home").Parse(`
<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<meta name="theme-color" content="#f1eee5">
	<title>SiteCheck — Link Audit</title>
` + sharedStyles + `
	<style>
		.site-shell { display:grid; grid-template-rows:76px minmax(0,1fr) 46px; height:100dvh; padding:0 24px; }
		.footer { display:flex; align-items:center; justify-content:space-between; color:var(--muted); font-family:"Courier New",monospace; font-size:10px; font-weight:700; letter-spacing:.1em; text-transform:uppercase; }
		.footer strong { color:var(--ink); }
		.system-state { display:flex; align-items:center; gap:9px; }
		.state-dot { width:8px; height:8px; background:var(--signal); border-radius:50%; box-shadow:0 0 0 4px rgba(255,77,24,.13); }
		.stage { display:grid; grid-template-columns:minmax(0,1.18fr) minmax(330px,.82fr); min-height:0; border-top:1px solid var(--ink); border-bottom:1px solid var(--ink); }
		.pitch { display:flex; flex-direction:column; justify-content:space-between; min-width:0; padding:clamp(28px,5.2vh,60px) clamp(28px,4.8vw,78px) clamp(28px,5vh,56px) 0; }
		.micro-label { display:flex; align-items:center; gap:12px; color:var(--muted); }
		.micro-label::before { content:""; width:34px; height:2px; background:var(--signal); }
		h1 { max-width:850px; margin:clamp(16px,3vh,34px) 0 18px; font-family:"Arial Narrow","Franklin Gothic Medium",Arial,sans-serif; font-size:clamp(54px,7.3vw,112px); font-stretch:condensed; font-weight:900; letter-spacing:-.07em; line-height:.82; text-transform:uppercase; }
		h1 .accent { color:var(--signal); }
		.lede { max-width:570px; margin:0; color:#393832; font-family:Georgia,"Times New Roman",serif; font-size:clamp(16px,1.35vw,21px); line-height:1.45; }
		.scan-form { max-width:760px; margin-top:clamp(24px,5vh,58px); }
		.form-label { display:block; margin-bottom:10px; }
		.input-row { display:grid; grid-template-columns:minmax(0,1fr) auto; background:#fffdf7; border:2px solid var(--ink); box-shadow:7px 7px 0 var(--ink); }
		.url-input { width:100%; min-width:0; padding:18px 20px; background:transparent; border:0; border-radius:0; color:var(--ink); font-family:"Courier New",monospace; font-size:15px; outline:none; }
		.url-input::placeholder { color:#8e8c82; }
		.url-input:focus { box-shadow:inset 0 -4px 0 var(--signal); }
		.scan-button { min-width:155px; padding:0 24px; background:var(--signal); border:0; border-left:2px solid var(--ink); border-radius:0; color:var(--ink); cursor:pointer; font-family:"Courier New",monospace; font-size:12px; font-weight:900; letter-spacing:.08em; text-transform:uppercase; }
		.scan-button:hover { background:var(--ink); color:var(--paper); }
		.scan-button:focus-visible { outline:3px solid var(--ink); outline-offset:3px; }
		.scan-button:disabled { cursor:wait; background:var(--ink); color:var(--paper); }
		.instrument { display:grid; grid-template-rows:auto minmax(0,1fr) auto; min-width:0; margin:18px 0 18px 18px; background:var(--ink); color:var(--paper); overflow:hidden; }
		.instrument-head,.instrument-foot { display:flex; align-items:center; justify-content:space-between; padding:17px 20px; }
		.instrument-head { border-bottom:1px solid rgba(241,238,229,.24); }
		.instrument-foot { border-top:1px solid rgba(241,238,229,.24); }
		.instrument-index { color:var(--signal); }
		.map { position:relative; min-height:0; overflow:hidden; }
		.map::before { content:""; position:absolute; inset:0; background-image:linear-gradient(rgba(241,238,229,.08) 1px,transparent 1px),linear-gradient(90deg,rgba(241,238,229,.08) 1px,transparent 1px); background-size:42px 42px; }
		.link-map { position:relative; z-index:1; width:100%; height:100%; }
		.map-line { fill:none; stroke:#7a7972; stroke-width:1.5; }
		.map-line.active { stroke:var(--signal); stroke-width:2.5; stroke-dasharray:8 8; animation:trace 8s linear infinite; }
		.map-node { fill:var(--ink); stroke:var(--paper); stroke-width:2; }
		.map-node.origin { fill:var(--signal); stroke:var(--signal); }
		.map-node.open { fill:var(--paper); }
		.map-text { fill:var(--paper); font-family:"Courier New",monospace; font-size:10px; font-weight:700; letter-spacing:.08em; }
		.map-number { fill:var(--signal); font-family:Georgia,serif; font-size:46px; }
		.method { display:grid; grid-template-columns:repeat(3,1fr); width:100%; }
		.method-step { padding:0 14px; border-left:1px solid rgba(241,238,229,.2); }
		.method-step:first-child { padding-left:0; border-left:0; }
		.method-number { display:block; margin-bottom:5px; color:var(--signal); font-family:Georgia,serif; font-size:19px; }
		.method-copy { display:block; color:#aaa89e; font-family:"Courier New",monospace; font-size:8px; line-height:1.45; text-transform:uppercase; }
		@keyframes trace { to { stroke-dashoffset:-64; } }
		@media (prefers-reduced-motion:reduce) { .map-line.active { animation:none; } }
		@media (max-width:800px) {
			.site-shell { grid-template-rows:64px minmax(0,1fr) 40px; padding:0 16px; }
			.stage { grid-template-columns:1fr; }
			.pitch { padding:28px 0; }
			.instrument { display:none; }
			h1 { max-width:680px; font-size:clamp(50px,15vw,96px); }
			.lede { font-size:17px; }
			.footer span:last-child { display:none; }
		}
		@media (max-width:520px) {
			.system-state { font-size:8px; }
			h1 { font-size:clamp(43px,14vw,72px); }
			.input-row { grid-template-columns:1fr; box-shadow:5px 5px 0 var(--ink); }
			.url-input { padding:14px; font-size:13px; }
			.scan-button { min-height:46px; border-top:2px solid var(--ink); border-left:0; }
		}
		@media (max-height:690px) and (min-width:801px) {
			.site-shell { grid-template-rows:60px minmax(0,1fr) 38px; }
			.pitch { padding-top:24px; padding-bottom:24px; }
			h1 { margin:18px 0 12px; font-size:clamp(52px,6.8vw,92px); }
			.scan-form { margin-top:22px; }
			.url-input { padding-top:14px; padding-bottom:14px; }
		}
	</style>
</head>
<body>
	<div class="site-shell">
		<header class="topbar">
			<div class="brand"><span class="brand-mark" aria-hidden="true"></span><span>Site / Check</span></div>
			<div class="system-state mono"><span class="state-dot" aria-hidden="true"></span><span>Scanner ready</span></div>
		</header>
		<main class="stage">
			<section class="pitch">
				<div>
					<div class="micro-label mono">Independent link audit</div>
					<h1>Find the links<br>your site <span class="accent">forgot.</span></h1>
					<p class="lede">Paste a starting URL. SiteCheck follows its internal pages, then tests every unique link with ten concurrent workers.</p>
				</div>
				<form class="scan-form" action="/scan" method="POST">
					<label class="form-label mono" for="url">Start with a complete address</label>
					<div class="input-row">
						<input class="url-input" id="url" type="url" name="url" placeholder="https://your-site.com" pattern="https?://.+" title="Enter a full address starting with http:// or https://" autocomplete="url" spellcheck="false" required autofocus>
						<button class="scan-button" type="submit">Run link audit</button>
					</div>
				</form>
			</section>
			<aside class="instrument" aria-label="How the scanner works">
				<div class="instrument-head mono"><span>Crawl topology</span><span class="instrument-index">10 workers / 10s</span></div>
				<div class="map">
					<svg class="link-map" viewBox="0 0 520 520" preserveAspectRatio="xMidYMid meet" aria-hidden="true">
						<path class="map-line active" d="M96 242 L216 142 L335 208 L438 110"></path>
						<path class="map-line" d="M96 242 L195 344 L330 370 L426 306"></path>
						<path class="map-line" d="M216 142 L195 344 M335 208 L330 370 M335 208 L426 306"></path>
						<circle class="map-node origin" cx="96" cy="242" r="10"></circle><circle class="map-node open" cx="216" cy="142" r="8"></circle><circle class="map-node" cx="335" cy="208" r="8"></circle><circle class="map-node open" cx="438" cy="110" r="8"></circle><circle class="map-node" cx="195" cy="344" r="8"></circle><circle class="map-node open" cx="330" cy="370" r="8"></circle><circle class="map-node" cx="426" cy="306" r="8"></circle>
						<text class="map-text" x="70" y="279">ORIGIN</text><text class="map-text" x="190" y="117">INTERNAL</text><text class="map-text" x="316" y="187">LINK</text><text class="map-text" x="399" y="87">STATUS</text><text class="map-number" x="62" y="92">01</text>
					</svg>
				</div>
				<div class="instrument-foot">
					<div class="method">
						<div class="method-step"><span class="method-number">A</span><span class="method-copy">Read the<br>start page</span></div>
						<div class="method-step"><span class="method-number">B</span><span class="method-copy">Follow internal<br>connections</span></div>
						<div class="method-step"><span class="method-number">C</span><span class="method-copy">Report every<br>response</span></div>
					</div>
				</div>
			</aside>
		</main>
		<footer class="footer"><span><strong>No account.</strong> One URL is enough.</span><span>HTTP / HTTPS · Internal + external links</span></footer>
	</div>
	<script>
		const form = document.querySelector('.scan-form')
		const button = document.querySelector('.scan-button')
		const state = document.querySelector('.system-state span:last-child')
		form.addEventListener('submit', () => {
			button.disabled = true
			button.textContent = 'Scanning...'
			state.textContent = 'Audit in progress'
		})
	</script>
</body>
</html>
`))

var reportTemplate = template.Must(template.New("report").Parse(`
<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<meta name="theme-color" content="#f1eee5">
	<title>SiteCheck — Audit Complete</title>
` + sharedStyles + `
	<style>
		.report-shell { display:grid; grid-template-rows:70px minmax(0,1fr); height:100dvh; padding:0 24px 24px; }
		.new-scan { padding:10px 14px; border:1px solid var(--ink); color:var(--ink); font-family:"Courier New",monospace; font-size:10px; font-weight:700; letter-spacing:.1em; text-decoration:none; text-transform:uppercase; }
		.new-scan:hover,.new-scan:focus-visible { background:var(--ink); color:var(--paper); outline:none; }
		.report-main { display:grid; grid-template-rows:auto auto minmax(0,1fr); min-height:0; border:1px solid var(--ink); }
		.report-heading { display:flex; align-items:end; justify-content:space-between; gap:24px; padding:clamp(22px,3.5vh,40px); border-bottom:1px solid var(--ink); }
		.eyebrow { margin-bottom:10px; color:var(--signal); }
		h1 { margin:0; font-family:"Arial Narrow","Franklin Gothic Medium",Arial,sans-serif; font-size:clamp(42px,6vw,88px); font-weight:900; letter-spacing:-.065em; line-height:.85; text-transform:uppercase; }
		.result-count { max-width:220px; color:var(--muted); line-height:1.6; text-align:right; }
		.metrics { display:grid; grid-template-columns:repeat(6,1fr); border-bottom:1px solid var(--ink); }
		.metric { min-width:0; padding:clamp(14px,2vw,24px); border-left:1px solid var(--ink); }
		.metric:first-child { border-left:0; background:var(--ink); color:var(--paper); }
		.metric-value { display:block; margin-bottom:7px; font-family:Georgia,"Times New Roman",serif; font-size:clamp(26px,3vw,45px); line-height:1; }
		.metric.alert .metric-value { color:var(--signal); }
		.metric-label { display:block; overflow:hidden; color:var(--muted); text-overflow:ellipsis; white-space:nowrap; }
		.metric:first-child .metric-label { color:#aaa89e; }
		.results { display:grid; grid-template-rows:42px minmax(0,1fr); min-height:0; }
		.results-head { display:flex; align-items:center; justify-content:space-between; padding:0 16px; border-bottom:1px solid var(--ink); }
		.results-note { color:var(--muted); font-family:Georgia,serif; font-size:12px; font-style:italic; }
		.table-wrap { min-height:0; overflow:auto; background:#fffdf7; }
		table { width:100%; border-collapse:collapse; table-layout:fixed; }
		th { position:sticky; top:0; z-index:1; padding:12px 16px; background:var(--paper); border-bottom:1px solid var(--ink); text-align:left; }
		th:nth-child(1) { width:68%; } th:nth-child(2) { width:12%; } th:nth-child(3) { width:20%; }
		td { padding:12px 16px; border-bottom:1px solid var(--rule); font-family:"Courier New",monospace; font-size:12px; vertical-align:middle; }
		tbody tr:hover { background:#e6e2d7; }
		.url-cell,.error-cell { overflow:hidden; text-overflow:ellipsis; white-space:nowrap; }
		.status { display:inline-block; min-width:48px; padding:4px 6px; background:var(--ink); color:var(--paper); text-align:center; }
		.error-cell { color:var(--signal); }
		.empty-state { padding:34px 16px; color:var(--muted); font-family:Georgia,"Times New Roman",serif; font-size:16px; font-style:italic; text-align:center; }
		::-webkit-scrollbar { width:10px; height:10px; }
		::-webkit-scrollbar-track { background:var(--paper); border-left:1px solid var(--ink); }
		::-webkit-scrollbar-thumb { background:var(--ink); }
		@media (max-width:760px) {
			.report-shell { grid-template-rows:58px minmax(0,1fr); padding:0 12px 12px; }
			.report-heading { align-items:start; padding:18px; }
			.result-count { display:none; }
			.metrics { grid-template-columns:repeat(3,1fr); }
			.metric { padding:11px; border-bottom:1px solid var(--ink); }
			.metric:nth-child(4) { border-left:0; }
			.metric:nth-child(n+4) { border-bottom:0; }
			.metric-value { font-size:24px; }
			.metric-label { font-size:8px; }
			th:nth-child(1) { width:55%; } th:nth-child(2) { width:20%; } th:nth-child(3) { width:25%; }
			th,td { padding:10px; }
		}
		@media (max-height:680px) and (min-width:761px) {
			.report-shell { grid-template-rows:58px minmax(0,1fr); padding-bottom:16px; }
			.report-heading { padding:20px 26px; }
			h1 { font-size:52px; }
			.metric { padding:13px 18px; }
			.metric-value { font-size:28px; }
		}
	</style>
</head>
<body>
	<div class="report-shell">
		<header class="topbar">
			<div class="brand"><span class="brand-mark" aria-hidden="true"></span><span>Site / Check</span></div>
			<a class="new-scan" href="/">New audit</a>
		</header>
		<main class="report-main">
			<section class="report-heading">
				<div><div class="eyebrow mono">Audit finished</div><h1>Scan complete.</h1></div>
				<div class="result-count mono">{{len .Results}} unique links answered the check</div>
			</section>
			<section class="metrics" aria-label="Scan summary">
				<div class="metric"><span class="metric-value">{{len .Results}}</span><span class="metric-label mono">Links checked</span></div>
				<div class="metric"><span class="metric-value">{{.Working}}</span><span class="metric-label mono">Working</span></div>
				<div class="metric"><span class="metric-value">{{.Redirects}}</span><span class="metric-label mono">Redirects</span></div>
				<div class="metric alert"><span class="metric-value">{{.Broken}}</span><span class="metric-label mono">Broken</span></div>
				<div class="metric alert"><span class="metric-value">{{.ServerErrors}}</span><span class="metric-label mono">Server errors</span></div>
				<div class="metric alert"><span class="metric-value">{{.Errors}}</span><span class="metric-label mono">Timeouts</span></div>
			</section>
			<section class="results">
				<div class="results-head"><span class="mono">Response ledger</span><span class="results-note">Every unique link, once.</span></div>
				<div class="table-wrap">
					<table>
						<thead><tr><th class="mono">URL</th><th class="mono">Status</th><th class="mono">Error</th></tr></thead>
						<tbody>
							{{range .Results}}
							<tr>
								<td class="url-cell" title="{{.Url}}">{{.Url}}</td>
								<td><span class="status mono">{{if .Err}}ERR{{else}}{{.Status}}{{end}}</span></td>
								<td class="error-cell" title="{{if .Err}}{{.Err}}{{end}}">{{if .Err}}{{.Err}}{{else}}—{{end}}</td>
							</tr>
							{{else}}
							<tr><td class="empty-state" colspan="3">No links were found on the starting page.</td></tr>
							{{end}}
						</tbody>
					</table>
				</div>
			</section>
		</main>
	</div>
</body>
</html>
`))
