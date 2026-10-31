import { readSession } from "../../feature/connect/frontend/session.js";
import { isAddress } from "./account.js";
import { BaseController } from "./controller.js";
import { linkIsSelf, MAX_TEXT_NODES } from "./you-badge.js";

// YouBadgeController marks the connected address where a realm renders it.
// It inserts its own element only; realm text is never re-parsed.
export class YouBadgeController extends BaseController {
	protected connect(): void {
		const session = readSession();
		const root = document.querySelector("md-renderer");
		if (!session || !root || !isAddress(session.address)) return;
		const { address, username } = session;

		const links: Element[] = [];
		root.querySelectorAll("a[href]").forEach((a) => {
			if (a.closest("pre, code")) return;
			const href = a.getAttribute("href") ?? "";
			if (linkIsSelf(href, window.location.href, address, username)) {
				links.push(a);
			}
		});

		const texts: Text[] = [];
		const walker = document.createTreeWalker(root, NodeFilter.SHOW_TEXT);
		for (let i = 0; i < MAX_TEXT_NODES; i++) {
			const node = walker.nextNode() as Text | null;
			if (!node) break;
			// Links are badged once, above; code shows data, not people.
			if (node.parentElement?.closest("pre, code, a")) continue;
			if (node.data.includes(address)) texts.push(node);
		}

		links.forEach((a) => {
			a.after(badge());
		});
		texts.forEach((node) => {
			node.splitText(node.data.indexOf(address) + address.length);
			node.after(badge());
		});
	}
}

function badge(): HTMLElement {
	const el = document.createElement("span");
	el.className = "b-you";
	el.textContent = "you";
	return el;
}
