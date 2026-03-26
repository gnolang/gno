import { BigNumber } from 'bignumber.js'

import { CURRENCIES } from '../config'

// Prevent exponential notation
BigNumber.config({ EXPONENTIAL_AT: [-20, 24] })

interface Coin {
  amount: string
  denom: string
}

export function parseBalanceToCoins(balanceText: string): Coin[] {
  return balanceText
    .replace(/\s/g, '')
    .split(',')
    .filter(Boolean)
    .map((part) => {
      const match = part.match(/^(\d+)?([a-zA-Z0-9]+(?:\/[a-zA-Z0-9]+)*)?$/)

      if (!match) return null

      const amount = match[1] || '0'
      const denom = match[2] || ''

      return { amount, denom }
    })
    .filter(Boolean) as Coin[]
}

function intlNumberFormat(value: string, unit: number): string {
  return new Intl.NumberFormat('en-US', {
    minimumFractionDigits: unit,
    maximumFractionDigits: unit,
  }).format(Number(value))
}

function mapDenomToCurrency(denom: string): string {
  return CURRENCIES[denom]?.currency ?? ''
}

export function convertUnitToDecimal(amount: string, denom: string) {
  const unit = CURRENCIES[denom]?.unit ?? 8
  return new BigNumber(amount).shiftedBy(-unit).decimalPlaces(unit).toString()
}

export function convertDecimalToUnit(amount: string, denom: string) {
  const unit = CURRENCIES[denom]?.unit ?? 8
  return new BigNumber(amount).shiftedBy(unit).decimalPlaces(unit).toString()
}

export function formatUnitAmount(amount: string, denom: string) {
  const unit = CURRENCIES[denom]?.unit ?? 8
  const currency = mapDenomToCurrency(denom)
  const decimal = convertUnitToDecimal(amount, denom)

  return `${intlNumberFormat(decimal, unit)} ${currency || 'Unknown'}`
}

export function getGnotCoin(balance: string): Coin | undefined {
  const coins = parseBalanceToCoins(balance)
  const defaultCoin = CURRENCIES.ugnot

  return coins.find((coin) => coin.denom === defaultCoin.denom)
}

export function formatBalanceAmount(balance: string): string[] {
  const coins = parseBalanceToCoins(balance)

  return coins.map((coin) => formatUnitAmount(coin.amount, coin.denom))
}
