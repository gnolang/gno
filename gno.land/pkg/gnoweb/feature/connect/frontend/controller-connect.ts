import {
	type AccountView,
	accountView,
	chainMatches,
	isAddress,
	loadAccount,
	loadUsername,
	refresh,
} from "../../../frontend/js/account.js";
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
	displayName,
	type GnoSession,
	onSessionChange,
	readSession,
	reconcile,
	setUsername,
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
		// "Change wallet" always shows the list, even with one wallet installed:
		// picking is the whole point, unlike login, which auto-picks a lone wallet.
		this.getTarget("switch")?.addEventListener("click", () =>
			this._pickWallet("Change wallet", true),
		);
		this.getTarget("disconnect")?.addEventListener("click", () => {
			clearSession();
			this._closeMenu();
		});
		// Other controllers (the tx-outcome notice) ask to switch accounts.
		this.on("connect:pick", () => {
			void this._pickWallet("Change wallet", true);
		});
		// Close the menu on an outside click, like the other header popups.
		document.addEventListener("click", (event) => {
			if (!this.element.contains(event.target as Node)) this._closeMenu();
		});

		this._unsubscribe = onSessionChange((session) => this._apply(session));

		// Render the remembered address immediately — no flash of "Connect" on
		// every navigation — then reconcile asynchronously. The address is the
		// user's own and is display-only, so a brief unverified render costs
		// nothing.
		this._render(readSession());
		void reconcile().then((session) => this._apply(session));
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
		if (address) address.textContent = displayName(session);
		const avatar = this.getTarget("avatar");
		// The SVG is generated from a numeric PRNG, never from the address text.
		if (avatar) avatar.innerHTML = avatarSVG(session.address);

		const toggle = this.getTarget("toggle");
		toggle?.setAttribute("title", session.address);
		toggle?.setAttribute(
			"aria-label",
			session.username
				? `Connected as @${session.username} (${session.address})`
				: `Connected as ${session.address}`,
		);

		const link = this.getTarget("username-link") as HTMLAnchorElement | null;
		if (link) {
			link.hidden = !session.username;
			link.textContent = session.username ? `@${session.username}` : "";
			link.href = session.username
				? `/u/${encodeURIComponent(session.username)}`
				: "#";
		}

		const menuAvatar = this.getTarget("menu-avatar");
		if (menuAvatar) menuAvatar.innerHTML = avatarSVG(session.address);
		const short = this.getTarget("short-address");
		if (short) {
			short.textContent = truncate(session.address);
			short.title = session.address;
		}
		// The copy controller on this button reads the value at click time.
		this.getTarget("copy")?.setAttribute(
			"data-copy-text-value",
			session.address,
		);
		const via = this.getTarget("via");
		if (via) {
			via.textContent = [session.name, session.chainid]
				.filter(Boolean)
				.join(" · ");
		}
	}

	private _apply(session: GnoSession | null): void {
		this._render(session);
		void this._resolveUsername(session);
	}

	// Whether chain reads about session can run against this page's RPC.
	private _queryable(session: GnoSession | null): session is GnoSession {
		return (
			!!session && isAddress(session.address) && chainMatches(session.chainid)
		);
	}

	// Check the stored name once per page; a failed lookup keeps it.
	private async _resolveUsername(session: GnoSession | null): Promise<void> {
		if (!this._queryable(session)) return;
		try {
			setUsername(session.address, await loadUsername(session.address));
		} catch (err) {
			this.warn("username lookup failed", err);
		}
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
		void this._renderAccount();
	}

	// Fresh numbers on every open: balance and sequence are what move.
	private async _renderAccount(): Promise<void> {
		const session = readSession();
		const block = this.getTarget("balance-block");
		const rows = this.getTarget("account-rows");
		if (!session || !block || !rows) return;

		block.hidden = true;
		rows.hidden = true;
		if (!this._queryable(session)) return;

		this._fillAccount({
			amount: "…",
			extra: null,
			accountNumber: "…",
			sequence: "…",
		});
		block.hidden = false;
		refresh(session.address);
		let view: AccountView;
		try {
			view = accountView(await loadAccount(session.address));
		} catch (err) {
			this.warn("account lookup failed", err);
			block.hidden = true;
			rows.hidden = true;
			return;
		}
		if (readSession()?.address !== session.address) return; // switched meanwhile
		this._fillAccount(view);
	}

	private _fillAccount(view: AccountView): void {
		const text = (name: string, value: string | null) => {
			const el = this.getTarget(name);
			if (!el) return;
			el.textContent = value ?? "";
			el.hidden = value === null;
		};
		text("amount", view.amount);
		text("extra", view.extra);
		text("account-number", view.accountNumber);
		text("sequence", view.sequence);
		const rows = this.getTarget("account-rows");
		if (rows) rows.hidden = view.accountNumber === null;
	}

	private _closeMenu(): void {
		const menu = this.getTarget("menu");
		if (!menu) return;
		menu.hidden = true;
		this.getTarget("toggle")?.setAttribute("aria-expanded", "false");
	}
}
