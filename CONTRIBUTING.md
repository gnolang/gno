[![GitHub repo Good Issues for newbies](https://img.shields.io/github/issues/gnolang/gno/good%20first%20issue?style=flat&logo=github&logoColor=green&label=Good%20First%20issues)](https://github.com/gnolang/gno/issues?q=is%3Aopen+is%3Aissue+label%3A%22good+first+issue%22) [![GitHub Help Wanted issues](https://img.shields.io/github/issues/gnolang/gno/help%20wanted?style=flat&logo=github&logoColor=b545d1&label=%22Help%20Wanted%22%20issues)](https://github.com/gnolang/gno/issues?q=is%3Aopen+is%3Aissue+label%3A%22help+wanted%22) [![GitHub Help Wanted PRs](https://img.shields.io/github/issues-pr/gnolang/gno/help%20wanted?style=flat&logo=github&logoColor=b545d1&label=%22Help%20Wanted%22%20PRs)](https://github.com/gnolang/gno/pulls?q=is%3Aopen+is%3Aissue+label%3A%22help+wanted%22) [![GitHub repo Issues](https://img.shields.io/github/issues/gnolang/gno?style=flat&logo=github&logoColor=red&label=Issues)](https://github.com/gnolang/gno/issues?q=is%3Aopen)

# Contributing to Gno

Thank you for looking to contribute to the Gno project.
We appreciate every open-source contribution, as it helps us improve and enhance Gno for the benefit of the community.

This document outlines some basic pointers on making your future contribution a great experience. It outlines basic PR
etiquette employed by the core gno team. It lays out coding styles, simple how-to guides and tools to get you up and
running and pushing code.

If you are unsure about something, please don’t hesitate to reach out for help by opening an issue here or discuss on
Discord.
Likewise, if you have an idea on how to improve this guide, go for it as well.

## Table of Contents

- [Important Resources](#important-resources)
- [Getting Started](#getting-started)
    - [Environment](#environment)
    - [Local Setup](#local-setup)
    - [Testing](#testing)
      - [Running locally](#running-locally)
      - [Running test workflows](#running-test-workflows)
      - [Testing Gno code](#testing-gno-code)
    - [Repository Structure](#repository-structure)
- [How do I?](#how-do-i)
    - [How do I submit changes?](#how-do-i-submit-changes)
      - [A Word on Rebasing](#a-word-on-rebasing)
    - [How do I report a bug?](#how-do-i-report-a-bug)
    - [How do I request a feature?](#how-do-i-request-a-feature)
- [Style Guides](#style-guides)
    - [Git Commit Messages](#git-commit-messages)
    - [Go/Gno Style Guide](#gogno-style-guide)
    - [Documentation Style Guide](#documentation-style-guide)
- [Additional Notes](#additional-notes)
    - [Issue and Pull Request Labels](#issue-and-pull-request-labels)
      - [Labels for Pull Requests](#labels-for-pull-requests)
      - [Labels for Issues](#labels-for-issues)
    - [Notable Contributions](#notable-contributions)

## Important Resources

- **[Discord](https://discord.gg/YFtMjWwUN7)** - we are very active on Discord. Join today and start discussing all
  things gno with fellow engineers and enthusiasts.
- **[Awesome Gno](https://github.com/gnolang/awesome-gno)** - check out the list of compiled resources for helping you
  understand the gno ecosystem
- **[Active Staging](https://staging.gno.land/)** - use the currently available staging environment to play around with a
  production network. If you want to interact with a local instance, refer to the [Local Setup](#local-setup) guide.
- **[Twitter](https://twitter.com/_gnoland)** - follow us on Twitter to get the latest scoop
- **[Telegram](https://t.me/gnoland)** - join our official Telegram group to start a conversation about gno

## Getting Started

### Environment

The gno repository is primarily based on Go (Golang) and Gno.

The primary tech stack for working on the repository:

- Go (version 1.22+)
- make (for using Makefile configurations)

It is recommended to work on a Unix environment, as most of the tooling is built around ready-made tools in Unix (WSL2
for Windows / Linux / macOS).

For Gno, there is no specific tooling that needs to be installed, that’s not already provided with the repo itself.
You can utilize the `gno` command to facilitate Gno support when writing Smart Contracts in Gno, by installing it
with `make install_gno`.

If you are working on Go source code on this repository, `pkg.go.dev` will not
render our documentation as it has a license it does not recognise. Instead, use
the `go doc` command, or use our statically-generated documentation:
https://gnolang.github.io/gno/github.com/gnolang/gno.html

Additionally, you can also configure your editor to recognize `.gno` files as `.go` files, to get the benefit of syntax
highlighting.

#### Visual Studio Code

There currently is an unofficial [Visual Studio Code](https://marketplace.visualstudio.com/items?itemName=harry-hov.gno)
extension (primarily developed by a core team member) for working with `*.gno`
files.

#### ViM Support (without LSP)

Add to your `.vimrc` file:

```vim
function! GnoFmt()
	cexpr system('gofmt -e -w ' . expand('%')) " or replace with gofumpt, see below
	edit!
	set syntax=go
endfunction
command! GnoFmt call GnoFmt()
augroup gno_autocmd
	autocmd!
	autocmd BufNewFile,BufRead *.gno set syntax=go
	autocmd BufWritePost *.gno GnoFmt
augroup END
```

To use *gofumpt* instead of *gofmt*, as hinted in the comment, you may either have `gofumpt` in your PATH or substitute the cexpr line above with the following (please make sure to replace `<path/to/gno>` with the path to your local gno repository):

```vim
cexpr system('go run -modfile </path/to/gno>/misc/devdeps/go.mod mvdan.cc/gofumpt -w ' . expand('%'))
```

##### ViM Linting Support

To integrate GNO linting in Vim, you can use Vim's `:make` command with a custom `makeprg` and `errorformat` to run the GNO linter and parse its output. Add the following configuration to your `.vimrc` file:

```vim
autocmd FileType gno setlocal makeprg=gno\ lint\ %
autocmd FileType gno setlocal errorformat=%f:%l:\ %m

" Optional: Key binding to run :make on the current file
autocmd FileType gno nnoremap <buffer> <F5> :make<CR>
```

### ViM Support (with LSP)

There is an experimental and unofficial [Gno Language Server](https://github.com/jdkato/gnols)
developed by the community, with an installation guide for Neovim.

For ViM purists, you have to install the [`vim-lsp`](https://github.com/prabirshrestha/vim-lsp)
plugin and then register the LSP server in your `.vimrc` file:

```vim
augroup gno_autocmd
    autocmd!
    autocmd BufNewFile,BufRead *.gno
        \ set filetype=gno |
        \ set syntax=go
augroup END

if (executable('gnols'))
    au User lsp_setup call lsp#register_server({
        \ 'name': 'gnols',
        \ 'cmd': ['gnols'],
        \ 'allowlist': ['gno'],
        \ 'config': {},
        \ 'workspace_config': {
        \   'root' : '/path/to/gno_repo',
        \	'gno'  : '/path/to/gno_bin',
        \   'precompileOnSave' : v:true,
        \   'buildOnSave'      : v:false,
        \ },
        \ 'languageId': {server_info->'gno'},
    \ })
else
	echomsg 'gnols binary not found: LSP disabled for Gno files'
endif

function! s:on_lsp_buffer_enabled() abort
    " Autocompletion
    setlocal omnifunc=lsp#complete
    " Format on save
    autocmd BufWritePre <buffer> LspDocumentFormatSync
    " Some optional mappings
    nmap <buffer> <leader>i <Plug>(lsp-hover)
    " Following mappings are not supported yet by gnols
    " nmap <buffer> gd <plug>(lsp-definition)
    " nmap <buffer> <leader>rr <plug>(lsp-rename)
endfunction
augroup lsp_install
    au!
    autocmd User lsp_buffer_enabled call s:on_lsp_buffer_enabled()
augroup END
```

Note that unlike the previous ViM setup without LSP, here it is required by
`vim-lsp` to have a specific `filetype=gno`. Syntax highlighting is preserved
thanks to `syntax=go`.

Inside `lsp#register_server()`, you also have to replace
`workspace_config.root` and `workspace_config.gno` with the correct directories
from your machine.

Additionally, it's not possible to use `gofumpt` for code formatting with
`gnols` for now.

#### Emacs Support

1. Install [go-mode.el](https://github.com/dominikh/go-mode.el).
2. Add to your emacs configuration file:

```lisp
(define-derived-mode gno-mode go-mode "GNO"
  "Major mode for GNO files, an alias for go-mode."
  (setq-local tab-width 8))
(define-derived-mode gno-dot-mod-mode go-dot-mod-mode "GNO Mod"
  "Major mode for GNO mod files, an alias for go-dot-mod-mode."
  )
```

3. To integrate GNO linting with Flycheck, add the following to your Emacs configuration:
```lisp
(require 'flycheck)

(flycheck-define-checker gno-lint
  "A GNO syntax checker using the gno lint tool."
  :command ("gno" "lint" source-original)
  :error-patterns (;; ./file.gno:32: error message (code=1)
                   (error line-start (file-name) ":" line ": " (message) " (code=" (id (one-or-more digit)) ")." line-end))
  ;; Ensure the file is saved, to work around
  ;; https://github.com/python/mypy/issues/4746.
  :predicate (lambda ()
               (and (not (bound-and-true-p polymode-mode))
                    (flycheck-buffer-saved-p)))
  :modes gno-mode)

(add-to-list 'flycheck-checkers 'gno-lint)
```

#### Sublime Text

There is an experimental and unofficial [Gno Language Server](https://github.com/jdkato/gnols)
developed by the community, with an installation guide for Sublime Text.

### Local Setup

To get started with Gno development, the process is relatively straightforward.

Clone the repo:
`git clone https://github.com/gnolang/gno.git`

Build / install base commands:
`make install`

If you haven't already, you may need to add the directory where [`go install`
places its binaries](https://pkg.go.dev/cmd/go#hdr-Compile_and_install_packages_and_dependencies)
to your `PATH`. If you haven't configured `GOBIN` or `GOPATH` differently, this
command should suffice:

```
echo 'export PATH="$HOME/go/bin:$PATH"' >> ~/.profile
source ~/.profile # reload ~/.profile in the current shell
```

After that, you should be good to go to use `gno` and `gnokey`, straight from
your command line! The following commands should list the help messages for
each:

```console
$ gno --help
USAGE
  <subcommand> [flags] [<arg>...]

Runs the gno development toolkit
[...]
$ gnokey --help
USAGE
  <subcommand> [flags] [<arg>...]

Manages private keys for the node
[...]
```

### Testing

There are essentially 2 ways to execute the entire test suite:

1. Using the base `go test` commands (running locally),
2. Using a tool like [act](https://github.com/nektos/act) to run workflow files in Docker containers

#### Running locally

To run the entire test suite locally, run the following command:

    make test

This will execute the full test suite, that includes tests for `.gno` files, as well as project `.go` tests.

#### Running test workflows

Using a tool like [act](https://github.com/nektos/act) can enable you to run specific repository workflow files locally,
using Docker. The workflow configurations contain different `go` versions, so it might be worth running if you’re
worried about compatibility.

To run the entire test suite through workflow files, run the following command:

    act -v -j go-test

#### Testing Gno code

If you wish to test a `.gno` Realm or Package, you can utilize the `gno` tool.

1. To install it, simply run:

    make install_gno

2. Now, you can point to the directory containing the `*_test.gno` files:

    gno test <path-to-dir> -v


To learn more about how `gno` can help you when developing gno code, you can look into the available
subcommands by running:

    gno --help

#### Adding new tests

Most packages will follow the convention established with Go: each package
contains within its file many files suffixed with `_test.go` which test its
functionality. As a general rule, you should follow this convention, and in
every PR you make you should ensure all the code you added is appropriately
covered by tests ([Codecov](https://about.codecov.io/) will loudly complain in
your PR's comments if you don't).

Additionally, we have a few testing systems that stray from this general rule;
at the time of writing, these are for integration tests and language tests. You
can find more documentation about them [on this guide](docs/testing-guide.md).

### Repository Structure

The repository structure can seem tricky at first, but it’s simple if you consider the philosophy that the gno project
employs (check out [PHILOSOPHY.md](./PHILOSOPHY.md)).

The gno project currently favors a mono-repo structure, as it’s easier to manage contributions and keep everyone
aligned. In the future, this may change, but in the meantime the majority of gno resources and source code will be
centralized here.

- `examples` - contains the example `.gno` realms and packages. This is the central point for adding user-defined realms
  and packages.
- `gno.land` - contains the base source code for bootstrapping the Gnoland node,
  using `tm2` and `gnovm`.
- `gnovm` - contains the implementation of the Gno programming language and its
  Virtual Machine, together with their standard libraries and tests.
- `tm2` - contains a fork of the [Tendermint consensus engine](https://docs.tendermint.com/v0.34/introduction/what-is-tendermint.html) with different expectations.

## How do I?

### How do I submit changes?

Pull Requests serve primarily as an addition to the codebase, but also as a central point of discussion for potential features.
Instead of opening issues, contributors are encouraged to open PRs with their suggested changes, however small, in order to
get feedback from the community on their work.

The key to successfully submitting upstream changes is providing the adequate context, with the correct implementation,
of course.

To let the maintainers know about _what_ your PR does, and _why_ it does it, you should take a second to fill out the PR
template provided on the repo.

Once someone leaves a PR review (with open comments / discussions), it is on the PR _creator_ to do their best in trying
to resolve all comments.
The PR _reviewer_ is in charge of resolving each conversation (hitting the `Resolve Conversation` button), after the PR
_creator_ implements or discusses
the requested changes.

If the PR _creator_ pushed code as a result of some conversation, they should link the commit hash in the relevant
comment reply, ex:

```Markdown
Fixed in:

[0a0577c](https://github.com/gnolang/gno/commit/0a0577ccdeb951a6621d6fbe1c04ac4e64a529c1)
```

#### A Word on Rebasing

Avoid rebasing after you open your PRs to reviews. Instead, add more commits to your PR.
It's OK to do force pushes if PR was never opened for reviews before.

A reviewer may like to see a linear commit history while reviewing. If you tend to force push from an older commit,
a reviewer might lose track in your recent changes and will have to start reviewing from scratch.

Don't worry about adding too many commits. The commits are squashed into a single commit while merging (if needed).

#### PR template

The [PR template](https://github.com/gnolang/gno/blob/master/.github/pull_request_template.md) is by default a
simple template meant to illustrate quickly what’s the context of the PR.

If you've run a manual test, please provide the exact steps taken.

### How do I report a bug?

Found something funky while using gno? You can report it to the team using GitHub issues.

Before opening an issue, please check for existing open and closed Issues to see if that bug/feature has already been
reported/requested. If you find a relevant topic, you can comment on that issue.

In the issue, please provide as much information as possible, for example:

- what you’ve experienced
- what should be the expected behavior
- which version of gno are you running
- any logs that might be relevant

Once you open up an issue, a core team member will take a look and start a conversation with you about resolving it.

### How do I request a feature?

The gno project operates in the public. This means that feature requests are not just reviewed by the core team, but
also by the general user base and other contributors.

If you would like to propose a new feature (submit a feature request), you can do so by opening up an issue on GitHub,
and providing ample context as to why this feature should be implemented. Even better, if you are able to propose your
feature request in code, you should open a PR instead. This allows the community to easily discuss and build upon the idea
in a central place.

Additionally, it is encouraged that users suggest their idea on the official communication channels like Discord, to get
a feel for what other contributors think.

## Style Guides

### Git Commit Messages

The gno project tends to use the [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) standard for any
commit that goes into the main code stream (currently, the `master` branch).

Each PR is squashed and merged into the main code stream, which means PR _titles_ should adhere to the Conventional
Commit standard, and be short and precise. This is practically enforced using a
[linter check](https://github.com/amannn/action-semantic-pull-request). The type
of PRs that are allowed are the following:

* **build**\
  Changes that affect the build system or external dependencies.
* **chore**\
  Other changes that don't modify source or test files.
* **ci**\
  Changes to our CI configuration files (ie. GitHub Actions) and scripts.
* **docs**\
  Documentation only changes.
* **feat**\
  A new feature.
* **fix**\
  A bug fix.
* **perf**\
  A code change that improves performance.
* **refactor**\
  A code change that neither fixes a bug nor adds a feature.
* **revert**\
  Reverts a previous commit.
* **style**\
  Changes that do not affect the meaning of the code (white-space, formatting, missing semi-colons, etc).
* **test**\
  Adding missing tests or correcting existing tests.

A general rule of thumb:

- Never favor rewriting history in PRs (rebases have very few exceptions, like implementation rewrites;
see [A Word on Rebasing](#a-word-on-rebasing))
- Tend to make separate commits for logically separate changes

### Go/Gno Style Guide

The gno codebase currently doesn’t follow any specific standard in terms of code style. The core team, and core
contributors, favor using the following code style guides:

- [Effective Go](https://go.dev/doc/effective_go)
- [Uber’s Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)
- [Google’s Go Style Guide](https://google.github.io/styleguide/go/guide)
  and [Google’s Go Best Practices](https://google.github.io/eng-practices/review/)

If unsure, you should always follow the existing code style of the repository.

Additionally, the gno codebase uses linters to enforce some common coding style etiquette. Any PR that aims to modify
the gno codebase should make sure that the linter checks pass; otherwise they won’t be merged into the main code stream.

### Documentation Style Guide

When writing in-code documentation, always favor to stay aligned with the godoc standard. Gno’s development philosophy
stays true to idiomatic Go, and this transfers to the documentation as well.

Resources for idiomatic Go docs:

- [godoc](https://go.dev/blog/godoc)
- [Go Doc Comments](https://tip.golang.org/doc/comment)

## Additional Notes

### Issue and Pull Request Labels

The gno project uses a set of labels for managing ongoing issues and pull requests.

If you’d like to modify the current label structure, please submit a PR modifying
the [labels.json](https://github.com/gnolang/gno/blob/master/.github/labels.json) file. The gno project utilizes
automatic label management.

#### Labels for Pull Requests

| Label Name      | Description                                                     |
|-----------------|-----------------------------------------------------------------|
| breaking change | PR introduces backwards incompatible functionality              |
| bug fix         | PR introduces a fix for a known bug                             |
| dependencies    | PR introduces a package version bump                            |
| documentation   | PR introduces an improvement or addition to the docs            |
| don’t merge     | PR contains unstable functionality that shouldn't be merged yet |
| feature         | PR adds new functionality to Gno                                |
| hotfix          | PR applies a hotfix that should be merged ASAP                  |

#### Labels for Issues

| Label Name       | Description                                       |
|------------------|---------------------------------------------------|
| bug              | Issue confirms a bug is present in Gnoland        |
| duplicate        | Issue already exists, or has been posted before   |
| good first issue | Issue is a great introduction for newcomers       |
| help wanted      | Issue requires extra attention from the community |
| info needed      | Issue is lacking information needed for resolving |
| investigating    | Issue is still being investigated by the team     |
| question         | Issue starts a discussion or raises a question    |
