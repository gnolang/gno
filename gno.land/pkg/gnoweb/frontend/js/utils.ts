export function debounce<T extends (...args: any[]) => void>(func: T, delay: number = 250): (...args: Parameters<T>) => void {
  let timeoutId: ReturnType<typeof setTimeout> | undefined;

  return function (this: any, ...args: Parameters<T>) {
    if (timeoutId !== undefined) {
      clearTimeout(timeoutId);
    }
    timeoutId = setTimeout(() => {
      func.apply(this, args);
    }, delay);
  };
}

export const throttle = (fn: Function, wait: number) => {
  let lastTime = 0;
  return (...args: any[]) => {
    const now = Date.now();
    if (now - lastTime >= wait) {
      lastTime = now;
      fn(...args);
    }
  };
};

export function escapeShellSpecialChars(arg: string): string {
  return arg.replace(/([$`"\\!|&;<>*?{}()])/g, "\\$1");
}

class ViewportObserver {
  private width: number;

  constructor() {
    this.width = window.innerWidth;

    window.addEventListener(
      "resize",
      throttle(() => {
        this.width = window.innerWidth;
      }, 100)
    );
  }

  getWidth(): number {
    return this.width;
  }
}
export const viewport = new ViewportObserver();

