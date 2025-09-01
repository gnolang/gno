package grc20

// REVIEW: conversion to julia w/ Algebraic Petri Nets
// https://gist.github.com/stackdump/33dcd8e6f071d391c998008dfeac6ed9

var pflowLink = `
[![pflow](https://pflow.dev/img/zb2rhdVMFVdcBLYqpMJh8Vk4awKUf5GeWsuURmeTxRghdAmQx.svg)](https://pflow.dev/p/zb2rhdVMFVdcBLYqpMJh8Vk4awKUf5GeWsuURmeTxRghdAmQx/)
`

const version ="bafkreie4ceqfi5rnae75u2dp22utnwr6yrhvz5p5uvalgrrxfwwrvet2nq"

func init() {
    m := GRC20(modelOpts)
    if m.Cid() != version {
        panic("model version mismatch")
    }
}


func Render(path string) string {
	m := GRC20(modelOpts)
	return pflowLink + "\n\n Model ID: " + m.Cid() + "\n\n"  + m.IdentityHash() + "\n\n" + m.Markdown()
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
