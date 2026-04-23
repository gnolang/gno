// CustomEvent payloads dispatched on document by controllers and consumed by
// other controllers / analytics. Augmenting DocumentEventMap lets
// addEventListener narrow the event type without runtime casts.

declare global {
	interface DocumentEventMap {
		"mode:changed": CustomEvent<{ mode: string }>;
		"address:changed": CustomEvent<{ address: string }>;
		"theme:changed": CustomEvent<{ theme: string }>;
	}
}

export {};
