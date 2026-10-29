import { CodeEditor } from "@gnoweb/js/code-editor.js";
import {
	BaseController,
	downloadBlob,
	shellQuote,
} from "@gnoweb/js/controller.js";
import { heredocDelimiter } from "./shell.js";

// The page serves the script and both modes of the command whole. This
// controller turns the script into an editor and fills the command's slots
// from the fields; the mode and the key come from the action-header
// controller, which stores them for $help too.
export class RunController extends BaseController {
	// The script the page was served with, which Reset goes back to.
	private declare script: string;
	private declare gasWantedEl: HTMLInputElement;
	private declare gasFeeEl: HTMLInputElement;
	private declare sendEl: HTMLInputElement;
	private declare includeScriptEl: HTMLInputElement;
	private declare editor: CodeEditor;
	// The key or address typed in the header, empty for the placeholders.
	private declare key: string;
	private declare onMode: (e: Event) => void;
	private declare onAddress: (e: Event) => void;
	private declare onInput: () => void;

	protected connect(): void {
		this.script = this.getValue("script");
		this.gasWantedEl = this.getTarget("gasWanted") as HTMLInputElement;
		this.gasFeeEl = this.getTarget("gasFee") as HTMLInputElement;
		this.sendEl = this.getTarget("send") as HTMLInputElement;
		this.includeScriptEl = this.getTarget("includeScript") as HTMLInputElement;
		const editorEl = this.getTarget("editor");
		if (!editorEl) return;
		this.key = "";
		// Served hidden: these controls only work from here on.
		for (const el of this.getTargets("js-only")) el.hidden = false;

		// The editor replaces the highlighted script the page was served with.
		editorEl.replaceChildren();
		this.editor = new CodeEditor({
			parent: editorEl,
			content: this.script,
			fileName: "script.gno",
			label: "Run script editor",
			onChange: () => this._updateCommand(),
		});

		this.onInput = () => this._updateCommand();
		this.onMode = (e) => this._showMode((e as CustomEvent).detail.mode);
		this.onAddress = (e) => {
			this.key = String((e as CustomEvent).detail.address).trim();
			this._updateCommand();
		};
		for (const el of [this.gasWantedEl, this.gasFeeEl, this.sendEl]) {
			el.addEventListener("input", this.onInput);
		}
		this.includeScriptEl.addEventListener("change", this.onInput);
		this.on("mode:changed", this.onMode);
		this.on("address:changed", this.onAddress);
		// The browser may have restored the fields of a reloaded page.
		this._updateCommand();
	}

	protected disconnect(): void {
		for (const el of [this.gasWantedEl, this.gasFeeEl, this.sendEl]) {
			el?.removeEventListener("input", this.onInput);
		}
		this.includeScriptEl?.removeEventListener("change", this.onInput);
		document.removeEventListener("mode:changed", this.onMode);
		document.removeEventListener("address:changed", this.onAddress);
	}

	// Shows the command of the mode picked, and makes it the one Copy copies.
	private _showMode(mode: string): void {
		for (const el of this.getTargets("mode")) {
			const shown = this.getValue("mode", el) === mode;
			el.classList.toggle("u-inline", shown);
			el.classList.toggle("u-hidden", !shown);
			if (shown) el.dataset.copyTarget = "run-cmd";
			else delete el.dataset.copyTarget;
		}
	}

	private _fill(target: string, text: string): void {
		for (const el of this.getTargets(target)) el.textContent = text;
	}

	private _updateCommand(): void {
		// KEYNAME, not <key-name>: a shell reads "<" as a redirect.
		const key = this.key ? shellQuote(this.key) : "";
		this._fill("key-arg", key || "KEYNAME");
		this._fill("address-arg", key || "ADDRESS");
		// An emptied field falls back to the default the page was served with.
		this._fill(
			"gas-wanted-arg",
			shellQuote(
				this.gasWantedEl.value.trim() || this.gasWantedEl.defaultValue,
			),
		);
		this._fill(
			"gas-fee-arg",
			shellQuote(this.gasFeeEl.value.trim() || this.gasFeeEl.defaultValue),
		);
		const send = this.sendEl.value.trim();
		this._fill(
			"send-arg",
			send && send !== "0ugnot" ? ` -send ${shellQuote(send)}` : "",
		);
		this._fill("heredoc", this._heredoc());
	}

	// The heredoc that writes the editor's script from the same paste as the
	// command, or, with the script left out, the line that asks for the file.
	private _heredoc(): string {
		if (!this.includeScriptEl.checked) {
			return "# Save the script above as script.gno (Download), then run:\n";
		}
		const code = this.editor.getCode().replace(/\n+$/, "");
		const delim = heredocDelimiter(code);
		return `cat > script.gno <<'${delim}'\n${code}\n${delim}\n\n`;
	}

	public resetCode(): void {
		this.editor.setCode(this.script);
	}

	public downloadCode(): void {
		downloadBlob(
			new Blob([this.editor.getCode()], { type: "text/plain" }),
			"script.gno",
		);
	}
}
