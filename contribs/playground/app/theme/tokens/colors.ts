import { defineTokens } from '@pandacss/dev'

export const colors = defineTokens.colors({
  current: { value: 'currentColor' },
  dark: { value: '#111' },
  black: { value: '#000' },
  white: { value: '#fff' },
  gray: {
    200: { value: '#DADCDE' },
    300: { value: '#C7C7C7' },
    400: { value: '#939393' },
    500: { value: '#606060' },
    950: { value: '#1E1E1E' },
  },
})
