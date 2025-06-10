import { debounce } from './utils';

type ListItem = {
  element: HTMLElement;
  title: string;
  type: string;
};

class List {
  private DOM: {
    el: Element | null;
    range: HTMLElement | null;
    searchBar: HTMLInputElement | null;
    orderOption: HTMLElement | null;
    packagesCount: HTMLElement | null;
    realmsCount: HTMLElement | null;
    pureCount: HTMLElement | null;
  };

  private static SELECTORS = {
    range: '.js-list-range',
    searchBar: '.js-list-searchbar',
    orderOption: '.js-list-order-filter',
    itemTitle: '.js-list-range-title',
    packagesCount: '.js-list-packages-count',
    realmsCount: '.js-list-realms-count',
    pureCount: '.js-list-pure-count',
  };

  private items: Array<ListItem> = [];

  // Cache
  private currentFilter: string = '';
  private sortedItems: Array<ListItem> = [];

  constructor(el: Element | null) {
    if (!el) {
      console.warn('No element provided');
      return;
    }

    this.DOM = {
      el,
      range: el.querySelector<HTMLElement>(List.SELECTORS.range),
      searchBar: el.querySelector<HTMLInputElement>(List.SELECTORS.searchBar),
      orderOption: el.querySelector<HTMLElement>(List.SELECTORS.orderOption),
      packagesCount: el.querySelector<HTMLElement>(
        List.SELECTORS.packagesCount
      ),
      realmsCount: el.querySelector<HTMLElement>(List.SELECTORS.realmsCount),
      pureCount: el.querySelector<HTMLElement>(List.SELECTORS.pureCount),
    };

    if (!this.DOM.range) {
      console.warn('Range element not found');
      return;
    }

    // Store initial items with their titles and types
    this.items = Array.from(this.DOM.range.children).map(element => {
      const titleElement = (element as HTMLElement).querySelector(
        List.SELECTORS.itemTitle
      );
      const type = (element as HTMLElement).dataset.type || '';

      return {
        element: element as HTMLElement,
        title: (titleElement?.textContent || '').toLowerCase(),
        type,
      };
    });

    this.sortedItems = [...this.items];
    this.bindEvents();
  }

  private bindEvents(): void {
    const { searchBar, orderOption } = this.DOM;

    // Handle order change
    orderOption?.addEventListener('change', e => {
      const target = e.target as HTMLInputElement;

      // event delegation
      if (target.matches('input[name="order-mode"]')) {
        this.sortItems();
        this.updateDOM();
      }
    });

    searchBar?.addEventListener('input', e => {
      const target = e.target as HTMLInputElement;
      this.debouncedSearch(target.value);
    });
  }

  // Debounce search to avoid too many calls
  private debouncedSearch = debounce((value: string) => {
    this.currentFilter = value.toLowerCase();
    this.updateDOM();
  }, 150);

  // Sort items
  private sortItems(): void {
    this.sortedItems.reverse();
  }

  // Filter items
  private filterItems(): Array<ListItem> {
    if (!this.currentFilter) return this.sortedItems;

    return this.sortedItems.filter(item =>
      item.title.includes(this.currentFilter)
    );
  }

  // Update counts based on filtered items
  private updateCounts(items: Array<ListItem>): void {
    const { packagesCount, realmsCount, pureCount } = this.DOM;

    const counts = items.reduce(
      (acc, item) => {
        acc[item.type] = (acc[item.type] || 0) + 1;
        return acc;
      },
      {} as Record<string, number>
    );

    const realmCountValue = counts['1'] || 0;
    const pureCountValue = counts['2'] || 0;
    const totalPackages = realmCountValue + pureCountValue;

    if (packagesCount) {
      packagesCount.textContent = totalPackages.toString();
    }
    if (realmsCount) {
      realmsCount.textContent = realmCountValue.toString();
    }
    if (pureCount) {
      pureCount.textContent = pureCountValue.toString();
    }
  }

  // Update DOM with items
  private updateDOM(): void {
    const { range } = this.DOM;
    if (!range) return;

    const filteredItems = this.filterItems();

    // Create a document fragment for batch DOM updates
    const fragment = document.createDocumentFragment();
    filteredItems.forEach(item => fragment.appendChild(item.element));

    // Clear and update range with a single reflow
    range.textContent = '';
    range.appendChild(fragment);

    // Update counts based on filtered items
    this.updateCounts(filteredItems);
  }
}

export default (el: HTMLElement | null) => new List(el);
