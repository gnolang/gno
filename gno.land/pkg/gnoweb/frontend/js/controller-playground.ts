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
import { BaseController } from "./controller.js";

interface PlaygroundFile {
	name: string;
	content: string;
}

const GNOMOD_FILE = "gnomod.toml";
const DEFAULT_GNO_CONTENT = "package main\n";

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
		if (!this.mountEl || !this.outputEl || !this.tabsEl || !initialCodeEl) {
			return;
		}

		this._parseInitialCode(initialCodeEl.value);
		this._createEditor();
		this.renderTabs();

		this.on("theme:changed", () => {
			this.view.dispatch({
				effects: this.themeCompartment.reconfigure(this._getCodeEditorTheme()),
			});
		});
	}

	private _parseInitialCode(initialCode: string): void {
		if (initialCode.includes("// --- ") && initialCode.includes(" ---")) {
			const parts = initialCode.split(/^\/\/ --- (.+?) ---$/m);
			for (let i = 1; i < parts.length; i += 2) {
				const name = parts[i].trim();
				const content = (parts[i + 1] || "").trim();
				if (name) this.files.push({ name, content });
			}

			if (this.files.length === 0)
				this.files = [{ name: "main.gno", content: initialCode }];
		} else {
			this.files = [{ name: "main.gno", content: initialCode }];
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

	private _setOutput(text: string, isError: boolean = false): void {
		this.outputEl.textContent = text;
		this.outputEl.classList.toggle("u-color-danger", isError);
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

		const addBtn = document.createElement("button");
		addBtn.className = "b-playground-tab-add";
		addBtn.textContent = "+";
		addBtn.title = "Add file";
		addBtn.addEventListener("click", () => this.addFile());
		this.tabsEl.appendChild(addBtn);
	}

	public switchTab(event: Event & { params?: Record<string, unknown> }): void {
		const fileName = event.params?.file as string;
		if (fileName) this._switchToFile(fileName);
	}

	public addFile(): void {
		const name = prompt("File name (e.g. main.gno or gnomod.toml):");
		if (name == null) return;

		if (this._switchToFile(name)) return;

		const isGnomod = name === GNOMOD_FILE;
		if (!name.endsWith(".gno") && !isGnomod) return;

		const domain = this.getValue("domain") || "gno.land";
		let content = DEFAULT_GNO_CONTENT;
		if (isGnomod) {
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
		this._setOutput("Running...");

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
					this._setOutput(`Error: ${result.error}`, true);
				} else {
					this._setOutput(result.result);
				}
			} catch {
				this._setOutput(
					`Note: Server-side execution not available for scratch pad code.\n\nPackage: ${pkgName}\nFiles: ${this.files.map((f) => f.name).join(", ")}\n\nTo deploy and test:\n  gnokey maketx addpkg -pkgpath "${domain}/r/yourname/pkg" ...`,
				);
			}
		} else {
			this._setOutput(
				`Package: ${pkgName}\nFiles: ${this.files.map((f) => f.name).join(", ")}\n\nTo run locally:\n  gno run ${this.files.map((f) => f.name).join(" ")}\n\nTo test:\n  gno test .`,
			);
		}
	}

	public runTests(): void {
		this._setOutput(
			"Testing requires a running gno node.\n\nTo test locally:\n  gno test .",
		);
	}

	public formatCode(): void {
		this._setOutput(
			"Formatting requires server-side gno fmt (coming soon).\n\nTo format locally:\n  gno fmt -w " +
				this.files[this.activeFile].name,
		);
	}

	public shareCode(): void {
		this.files[this.activeFile].content = this._getCode();
		const code =
			this.files.length === 1
				? this.files[0].content
				: this.files
						.map((f) => `// --- ${f.name} ---\n${f.content}`)
						.join("\n\n");

		// Encode as base64; TextEncoder produces UTF-8 bytes, which are mapped to a
		// Latin-1 binary string so btoa() can handle non-ASCII chars, then the
		// resulting base64 is percent-encoded to be safe in a URL query parameter.
		const bytes = new TextEncoder().encode(code);
		const binary = Array.from(bytes, (b) => String.fromCharCode(b)).join("");
		const url = `${window.location.origin}/_/play?code=${encodeURIComponent(btoa(binary))}`;
		navigator.clipboard
			.writeText(url)
			.then(() => {
				this._setOutput("Share URL copied to clipboard!");
			})
			.catch(() => {
				this._setOutput(`Share URL:\n${url}`);
			});
	}

	public clearOutput(): void {
		this._setOutput("// Run code to see output here");
	}
}
