import { BaseController } from "./controller.js";

interface PlaygroundFile {
	name: string;
	content: string;
}

const GNOMOD_FILE = "gnomod.toml";
const DEFAULT_GNO_CONTENT = "package main\n";

export class PlaygroundController extends BaseController {
	private declare files: PlaygroundFile[];
	private declare activeFile: number;
	private declare codeEl: HTMLTextAreaElement;
	private declare outputEl: HTMLElement;
	private declare tabsEl: HTMLElement;

	protected connect(): void {
		this.files = [];
		this.activeFile = 0;
		this.codeEl = this.getTarget("code") as HTMLTextAreaElement;
		this.outputEl = this.getTarget("output") as HTMLElement;
		this.tabsEl = this.getTarget("tabs") as HTMLElement;
		if (!this.codeEl || !this.outputEl || !this.tabsEl) return;

		this._parseInitialCode();
		this._switchToDefaultFile();
		this._setupKeyboardShortcuts();
		this.renderTabs();
	}

	private _parseInitialCode(): void {
		const initialCode = this.codeEl.value;
		if (initialCode.includes("// --- ") && initialCode.includes(" ---")) {
			const parts = initialCode.split(/^\/\/ --- (.+?) ---$/m);
			for (let i = 1; i < parts.length; i += 2) {
				const name = parts[i].trim();
				const content = (parts[i + 1] || "").trim();
				if (name) this.files.push({ name, content });
			}

			if (this.files.length === 0)
				this.files = [{ name: "main.gno", content: initialCode }];

			this.codeEl.value = this.files[0].content;
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

	private _setupKeyboardShortcuts(): void {
		this.codeEl.addEventListener("keydown", (e: KeyboardEvent) => {
			if (e.ctrlKey && e.key === "Enter") {
				e.preventDefault();

				this.runCode();
				return;
			}

			if (e.key === "Tab" && !e.shiftKey) {
				e.preventDefault();

				const start = this.codeEl.selectionStart;
				const end = this.codeEl.selectionEnd;
				this.codeEl.value = `${this.codeEl.value.substring(0, start)}\t${this.codeEl.value.substring(end)}`;
				this.codeEl.selectionStart = this.codeEl.selectionEnd = start + 1;
			}
		});
	}

	private _setOutput(text: string, isError: boolean = false): void {
		this.outputEl.textContent = text;
		this.outputEl.classList.toggle("u-color-danger", isError);
	}

	private _switchToFile(fileName: string): boolean {
		this.files[this.activeFile].content = this.codeEl.value;
		const idx = this.files.findIndex((f) => f.name === fileName);
		if (idx >= 0) {
			this.activeFile = idx;
			this.codeEl.value = this.files[idx].content;
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
		const name = prompt("File name (e.g. helper.gno):");
		if (name == null) return;

		if (this._switchToFile(name)) return;

		const isGnomod = name === GNOMOD_FILE;
		if (!name.endsWith(".gno") && !isGnomod) return;

		const domain = this.getValue("domain") || "gno.land";
		let content = DEFAULT_GNO_CONTENT;
		if (isGnomod) {
			content = `module = "${domain}/r/yourname/pkg"\ngno = "0.9"`;
		}

		this.files[this.activeFile].content = this.codeEl.value;
		this.files.push({ name, content });
		this.activeFile = this.files.length - 1;
		this.codeEl.value = this.files[this.activeFile].content;
		this.renderTabs();
	}

	public async runCode(): Promise<void> {
		this.files[this.activeFile].content = this.codeEl.value;
		this._setOutput("Running...");

		const code = this.codeEl.value;
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
		this.files[this.activeFile].content = this.codeEl.value;
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
