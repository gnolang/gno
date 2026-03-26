type Debounce = <T extends unknown[]>(cb: (...args: T) => void, delay?: number) => (...args: T) => void

export const debounce: Debounce = (cb, delay = 300) => {
  let timeout = 0
  return function debounced(...args) {
    clearTimeout(timeout)
    timeout = window.setTimeout(() => cb(...args), delay)
  }
}
