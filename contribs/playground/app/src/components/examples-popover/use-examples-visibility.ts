import { useEffect, useState } from 'react'

export function useExamplesVisibility() {
  const [isVisible, setIsVisible] = useState(false)

  const toggleVisibility = () => {
    setIsVisible((prev) => !prev)
  }

  const hideExamples = () => {
    setIsVisible(false)
  }

  useEffect(() => {
    const handleKeyDown = (event: KeyboardEvent) => {
      if ((event.ctrlKey || event.metaKey) && event.key === 'e') {
        event.preventDefault()
        toggleVisibility()
      }

      if (event.key === 'Escape' && isVisible) {
        hideExamples()
      }
    }

    window.addEventListener('keydown', handleKeyDown)
    return () => {
      window.removeEventListener('keydown', handleKeyDown)
    }
  }, [isVisible])

  useEffect(() => {
    if (window.location.hash === '#examples') {
      setIsVisible(true)
      window.history.replaceState(null, '', window.location.pathname + window.location.search)
    }
  }, [])

  return {
    isVisible,
    toggleVisibility,
    hideExamples,
  }
}
