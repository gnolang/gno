import { defineSlotRecipe } from '@pandacss/dev'

export const inputRecipe = defineSlotRecipe({
  className: 'input',
  slots: ['root', 'input', 'left'],
  base: {
    root: {
      color: 'input',
      _placeholder: { color: 'input.placeholder' },
    },
    input: {
      appearance: 'none',
      w: 'full',
    },
    left: {
      mr: 2,
    },
  },
  variants: {
    body: {
      field: {
        root: {
          display: 'inline-flex',
          alignItems: 'center',
          backgroundColor: 'transparent',
          outline: '1px solid',
          outlineColor: 'border',
          borderColor: 'transparent',
          borderRightWidth: '10px',
          px: 3,
        },
      },
      strict: {
        root: {
          backgroundColor: 'transparent',
          border: 'none',
          outline: '1px solid',
          outlineColor: 'border',
          outlineOffset: '1px',
          borderRadius: '0',
        },
      },
    },
    size: {
      lg: {
        root: {
          h: '41px',
          lineHeight: '41px',
        },
      },
      md: {
        root: {
          h: '33px',
          lineHeight: '33px',
        },
      },
      xs: {
        root: {
          h: 'auto',
          lineHeight: '1',
        },
      },
    },
  },
  defaultVariants: {
    size: 'md',
    body: 'field',
  },
})
