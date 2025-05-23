# Examples

This folder showcases example Gno realms (smart contracts) and pure packages (libraries).
These examples provide a glimpse into the potential of gno.land and the capabilities of Gno,
while also serving as a test suite for the GnoVM.

Pure packages and realms in this folder are pre-deployed to gno.land testnets,
making them readily available for on-chain use. However, **there is no guarantee
that the code is bug-free, so it should be used with caution and an understanding of potential risks.**

## Structure

This folder mimics the gno.land package path system; the "root" of the system is
the `gno.land` folder. Next, it branches out to `p/` and `r/`, which contain
pure packages and realms, respectively.

## Personal Realms & Shared Content

**Prioritizing Shared Content:** As we expand our examples and use-cases, it's
essential to prioritize shared content that benefits the broader community.
These examples serve as a foundation and reference for all users.

**Personal Realms & Pure Packages:** We welcome personal realms that
exemplify best practices and inspire others. To maintain the organization
of the monorepo, some submissions may be declined. If so, consider uploading
[permissionlessly](../docs/users/interact-with-gnokey.md#addpackage)
and storing the source code in a separate repo. For higher
acceptance odds, offer useful and original examples.

**Recommended Approach:**
- Use `r/demo` and `p/demo` for generic examples and components that can be
  imported by others. These are meant to be easily referenced and utilized by the
  community.
- Packages under personal namespaces, such as in [r/leon](./gno.land/r/leon),
  are welcome if they are easily maintainable with the Continuous Integration (CI)
  system. If a personal realm becomes cumbersome to maintain or doesn't align with
  the CI's checks, it might be relocated to a less prominent location or even removed. 