import { useCallback, useState } from 'react'

export function useWriteClipboard() {
  const [isCopied, setIsCopied] = useState(false)

  const writeToClipboard = useCallback(
    async (text: string) => {
      await navigator.clipboard.writeText(text)
      setIsCopied(true)

      setTimeout(() => {
        setIsCopied(false)
      }, 1000)
    },
    [setIsCopied],
  )

  return [writeToClipboard, isCopied] as const
}
