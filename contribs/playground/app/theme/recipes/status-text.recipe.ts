import { defineSlotRecipe } from '@pandacss/dev'

export const statusTextRecipe = defineSlotRecipe({
  className: 'status-text',
  slots: ['root', 'text', 'tag', 'icon', 'progress'],
  base: {
    root: {
      display: 'flex',
      alignItems: 'center',
    },
    text: {
      display: 'block',
      whiteSpace: 'nowrap',
      overflow: 'hidden',
      textOverflow: 'ellipsis',
    },
    icon: {
      marginRight: '2',
      display: 'inline-block',
    },
    progress: {
      // This could be a variant but PandaCSS sometimes randomly
      // doesn't emit styles for pseudo-elements in variants.
      '&::before': {
        content: '"⠋"',
        display: 'inline-block',
        marginRight: '2',
        fontSize: '1.1rem',
        animation: 'textSpinner 1s infinite steps(8)',
      },
    },
    tag: {
      marginRight: '.3em',
      '&::after': {
        content: '": "',
      },
    },
  },
})
