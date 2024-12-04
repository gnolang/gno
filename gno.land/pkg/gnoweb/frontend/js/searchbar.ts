(() => {
  class SearchBar {
    private DOM: {
      el: HTMLElement | null;
      inputSearch: HTMLInputElement | null;
      breadcrumb: HTMLElement | null;
    };

    private baseUrl: string;

    constructor() {
      this.DOM = {
        el: document.querySelector("#header-searchbar"),
        inputSearch: null,
        breadcrumb: null,
      };

      this.baseUrl = window.location.origin;

      if (this.DOM.el) this.init();
    }

    private init() {
      const { el } = this.DOM;

      this.DOM.inputSearch = el?.querySelector<HTMLInputElement>("[data-role='header-input-search']") ?? null;
      this.DOM.breadcrumb = el?.querySelector<HTMLInputElement>("[data-role='header-breadcrumb-search']") ?? null;

      this.bindEvents();
    }

    private bindEvents() {
      this.DOM.el?.addEventListener("submit", (e) => {
        e.preventDefault();
        this.searchUrl();
      });
    }

    public searchUrl() {
      const input = this.DOM.inputSearch?.value.trim();

      if (input) {
        let url = input;

        if (!/^https?:\/\//i.test(url)) {
          url = `${this.baseUrl}${url.startsWith("/") ? "" : "/"}${url}`;
        }

        try {
          window.location.href = new URL(url).href;
        } catch (error) {
          console.error("Invalid URL. Please enter a valid URL starting with http:// or https://");
        }
      } else {
        console.error("Please enter a URL to search.");
      }
    }
  }

  document.addEventListener("DOMContentLoaded", () => new SearchBar());
})();
