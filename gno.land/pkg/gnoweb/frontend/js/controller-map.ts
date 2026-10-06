import { BaseController } from "./controller.js";

// Map inspector: names the tile under the pointer or the keyboard focus in a
// status line below the map. Mount on the map figure:
//   data-controller="map"
//   <p data-map-target="status" hidden>
// Each tile is an SVG <a class="b-map__tile"> whose <title> holds its full
// path and activity, the same text its native tooltip shows. Without this
// controller the status line stays hidden and the tooltip remains.
export class MapController extends BaseController {
	private declare status: HTMLElement | null;

	protected connect(): void {
		this.status = this.getTarget("status");
		if (!this.status) return;
		this.status.hidden = false;
		this.status.textContent = "Point at a tile to see its package.";

		const show = (event: Event): void => {
			const tile = (event.target as Element | null)?.closest(".b-map__tile");
			const title = tile?.querySelector("title")?.textContent;
			// Text only: the title is the server's escaped text, and goes in as
			// a text node, never as markup.
			if (title && this.status) this.status.textContent = title;
		};
		this.element.addEventListener("pointerover", show);
		this.element.addEventListener("focusin", show);
	}
}
