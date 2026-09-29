import { BaseController } from "./controller.js";

// The docs rail scrolls on its own, and on long sections the current page
// starts below its fold. Scroll the rail, not the page, so the entry and its
// outline are in view on load.
export class DocsNavController extends BaseController {
	protected connect(): void {
		const rail = this.element;
		const active = rail.querySelector<HTMLElement>('[aria-current="page"]');
		if (!active) return;

		const offset =
			active.getBoundingClientRect().top - rail.getBoundingClientRect().top;
		if (offset > rail.clientHeight / 2) {
			rail.scrollTop += offset - rail.clientHeight / 4;
		}
	}
}
