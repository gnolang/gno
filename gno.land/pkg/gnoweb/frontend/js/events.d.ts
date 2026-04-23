// Ambient types shared across frontend scripts.

declare global {
	// CustomEvent payloads dispatched on document by controllers and consumed
	// by other controllers / analytics. Augmenting DocumentEventMap lets
	// addEventListener narrow the event type without runtime casts.
	interface DocumentEventMap {
		"mode:changed": CustomEvent<{ mode: string }>;
		"address:changed": CustomEvent<{ address: string }>;
		"theme:changed": CustomEvent<{ theme: string }>;
	}

	// SimpleAnalytics globals injected by sa.gno.services/latest.js and the
	// sa-bootstrap loader. Declared here so analytics.ts and sa-bootstrap.ts
	// share the same shape.
	interface Window {
		sa_event?: (name: string, meta?: Record<string, string | boolean>) => void;
		sa_metadata?: Record<string, string>;
	}
}

export {};
