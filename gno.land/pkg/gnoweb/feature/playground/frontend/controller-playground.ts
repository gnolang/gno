import { CodeEditor } from "@gnoweb/js/code-editor.js";
import {
	BaseController,
	downloadBlob,
	makeCopyIcon,
	makeIcon,
	postEval,
} from "@gnoweb/js/controller.js";
import {
	copyFiles,
	createTar,
	GNOMOD_FILE,
	isValidFileName,
	renderText,
	shareLink,
	shareRefusal,
} from "./playground-files.js";

const DEFAULT_GNO_CONTENT = "package main\n";

export class PlaygroundController extends BaseController {
	private declare editorEl: HTMLElement;
	private declare outputEl: HTMLElement;
	private declare tabsEl: HTMLElement;
	private declare tabsWrapEl: HTMLElement;
	private declare prevBtnEl: HTMLButtonElement;
	private declare nextBtnEl: HTMLButtonElement;
	private declare runBtnEl: HTMLElement | null;
	// The editor holds every file (text and undo history); the controller
	// keeps no copy of its own.
	private declare editor: CodeEditor;
	// preview is true while Run can show the copied realm's own Render: the
	// page says it can, and the copy is unmodified.
	private declare preview: boolean;
	// outputSeq counts output resets, so a late Run result never replaces
	// what came after it.
	private declare outputSeq: number;
	private declare runAbort: AbortController | undefined;

	protected connect(): void {
		// The page serves the code once, as text, one <pre> per file: read
		// it before the editor replaces it.
		const served = this.getTargets("file").map((el) => ({
			name: this.getValue("name", el),
			content: el.textContent ?? "",
		}));

		this.editorEl = this.getTarget("editor") as HTMLElement;
		this.outputEl = this.getTarget("output") as HTMLElement;
		this.tabsEl = this.getTarget("tabs") as HTMLElement;
		this.tabsWrapEl = this.getTarget("tabs-wrap") as HTMLElement;
		this.prevBtnEl = this.getTarget("prev-button") as HTMLButtonElement;
		this.nextBtnEl = this.getTarget("next-button") as HTMLButtonElement;
		this.runBtnEl = this.getTarget("run");
		this.outputSeq = 0;
		if (!this.editorEl || !this.outputEl || !this.tabsEl) return;

		// Served hidden: the buttons, the file tabs and the output only work
		// from here on.
		for (const el of this.getTargets("js-only")) el.hidden = false;
		if (this.tabsWrapEl) this.tabsWrapEl.hidden = false;

		this.editorEl.addEventListener("focusin", () =>
			this._scrollActiveTabIntoView(),
		);

		const { files, skipped } = copyFiles(served);
		const [first, ...rest] = files.length
			? files
			: [{ name: "main.gno", content: DEFAULT_GNO_CONTENT }];
		// The editor replaces the code the page was served with as text.
		this.editorEl.replaceChildren();
		this.editor = new CodeEditor({
			parent: this.editorEl,
			content: first.content,
			fileName: first.name,
			label: "Playground editor",
			onRun: () => this.runCode(),
			onChange: () => this._endPreview(),
		});
		for (const f of rest) this.editor.addFile(f.name, f.content);
		this.preview = !!this.runBtnEl && this.hasValue("preview", this.runBtnEl);

		const defaultFile = this.getValue("default-file");
		if (defaultFile) this.editor.openFile(defaultFile);

		this._setupTabsScroll();
		this.renderTabs();
		if (skipped.length) {
			this._setErrorOutput(
				`Left out of this copy, as the playground takes no such file name: ${skipped.join(", ")}. Share and Download leave them out too.`,
			);
		} else {
			this._resetOutput("// Run code to see output here");
		}
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

	// _endPreview makes Run show the commands that run the code locally,
	// once the copy is edited, and the button say so in the page's words.
	private _endPreview(): void {
		if (!this.preview) return;
		this.preview = false;
		// A preview still in flight would show the realm as deployed, which
		// the edit has just made something else: drop it.
		if (this.runAbort) {
			this.runAbort.abort();
			this.runAbort = undefined;
			this._resetOutput("// Run code to see output here");
		}
		const btn = this.runBtnEl;
		if (!btn) return;
		const label = btn.querySelector("span");
		if (label) label.textContent = this.getValue("local-label", btn);
		btn.title = this.getValue("local-title", btn);
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
		this.outputSeq++;
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
		this._endPreview();
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
				`Error: "${name}" is not a valid file name. Use letters, digits, "_", "-" and "." ending in .gno (100 characters at most), or gnomod.toml.`,
			);
			return;
		}

		let content = DEFAULT_GNO_CONTENT;
		if (name === GNOMOD_FILE) {
			content = `module = "${this.getValue("domain")}/r/yourname/pkg"\ngno = "0.9"`;
		}

		this.editor.addFile(name, content);
		this._endPreview();
		this._openFile(name);
	}

	// runCode shows what the copied realm renders on the chain while the
	// copy is unmodified: edited code exists nowhere the node can run it, so
	// it shows the commands that run it locally instead.
	public async runCode(): Promise<void> {
		this.runAbort?.abort();
		this.runAbort = undefined;
		if (!this.preview) {
			this._resetOutput("Save the files in one directory, then run:");
			this._setOutput(" gno run .", true);
			this._setOutput("\n\nTo test:");
			this._setOutput(" gno test .", true);
			return;
		}

		this._resetOutput("Running...");
		const seq = this.outputSeq;
		this.runAbort = new AbortController();
		const { text, isError, unavailable } = await postEval(
			this.getValue("from"),
			'Render("")',
			this.runAbort.signal,
		);
		if (seq !== this.outputSeq) return;
		this.runAbort = undefined;
		const failed = isError || unavailable;
		this._resetOutput(failed ? text : renderText(text), false, failed);
	}

	public async shareCode(): Promise<void> {
		// Output written after this one (a Run) supersedes the share's.
		const seq = this.outputSeq;
		const { url, bytes } = await shareLink(
			window.location.origin,
			this.editor.files(),
		);
		if (seq !== this.outputSeq) return;
		const refusal = shareRefusal(bytes, url.length);
		if (refusal) {
			this._setErrorOutput(refusal);
			return;
		}

		// No Clipboard API outside a secure context (plain http off localhost).
		if (!navigator.clipboard) {
			this._resetOutput(`Share URL:\n${url}`);
			return;
		}
		const done = (text: string) => {
			if (seq === this.outputSeq) this._resetOutput(text);
		};
		navigator.clipboard
			.writeText(url)
			.then(() => done("Share URL copied to clipboard!"))
			.catch(() => done(`Share URL:\n${url}`));
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
