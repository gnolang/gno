import { defineSlotRecipe } from '@pandacss/dev'

export const menuRecipe = defineSlotRecipe({
  className: 'menu',
  slots: ['content', 'arrow', 'arrowTip', 'item'],
  base: {
    content: {
      p: 2,
      bg: 'header',
      borderWidth: '1px',
      borderColor: 'background',
    },
    arrow: {
      '--arrow-size': 'token(spacing.4)',
      '--arrow-background': 'token(colors.header)',
    },
    arrowTip: {
      borderTopWidth: '1px',
      borderLeftWidth: '1px',
      borderColor: 'background',
    },
    item: {
      display: 'flex',
      cursor: 'pointer',
      py: 3,
      px: 4,
      w: 'full',

      _highlighted: {
        outline: 'none',
        bg: 'primary',
      },

      _disabled: {
        opacity: '0.5',
      },
    },
  },
  variants: {
    size: {
      sm: {
        content: {
          w: '240px',
        },
      },
      md: {
        content: {
          w: '380px',
        },
      },
      lg: {
        content: {
          w: '420px',
        },
      },
    },
  },
  defaultVariants: {
    size: 'sm',
  },
})
