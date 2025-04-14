# Example of Usage

This directory contains three versions of a simple realm. When the first is installed, `Render()` can be called on it to see a result.

When the `v2` version of the realm is installed, the `Render()` in the original `v1` version will immediately start running the version introduced in `v2`.

When the `v3` version of the realm is subsequently installed, another upgrade will take place, and the `v2` code will be superceded by the `v3` code.

```go
gnokey maketx addpkg \
  -pkgpath gno.land/r/yournamespace/v1/mywebsite \
  -pkgdir v1 \
  -gas-fee 1000000ugnot \
  -gas-wanted 30000000 \
  -broadcast \
  -chainid dev \
  YourKey
```

```go
gnokey maketx call \
  -pkgpath "gno.land/p/yournamespace/v1/mywebsite" \
  -func "Render" \
  -args "abc123" \
  -gas-fee 1000000ugnot \
  -gas-wanted 3000000 \
  -broadcast \
  -chainid dev \
  YourKey
```

```go
gnokey maketx addpkg \
  -pkgpath gno.land/r/yournamespace/v2/mywebsite \
  -pkgdir v2 \
  -gas-fee 1000000ugnot \
  -gas-wanted 30000000 \
  -broadcast \
  -chainid dev \
  YourKey
```

```go
gnokey maketx call \
  -pkgpath "gno.land/p/yournamespace/v1/mywebsite" \
  -func "Render" \
  -args "abc123" \
  -gas-fee 1000000ugnot \
  -gas-wanted 3000000 \
  -broadcast \
  -chainid dev \
  YourKey
```

```go
gnokey maketx addpkg \
  -pkgpath gno.land/r/yournamespace/v3/mywebsite \
  -pkgdir v3 \
  -gas-fee 1000000ugnot \
  -gas-wanted 30000000 \
  -broadcast \
  -chainid dev \
  YourKey
```

```go
gnokey maketx call \
  -pkgpath "gno.land/p/yournamespace/v1/mywebsite" \
  -func "Render" \
  -args "abc123" \
  -gas-fee 1000000ugnot \
  -gas-wanted 3000000 \
  -broadcast \
  -chainid dev \
  YourKey
```
