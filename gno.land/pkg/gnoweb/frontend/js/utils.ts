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

export 	const throttle = <T extends unknown[]>(
  callback: (...args: T) => void,
  delay: number,
) => {
  let isWaiting = false;
 
  return (...args: T) => {
    if (isWaiting) {
      return;
    }
 
    callback(...args);
    isWaiting = true;
 
    setTimeout(() => {
      isWaiting = false;
    }, delay);
  };
};

export function escapeShellSpecialChars(arg: string): string {
  return arg.replace(/([$`"\\!|&;<>*?{}()])/g, "\\$1");
}
