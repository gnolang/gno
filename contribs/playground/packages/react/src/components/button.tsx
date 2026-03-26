import { forwardRef, type ComponentPropsWithoutRef } from 'react'

import { ark } from '@ark-ui/react'

export type ButtonProps = ComponentPropsWithoutRef<typeof ark.button>

export const Button = forwardRef<HTMLButtonElement, ButtonProps>((props, ref) => (
  <ark.button data-scope="button" data-part="root" {...props} ref={ref} />
))

Button.displayName = 'Button'
