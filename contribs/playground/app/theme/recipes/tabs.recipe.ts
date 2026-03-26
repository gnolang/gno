import { defineSlotRecipe } from '@pandacss/dev'

export const tabsRecipe = defineSlotRecipe({
  className: 'tabs',

  slots: ['root', 'list', 'trigger', 'close', 'textInput'],

  base: {
    root: {
      display: 'flex',
      flexDirection: 'column',
    },
    list: {
      flexShrink: 0,
      display: 'flex',
      overflowX: 'auto',
      scrollSnapType: 'x mandatory',
    },
    trigger: {
      userSelect: 'none',
      py: '2',
      px: '3',
      background: 'header',
      color: 'primary.foreground',
      border: '1px solid',
      borderColor: 'background',
      lineHeight: '1.215',
      display: 'flex',
      alignItems: 'center',

      _selected: {
        bg: 'background',
        color: 'foreground',
      },
    },
    textInput: {
      display: 'block',
      width: '32',
      height: '100%',
    },
    close: {
      p: 1,
      ml: 1,
    },
  },
})
