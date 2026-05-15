import {
	defaultKeymap,
	history,
	historyKeymap,
	indentWithTab,
} from "@codemirror/commands";
import {
	bracketMatching,
	defaultHighlightStyle,
	indentOnInput,
	indentUnit,
	StreamLanguage,
	syntaxHighlighting,
} from "@codemirror/language";
import { go } from "@codemirror/legacy-modes/mode/go";
import { toml } from "@codemirror/legacy-modes/mode/toml";
import { Compartment, EditorState } from "@codemirror/state";
import { oneDark } from "@codemirror/theme-one-dark";
import {
	drawSelection,
	EditorView,
	highlightActiveLine,
	highlightActiveLineGutter,
	keymap,
	lineNumbers,
} from "@codemirror/view";
import { BaseController, makeCopyIcon } from "./controller.js";

interface PlaygroundFile {
	name: string;
	content: string;
}

const GNOMOD_FILE = "gnomod.toml";
const DEFAULT_GNO_CONTENT = "package main\n";

// Max length for shared source code.
// It stays under the 8192-byte default limit of common web servers (nginx, Apache).
const MAX_SHARE_URL_LENGTH = 8_000;

const goLang = StreamLanguage.define(go);
const tomlLang = StreamLanguage.define(toml);

function languageFromFilename(name: string): StreamLanguage<unknown> {
	return name.endsWith(".toml") ? tomlLang : goLang;
}

export class PlaygroundController extends BaseController {
	private declare files: PlaygroundFile[];
	private declare activeFile: number;
	private declare mountEl: HTMLElement;
	private declare outputEl: HTMLElement;
	private declare tabsEl: HTMLElement;
	private declare tabsWrapEl: HTMLElement;
	private declare prevBtnEl: HTMLButtonElement;
	private declare nextBtnEl: HTMLButtonElement;
	private declare view: EditorView;
	private declare langCompartment: Compartment;
	private declare themeCompartment: Compartment;

	protected connect(): void {
		const initialCodeEl = this.getTarget("initial-code") as HTMLTextAreaElement;

		this.files = [];
		this.activeFile = 0;
		this.mountEl = this.getTarget("code") as HTMLElement;
		this.outputEl = this.getTarget("output") as HTMLElement;
		this.tabsEl = this.getTarget("tabs") as HTMLElement;
		this.tabsWrapEl = this.getTarget("tabs-wrap") as HTMLElement;
		this.prevBtnEl = this.getTarget("prev-button") as HTMLButtonElement;
		this.nextBtnEl = this.getTarget("next-button") as HTMLButtonElement;
		if (!this.mountEl || !this.outputEl || !this.tabsEl || !initialCodeEl)
			return;

		this.mountEl.addEventListener("focusin", () =>
			this._scrollActiveTabIntoView(),
		);

		this._parseInitialCode(initialCodeEl.value);
		this._createEditor();
		this._switchToDefaultFile();
		this._setupTabsScroll();
		this.renderTabs();
		this.clearOutput();

		this.on("theme:changed", () => {
			this.view.dispatch({
				effects: this.themeCompartment.reconfigure(this._getCodeEditorTheme()),
			});
		});
	}

	private _setupTabsScroll(): void {
		if (!this.tabsWrapEl || !this.prevBtnEl || !this.nextBtnEl) return;

		this.tabsEl.addEventListener("scroll", () => this._updateNavButtons(), {
			passive: true,
		});

		const observer = new ResizeObserver(() => this._updateNavButtons());
		observer.observe(this.tabsWrapEl);
		observer.observe(this.tabsEl);
	}

	private _updateNavButtons(): void {
		if (!this.tabsWrapEl || !this.prevBtnEl || !this.nextBtnEl) return;

		const overflows = this.tabsEl.scrollWidth > this.tabsWrapEl.clientWidth + 1;
		this.prevBtnEl.hidden = !overflows;
		this.nextBtnEl.hidden = !overflows;
		if (!overflows) return;

		const { scrollLeft, scrollWidth, clientWidth } = this.tabsEl;
		this.prevBtnEl.disabled = scrollLeft <= 0;
		this.nextBtnEl.disabled = scrollLeft + clientWidth >= scrollWidth - 1;
	}

	private _scrollByPage(direction: 1 | -1): void {
		// Calculate the scroll distance, keeping 70% of the tab bar visible,
		// keeping ~30% overlap, so user keeps visual context across clicks.
		// The 80px floor guarantees a meaningful jump when the tab bar is
		// very narrow, where 70% could be tiny.
		const amount = Math.max(this.tabsEl.clientWidth * 0.7, 80);
		this.tabsEl.scrollBy({ left: direction * amount, behavior: "smooth" });
	}

	private _scrollActiveTabIntoView(): void {
		const active = this.tabsEl.querySelector(
			".b-playground-tab--active",
		) as HTMLElement | null;
		if (!active) return;

		active.scrollIntoView({ inline: "nearest", block: "nearest" });
	}

	public scrollTabsPrev(): void {
		this._scrollByPage(-1);
	}

	public scrollTabsNext(): void {
		this._scrollByPage(1);
	}

	private _isValidFileName(name: string): boolean {
		return name.endsWith(".gno") || name === GNOMOD_FILE;
	}

	private _parseInitialCode(initialCode: string): void {
		if (initialCode.includes("// --- ") && initialCode.includes(" ---")) {
			const parts = initialCode.split(/^\/\/ --- (.+?) ---$/m);
			for (let i = 1; i < parts.length; i += 2) {
				const name = parts[i].trim();
				const content = (parts[i + 1] || "").trim();

				// Make sure file name has no path prefix
				if (name.includes("/") || name.includes("\\") || name.includes("..")) {
					console.warn(
						`PlaygroundControler: rejected unsafe file name: "${name}"`,
					);
					continue;
				}

				if (name && this._isValidFileName(name)) {
					this.files.push({ name, content });
				}
			}

			if (this.files.length === 0) {
				this.files = [{ name: "main.gno", content: initialCode }];
			}
		} else {
			this.files = [{ name: "main.gno", content: initialCode }];
		}
	}

	private _switchToDefaultFile(): void {
		const defaultFile = this.getValue("default-file");
		if (defaultFile) {
			this._switchToFile(defaultFile);
		}
	}

	private _createEditor(): void {
		this.langCompartment = new Compartment();
		this.themeCompartment = new Compartment();

		const runOnEnter = keymap.of([
			{
				key: "Mod-Enter",
				preventDefault: true,
				run: () => {
					this.runCode();
					return true;
				},
			},
		]);

		this.view = new EditorView({
			parent: this.mountEl,
			state: EditorState.create({
				doc: this.files[0].content,
				extensions: [
					lineNumbers(),
					highlightActiveLine(),
					highlightActiveLineGutter(),
					drawSelection(),
					history(),
					indentOnInput(),
					indentUnit.of("\t"),
					bracketMatching(),
					this.langCompartment.of(languageFromFilename(this.files[0].name)),
					this.themeCompartment.of(this._getCodeEditorTheme()),
					runOnEnter,
					keymap.of([indentWithTab, ...historyKeymap, ...defaultKeymap]),
				],
			}),
		});
	}

	private _isDarkMode(): boolean {
		return document.documentElement.getAttribute("data-theme") === "dark";
	}

	private _getCodeEditorTheme() {
		return this._isDarkMode()
			? oneDark
			: syntaxHighlighting(defaultHighlightStyle, { fallback: true });
	}

	private _getCode(): string {
		return this.view.state.doc.toString();
	}

	private _setCode(text: string): void {
		this.view.dispatch({
			changes: { from: 0, to: this.view.state.doc.length, insert: text },
		});
	}

	private _setLanguage(name: string): void {
		this.view.dispatch({
			effects: this.langCompartment.reconfigure(languageFromFilename(name)),
		});
	}

	private _resetOutput(
		text: string,
		copyable: boolean = false,
		isError: boolean = false,
	): void {
		while (this.outputEl.firstChild) {
			this.outputEl.removeChild(this.outputEl.firstChild);
		}
		this._setOutput(text, copyable, isError);
	}

	private _setOutput(
		text: string,
		copyable: boolean = false,
		isError: boolean = false,
	): void {
		const row = document.createElement("div");
		row.className = "b-playground-output-item";
		if (isError) row.classList.add("u-color-danger");

		const pre = document.createElement("pre");
		pre.className = "b-playground-output-item-text";
		pre.textContent = text;
		row.appendChild(pre);

		if (copyable) {
			const btn = document.createElement("button");
			btn.className = "b-playground-output-copy-btn";
			btn.title = "Copy to clipboard";
			btn.setAttribute("aria-label", "Copy to clipboard");
			btn.setAttribute("data-controller", "copy");
			btn.setAttribute("data-action", "click->copy#copy");
			btn.setAttribute("data-copy-text-value", text);
			btn.appendChild(makeCopyIcon());
			row.appendChild(btn);
		}

		this.outputEl.appendChild(row);
		this.outputEl.scrollIntoView({ behavior: "smooth", block: "nearest" });
	}

	private _setErrorOutput(text: string): void {
		this._resetOutput(text, false, true);
	}

	private _switchToFile(fileName: string): boolean {
		this.files[this.activeFile].content = this._getCode();
		const idx = this.files.findIndex((f) => f.name === fileName);
		if (idx >= 0) {
			this.activeFile = idx;
			this._setCode(this.files[idx].content);
			this._setLanguage(this.files[idx].name);
			this.renderTabs();
		}
		return idx >= 0;
	}

	private renderTabs(): void {
		while (this.tabsEl.firstChild)
			this.tabsEl.removeChild(this.tabsEl.firstChild);

		this.files.forEach((f, i) => {
			const btn = document.createElement("button");
			btn.className = `b-playground-tab${i === this.activeFile ? " b-playground-tab--active" : ""}`;
			btn.textContent = f.name;
			btn.addEventListener("click", () => this._switchToFile(f.name));
			this.tabsEl.appendChild(btn);
		});

		this._updateNavButtons();
		this._scrollActiveTabIntoView();
	}

	public switchTab(event: Event & { params?: Record<string, unknown> }): void {
		const fileName = event.params?.file as string;
		if (fileName) this._switchToFile(fileName);
	}

	public addFile(): void {
		const name = prompt("File name (e.g. main.gno or gnomod.toml):");
		if (name == null) return;

		// If a file with the same name exists, switch to it
		if (this._switchToFile(name)) return;

		if (!this._isValidFileName(name)) {
			console.error(
				`PlaygroundController: invalid name, file not added: ${name}`,
			);
			return;
		}

		let content = DEFAULT_GNO_CONTENT;
		if (name === GNOMOD_FILE) {
			const domain = this.getValue("domain") || "gno.land";
			content = `module = "${domain}/r/yourname/pkg"\ngno = "0.9"`;
		}

		this.files[this.activeFile].content = this._getCode();
		this.files.push({ name, content });
		this.activeFile = this.files.length - 1;
		this._setCode(this.files[this.activeFile].content);
		this._setLanguage(this.files[this.activeFile].name);
		this.renderTabs();
	}

	public async runCode(): Promise<void> {
		this.files[this.activeFile].content = this._getCode();
		this._resetOutput("Running...");

		const code = this._getCode();
		const pkgMatch = code.match(/^package\s+(\w+)/m);
		const pkgName = pkgMatch ? pkgMatch[1] : "main";
		const domain = this.getValue("domain") || "gno.land";

		if (code.includes("func Render(")) {
			try {
				const resp = await fetch("/_/api/eval", {
					method: "POST",
					headers: { "Content-Type": "application/json" },
					body: JSON.stringify({
						pkg_path: `${domain}/r/playground_preview`,
						expression: 'Render("")',
					}),
				});
				const result = await resp.json();
				if (result.error) {
					this._setErrorOutput(`Error: ${result.error}`);
				} else {
					this._resetOutput(result.result);
				}
			} catch {
				this._resetOutput(
					`Note: Server-side execution not available for scratch pad code.\n\nPackage: ${pkgName}\nFiles: ${this.files.map((f) => f.name).join(", ")}\n\nTo deploy and test:\n`,
				);
				this._setOutput(
					` gnokey maketx addpkg -pkgpath "${domain}/r/yourname/pkg" ...`,
					true,
				);
			}
		} else {
			this._resetOutput(
				`Package: ${pkgName}\nFiles: ${this.files.map((f) => f.name).join(", ")}\n\nTo run locally:`,
			);
			this._setOutput(
				` gno run ${this.files.map((f) => f.name).join(" ")}`,
				true,
			);
			this._setOutput("\n\nTo test:");
			this._setOutput(" gno test .", true);
		}
	}

	public runTests(): void {
		this._resetOutput(
			"Testing requires a running gno node.\n\nTo test locally:",
		);
		this._setOutput(" gno test .", true);
	}

	public formatCode(): void {
		this._resetOutput(
			"Formatting requires server-side gno fmt (coming soon).\n\nTo format locally:",
		);
		this._setOutput(` gno fmt -w ${this.files[this.activeFile].name}`, true);
	}

	public async shareCode(): Promise<void> {
		this.files[this.activeFile].content = this._getCode();

		const code =
			this.files.length === 1
				? this.files[0].content
				: this.files
						.map((f) => `// --- ${f.name} ---\n${f.content}`)
						.join("\n\n");

		// Compress code before sharing it
		const bytes = new TextEncoder().encode(code);
		const cs = new CompressionStream("deflate-raw");
		const writer = cs.writable.getWriter();
		writer.write(bytes);
		writer.close();

		// Use Response to drain the stream into an ArrayBuffer for compatibility with
		// browsers older than ~2 years. ReadableStream.bytes() would a simpler alternative
		// but it's only available for (Chrome 124+, Firefox 128+, Safari 18+).
		const compressed = await new Response(cs.readable).arrayBuffer();
		const binary = Array.from(new Uint8Array(compressed), (b) =>
			String.fromCharCode(b),
		).join("");

		// Share compressed code
		const url = `${window.location.origin}/_/play?code=${encodeURIComponent(btoa(binary))}&z`;
		if (url.length > MAX_SHARE_URL_LENGTH) {
			this._setErrorOutput(
				`Error: code is too large to share via URL.\n\nTry reducing the code or splitting into a deployed package.`,
			);
			return;
		}

		navigator.clipboard
			.writeText(url)
			.then(() => this._resetOutput("Share URL copied to clipboard!"))
			.catch(() => this._resetOutput(`Share URL:\n${url}`));
	}

	public downloadFiles(): void {
		// Make sure current file content is the latest when downloading
		this.files[this.activeFile].content = this._getCode();

		if (this.files.length === 1) {
			this._triggerDownload(
				new Blob([this.files[0].content], { type: "text/plain" }),
				this.files[0].name,
			);
		} else {
			this._triggerDownload(
				new Blob([createTar(this.files)], { type: "application/x-tar" }),
				"playground-gno-source-code.tar",
			);
		}
	}

	private _triggerDownload(blob: Blob, filename: string): void {
		const url = URL.createObjectURL(blob);
		const a = document.createElement("a");
		a.href = url;
		a.download = filename;
		a.click();
		URL.revokeObjectURL(url);
	}

	public clearOutput(): void {
		this._resetOutput("// Run code to see output here");
	}
}

function createTar(
	files: { name: string; content: string }[],
): Uint8Array<ArrayBuffer> {
	const encoder = new TextEncoder();
	const blocks: Uint8Array[] = [];

	for (const file of files) {
		if (
			file.name.includes("/") ||
			file.name.includes("\\") ||
			file.name.includes("..")
		) {
			console.error(
				`PlaygroundController: skipped file with unsafe name in tar: "${file.name}"`,
			);
			continue;
		}

		const data = encoder.encode(file.content);

		// TAR's numeric fields are ASCII octal strings
		const timestamp = Math.floor(Date.now() / 1000).toString(8);

		// Unix Standard TAR header: 512-byte fixed-field block preceding each file's data.
		// Spec: https://www.gnu.org/software/tar/manual/html_node/Standard.html
		// Fields: https://en.wikipedia.org/wiki/Tar_(computing)#UStar_format
		const header = new Uint8Array(512);

		header.set(encoder.encode(file.name).slice(0, 100), 0);
		header.set(encoder.encode("0000644\0"), 100); // file mode
		header.set(encoder.encode("0000000\0"), 108); // uid
		header.set(encoder.encode("0000000\0"), 116); // gid
		header.set(
			encoder.encode(`${data.length.toString(8).padStart(11, "0")}\0`),
			124,
		); // file size
		header.set(encoder.encode(`${timestamp.padStart(11, "0")}\0`), 136); // mtime
		header.fill(0x20, 148, 156); // checksum placeholder (spaces)
		header[156] = 0x30; // typeflag '0' = regular file
		header.set(encoder.encode("ustar\0"), 257); // magic
		header.set(encoder.encode("00"), 263); // version

		// Sum all 512 header bytes (the 8 checksum bytes count as spaces, already
		// set above), then write the result as 6 octal digits + NUL + space
		// back into the checksum field at offset 148.
		let sum = 0;
		for (let i = 0; i < 512; i++) sum += header[i];
		header.set(encoder.encode(`${sum.toString(8).padStart(6, "0")}\0 `), 148);

		// Round up to the next 512-byte boundary because TAR requires
		// file data to be padded to a multiple of 512 bytes.
		// Extra space is padded with zeros.
		const padded = new Uint8Array(Math.ceil(data.length / 512) * 512);
		padded.set(data);

		blocks.push(header);
		blocks.push(padded);
	}

	// EOF, marked by two zero blocks
	blocks.push(new Uint8Array(1024));

	// Concatenate all blocks into a single flat byte array. Each block is
	// copied into its position in the output buffer using a running offset.
	const total = blocks.reduce((n, b) => n + b.length, 0);
	const out = new Uint8Array(total);
	let off = 0;
	for (const b of blocks) {
		out.set(b, off);
		off += b.length;
	}
	return out;
}
