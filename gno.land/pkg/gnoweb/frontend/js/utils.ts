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

export function escapeShellContent(arg: string): string {
  return arg.replace(/([$`"\\!|&;<>*?{}()])/g, "\\$1");
}
