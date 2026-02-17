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
	private sun: HTMLElement | null = null;
	private moon: HTMLElement | null = null;

	private get theme(): Theme {
		const data = document.documentElement.getAttribute("data-theme");
		return isTheme(data) ? data : DEFAULT_THEME;
	}

	protected connect(): void {
		const { sun, moon } = this.initializeDOM({
			sun: this.getTarget("sun"),
			moon: this.getTarget("moon"),
		});
		this.sun = sun;
		this.moon = moon;

		try {
			const storedTheme = localStorage.getItem(STORAGE_KEY);
			if (isTheme(storedTheme)) {
				document.documentElement.setAttribute("data-theme", storedTheme);
			}
		} catch {
			// localStorage unavailable (private browsing, etc.)
		}

		this.syncUI(this.theme);
	}

	public toggle(): void {
		const nextTheme = this.theme === Theme.Dark ? Theme.Light : Theme.Dark;
		document.documentElement.setAttribute("data-theme", nextTheme);

		try {
			localStorage.setItem(STORAGE_KEY, nextTheme);
		} catch {
			// localStorage unavailable — theme changes but won't persist
		}

		this.syncUI(nextTheme);
	}

	private syncUI(theme: Theme): void {
		const isDark = theme === Theme.Dark;
		if (this.sun) this.sun.classList.toggle("u-hidden", !isDark);
		if (this.moon) this.moon.classList.toggle("u-hidden", isDark);
		this.element.setAttribute("aria-pressed", String(isDark));
	}
}
