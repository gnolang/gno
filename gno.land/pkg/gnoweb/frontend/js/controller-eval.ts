import { BaseController } from "./controller.js";

interface HistoryEntry {
	expression: string;
	result: string;
	isError: boolean;
}

export class EvalController extends BaseController {
	private declare history: HistoryEntry[];
	private declare inputEl: HTMLInputElement;
	private declare resultEl: HTMLElement;
	private declare historyListEl: HTMLElement | null;
	private declare historySection: HTMLElement | null;

	protected connect(): void {
		this.history = [];
		this.inputEl = this.getTarget("input") as HTMLInputElement;
		this.resultEl = this.getTarget("result") as HTMLElement;
		this.historyListEl = this.getTarget("history-list") as HTMLElement | null;
		this.historySection = this.getTarget("history") as HTMLElement | null;

		if (!this.inputEl || !this.resultEl) return;

		this._setupKeyboardShortcuts();
		this.inputEl.focus();
	}

	private _setupKeyboardShortcuts(): void {
		this.inputEl.addEventListener("keydown", (e: KeyboardEvent) => {
			if (e.key === "ArrowUp" && this.history.length > 0) {
				e.preventDefault();
				this.inputEl.value = this.history[this.history.length - 1].expression;
			}
		});
	}

	private async _doEval(expression: string): Promise<void> {
		this.resultEl.textContent = "Evaluating...";
		this.resultEl.classList.remove("u-color-danger");

		try {
			const pkgPath = this.getValue("pkg-path");
			const domain = this.getValue("domain") || "gno.land";
			const relPath = pkgPath.startsWith(`${domain}/`)
				? pkgPath.slice(domain.length + 1)
				: pkgPath;

			const response = await fetch("/_/api/eval", {
				method: "POST",
				headers: { "Content-Type": "application/json" },
				body: JSON.stringify({ pkg_path: relPath, expression }),
			});

			if (response.status === 429)
				throw new Error("rate limit exceeded — please wait a moment");

			if (!response.ok) throw new Error(`HTTP ${response.status}`);
			const json = await response.json();

			let result: string;
			let isError: boolean;
			if (json.error) {
				result = `Error: ${json.error}`;
				isError = true;
			} else {
				result = json.result || "(empty result)";
				isError = false;
			}

			this.resultEl.textContent = result;
			this.resultEl.classList.toggle("u-color-danger", isError);
			this._addToHistory(expression, result, isError);
		} catch (err) {
			const errMsg = `Error: ${err instanceof Error ? err.message : String(err)}`;
			this.resultEl.textContent = errMsg;
			this.resultEl.classList.add("u-color-danger");
			this._addToHistory(expression, errMsg, true);
		}
	}

	private _addToHistory(
		expression: string,
		result: string,
		isError: boolean,
	): void {
		this.history.push({ expression, result, isError });
		if (!this.historyListEl) return;

		// Reveal the history section on first entry
		if (this.historySection) this.historySection.removeAttribute("hidden");

		const entry = document.createElement("div");
		entry.className = "b-eval-history-entry";

		const exprDiv = document.createElement("div");
		exprDiv.className = "b-eval-history-expr";

		const codeEl = document.createElement("code");
		codeEl.className = "u-font-mono";
		codeEl.textContent = expression;
		exprDiv.appendChild(codeEl);

		const rerunBtn = document.createElement("button");
		rerunBtn.className = "b-eval-history-rerun";
		rerunBtn.textContent = "↩";
		rerunBtn.title = "Re-run";
		rerunBtn.addEventListener("click", () => {
			this.inputEl.value = expression;
			this._doEval(expression);
		});
		exprDiv.appendChild(rerunBtn);

		const resultPre = document.createElement("pre");
		resultPre.className = `b-eval-history-result${isError ? " u-color-danger" : ""}`;
		resultPre.textContent =
			result.length > 200 ? `${result.substring(0, 200)}...` : result;

		entry.appendChild(exprDiv);
		entry.appendChild(resultPre);
		this.historyListEl.prepend(entry);
	}

	public evalExpression(event: Event): void {
		event.preventDefault();
		const expr = this.inputEl.value.trim();
		if (expr) this._doEval(expr);
	}

	public clearResult(): void {
		this.resultEl.textContent = "// Enter an expression above";
		this.resultEl.classList.remove("u-color-danger");
	}

	public quickCall(event: Event & { params?: Record<string, unknown> }): void {
		const funcName = (event.params?.funcName as string) || "";
		const funcSig = (event.params?.funcSig as string) || "";
		if (!funcName) return;

		const paramMatch = funcSig.match(/\(([^)]*)\)/);
		const params = paramMatch ? paramMatch[1] : "";
		if (params) {
			this.inputEl.value = `${funcName}(${params
				.split(",")
				.map((p: string) => {
					const parts = p.trim().split(/\s+/);
					const type = parts[parts.length - 1];
					return type === "string" ? '""' : "0";
				})
				.join(", ")})`;

			this.inputEl.focus();
			this.inputEl.setSelectionRange(
				funcName.length + 1,
				this.inputEl.value.length - 1,
			);
		} else {
			this.inputEl.value = `${funcName}()`;
			this._doEval(`${funcName}()`);
		}
	}
}
