## Guide

### Getting started with ViM

Add to your .vimrc file:

```vim
function! GnoFmt()
	cexpr system('gofmt -e -w ' . expand('%')) "or replace with gofumpt
	edit!
endfunction
command! GnoFmt call GnoFmt()
augroup gno_autocmd
	autocmd!
	autocmd BufNewFile,BufRead *.gno set filetype=go
	autocmd BufWritePost *.gno GnoFmt
augroup END
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

- [#208](https://github.com/gnolang/gno/pull/208) - @anarcher, gnodev test with testing.T
- [#167](https://github.com/gnolang/gno/pull/167) - @loicttn, website: Add syntax highlighting + security practices
- [#136](https://github.com/gnolang/gno/pull/136) - @moul, foo20, a grc20 example smart contract
- [#126](https://github.com/gnolang/gno/pull/126) - @moul, feat: use the new Precompile in gnodev and in the addpkg/execution flow (2/2)
- [#119](https://github.com/gnolang/gno/pull/119) - @moul, add a Gno2Go precompiler (1/2)
- [#112](https://github.com/gnolang/gno/pull/112) - @moul, feat: add 'gnokey maketx --broadcast' option
- [#110](https://github.com/gnolang/gno/pull/110), [#109](https://github.com/gnolang/gno/pull/109), [#108](https://github.com/gnolang/gno/pull/108), [#106](https://github.com/gnolang/gno/pull/106), [#103](https://github.com/gnolang/gno/pull/103), [#102](https://github.com/gnolang/gno/pull/102), [#101](https://github.com/gnolang/gno/pull/101) - @moul, various chores.
