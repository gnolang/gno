import { BaseController } from "./controller.js";

type PathsResponse = {
	realms: string[];
	packages: string[];
};

// Shapes returned by `<path>$search&…&json`. The controller knows no
// qualifier by name: the server owns the list, so a deployment without an
// indexer simply advertises fewer of them and the omnibar offers fewer.
type Selector = {
	name: string;
	hint: string;
	label: string;
	scope: string;
	source: string;
	bare?: boolean;
};

type SearchResult = {
	title: string;
	detail?: string;
	href?: string;
	tags?: string[];
};

type SearchGroup = {
	label: string;
	source: string;
	results: SearchResult[];
	error?: string;
};

type SearchResponse = {
	selectors?: Selector[];
	groups?: SearchGroup[];
	unknown_filter?: string;
};

type PageMatch = {
	node: Text;
	index: number;
	length: number;
	snippet: string;
};

const CONTENT_SELECTOR = "md-renderer, .b-source-code";
const SEARCH_ENDPOINT = "/search.json";
const MIN_QUERY_LENGTH = 2;
const RESULTS_PER_GROUP = 5;
const MAX_PAGE_MATCHES = 10;
const SNIPPET_RADIUS = 32;
const SEARCH_DELAY = 120;
// A qualified query costs a round trip, so it waits longer than the local
// path filter, which costs nothing.
const QUALIFIED_SEARCH_DELAY = 220;

// Full Amino object ID: 40 hex + ":" + uint index (ObjectID.MarshalAmino).
// Anchored to {40} to mirror the server's ValidateOID; shorter inputs fall
// through to a normal text search.
const OID_PATTERN = /^[a-f0-9]{40}:\d+$/i;

export class SearchbarController extends BaseController {
	private realms: string[] = [];
	private packages: string[] = [];
	private loading: Promise<void> | null = null;
	private loaded = false;
	private items: HTMLElement[] = [];
	private activeIndex = -1;
	private pageMatches: PageMatch[] = [];
	private highlightEl: HTMLElement | null = null;
	private runId = 0;
	private selectors: Selector[] = [];
	private selectorsLoaded = false;
	// null until the first probe answers. false means this deployment serves
	// no `$search` endpoint, and the controller stops asking.
	private searchAvailable: boolean | null = null;
	private selectorProbe: Promise<void> | null = null;
	private timer: ReturnType<typeof setTimeout> | undefined;

	protected connect(): void {
		this.initializeDOM({
			input: this.getTarget("input"),
			results: this.getTarget("results"),
		});

		const input = this.getDOMElement("input");
		input?.addEventListener("keydown", this.keynav.bind(this));
		input?.addEventListener("focus", this.selectInput.bind(this));
		document.addEventListener("click", this.onOutsideClick.bind(this));
		document.addEventListener("keydown", this.onKeyShortcut.bind(this));
	}

	// search filters the (once-fetched) path list for the current input, or
	// asks the server when the input carries a qualifier.
	public search(): void {
		const input = this.getDOMElement("input") as HTMLInputElement | null;
		this.schedule(input?.value.trim() ?? "");
	}

	// searchUrl keeps the bar usable as a direct path navigator on submit.
	// preventDefault is called only on a path the controller actually
	// handles. Calling it up front meant an unhandled input swallowed the
	// submit and did nothing at all — worst of all on a deployment with no
	// `$search` endpoint, where the form's own action would have worked.
	public searchUrl(e: Event): void {
		const input = this.getDOMElement("input") as HTMLInputElement | null;
		const raw = input?.value.trim();
		if (!raw) return;

		const go = (href: string): void => {
			e.preventDefault();
			window.location.href = href;
		};

		// OID-shaped input redirects to the state view for that object.
		if (OID_PATTERN.test(raw) && !raw.startsWith("/")) {
			const realmPath = this.currentRealmPath();
			if (realmPath) {
				go(`${realmPath}$state&oid=${encodeURIComponent(raw)}`);
				return;
			}
		}

		// A qualified query is a search, not a path: send it to the results
		// page, which is also the view a reader without JavaScript gets.
		if (this.isQualified(raw)) {
			go(this.searchHref(raw));
			return;
		}

		const target = SearchbarController.resolveTarget(raw);
		if (target) go(target);
	}

	private currentRealmPath(): string | null {
		const match = window.location.pathname.match(/^(\/r\/[^$]+)/);
		return match ? match[1] : null;
	}

	// keynav moves through results, opens the active one, or closes the list.
	public keynav(e: KeyboardEvent): void {
		switch (e.key) {
			case "ArrowDown":
				e.preventDefault();
				this.move(1);
				break;
			case "ArrowUp":
				e.preventDefault();
				this.move(-1);
				break;
			case "Enter":
				if (this.activeIndex >= 0) {
					e.preventDefault();
					this.items[this.activeIndex]?.click();
				}
				break;
			case "Escape":
				this.close();
				break;
		}
	}

	// One timer, with the delay chosen when it is scheduled. Two independent
	// debounce closures could not cancel each other, so the keystroke that
	// turned a query into a qualified one fired both of them: two identical
	// requests, and two rate-limit tokens for one search.
	private schedule(q: string): void {
		clearTimeout(this.timer);
		this.timer = setTimeout(
			() => {
				const input = this.getDOMElement("input") as HTMLInputElement | null;
				const current = input?.value.trim() ?? "";
				if (current.length < MIN_QUERY_LENGTH) {
					this.close();
					return;
				}
				void this.run(current);
			},
			this.isQualified(q) ? QUALIFIED_SEARCH_DELAY : SEARCH_DELAY,
		);
	}

	private async run(q: string): Promise<void> {
		const id = ++this.runId;
		const pageMatches = this.scanPage(q);

		if (this.isQualified(q)) {
			const res = await this.fetchSearch(q);
			if (id !== this.runId) return; // a newer keystroke superseded us
			if (res) {
				this.pageMatches = pageMatches;
				this.drawGroups(q, res, pageMatches);
				return;
			}
			// The endpoint did not answer: fall through to the local path
			// filter rather than showing the reader an empty dropdown.
		}

		await this.ensureLoaded();
		if (id !== this.runId) return;
		this.pageMatches = pageMatches;
		this.draw(q, this.filter(q), pageMatches);
	}

	// scopeBase strips any `$webquery` so `$search` hangs off the bare path.
	private scopeBase(): string {
		const p = window.location.pathname;
		const at = p.indexOf("$");
		return at >= 0 ? p.slice(0, at) : p;
	}

	private searchHref(q: string): string {
		return `${this.scopeBase()}$search&q=${encodeURIComponent(q)}`;
	}

	// isQualified reports whether a query carries a `key:value` token or names
	// a no-argument selector, i.e. whether the server has to answer it.
	private isQualified(q: string): boolean {
		if (this.searchAvailable === false) return false;
		// Mirrors ParseQuery exactly. The key must be a bare word: a gno path
		// with arguments (/r/gnoland/pages:p/about) carries a colon too, and
		// the bar is prefilled with the current path — treating that as a
		// search turned Enter from "go there" into "search for a qualifier
		// that does not exist".
		if (q.split(/\s+/).some(SearchbarController.isQualifierToken)) return true;
		const word = q.trim().toLowerCase();
		return this.selectors.some((s) => s.bare && s.name === word);
	}

	// isQualifierToken must stay in step with ParseQuery's isQualifierKey.
	static isQualifierToken(token: string): boolean {
		const at = token.indexOf(":");
		if (at < 0) return false;
		if (token.slice(at + 1).startsWith("//")) return false;
		return /^[a-z][a-z0-9_-]*$/.test(token.slice(0, at).toLowerCase());
	}

	// fetchSearch asks the server. A failure returns null and is not retried
	// for the life of the page when the endpoint itself is absent: an older
	// gnoweb, or one behind a proxy that does not route `$search`, must not
	// cost a request per keystroke.
	private async fetchSearch(q: string): Promise<SearchResponse | null> {
		try {
			const res = await fetch(`${this.searchHref(q)}&json`, {
				headers: { Accept: "application/json" },
			});
			if (res.status === 404) {
				this.searchAvailable = false;
				return null;
			}
			if (!res.ok) return null;
			const data = (await res.json()) as SearchResponse;
			this.searchAvailable = true;
			if (data.selectors) {
				this.selectors = data.selectors;
				this.selectorsLoaded = true;
			}
			return data;
		} catch {
			return null;
		}
	}

	// loadSelectors probes the endpoint once with an empty query so the hint
	// list reflects what this deployment can actually answer.
	private async loadSelectors(): Promise<void> {
		if (this.selectorsLoaded || this.searchAvailable === false) return;
		if (this.selectorProbe) return this.selectorProbe;
		// The flag is set by fetchSearch on success, never here: setting it
		// before the await meant one offline moment silently disabled every
		// bare qualifier for the life of the page, with no retry.
		this.selectorProbe = this.fetchSearch("")
			.then(() => undefined)
			.finally(() => {
				this.selectorProbe = null;
			});
		return this.selectorProbe;
	}

	// beginDraw resets the dropdown for a fresh render; endDraw reveals it.
	// Both render paths go through them so ARIA state cannot drift between
	// the local filter and the server-answered one.
	private beginDraw(): HTMLElement | null {
		const results = this.getDOMElement("results");
		if (!results) return null;
		this.clearHighlight();
		this.getDOMElement("input")?.removeAttribute("aria-activedescendant");
		results.textContent = "";
		this.items = [];
		this.activeIndex = -1;
		return results;
	}

	private endDraw(results: HTMLElement): void {
		// In the combobox pattern DOM focus stays on the input; options are
		// reached with the arrow keys. Without this, Tab walks every result.
		for (const el of this.items) el.tabIndex = -1;
		results.hidden = false;
		this.getDOMElement("input")?.setAttribute("aria-expanded", "true");
	}

	private drawGroups(
		q: string,
		res: SearchResponse,
		pageMatches: PageMatch[],
	): void {
		const results = this.beginDraw();
		if (!results) return;

		if (res.unknown_filter) {
			results.appendChild(
				this.buildNotice(
					`"${res.unknown_filter}:" is not a search qualifier here`,
				),
			);
		}

		for (const group of res.groups ?? []) {
			const label =
				group.source === "indexer"
					? `${group.label}, from indexer`
					: group.label;
			const section = this.sectionWithLabel(label);

			if (group.error) {
				section.appendChild(this.buildNotice(group.error));
			}

			for (const r of group.results ?? []) {
				section.appendChild(this.buildRemoteItem(r, q, group.source));
			}
			results.appendChild(section);
		}

		// A qualified query always offers its own results page, which is the
		// only view that works without JavaScript.
		results.appendChild(this.buildFullPageItem(q));

		if (pageMatches.length > 0) {
			results.appendChild(this.buildPageSection(q, pageMatches));
		}

		this.endDraw(results);
	}

	// buildNotice renders a message inside the listbox. It is an option
	// because role="listbox" owns only options and groups — and a disabled
	// one, because it is not reachable by the arrow keys and must not inflate
	// the option count a screen reader announces.
	private buildNotice(text: string): HTMLElement {
		const el = document.createElement("div");
		el.className = "b-omnisearch-empty";
		el.setAttribute("role", "option");
		el.setAttribute("aria-disabled", "true");
		el.textContent = text;
		return el;
	}

	// Provenance is a property of the group, not of the row: the server marks
	// the group once and every row in it inherits that mark.
	private buildRemoteItem(
		r: SearchResult,
		q: string,
		source: string,
	): HTMLElement {
		// A result without a destination (a transaction has no page in gnoweb)
		// renders as a non-interactive row rather than as a dead link.
		const item = document.createElement(r.href ? "a" : "div");
		item.className = "b-omnisearch-item";
		item.setAttribute("role", "option");
		// A row with no destination (a transaction has no page in gnoweb) is
		// announced but never reachable by the arrow keys, so it must say so
		// rather than inflating the option count.
		if (r.href?.startsWith("/")) {
			(item as HTMLAnchorElement).href = r.href;
		} else {
			item.setAttribute("aria-disabled", "true");
		}
		item.setAttribute(
			"aria-label",
			source === "indexer" ? `from indexer: ${r.title}` : r.title,
		);

		if (source === "indexer") {
			const tag = document.createElement("span");
			tag.className = "b-omnisearch-tag";
			tag.dataset.type = "i";
			tag.textContent = "i";
			tag.title = "From an indexer, not from the chain";
			// The row's aria-label already says it; a bare "i" in the
			// accessible name reads as noise.
			tag.setAttribute("aria-hidden", "true");
			item.appendChild(tag);
		}

		const text = document.createElement("span");
		text.className = "b-omnisearch-text";
		this.fillHighlighted(text, r.title, q);
		if (r.detail) {
			const detail = document.createElement("small");
			detail.textContent = ` ${r.detail}`;
			text.appendChild(detail);
		}
		item.appendChild(text);

		if (r.href?.startsWith("/")) this.items.push(item);
		return item;
	}

	private buildFullPageItem(q: string): HTMLElement {
		const section = this.sectionWithLabel("All results");
		const item = document.createElement("a");
		item.className = "b-omnisearch-item";
		item.setAttribute("role", "option");
		item.href = this.searchHref(q);
		const text = document.createElement("span");
		text.className = "b-omnisearch-text";
		text.textContent = `Open the full results page for "${q}"`;
		item.appendChild(text);
		section.appendChild(item);
		this.items.push(item);
		return section;
	}

	// ensureLoaded fetches the path list once and reuses it for every keystroke.
	// Concurrent callers share the single in-flight request. On failure `loaded`
	// stays false so the next keystroke retries instead of silently giving up.
	private ensureLoaded(): Promise<void> {
		if (this.loaded) return Promise.resolve();
		if (this.loading) return this.loading;
		this.loading = fetch(SEARCH_ENDPOINT)
			.then((res) => {
				if (!res.ok) throw new Error(`search.json: HTTP ${res.status}`);
				return res.json();
			})
			.then((data: Partial<PathsResponse>) => {
				this.realms = data.realms ?? [];
				this.packages = data.packages ?? [];
				this.loaded = true;
			})
			.catch(() => {
				// swallow: leave loaded=false so a later keystroke can retry
			})
			.finally(() => {
				this.loading = null;
			});
		return this.loading;
	}

	// filter ranks the cached lists by relevance and keeps the top results per
	// group. Users are derived from the first path segment of the ranked matches.
	private filter(q: string): {
		apps: string[];
		packages: string[];
		users: string[];
	} {
		const needle = q.toLowerCase();
		const apps = SearchbarController.rank(this.realms, needle);
		const packages = SearchbarController.rank(this.packages, needle);

		const users: string[] = [];
		const seen = new Set<string>();
		for (const p of apps.concat(packages)) {
			const name = SearchbarController.firstSegment(p);
			if (name && !seen.has(name)) {
				seen.add(name);
				users.push(`/u/${name}`);
			}
		}

		return {
			apps: apps.slice(0, RESULTS_PER_GROUP),
			packages: packages.slice(0, RESULTS_PER_GROUP),
			users: users.slice(0, RESULTS_PER_GROUP),
		};
	}

	// rank returns the paths matching needle, most-relevant first: a match in the
	// realm name (last segment) outranks one in the namespace/address, and a
	// shallower path edges out a deeper one.
	static rank(paths: string[], needle: string): string[] {
		const scored: Array<{ path: string; score: number }> = [];
		for (const p of paths) {
			const score = SearchbarController.relevance(p, needle);
			if (score > 0) scored.push({ path: p, score });
		}
		scored.sort(
			(a, b) =>
				b.score - a.score ||
				a.path.length - b.path.length ||
				a.path.localeCompare(b.path),
		);
		return scored.map((s) => s.path);
	}

	// relevance scores how well a path matches needle; 0 means no match.
	static relevance(path: string, needle: string): number {
		const lower = path.toLowerCase();
		if (!lower.includes(needle)) return 0;
		const segs = lower.split("/").filter(Boolean);
		const name = segs[segs.length - 1] ?? "";
		let score: number;
		if (name === needle) score = 100;
		else if (name.startsWith(needle)) score = 80;
		else if (name.includes(needle)) score = 60;
		else if (segs.some((s) => s.startsWith(needle))) score = 40;
		else score = 20;
		return score - segs.length;
	}

	private draw(
		q: string,
		groups: { apps: string[]; packages: string[]; users: string[] },
		pageMatches: PageMatch[],
	): void {
		const results = this.beginDraw();
		if (!results) return;

		const gnoland = this.buildLinkSection("gno.land", q, [
			["Apps", groups.apps],
			["Packages", groups.packages],
			["Users", groups.users],
		]);
		if (gnoland) results.appendChild(gnoland);

		if (pageMatches.length > 0) {
			results.appendChild(this.buildPageSection(q, pageMatches));
		}

		if (this.items.length === 0) {
			results.appendChild(this.buildNotice("No results"));
		}
		this.endDraw(results);
	}

	private buildLinkSection(
		title: string,
		q: string,
		groups: Array<[string, string[]]>,
	): HTMLElement | null {
		const filled = groups.filter(([, paths]) => paths.length > 0);
		if (filled.length === 0) return null;

		const section = this.sectionWithLabel(title);
		for (const [label, paths] of filled) {
			for (const path of paths) {
				if (!path.startsWith("/")) continue;
				const item = document.createElement("a");
				item.className = "b-omnisearch-item";
				item.href = path;
				item.setAttribute("role", "option");
				item.setAttribute("aria-label", `${label}: ${path}`);

				const type = SearchbarController.typeOf(path);
				if (type) {
					const tag = document.createElement("span");
					tag.className = "b-omnisearch-tag";
					tag.dataset.type = type;
					tag.textContent = type;
					item.appendChild(tag);
				}

				const text = document.createElement("span");
				text.className = "b-omnisearch-text";
				this.fillHighlighted(text, path, q);
				item.appendChild(text);

				section.appendChild(item);
				this.items.push(item);
			}
		}
		return section;
	}

	private buildPageSection(q: string, matches: PageMatch[]): HTMLElement {
		const section = this.sectionWithLabel(
			`This page · ${matches.length} matches`,
		);
		matches.forEach((match, i) => {
			const item = document.createElement("button");
			item.type = "button";
			item.className = "b-omnisearch-item";
			item.setAttribute("role", "option");

			const text = document.createElement("span");
			text.className = "b-omnisearch-text";
			this.fillHighlighted(text, match.snippet, q);
			item.appendChild(text);

			item.addEventListener("click", () => this.reveal(i));
			section.appendChild(item);
			this.items.push(item);
		});
		return section;
	}

	private sectionWithLabel(title: string): HTMLElement {
		const section = document.createElement("div");
		section.className = "b-omnisearch-section";
		section.setAttribute("role", "group");
		section.setAttribute("aria-label", title);
		const label = document.createElement("span");
		label.className = "b-omnisearch-label";
		label.textContent = title;
		// role="group" owns only options; the group already carries this
		// text as its aria-label, so the span is decoration.
		label.setAttribute("aria-hidden", "true");
		section.appendChild(label);
		return section;
	}

	// fillHighlighted writes text into el, emphasizing the first match of q.
	private fillHighlighted(el: HTMLElement, text: string, q: string): void {
		const at = text.toLowerCase().indexOf(q.toLowerCase());
		if (at < 0) {
			el.textContent = text;
			return;
		}
		el.appendChild(document.createTextNode(text.slice(0, at)));
		const mark = document.createElement("mark");
		mark.textContent = text.slice(at, at + q.length);
		el.appendChild(mark);
		el.appendChild(document.createTextNode(text.slice(at + q.length)));
	}

	private move(delta: number): void {
		if (this.items.length === 0) return;
		const prev = this.items[this.activeIndex];
		prev?.removeAttribute("aria-selected");
		prev?.removeAttribute("id");
		this.activeIndex =
			(this.activeIndex + delta + this.items.length) % this.items.length;
		const active = this.items[this.activeIndex];
		if (!active) return;
		active.id = `omnisearch-opt-${this.activeIndex}`;
		active.setAttribute("aria-selected", "true");
		active.scrollIntoView({ block: "nearest" });
		this.getDOMElement("input")?.setAttribute(
			"aria-activedescendant",
			active.id,
		);
	}

	private close(): void {
		this.runId++;
		const results = this.getDOMElement("results");
		if (results) {
			results.hidden = true;
			results.textContent = "";
		}
		this.items = [];
		this.activeIndex = -1;
		const input = this.getDOMElement("input");
		input?.setAttribute("aria-expanded", "false");
		input?.removeAttribute("aria-activedescendant");
	}

	private onOutsideClick(e: MouseEvent): void {
		if (!this.element.contains(e.target as Node)) this.close();
	}

	// onKeyShortcut focuses the search bar when "/" is pressed outside any
	// editable element, then opens the dropdown against the prefilled path so
	// the action has visible feedback (a bare select() reads as a no-op).
	// Skips when modifiers are held or during IME composition so chorded
	// shortcuts and CJK input keep working.
	//
	// TODO: extract to a shared keyboard-shortcut helper on BaseController when
	// a second controller needs a global key binding.
	private onKeyShortcut(e: KeyboardEvent): void {
		if (e.key !== "/" || e.ctrlKey || e.metaKey || e.altKey || e.isComposing)
			return;
		const target = e.target as HTMLElement | null;
		if (target?.matches?.("input, textarea, [contenteditable='true']")) return;
		e.preventDefault();
		(this.getDOMElement("input") as HTMLInputElement | null)?.focus();
		this.search();
	}

	// selectInput selects the whole input on focus so typing replaces the
	// prefilled path instead of appending to it. Deferred via rAF so a mouse
	// click's caret placement doesn't collapse the selection (a plain
	// synchronous select() is lost on mousedown→focus→mouseup).
	private selectInput(): void {
		const input = this.getDOMElement("input") as HTMLInputElement | null;
		requestAnimationFrame(() => input?.select());
		// Learn the qualifier list on first focus, not on page load: a reader
		// who never opens the bar costs no request.
		void this.loadSelectors();
	}

	// scanPage collects up to MAX_PAGE_MATCHES occurrences of q in the page
	// content region, each with a short surrounding snippet.
	private scanPage(q: string): PageMatch[] {
		const root = document.querySelector<HTMLElement>(CONTENT_SELECTOR);
		if (!root) return [];

		const needle = q.toLowerCase();
		const matches: PageMatch[] = [];
		const walker = document.createTreeWalker(root, NodeFilter.SHOW_TEXT);
		let node = walker.nextNode() as Text | null;
		while (node && matches.length < MAX_PAGE_MATCHES) {
			const text = node.nodeValue ?? "";
			const at = text.toLowerCase().indexOf(needle);
			if (at >= 0) {
				matches.push({
					node,
					index: at,
					length: q.length,
					snippet: SearchbarController.snippet(text, at, q.length),
				});
			}
			node = walker.nextNode() as Text | null;
		}
		return matches;
	}

	// reveal scrolls to a page match and wraps it in a transient highlight.
	// The DOM may have changed since scan (other controllers, re-render); skip
	// silently if the captured range is no longer valid.
	private reveal(i: number): void {
		const match = this.pageMatches[i];
		if (!match?.node.parentNode) return;
		this.clearHighlight();

		const end = Math.min(match.index + match.length, match.node.length);
		if (end <= match.index) {
			this.close();
			return;
		}

		const mark = document.createElement("mark");
		mark.className = "b-omnisearch-hl";
		try {
			const range = document.createRange();
			range.setStart(match.node, match.index);
			range.setEnd(match.node, end);
			range.surroundContents(mark);
		} catch {
			this.close();
			return;
		}
		this.highlightEl = mark;
		mark.scrollIntoView({ block: "center" });
		this.close();
	}

	private clearHighlight(): void {
		const mark = this.highlightEl;
		if (!mark?.parentNode) return;
		const parent = mark.parentNode;
		while (mark.firstChild) parent.insertBefore(mark.firstChild, mark);
		parent.removeChild(mark);
		parent.normalize();
		this.highlightEl = null;
	}

	static snippet(text: string, at: number, len: number): string {
		const start = Math.max(0, at - SNIPPET_RADIUS);
		const end = Math.min(text.length, at + len + SNIPPET_RADIUS);
		return (
			(start > 0 ? "…" : "") +
			text.slice(start, end).trim() +
			(end < text.length ? "…" : "")
		);
	}

	// typeOf returns the single-letter on-chain kind of a path: r, p, or u.
	static typeOf(path: string): string {
		if (path.startsWith("/r/")) return "r";
		if (path.startsWith("/p/")) return "p";
		if (path.startsWith("/u/")) return "u";
		return "";
	}

	// firstSegment returns the path element after /r/ or /p/,
	// e.g. "/r/demo/boards" -> "demo".
	static firstSegment(p: string): string {
		const parts = p.replace(/^\//, "").split("/");
		return parts.length >= 2 ? parts[1] : "";
	}

	// resolveTarget strips a leading `gno.land` host (with or without scheme) so
	// realm paths copied from anywhere resolve locally; non-`gno.land` absolute
	// URLs pass through, and relatives resolve against the origin.
	static resolveTarget(input: string): string | null {
		const stripped = input.replace(
			/^(?:https?:\/\/)?gno\.land(?=\/|$|\?|#)/i,
			"",
		);
		try {
			const url = new URL(stripped, window.location.origin);
			if (url.protocol !== "http:" && url.protocol !== "https:") return null;
			return url.href;
		} catch {
			return null;
		}
	}
}
