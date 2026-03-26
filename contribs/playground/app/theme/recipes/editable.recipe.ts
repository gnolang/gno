import { defineSlotRecipe } from '@pandacss/dev'

export const editableRecipe = defineSlotRecipe({
  className: 'editable',
  slots: ['root', 'area', 'input', 'preview', 'control', 'submitTrigger'],
  base: {
    root: {
      display: 'inline-flex',
    },
    input: {
      bg: 'background',
    },
    preview: {
      outline: '1px solid transparent',
      outlineOffset: '1px',
      _hover: {
        outlineColor: 'border',
      },
    },
    control: {
      display: 'inline-flex',
    },
    submitTrigger: {
      height: '100%',
      px: 0.5,

      _hover: {
        bg: 'primary.background',
      },
    },
  },
})
