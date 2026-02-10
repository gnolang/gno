import {
	debounce,
	escapeShellSpecialChars,
	findAllInclusive,
	findFirstInclusive,
	toCamelCase,
	toKebabCase,
} from "./utils.js";

export abstract class BaseController {
	protected element: HTMLElement;
	protected initialized = false;
	protected DOM: Record<string, HTMLElement | HTMLElement[] | null> = {};
	protected controllerName: string;
	protected controllerKebabName: string;

	constructor(element: HTMLElement) {
		this.element = element;
		this.controllerName = toCamelCase(this.getControllerName());
		this.controllerKebabName = toKebabCase(this.controllerName);
		this.init();
	}

	// connect and disconnect the controller
	protected abstract connect(): void;
	protected disconnect?(): void;

	protected init(): void {
		if (!this.initialized) {
			this.connect();
			this.setupActions();
			this.initialized = true;
		}
	}

	protected initializeDOM<
		U extends Record<string, HTMLElement | HTMLElement[] | null>,
	>(domStructure: U): U {
		this.DOM = { el: this.element, ...domStructure } as Record<
			string,
			HTMLElement | HTMLElement[] | null
		>;
		return domStructure;
	}

	// Helper method to get DOM element with automatic typing
	protected getDOMElement(key: string): HTMLElement | null {
		const element = Array.isArray(this.DOM[key])
			? this.DOM[key][0]
			: this.DOM[key];
		return element;
	}

	// Helper method to get DOM element as array (for forEach usage)
	protected getDOMArray(key: string): HTMLElement[] {
		const element = this.DOM[key];
		if (!element) return [];
		return Array.isArray(element) ? element : [element];
	}

	// get the element with the same attribute
	// eg. getTarget("source-code") -> <div data-[CONTROLLER]-target="source-code">
	protected getTarget(
		attribute: string,
		scope?: HTMLElement,
	): HTMLElement | null {
		const searchScope = scope || this.element;
		const selector = `[data-${this.controllerKebabName}-target="${attribute}"]`;
		const result = findFirstInclusive(searchScope, selector);
		return result;
	}

	// get all elements with the same attribute
	// eg. getTargets("source-code") -> [<div data-[CONTROLLER]-target="source-code">, <div data-[CONTROLLER]-target="source-code">]
	protected getTargets(attribute: string, scope?: HTMLElement): HTMLElement[] {
		const searchScope = scope || this.element;
		return findAllInclusive(
			searchScope,
			`[data-${this.controllerKebabName}-target="${attribute}"]`,
		);
	}

	// get target by attribute globally
	// eg. getGlobalTarget("source-code") -> <div data-[CONTROLLER]-target="source-code">
	protected getGlobalTarget(attribute: string): HTMLElement | null {
		return document.querySelector(
			`[data-${this.controllerKebabName}-target="${attribute}"]`,
		);
	}

	// add event listeners to the element (outter of the controller element)
	// eg. setupActions() -> <button data-action="click->copy#copy">
	private setupActions(): void {
		const actions = findAllInclusive(this.element, "[data-action]");
		actions.forEach((element) => {
			const actionAttr = element.getAttribute("data-action");
			if (actionAttr) {
				const [event, action] = actionAttr.split("->");
				const [controllerName, methodName] = action.split("#");

				if (controllerName === this.controllerKebabName) {
					const method = (this as Record<string, unknown>)[methodName] as
						| ((event: Event & { params?: Record<string, unknown> }) => void)
						| undefined;
					if (typeof method === "function") {
						element.addEventListener(event, (event) => {
							// Extract params from data attributes
							const params = this.extractActionParams(element);
							// Add params to event object
							(event as Event & { params?: Record<string, unknown> }).params =
								params;
							method.call(this, event);
						});
					}
				}
			}
		});
	}

	// Extract action parameters from data attributes
	// eg. data-item-id-param="12345" -> { id: 12345 }
	private extractActionParams(element: HTMLElement): Record<string, unknown> {
		const params: Record<string, unknown> = {};
		for (const [key, value] of Object.entries(element.dataset)) {
			if (key.startsWith(this.controllerName) && key.endsWith("Param")) {
				// arg to be used in CamelCase (eg. actionFunctionSendParam -> send)
				const paramName = toCamelCase(
					key.slice(this.controllerName.length, -"Param".length),
				);
				params[paramName] = this.parseParamValue(value);
			}
		}

		return params;
	}

	// Parse parameter value with automatic type casting
	// eg. data-controller-param="true" -> true
	// eg. data-controller-param="123" -> 123
	// eg. data-controller-param="{\"foo\":\"bar\"}" -> { foo: "bar" }
	// eg. data-controller-param="foo" -> "foo"
	private parseParamValue(
		value: string | undefined,
	): string | number | boolean | object | null | undefined {
		if (value == null || value.trim() === "") {
			return "";
		}
		const s = value.trim();

		// Boolean (case insensitive)
		if (/^(true|false)$/i.test(s)) {
			return s.toLowerCase() === "true";
		}

		// Number (integers, floats, exponents)
		if (/^[+-]?\d+(\.\d+)?([eE][+-]?\d+)?$/.test(s)) {
			const n = Number(s);
			if (!Number.isNaN(n) && Number.isFinite(n)) {
				return n;
			}
		}

		// null / undefined literals
		if (/^(null|undefined)$/i.test(s)) {
			return s.toLowerCase() === "null" ? null : undefined;
		}

		// JSON (object, array, JSON string)
		if (/^[{[]/.test(s) || /^".*"$/.test(s)) {
			try {
				return JSON.parse(s);
			} catch {
				// if JSON is invalid, return string
			}
		}

		// Fallback : string
		return value;
	}

	// get the value of the attribute
	// eg. getValue("remote") -> <div data-[CONTROLLER]-remote-value="source-code">
	protected getValue(name: string, scope?: HTMLElement): string {
		const attribute = `data-${this.controllerKebabName}-${name}-value`;
		const searchScope = scope || this.element;
		const element = findFirstInclusive(searchScope, `[${attribute}]`);
		return element ? element.getAttribute(attribute) || "" : "";
	}

	// set the value of the attribute
	// eg. setValue("remote", "foo") -> <div data-[CONTROLLER]-remote-value="foo">
	protected setValue(name: string, value: string, scope?: HTMLElement): void {
		const attribute = `data-${this.controllerKebabName}-${name}-value`;
		const targetElement = scope || this.element;
		targetElement.setAttribute(attribute, value);
	}

	// check if the attribute has a value
	// eg. hasValue("remote") -> true
	protected hasValue(name: string, scope?: HTMLElement): boolean {
		const attribute = `data-${this.controllerKebabName}-${name}-value`;
		const searchScope = scope || this.element;
		return findFirstInclusive(searchScope, `[${attribute}]`) !== null;
	}

	// Events
	// add an event listener
	// eg. on("click", () => {})
	protected on(event: string, callback: EventListener): void {
		document.addEventListener(event, callback);
	}

	// dispatch an event
	// eg. dispatch("click", { detail: { foo: "bar" } })
	protected dispatch(
		eventName: string,
		detail: Record<string, unknown> = {},
	): void {
		const event = new CustomEvent(eventName, { detail, bubbles: true });
		document.dispatchEvent(event);
	}

	// get the controller name
	// eg. getControllerName() -> "Copy" (no Controller suffix) PascalCase
	private getControllerName(): string {
		// Use data-controller attribute for minified code
		const controllerAttr = this.element.getAttribute("data-controller");
		if (controllerAttr) {
			return controllerAttr;
		}

		// Fallback to constructor name (for non-minified code)
		const className = this.constructor.name;
		return className
			.replace(/^_/, "") // remove leading underscore
			.replace(/Controller$/, ""); // remove Controller suffix
	}
}

// Re-export utilities for convenience
export {
	debounce,
	escapeShellSpecialChars,
	findAllInclusive,
	findFirstInclusive,
	toCamelCase,
	toKebabCase,
};
