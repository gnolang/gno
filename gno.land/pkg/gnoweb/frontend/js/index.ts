import { toCamelCase, toKebabCase } from "./utils.js";

declare const process: { env: { NODE_ENV: string } };

(() => {
	// TODO: Make CONTROLLER_PATH build-safe (BASE_URL, CDN, hashing, etc.)
	const CONTROLLER_PATH = "/public/js/controller-";
	const modulePromises = new Map<string, Promise<Record<string, unknown>>>();

	// load one controller for a provided set of elements (no re-query)
	const loadController = async (
		controllerName: string,
		elements: HTMLElement[],
	): Promise<void> => {
		if (elements.length === 0) return;

		// normalize the controller name to kebab-case
		const kebab = toKebabCase(controllerName);
		if (!/^[a-z0-9-]+$/.test(kebab)) {
			console.error(`❌ Invalid controller name: ${controllerName}`);
			return;
		}

		// normalize the controller name to camelCase
		const camel = toCamelCase(controllerName);
		const pascal = camel.charAt(0).toUpperCase() + camel.slice(1);

		// Only kebab-case file naming, prefixed with "controller-"
		const path = `${CONTROLLER_PATH}${kebab}.js`;

		// import the controller module with promise cache (dedupe concurrent imports)
		let modulePromise = modulePromises.get(path);
		if (!modulePromise) {
			modulePromise = import(path) as Promise<Record<string, unknown>>;
			modulePromises.set(path, modulePromise);
		}
		let module: Record<string, unknown> | undefined;
		try {
			module = await modulePromise;
		} catch (err) {
			modulePromises.delete(path);
			console.error(`❌ Failed to load ${path}:`, err);
			return;
		}

		// Resolve controller factory (class or function)
		let controller: ((element: HTMLElement) => void) | undefined;
		const def = (module as { default?: unknown }).default;
		if (typeof def === "function") {
			// Static class detection to avoid try/catch masking runtime errors
			const isClass = /^\s*class\b/.test(Function.prototype.toString.call(def));
			controller = isClass
				? (el: HTMLElement) => new (def as new (el: HTMLElement) => unknown)(el)
				: (def as (el: HTMLElement) => void);
		} else {
			const ControllerClass = (module as Record<string, unknown>)[
				`${pascal}Controller`
			] as (new (el: HTMLElement) => unknown) | undefined;
			if (ControllerClass)
				controller = (el: HTMLElement) => new ControllerClass(el);
		}

		if (typeof controller !== "function") {
			console.error(
				`❌ Invalid controller export for ${controllerName}. Expected default or named class "${pascal}Controller"`,
			);
			return;
		}

		// Idempotent init per element (safe if re-run)
		const flag = `data-controller-initialized-${kebab}`;
		const targets = elements.filter((el) => !el.hasAttribute(flag));
		if (targets.length === 0) return;

		targets.forEach((el) => {
			try {
				controller(el);
				el.setAttribute(flag, "1");
			} catch (err) {
				console.error(
					`❌ Controller runtime error for ${controllerName}:`,
					err,
					el,
				);
			}
		});

		if (process.env.NODE_ENV === "development")
			console.log(`✅ js - Loaded: ${controllerName} (${path})`);
	};

	// Collect controllers once and pass elements directly
	const collectControllers = (root: ParentNode): Map<string, HTMLElement[]> => {
		const map = new Map<string, HTMLElement[]>();
		root.querySelectorAll("[data-controller]").forEach((el) => {
			const value = el.getAttribute("data-controller");
			if (!value) return;
			value
				.split(/\s+/)
				.filter(Boolean)
				.forEach((name) => {
					const arr = map.get(name) || [];
					arr.push(el as HTMLElement);
					map.set(name, arr);
				});
		});
		return map;
	};

	// Init modules and start observer after DOMContentLoaded
	const initModules = async (): Promise<void> => {
		// get all used controllers and their elements
		const controllers = collectControllers(document);
		if (controllers.size === 0) return;

		await Promise.all(
			Array.from(controllers.entries()).map(([name, els]) =>
				loadController(name, els),
			),
		);

		// dispatch an event to notify other controllers that they are ready
		document.dispatchEvent(
			new CustomEvent("controllers:ready", {
				detail: { names: Array.from(controllers.keys()) },
			}),
		);
	};

	// Start observer to collect controllers in the DOM
	const startObserver = (): void => {
		const queue = new Map<string, Set<HTMLElement>>();
		let scheduled = false;

		const flush = () => {
			scheduled = false;
			if (queue.size === 0) return;
			const tasks: Promise<void>[] = [];
			for (const [name, set] of queue)
				tasks.push(loadController(name, [...set]));
			queue.clear();
			Promise.all(tasks).catch((e) =>
				console.error("Observer batch error:", e),
			);
		};

		const schedule = () => {
			if (!scheduled) {
				scheduled = true;
				queueMicrotask(flush); // setTimeout(flush, 0)
			}
		};

		// Collect controllers in a root element in order to avoid re-querying the DOM
		const handleRoot = (root: ParentNode) => {
			// short-circuit if no controllers are found
			if (!(root as Element).querySelector?.("[data-controller]")) return;

			// collect the controllers in the root element
			const map = collectControllers(root);
			for (const [name, els] of map) {
				const kebab = toKebabCase(name);
				const flag = `data-controller-initialized-${kebab}`;
				const filtered = els.filter((el) => !el.hasAttribute(flag));
				if (filtered.length === 0) continue;

				const set = queue.get(name) ?? new Set<HTMLElement>();
				filtered.forEach((el) => set.add(el));
				queue.set(name, set);
			}
			if (map.size) schedule();
		};

		// create a mutation observer to observe the document for new controllers
		const observer = new MutationObserver((muts) => {
			for (const m of muts) {
				if (m.type === "childList") {
					m.addedNodes.forEach((n) => {
						if (n.nodeType === 1) handleRoot(n as ParentNode);
					});
				} else if (m.type === "attributes" && m.target instanceof HTMLElement) {
					handleRoot(m.target);
				}
			}
		});

		// observe the document for new controllers
		observer.observe(document.documentElement, {
			childList: true,
			subtree: true,
			attributes: true,
			attributeFilter: ["data-controller"],
		});
	};

	// Init modules and start observer after DOMContentLoaded
	document.addEventListener("DOMContentLoaded", async () => {
		await initModules();
		startObserver();
	});
})();
