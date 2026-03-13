import type {
	AminoFuncValue,
	QobjectResponse,
	QpkgResponse,
	QtypeResponse,
	StateNode,
} from "@gnojs/amino";
import {
	decodeFuncObject,
	decodeObject,
	decodePkg,
	structFieldNames,
} from "@gnojs/amino";
import { BaseController } from "./controller.js";

const ARROW_RIGHT = "\u25B6";
const ARROW_DOWN = "\u25BC";

export class StateExplorerController extends BaseController {
	private declare pkgPath: string;
	private declare typeCache: Map<string, string[]>;
	private declare sourceCache: Map<string, string>;

	protected connect(): void {
		this.pkgPath = this.getValue("pkg-path");
		this.typeCache = new Map();
		this.sourceCache = new Map();

		// Show realm path when navigated via OID (gnoweb uses $state&oid=... in path)
		if (window.location.pathname.includes("oid=")) {
			this._showPathInfo();
		}

		const dataEl = this.getTarget("initial-data");
		if (dataEl?.textContent) {
			try {
				const raw: QpkgResponse = JSON.parse(dataEl.textContent);
				const nodes = decodePkg(raw);
				const tree = this.getTarget("tree");
				if (tree) {
					this._renderNodes(nodes, tree, 0);
					this._updateCount(nodes.length);
				}
			} catch (err) {
				console.error("Failed to parse initial state data:", err);
			}
		}
	}

	private _updateCount(n: number): void {
		const countEl = this.getTarget("count");
		if (countEl) {
			const kind = this.pkgPath.startsWith("/r/") ? "Realm" : "Package";
			countEl.textContent = `${kind} top-level declarations (${n})`;
		}
	}

	private _showPathInfo(): void {
		const el = this.getTarget("path-info");
		if (!el) return;
		const link = document.createElement("a");
		link.href = this.pkgPath;
		link.textContent = this.pkgPath;
		link.className = "b-state-explorer__path-link";
		el.textContent = "Realm: ";
		el.appendChild(link);
	}

	private _renderNodes(
		nodes: StateNode[],
		container: HTMLElement,
		depth: number,
	): void {
		const fragment = document.createDocumentFragment();
		for (const node of nodes) {
			fragment.appendChild(this._createRow(node, depth));
		}
		container.appendChild(fragment);
	}

	private _createRow(node: StateNode, depth: number): HTMLElement {
		const row = document.createElement("div");
		row.className = "b-state-row";

		const line = document.createElement("div");
		line.className = "b-state-row__line";
		line.style.paddingLeft = `${depth * 1.25 + 0.25}rem`;

		// Toggle arrow
		const toggle = document.createElement("span");
		toggle.className = "b-state-toggle";
		if (node.expandable || (node.children && node.children.length > 0)) {
			toggle.textContent = ARROW_RIGHT;
			toggle.addEventListener("click", () =>
				this._toggle(toggle, row, node, depth),
			);
		}
		line.appendChild(toggle);

		// Name
		const nameEl = document.createElement("span");
		nameEl.className = "b-state-name";
		nameEl.textContent = node.name;
		line.appendChild(nameEl);

		// Separator
		line.appendChild(this._sep(":"));

		// Type
		const typeEl = document.createElement("span");
		typeEl.className = `b-state-type b-state-kind--${node.kind}`;
		typeEl.textContent = node.type;
		line.appendChild(typeEl);

		// Length
		if (node.length !== undefined && node.length > 0) {
			const lenEl = document.createElement("span");
			lenEl.className = "b-state-meta";
			lenEl.textContent = `(len=${node.length})`;
			line.appendChild(lenEl);
		}

		// Value
		if (node.value !== undefined && node.value !== "") {
			line.appendChild(this._sep("="));
			const valEl = document.createElement("span");
			valEl.className = `b-state-val b-state-val--${node.kind}`;
			valEl.textContent = node.value;
			line.appendChild(valEl);
		}

		// Source location (for functions) — clickable link to source view
		if (node.source) {
			const srcLink = document.createElement("a");
			srcLink.className = "b-state-meta b-state-srclink";
			srcLink.textContent = `${node.source.file}:${node.source.startLine}`;
			srcLink.href = `${this.pkgPath}$source&file=${encodeURIComponent(node.source.file)}#L${node.source.startLine}`;
			srcLink.title = "View source";
			line.appendChild(srcLink);
		}

		// ObjectID (subtle, on hover)
		if (node.objectId) {
			const oid = node.objectId;
			const oidEl = document.createElement("span");
			oidEl.className = "b-state-oid";
			oidEl.textContent = oid;
			oidEl.title = "Object ID \u2014 click to copy";
			oidEl.addEventListener("click", (e) => {
				e.stopPropagation();
				navigator.clipboard.writeText(oid);
				oidEl.textContent = "copied!";
				setTimeout(() => {
					oidEl.textContent = oid;
				}, 1000);
			});
			line.appendChild(oidEl);
		}

		row.appendChild(line);

		// Children container
		const kids = document.createElement("div");
		kids.className = "b-state-kids";
		if (node.children && node.children.length > 0) {
			this._renderNodes(node.children, kids, depth + 1);
		} else {
			kids.hidden = true;
		}
		row.appendChild(kids);

		return row;
	}

	private _sep(char: string): HTMLElement {
		const s = document.createElement("span");
		s.className = "b-state-sep";
		s.textContent = char;
		return s;
	}

	private async _toggle(
		toggle: HTMLElement,
		row: HTMLElement,
		node: StateNode,
		depth: number,
	): Promise<void> {
		const kids = row.querySelector(".b-state-kids") as HTMLElement;
		if (!kids) return;

		const isHidden = kids.hidden;

		if (isHidden) {
			// Expand — lazy-fetch if needed
			if (kids.children.length === 0) {
				// Closures with inline children: render source + captures
				if (
					node.kind === "closure" &&
					node.children &&
					node.children.length > 0
				) {
					if (node.source) {
						await this._renderSourceBlock(
							node.source.file,
							node.source.startLine,
							node.source.endLine,
							kids,
							depth,
						);
					}
					// Render capture label + children
					const label = document.createElement("div");
					label.className = "b-state-captures-label";
					label.style.paddingLeft = `${(depth + 1) * 1.25 + 0.25}rem`;
					label.textContent = "Captured variables:";
					kids.appendChild(label);
					this._renderNodes(node.children, kids, depth + 1);
				} else if (node.objectId) {
					toggle.classList.add("b-state-toggle--loading");
					try {
						const url = `${this.pkgPath}$state&oid=${encodeURIComponent(node.objectId)}&json`;
						const resp = await fetch(url);
						if (!resp.ok) throw new Error(`HTTP ${resp.status}`);
						const raw: QobjectResponse = await resp.json();

						if (node.kind === "func" || node.kind === "closure") {
							await this._expandFunc(
								raw.value as AminoFuncValue,
								kids,
								depth,
							);
						} else {
							let childNodes = decodeObject(raw);

							// Resolve struct field names via qtype_json if parent has a typeId
							if (node.typeId && childNodes.length > 0) {
								childNodes = await this._resolveFieldNames(
									node.typeId,
									childNodes,
								);
							}

							this._renderNodes(childNodes, kids, depth + 1);
						}
					} catch (err) {
						console.error("State fetch error:", err);
						const errEl = document.createElement("span");
						errEl.className = "b-state-err";
						errEl.textContent = "Failed to load";
						kids.appendChild(errEl);
					}
					toggle.classList.remove("b-state-toggle--loading");
				}
			}
			kids.hidden = false;
			toggle.textContent = ARROW_DOWN;
		} else {
			kids.hidden = true;
			toggle.textContent = ARROW_RIGHT;
		}
	}

	// Render a syntax-highlighted source block inline.
	private async _renderSourceBlock(
		file: string,
		startLine: number,
		endLine: number,
		container: HTMLElement,
		depth: number,
	): Promise<void> {
		try {
			const html = await this._fetchSourceHTML(file, startLine, endLine);

			const wrapper = document.createElement("div");
			wrapper.className = "b-state-source";
			wrapper.style.paddingLeft = `${(depth + 1) * 1.25 + 0.25}rem`;

			// Source link in top-right corner
			const link = document.createElement("a");
			link.className = "b-state-source__link";
			link.textContent = `${file}:${startLine}`;
			link.href = `${this.pkgPath}$source&file=${encodeURIComponent(file)}#L${startLine}`;
			link.title = "View source";
			wrapper.appendChild(link);

			const code = document.createElement("div");
			code.innerHTML = html;
			wrapper.appendChild(code);

			container.appendChild(wrapper);
		} catch (err) {
			console.error("Source fetch error:", err);
			const errEl = document.createElement("span");
			errEl.className = "b-state-err";
			errEl.textContent = `Failed to load source: ${err instanceof Error ? err.message : String(err)}`;
			container.appendChild(errEl);
		}
	}

	// Fetch syntax-highlighted source and display it inline, plus captures for closures.
	private async _expandFunc(
		fv: AminoFuncValue,
		container: HTMLElement,
		depth: number,
	): Promise<void> {
		const loc = fv.Source?.Location;
		if (!loc?.File || !loc?.Span) {
			const info = decodeFuncObject(fv);
			if (info.source) {
				const el = document.createElement("a");
				el.className = "b-state-meta b-state-srclink";
				el.textContent = `${info.source.file}:${info.source.startLine}`;
				el.href = `${this.pkgPath}$source&file=${encodeURIComponent(info.source.file)}#L${info.source.startLine}`;
				container.appendChild(el);
			}
			// Still show captures even without source location
			if (fv.Captures && fv.Captures.length > 0) {
				this._renderCaptures(fv, container, depth);
			}
			return;
		}

		const file = loc.File;
		const startLine = parseInt(loc.Span.Pos.Line) || 1;
		const endLine = parseInt(loc.Span.End.Line) || startLine;

		await this._renderSourceBlock(file, startLine, endLine, container, depth);

		// Show captures for closures
		if (fv.Captures && fv.Captures.length > 0) {
			this._renderCaptures(fv, container, depth);
		}
	}

	// Render closure captures as child nodes.
	private _renderCaptures(
		fv: AminoFuncValue,
		container: HTMLElement,
		depth: number,
	): void {
		const label = document.createElement("div");
		label.className = "b-state-captures-label";
		label.style.paddingLeft = `${(depth + 1) * 1.25 + 0.25}rem`;
		label.textContent = "Captured variables:";
		container.appendChild(label);

		const decoded = decodeFuncObject(fv);
		if (decoded.children && decoded.children.length > 0) {
			this._renderNodes(decoded.children, container, depth + 1);
		}
	}

	// Fetch syntax-highlighted HTML for a line range of a source file.
	private async _fetchSourceHTML(
		fileName: string,
		start: number,
		end: number,
	): Promise<string> {
		const cacheKey = `${fileName}:${start}-${end}`;
		const cached = this.sourceCache.get(cacheKey);
		if (cached !== undefined) return cached;

		const url = `${this.pkgPath}$state&file=${encodeURIComponent(fileName)}&start=${start}&end=${end}&json`;
		const resp = await fetch(url);
		if (!resp.ok) throw new Error(`HTTP ${resp.status}`);
		const html = await resp.text();
		this.sourceCache.set(cacheKey, html);
		return html;
	}

	// Resolve struct field names from type info.
	private async _resolveFieldNames(
		typeId: string,
		children: StateNode[],
	): Promise<StateNode[]> {
		// Skip stdlib types — field names for stdlib structs (e.g. time.Time)
		// are rarely useful and the extra round-trip isn't worth it.
		if (!typeId.includes("/")) {
			return children;
		}

		let names = this.typeCache.get(typeId);
		if (!names) {
			try {
				const url = `${this.pkgPath}$state&tid=${encodeURIComponent(typeId)}&json`;
				const resp = await fetch(url);
				if (resp.ok) {
					const raw: QtypeResponse = await resp.json();
					const resolved = structFieldNames(raw.type);
					if (resolved) {
						names = resolved;
						this.typeCache.set(typeId, names);
					}
				}
			} catch {
				// Type resolution failed — keep index-based names
			}
		}
		if (names) {
			for (let i = 0; i < children.length && i < names.length; i++) {
				children[i].name = names[i];
			}
		}
		return children;
	}
}
