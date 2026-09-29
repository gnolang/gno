// The wallet chooser dialog, shared by the Execute path and the header
// identity control. The markup lives in layouts/header.html, so it is present
// on every page; this module owns everything that happens inside it.

import { readMeta } from "./chain.js";
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
	browser?: { label: string; onPick?: () => void } | null;
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
	// Undo a previous error report: the picker is the default state.
	dialog.classList.remove("b-wallet-chooser--error");
	const install = el("chooser-install");
	if (install) install.hidden = false;
	const cancelLabel = el("chooser-cancel");
	if (cancelLabel) cancelLabel.textContent = "Cancel";

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

// What a wallet's failure means for the user, per the standard's codes.
// Only a known code is ever shown; anything else gets the generic sentence.
const ERROR_MESSAGES: Record<
	string,
	(wallet: string, chain: string) => string
> = {
	no_signer: (w) =>
		`${w} has no account to sign with. Create or import one in ${w}, then click Execute again.`,
	network_declined: (w, chain) =>
		`${w} didn't switch to this network${chain ? ` (${chain})` : ""}. Select it in ${w}, then try again.`,
	signer_unavailable: (w) =>
		`${w} isn't on the account you're connected as. Switch to it in ${w}, or reconnect.`,
	tx_failed: (w) => `${w} sent the transaction, but it failed.`,
};

export interface ChooserErrorView {
	title: string;
	message: string;
	code: string | null;
}

export function chooserErrorView(
	walletName: string,
	err: unknown,
	chain = "",
): ChooserErrorView {
	const raw = (err as { code?: unknown } | null)?.code;
	const known =
		typeof raw === "string" && Object.hasOwn(ERROR_MESSAGES, raw) ? raw : null;
	return {
		title: `${walletName} couldn't sign`,
		message: known
			? ERROR_MESSAGES[known](walletName, chain)
			: `${walletName} couldn't handle this request.`,
		code: known,
	};
}

// Reopen the chooser as an error report, then resolve when it closes.
export function reportChooserError(
	walletName: string,
	err: unknown,
): Promise<void> {
	const dialog = el("chooser") as HTMLDialogElement | null;
	const list = el("chooser-list");
	if (!dialog || !list) return Promise.resolve();

	const view = chooserErrorView(
		walletName,
		err,
		readMeta("gnoconnect:chainid"),
	);
	dialog.classList.add("b-wallet-chooser--error");
	const title = el("chooser-title");
	if (title) title.textContent = view.title;
	const install = el("chooser-install");
	if (install) install.hidden = true;
	const browser = el("chooser-browser");
	if (browser) browser.hidden = true;
	const close = el("chooser-cancel") as HTMLButtonElement | null;
	if (close) {
		close.textContent = "Close";
		close.onclick = () => dialog.close();
	}

	list.textContent = "";
	list.append(errorPanel(view), hint());

	return new Promise<void>((resolve) => {
		dialog.addEventListener("close", () => resolve(), { once: true });
		if (!dialog.open) show(dialog);
		close?.focus();
	});
}

function errorPanel(view: ChooserErrorView): HTMLElement {
	const li = document.createElement("li");
	li.className = "b-wallet-chooser__error";
	li.setAttribute("role", "alert");

	const svg = document.createElementNS("http://www.w3.org/2000/svg", "svg");
	svg.setAttribute("class", "c-icon");
	svg.setAttribute("aria-hidden", "true");
	const use = document.createElementNS("http://www.w3.org/2000/svg", "use");
	use.setAttribute("href", "#ico-warning");
	svg.appendChild(use);

	const body = document.createElement("div");
	const message = document.createElement("p");
	message.textContent = view.message;
	body.appendChild(message);
	if (view.code) {
		const code = document.createElement("code");
		code.className = "b-wallet-chooser__code";
		code.textContent = view.code;
		body.appendChild(code);
	}
	li.append(svg, body);
	return li;
}

function hint(): HTMLElement {
	const li = document.createElement("li");
	li.className = "b-wallet-chooser__hint";
	li.textContent = "You can also sign with the gnokey command on this page.";
	return li;
}
