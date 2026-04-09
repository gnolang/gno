// Expression evaluator controller — standalone, no BaseController dependency
interface HistoryEntry {
	expression: string;
	result: string;
	isError: boolean;
}

export default function EvalController(element: HTMLElement): void {
	const inputEl = element.querySelector(
		'[data-eval-target="input"]',
	) as HTMLInputElement;
	const resultEl = element.querySelector(
		'[data-eval-target="result"]',
	) as HTMLElement;
	const historyListEl = element.querySelector(
		'[data-eval-target="history-list"]',
	) as HTMLElement;

	if (!inputEl || !resultEl) return;

	const pkgPath = element.getAttribute("data-eval-pkg-path-value") || "";
	const domain = element.getAttribute("data-eval-domain-value") || "gno.land";
	const history: HistoryEntry[] = [];

	inputEl.focus();

	async function doEval(expression: string): Promise<void> {
		resultEl.textContent = "Evaluating...";
		resultEl.classList.remove("u-color-danger");

		try {
			const relPath = pkgPath.startsWith(domain + "/")
				? pkgPath.slice(domain.length + 1)
				: pkgPath;

			const response = await fetch("/_/api/eval", {
				method: "POST",
				headers: { "Content-Type": "application/json" },
				body: JSON.stringify({ pkg_path: relPath, expression }),
			});

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

			resultEl.textContent = result;
			resultEl.classList.toggle("u-color-danger", isError);
			addToHistory(expression, result, isError);
		} catch (err) {
			const errMsg = `Error: ${err instanceof Error ? err.message : String(err)}`;
			resultEl.textContent = errMsg;
			resultEl.classList.add("u-color-danger");
			addToHistory(expression, errMsg, true);
		}
	}

	function addToHistory(
		expression: string,
		result: string,
		isError: boolean,
	): void {
		history.push({ expression, result, isError });
		if (!historyListEl) return;

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
		rerunBtn.textContent = "\u21A9";
		rerunBtn.title = "Re-run";
		rerunBtn.addEventListener("click", () => {
			inputEl.value = expression;
			doEval(expression);
		});
		exprDiv.appendChild(rerunBtn);

		const resultPre = document.createElement("pre");
		resultPre.className = `b-eval-history-result${isError ? " u-color-danger" : ""}`;
		resultPre.textContent =
			result.length > 200 ? result.substring(0, 200) + "..." : result;

		entry.appendChild(exprDiv);
		entry.appendChild(resultPre);
		historyListEl.prepend(entry);
	}

	// Enter key submits
	inputEl.addEventListener("keydown", (e: KeyboardEvent) => {
		if (e.key === "ArrowUp" && history.length > 0) {
			e.preventDefault();
			inputEl.value = history[history.length - 1].expression;
		}
		if (e.key === "Enter") {
			e.preventDefault();
			const expr = inputEl.value.trim();
			if (expr) doEval(expr);
		}
	});

	// Form submit
	const form = element.querySelector("form");
	if (form) {
		form.addEventListener("submit", (e: Event) => {
			e.preventDefault();
			const expr = inputEl.value.trim();
			if (expr) doEval(expr);
		});
	}

	// Clear button
	element.querySelectorAll("[data-action]").forEach((el) => {
		const attr = el.getAttribute("data-action");
		if (attr === "click->eval#clearResult") {
			el.addEventListener("click", () => {
				resultEl.textContent = "// Enter an expression above";
				resultEl.classList.remove("u-color-danger");
			});
		}
	});

	// Quick call buttons
	element.querySelectorAll("[data-action]").forEach((el) => {
		const attr = el.getAttribute("data-action");
		if (attr !== "click->eval#quickCall") return;
		el.addEventListener("click", () => {
			const funcName = (el as HTMLElement).dataset.evalFuncNameParam || "";
			const funcSig = (el as HTMLElement).dataset.evalFuncSigParam || "";
			if (!funcName) return;

			const paramMatch = funcSig.match(/\(([^)]*)\)/);
			const params = paramMatch ? paramMatch[1] : "";
			if (params) {
				inputEl.value = `${funcName}(${params
					.split(",")
					.map((p: string) => {
						const parts = p.trim().split(/\s+/);
						const type = parts[parts.length - 1];
						return type === "string" ? '""' : "0";
					})
					.join(", ")})`;
				inputEl.focus();
				inputEl.setSelectionRange(
					funcName.length + 1,
					inputEl.value.length - 1,
				);
			} else {
				inputEl.value = `${funcName}()`;
				doEval(`${funcName}()`);
			}
		});
	});
}
