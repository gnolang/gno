import { BaseController } from "./controller.js";

export class RunController extends BaseController {
	private declare pkgPath: string;
	private declare pkgAlias: string;
	private declare remote: string;
	private declare chainId: string;
	private declare codeEl: HTMLTextAreaElement;
	private declare keyEl: HTMLInputElement;
	private declare gasWantedEl: HTMLInputElement;
	private declare gasFeeEl: HTMLInputElement;
	private declare sendEl: HTMLInputElement;
	private declare dryRunCmdEl: HTMLElement;
	private declare executeCmdEl: HTMLElement;

	protected connect(): void {
		this.pkgPath = this.getValue("pkg-path");
		this.pkgAlias = this.getValue("pkg-alias") || "pkg";
		this.remote = this.getValue("remote");
		this.chainId = this.getValue("chain-id");
		this.codeEl = this.getTarget("code") as HTMLTextAreaElement;
		this.keyEl = this.getTarget("key") as HTMLInputElement;
		this.gasWantedEl = this.getTarget("gasWanted") as HTMLInputElement;
		this.gasFeeEl = this.getTarget("gasFee") as HTMLInputElement;
		this.sendEl = this.getTarget("send") as HTMLInputElement;
		this.dryRunCmdEl = this.getTarget("dryRunCmd") as HTMLElement;
		this.executeCmdEl = this.getTarget("executeCmd") as HTMLElement;

		if (!this.codeEl || !this.dryRunCmdEl || !this.executeCmdEl) return;

		this.codeEl.value = this._buildTemplate();

		this._setupKeyboardShortcuts();
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

	private _setupKeyboardShortcuts(): void {
		this.codeEl.addEventListener("keydown", (e: KeyboardEvent) => {
			if (e.key === "Tab" && !e.shiftKey) {
				e.preventDefault();
				const start = this.codeEl.selectionStart;
				const end = this.codeEl.selectionEnd;
				this.codeEl.value = `${this.codeEl.value.substring(0, start)}\t${this.codeEl.value.substring(end)}`;
				this.codeEl.selectionStart = this.codeEl.selectionEnd = start + 1;
			}
		});
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

		if (send && send !== "0ugnot") parts.push(`  -send "${send}"`);
		parts.push("  -broadcast");
		if (dryRun) parts.push("  -simulate only");
		if (this.chainId) parts.push(`  -chainid ${this.chainId}`);
		if (this.remote) parts.push(`  -remote "${this.remote}"`);
		parts.push(`  ${key} script.gno`);

		return parts.join(" \\\n");
	}

	private _updateCommands(): void {
		this.dryRunCmdEl.textContent = this._buildCmd(true);
		this.executeCmdEl.textContent = this._buildCmd(false);
	}

	private _copyToClipboard(text: string, btn: HTMLElement): void {
		navigator.clipboard
			.writeText(text)
			.then(() => {
				const span = btn.querySelector("span");
				if (!span) return;
				const orig = span.textContent;
				span.textContent = "Copied!";
				setTimeout(() => {
					span.textContent = orig;
				}, 1500);
			})
			.catch(() => {
				// fallback: select the pre element content
				const pre = btn.closest(".b-run-command-block")?.querySelector("pre");
				if (pre) {
					const sel = window.getSelection();
					const range = document.createRange();
					range.selectNodeContents(pre);
					sel?.removeAllRanges();
					sel?.addRange(range);
				}
			});
	}

	public resetCode(): void {
		this.codeEl.value = this._buildTemplate();
	}

	public downloadCode(): void {
		const blob = new Blob([this.codeEl.value], { type: "text/plain" });
		const url = URL.createObjectURL(blob);
		const a = document.createElement("a");
		a.href = url;
		a.download = "script.gno";
		a.click();
		URL.revokeObjectURL(url);
	}

	public copyDryRun(event: Event): void {
		this._copyToClipboard(
			this._buildCmd(true),
			event.currentTarget as HTMLElement,
		);
	}

	public copyExecute(event: Event): void {
		this._copyToClipboard(
			this._buildCmd(false),
			event.currentTarget as HTMLElement,
		);
	}
}
