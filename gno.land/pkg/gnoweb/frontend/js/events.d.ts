// Ambient types shared across frontend scripts.

import type { GnoWallet } from "./wallet-discovery.js";

declare global {
	// CustomEvent payloads dispatched on document by controllers and consumed
	// by other controllers / analytics. Augmenting DocumentEventMap lets
	// addEventListener narrow the event type without runtime casts.
	interface DocumentEventMap {
		"mode:changed": CustomEvent<{ mode: string }>;
		"address:changed": CustomEvent<{ address: string }>;
		"theme:changed": CustomEvent<{ theme: string }>;
	}

	// GnoConnect in-page wallet discovery, dispatched on window rather than
	// document: an extension announces itself before any page script runs,
	// and window is the surface both sides can rely on.
	interface WindowEventMap {
		// Wallet → page: "I am here, and this is how to call me."
		"gno:registerWallet": CustomEvent<GnoWallet>;
		// Page → wallets: "announce yourselves" — for wallets that loaded
		// before the page was listening.
		"gno:requestWallet": Event;
	}

	// SimpleAnalytics globals injected by sa.gno.services/latest.js and the
	// sa-bootstrap loader. Declared here so analytics.ts and sa-bootstrap.ts
	// share the same shape.
	interface Window {
		sa_event?: (name: string, meta?: Record<string, string | boolean>) => void;
		sa_metadata?: Record<string, string>;
		// Path-overwriter read by latest.js via data-path-overwriter. Returns the
		// server-built pageview path that latest.js records in place of the raw
		// location.pathname.
		gnoSaPath?: () => string;
	}
}
