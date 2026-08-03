import { BaseController } from "./controller.js";
import {
	type GnoTxRequest,
	type GnoWallet,
	getWallets,
} from "./wallet-discovery.js";

// Entry shape of the server-embedded registry (components/wallets.json).
// platforms/install_url are informational for now (no filtering/fallback yet).
interface Wallet {
	name: string;
	id: string;
	icon: string;
	scheme: string; // bare, e.g. "land.gno.gnokey"; this controller appends "://tx?..."
	platforms: string[];
	install_url: string;
}

// A legacy extension: it announces nothing and exposes no provider surface,
// it just writes a global and cancels the submit. Identified by that global,
// which is also the only name we can show for it.
interface LegacyWallet {
	name: string;
	icon: string; // always empty: there is nothing to read it from
}

// `rdns` is what the same wallet announces under once it speaks the protocol.
// A wallet that does both — keeps its global for existing dapps *and* announces,
// which is the additive way to adopt this — would otherwise be listed twice: once
// with its icon, once as a legacy entry without one, indistinguishable to the
// user and pointing at the same extension.
const LEGACY_PROVIDERS: { global: string; name: string; rdns: string }[] = [
	{ global: "adena", name: "Adena", rdns: "land.gno.adena" },
	{ global: "gnoconnect", name: "GnoConnect", rdns: "land.gno.gnoconnect" },
];

// The three kinds of candidate the chooser merges: a wallet that announced
// itself in the page (extension, called directly), a registry entry (external
// app, reached by launch link), and a legacy extension (reached by handing it
// back the submit it expects to intercept).
type Candidate =
	| { kind: "in-page"; wallet: GnoWallet }
	| { kind: "external"; wallet: Wallet }
	| { kind: "legacy"; wallet: LegacyWallet };

// The registry is parsed lazily on first submit and shared across the
// per-function controller instances.
let registryCache: Wallet[] | undefined;

// One window listener for the page, not one per function form: a $help page
// has a controller per function, and every submit is dispatched to all
// window-capture listeners. Handlers are keyed by their form.
//
// Capture phase: a legacy extension listens on document and cancels the event
// there, which is downstream of this. Capture always reaches window first,
// whatever order the listeners were added in, so this is the only place the
// page is certain to be asked at all.
const handlers = new Map<HTMLElement, (event: Event) => void>();
let listening = false;

function listen(): void {
	if (listening) return;
	listening = true;
	window.addEventListener(
		"submit",
		(event) => handlers.get(event.target as HTMLElement)?.(event),
		true,
	);
}

// WalletLaunchController routes the Execute submit to a wallet: an in-page
// one announced over gno:registerWallet (any device), or an external one from
// the embedded registry via a GnoConnect launch link (mobile only). With no
// candidate — desktop without an announcing extension, empty registry — the
// native submit proceeds untouched.
export class WalletLaunchController extends BaseController {
	declare _funcName: string;
	declare _pkgPath: string;
	declare _discovery: ReturnType<typeof getWallets>;
	// Set while re-dispatching a submit meant for a legacy extension, so the
	// handler below lets that one through instead of reopening the chooser.
	private _passThrough = false;

	protected connect(): void {
		this.initializeDOM({});

		// Ask at page load, not at submit: an extension may answer
		// asynchronously, and the chooser should not wait on it to appear.
		this._discovery = getWallets();
		this._discovery.request();

		// Attached to the params <form> (one controller per element); the
		// function name/pkgpath live on the enclosing article.
		const article = this.element.closest<HTMLElement>(
			"[data-action-function-name-value]",
		);
		this._funcName =
			article?.getAttribute("data-action-function-name-value") || "";
		this._pkgPath =
			article?.getAttribute("data-action-function-pkgpath-value") || "";

		handlers.set(this.element, this._onSubmit.bind(this));
		listen();
	}

	// A missing/malformed registry disables external-wallet routing.
	private _wallets(): Wallet[] {
		if (registryCache) return registryCache;
		registryCache = [];
		const script = this.getGlobalTarget("wallet-registry");
		if (!script?.textContent) return registryCache;
		try {
			const parsed = JSON.parse(script.textContent);
			if (Array.isArray(parsed)) registryCache = parsed as Wallet[];
		} catch {
			this.warn("invalid wallet registry JSON");
		}
		return registryCache;
	}

	// Parameter name/value pairs read live from the inputs at submit time
	// (checked boxes of a checkbox group are comma-joined).
	private _readArgs(): Map<string, string> {
		const values = new Map<string, string>();
		this.element
			.querySelectorAll<HTMLInputElement>("[data-action-function-param-value]")
			.forEach((input) => {
				const name =
					input.getAttribute("data-action-function-param-value") || "";
				if (!name) return;
				if (input.type === "checkbox" || input.type === "radio") {
					const prev = values.get(name) ?? "";
					if (input.checked) {
						values.set(
							name,
							prev ? `${prev},${input.value.trim()}` : input.value.trim(),
						);
					} else if (!values.has(name)) {
						values.set(name, "");
					}
				} else {
					values.set(name, input.value.trim());
				}
			});
		return values;
	}

	// Send coins, if the send checkbox is toggled on.
	private _readSend(): string | undefined {
		const box = this.element.querySelector<HTMLInputElement>(
			'input[type="checkbox"][data-action-function-send-value]',
		);
		if (box?.checked) {
			return box.getAttribute("data-action-function-send-value") || undefined;
		}
		return undefined;
	}

	private _meta(name: string): string {
		const el = document.querySelector<HTMLMetaElement>(`meta[name="${name}"]`);
		return el?.content?.trim() || "";
	}

	// Coarse primary pointer only: maxTouchPoints would also match touchscreen
	// laptops, where a failed custom-scheme launch would break Execute.
	private _isMobile(): boolean {
		return window.matchMedia?.("(pointer: coarse)").matches === true;
	}

	// A legacy extension that owns the submit by intercepting it, without
	// announcing itself. One that *has* announced is skipped: it is already in
	// the list, reachable by being called rather than by being handed the event.
	private _legacyProvider(announced: GnoWallet[]): Candidate | null {
		const w = window as unknown as Record<string, unknown>;
		const spoken = new Set(announced.map((a) => a.info?.rdns));
		const found = LEGACY_PROVIDERS.find(
			({ global, rdns }) => Boolean(w[global]) && !spoken.has(rdns),
		);
		return found
			? { kind: "legacy", wallet: { name: found.name, icon: "" } }
			: null;
	}

	// Hand the submit back so a legacy extension's own interceptor sees it.
	// Reaching it any other way is impossible: it exposes nothing to call, and
	// it only acts on the event this page has just taken.
	private _passToLegacy(): void {
		const form = this.element as HTMLFormElement;
		this._passThrough = true;
		if (typeof form.requestSubmit === "function") {
			form.requestSubmit();
			return;
		}
		// Older Safari: re-dispatch by hand, and submit natively if the
		// extension turns out not to claim it after all.
		const event = new Event("submit", { bubbles: true, cancelable: true });
		if (form.dispatchEvent(event)) form.submit();
	}

	// Current page URL minus wallet result params, so repeated round trips
	// don't accumulate stale status/hash.
	private _callbackURL(): string {
		const url = new URL(window.location.href);
		url.searchParams.delete("status");
		url.searchParams.delete("hash");
		return url.toString();
	}

	// The transaction intent, read live from the form. Both transports carry
	// it: the launch link serializes it, an in-page provider is handed it.
	private _txRequest(): GnoTxRequest {
		const tx: GnoTxRequest = {
			path: this._pkgPath,
			func: this._funcName,
			args: Array.from(this._readArgs(), ([name, value]) => ({ name, value })),
		};
		const send = this._readSend();
		if (send) tx.send = send;

		const rpc = this._meta("gnoconnect:rpc");
		const chainid = this._meta("gnoconnect:chainid");
		if (rpc) tx.rpc = rpc;
		if (chainid) tx.chainid = chainid;
		return tx;
	}

	// Compose "<scheme>://sendtx?path=&func=&arg.<name>=&...". Args are named,
	// prefixed "arg." so realm parameter names can't collide with the link's
	// own keys (path, func, send, rpc, chainid, callback).
	private _buildLink(wallet: Wallet): string {
		const enc = encodeURIComponent;
		const tx = this._txRequest();
		const parts: string[] = [`path=${enc(tx.path)}`, `func=${enc(tx.func)}`];
		for (const { name, value } of tx.args) {
			parts.push(`arg.${enc(name)}=${enc(value)}`);
		}
		if (tx.send) parts.push(`send=${enc(tx.send)}`);
		if (tx.rpc) parts.push(`rpc=${enc(tx.rpc)}`);
		if (tx.chainid) parts.push(`chainid=${enc(tx.chainid)}`);
		parts.push(`callback=${enc(this._callbackURL())}`);

		return `${wallet.scheme}://sendtx?${parts.join("&")}`;
	}

	private _openWallet(wallet: Wallet): void {
		window.location.href = this._buildLink(wallet);
	}

	// Hand the intent to an announced wallet. A wallet may announce itself
	// without implementing the tx surface, so a missing method falls back to
	// the native submit rather than dead-ending Execute.
	private async _signInPage(wallet: GnoWallet): Promise<void> {
		const sign = wallet.provider?.sendTx;
		if (typeof sign !== "function") {
			this.warn(
				`wallet "${wallet.info.name}" announced no sendTx; continuing in browser`,
			);
			(this.element as HTMLFormElement).submit();
			return;
		}

		let response: Awaited<ReturnType<typeof sign>>;
		try {
			response = await sign.call(wallet.provider, this._txRequest());
		} catch (err) {
			// The assumption here was that the wallet had already shown the user
			// its own error, so the page owed them nothing. That is only true
			// once the wallet has a window open: a request it refuses outright —
			// an unestablished origin, a chain it does not have, a malformed
			// intent — never reaches a screen, so the user clicked a wallet and
			// watched nothing happen. Show the reason and reveal the copy-paste
			// command, which works regardless of what the wallet decided.
			this.warn(`wallet "${wallet.info.name}" failed to sign`, err);
			this._reportWalletError(wallet.info.name, err);
			return;
		}
		// Rejected is an answer, not a failure: the user declined, the page
		// stays exactly as it was.
		if (response?.status !== "Approved") return;

		// Same landing state as an external wallet's callback, so both
		// transports end the round trip on one URL shape.
		const url = new URL(this._callbackURL());
		url.searchParams.set("status", "success");
		if (response.args?.hash) url.searchParams.set("hash", response.args.hash);
		window.location.href = url.toString();
	}

	// Reopen the chooser to say why the wallet did nothing.
	//
	// The dialog is closed before the wallet is called, so a request refused
	// without ever reaching a wallet screen left the user staring at an
	// unchanged page. `code` is the standard's enumerated reason and is rendered
	// as text — like an announced name, nothing coming back from a wallet is
	// trusted as markup.
	private _reportWalletError(walletName: string, err: unknown): void {
		const dialog = this.getGlobalTarget("chooser") as HTMLDialogElement | null;
		const list = this.getGlobalTarget("chooser-list");
		if (!dialog || !list) return;

		const code = (err as { code?: unknown })?.code;
		const reason = typeof code === "string" ? ` (${code})` : "";

		list.textContent = "";
		const li = document.createElement("li");
		li.className = "b-wallet-chooser__error";
		li.textContent = `${walletName} could not take this transaction${reason}. The gnokey command below works without a wallet.`;
		list.appendChild(li);

		if (dialog.open) return;
		if (typeof dialog.showModal === "function") {
			dialog.showModal();
			this._centerInVisualViewport(dialog);
		} else {
			dialog.setAttribute("open", "");
		}
	}

	// Candidates for this submit: wallets announced in the page (any device)
	// plus, on mobile, the registry's external apps. Desktop external wallets
	// need the cross-device QR, a deferred follow-up.
	//
	// A legacy extension is listed only alongside them, never alone: on its
	// own it still owns the submit by intercepting it, which is one tap fewer
	// and today's behaviour. It is appended last — it is the entry we know
	// least about.
	private _candidates(): Candidate[] {
		const announced = this._discovery.get();
		const inPage: Candidate[] = announced.map((wallet) => ({
			kind: "in-page",
			wallet,
		}));
		const external: Candidate[] = this._isMobile()
			? this._wallets().map((wallet) => ({ kind: "external", wallet }))
			: [];
		if (inPage.length === 0 && external.length === 0) return [];

		const legacy = this._legacyProvider(announced);
		return legacy ? [...inPage, ...external, legacy] : [...inPage, ...external];
	}

	private _onSubmit(event: Event): void {
		// Our own re-dispatch on the way to a legacy extension.
		if (this._passThrough) {
			this._passThrough = false;
			return;
		}

		const candidates = this._candidates();
		if (candidates.length === 0) {
			// Nothing to route to: the native submit (TxLink navigation), and
			// with it any legacy extension interception, proceeds untouched.
			return;
		}

		event.preventDefault();
		// Claim the event before the document-capture listener a legacy
		// extension installs — otherwise it signs whatever the user was still
		// choosing between. stopPropagation, not stopImmediatePropagation:
		// other window-capture listeners (analytics) are not competing for it.
		event.stopPropagation();
		this._openChooser(candidates);
	}

	// Populate and show the page-level chooser dialog. Always shown, even for
	// a single wallet: "Continue in browser" is the only way back to the
	// native submit when the wallet isn't installed (a failed custom-scheme
	// launch is silent).
	private _openChooser(candidates: Candidate[]): void {
		const dialog = this.getGlobalTarget("chooser") as HTMLDialogElement | null;
		const list = this.getGlobalTarget("chooser-list");
		if (!dialog || !list) {
			this._pick(candidates[0]); // no dialog — fail open
			return;
		}

		this._renderCandidates(dialog, list, candidates);

		// A wallet that loads late announces late; re-render so it appears
		// instead of being missing for the life of the dialog.
		const unsubscribe = this._discovery.on(() => {
			this._renderCandidates(dialog, list, this._candidates());
		});
		this._discovery.request();

		// Assignment (not addEventListener) so reopening doesn't stack handlers
		// or submit a previously opened form.
		const browser = this.getGlobalTarget("chooser-browser");
		if (browser) {
			browser.onclick = () => {
				dialog.close();
				// Native submit; bypasses submit listeners, so no re-interception.
				(this.element as HTMLFormElement).submit();
			};
		}
		const cancel = this.getGlobalTarget("chooser-cancel");
		if (cancel) cancel.onclick = () => dialog.close();
		dialog.addEventListener("close", unsubscribe, { once: true });

		if (typeof dialog.showModal === "function") {
			dialog.showModal();
			this._centerInVisualViewport(dialog);
		} else {
			dialog.setAttribute("open", "");
		}
	}

	private _renderCandidates(
		dialog: HTMLDialogElement,
		list: HTMLElement,
		candidates: Candidate[],
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
			// Which transport signs matters to the user: an extension signs
			// here, an app takes them out of the browser and back.
			const kind = document.createElement("span");
			kind.className = "b-wallet-chooser__kind";
			kind.textContent = candidate.kind === "external" ? "App" : "Extension";
			btn.appendChild(kind);

			btn.addEventListener("click", () => {
				dialog.close();
				this._pick(candidate);
			});
			li.appendChild(btn);
			list.appendChild(li);
		});
	}

	private _pick(candidate: Candidate): void {
		if (candidate.kind === "in-page") {
			void this._signInPage(candidate.wallet);
		} else if (candidate.kind === "external") {
			this._openWallet(candidate.wallet);
		} else {
			this._passToLegacy();
		}
	}

	// showModal() centers the dialog in the layout viewport, but a zoomed
	// mobile page (e.g. iOS auto-zoom on sub-16px inputs) only shows part of
	// it, so the dialog can land half off-screen. Shift it to the center of
	// the visual viewport instead, tracking zoom/scroll while open.
	private _centerInVisualViewport(dialog: HTMLDialogElement): void {
		const vv = window.visualViewport;
		if (!vv) return;

		const center = () => {
			const root = document.documentElement;
			// Cap to the visible area: the inner's 90vw is layout-viewport based
			// and overflows the visible width once the page is zoomed.
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
}
