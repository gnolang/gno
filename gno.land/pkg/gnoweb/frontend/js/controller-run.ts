// Run (maketx run) controller — pre-fills editor with import template, generates gnokey commands
export default function RunController(element: HTMLElement): void {
	const codeEl = element.querySelector('[data-run-target="code"]') as HTMLTextAreaElement;
	const keyEl = element.querySelector('[data-run-target="key"]') as HTMLInputElement;
	const gasWantedEl = element.querySelector('[data-run-target="gasWanted"]') as HTMLInputElement;
	const gasFeeEl = element.querySelector('[data-run-target="gasFee"]') as HTMLInputElement;
	const sendEl = element.querySelector('[data-run-target="send"]') as HTMLInputElement;
	const dryRunCmdEl = element.querySelector('[data-run-target="dryRunCmd"]') as HTMLElement;
	const executeCmdEl = element.querySelector('[data-run-target="executeCmd"]') as HTMLElement;

	if (!codeEl || !dryRunCmdEl || !executeCmdEl) return;

	const pkgPath = element.getAttribute("data-run-pkg-path-value") || "";
	const pkgAlias = element.getAttribute("data-run-pkg-alias-value") || "pkg";
	const remote = element.getAttribute("data-run-remote-value") || "";
	const chainId = element.getAttribute("data-run-chain-id-value") || "";

	const initialCode = buildTemplate(pkgPath, pkgAlias);
	codeEl.value = initialCode;

	// Tab key inserts spaces instead of changing focus
	codeEl.addEventListener("keydown", (e: KeyboardEvent) => {
		if (e.key === "Tab" && !e.shiftKey) {
			e.preventDefault();
			const start = codeEl.selectionStart;
			const end = codeEl.selectionEnd;
			codeEl.value = codeEl.value.substring(0, start) + "\t" + codeEl.value.substring(end);
			codeEl.selectionStart = codeEl.selectionEnd = start + 1;
		}
	});

	function buildTemplate(pkg: string, alias: string): string {
		return `package main

import "${pkg}"

func main() {
\t// Call ${alias} functions here, e.g.:
\t// ${alias}.Render("")
}
`;
	}

	function buildCmd(dryRun: boolean): string {
		const key = keyEl?.value.trim() || "<key-name>";
		const gasWanted = gasWantedEl?.value.trim() || "2000000";
		const gasFee = gasFeeEl?.value.trim() || "1000000ugnot";
		const send = sendEl?.value.trim();

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

		if (chainId) {
			parts.push(`  -chainid ${chainId}`);
		}

		if (remote) {
			parts.push(`  -remote "${remote}"`);
		}

		parts.push(`  ${key} script.gno`);

		return parts.join(" \\\n");
	}

	function updateCommands(): void {
		dryRunCmdEl.textContent = buildCmd(true);
		executeCmdEl.textContent = buildCmd(false);
	}

	// Update commands on any input change
	[keyEl, gasWantedEl, gasFeeEl, sendEl].forEach((el) => {
		if (el) el.addEventListener("input", updateCommands);
	});

	// Wire actions
	element.querySelectorAll("[data-action]").forEach((el) => {
		const attr = el.getAttribute("data-action");
		if (!attr) return;
		const match = attr.match(/^(\w+)->run#(\w+)$/);
		if (!match) return;
		const [, evt, method] = match;

		const handlers: Record<string, () => void> = {
			resetCode: () => {
				codeEl.value = buildTemplate(pkgPath, pkgAlias);
			},
			downloadCode: () => {
				const blob = new Blob([codeEl.value], { type: "text/plain" });
				const url = URL.createObjectURL(blob);
				const a = document.createElement("a");
				a.href = url;
				a.download = "script.gno";
				a.click();
				URL.revokeObjectURL(url);
			},
			copyDryRun: () => copyToClipboard(buildCmd(true), el as HTMLElement),
			copyExecute: () => copyToClipboard(buildCmd(false), el as HTMLElement),
		};

		const fn = handlers[method];
		if (fn) el.addEventListener(evt, fn);
	});

	function copyToClipboard(text: string, btn: HTMLElement): void {
		navigator.clipboard.writeText(text).then(() => {
			const span = btn.querySelector("span");
			if (!span) return;
			const orig = span.textContent;
			span.textContent = "Copied!";
			setTimeout(() => { span.textContent = orig; }, 1500);
		}).catch(() => {
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

	updateCommands();
}
