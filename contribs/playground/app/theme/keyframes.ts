import { defineKeyframes } from '@pandacss/dev'

export const keyframes = defineKeyframes({
  rotate: {
    from: {
      transform: 'rotate(0deg)',
    },
    to: {
      transform: 'rotate(360deg)',
    },
  },
  textSpinner: {
    '0%': { content: '"⠋"' },
    '12.5%': { content: '"⠙"' },
    '25%': { content: '"⠹"' },
    '37.5%': { content: '"⠸"' },
    '50%': { content: '"⠼"' },
    '62.5%': { content: '"⠴"' },
    '75%': { content: '"⠦"' },
    '87.5%': { content: '"⠧"' },
    '100%': { content: '"⠏"' },
  },
})
