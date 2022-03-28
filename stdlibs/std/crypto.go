package std

type Address string // NOTE: bech32

const RawAddressSize = 20

type RawAddress [RawAddressSize]byte
