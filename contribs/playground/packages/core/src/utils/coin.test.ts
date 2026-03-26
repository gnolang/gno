import { expect, it } from 'vitest'

import { convertDecimalToUnit, convertUnitToDecimal, parseBalanceToCoins } from './coin'

it('should return the correct amount and denom', () => {
  expect(parseBalanceToCoins('1000uatom')).toEqual([{ amount: '1000', denom: 'uatom' }])
  expect(parseBalanceToCoins('1000')).toEqual([{ amount: '1000', denom: '' }])
  expect(parseBalanceToCoins('uatom')).toEqual([{ amount: '0', denom: 'uatom' }])
  expect(parseBalanceToCoins('123uatom,456ugnot')).toEqual([
    { amount: '123', denom: 'uatom' },
    { amount: '456', denom: 'ugnot' },
  ])
  expect(parseBalanceToCoins('123456789ibc/7F1D3FCF4AE79E1554D670D1AD949A9BA4E4A3C76C63093E17E446A46061A7A1')).toEqual([
    {
      amount: '123456789',
      denom: 'ibc/7F1D3FCF4AE79E1554D670D1AD949A9BA4E4A3C76C63093E17E446A46061A7A1',
    },
  ])
})

it('should return empty when parse fails', () => {
  expect(parseBalanceToCoins('1.000uatom1000uatom')).toEqual([])
})

it('should convert the unit to decimal', () => {
  expect(convertUnitToDecimal('1', 'uatom')).toBe('0.00000001')
  expect(convertUnitToDecimal('123', 'uosmo')).toBe('0.00000123')
  expect(convertUnitToDecimal('1000', 'ugnot')).toBe('0.001')
  expect(convertUnitToDecimal('1000', 'unknown')).toBe('0.00001')
})

it('should convert the decimal to unit', () => {
  expect(convertDecimalToUnit('0.00000001', 'uatom')).toBe('1')
  expect(convertDecimalToUnit('0.00000123', 'uosmo')).toBe('123')
  expect(convertDecimalToUnit('0.001', 'ugnot')).toBe('1000')
  expect(convertDecimalToUnit('0.00001', 'unknown')).toBe('1000')
})
