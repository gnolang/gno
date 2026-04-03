import { BaseController } from "./controller.js";

interface HistoryEntry {
	expression: string;
	result: string;
	isError: boolean;
}

export class EvalController extends BaseController {
	private _inputEl!: HTMLInputElement;
	private _resultEl!: HTMLElement;
	private _historyListEl!: HTMLElement;
	private _history: HistoryEntry[] = [];

	protected connect(): void {
		this._inputEl = this.getTarget("input") as HTMLInputElement;
		this._resultEl = this.getTarget("result") as HTMLElement;
		this._historyListEl = this.getTarget("history-list") as HTMLElement;

		this._inputEl?.focus();

		this._inputEl?.addEventListener("keydown", (e: KeyboardEvent) => {
			if (e.key === "ArrowUp" && this._history.length > 0) {
				e.preventDefault();
				this._inputEl.value =
					this._history[this._history.length - 1].expression;
			}
		});
	}

	public evalExpression(event: Event): void {
		event.preventDefault();
		const expr = this._inputEl.value.trim();
		if (!expr) return;
		this._doEval(expr);
	}

	public quickCall(event: Event & { params?: Record<string, unknown> }): void {
		const funcName = event.params?.funcName as string;
		const funcSig = event.params?.funcSig as string;

		if (funcName) {
			const paramMatch = funcSig?.match(/\(([^)]*)\)/);
			const params = paramMatch ? paramMatch[1] : "";

			if (params) {
				this._inputEl.value = `${funcName}(${params
					.split(",")
					.map((p: string) => {
						const parts = p.trim().split(/\s+/);
						const type = parts[parts.length - 1];
						return type === "string" ? '""' : "0";
					})
					.join(", ")})`;
				this._inputEl.focus();
				this._inputEl.setSelectionRange(
					funcName.length + 1,
					this._inputEl.value.length - 1,
				);
			} else {
				this._inputEl.value = `${funcName}()`;
				this._doEval(`${funcName}()`);
			}
		}
	}

	public clearResult(): void {
		this._resultEl.textContent = "// Enter an expression above";
		this._resultEl.classList.remove("u-color-danger");
	}

	private async _doEval(expression: string): Promise<void> {
		const pkgPath = this.getValue("pkg-path");
		const domain = this.getValue("domain");

		this._resultEl.textContent = "Evaluating...";
		this._resultEl.classList.remove("u-color-danger");

		try {
			// Strip domain prefix from pkgPath for the API (it expects relative path)
			const relPath = pkgPath.startsWith(domain + "/")
				? pkgPath.slice(domain.length + 1)
				: pkgPath;

			const response = await fetch("/_/api/eval", {
				method: "POST",
				headers: { "Content-Type": "application/json" },
				body: JSON.stringify({
					pkg_path: relPath,
					expression: expression,
				}),
			});

			if (!response.ok) {
				throw new Error(`HTTP ${response.status}`);
			}

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

			this._resultEl.textContent = result;
			this._resultEl.classList.toggle("u-color-danger", isError);

			this._addToHistory(expression, result, isError);
		} catch (err) {
			const errMsg = `Error: ${err instanceof Error ? err.message : String(err)}`;
			this._resultEl.textContent = errMsg;
			this._resultEl.classList.add("u-color-danger");
			this._addToHistory(expression, errMsg, true);
		}
	}

	private _addToHistory(
		expression: string,
		result: string,
		isError: boolean,
	): void {
		this._history.push({ expression, result, isError });

		if (this._historyListEl) {
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
			rerunBtn.textContent = "\u21A9"; // ↩
			rerunBtn.title = "Re-run";
			rerunBtn.addEventListener("click", () => {
				this._inputEl.value = expression;
				this._doEval(expression);
			});
			exprDiv.appendChild(rerunBtn);

			const resultPre = document.createElement("pre");
			resultPre.className = `b-eval-history-result${isError ? " u-color-danger" : ""}`;
			resultPre.textContent =
				result.length > 200 ? result.substring(0, 200) + "..." : result;

			entry.appendChild(exprDiv);
			entry.appendChild(resultPre);
			this._historyListEl.prepend(entry);
		}
	}
}
