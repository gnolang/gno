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
	// What the keyword does, shown beside its hint.
	help: string;
	scope: string;
	// The heading of one that works anywhere: a KeywordGroup key.
	group?: string;
	source: string;
	bare?: boolean;
	// A value this selector answers typed alone (a pasted address), as a
	// pattern. The server matches the same one.
	accepts?: string;
};

// A narrowing qualifier (`author:`, `is:`), also listed by the server. It
// names where its values come from, so the controller needs no list of its
// own: a fixed set, or one the bar already holds.
type Filter = {
	name: string;
	hint: string;
	help: string;
	group: string;
	values?: string[];
	complete?: string;
};

// KeywordGroup heads the keywords that work anywhere, in the order shown;
// the results page's popover uses the same list.
type KeywordGroup = {
	key: string;
	label: string;
};

// Keyword is what the omnibar suggests: a selector, or a filter, which works
// anywhere and is answered by the chain.
type Keyword = Selector & Pick<Filter, "values" | "complete">;

// Choice is a suggestion row. Picking it inserts `complete` rather than
// following a link, or with `run` searches for it at once; Tab only inserts.
type Choice = {
	title: string;
	detail?: string;
	indexer?: boolean;
	complete: string;
	run?: boolean;
};

// ChoiceSection is a group of suggestions. Provenance is a row's, never
// the group's: an indexer row carries its own pill, in the marks column.
type ChoiceSection = {
	label: string;
	choices: Choice[];
	notice?: string;
	highlight?: string;
};

// Token is a span of the input: the selection, or the word under the caret.
type Token = { start: number; end: number; text: string };

type SearchResult = {
	title: string;
	detail?: string;
	href?: string;
	tags?: string[];
};

type SearchGroup = {
	label: string;
	source: string;
	results?: SearchResult[];
	error?: string;
	// notice says why a partial answer is partial; truncated marks one.
	notice?: string;
	truncated?: boolean;
};

type SearchResponse = {
	selectors?: Selector[];
	filters?: Filter[];
	keyword_groups?: KeywordGroup[];
	groups?: SearchGroup[];
	unknown_filter?: string;
	// error is the server refusing the query (429, 400), not a group's.
	error?: string;
};

// isSearchResponse checks the fields the renderer dereferences. The answer
// crosses a network (a proxy, a version skew), and one malformed field used
// to throw on every keystroke for the life of the page.
function isSearchResponse(v: unknown): v is SearchResponse {
	const obj = (x: unknown): x is Record<string, unknown> =>
		typeof x === "object" && x !== null;
	const list = (x: unknown, ok: (e: unknown) => boolean): boolean =>
		x === undefined || (Array.isArray(x) && x.every(ok));
	const str = (x: unknown): boolean => typeof x === "string";
	const opt = (x: unknown): boolean => x === undefined || str(x);
	return (
		obj(v) &&
		list(
			v.selectors,
			(s) =>
				obj(s) && str(s.name) && str(s.hint) && str(s.help) && opt(s.group),
		) &&
		list(
			v.filters,
			(f) =>
				obj(f) &&
				str(f.name) &&
				str(f.hint) &&
				str(f.help) &&
				str(f.group) &&
				opt(f.complete) &&
				list(f.values, str),
		) &&
		list(v.keyword_groups, (g) => obj(g) && str(g.key) && str(g.label)) &&
		list(
			v.groups,
			(g) =>
				obj(g) &&
				list(
					g.results,
					(r) =>
						obj(r) &&
						typeof r.title === "string" &&
						(r.href === undefined || typeof r.href === "string"),
				),
		)
	);
}

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
// Keyword rows sit above the results while typing: a hint, not the answer.
const MAX_KEYWORD_ROWS = 4;
// Provenance, in a group's label or spoken on a row.
const FROM_INDEXER = "from the indexer";
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

// A bech32 account address, as a namespace: ranked as a throwaway deploy.
// Same shape as omnisearch addressPattern.
const ADDRESS_PATTERN = /^g1[02-9ac-hj-np-z]{38}$/;

// A version segment such as v0 or v12, skipped for a package's name.
const VERSION_SEGMENT = /^v(0|[1-9]\d*)$/;

// A leading `gno.land` host, with or without scheme.
const GNO_LAND_HOST = /^(?:https?:\/\/)?gno\.land(?=\/|$|\?|#)/i;

export class SearchbarController extends BaseController {
	private realms: string[] = [];
	private packages: string[] = [];
	// The namespaces of those paths, for `author:`.
	private namespaces: string[] = [];
	private loading: Promise<void> | null = null;
	private loaded = false;
	private items: HTMLElement[] = [];
	private activeIndex = -1;
	private pageMatches: PageMatch[] = [];
	private highlightEl: HTMLElement | null = null;
	private runId = 0;
	private selectors: Selector[] = [];
	private filters: Filter[] = [];
	private groups: KeywordGroup[] = [];
	// The open list is the keyword help, not an answer to typing: Tab must
	// leave the bar there, not run the first keyword.
	private helping = false;
	// The token the list was drawn for, which a choice replaces, and whether
	// a keystroke still waits for its search: Tab completes only from a list
	// that answers what is typed. Null in the help, and once the input
	// changes: a choice then replaces the token under the caret.
	private listToken: Token | null = null;
	private pending = false;
	private selectorsLoaded = false;
	// null until the first probe answers. false means this deployment serves
	// no `$search` endpoint, and the controller stops asking.
	private searchAvailable: boolean | null = null;
	private selectorProbe: Promise<void> | null = null;
	private timer: ReturnType<typeof setTimeout> | undefined;
	// The in-flight `$search&json` request. Aborting it also cancels the
	// server's r.Context(), which matters for indexer-backed selectors.
	private inflight: AbortController | null = null;

	protected connect(): void {
		this.initializeDOM({
			input: this.getTarget("input"),
			results: this.getTarget("results"),
		});

		const input = this.getDOMElement("input");
		input?.addEventListener("keydown", this.keynav.bind(this));
		input?.addEventListener("focus", this.selectInput.bind(this));
		// Focus leaving the omnibar closes it: a click outside, Tab out.
		this.element.addEventListener("focusout", (e) => {
			if (!this.element.contains(e.relatedTarget as Node | null)) this.close();
		});
		(input as HTMLInputElement | null)?.form?.addEventListener(
			"mousedown",
			this.focusFromForm.bind(this),
		);
		// Focus stays in the input, as the combobox pattern wants, so a choice
		// can edit it and leave the caret where the reader types next.
		const results = this.getDOMElement("results");
		results?.addEventListener("mousedown", (e) => e.preventDefault());
		// The pointer moves the one highlight the arrow keys move. Not
		// pointerover: a list redrawn under a still pointer would take it.
		results?.addEventListener("pointermove", (e) => {
			const row = (e.target as Element).closest(".b-omnisearch-item");
			const i = this.items.indexOf(row as HTMLElement);
			if (i >= 0 && i !== this.activeIndex) this.activate(i);
		});
		document.addEventListener("keydown", this.onKeyShortcut.bind(this));
	}

	// search filters the (once-fetched) path list for the current input, or
	// asks the server when the input carries a qualifier.
	public search(): void {
		const input = this.getDOMElement("input") as HTMLInputElement | null;
		// The list on screen no longer answers the input: Enter before the
		// next render must submit what was typed, not open an old row.
		this.deselect();
		this.activeIndex = -1;
		this.listToken = null;
		input?.removeAttribute("aria-activedescendant");
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
			clearTimeout(this.timer);
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

		// A qualified query, or free text that names no path, is a search:
		// send it to the results page, which is also the view a reader without
		// JavaScript gets. Navigating `counter` to `/counter` was a dead end.
		if (
			this.isQualified(raw) ||
			(this.searchAvailable !== false && !SearchbarController.isPathLike(raw))
		) {
			go(this.searchHref(raw));
			return;
		}

		const target = SearchbarController.resolveTarget(raw);
		if (target) go(target);
	}

	// isPathLike reports input the bar navigates to rather than searches: a
	// path (with or without its leading slash), an anchor or query (resolved
	// against the site root), a gno.land link, or a URL.
	static isPathLike(raw: string): boolean {
		return (
			/^[rpu]\//.test(raw) ||
			/^[/#?]/.test(raw) ||
			GNO_LAND_HOST.test(raw) ||
			/^[a-z][a-z0-9+.-]*:\/\//i.test(raw)
		);
	}

	private currentRealmPath(): string | null {
		const path = this.packagePath();
		return path.startsWith("/r/") ? path : null;
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
			case "Tab": {
				if (e.shiftKey) break;
				// Tab completes, as in a shell: the highlighted choice, else
				// the first one typing produced. Not from a list drawn for
				// another word, or before the last keystroke's search.
				const fresh =
					!this.pending &&
					(this.helping || this.caretToken().text === this.listToken?.text);
				if (!fresh) break;
				const item =
					this.items[this.activeIndex] ??
					(this.helping ? undefined : this.items[0]);
				const text = item?.dataset.complete;
				if (text !== undefined) {
					e.preventDefault();
					this.complete(text);
				}
				break;
			}
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
		this.inflight?.abort();
		this.runId++;
		this.pending = true;
		this.timer = setTimeout(
			() => {
				this.pending = false;
				const input = this.getDOMElement("input") as HTMLInputElement | null;
				const current = input?.value.trim() ?? "";
				if (current.length < MIN_QUERY_LENGTH) {
					if (!current && this.showKeywords()) return;
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
		const token = this.caretToken();
		const { lead, waiting } = await this.suggest();
		if (id !== this.runId) return;
		// A query the server must refuse is not sent: the dropdown says what
		// to add, unless the keyword under the caret already does.
		const refusal = this.isQualified(q) ? this.unsendable(q) : "";
		if (refusal) {
			this.render(q, [], [], {
				lead,
				token,
				waiting: true,
				notice: waiting ? undefined : refusal,
			});
			return;
		}
		const pageMatches = this.scanPage(q);

		if (this.isQualified(q)) {
			const ctl = new AbortController();
			this.inflight = ctl;
			const res = await this.fetchSearch(q, ctl.signal);
			// A newer keystroke, or close(), superseded us.
			if (id !== this.runId || ctl.signal.aborted) return;
			if (res) {
				this.pageMatches = pageMatches;
				this.render(q, res.groups ?? [], pageMatches, {
					unknownFilter: res.unknown_filter,
					error: res.error,
					fullPage: true,
					highlight: this.highlightTerm(q),
					lead,
					token,
				});
				return;
			}
			// The endpoint did not answer: fall through to the local path
			// filter rather than showing the reader an empty dropdown.
		}

		await this.ensureLoaded();
		if (id !== this.runId) return;
		this.pageMatches = pageMatches;
		this.render(q, this.filter(q), pageMatches, { lead, token });
	}

	// keywords are the selectors and filters the reader can use here: those
	// about a package only on its page, or once `in:` names one.
	private keywords(): Keyword[] {
		const input = this.getDOMElement("input") as HTMLInputElement | null;
		const scoped =
			this.packagePath() !== "" ||
			(input?.value ?? "")
				.split(/\s+/)
				.some(
					(t) => SearchbarController.isQualifierToken(t) && /^in:./i.test(t),
				);
		return [
			...this.selectors,
			...this.filters.map((f) => ({
				...f,
				label: f.help,
				scope: "global",
				source: "chain",
			})),
		].filter((k) => scoped || k.scope !== "package");
	}

	// packagePath reads the form's action, which the server points at
	// `<path>$search` on a realm or package page.
	private packagePath(): string {
		const input = this.getDOMElement("input") as HTMLInputElement | null;
		const action = input?.form?.getAttribute("action") ?? "";
		return /^(\/[rp]\/.+)\$search$/.exec(action)?.[1] ?? "";
	}

	// showKeywords opens an untouched bar on what it can do: on a package
	// page the keywords about it, those that run alone first; then a few
	// that work anywhere, under what they do, and the page that lists them
	// all with examples. It reports whether there was anything to show.
	private showKeywords(): boolean {
		const all = this.keywords();
		if (all.length === 0) return false;
		const local = all.filter((k) => k.scope === "package");
		const sections: ChoiceSection[] = [];
		let room = Infinity;
		if (this.packagePath() && local.length > 0) {
			local.sort((a, b) => Number(!!b.bare) - Number(!!a.bare));
			sections.push({
				label: "On this package",
				choices: local.map((k) => this.keywordChoice(k)),
			});
			room = MAX_KEYWORD_ROWS;
		}
		for (const g of this.groups) {
			const ks = all
				.filter((k) => k.scope !== "package" && k.group === g.key)
				.slice(0, room);
			room -= ks.length;
			if (ks.length > 0) {
				sections.push({
					label: g.label,
					choices: ks.map((k) => this.keywordChoice(k)),
				});
			}
		}

		// A run still waiting on the network must not draw over the help.
		this.runId++;
		this.inflight?.abort();
		const results = this.beginDraw();
		if (!results) return false;
		results.appendChild(this.buildKeywordsHead());
		for (const s of sections) results.appendChild(this.buildChoiceSection(s));
		const foot = this.buildLinkItem(
			"/$search",
			"All keywords and examples",
			this.buildIcon("ico-arrow-right"),
		);
		foot.classList.add("b-omnisearch-foot");
		foot.firstElementChild?.classList.add("c-with-icon--inline");
		results.appendChild(foot);
		this.helping = true;
		this.endDraw(results);
		return true;
	}

	// suggest reads the token under the caret: the start of a keyword's name
	// offers the keyword; a keyword and its colon, the value it takes.
	// waiting marks a keyword typed with no value yet.
	private async suggest(): Promise<{
		lead: ChoiceSection[];
		waiting?: boolean;
	}> {
		const none = { lead: [] };
		const { text } = this.caretToken();
		const at = text.indexOf(":");
		if (at < 0) {
			if (!text || SearchbarController.isPathLike(text)) return none;
			const word = text.toLowerCase();
			const hits = this.keywords()
				// A keyword typed whole and running alone is already the query.
				.filter((k) => k.name.startsWith(word) && !(k.bare && k.name === word))
				.slice(0, MAX_KEYWORD_ROWS);
			if (hits.length === 0) return none;
			const choices = hits.map((k) => this.keywordChoice(k));
			return { lead: [{ label: "Keywords", choices, highlight: word }] };
		}

		const key = text.slice(0, at).toLowerCase();
		const k = this.keywords().find((s) => s.name === key);
		if (!SearchbarController.isQualifierToken(text) || !k) return none;
		const value = text.slice(at + 1);
		const choices = (await this.valuesFor(k, value)).map((v) => ({
			title: v,
			indexer: k.source === "indexer",
			complete: `${k.name}:${v} `,
		}));
		// What the value is, from the hint: `account:<address>`.
		const expected = /<([^>]+)>/.exec(k.hint)?.[1];
		// A fixed set of values is listed instead: `is:` says realm or package.
		const notice =
			!value && expected && !k.values ? `Type the ${expected}` : undefined;
		if (choices.length === 0 && !notice) return none;
		return {
			lead: [{ label: k.label, choices, notice, highlight: value }],
			waiting: !value,
		};
	}

	// valuesFor lists the values a keyword takes that start with value, from
	// what the bar already holds: no request of its own beyond the path list.
	private async valuesFor(k: Keyword, value: string): Promise<string[]> {
		const v = value.toLowerCase();
		if (k.values) return k.values.filter((x) => x.startsWith(v));
		if (k.complete !== "namespace" && k.complete !== "path") return [];
		await this.ensureLoaded();
		let list = this.namespaces;
		let prefix = v;
		if (k.complete === "path") {
			list = this.realms.concat(this.packages);
			// A path copied with its host: `gno.land/r/demo`.
			prefix = v.replace(GNO_LAND_HOST, "");
			if (!prefix.startsWith("/")) prefix = `/${prefix}`;
		}
		// Shortest first, ties in byte order: a name before the addresses
		// deployed under it.
		return list
			.filter((x) => x.toLowerCase().startsWith(prefix))
			.sort((a, b) => a.length - b.length || (a < b ? -1 : a > b ? 1 : 0))
			.slice(0, RESULTS_PER_GROUP);
	}

	private keywordChoice(k: Keyword): Choice {
		return {
			title: k.hint,
			detail: k.help,
			indexer: k.source === "indexer",
			complete: k.bare ? k.name : `${k.name}:`,
			run: k.bare,
		};
	}

	// caretToken is the selection, or else the word under the caret.
	private caretToken(): Token {
		const input = this.getDOMElement("input") as HTMLInputElement | null;
		const v = input?.value ?? "";
		let start = input?.selectionStart ?? v.length;
		let end = input?.selectionEnd ?? v.length;
		if (start === end) {
			start -= /\S*$/.exec(v.slice(0, start))?.[0].length ?? 0;
			end += /^\S*/.exec(v.slice(end))?.[0].length ?? 0;
		}
		return { start, end, text: v.slice(start, end) };
	}

	// replaceToken puts text in place of the token the list was drawn for,
	// else the caret's, the caret after it.
	private replaceToken(text: string): string {
		const input = this.getDOMElement("input") as HTMLInputElement | null;
		if (!input) return "";
		const { start, end } = this.listToken ?? this.caretToken();
		const rest = input.value.slice(end);
		input.value =
			input.value.slice(0, start) +
			text +
			(text.endsWith(" ") ? rest.trimStart() : rest);
		const caret = start + text.length;
		input.setSelectionRange(caret, caret);
		return input.value;
	}

	// complete edits the input and searches again: a keyword then shows the
	// value it takes, a value the answer.
	private complete(text: string): void {
		this.replaceToken(text);
		this.search();
	}

	// runKeyword runs a keyword that takes no value, as Enter on it would.
	private runKeyword(name: string): void {
		const q = this.replaceToken(name).trim();
		clearTimeout(this.timer);
		window.location.href = this.searchHref(q);
	}

	// searchHref builds on the form's action, the server's `SearchAction`, so
	// the JS and no-JS paths share one source. The server already falls back
	// to `/$search` on a path it cannot parse, where `<path>$search` is a 404.
	private searchHref(q: string): string {
		const input = this.getDOMElement("input") as HTMLInputElement | null;
		const action = input?.form?.getAttribute("action") || "/$search";
		return `${action}&q=${encodeURIComponent(q)}`;
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
		return this.namesSelector(q);
	}

	// namesSelector reports a query that is one selector's search without a
	// `key:`: a bare keyword, or a value it accepts typed alone.
	private namesSelector(q: string): boolean {
		const word = q.trim().toLowerCase();
		return this.selectors.some(
			(s) =>
				(s.bare && s.name === word) ||
				(s.accepts !== undefined && new RegExp(s.accepts).test(q.trim())),
		);
	}

	// unsendable says what the server would refuse in a qualified query, or
	// "": a keyword with no value yet, or, with no selector to run, nothing
	// to search by. Mirrors discover's floor: two characters of text, of an
	// `author:`, or an `in:` package.
	private unsendable(q: string): string {
		const tokens = q.split(/\s+/).filter(Boolean);
		const quals = SearchbarController.qualifiers(tokens);
		const empty = quals.find((f) => f.value === "");
		if (empty) return `Add a value after ${empty.key}:`;
		if (
			this.namesSelector(q) ||
			quals.some((f) => this.selectors.some((s) => s.name === f.key))
		)
			return "";
		const value = (key: string): string =>
			quals.find((f) => f.key === key)?.value ?? "";
		const text = tokens.filter((t) => !SearchbarController.isQualifierToken(t));
		if (
			text.join(" ").length >= MIN_QUERY_LENGTH ||
			value("author").length >= MIN_QUERY_LENGTH ||
			/^\/?[rp]\/./.test(value("in").replace(GNO_LAND_HOST, ""))
		)
			return "";
		return "Add a word, or author: / in:";
	}

	// highlightTerm is what a server-answered row can contain: the free text,
	// else the selector's value. No title contains `func:Render` as typed.
	private highlightTerm(q: string): string {
		const tokens = q.split(/\s+/).filter(Boolean);
		const text = tokens.filter(
			(t) =>
				!SearchbarController.isQualifierToken(t) &&
				!this.selectors.some((s) => s.bare && s.name === t.toLowerCase()),
		);
		if (text.length > 0) return text.join(" ");
		const values = SearchbarController.qualifiers(tokens);
		const named = values.find((v) =>
			this.selectors.some((s) => s.name === v.key),
		);
		return (named ?? values[0])?.value ?? "";
	}

	// qualifiers splits the `key:value` tokens, as ParseQuery does.
	static qualifiers(tokens: string[]): { key: string; value: string }[] {
		return tokens.filter(SearchbarController.isQualifierToken).map((t) => {
			const at = t.indexOf(":");
			return { key: t.slice(0, at).toLowerCase(), value: t.slice(at + 1) };
		});
	}

	// isQualifierToken must stay in step with ParseQuery's isQualifierKey.
	static isQualifierToken(token: string): boolean {
		const at = token.indexOf(":");
		// An object ID led by a hex letter has a bare-word key too.
		if (at < 0 || OID_PATTERN.test(token)) return false;
		if (token.slice(at + 1).startsWith("//")) return false;
		return /^[a-z][a-z0-9_-]*$/.test(token.slice(0, at).toLowerCase());
	}

	// fetchSearch asks the server. A failure returns null and is not retried
	// for the life of the page when the endpoint itself is absent: an older
	// gnoweb, or one behind a proxy that does not route `$search`, must not
	// cost a request per keystroke.
	private async fetchSearch(
		q: string,
		signal?: AbortSignal,
	): Promise<SearchResponse | null> {
		try {
			const res = await fetch(`${this.searchHref(q)}&json`, {
				headers: { Accept: "application/json" },
				signal,
			});
			if (res.status === 404) {
				this.searchAvailable = false;
				return null;
			}
			// A refusal is the server's answer: shown, not replaced by the
			// local path filter reading as "No results".
			if (res.status === 429 || res.status === 400) {
				const body = (await res.json().catch(() => null)) as {
					error?: string;
				} | null;
				return { error: body?.error || `search refused (${res.status})` };
			}
			if (!res.ok) return null;
			const data: unknown = await res.json();
			if (!isSearchResponse(data)) return null;
			this.searchAvailable = true;
			if (data.selectors) {
				this.selectors = data.selectors;
				this.selectorsLoaded = true;
			}
			if (data.filters) this.filters = data.filters;
			if (data.keyword_groups) this.groups = data.keyword_groups;
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
		this.helping = false;
		this.listToken = null;
		return results;
	}

	private endDraw(results: HTMLElement): void {
		// In the combobox pattern DOM focus stays on the input; options are
		// reached with the arrow keys. Without this, Tab walks every result.
		for (const el of this.items) el.tabIndex = -1;
		results.hidden = false;
		this.getDOMElement("input")?.setAttribute("aria-expanded", "true");
	}

	// render is the single dropdown renderer. Both paths — the local path
	// filter and the server's answer — hand it the same SearchGroup[], so
	// ARIA bookkeeping and item collection cannot drift between them.
	private render(
		q: string,
		groups: SearchGroup[],
		pageMatches: PageMatch[],
		opts: {
			unknownFilter?: string;
			error?: string;
			fullPage?: boolean;
			highlight?: string;
			lead?: ChoiceSection[];
			token?: Token;
			// waiting: q was not sent; the lead is the whole answer.
			waiting?: boolean;
			notice?: string;
		} = {},
	): void {
		const results = this.beginDraw();
		if (!results) return;
		this.listToken = opts.token ?? null;

		// Suggestions lead: they change what is typed, so they come first.
		const lead = opts.lead ?? [];
		for (const s of lead) results.appendChild(this.buildChoiceSection(s));

		// A query that could not be answered is not one that found nothing.
		let unanswered = false;
		for (const text of [opts.notice, opts.error]) {
			if (!text) continue;
			results.appendChild(this.buildNotice(text));
			unanswered = true;
		}
		if (opts.unknownFilter) {
			results.appendChild(
				this.buildNotice(
					`"${opts.unknownFilter}:" is not a search qualifier here`,
				),
			);
			unanswered = true;
		}

		for (const group of groups) {
			const section = this.sectionWithLabel(group.label);

			if (group.error) {
				section.appendChild(this.buildNotice(group.error));
				unanswered = true;
			}
			for (const r of group.results ?? []) {
				section.appendChild(this.buildItem(r, opts.highlight ?? q, group));
			}
			if (group.truncated && group.notice) {
				section.appendChild(this.buildNotice(group.notice));
			}
			results.appendChild(section);
		}

		// Counted before the full-results item, which is always there for a
		// qualified query and would otherwise hide that nothing matched.
		const answered =
			groups.some((g) => (g.results?.length ?? 0) > 0) ||
			pageMatches.length > 0 ||
			(opts.waiting === true && lead.length > 0);
		if (!answered && !unanswered) {
			results.appendChild(this.buildNotice("No results"));
		}

		// A qualified query always offers its own results page, the only view
		// that works without JavaScript.
		if (opts.fullPage) {
			const section = this.sectionWithLabel("All results");
			section.appendChild(
				this.buildLinkItem(
					this.searchHref(q),
					`Open the full results page for "${q}"`,
				),
			);
			results.appendChild(section);
		}

		if (pageMatches.length > 0) {
			results.appendChild(this.buildPageSection(q, pageMatches));
		}
		this.endDraw(results);
	}

	// buildItem renders one row. Its group's provenance is said on the row,
	// as on a suggestion: the pill after the text, spoken before it.
	private buildItem(
		r: SearchResult,
		q: string,
		group: SearchGroup,
	): HTMLElement {
		// Same-origin only: `//host` and `/\\host` leave it.
		const linkable = /^\/(?![/\\])/.test(r.href ?? "");
		const item = document.createElement(linkable ? "a" : "div");
		item.className = "b-omnisearch-item";
		item.setAttribute("role", "option");

		if (linkable && r.href) {
			(item as HTMLAnchorElement).href = r.href;
		} else {
			// Announced but unreachable by the arrow keys — a transaction has
			// no page in gnoweb — so it must not inflate the option count.
			item.setAttribute("aria-disabled", "true");
		}
		// The name comes from the row's content, so the detail is read too;
		// an aria-label would replace it. Marks after the text are glyphs:
		// what they mean is spoken here.
		const kind = SearchbarController.typeOf(r.href ?? "");
		if (kind) item.appendChild(this.buildTag(kind));
		const indexer = group.source === "indexer";
		if (indexer) item.appendChild(this.srOnly(`${FROM_INDEXER}: `));

		const text = document.createElement("span");
		text.className = "b-omnisearch-text";
		this.fillHighlighted(text, r.title, q);
		if (r.detail) {
			const detail = document.createElement("small");
			detail.textContent = ` ${r.detail}`;
			text.appendChild(detail);
		}
		item.appendChild(text);

		if (indexer) item.appendChild(this.buildIndexerPill());

		if (linkable) this.items.push(item);
		return item;
	}

	// buildChoice renders a suggestion: what it inserts, set as typed, in a
	// column the panel shares so the descriptions line up beside it.
	private buildChoice(c: Choice, q: string): HTMLElement {
		const item = document.createElement("button");
		item.type = "button";
		item.className = "b-omnisearch-item b-omnisearch-choice";
		item.setAttribute("role", "option");
		item.dataset.complete = c.complete;
		item.addEventListener("click", () =>
			c.run ? this.runKeyword(c.complete) : this.complete(c.complete),
		);
		if (c.indexer) item.appendChild(this.srOnly(`${FROM_INDEXER}: `));

		// The keyword's name is what to read; the value it takes, muted. Its
		// brackets are not read out.
		const syntax = document.createElement("code");
		syntax.className = "b-omnisearch-syntax";
		const at = c.title.indexOf(":");
		this.fillHighlighted(syntax, at < 0 ? c.title : c.title.slice(0, at), q);
		if (at >= 0) {
			const arg = document.createElement("span");
			arg.className = "b-omnisearch-arg";
			for (const part of c.title.slice(at).split(/([<>])/)) {
				if (part === "<") arg.append(this.decor("‹"));
				else if (part === ">") arg.append(this.decor("›"));
				else arg.append(part);
			}
			syntax.appendChild(arg);
		}
		item.appendChild(syntax);
		if (c.detail) {
			const detail = document.createElement("span");
			detail.className = "b-omnisearch-detail";
			detail.textContent = ` ${c.detail}`;
			item.appendChild(detail);
		}

		// One trailing cell: provenance, then "runs at once".
		const marks = document.createElement("span");
		marks.className = "b-omnisearch-marks";
		if (c.indexer) marks.appendChild(this.buildIndexerPill());
		if (c.run) {
			const run = this.decor("↵", "kbd");
			run.className = "b-omnisearch-run";
			run.title = "Runs the search at once";
			marks.appendChild(run);
		}
		if (marks.childElementCount > 0) item.appendChild(marks);
		if (c.run) item.appendChild(this.srOnly(", runs at once"));

		this.items.push(item);
		return item;
	}

	private buildChoiceSection(s: ChoiceSection): HTMLElement {
		const section = this.sectionWithLabel(s.label);
		if (s.notice) section.appendChild(this.buildNotice(s.notice));
		for (const c of s.choices) {
			section.appendChild(this.buildChoice(c, s.highlight ?? ""));
		}
		return section;
	}

	// buildKeywordsHead titles the keyword panel and says how to use it. The
	// keys are what the combobox pattern already announces, so it is hidden
	// from assistive technology, as the group labels are.
	private buildKeywordsHead(): HTMLElement {
		const head = document.createElement("div");
		head.className = "b-omnisearch-head";
		head.setAttribute("aria-hidden", "true");
		const title = document.createElement("span");
		title.className = "b-omnisearch-title";
		title.textContent = "Search keywords";
		const keys = document.createElement("span");
		keys.className = "b-omnisearch-keys";
		[
			["↑↓", " move · "],
			["Tab", " insert · "],
			["↵", " search"],
		].forEach(([key, what]) => {
			const kbd = document.createElement("kbd");
			kbd.className = "b-tag";
			kbd.textContent = key;
			keys.append(kbd, what);
		});
		head.append(title, keys);
		return head;
	}

	// buildLinkItem is a row that opens a page, an option the arrow keys
	// reach.
	private buildLinkItem(
		href: string,
		...content: (string | Node)[]
	): HTMLElement {
		const item = document.createElement("a");
		item.className = "b-omnisearch-item";
		item.setAttribute("role", "option");
		item.href = href;
		const text = document.createElement("span");
		text.className = "b-omnisearch-text";
		text.append(...content);
		item.appendChild(text);
		this.items.push(item);
		return item;
	}

	// buildIcon draws a symbol from the page's icon sprite, as decoration.
	private buildIcon(id: string): SVGSVGElement {
		const svg = document.createElementNS("http://www.w3.org/2000/svg", "svg");
		svg.setAttribute("class", "c-icon");
		svg.setAttribute("aria-hidden", "true");
		const use = document.createElementNS(svg.namespaceURI, "use");
		use.setAttribute("href", `#${id}`);
		svg.appendChild(use);
		return svg;
	}

	// buildIndexerPill is the visible half of a row's provenance; the
	// FROM_INDEXER text before the row's name is the spoken half.
	private buildIndexerPill(): HTMLElement {
		const pill = this.decor("indexer");
		pill.className = "b-tag b-tag--indexer";
		pill.title = "From an indexer, not from the chain";
		return pill;
	}

	// buildTag marks the on-chain kind (r/p/u) a row leads to. The path
	// already says it, so the glyph is decoration.
	private buildTag(type: string): HTMLElement {
		const tag = this.decor(type);
		tag.className = "b-omnisearch-tag";
		tag.dataset.type = type;
		return tag;
	}

	private srOnly(text: string): HTMLElement {
		const el = document.createElement("span");
		el.className = "u-sr-only";
		el.textContent = text;
		return el;
	}

	// decor is a glyph the row's text already says.
	private decor(text: string, tag = "span"): HTMLElement {
		const el = document.createElement(tag);
		el.setAttribute("aria-hidden", "true");
		el.textContent = text;
		return el;
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
				this.namespaces = [
					...new Set(
						this.realms
							.concat(this.packages)
							.map(SearchbarController.firstSegment),
					),
				].filter(Boolean);
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

	// filter ranks the cached lists locally and emits the same shape the
	// server returns, so one renderer serves both paths.
	private filter(q: string): SearchGroup[] {
		const needle = q.toLowerCase();
		const apps = SearchbarController.rank(this.realms, needle);
		const packages = SearchbarController.rank(this.packages, needle);

		const users: string[] = [];
		const seen = new Set<string>();
		// A pasted address names a user even with nothing deployed under it.
		if (SearchbarController.isAddress(needle)) {
			seen.add(needle);
			users.push(`/u/${needle}`);
		}
		for (const p of apps.concat(packages)) {
			const name = SearchbarController.firstSegment(p);
			if (name && !seen.has(name)) {
				seen.add(name);
				users.push(`/u/${name}`);
			}
		}

		const toGroup = (label: string, paths: string[]): SearchGroup => ({
			label,
			source: "chain",
			results: paths
				.slice(0, RESULTS_PER_GROUP)
				.map((path) => ({ title: path, href: path })),
		});

		return [
			toGroup("Apps", apps),
			toGroup("Packages", packages),
			toGroup("Users", users),
		].filter((g) => (g.results?.length ?? 0) > 0);
	}

	// rank returns the paths matching needle, most-relevant first: throwaway
	// deploys last, then a match in the package name outranks one in the
	// namespace/address, and a shallower path edges out a deeper one. The
	// results page ranks with the same rules (omnisearch sortByRelevance).
	static rank(paths: string[], needle: string): string[] {
		const scored: Array<{ path: string; score: number; throwaway: boolean }> =
			[];
		for (const p of paths) {
			const score = SearchbarController.relevance(p, needle);
			if (score > 0)
				scored.push({
					path: p,
					score,
					throwaway: SearchbarController.isThrowaway(p, needle),
				});
		}
		scored.sort(
			(a, b) =>
				Number(a.throwaway) - Number(b.throwaway) ||
				b.score - a.score ||
				a.path.length - b.path.length ||
				(a.path < b.path ? -1 : a.path > b.path ? 1 : 0),
		);
		return scored.map((s) => s.path);
	}

	// isAddress checks a bech32 address's checksum (BIP-173), as the server
	// does before linking /u/<address>: the shape alone lets a typo through.
	static isAddress(s: string): boolean {
		if (!ADDRESS_PATTERN.test(s)) return false;
		const charset = "qpzry9x8gf2tvdw0s3jn54khce6mua7l";
		const sep = s.lastIndexOf("1");
		const hrp = [...s.slice(0, sep)].map((c) => c.charCodeAt(0));
		const values = [
			...hrp.map((c) => c >> 5),
			0,
			...hrp.map((c) => c & 31),
			...[...s.slice(sep + 1)].map((c) => charset.indexOf(c)),
		];
		const gen = [0x3b6a57b2, 0x26508e6d, 0x1ea119fa, 0x3d4233dd, 0x2a1462b3];
		let chk = 1;
		for (const v of values) {
			const top = chk >>> 25;
			chk = ((chk & 0x1ffffff) << 5) ^ v;
			for (let i = 0; i < 5; i++) if ((top >>> i) & 1) chk ^= gen[i];
		}
		return chk === 1;
	}

	// isThrowaway mirrors omnisearch isThrowaway: a raw address namespace, a
	// namespace with "test" in it, or a test package; not when asked for.
	static isThrowaway(path: string, needle: string): boolean {
		const segs = path.toLowerCase().split("/").filter(Boolean);
		const ns = segs[1] ?? "";
		// A pasted address asks for that namespace; a needle inside it does not.
		if (ADDRESS_PATTERN.test(ns) && ns !== needle) return true;
		if (needle.includes("test")) return false;
		if (ns.includes("test")) return true;
		// Not "latest" or "attestation".
		return segs.some(
			(s) => s.startsWith("test") || s.endsWith("_test") || s.endsWith("-test"),
		);
	}

	// relevance scores how well a path matches needle; 0 means no match.
	static relevance(path: string, needle: string): number {
		const lower = path.toLowerCase();
		if (!lower.includes(needle)) return 0;
		const segs = lower.split("/").filter(Boolean);
		// A version segment (`/v0`) is not depth: the name is the one before.
		if (segs.length > 1 && VERSION_SEGMENT.test(segs[segs.length - 1])) {
			segs.pop();
		}
		const name = segs[segs.length - 1] ?? "";
		let score: number;
		if (name === needle) score = 100;
		else if (segs.includes(needle)) score = 90;
		else if (name.startsWith(needle)) score = 80;
		else if (name.includes(needle)) score = 60;
		else if (segs.some((s) => s.startsWith(needle))) score = 40;
		else score = 20;
		return score - segs.length;
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

	// sectionWithLabel opens a group; indexer says in its label that all of
	// it is from the indexer.
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
		const at = q ? text.toLowerCase().indexOf(q.toLowerCase()) : -1;
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
		this.activate(
			(this.activeIndex + delta + this.items.length) % this.items.length,
		);
		this.items[this.activeIndex]?.scrollIntoView({ block: "nearest" });
	}

	// activate highlights row i, the one Enter opens.
	private activate(i: number): void {
		this.deselect();
		this.activeIndex = i;
		const active = this.items[i];
		if (!active) return;
		active.id = `omnisearch-opt-${i}`;
		active.setAttribute("aria-selected", "true");
		this.getDOMElement("input")?.setAttribute(
			"aria-activedescendant",
			active.id,
		);
	}

	// deselect drops the active row's option state; the list stays.
	private deselect(): void {
		const prev = this.items[this.activeIndex];
		prev?.removeAttribute("aria-selected");
		prev?.removeAttribute("id");
	}

	private close(): void {
		clearTimeout(this.timer);
		this.pending = false;
		this.inflight?.abort();
		this.runId++;
		const results = this.getDOMElement("results");
		if (results) {
			results.hidden = true;
			results.textContent = "";
		}
		this.items = [];
		this.activeIndex = -1;
		this.listToken = null;
		const input = this.getDOMElement("input");
		input?.setAttribute("aria-expanded", "false");
		input?.removeAttribute("aria-activedescendant");
	}

	// onKeyShortcut focuses the search bar when "/" is pressed outside any
	// editable element; the focus opens the keyword help, or, when that does
	// not open, the dropdown against what the bar holds, so the action has
	// visible feedback (a bare select() reads as a no-op).
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
		void this.loadSelectors().then(() => {
			if (this.getDOMElement("results")?.hidden) this.search();
		});
	}

	// focusFromForm focuses the input when a press lands elsewhere in the bar,
	// such as the "/" hint or the gap after it, which sit beside the input.
	private focusFromForm(e: MouseEvent): void {
		const input = this.getDOMElement("input");
		if (!input || e.target === input) return;
		e.preventDefault();
		input.focus();
	}

	// selectInput selects the whole input on focus so typing replaces the
	// prefilled path instead of appending to it. Deferred via rAF so a mouse
	// click's caret placement doesn't collapse the selection (a plain
	// synchronous select() is lost on mousedown→focus→mouseup). Only while
	// focused: select() pulls focus back.
	private selectInput(): void {
		const input = this.getDOMElement("input") as HTMLInputElement | null;
		requestAnimationFrame(() => {
			if (document.activeElement === input) input?.select();
		});
		// Learn the qualifier list on first focus, not on page load: a reader
		// who never opens the bar costs no request. A bar not yet typed in
		// opens on the keywords: nothing else says they exist.
		void this.loadSelectors().then(() => {
			const untouched = !input?.value || input.value === input.defaultValue;
			if (document.activeElement === input && untouched) this.showKeywords();
		});
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
		const stripped = input.replace(GNO_LAND_HOST, "");
		try {
			const url = new URL(stripped, window.location.origin);
			if (url.protocol !== "http:" && url.protocol !== "https:") return null;
			return url.href;
		} catch {
			return null;
		}
	}
}
