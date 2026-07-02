import { CodeEditor, isDarkMode } from "@gnoweb/js/code-editor.js";
import { BaseController } from "@gnoweb/js/controller.js";

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
	private declare dryRunCmdEl: HTMLElement;
	private declare executeCmdEl: HTMLElement;
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
		this.dryRunCmdEl = this.getTarget("dryRunCmd") as HTMLElement;
		this.executeCmdEl = this.getTarget("executeCmd") as HTMLElement;

		if (!this.editorEl || !this.dryRunCmdEl || !this.executeCmdEl) return;

		this.editor = new CodeEditor({
			parent: this.editorEl,
			content: this._buildTemplate(),
			fileName: "script.gno",
			isDarkMode: isDarkMode(),
		});

		this.on("theme:changed", () => {
			this.editor.changeTheme(isDarkMode());
		});

		this._setupInputListeners();
		this._updateCommands();
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
		const update = (): void => this._updateCommands();
		this.keyEl.addEventListener("input", update);
		this.gasWantedEl.addEventListener("input", update);
		this.gasFeeEl.addEventListener("input", update);
		this.sendEl.addEventListener("input", update);
	}

	private _buildCmd(dryRun: boolean): string {
		const key = this.keyEl.value.trim() || "<key-name>";
		const gasWanted = this.gasWantedEl.value.trim() || "2000000";
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

		if (dryRun) {
			parts.push("  -simulate only");
		}

		if (this.chainId) {
			parts.push(`  -chainid ${this.chainId}`);
		}

		if (this.remote) {
			parts.push(`  -remote "${this.remote}"`);
		}

		parts.push(`  ${key} script.gno`);
		return parts.join(" \\\n");
	}

	private _updateCommands(): void {
		this.dryRunCmdEl.textContent = this._buildCmd(true);
		this.executeCmdEl.textContent = this._buildCmd(false);
	}

	public resetCode(): void {
		this.editor.setCode(this._buildTemplate());
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
