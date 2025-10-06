import { BaseController } from "./controller.js";

export class TooltipController extends BaseController {
	protected connect(): void {
		this.initializeDOM({
			tooltip: this.getTargets("info"),
		});

		if (this.getDOMArray("tooltip").length > 0) {
			this._positionAllTooltips();
			window.addEventListener("resize", this._positionAllTooltips.bind(this));
		}
	}

	// position all tooltips
	private _positionAllTooltips(): void {
		const screenWidth = window.innerWidth;

		this.getDOMArray("tooltip").forEach((tooltip) => {
			const tooltipLeft = tooltip.getBoundingClientRect().left;
			const isRightSide = tooltipLeft > screenWidth / 2;

			tooltip.style.setProperty(
				"--tooltip-left",
				isRightSide ? "initial" : "0",
			);
			tooltip.style.setProperty(
				"--tooltip-right",
				isRightSide ? "0" : "initial",
			);
		});
	}
}
