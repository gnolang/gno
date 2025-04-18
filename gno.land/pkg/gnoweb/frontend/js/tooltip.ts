import { throttle } from "./utils";

class Tooltip {
    private DOM: {
      el: HTMLElement | null;
      tooltip: HTMLSpanElement[];
    };
    
    private static SELECTORS = {
      tooltip: "[data-tooltip]",
    };

    private screenWidth: number;
  
    constructor() {
      this.DOM = {
        el: document.querySelector<HTMLElement>("main"),
        tooltip: [...document.querySelectorAll<HTMLSpanElement>(Tooltip.SELECTORS.tooltip)]
      };
  
      if (this.DOM.el) {
        this.init();
      } else {
        console.warn("Copy: Main container not found.");
      }
    }
  
    private init(): void {
      this.bindEvents();

      this.positionTooltip();
    }
  
    private bindEvents(): void {
      window.addEventListener("resize", throttle(this.positionTooltip.bind(this), 100));
    }
  
    private positionTooltip(): void {
        this.screenWidth = window.innerWidth;
        this.DOM.tooltip.forEach((tooltip) => {
           const tooltipLeft = tooltip.getBoundingClientRect().left;

           if(tooltipLeft > this.screenWidth / 2) {
            tooltip.style.setProperty("--tooltip-left", `initial`);
            tooltip.style.setProperty("--tooltip-right", `0`);
           } else {
            tooltip.style.setProperty("--tooltip-left", `0`);
            tooltip.style.setProperty("--tooltip-right", `initial`);
           }
        });
    }
  }
  
  export default () => new Tooltip();
  