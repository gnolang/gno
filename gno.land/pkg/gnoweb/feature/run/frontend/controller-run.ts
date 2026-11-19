import { CodeEditor } from "@gnoweb/js/code-editor.js";
import {
	addressSlotText,
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
	// The address typed in the header, empty for the placeholder.
	private declare address: string;

	protected connect(): void {
		this.script = this.getValue("script");
		this.gasWantedEl = this.getTarget("gasWanted") as HTMLInputElement;
		this.gasFeeEl = this.getTarget("gasFee") as HTMLInputElement;
		this.sendEl = this.getTarget("send") as HTMLInputElement;
		this.includeScriptEl = this.getTarget("includeScript") as HTMLInputElement;
		const editorEl = this.getTarget("editor");
		if (!editorEl) return;
		this.address = "";
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

		const update = () => this._updateCommand();
		for (const el of [this.gasWantedEl, this.gasFeeEl, this.sendEl]) {
			el.addEventListener("input", update);
		}
		this.includeScriptEl.addEventListener("change", update);
		this.on("mode:changed", (e) =>
			this._showMode((e as CustomEvent).detail.mode),
		);
		this.on("address:changed", (e) => {
			this.address = String((e as CustomEvent).detail.address);
			this._updateCommand();
		});
		// The browser may have restored the fields of a reloaded page.
		this._updateCommand();
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
		this._fill("address-arg", addressSlotText(this.address));
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
