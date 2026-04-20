// Playground controller — standalone, does not rely on BaseController action wiring
interface PlaygroundFile {
	name: string;
	content: string;
}

export default function PlaygroundController(element: HTMLElement): void {
	let files: PlaygroundFile[] = [];
	let activeFile = 0;

	const codeEl = element.querySelector(
		'[data-playground-target="code"]',
	) as HTMLTextAreaElement;
	const outputEl = element.querySelector(
		'[data-playground-target="output"]',
	) as HTMLElement;
	const tabsEl = element.querySelector(
		'[data-playground-target="tabs"]',
	) as HTMLElement;

	if (!codeEl || !outputEl || !tabsEl) return;

	const domain =
		element.getAttribute("data-playground-domain-value") || "gno.land";

	// Parse initial code
	const initialCode = codeEl.value;
	if (initialCode.includes("// --- ") && initialCode.includes(" ---")) {
		const parts = initialCode.split(/^\/\/ --- (.+?) ---$/m);
		for (let i = 1; i < parts.length; i += 2) {
			const name = parts[i].trim();
			const content = (parts[i + 1] || "").trim();
			if (name) files.push({ name, content });
		}
		if (files.length === 0)
			files = [{ name: "main.gno", content: initialCode }];
		codeEl.value = files[0].content;
	} else {
		files = [{ name: "main.gno", content: initialCode }];
	}

	function renderTabs(): void {
		while (tabsEl.firstChild) tabsEl.removeChild(tabsEl.firstChild);
		files.forEach((f, i) => {
			const btn = document.createElement("button");
			btn.className = `b-playground-tab${i === activeFile ? " b-playground-tab--active" : ""}`;
			btn.textContent = f.name;
			btn.addEventListener("click", () => switchToFile(f.name));
			tabsEl.appendChild(btn);
		});
		const addBtn = document.createElement("button");
		addBtn.className = "b-playground-tab-add";
		addBtn.textContent = "+";
		addBtn.title = "Add file";
		addBtn.addEventListener("click", addFile);
		tabsEl.appendChild(addBtn);
	}

	function switchToFile(fileName: string): void {
		files[activeFile].content = codeEl.value;
		const idx = files.findIndex((f) => f.name === fileName);
		if (idx >= 0) {
			activeFile = idx;
			codeEl.value = files[idx].content;
			renderTabs();
		}
	}

	function addFile(): void {
		const name = prompt("File name (e.g. helper.gno):");
		if (!name || !name.endsWith(".gno")) return;
		if (files.some((f) => f.name === name)) return;
		files[activeFile].content = codeEl.value;
		files.push({ name, content: "package main\n" });
		activeFile = files.length - 1;
		codeEl.value = files[activeFile].content;
		renderTabs();
	}

	async function runCode(): Promise<void> {
		files[activeFile].content = codeEl.value;
		outputEl.textContent = "Running...";
		outputEl.classList.remove("u-color-danger");

		const code = codeEl.value;
		const pkgMatch = code.match(/^package\s+(\w+)/m);
		const pkgName = pkgMatch ? pkgMatch[1] : "main";

		if (code.includes("func Render(")) {
			try {
				const resp = await fetch("/_/api/eval", {
					method: "POST",
					headers: { "Content-Type": "application/json" },
					body: JSON.stringify({
						pkg_path: `${domain}/r/playground_preview`,
						expression: 'Render("")',
					}),
				});
				const result = await resp.json();
				if (result.error) {
					outputEl.textContent = `Error: ${result.error}`;
					outputEl.classList.add("u-color-danger");
				} else {
					outputEl.textContent = result.result;
				}
			} catch {
				outputEl.textContent = `Note: Server-side execution not available for scratch pad code.\n\nPackage: ${pkgName}\nFiles: ${files.map((f) => f.name).join(", ")}\n\nTo deploy and test:\n  gnokey maketx addpkg -pkgpath "${domain}/r/yourname/pkg" ...`;
			}
		} else {
			outputEl.textContent = `Package: ${pkgName}\nFiles: ${files.map((f) => f.name).join(", ")}\n\nTo run locally:\n  gno run ${files.map((f) => f.name).join(" ")}\n\nTo test:\n  gno test .`;
		}
	}

	function runTests(): void {
		outputEl.textContent =
			"Testing requires a running gno node.\n\nTo test locally:\n  gno test .";
	}

	function formatCode(): void {
		outputEl.textContent =
			"Formatting requires server-side gno fmt (coming soon).\n\nTo format locally:\n  gno fmt -w " +
			files[activeFile].name;
	}

	function shareCode(): void {
		files[activeFile].content = codeEl.value;
		const code =
			files.length === 1
				? files[0].content
				: files.map((f) => `// --- ${f.name} ---\n${f.content}`).join("\n\n");

		// Encode as base64; TextEncoder produces UTF-8 bytes, which are mapped to a
		// Latin-1 binary string so btoa() can handle non-ASCII chars, then the
		// resulting base64 is percent-encoded to be safe in a URL query parameter.
		const bytes = new TextEncoder().encode(code);
		const binary = Array.from(bytes, (b) => String.fromCharCode(b)).join("");
		const url = `${window.location.origin}/_/play?code=${encodeURIComponent(btoa(binary))}`;
		navigator.clipboard
			.writeText(url)
			.then(() => {
				outputEl.textContent = "Share URL copied to clipboard!";
			})
			.catch(() => {
				outputEl.textContent = `Share URL:\n${url}`;
			});
	}

	function clearOutput(): void {
		outputEl.textContent = "// Run code to see output here";
		outputEl.classList.remove("u-color-danger");
	}

	// Keyboard shortcuts
	codeEl.addEventListener("keydown", (e: KeyboardEvent) => {
		if (e.ctrlKey && e.key === "Enter") {
			e.preventDefault();
			runCode();
			return;
		}
		if (e.key === "Tab" && !e.shiftKey) {
			e.preventDefault();
			const start = codeEl.selectionStart;
			const end = codeEl.selectionEnd;
			codeEl.value =
				codeEl.value.substring(0, start) + "\t" + codeEl.value.substring(end);
			codeEl.selectionStart = codeEl.selectionEnd = start + 1;
		}
	});

	// Wire buttons by data-action attribute
	const actions: Record<string, () => void> = {
		runCode,
		runTests,
		formatCode,
		shareCode,
		clearOutput,
	};
	element.querySelectorAll("[data-action]").forEach((el) => {
		const attr = el.getAttribute("data-action");
		if (!attr) return;
		const match = attr.match(/^(\w+)->playground#(\w+)$/);
		if (!match) return;
		const [, evt, method] = match;
		const fn = actions[method];
		if (fn) el.addEventListener(evt, fn);
	});

	renderTabs();
}
