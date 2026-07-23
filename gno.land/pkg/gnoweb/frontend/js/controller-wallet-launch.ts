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

// The two kinds of candidate the chooser merges: a wallet that announced
// itself in the page (extension, called directly) and a registry entry
// (external app, reached by launch link).
type Candidate =
	| { kind: "in-page"; wallet: GnoWallet }
	| { kind: "external"; wallet: Wallet };

// The registry is parsed lazily on first submit and shared across the
// per-function controller instances.
let registryCache: Wallet[] | undefined;

// WalletLaunchController routes the Execute submit to a wallet: an in-page
// one announced over gno:registerWallet (any device), or an external one from
// the embedded registry via a GnoConnect launch link (mobile only). With no
// candidate — desktop without an announcing extension, empty registry — the
// native submit proceeds untouched.
export class WalletLaunchController extends BaseController {
	declare _funcName: string;
	declare _pkgPath: string;
	declare _discovery: ReturnType<typeof getWallets>;

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

		this.element.addEventListener("submit", this._onSubmit.bind(this));
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
	// announcing itself. Only consulted when nothing announced: a wallet that
	// speaks the announce protocol is chosen explicitly instead.
	private _hasLegacyProvider(): boolean {
		const w = window as unknown as Record<string, unknown>;
		return Boolean(w.adena || w.gnoconnect);
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

	// Compose "<scheme>://tx?path=&func=&arg.<name>=&...". Args are named,
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

		return `${wallet.scheme}://tx?${parts.join("&")}`;
	}

	private _openWallet(wallet: Wallet): void {
		window.location.href = this._buildLink(wallet);
	}

	// Hand the intent to an announced wallet. A wallet may announce itself
	// without implementing the tx surface, so a missing method falls back to
	// the native submit rather than dead-ending Execute.
	private async _signInPage(wallet: GnoWallet): Promise<void> {
		const sign = wallet.provider?.signAndSubmitTransaction;
		if (typeof sign !== "function") {
			this.warn(
				`wallet "${wallet.info.name}" announced no signAndSubmitTransaction; continuing in browser`,
			);
			(this.element as HTMLFormElement).submit();
			return;
		}

		let response: Awaited<ReturnType<typeof sign>>;
		try {
			response = await sign.call(wallet.provider, this._txRequest());
		} catch (err) {
			// The wallet showed the user its own error; leave the form intact
			// so they can retry or copy the command.
			this.warn(`wallet "${wallet.info.name}" failed to sign`, err);
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

	// Candidates for this submit: wallets announced in the page (any device)
	// plus, on mobile, the registry's external apps. Desktop external wallets
	// need the cross-device QR, a deferred follow-up.
	private _candidates(): Candidate[] {
		const inPage: Candidate[] = this._discovery
			.get()
			.map((wallet) => ({ kind: "in-page", wallet }));
		const external: Candidate[] = this._isMobile()
			? this._wallets().map((wallet) => ({ kind: "external", wallet }))
			: [];
		return [...inPage, ...external];
	}

	private _onSubmit(event: Event): void {
		const candidates = this._candidates();
		if (candidates.length === 0) {
			// Nothing to route to: the native submit (TxLink navigation), and
			// with it any legacy extension interception, proceeds untouched.
			return;
		}
		// A legacy extension owns the submit, but only while no wallet has
		// announced itself — an announced wallet is picked by the user here.
		if (
			this._hasLegacyProvider() &&
			!candidates.some((c) => c.kind === "in-page")
		) {
			return;
		}

		event.preventDefault();
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
			kind.textContent = candidate.kind === "in-page" ? "Extension" : "App";
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
		} else {
			this._openWallet(candidate.wallet);
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
