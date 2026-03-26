import { useEffect, useRef } from 'react'

type DebounceCallback<T extends any[]> = (...args: T) => void
type DebounceFunc<T extends any[]> = (...args: T) => void
type TimeoutID = ReturnType<typeof setTimeout>
type Result<T extends any[]> = [DebounceFunc<T>, () => void]

/**
 * Creates a new debouncer for a given callback with timeout.
 *
 * Returns a tuple of executor and cancel functions.
 */
export const useDebounce = <T extends any[]>(delay: number, callback: DebounceCallback<T>): Result<T> => {
  const timeoutRef = useRef<TimeoutID | null>(null)

  useEffect(() => {
    // Make sure callback is not called after the component is unmounted
    return () => {
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current)
      }
    }
  }, [])

  const delayCallback = (...args: T) => {
    if (timeoutRef.current) {
      clearTimeout(timeoutRef.current)
    }

    timeoutRef.current = setTimeout(() => {
      timeoutRef.current = null
      callback(...args)
    }, delay)

    return timeoutRef
  }

  const callFunc = (...args: T) => delayCallback(...args)
  const cancelFunc = () => timeoutRef.current && clearTimeout(timeoutRef.current)
  return [callFunc, cancelFunc]
}
