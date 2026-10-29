import { BaseController, postEval } from "./controller.js";

export class ActionEvalController extends BaseController {
	// Past expressions, for ArrowUp recall.
	private declare history: string[];
	private declare inputEl: HTMLInputElement;
	private declare resultEl: HTMLElement;
	private declare historyListEl: HTMLElement | null;
	private declare historySection: HTMLElement | null;
	// The template's placeholder, shown again when a new expression arrives.
	private declare placeholder: string;
	// Bumped by every evaluation, so a slow response never overwrites a newer one.
	private declare evalSeq: number;
	// The evaluation in flight, aborted when a newer one starts.
	private declare evalAbort: AbortController | null;

	protected connect(): void {
		this.history = [];
		this.inputEl = this.getTarget("input") as HTMLInputElement;
		this.resultEl = this.getTarget("result") as HTMLElement;
		this.historyListEl = this.getTarget("history-list") as HTMLElement | null;
		this.historySection = this.getTarget("history") as HTMLElement | null;
		this.placeholder = this.resultEl?.textContent ?? "";
		this.evalSeq = 0;
		this.evalAbort = null;

		if (!this.inputEl || !this.resultEl) return;

		this._setupKeyboardShortcuts();
		this._listenForEvents();
	}

	protected disconnect(): void {
		this.evalAbort?.abort();
	}

	private _listenForEvents(): void {
		this.on("eval:expression", (e: Event) => {
			this.evalSeq++;
			this._show(this.placeholder, false);

			const { expression, autoEval } = (e as CustomEvent).detail;
			this.inputEl.value = expression;
			this.inputEl.scrollIntoView({ behavior: "smooth", block: "center" });

			if (autoEval) {
				this._doEval(expression);
			} else {
				this.inputEl.focus();
				const start = expression.indexOf("(") + 1;
				const end = expression.lastIndexOf(")");
				this.inputEl.setSelectionRange(start, end);
			}
		});
	}

	private _setupKeyboardShortcuts(): void {
		this.inputEl.addEventListener("keydown", (e: KeyboardEvent) => {
			if (e.key === "ArrowUp" && this.history.length > 0) {
				e.preventDefault();
				this.inputEl.value = this.history[this.history.length - 1];
			}
		});
	}

	private async _doEval(expression: string): Promise<void> {
		const seq = ++this.evalSeq;
		this.evalAbort?.abort();
		const abort = new AbortController();
		this.evalAbort = abort;
		this._show("Evaluating...", false);
		const { text, isError, unavailable } = await postEval(
			this.getValue("pkg-path"),
			expression,
			abort.signal,
		);
		if (seq !== this.evalSeq) return;
		this.evalAbort = null;
		const shown = text || "(empty result)";
		this._show(shown, isError);
		// A rate limit or an outage says nothing about the expression: keep
		// it out of history.
		if (!unavailable) this._addToHistory(expression, shown, isError);
	}

	private _show(text: string, isError: boolean): void {
		this.resultEl.textContent = text;
		this.resultEl.classList.toggle("u-color-danger", isError);
	}

	private _addToHistory(
		expression: string,
		result: string,
		isError: boolean,
	): void {
		this.history.push(expression);
		if (!this.historyListEl) return;

		if (this.historySection) this.historySection.removeAttribute("hidden");

		const entry = document.createElement("div");
		entry.className = "b-action-eval-history-entry";

		const exprDiv = document.createElement("div");
		exprDiv.className = "b-action-eval-history-expr c-between";

		const codeEl = document.createElement("code");
		codeEl.className = "u-font-mono";
		codeEl.textContent = expression;
		exprDiv.appendChild(codeEl);

		const rerunBtn = document.createElement("button");
		rerunBtn.className = "b-action-eval-history-rerun";
		rerunBtn.type = "button";
		rerunBtn.textContent = "↩";
		rerunBtn.title = "Re-run";
		rerunBtn.setAttribute("aria-label", `Re-run ${expression}`);
		rerunBtn.addEventListener("click", () => {
			this.inputEl.value = expression;
			this._doEval(expression);
		});
		exprDiv.appendChild(rerunBtn);

		const resultPre = document.createElement("pre");
		resultPre.className = `b-action-eval-history-result${isError ? " u-color-danger" : ""}`;
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
}
