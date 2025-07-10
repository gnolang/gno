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

export function escapeShellSpecialChars(arg: string): string {
  return arg.replace(/([$`"\\!|&;<>*?{}()])/g, "\\$1");
}

// Cookie utility functions
export function setCookie(
  name: string,
  value: string,
  days: number = 30
): void {
  const expires = new Date();
  expires.setTime(expires.getTime() + days * 24 * 60 * 60 * 1000);
  document.cookie = `${name}=${encodeURIComponent(
    value
  )};expires=${expires.toUTCString()};path=/;SameSite=Strict`;
}

export function getCookie(name: string): string | null {
  const cookies = document.cookie.split("; ").map((c) => c.split("="));
  for (const [key, val] of cookies) {
    if (key === name) return decodeURIComponent(val || "");
  }
  return null;
}