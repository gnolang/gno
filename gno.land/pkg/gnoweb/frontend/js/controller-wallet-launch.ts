import {
	connectWallet,
	findAnnounced,
	type GnoSession,
	readSession,
} from "../../feature/connect/frontend/session.js";
import { BaseController } from "./controller.js";
import {
	type Candidate,
	legacyCandidate,
	loadRegistry,
	openChooser,
	type RegistryWallet,
	reportChooserError,
} from "./wallet-chooser.js";
import {
	type GnoTxRequest,
	type GnoWallet,
	getWallets,
} from "./wallet-discovery.js";

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

// WalletLaunchController routes the Execute submit to a wallet: an in-page one
// announced over gno:registerWallet (any device), or an external one from the
// embedded registry via a GnoConnect launch link (mobile only). With no
// candidate the native submit proceeds untouched.
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

	// Hand the submit back so a legacy extension's own interceptor sees it.
	// The one path that does not navigate: the extension owns the event, and
	// navigating underneath it would cancel the approval it is opening.
	private _passToLegacy(): void {
		const form = this.element as HTMLFormElement;
		this._passThrough = true;
		if (typeof form.requestSubmit === "function") {
			form.requestSubmit();
			return;
		}
		const event = new Event("submit", { bubbles: true, cancelable: true });
		if (form.dispatchEvent(event)) form.submit();
	}

	// The function help page for this form: `$help&func=Name&p1=v1&...`, live
	// from the form action, plus the anchor so scroll position survives.
	private _helpURL(): string {
		const action =
			(this.element as HTMLFormElement).getAttribute("action") ||
			window.location.href;
		const url = new URL(action, window.location.href);
		url.searchParams.delete("status");
		url.searchParams.delete("hash");
		url.hash = `func-${this._funcName}`;
		return url.toString();
	}

	// Where an external wallet returns to, and the base for the success
	// landing. Same URL as _helpURL without the anchor, which a callback
	// round trip would only lose anyway.
	private _callbackURL(): string {
		const url = new URL(this._helpURL());
		url.hash = "";
		return url.toString();
	}

	private _navigate(url: string): void {
		window.location.assign(url);
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

		// Pin the identity the user picked, so the wallet signs as that account
		// rather than whichever is to hand.
		const session = readSession();
		if (session?.address) tx.signer = session.address;
		return tx;
	}

	// Compose "<scheme>://sendtx?path=&func=&arg.<name>=&...". Args are named,
	// prefixed "arg." so realm parameter names can't collide with the link's
	// own keys (path, func, send, rpc, chainid, signer, callback).
	private _buildLink(wallet: RegistryWallet): string {
		const enc = encodeURIComponent;
		const tx = this._txRequest();
		const parts: string[] = [`path=${enc(tx.path)}`, `func=${enc(tx.func)}`];
		for (const { name, value } of tx.args) {
			parts.push(`arg.${enc(name)}=${enc(value)}`);
		}
		if (tx.send) parts.push(`send=${enc(tx.send)}`);
		if (tx.rpc) parts.push(`rpc=${enc(tx.rpc)}`);
		if (tx.chainid) parts.push(`chainid=${enc(tx.chainid)}`);
		if (tx.signer) parts.push(`signer=${enc(tx.signer)}`);
		parts.push(`callback=${enc(this._callbackURL())}`);

		return `${wallet.scheme}://sendtx?${parts.join("&")}`;
	}

	// Pin the args in the URL, then fire the launch link — and do not navigate
	// after it. iOS gates a custom scheme behind a system "Open in …?" prompt,
	// and any navigation dismisses that prompt before the user can answer, so
	// the wallet never opens; deferring it to a task does not help. Nothing is
	// lost by staying: the destination was this same page with the args pinned,
	// which replaceState reaches without tearing the page down, and the result
	// comes back through the callback URL.
	private _openWallet(wallet: RegistryWallet): void {
		window.history.replaceState(null, "", this._helpURL());
		window.location.href = this._buildLink(wallet);
	}

	// Sign first, navigate on the outcome. sendTx() returns a Promise in the
	// live page; navigating first would destroy it, and some extensions cancel
	// a pending approval when the requesting tab navigates away.
	private async _signInPage(wallet: GnoWallet, retried = false): Promise<void> {
		const sign = wallet.provider?.sendTx;
		if (typeof sign !== "function") {
			// A wallet announcing without the tx surface is non-conforming, but
			// gnoweb still must not dead-end: go where a rejection would.
			this.warn(
				`wallet "${wallet.info.name}" announced no sendTx; continuing in browser`,
			);
			this._navigate(this._helpURL());
			return;
		}

		let response: Awaited<ReturnType<typeof sign>>;
		try {
			response = await sign.call(wallet.provider, this._txRequest());
		} catch (err) {
			this.warn(`wallet "${wallet.info.name}" failed to sign`, err);
			await reportChooserError(wallet.info.name, err);
			this._navigate(this._helpURL());
			return;
		}

		if (response?.status === "Approved") {
			// The landing state both transports converge on.
			const url = new URL(this._callbackURL());
			url.searchParams.set("status", "success");
			if (response.args?.hash) url.searchParams.set("hash", response.args.hash);
			url.hash = `func-${this._funcName}`;
			// sendTx returns a hash, not an address, so signing while unconnected
			// leaves the user unconnected: adopt the identity before leaving.
			if (!readSession()) await connectWallet(wallet);
			this._navigate(url.toString());
			return;
		}

		// not_connected is never surfaced: connect and retry once. Only a
		// declined connect becomes a plain Rejected.
		if (response?.code === "not_connected" && !retried) {
			const session = await connectWallet(wallet);
			if (session) {
				await this._signInPage(wallet, true);
				return;
			}
		}

		// Rejected, or anything else: the function help page, args pinned.
		this._navigate(this._helpURL());
	}

	// Candidates for this submit: wallets announced in the page (any device)
	// plus, on mobile, the registry's launchable apps. A scheme-less entry is
	// never launched.
	private _candidates(): Candidate[] {
		const announced = this._discovery.get();
		const inPage: Candidate[] = announced.map((wallet) => ({
			kind: "in-page",
			wallet,
		}));
		const external: Candidate[] = this._isMobile()
			? loadRegistry()
					.filter((w) => w.kind === "app" && w.scheme !== "")
					.map((wallet) => ({ kind: "external", wallet }))
			: [];
		if (inPage.length === 0 && external.length === 0) return [];

		const legacy = legacyCandidate(announced);
		return legacy ? [...inPage, ...external, legacy] : [...inPage, ...external];
	}

	// The remembered wallet for this session, if it is reachable right now.
	private _sessionCandidate(session: GnoSession): Candidate | null {
		const announced = findAnnounced(session.rdns);
		if (announced) return { kind: "in-page", wallet: announced };
		const entry = loadRegistry().find(
			(w) => w.rdns === session.rdns && w.kind === "app" && w.scheme !== "",
		);
		return entry ? { kind: "external", wallet: entry } : null;
	}

	private _onSubmit(event: Event): void {
		// Our own re-dispatch on the way to a legacy extension.
		if (this._passThrough) {
			this._passThrough = false;
			return;
		}

		const candidates = this._candidates();
		if (candidates.length === 0) {
			// Nothing to route to: the native submit navigates to the help page
			// on its own, and any legacy extension interception is untouched.
			return;
		}

		event.preventDefault();
		// Claim the event before the document-capture listener a legacy
		// extension installs. stopPropagation, not stopImmediatePropagation:
		// other window-capture listeners (analytics) are not competing for it.
		event.stopPropagation();

		// Connected: straight to the remembered wallet, no chooser.
		const session = readSession();
		const remembered = session ? this._sessionCandidate(session) : null;
		if (remembered) {
			this._pick(remembered);
			return;
		}

		// Exactly one wallet available: it fires, and the navigation follows.
		if (candidates.length === 1) {
			this._pick(candidates[0]);
			return;
		}

		void openChooser({
			refresh: () => this._candidates(),
			// Re-dispatch gno:requestWallet on open: a wallet that only answers
			// explicit requests may have missed the connect-time one.
			onOpen: () => this._discovery.request(),
			// Dismissing the dialog already lands on the help page with the args
			// pinned, so the fallback needs no action of its own — it only says
			// that out loud. Not the native submit: that rebuilds the query from
			// the form's unnamed inputs and drops the args on the floor.
			browser: { label: "Continue in browser" },
		}).then((picked) => {
			if (picked) {
				this._pick(picked);
				return;
			}
			// Cancel, Esc, backdrop, browser fallback: all still a navigation, and
			// the args round-trip through the URL, so nothing is lost.
			this._navigate(this._helpURL());
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
}
