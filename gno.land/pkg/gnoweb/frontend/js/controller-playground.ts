import { BaseController } from "./controller.js";

interface PlaygroundFile {
	name: string;
	content: string;
}

export class PlaygroundController extends BaseController {
	private files: PlaygroundFile[] = [];
	private activeFile = 0;
	private _codeEl!: HTMLTextAreaElement;
	private _outputEl!: HTMLElement;
	private _tabsEl!: HTMLElement;

	protected connect(): void {
		this._codeEl = this.getTarget("code") as HTMLTextAreaElement;
		this._outputEl = this.getTarget("output") as HTMLElement;
		this._tabsEl = this.getTarget("tabs") as HTMLElement;

		// Initialize files from the editor content
		const initialCode = this._codeEl.value;
		if (initialCode.includes("// --- ") && initialCode.includes(" ---")) {
			this._parseForkedFiles(initialCode);
		} else {
			this.files = [{ name: "main.gno", content: initialCode }];
		}

		this._renderTabs();
		this._setupKeyboard();
		this._bindButtons();
	}

	// Bind toolbar buttons directly (BaseController.setupActions can be unreliable)
	private _bindButtons(): void {
		this.element.querySelectorAll("[data-action]").forEach((el) => {
			const attr = el.getAttribute("data-action");
			if (!attr) return;
			const match = attr.match(/^(\w+)->playground#(\w+)$/);
			if (!match) return;
			const [, event, method] = match;
			const fn = (this as Record<string, unknown>)[method];
			if (typeof fn === "function") {
				el.addEventListener(event, fn.bind(this));
			}
		});
	}

	private _parseForkedFiles(code: string): void {
		const parts = code.split(/^\/\/ --- (.+?) ---$/m);
		this.files = [];
		for (let i = 1; i < parts.length; i += 2) {
			const name = parts[i].trim();
			const content = (parts[i + 1] || "").trim();
			if (name) {
				this.files.push({ name, content });
			}
		}
		if (this.files.length === 0) {
			this.files = [{ name: "main.gno", content: code }];
		}
		this._codeEl.value = this.files[0].content;
	}

	private _setupKeyboard(): void {
		this._codeEl.addEventListener("keydown", (e: KeyboardEvent) => {
			if (e.ctrlKey && e.key === "Enter") {
				e.preventDefault();
				this.runCode();
				return;
			}
			if (e.key === "Tab" && !e.shiftKey) {
				e.preventDefault();
				const start = this._codeEl.selectionStart;
				const end = this._codeEl.selectionEnd;
				this._codeEl.value =
					this._codeEl.value.substring(0, start) +
					"\t" +
					this._codeEl.value.substring(end);
				this._codeEl.selectionStart = this._codeEl.selectionEnd = start + 1;
			}
		});
	}

	private _renderTabs(): void {
		// Clear existing tabs using safe DOM methods
		while (this._tabsEl.firstChild) {
			this._tabsEl.removeChild(this._tabsEl.firstChild);
		}

		this.files.forEach((f, i) => {
			const btn = document.createElement("button");
			btn.className = `b-playground-tab${i === this.activeFile ? " b-playground-tab--active" : ""}`;
			btn.textContent = f.name;
			btn.addEventListener("click", () => this._switchToFile(f.name));
			this._tabsEl.appendChild(btn);
		});

		const addBtn = document.createElement("button");
		addBtn.className = "b-playground-tab-add";
		addBtn.textContent = "+";
		addBtn.title = "Add file";
		addBtn.addEventListener("click", () => this.addFile());
		this._tabsEl.appendChild(addBtn);
	}

	private _switchToFile(fileName: string): void {
		this.files[this.activeFile].content = this._codeEl.value;
		const idx = this.files.findIndex((f) => f.name === fileName);
		if (idx >= 0) {
			this.activeFile = idx;
			this._codeEl.value = this.files[idx].content;
			this._renderTabs();
		}
	}

	public switchTab(event: Event): void {
		const target = event.currentTarget as HTMLElement;
		const fileName = target.dataset.playgroundFileParam || "";
		this._switchToFile(fileName);
	}

	public addFile(): void {
		const name = prompt("File name (e.g. helper.gno):");
		if (!name) return;
		if (!name.endsWith(".gno")) {
			alert("File name must end with .gno");
			return;
		}
		if (this.files.some((f) => f.name === name)) {
			alert("File already exists");
			return;
		}

		this.files[this.activeFile].content = this._codeEl.value;
		this.files.push({ name, content: `package main\n` });
		this.activeFile = this.files.length - 1;
		this._codeEl.value = this.files[this.activeFile].content;
		this._renderTabs();
	}

	public async runCode(): Promise<void> {
		this.files[this.activeFile].content = this._codeEl.value;
		this._outputEl.textContent = "Running...";

		const remote = this.getValue("remote");
		const domain = this.getValue("domain");
		const code = this._codeEl.value;
		const pkgMatch = code.match(/^package\s+(\w+)/m);
		const pkgName = pkgMatch ? pkgMatch[1] : "main";

		if (code.includes("func Render(")) {
			try {
				const resp = await fetch("/_/api/eval", {
					method: "POST",
					headers: { "Content-Type": "application/json" },
					body: JSON.stringify({
						pkg_path: `${domain}/r/playground_preview`,
						expression: `Render("")`,
					}),
				});
				const result = await resp.json();
				if (result.error) {
					this._outputEl.textContent = `Error: ${result.error}`;
					this._outputEl.classList.add("u-color-danger");
				} else {
					this._outputEl.textContent = result.result;
					this._outputEl.classList.remove("u-color-danger");
				}
			} catch {
				this._outputEl.textContent = `Note: Server-side execution requires a running gno node.\n\nPackage: ${pkgName}\nFiles: ${this.files.map((f) => f.name).join(", ")}\n\nTo deploy and test, use:\n  gnokey maketx addpkg -pkgpath "${domain}/r/yourname/pkg" ...`;
				this._outputEl.classList.remove("u-color-danger");
			}
		} else {
			this._outputEl.textContent = `Package: ${pkgName}\nFiles: ${this.files.map((f) => f.name).join(", ")}\n\nTo run locally:\n  gno run ${this.files.map((f) => f.name).join(" ")}\n\nTo test:\n  gno test .`;
			this._outputEl.classList.remove("u-color-danger");
		}
	}

	public runTests(): void {
		this._outputEl.textContent =
			"Testing requires a running gno node.\n\nTo test locally:\n  gno test .";
	}

	public formatCode(): void {
		this._outputEl.textContent =
			"Formatting requires server-side gno fmt (coming soon).\n\nTo format locally:\n  gno fmt -w " +
			this.files[this.activeFile].name;
	}

	public shareCode(): void {
		this.files[this.activeFile].content = this._codeEl.value;

		const code =
			this.files.length === 1
				? this.files[0].content
				: this.files
						.map((f) => `// --- ${f.name} ---\n${f.content}`)
						.join("\n\n");

		const encoded = encodeURIComponent(code);
		const url = `${window.location.origin}/_/play?code=${encoded}`;

		navigator.clipboard
			.writeText(url)
			.then(() => {
				this._outputEl.textContent = "Share URL copied to clipboard!";
			})
			.catch(() => {
				this._outputEl.textContent = `Share URL:\n${url}`;
			});
	}

	public clearOutput(): void {
		this._outputEl.textContent = "// Run code to see output here";
		this._outputEl.classList.remove("u-color-danger");
	}
}
