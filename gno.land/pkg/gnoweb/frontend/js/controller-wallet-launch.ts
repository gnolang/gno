import { BaseController } from "./controller.js";

// Wallet mirrors the entry shape from the server-embedded registry
// (components/wallets.json). The scheme is stored bare (e.g.
// "land.gno.gnokey"); this controller composes the "<scheme>://tx?..." prefix.
interface Wallet {
	name: string;
	id: string;
	icon: string;
	scheme: string;
	platforms: string[];
	install_url: string;
}

// WalletLaunchController intercepts the Execute submit of a b-action-function
// article and, on a touch device with at least one registered external wallet,
// opens a GnoConnect launch link instead of the default TxLink GET navigation.
//
// It is strictly additive: whenever no external wallet applies — desktop (QR is
// a deferred follow-up), no registered wallet, or an in-page provider
// (extension) is present — it does nothing and lets the native submit proceed,
// preserving today's TxLink navigation and any extension interception.
export class WalletLaunchController extends BaseController {
	declare _funcName: string;
	declare _pkgPath: string;
	declare _params: Record<string, string>;
	declare _send: string | undefined;
	declare _wallets: Wallet[];
	declare _form: HTMLFormElement | null;

	protected connect(): void {
		this.initializeDOM({});

		// Static identity, read from the co-located action-function article.
		this._funcName =
			this.element.getAttribute("data-action-function-name-value") || "";
		this._pkgPath =
			this.element.getAttribute("data-action-function-pkgpath-value") || "";

		this._wallets = this._loadRegistry();

		// Seed live state from the DOM so an Execute click before any input
		// still produces a correct link (the initial params:changed event may
		// fire before this controller subscribes).
		this._params = this._readParams();
		this._send = this._readSend();

		// Keep state fresh. ActionFunctionController already resolves
		// checkbox/radio values for us, so prefer its payload.
		this.on("params:changed", (event: Event) => {
			const detail = (event as CustomEvent).detail;
			if (detail.funcName !== this._funcName) return;
			this._params = { ...detail.params };
			this._send = detail.send;
		});

		this._form = document.getElementById(
			`form-${this._funcName}`,
		) as HTMLFormElement | null;
		this._form?.addEventListener("submit", this._onSubmit.bind(this));
	}

	// Parse the embedded wallet registry. A missing/malformed registry simply
	// disables external-wallet routing (fall through to native submit).
	private _loadRegistry(): Wallet[] {
		const script = this.getGlobalTarget("wallet-registry");
		if (!script?.textContent) return [];
		try {
			const parsed = JSON.parse(script.textContent);
			return Array.isArray(parsed) ? (parsed as Wallet[]) : [];
		} catch {
			this.warn("invalid wallet registry JSON");
			return [];
		}
	}

	// Ordered parameter names, in declaration (DOM) order. Deriving order from
	// the inputs avoids relying on object key order in the params payload.
	private _orderedParamNames(): string[] {
		const names: string[] = [];
		const seen = new Set<string>();
		this.element
			.querySelectorAll<HTMLElement>("[data-action-function-param-value]")
			.forEach((el) => {
				const name = el.getAttribute("data-action-function-param-value") || "";
				if (name && !seen.has(name)) {
					seen.add(name);
					names.push(name);
				}
			});
		return names;
	}

	// Live parameter values straight from the inputs (seed / fallback).
	private _readParams(): Record<string, string> {
		const params: Record<string, string> = {};
		this.element
			.querySelectorAll<HTMLInputElement>("[data-action-function-param-value]")
			.forEach((input) => {
				const name =
					input.getAttribute("data-action-function-param-value") || "";
				if (!name) return;
				if (input.type === "checkbox" || input.type === "radio") {
					if (input.checked) {
						const prev = params[name];
						params[name] = prev
							? `${prev},${input.value.trim()}`
							: input.value.trim();
					} else if (!(name in params)) {
						params[name] = "";
					}
				} else {
					params[name] = input.value.trim();
				}
			});
		return params;
	}

	// Live send coins from the send checkbox, if toggled on.
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

	// Touch device heuristic — coarse pointer or touch points.
	private _isMobile(): boolean {
		return (
			window.matchMedia?.("(pointer: coarse)").matches === true ||
			navigator.maxTouchPoints > 0
		);
	}

	// An installed in-page provider (browser extension) should own the submit.
	// Never break that path: if one is present, do nothing.
	private _hasInPageProvider(): boolean {
		const w = window as unknown as Record<string, unknown>;
		return Boolean(w.adena || w.gnoconnect);
	}

	// Compose "<scheme>://tx?path=&func=&args=&...&rpc=&chainid=&callback=".
	// Values are percent-encoded; args are repeated once per positional param
	// in declaration order (empty values included to keep positions aligned).
	private _buildLink(wallet: Wallet): string {
		const enc = encodeURIComponent;
		const parts: string[] = [
			`path=${enc(this._pkgPath)}`,
			`func=${enc(this._funcName)}`,
		];
		for (const name of this._orderedParamNames()) {
			parts.push(`args=${enc(this._params[name] ?? "")}`);
		}
		if (this._send) parts.push(`send=${enc(this._send)}`);

		const rpc = this._meta("gnoconnect:rpc");
		const chainid = this._meta("gnoconnect:chainid");
		if (rpc) parts.push(`rpc=${enc(rpc)}`);
		if (chainid) parts.push(`chainid=${enc(chainid)}`);

		// Return URL so the wallet can hand the result back to this page.
		parts.push(`callback=${enc(window.location.href)}`);

		return `${wallet.scheme}://tx?${parts.join("&")}`;
	}

	private _openWallet(wallet: Wallet): void {
		window.location.href = this._buildLink(wallet);
	}

	private _onSubmit(event: Event): void {
		// Fall through to the native submit (today's TxLink navigation and any
		// extension interception) whenever external-wallet routing doesn't apply.
		if (this._hasInPageProvider()) return;
		if (!this._isMobile()) return; // desktop: QR is a deferred follow-up
		if (this._wallets.length === 0) return;

		event.preventDefault();

		if (this._wallets.length === 1) {
			this._openWallet(this._wallets[0]);
			return;
		}
		this._openChooser();
	}

	// Populate and show the chooser dialog for the >1 wallet case.
	private _openChooser(): void {
		const dialog = this.getTarget("chooser") as HTMLDialogElement | null;
		const list = this.getTarget("chooser-list");
		if (!dialog || !list) {
			// No dialog available — fail open to the first wallet.
			this._openWallet(this._wallets[0]);
			return;
		}

		list.textContent = "";
		this._wallets.forEach((wallet) => {
			const li = document.createElement("li");
			const btn = document.createElement("button");
			btn.type = "button";
			btn.className = "b-wallet-chooser__item";
			if (wallet.icon) {
				const img = document.createElement("img");
				img.src = wallet.icon;
				img.alt = "";
				img.className = "b-wallet-chooser__icon";
				btn.appendChild(img);
			}
			const label = document.createElement("span");
			label.textContent = wallet.name;
			btn.appendChild(label);
			btn.addEventListener("click", () => {
				dialog.close();
				this._openWallet(wallet);
			});
			li.appendChild(btn);
			list.appendChild(li);
		});

		if (typeof dialog.showModal === "function") {
			dialog.showModal();
		} else {
			dialog.setAttribute("open", "");
		}
	}

	// DOM ACTION — Cancel button in the chooser dialog.
	public closeChooser(): void {
		const dialog = this.getTarget("chooser") as HTMLDialogElement | null;
		dialog?.close();
	}
}

export default WalletLaunchController;
