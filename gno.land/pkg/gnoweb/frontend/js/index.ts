import { toCamelCase, toKebabCase } from "./utils.js";

declare const __DEV__: boolean;

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

		// normalize the controller name
		const kebab = toKebabCase(controllerName);
		if (!/^[a-z0-9-]+$/.test(kebab)) {
			console.error(`❌ Invalid controller name: ${controllerName}`);
			return;
		}
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

		if (__DEV__) console.log(`✅ Loaded: ${controllerName} (${path})`);
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

	document.addEventListener("DOMContentLoaded", initModules);
})();
