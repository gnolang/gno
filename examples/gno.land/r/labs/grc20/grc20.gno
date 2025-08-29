package grc20

// REVIEW: conversion to julia w/ Algebraic Petri Nets
// https://gist.github.com/stackdump/33dcd8e6f071d391c998008dfeac6ed9

var pflowLink = `
[![pflow](https://pflow.dev/img/zb2rhdVMFVdcBLYqpMJh8Vk4awKUf5GeWsuURmeTxRghdAmQx.svg)](https://pflow.dev/p/zb2rhdVMFVdcBLYqpMJh8Vk4awKUf5GeWsuURmeTxRghdAmQx/)`

func Render(path string) string {
	m := GRC20(modelOpts)
	return pflowLink + "\n\n Model ID: " + m.Cid() + "\n\n" + m.Markdown()
}

/*
func (tok Token) RenderHome() string {
    return ufmt.Sprintf(`# %s ($%s)
* **Version**: v0
* **Decimals**: %d
* **Total supply**: %d
* **Known accounts**: %d
* **Model digest**: %s
* **Proof digest**: %s
`, tok.name, tok.symbol, tok.decimals, tok.ledger.totalSupply, tok.KnownAccounts(),
   tok.ModelDigest(), tok.ProofDigest())
}
*/
