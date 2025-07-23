This package exists because we add many things into `testing` beyond Go's
`testing` module, and this creates circular dependencies in our own standard
libraries. Standard library packages like crypto/bech32 need to import
`testing/base` instead of `testing`.
