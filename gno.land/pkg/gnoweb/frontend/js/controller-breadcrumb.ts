import { BaseController } from "./controller.js";

export class BreadcrumbController extends BaseController {
	static controllerIdentifier = "breadcrumb";

	protected connect(): void {}

	// DOM ACTIONS
	public focus(event: Event): void {
		const target = event.target as HTMLInputElement;
		target.select();
	}
}
