class Tooltip {
  private DOM = {
    tooltip: [] as HTMLSpanElement[],
  };

  private static SELECTORS = {
    tooltip: "[data-tooltip]",
  };

  constructor() {
    this.DOM.tooltip = [
      ...document.querySelectorAll<HTMLSpanElement>(Tooltip.SELECTORS.tooltip),
    ];
    
    this.positionTooltip();
    window.addEventListener("resize", this.positionTooltip.bind(this));
  }

  private positionTooltip(): void {
    const screenWidth = window.innerWidth;

    this.DOM.tooltip.forEach((tooltip) => {
      const tooltipLeft = tooltip.getBoundingClientRect().left;
      const isRightSide = tooltipLeft > screenWidth / 2;

      tooltip.style.setProperty("--tooltip-left", isRightSide ? "initial" : "0");
      tooltip.style.setProperty("--tooltip-right", isRightSide ? "0" : "initial");
    });
  }
}

export default () => new Tooltip();
