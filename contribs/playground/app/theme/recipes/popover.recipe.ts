import { defineSlotRecipe } from '@pandacss/dev'

export const popoverRecipe = defineSlotRecipe({
  className: 'popover',
  slots: ['content', 'title', 'close', 'arrow', 'arrowTip', 'section', 'sectionLabel'],
  base: {
    content: {
      p: 5,
      w: '300px',
      bg: 'header',
      borderWidth: '1px',
      borderColor: 'background',
      maxH: '80vh',
      overflowY: 'auto',
    },
    title: {
      fontSize: 'lg',
      fontWeight: 'bold',
      display: 'flex',
      alignItems: 'center',
    },
    close: {
      position: 'absolute',
      right: '2',
      top: '2',
      p: '1.5',
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
    section: {
      p: 3,
      borderBottomWidth: '1px',
      borderBottomColor: 'gray.300',
      '&:last-child': {
        borderBottom: 'none',
      },
    },
    sectionLabel: {
      display: 'block',
      fontSize: 'xs',
      pb: 3,
    },
  },
  variants: {
    size: {
      xs: {
        content: {
          w: '200px',
        },
      },
      sm: {
        content: {
          w: '300px',
        },
      },
      md: {
        content: {
          w: '430px',
        },
      },
      lg: {
        content: {
          w: '520px',
        },
      },
    },
    layout: {
      sections: {
        content: {
          p: 0,
        },
      },
    },
  },
  defaultVariants: {
    size: 'sm',
  },
})
