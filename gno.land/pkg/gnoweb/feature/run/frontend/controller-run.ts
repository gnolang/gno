import { CodeEditor } from "@gnoweb/js/code-editor.js";
import { BaseController, downloadBlob } from "@gnoweb/js/controller.js";
import { heredocDelimiter, shellQuote } from "./shell.js";

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

		if (!this.editorEl || !this.cmdEl) return;

		this.editor = new CodeEditor({
			parent: this.editorEl,
			content: this._buildTemplate(),
			fileName: "script.gno",
			label: "Run script editor",
			onChange: () => this._updateCommand(),
		});

		this._setupInputListeners();
		this._updateCommand();
	}

	private _buildTemplate(): string {
		return `package main

import "${this.pkgPath}"

func main() {
\t// Call ${this.pkgAlias} functions here, e.g.:
\tprintln(${this.pkgAlias}.Render(""))
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
		// KEYNAME, not <key-name>: a shell reads "<" as a redirect.
		const keyName = this.keyEl.value.trim();
		const key = keyName ? shellQuote(keyName) : "KEYNAME";
		const gasWanted = shellQuote(
			this.gasWantedEl.value.trim() || "1_000_000_000",
		);
		const gasFee = shellQuote(this.gasFeeEl.value.trim() || "1000000ugnot");
		const send = this.sendEl.value.trim();

		const parts = [
			"gnokey maketx run",
			`  -gas-wanted ${gasWanted}`,
			`  -gas-fee ${gasFee}`,
		];

		if (send && send !== "0ugnot") {
			parts.push(`  -send ${shellQuote(send)}`);
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
		const delim = heredocDelimiter(code);
		return `cat > script.gno <<'${delim}'\n${code}\n${delim}`;
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

	public downloadCode(): void {
		downloadBlob(
			new Blob([this.editor.getCode()], { type: "text/plain" }),
			"script.gno",
		);
	}
}
