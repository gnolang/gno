## Guide

### Getting started with ViM

Add to your .vimrc file:

```vim
au BufRead,BufNewFile *.gno set filetype=go
```

TODO: other vim tweaks to make work with vim-go etc.

### Getting started with Emacs

Install [go-mode.el](https://github.com/dominikh/go-mode.el).

Add to your emacs configuration file:

```lisp
(add-to-list 'auto-mode-alist '("\\.gno\\'" . go-mode))
```

## Notable Contributions

Notable contributions of fixes/features/refactors:

* [https://github.com/gnolang/gno/pull/208](#208) - @anarcher, gnodev test with testing.T
* [https://github.com/gnolang/gno/pull/167](#167) - @loicttn, website: Add syntax highlighting + security practices
* [https://github.com/gnolang/gno/pull/136](#136) - @moul, foo20, a grc20 example smart contract
* [https://github.com/gnolang/gno/pull/126](#126) - @moul, feat: use the new Precompile in gnodev and in the addpkg/execution flow (2/2)
* [https://github.com/gnolang/gno/pull/119](#119) - @moul, add a Gno2Go precompiler (1/2)
* [https://github.com/gnolang/gno/pull/112](#112) - @moul, feat: add 'gnokey maketx --broadcast' option
* [https://github.com/gnolang/gno/pull/110](#110), [https://github.com/gnolang/gno/pull/109](#109), [https://github.com/gnolang/gno/pull/108](#108), [https://github.com/gnolang/gno/pull/106](#106), [https://github.com/gnolang/gno/pull/103](#103), [https://github.com/gnolang/gno/pull/102](#102), [https://github.com/gnolang/gno/pull/101](#101) - @moul, various chores.
