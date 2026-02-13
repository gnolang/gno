import { BaseController } from "./controller.js";

enum Theme {
	Light = "light",
	Dark = "dark",
}

const STORAGE_KEY = "theme";
const DEFAULT_THEME = Theme.Light;

function isTheme(value: string | null): value is Theme {
	return value === Theme.Light || value === Theme.Dark;
}

export class ThemeController extends BaseController {
	private get theme(): Theme {
		const data = document.documentElement.getAttribute("data-theme");
		return isTheme(data) ? data : DEFAULT_THEME;
	}

	protected connect(): void {
		const storedTheme = localStorage.getItem(STORAGE_KEY);
		if (isTheme(storedTheme))
			document.documentElement.setAttribute("data-theme", storedTheme);
		this.updateIcon(this.theme);
	}

	public toggle(): void {
		const nextTheme = this.theme === Theme.Dark ? Theme.Light : Theme.Dark;
		document.documentElement.setAttribute("data-theme", nextTheme);
		localStorage.setItem(STORAGE_KEY, nextTheme);
		this.updateIcon(nextTheme);
	}

	private updateIcon(theme: Theme): void {
		const sun = this.getTarget("sun");
		const moon = this.getTarget("moon");
		if (sun) sun.classList.toggle("u-hidden", theme === Theme.Light);
		if (moon) moon.classList.toggle("u-hidden", theme === Theme.Dark);
	}
}
