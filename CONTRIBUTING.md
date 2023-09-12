# Contributing to GNO

Thank you for looking to contribute to the GNO project.
We appreciate every open-source contribution, as it helps us improve and enhance gno for the benefit of the community.

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
      - [Testing GNO code](#testing-gno-code)
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
- **[Awesome GNO](https://github.com/gnolang/awesome-gno)** - check out the list of compiled resources for helping you
  understand the gno ecosystem
- **[Active Staging](https://gno.land/)** - use the currently available staging environment to play around with a
  production network. If you want to interact with a local instance, refer to the [Local Setup](#local-setup) guide.
- **[Twitter](https://twitter.com/_gnoland)** - follow us on Twitter to get the latest scoop
- **[Telegram](https://t.me/gnoland)** - join our official Telegram group to start a conversation about gno

## Getting Started

### Environment

The gno repository is primarily based on Golang (Go), and Gnolang (Gno).

The primary tech stack for working on the repository:

- Go (version 1.20+)
- make (for using Makefile configurations)
- Docker (for using the official Docker setup files)

It is recommended to work on a Unix environment, as most of the tooling is built around ready-made tools in Unix (WSL2
for Windows / Linux / macOS).

For Gno, there is no specific tooling that needs to be installed, that’s not already provided with the repo itself.
You can utilize the `gno` command to facilitate Gnolang support when writing Smart Contracts in Gno, by installing it
with `make install_gno`.

Additionally, you can also configure your editor to recognize `.gno` files as `.go` files, to get the benefit of syntax
highlighting.

Currently, we support a [Visual Studio Code](https://marketplace.visualstudio.com/items?itemName=harry-hov.gno) extension
(eventually official in the future) for Gnolang.

#### ViM Support

Add to your `.vimrc` file:

```vim
function! GnoFmt()
	cexpr system('gofmt -e -w ' . expand('%')) " or replace with gofumpt
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

There is an experimental and unofficial [Gno Language Server](https://github.com/jdkato/gnols)
developed by the community, with an installation guide for Neovim.

#### Emacs Support

1. Install [go-mode.el](https://github.com/dominikh/go-mode.el).
2. Add to your emacs configuration file:

```lisp
(add-to-list 'auto-mode-alist '("\\.gno\\'" . go-mode))
```

#### Sublime Text

There is an experimental and unofficial [Gno Language Server](https://github.com/jdkato/gnols)
developed by the community, with an installation guide for Sublime Text.

### Local Setup

To get started with Gno development, the process is relatively straightforward.

Clone the repo:
`git clone https://github.com/gnolang/gno.git`

Build / install base commands:
`make build `

That’s it!

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

#### Testing GNO code

If you wish to test a `.gno` Realm or Package, you can utilize the `gno` tool.

1. To install it, simply run:

    make install_gno

2. Now, you can point to the directory containing the `*_test.gno` files:

    gno test <path-to-dir> --verbose


To learn more about how `gno` can help you when developing gno code, you can look into the available
subcommands by running:

    gno --help

### Repository Structure

The repository structure can seem tricky at first, but it’s simple if you consider the philosophy that the gno project
employs (check out [PHILOSOPHY.md](https://github.com/gnolang/gno/blob/master/PHILOSOPHY.md)).

The gno project currently favors a mono-repo structure, as it’s easier to manage contributions and keep everyone
aligned. In the future, this may change, but in the meantime the majority of gno resources and source code will be
centralized here.

- `cmd` - contains the base command implementations for tools like `gnokey`, `gnotxport`, etc. The actual underlying
  logic is located within the `pkgs` subdirectories.
- `examples` - contains the example `.gno` realms and packages. This is the central point for adding user-defined realms
  and packages.
- `gnoland` - contains the base source code for bootstrapping the Gnoland node
- `pkgs` - contains the dev-audited packages used throughout the gno codebase
- `stdlibs` - contains the standard library packages used (imported) in `.gno` Smart Contracts. These packages are
  themselves `.gno` files.
- `tests` - contains the standard language tests for Gnolang

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
