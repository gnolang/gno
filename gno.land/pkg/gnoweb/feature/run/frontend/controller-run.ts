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
	// Absent when the block is embedded in a page that already has a key input
	// (the action page); then the key arrives via "address:changed" instead.
	private declare keyEl: HTMLInputElement | null;
	// `declare` + assigned in connect(): connect() runs during super(), so a
	// field initialiser here would not exist yet when the first render reads it.
	private declare broadcastKey: string;
	private declare gasWantedEl: HTMLInputElement;
	private declare gasFeeEl: HTMLInputElement;
	private declare sendEl: HTMLInputElement;
	private declare dryRunEl: HTMLInputElement;
	private declare includeScriptEl: HTMLInputElement;
	private declare cmdEl: HTMLElement;
	private declare editor: CodeEditor;

	protected connect(): void {
		this.pkgPath = this.getValue("pkg-path");
		this.pkgAlias = this.getValue("pkg-alias") || "pkg";
		this.remote = this.getValue("remote");
		this.chainId = this.getValue("chain-id");
		this.editorEl = this.getTarget("editor") as HTMLElement;
		this.keyEl = this.getTarget("key") as HTMLInputElement | null;
		this.broadcastKey = "";
		this.gasWantedEl = this.getTarget("gasWanted") as HTMLInputElement;
		this.gasFeeEl = this.getTarget("gasFee") as HTMLInputElement;
		this.sendEl = this.getTarget("send") as HTMLInputElement;
		this.dryRunEl = this.getTarget("dryRun") as HTMLInputElement;
		this.includeScriptEl = this.getTarget("includeScript") as HTMLInputElement;
		this.cmdEl = this.getTarget("cmd") as HTMLElement;

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

		// Without a key field of its own, follow the page-level Key/Address input
		// that the action-header controller broadcasts (it replays the stored
		// value once every controller is ready).
		if (!this.keyEl) {
			this.on("address:changed", (event: Event) => {
				this.broadcastKey = (event as CustomEvent).detail?.address ?? "";
				this._updateCommand();
			});
		}

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
		this.keyEl?.addEventListener("input", update);
		this.gasWantedEl.addEventListener("input", update);
		this.gasFeeEl.addEventListener("input", update);
		this.sendEl.addEventListener("input", update);
		this.dryRunEl.addEventListener("change", update);
		this.includeScriptEl.addEventListener("change", update);
	}

	private _buildCmd(dryRun: boolean): string {
		const key = (this.keyEl?.value ?? this.broadcastKey).trim() || "ADDRESS";
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

	// Wraps the editor content in a quoted heredoc so the script can be written
	// from the same paste as the command.
	private _buildHeredoc(): string {
		const code = this.editor.getCode().replace(/\n+$/, "");
		return `cat > script.gno <<'${HEREDOC_DELIMITER}'\n${code}\n${HEREDOC_DELIMITER}`;
	}

	private _updateCommand(): void {
		const cmd = this._buildCmd(this.dryRunEl.checked);
		this.cmdEl.textContent = this.includeScriptEl.checked
			? `${this._buildHeredoc()}\n\n${cmd}`
			: cmd;
	}

	// When the block sits inside a collapsed <details>, the editor is built in a
	// container with no layout, so it measures itself as empty. Re-measure once
	// the section is revealed.
	public refreshEditor(): void {
		this.editor?.view.requestMeasure();
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
