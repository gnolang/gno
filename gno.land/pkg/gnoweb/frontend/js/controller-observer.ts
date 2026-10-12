import { BaseController } from "./controller.js";

// Agnostic viewport observer. Mount on ANY element that contains `<a href="#id">`
// links pointing at on-page targets:
//   data-controller="observer"
//   data-observer-margin-value="0px 0px -75% 0px"  (optional IntersectionObserver rootMargin)
//
// As you scroll it sets `aria-current="true"` on the link whose target is the
// topmost one in view. A click PINS that link and holds it until the next real
// user scroll (wheel / touch / scroll keys) — never the programmatic jump's own
// `scroll` events. So clicking a section near the page bottom (which can't reach
// the top band) keeps its marker instead of snapping to the last section.
// No app-specific assumptions — reuse it for section navs, TOCs, step indicators.
export class ObserverController extends BaseController {
	private declare links: Map<string, HTMLAnchorElement>;
	private declare ids: string[];
	private declare visible: Set<string>;
	private declare observer: IntersectionObserver | null;
	private declare locked: boolean;
	private declare ac: AbortController;

	// Keys that scroll the page — pressing one releases a pinned click.
	private static readonly SCROLL_KEYS = new Set([
		"ArrowUp",
		"ArrowDown",
		"PageUp",
		"PageDown",
		"Home",
		"End",
		" ",
	]);

	protected connect(): void {
		this.links = new Map();
		this.ids = [];
		this.visible = new Set();
		this.observer = null;
		this.locked = false;
		this.ac = new AbortController();

		const targets: HTMLElement[] = [];
		const anchors = Array.from(
			this.element.querySelectorAll<HTMLAnchorElement>('a[href^="#"]'),
		);
		for (const anchor of anchors) {
			let id: string;
			try {
				id = decodeURIComponent(anchor.hash.slice(1));
			} catch {
				continue; // malformed percent-encoding in the fragment — skip
			}
			if (!id || this.links.has(id)) continue;
			const target = document.getElementById(id);
			if (!target) continue;
			this.links.set(id, anchor);
			this.ids.push(id);
			targets.push(target);
			anchor.addEventListener("click", () => this.pin(id), {
				signal: this.ac.signal,
			});
		}
		if (targets.length === 0) return;

		// Release a pinned click only on genuine user-driven scrolling. The
		// programmatic jump fires `scroll` events, which we deliberately ignore.
		const release = (): void => {
			if (!this.locked) return;
			this.locked = false;
			this.update();
		};
		const opts: AddEventListenerOptions = {
			passive: true,
			signal: this.ac.signal,
		};
		window.addEventListener("wheel", release, opts);
		window.addEventListener("touchmove", release, opts);
		window.addEventListener(
			"keydown",
			(e) => {
				if (ObserverController.SCROLL_KEYS.has(e.key)) release();
			},
			{ signal: this.ac.signal },
		);

		// Default band: a target is "in view" once it reaches the top quarter of the
		// viewport (works under a sticky bar without needing its exact height).
		const rootMargin = this.getValue("margin") || "0px 0px -75% 0px";
		this.observer = new IntersectionObserver(
			(entries) => this.onIntersect(entries),
			{ rootMargin, threshold: 0 },
		);
		for (const target of targets) this.observer.observe(target);
	}

	protected disconnect(): void {
		this.observer?.disconnect();
		this.ac?.abort();
	}

	// A click is an explicit choice: mark it active and hold until a user scroll.
	private pin(id: string): void {
		this.locked = true;
		this.setActive(id);
	}

	private onIntersect(entries: IntersectionObserverEntry[]): void {
		for (const entry of entries) {
			const id = entry.target.id;
			if (entry.isIntersecting) this.visible.add(id);
			else this.visible.delete(id);
		}
		if (!this.locked) this.update();
	}

	// Topmost in-view target wins. Exception: once scrolled to the bottom the last
	// target can never reach the top band, so force it active there.
	private update(): void {
		let activeId: string | null = null;
		const atBottom =
			window.innerHeight + window.scrollY >=
			document.documentElement.scrollHeight - 2;
		if (atBottom && this.ids.length > 0) {
			activeId = this.ids[this.ids.length - 1];
		} else {
			for (const id of this.ids) {
				if (this.visible.has(id)) {
					activeId = id;
					break;
				}
			}
		}
		this.setActive(activeId);
	}

	private setActive(activeId: string | null): void {
		for (const [id, anchor] of this.links) {
			if (id === activeId) {
				anchor.setAttribute("aria-current", "true");
			} else {
				anchor.removeAttribute("aria-current");
			}
		}
	}
}
