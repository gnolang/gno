import { CodeEditor, isDarkMode } from "@gnoweb/js/code-editor.js";
import { BaseController } from "@gnoweb/js/controller.js";

// Terminates the heredoc that carries the script.
const HEREDOC_DELIMITER = "__GNO_EOF__";

export class RunController extends BaseController {
	private declare pkgPath: string;
	private declare pkgAlias: string;
	private declare remote: string;
	private declare chainId: string;
	private declare editorEl: HTMLElement;
	private declare keyEl: HTMLInputElement;
	private declare gasWantedEl: HTMLInputElement;
	private declare gasFeeEl: HTMLInputElement;
	private declare sendEl: HTMLInputElement;
	private declare includeScriptEl: HTMLInputElement;
	private declare cmdEl: HTMLElement;
	private declare resultEl: HTMLElement;
	private declare editor: CodeEditor;

	protected connect(): void {
		this.pkgPath = this.getValue("pkg-path");
		this.pkgAlias = this.getValue("pkg-alias") || "pkg";
		this.remote = this.getValue("remote");
		this.chainId = this.getValue("chain-id");
		this.editorEl = this.getTarget("editor") as HTMLElement;
		this.keyEl = this.getTarget("key") as HTMLInputElement;
		this.gasWantedEl = this.getTarget("gasWanted") as HTMLInputElement;
		this.gasFeeEl = this.getTarget("gasFee") as HTMLInputElement;
		this.sendEl = this.getTarget("send") as HTMLInputElement;
		this.includeScriptEl = this.getTarget("includeScript") as HTMLInputElement;
		this.cmdEl = this.getTarget("cmd") as HTMLElement;
		this.resultEl = this.getTarget("result") as HTMLElement;

		if (!this.editorEl || !this.cmdEl) return;

		this.editor = new CodeEditor({
			parent: this.editorEl,
			content: this._buildTemplate(),
			fileName: "script.gno",
			isDarkMode: isDarkMode(),
			onChange: () => this._updateCommand(),
		});

		this.on("theme:changed", () => {
			this.editor.changeTheme(isDarkMode());
		});

		this._setupInputListeners();
		this._updateCommand();
	}

	private _buildTemplate(): string {
		return `package main

import "${this.pkgPath}"

func main() {
\t// Call ${this.pkgAlias} functions here, e.g.:
\t// ${this.pkgAlias}.Render("")
}
`;
	}

	private _setupInputListeners(): void {
		const update = (): void => this._updateCommand();
		this.keyEl.addEventListener("input", update);
		this.gasWantedEl.addEventListener("input", update);
		this.gasFeeEl.addEventListener("input", update);
		this.sendEl.addEventListener("input", update);
		this.includeScriptEl.addEventListener("change", update);
	}

	private _buildCmd(): string {
		const key = this.keyEl.value.trim() || "<key-name>";
		const gasWanted = this.gasWantedEl.value.trim() || "1_000_000_000";
		const gasFee = this.gasFeeEl.value.trim() || "1000000ugnot";
		const send = this.sendEl.value.trim();

		const parts = [
			"gnokey maketx run",
			`  -gas-wanted ${gasWanted}`,
			`  -gas-fee ${gasFee}`,
		];

		if (send && send !== "0ugnot") {
			parts.push(`  -send "${send}"`);
		}

		parts.push("  -broadcast");

		if (this.chainId) {
			parts.push(`  -chainid ${this.chainId}`);
		}

		if (this.remote) {
			parts.push(`  -remote "${this.remote}"`);
		}

		parts.push(`  ${key} script.gno`);
		return parts.join(" \\\n");
	}

	// Wraps the editor content in a quoted heredoc so the script can be written
	// from the same paste as the command.
	private _buildHeredoc(): string {
		const code = this.editor.getCode().replace(/\n+$/, "");
		return `cat > script.gno <<'${HEREDOC_DELIMITER}'\n${code}\n${HEREDOC_DELIMITER}`;
	}

	private _updateCommand(): void {
		const cmd = this._buildCmd();
		this.cmdEl.textContent = this.includeScriptEl.checked
			? `${this._buildHeredoc()}\n\n${cmd}`
			: cmd;
	}

	public resetCode(): void {
		this.editor.setCode(this._buildTemplate());
	}

	// Simulates the script against the remote node without broadcasting.
	public async dryRun(): Promise<void> {
		this._setResult("Running...", false);

		try {
			const response = await fetch("/_/api/dryrun", {
				method: "POST",
				headers: { "Content-Type": "application/json" },
				body: JSON.stringify({
					pkg_path: this.pkgPath,
					script: this.editor.getCode(),
					address: this.keyEl.value.trim(),
				}),
			});

			if (response.status === 429)
				throw new Error("rate limit exceeded — please wait a moment");

			const json = await response.json();
			if (json.error) {
				this._setResult(`Error: ${json.error}`, true);
			} else {
				this._setResult(json.result || "(no output)", false);
			}
		} catch (err) {
			const msg = err instanceof Error ? err.message : String(err);
			this._setResult(`Error: ${msg}`, true);
		}
	}

	private _setResult(text: string, isError: boolean): void {
		this.resultEl.textContent = text;
		this.resultEl.classList.toggle("u-color-danger", isError);
	}

	public clearResult(): void {
		this._setResult("// Press Dry Run to simulate the script above", false);
	}

	public downloadCode(): void {
		const blob = new Blob([this.editor.getCode()], { type: "text/plain" });
		const url = URL.createObjectURL(blob);
		const a = document.createElement("a");
		a.href = url;
		a.download = "script.gno";
		a.click();
		URL.revokeObjectURL(url);
	}
}
