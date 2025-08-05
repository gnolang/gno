class Wallet {
	private DOM: {
		el: HTMLElement | null;
	};

	private static button: string = ".js-wallet-button";

	constructor() {
		this.DOM = {
			el: document.querySelector<HTMLElement>(Wallet.button),
		};

		if (this.DOM.el) {
			this.DOM.el.innerHTML = `<p>Connect wallet</p>`;
			this.bindEvents();
		} else {
			console.error("Wallet button not found");
		}
	}

	private bindEvents() {
		this.DOM.el?.addEventListener("click", (e) => {
			e.preventDefault();
			this.connect();
		});
	}

	public connect() {
		console.log("Connecting wallet...");
		if (window.adena) {
			try {
				const res = adena.AddEstablish(window.location.origin);
				this.DOM.el?.setAttribute("disabled", "true");
				this.displayInfos();
			} catch (error) {
				console.error("Failed to connect wallet:", error);
			}
		} else {
			console.error("Adena not found");
		}
	}

	private async displayInfos() {
		const infos = await adena.GetAccount();
		if (this.DOM.el) {
			this.DOM.el.innerHTML = `<p>${infos.data.address}</p>`;
			window.location.href = window.location.href + ":" + infos.data.address;
		} else {
			console.error("Wallet output not found");
		}
	}
}

export default () => new Wallet();
