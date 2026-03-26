import { defineRecipe } from '@pandacss/dev'

export const buttonRecipe = defineRecipe({
  className: 'button',
  base: {
    display: 'inline-flex',
    alignItems: 'center',
    justifyContent: 'center',
    letterSpacing: 'wider',
    fontWeight: 'medium',
    _disabled: {
      opacity: '0.5',
      cursor: 'not-allowed',
    },
  },
  variants: {
    variant: {
      solid: {
        bg: 'input',
        color: 'input.foreground',
      },
      outline: {
        borderWidth: '1px',
        borderColor: 'input',
        color: 'input',
      },
      ghost: {
        bg: 'transparent',
        color: 'input',
      },
    },
    block: {
      true: {
        width: '100%',
      },
    },
    size: {
      sm: {
        fontSize: 'xs',
        h: '24px',
        px: '2',
      },
      md: {
        fontSize: 'sm',
        h: '33px',
        px: '3',
      },
      lg: {
        fontSize: 'md',
        h: '40px',
        px: '4',
      },
    },
  },
  defaultVariants: {
    variant: 'solid',
    size: 'md',
  },
})
