import initTooltip from './tooltip';

(() => {
  interface Module {
    selector: string;
    path: string;
  }

  const modules: Record<string, Module> = {
    copy: {
      selector: ".js-copy-btn",
      path: "/public/js/copy.js",
    },
    help: {
      selector: ".js-help-view",
      path: "/public/js/realmhelp.js",
    },
    searchBar: {
      selector: ".js-header-searchbar",
      path: "/public/js/searchbar.js",
    },
    tooltip: {
      selector: ".js-tooltip",
      path: "/public/js/tooltip.js",
    },
  };

  const loadModuleIfExists = async ({ selector, path }: Module): Promise<void> => {
    const element = document.querySelector(selector);
    if (element) {
      try {
        const module = await import(path);
        module.default();
      } catch (err) {
        console.error(`Error while loading script ${path}:`, err);
      }
    } else {
      console.warn(`Module not loaded: no element matches selector "${selector}"`);
    }
  };

  const initModules = async (): Promise<void> => {
    const promises = Object.values(modules).map((module) => loadModuleIfExists(module));
    await Promise.all(promises);
  };

  document.addEventListener("DOMContentLoaded", initModules);
})();
