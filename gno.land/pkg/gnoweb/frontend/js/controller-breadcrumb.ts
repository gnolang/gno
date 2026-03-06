import { BaseController } from "./controller.js";

export class BreadcrumbController extends BaseController {
	protected connect(): void {}

	// DOM ACTIONS
	public focus(event: Event): void {
		const target = event.target as HTMLInputElement;
		target.select();
	}
}
