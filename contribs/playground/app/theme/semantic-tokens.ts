import { defineSemanticTokens } from '@pandacss/dev'

export const semanticTokens = defineSemanticTokens({
  colors: {
    background: {
      value: {
        base: '{colors.white}',
        _dark: '{colors.gray.950}',
      },
    },
    foreground: {
      DEFAULT: {
        value: { base: '{colors.black}', _dark: '{colors.gray.200}' },
      },
      muted: {
        value: { base: '{colors.gray.500}', _dark: '{colors.gray.400}' },
      },
    },
    border: {
      DEFAULT: {
        value: { base: '{colors.gray.400}', _dark: '{colors.gray.300}' },
      },
      highlight: {
        value: { base: '{colors.blue.500}', _dark: '{colors.blue.500}' },
      },
    },
    header: {
      value: { base: '{colors.gray.200}', _dark: '{colors.gray.500}' },
    },
    link: {
      value: { base: 'black', _dark: '{colors.gray.200}' },
    },
    input: {
      DEFAULT: {
        value: { base: '{colors.black}', _dark: '{colors.white}' },
      },
      foreground: {
        value: { base: '{colors.gray.200}', _dark: '{colors.gray.950}' },
      },
      placeholder: {
        value: { base: '{colors.gray.500}', _dark: '{colors.gray.300}' },
      },
    },
    primary: {
      DEFAULT: {
        value: { base: '{colors.gray.300}', _dark: '{colors.gray.500}' },
      },
      foreground: {
        value: { base: '{colors.black}', _dark: '{colors.gray.200}' },
      },
    },
  },
})
