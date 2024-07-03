// the boot package is similar to linux' /boot/cmdline.
//
// it contains values that were used to initialize the chain. as a pure-package,
// its content is immutable, which makes this package having a high potential of
// quickly becoming outdated. however, it can be used safely by initial
// contracts' init() functions.
//
// per-context expected usages:
// - gnodev:
//   - register current node's on poa
//   - dynamically load local gnokey's information
// - gnoland: static file manually patched by the chain architects.
package boot

/*
  {{.}}
*/