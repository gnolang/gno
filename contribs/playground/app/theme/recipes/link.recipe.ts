import { defineRecipe } from '@pandacss/dev'

export const linkRecipe = defineRecipe({
  className: 'link',
  base: {
    textDecoration: 'underline',
    overflowWrap: 'break-word',
    cursor: 'pointer',
    color: 'link',

    _disabled: {
      cursor: 'not-allowed',
      color: 'foreground.muted',
    },
  },
})
