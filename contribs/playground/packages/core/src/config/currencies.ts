interface CurrencyInfo {
  currency: string
  denom: string
  unit: number
}

export const CURRENCIES: Record<string, CurrencyInfo> = {
  ugnot: {
    currency: 'GNOT',
    denom: 'ugnot',
    unit: 6,
  },
} as const

export const FeeToken = CURRENCIES.ugnot
