import { BaseController } from "../../../frontend/js/controller.js";
import {
	type Candidate,
	legacyCandidate,
	openChooser,
} from "../../../frontend/js/wallet-chooser.js";
import { getWallets } from "../../../frontend/js/wallet-discovery.js";
import { avatarSVG } from "./avatar.js";
import {
	clearSession,
	connectWallet,
	type GnoSession,
	onSessionChange,
	readSession,
	reconcile,
	truncate,
} from "./session.js";

// ConnectController owns the header identity: it renders the remembered
// account, opens the chooser to connect or switch, and drops the session on
// disconnect. The session itself lives in session.ts.
export class ConnectController extends BaseController {
	declare _discovery: ReturnType<typeof getWallets>;
	declare _unsubscribe: () => void;

	protected connect(): void {
		this.initializeDOM({});
		this._discovery = getWallets();
		this._discovery.request();

		this.getTarget("connect-btn")?.addEventListener("click", () =>
			this._pickWallet("Connect a wallet", false),
		);
		this.getTarget("toggle")?.addEventListener("click", () =>
			this._toggleMenu(),
		);
		this.getTarget("copy")?.addEventListener("click", () =>
			this._copyAddress(),
		);
		// "Change wallet" always shows the list, even with one wallet installed:
		// picking is the whole point, unlike login, which auto-picks a lone wallet.
		this.getTarget("switch")?.addEventListener("click", () =>
			this._pickWallet("Change wallet", true),
		);
		this.getTarget("disconnect")?.addEventListener("click", () => {
			clearSession();
			this._closeMenu();
		});
		// Close the menu on an outside click, like the other header popups.
		document.addEventListener("click", (event) => {
			if (!this.element.contains(event.target as Node)) this._closeMenu();
		});

		this._unsubscribe = onSessionChange((session) => this._render(session));

		// Render the remembered address immediately — no flash of "Connect" on
		// every navigation — then reconcile asynchronously. The address is the
		// user's own and is display-only, so a brief unverified render costs
		// nothing.
		this._render(readSession());
		void reconcile().then((session) => this._render(session));
	}

	protected disconnect(): void {
		this._unsubscribe?.();
	}

	private _render(session: GnoSession | null): void {
		const button = this.getTarget("connect-btn");
		const account = this.getTarget("account");
		if (!button || !account) return;

		if (!session) {
			button.hidden = false;
			account.hidden = true;
			this._closeMenu();
			return;
		}

		button.hidden = true;
		account.hidden = false;

		const address = this.getTarget("address");
		if (address) address.textContent = truncate(session.address);
		const avatar = this.getTarget("avatar");
		// The SVG is generated from a numeric PRNG, never from the address text.
		if (avatar) avatar.innerHTML = avatarSVG(session.address);

		const toggle = this.getTarget("toggle");
		toggle?.setAttribute("title", session.address);
		toggle?.setAttribute("aria-label", `Connected as ${session.address}`);
	}

	// Candidates for logging in: announced wallets, plus a legacy extension
	// only alongside them. External apps are excluded — a launch-link connect
	// leaves the page, and the header is not where that round trip belongs.
	private _candidates(): Candidate[] {
		const announced = this._discovery.get();
		const inPage: Candidate[] = announced.map((wallet) => ({
			kind: "in-page",
			wallet,
		}));
		const legacy = legacyCandidate(announced);
		return legacy ? [...inPage, legacy] : inPage;
	}

	private async _pickWallet(title: string, always: boolean): Promise<void> {
		this._closeMenu();
		this._discovery.request();

		const candidates = this._candidates();
		if (candidates.length === 0) {
			// Nothing installed: the install page is the only useful answer.
			window.location.href = "/wallets";
			return;
		}

		let picked: Candidate | null = candidates[0];
		if (always || candidates.length > 1) {
			picked = await openChooser({ refresh: () => this._candidates(), title });
		}
		if (!picked || picked.kind !== "in-page") {
			// A legacy extension exposes nothing to call, so it cannot be logged
			// into; it stays reachable for signing only.
			if (picked?.kind === "legacy") {
				this.warn(
					`wallet "${picked.wallet.name}" cannot be connected to; it only intercepts submits`,
				);
			}
			return;
		}
		await connectWallet(picked.wallet);
	}

	private _copyAddress(): void {
		const session = readSession();
		if (!session) return;
		void navigator.clipboard?.writeText(session.address);
		this._closeMenu();
	}

	private _toggleMenu(): void {
		const menu = this.getTarget("menu");
		if (!menu) return;
		menu.hidden ? this._openMenu() : this._closeMenu();
	}

	private _openMenu(): void {
		const menu = this.getTarget("menu");
		if (!menu) return;
		menu.hidden = false;
		this.getTarget("toggle")?.setAttribute("aria-expanded", "true");
	}

	private _closeMenu(): void {
		const menu = this.getTarget("menu");
		if (!menu) return;
		menu.hidden = true;
		this.getTarget("toggle")?.setAttribute("aria-expanded", "false");
	}
}
