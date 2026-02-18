import { BaseController } from "./controller.js";

enum Preference {
	System = "system",
	Light = "light",
	Dark = "dark",
}

const STORAGE_KEY = "theme";
const DEFAULT_PREFERENCE = Preference.System;
const CYCLE_ORDER = [
	Preference.System,
	Preference.Light,
	Preference.Dark,
] as const;

function isPreference(value: string | null): value is Preference {
	return (
		value === Preference.System ||
		value === Preference.Light ||
		value === Preference.Dark
	);
}

export class ThemeController extends BaseController {
	declare sun: HTMLElement | null;
	declare moon: HTMLElement | null;
	declare system: HTMLElement | null;
	declare preference: Preference;

	protected connect(): void {
		const { sun, moon, system } = this.initializeDOM({
			sun: this.getTarget("sun"),
			moon: this.getTarget("moon"),
			system: this.getTarget("system"),
		});
		this.sun = sun;
		this.moon = moon;
		this.system = system;

		try {
			const stored = localStorage.getItem(STORAGE_KEY);
			this.preference = isPreference(stored) ? stored : DEFAULT_PREFERENCE;
		} catch {
			this.preference = DEFAULT_PREFERENCE;
		}

		window
			.matchMedia("(prefers-color-scheme: dark)")
			.addEventListener("change", () => {
				if (this.preference === Preference.System) {
					this.applyTheme();
				}
			});

		this.applyTheme();
	}

	public toggle(): void {
		const currentIndex = CYCLE_ORDER.indexOf(this.preference);
		this.preference = CYCLE_ORDER[(currentIndex + 1) % CYCLE_ORDER.length];

		try {
			localStorage.setItem(STORAGE_KEY, this.preference);
		} catch {
			// localStorage unavailable — preference won't persist
		}

		this.applyTheme();
	}

	private applyTheme(): void {
		const theme =
			this.preference === Preference.System
				? window.matchMedia("(prefers-color-scheme: dark)").matches
					? "dark"
					: "light"
				: this.preference;

		document.documentElement.setAttribute("data-theme", theme);

		for (const [el, pref] of [
			[this.sun, Preference.Light],
			[this.moon, Preference.Dark],
			[this.system, Preference.System],
		] as [HTMLElement | null, Preference][]) {
			el?.classList.toggle("u-hidden", this.preference !== pref);
		}
	}
}
