class Theme {
    private DOM: {
      el: HTMLElement | null;
      toggle: HTMLElement | null;
    };
  
    private static SELECTORS = {
      toggle: ".js-theme-toggle",
      root: "html",
    };
  
    private static STORAGE_KEY = "gnoweb-theme-preference";
    private static DARK_CLASS = "dark";
  
    constructor() {
      this.DOM = {
        el: document.querySelector<HTMLElement>("main"),
        toggle: document.querySelector<HTMLElement>(Theme.SELECTORS.toggle),
      };
  
      if (this.DOM.el && this.DOM.toggle) {
        this.init();
      } else {
        console.warn("Theme: Required elements not found.");
      }
    }
  
    private init(): void {
      this.applyInitialTheme();
      this.bindEvents();
    }
  
    private bindEvents(): void {
      this.DOM.toggle?.addEventListener("click", this.toggleTheme.bind(this));
    }
  
    private applyInitialTheme(): void {
      const savedTheme = this.getSavedTheme();
      const systemPreference = window.matchMedia("(prefers-color-scheme: dark)");
  
      if (savedTheme) {
        this.setTheme(savedTheme);
      } else {
        this.setTheme(systemPreference.matches ? "dark" : "light");
      }
  
      // Listen for system theme changes
      systemPreference.addEventListener("change", (e) => {
        if (!this.getSavedTheme()) {
          this.setTheme(e.matches ? "dark" : "light");
        }
      });
    }
  
    private toggleTheme(): void {
      const root = document.documentElement;
      const currentTheme = root.classList.contains(Theme.DARK_CLASS) ? "dark" : "light";
      const newTheme = currentTheme === "dark" ? "light" : "dark";
      
      this.setTheme(newTheme);
    }
  
    private setTheme(theme: "light" | "dark"): void {
      const root = document.documentElement;
  
      if (theme === "dark") {
        root.classList.add(Theme.DARK_CLASS);
      } else {
        root.classList.remove(Theme.DARK_CLASS);
      }
  
      // Save theme preference
      localStorage.setItem(Theme.STORAGE_KEY, theme);
    }
  
    private getSavedTheme(): "light" | "dark" | null {
      const savedTheme = localStorage.getItem(Theme.STORAGE_KEY);
      return savedTheme === "dark" || savedTheme === "light" ? savedTheme : null;
    }
  }
  
  export default () => new Theme();