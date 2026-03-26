import { defineRecipe } from '@pandacss/dev'

export const textRecipe = defineRecipe({
  className: 'text',
  base: {
    fontSize: 'md',
  },
  variants: {
    align: {
      left: {
        textAlign: 'left',
      },
      right: {
        textAlign: 'right',
      },
      center: {
        textAlign: 'center',
      },
    },
    ellipsis: {
      true: {
        overflow: 'hidden',
        textOverflow: 'ellipsis',
        whiteSpace: 'nowrap',
      },
    },
    transform: {
      uppercase: {
        textTransform: 'uppercase',
      },
    },
    size: {
      xs: {
        fontSize: 'xs',
      },
      sm: {
        fontSize: 'sm',
      },
      md: {
        fontSize: 'md',
      },
      lg: {
        fontSize: 'lg',
      },
    },
  },
})
