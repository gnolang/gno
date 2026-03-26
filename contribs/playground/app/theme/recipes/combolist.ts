import { defineSlotRecipe } from '@pandacss/dev'

export const comboListRecipe = defineSlotRecipe({
  className: 'combo-list',
  slots: ['container', 'item', 'label', 'indicator', 'action', 'actions'],
  base: {
    container: {},
    item: {
      position: 'relative',
      display: 'flex',
      flexDirection: 'row',
      alignItems: 'center',
      marginBottom: '12px',
      '&:last-child': {
        marginBottom: '0',
      },
    },
    label: {
      overflow: 'hidden',
      textOverflow: 'ellipsis',
      whiteSpace: 'nowrap',
      flex: '1 1 auto',
      textAlign: 'left',
      border: 'none',
      background: 'none',
      '&:focus-visible': {
        border: 'none',
        outline: '2px dashed',
      },
    },
    indicator: {
      marginRight: '8px',
    },
    actions: {
      display: 'flex',
      flexDirection: 'row',
      alignItems: 'center',
    },
    action: {
      border: 'none',
      background: 'none',
      color: 'gray.400',
      marginLeft: '8px',
      '&:hover, &:active, &:focus': {
        color: 'currentColor',
      },
      '&:focus-visible': {
        border: 'none',
        outline: '2px dashed',
      },
    },
  },
  variants: {
    /**
     * Display secondary actions above the text.
     */
    actionsOverflow: {
      true: {
        item: {
          position: 'relative',
        },
        actions: {
          position: 'absolute',
          top: 0,
          right: 0,
          bottom: 0,
        },
      },
    },
  },
})
