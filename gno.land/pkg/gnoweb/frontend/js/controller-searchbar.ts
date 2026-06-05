import { BaseController, debounce } from "./controller.js";

type PathsResponse = {
	realms: string[];
	packages: string[];
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

	// search filters the (once-fetched) path list for the current input.
	public search(): void {
		this.debouncedSearch();
	}

	// searchUrl keeps the bar usable as a direct path navigator on submit.
	public searchUrl(e: Event): void {
		e.preventDefault();
		const input = this.getDOMElement("input") as HTMLInputElement | null;
		const raw = input?.value.trim();
		if (!raw) return;
		const target = SearchbarController.resolveTarget(raw);
		if (target) window.location.href = target;
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

	private debouncedSearch = debounce((): void => {
		const input = this.getDOMElement("input") as HTMLInputElement | null;
		const q = input?.value.trim() ?? "";
		if (q.length < MIN_QUERY_LENGTH) {
			this.close();
			return;
		}
		void this.run(q);
	}, SEARCH_DELAY);

	private async run(q: string): Promise<void> {
		const id = ++this.runId;
		const pageMatches = this.scanPage(q);
		await this.ensureLoaded();
		if (id !== this.runId) return; // a newer keystroke superseded us
		this.pageMatches = pageMatches;
		this.draw(q, this.filter(q), pageMatches);
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
		const results = this.getDOMElement("results");
		const input = this.getDOMElement("input");
		if (!results) return;

		this.clearHighlight();
		input?.removeAttribute("aria-activedescendant");
		results.textContent = "";
		this.items = [];
		this.activeIndex = -1;

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
			const empty = document.createElement("div");
			empty.className = "b-omnisearch-empty";
			empty.textContent = "No results";
			results.appendChild(empty);
		}
		results.hidden = false;
		input?.setAttribute("aria-expanded", "true");
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
	// editable element (mirrors GitHub/Linear/Vercel). Skips when modifiers are
	// held or during IME composition, so chorded shortcuts and CJK input keep
	// working. The existing focus handler then selects the prefilled path.
	//
	// TODO: extract to a shared keyboard-shortcut helper on BaseController when
	// a second controller needs a global key binding (sole consumer for now —
	// premature centralization would be YAGNI).
	private onKeyShortcut(e: KeyboardEvent): void {
		if (e.key !== "/" || e.ctrlKey || e.metaKey || e.altKey || e.isComposing)
			return;
		const target = e.target as HTMLElement | null;
		if (target?.matches?.("input, textarea, [contenteditable='true']")) return;
		e.preventDefault();
		(this.getDOMElement("input") as HTMLInputElement | null)?.focus();
	}

	// selectInput selects the whole input on focus so typing replaces the
	// prefilled path instead of appending to it. Deferred via rAF so a mouse
	// click's caret placement doesn't collapse the selection (a plain
	// synchronous select() is lost on mousedown→focus→mouseup).
	private selectInput(): void {
		const input = this.getDOMElement("input") as HTMLInputElement | null;
		requestAnimationFrame(() => input?.select());
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
	// URLs pass through, and relatives resolve against the origin. Uses
	// `new URL` over the (Baseline 2024) `URL.parse` for older-browser reach.
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
