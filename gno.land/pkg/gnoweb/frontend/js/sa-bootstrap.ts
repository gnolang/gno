// CSP-safe SimpleAnalytics metadata loader. Loaded as a synchronous classic
// script so it blocks HTML parsing until executed — guaranteeing that
// window.sa_metadata and the path-overwriter are set before SA's async
// latest.js can even start fetching (async/defer scripts can only begin
// loading after the parser resumes). Values are read from data-* attributes on
// this script tag, which lets the server inject them without an inline <script>.

(() => {
	const el = document.currentScript;
	if (!(el instanceof HTMLScriptElement)) return;
	window.sa_metadata = {
		// Bump when the event taxonomy or metadata shape changes incompatibly,
		// so consumers can distinguish shapes (see SIMPLEANALYTICS.md, #5467).
		schema_version: "1",
		page_type: el.dataset.pageType ?? "",
		chain_id: el.dataset.chainId ?? "",
	};
	// Report the server-built path instead of the raw URL, whose pathname can
	// carry user-supplied function arguments. An empty value lets SA fall back
	// to its default path.
	const saPath = el.dataset.saPath ?? "";
	window.gnoSaPath = () => saPath;
})();
