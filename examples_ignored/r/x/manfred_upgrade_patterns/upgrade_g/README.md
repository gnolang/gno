# `upgrade_g` -- versioned redeploy (PR #6253)

The same path, redeployed. No wrapper, proxy, registry or copy: `v1/` and
`v2/` both declare `module = ".../upgrade_g/vault"`, and the chain accepts v2
over the live v1 because `version` steps by one and the sender is the
`[upgrade] authority`. Globals are carried by name, declared types keep their
identity, `migrate(from int)` runs in place of `init()`, and importers are
re-preprocessed in the same transaction.

Filetests cannot redeploy, so this pattern is executed by
`gno.land/pkg/integration/testdata/redeploy_public_authority.txtar` and
`redeploy_public_importers.txtar`. The sequence they run, read here:

```
addpkg v1 as vault             total = 0
addpkg client                  client imports vault
client.Deposit(5)              vault.Total() = 5, client.Last() = "receipt for 5 (v1)"
addpkg v2 as vault             carried: total = 5; migrate(1): count = 1
vault.Count()                  1
client.Deposit(5)              v2 credits double: vault.Total() = 15
client.Last()                  "receipt for 10 (v2)"
client.First()                 "receipt for 5 (v2)"   the receipt v1 persisted, v2's method
restart the node               the same answers
addpkg v3 without [upgrade]    accepted once, then the path is permanent
```

Refused, with the realm left as it was: a `version` that does not step by
one; a stranger as sender; a global retyped; a declared type removed or
reshaped; for a public realm, any declaration or method moved or removed.
