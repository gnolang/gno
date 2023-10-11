# Gnolang examples

This folder showcases Gnolang realms and library demos. These examples not only aid in engine testing but also provide a glimpse into the potential of Gnolang's capabilities.

While sharing contracts here can enhance engine testing, it's not mandatory. If considering a separate repository for contracts, be aware that this might restrict the experience due to the continuous efforts around `gno mod` support. A key point to note is that the main repository cannot reference separate code, which might pose developmental challenges.

## Personal Realms & Shared Content

**Prioritizing Shared Content:** As we expand our examples and use-cases, it's essential to prioritize shared content that benefits the broader community. These examples serve as a foundation and reference for all users.

**Personal Realms Inclusion:** We're open to personal realms, but they must exemplify best practices and inspire others. To maintain our repository's organization, we may decline some realms. If so, consider uploading onchain and keeping source code separately. For higher acceptance odds, offer useful or original examples.

**Recommended Approach:** 
- Use `r/demo` and `p/demo` for generic examples and components that can be imported by others. These are meant to be easily referenced and utilized by the community.
- Personal realms are welcomed if they are easily maintainable with the Continuous Integration (CI) system. If a personal realm becomes cumbersome to maintain or doesn't align with the CI's checks, it might be relocated to a less prominent location or even removed. 

## Usage

Our recommendation is to use the [gno](../gnovm/cmd/gno) utility to develop contracts locally before publishing them on-chain. This approach offers a faster and streamlined workflow, along with additional debugging features. Simply fork or create new contracts and refer to the Makefile. Once everything looks good locally, you can then publish it on a localnet or testnet.

For further guidance and insights, please refer to the [`awesome-gno` tutorials](https://github.com/gnolang/awesome-gno#tutorials).
