// CSP-safe SimpleAnalytics metadata loader. Loaded as a synchronous classic
// script so it blocks HTML parsing until executed — guaranteeing that
// window.sa_metadata is set before SA's async latest.js can even start
// fetching (async/defer scripts can only begin loading after the parser
// resumes). Values are read from data-* attributes on this script tag,
// which lets the server inject them without an inline <script>.

(() => {
	const el = document.currentScript as HTMLScriptElement | null;
	if (!el) return;
	window.sa_metadata = {
		page_type: el.dataset.pageType ?? "",
		chain_id: el.dataset.chainId ?? "",
	};
})();
