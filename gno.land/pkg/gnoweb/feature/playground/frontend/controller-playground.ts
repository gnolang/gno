import { CodeEditor, type EditorFile } from "@gnoweb/js/code-editor.js";
import {
	BaseController,
	downloadBlob,
	makeCopyIcon,
	makeIcon,
	postEval,
} from "@gnoweb/js/controller.js";

const GNOMOD_FILE = "gnomod.toml";
const DEFAULT_GNO_CONTENT = "package main\n";

// Max length for shared source code.
// It stays under the 8192-byte default limit of common web servers (nginx, Apache).
const MAX_SHARE_URL_LENGTH = 8_000;

// isValidFileName accepts Gno sources and gnomod.toml, with no path part, so
// a name can never escape the archive Download builds.
function isValidFileName(name: string): boolean {
	if (name.includes("/") || name.includes("\\") || name.includes("..")) {
		return false;
	}
	return name.endsWith(".gno") || name === GNOMOD_FILE;
}

export class PlaygroundController extends BaseController {
	private declare editorEl: HTMLElement;
	private declare outputEl: HTMLElement;
	private declare tabsEl: HTMLElement;
	private declare tabsWrapEl: HTMLElement;
	private declare prevBtnEl: HTMLButtonElement;
	private declare nextBtnEl: HTMLButtonElement;
	// The editor holds every file (text and undo history); the controller
	// keeps no copy of its own.
	private declare editor: CodeEditor;

	protected connect(): void {
		const initialCodeEl = this.getTarget("initial-code") as HTMLTextAreaElement;

		this.editorEl = this.getTarget("editor") as HTMLElement;
		this.outputEl = this.getTarget("output") as HTMLElement;
		this.tabsEl = this.getTarget("tabs") as HTMLElement;
		this.tabsWrapEl = this.getTarget("tabs-wrap") as HTMLElement;
		this.prevBtnEl = this.getTarget("prev-button") as HTMLButtonElement;
		this.nextBtnEl = this.getTarget("next-button") as HTMLButtonElement;
		if (!this.editorEl || !this.outputEl || !this.tabsEl || !initialCodeEl) {
			return;
		}

		this.editorEl.addEventListener("focusin", () =>
			this._scrollActiveTabIntoView(),
		);

		const [first, ...rest] = parseInitialCode(initialCodeEl.value);
		this.editor = new CodeEditor({
			parent: this.editorEl,
			content: first.content,
			fileName: first.name,
			label: "Playground editor",
			onRun: () => this.runCode(),
		});
		for (const f of rest) this.editor.addFile(f.name, f.content);

		const defaultFile = this.getValue("default-file");
		if (defaultFile) this.editor.openFile(defaultFile);

		this._setupTabsScroll();
		this.renderTabs();
		this._resetOutput("// Run code to see output here");
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

	private _resetOutput(
		text: string,
		copyable: boolean = false,
		isError: boolean = false,
	): void {
		this.outputEl.replaceChildren();
		this._setOutput(text, copyable, isError);
	}

	private _setOutput(
		text: string,
		copyable: boolean = false,
		isError: boolean = false,
	): void {
		const row = document.createElement("div");
		row.className = "b-playground-output-item";
		if (isError) {
			row.classList.add("u-color-danger");
		}

		const line = document.createElement("div");
		line.className = "b-playground-output-item-text";
		line.textContent = text;
		row.appendChild(line);

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
	}

	private _setErrorOutput(text: string): void {
		this._resetOutput(text, false, true);
	}

	private _openFile(fileName: string): void {
		this.editor.openFile(fileName);
		this.renderTabs();
	}

	private renderTabs(): void {
		const names = this.editor.fileNames();
		const closable = names.length > 1;
		this.tabsEl.replaceChildren(
			...names.map((name) => {
				const tab = document.createElement("div");
				tab.className = `b-playground-tab${name === this.editor.fileName ? " b-playground-tab--active" : ""}`;

				const open = document.createElement("button");
				open.type = "button";
				open.className = "b-playground-tab-open";
				if (name === this.editor.fileName) {
					open.setAttribute("aria-current", "true");
				}
				open.append(makeIcon("file"), name);
				open.addEventListener("click", () => this._openFile(name));
				tab.appendChild(open);

				if (closable) {
					const close = document.createElement("button");
					close.type = "button";
					close.className = "b-playground-tab-close";
					close.title = `Close ${name}`;
					close.setAttribute("aria-label", `Close ${name}`);
					close.appendChild(makeIcon("cross"));
					close.addEventListener("click", () => this._closeFile(name));
					tab.appendChild(close);
				}
				return tab;
			}),
		);

		this._updateNavButtons();
		this._scrollActiveTabIntoView();
	}

	// _closeFile removes a file from the playground; it never removes the last one.
	private _closeFile(name: string): void {
		const names = this.editor.fileNames();
		if (names.length < 2) return;
		if (
			this.editor.content(name).trim() &&
			!confirm(`Remove ${name} from the playground?`)
		) {
			return;
		}

		if (name === this.editor.fileName) {
			const i = names.indexOf(name);
			this.editor.openFile(names[i + 1] ?? names[i - 1]);
		}
		this.editor.closeFile(name);
		this.renderTabs();
		// The closed tab's button is gone: keep keyboard focus in the tab bar.
		this.tabsEl
			.querySelector<HTMLButtonElement>('[aria-current="true"]')
			?.focus();
	}

	public addFile(): void {
		const name = prompt("File name (e.g. main.gno or gnomod.toml):");
		if (name == null) return;

		if (this.editor.fileNames().includes(name)) {
			this._openFile(name);
			return;
		}

		if (!isValidFileName(name)) {
			this._setErrorOutput(
				`Error: "${name}" is not a valid file name. Use a name ending in .gno, or gnomod.toml.`,
			);
			return;
		}

		let content = DEFAULT_GNO_CONTENT;
		if (name === GNOMOD_FILE) {
			content = `module = "${this.getValue("domain")}/r/yourname/pkg"\ngno = "0.9"`;
		}

		this.editor.addFile(name, content);
		this._openFile(name);
	}

	public async runCode(): Promise<void> {
		this._resetOutput("Running...");

		const code = this.editor.getCode();
		const pkgMatch = code.match(/^package\s+(\w+)/m);
		const pkgName = pkgMatch ? pkgMatch[1] : "main";
		const names = this.editor.fileNames();

		if (code.includes("func Render(")) {
			const { text, isError, unavailable } = await postEval(
				"r/playground_preview",
				'Render("")',
			);
			this._resetOutput(text, false, isError || unavailable);
		} else {
			this._resetOutput(
				`Package: ${pkgName}\nFiles: ${names.join(", ")}\n\nTo run locally:`,
			);
			this._setOutput(` gno run ${names.join(" ")}`, true);
			this._setOutput("\n\nTo test:");
			this._setOutput(" gno test .", true);
		}
	}

	public async shareCode(): Promise<void> {
		const files = this.editor.files();
		const code =
			files.length === 1
				? files[0].content
				: files.map((f) => `// --- ${f.name} ---\n${f.content}`).join("\n\n");

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
		const files = this.editor.files();
		if (files.length === 1) {
			downloadBlob(
				new Blob([files[0].content], { type: "text/plain" }),
				files[0].name,
			);
		} else {
			downloadBlob(
				new Blob([createTar(files)], { type: "application/x-tar" }),
				"playground-gno-source-code.tar",
			);
		}
	}
}

// parseInitialCode splits shared or copied code on its "// --- name ---"
// separators. Code without separators, or whose names are all rejected, is
// a single main.gno.
function parseInitialCode(initialCode: string): EditorFile[] {
	const files: EditorFile[] = [];
	if (initialCode.includes("// --- ") && initialCode.includes(" ---")) {
		const parts = initialCode.split(/^\/\/ --- (.+?) ---$/m);
		for (let i = 1; i < parts.length; i += 2) {
			const name = parts[i].trim();
			if (!isValidFileName(name) || files.some((f) => f.name === name)) {
				console.warn(`PlaygroundController: skipped file "${name}"`);
				continue;
			}
			// Drop the separator's blank lines, keep one final newline.
			const content = (parts[i + 1] || "").replace(/^\n+|\n+$/g, "");
			files.push({ name, content: `${content}\n` });
		}
	}
	return files.length ? files : [{ name: "main.gno", content: initialCode }];
}

function createTar(files: EditorFile[]): Uint8Array<ArrayBuffer> {
	const encoder = new TextEncoder();
	const blocks: Uint8Array[] = [];

	for (const file of files) {
		if (!isValidFileName(file.name)) {
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
