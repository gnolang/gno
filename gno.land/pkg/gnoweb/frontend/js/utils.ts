export function debounce<T extends (...args: unknown[]) => void>(
	func: T,
	delay = 250,
): (...args: Parameters<T>) => void {
	let timeoutId: ReturnType<typeof setTimeout> | undefined;
	return function (this: unknown, ...args: Parameters<T>) {
		if (timeoutId !== undefined) clearTimeout(timeoutId);
		timeoutId = setTimeout(() => func.apply(this, args), delay);
	};
}

export function escapeShellSpecialChars(arg: string): string {
	return arg.replace(/([$`"\\!|&;<>*?{}()])/g, "\\$1");
}

export function toKebabCase(str: string): string {
	return str
		.replace(/([a-z])([A-Z])/g, "$1-$2") // replace camelCase with dash-separated words
		.toLowerCase(); // convert to lowercase
}

export function toCamelCase(str: string): string {
	str = str.replace(/-([a-z])/g, (_, letter) => letter.toUpperCase());
	return str.charAt(0).toLowerCase() + str.slice(1);
}

export function findFirstInclusive(
	root: HTMLElement,
	selector: string,
): HTMLElement | null {
	return root.matches(selector) ? root : root.querySelector(selector);
}

export function findAllInclusive(
	root: HTMLElement,
	selector: string,
): HTMLElement[] {
	const result: HTMLElement[] = [];
	if (root.matches(selector)) result.push(root);
	result.push(
		...(Array.from(root.querySelectorAll(selector)) as HTMLElement[]),
	);
	return result;
}
