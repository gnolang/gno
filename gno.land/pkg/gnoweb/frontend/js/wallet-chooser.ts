// The wallet chooser dialog, shared by the Execute path and the header
// identity control. The markup lives in layouts/header.html, so it is present
// on every page; this module owns everything that happens inside it.

import type { GnoWallet } from "./wallet-discovery.js";

// Entry shape of the server-embedded registry (components/wallets.json).
export interface RegistryWallet {
	name: string;
	id: string;
	icon: string;
	rdns: string;
	kind: "extension" | "app";
	scheme: string; // bare, e.g. "land.gno.gnokey"; callers append "://sendtx?..."
	global: string; // window key of a legacy extension, "" otherwise
	platforms: string[];
	install: { label: string; url: string }[];
}

// The three kinds of candidate the chooser merges: a wallet that announced
// itself in the page (extension, called directly), a registry entry (external
// app, reached by launch link), and a legacy extension (reached by handing it
// back the submit it expects to intercept).
export type Candidate =
	| { kind: "in-page"; wallet: GnoWallet }
	| { kind: "external"; wallet: RegistryWallet }
	| { kind: "legacy"; wallet: { name: string; icon: string; rdns: string } };

// Parsed once per page load and shared by every caller.
let registryCache: RegistryWallet[] | undefined;

// A missing/malformed registry disables external-wallet routing rather than
// dead-ending the page.
export function loadRegistry(): RegistryWallet[] {
	if (registryCache) return registryCache;
	registryCache = [];
	const script = document.querySelector(
		'[data-wallet-launch-target="wallet-registry"]',
	);
	if (!script?.textContent) return registryCache;
	try {
		const parsed = JSON.parse(script.textContent);
		if (Array.isArray(parsed)) registryCache = parsed as RegistryWallet[];
	} catch {
		console.warn("wallet-chooser: invalid wallet registry JSON");
	}
	return registryCache;
}

// A legacy extension owns the submit by intercepting it, without announcing
// itself. One that has announced is skipped: it is already in the list,
// reachable by being called rather than by being handed the event.
export function legacyCandidate(announced: GnoWallet[]): Candidate | null {
	const w = window as unknown as Record<string, unknown>;
	const spoken = new Set(announced.map((a) => a.info?.rdns));
	const found = loadRegistry().find(
		(entry) =>
			entry.global !== "" &&
			Boolean(w[entry.global]) &&
			!spoken.has(entry.rdns),
	);
	return found
		? {
				kind: "legacy",
				wallet: { name: found.name, icon: "", rdns: found.rdns },
			}
		: null;
}

function el(target: string): HTMLElement | null {
	return document.querySelector(`[data-wallet-launch-target="${target}"]`);
}

function render(
	list: HTMLElement,
	candidates: Candidate[],
	pick: (c: Candidate) => void,
): void {
	list.textContent = "";
	candidates.forEach((candidate) => {
		const { name, icon } =
			candidate.kind === "in-page" ? candidate.wallet.info : candidate.wallet;

		const li = document.createElement("li");
		const btn = document.createElement("button");
		btn.type = "button";
		btn.className = "b-wallet-chooser__item";
		if (icon) {
			const img = document.createElement("img");
			img.src = icon;
			img.alt = "";
			img.className = "b-wallet-chooser__icon";
			btn.appendChild(img);
		}
		// textContent, never innerHTML: an announced name is untrusted,
		// anything in the page can dispatch an announcement.
		const label = document.createElement("span");
		label.textContent = name;
		btn.appendChild(label);
		// Which transport signs matters to the user: an extension signs here,
		// an app takes them out of the browser and back.
		const kind = document.createElement("span");
		kind.className = "b-wallet-chooser__kind";
		kind.textContent = candidate.kind === "external" ? "App" : "Extension";
		btn.appendChild(kind);

		btn.addEventListener("click", () => pick(candidate));
		li.appendChild(btn);
		list.appendChild(li);
	});
}

// showModal() centers the dialog in the layout viewport, but a zoomed mobile
// page only shows part of it, so the dialog can land half off-screen. Shift it
// to the center of the visual viewport instead, tracking zoom/scroll.
function centerInVisualViewport(dialog: HTMLDialogElement): void {
	const vv = window.visualViewport;
	if (!vv) return;

	const center = () => {
		const root = document.documentElement;
		dialog.style.maxWidth = `${vv.width * 0.9}px`;
		dialog.style.maxHeight = `${vv.height * 0.9}px`;
		const dx = vv.offsetLeft + (vv.width - root.clientWidth) / 2;
		const dy = vv.offsetTop + (vv.height - root.clientHeight) / 2;
		dialog.style.transform = `translate(${dx}px, ${dy}px)`;
	};
	center();
	vv.addEventListener("resize", center);
	vv.addEventListener("scroll", center);
	dialog.addEventListener(
		"close",
		() => {
			vv.removeEventListener("resize", center);
			vv.removeEventListener("scroll", center);
			dialog.style.transform = "";
			dialog.style.maxWidth = "";
			dialog.style.maxHeight = "";
		},
		{ once: true },
	);
}

function show(dialog: HTMLDialogElement): void {
	if (typeof dialog.showModal === "function") {
		dialog.showModal();
		centerInVisualViewport(dialog);
	} else {
		dialog.setAttribute("open", "");
	}
}

// Open the chooser and resolve with the picked candidate, or null when the
// user cancels, closes it, or takes the browser fallback. `refresh` is called
// again whenever a wallet announces late, so a slow extension still appears.
export function openChooser(opts: {
	refresh: () => Candidate[];
	title?: string;
	browser?: { label: string; onPick: () => void } | null;
	onOpen?: () => void;
}): Promise<Candidate | null> {
	const dialog = el("chooser") as HTMLDialogElement | null;
	const list = el("chooser-list");
	if (!dialog || !list) {
		// No dialog in the page — fail open on the first candidate.
		return Promise.resolve(opts.refresh()[0] ?? null);
	}

	const title = el("chooser-title");
	if (title) title.textContent = opts.title ?? "Open with a wallet";

	return new Promise<Candidate | null>((resolve) => {
		let settled = false;
		const finish = (value: Candidate | null) => {
			if (settled) return;
			settled = true;
			dialog.close();
			resolve(value);
		};

		render(list, opts.refresh(), finish);

		// Assignment (not addEventListener) so reopening does not stack handlers.
		const browser = el("chooser-browser") as HTMLButtonElement | null;
		if (browser) {
			browser.hidden = !opts.browser;
			if (opts.browser) {
				browser.textContent = opts.browser.label;
				browser.onclick = () => {
					const run = opts.browser?.onPick;
					finish(null);
					run?.();
				};
			} else {
				browser.onclick = null;
			}
		}
		const cancel = el("chooser-cancel") as HTMLButtonElement | null;
		if (cancel) cancel.onclick = () => finish(null);

		// Esc, backdrop dismissal, anything else that closes the dialog.
		dialog.addEventListener("close", () => finish(null), { once: true });

		opts.onOpen?.();
		show(dialog);

		// Re-render on late announcements while the dialog is open.
		const rerender = () => {
			if (!settled) render(list, opts.refresh(), finish);
		};
		window.addEventListener("gno:registerWallet", rerender);
		dialog.addEventListener(
			"close",
			() => window.removeEventListener("gno:registerWallet", rerender),
			{ once: true },
		);
	});
}

// Reopen the chooser to say why a wallet did nothing. A request refused
// without ever reaching a wallet screen otherwise leaves the user staring at
// an unchanged page. `code` is the standard's enumerated reason, rendered as
// text — nothing coming back from a wallet is trusted as markup. Resolves
// when the dialog closes, so a caller can navigate only once the user has
// actually seen the message; resolves immediately if there is no dialog to
// show, so a page without the chooser markup is never stranded.
export function reportChooserError(
	walletName: string,
	err: unknown,
): Promise<void> {
	const dialog = el("chooser") as HTMLDialogElement | null;
	const list = el("chooser-list");
	if (!dialog || !list) return Promise.resolve();

	const code = (err as { code?: unknown })?.code;
	const reason = typeof code === "string" ? ` (${code})` : "";

	list.textContent = "";
	const li = document.createElement("li");
	li.className = "b-wallet-chooser__error";
	li.textContent = `${walletName} could not take this transaction${reason}. The gnokey command below works without a wallet.`;
	list.appendChild(li);

	return new Promise<void>((resolve) => {
		dialog.addEventListener("close", () => resolve(), { once: true });
		if (!dialog.open) show(dialog);
	});
}
