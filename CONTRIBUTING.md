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
    - [Repository Structure](#repository-structure)
- [How do I?](#how-do-i)
    - [How do I submit changes?](#how-do-i-submit-changes)
    - [How do I report a bug?](#how-do-i-report-a-bug)
    - [How do I request a feature?](#how-do-i-request-a-feature)
- [Style Guides](#style-guides)
    - [Git Commit Messages](#git-commit-messages)
    - [Go Style Guide](#go-style-guide)
    - [Documentation Style Guide](#documentation-style-guide)
- [Additional Notes](#additional-notes)
    - [Issue and Pull Request Labels](#issue-and-pull-request-labels)
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

- Go (version 1.18+)
- make (for using Makefile configurations)
- Docker (for using the official Docker setup files)

It is recommended to work on a Unix environment, as most of the tooling is built around ready-made tools in Unix (WSL2
for Windows / Linux / macOS).

For Gno, there is no specific tooling that needs to be installed, that’s not already provided with the repo itself.
You can utilize the `gnodev` command to facilitate Gnolang support when writing Smart Contracts in Gno, by installing it
with `make install gnodev`.

Additionally, you can also configure your editor to recognize `.gno` files as `.go` files, to get the benefit of syntax
highlighting.

Currently, we support a [Visual Studio Code](https://marketplace.visualstudio.com/items?itemName=harry-hov.gno) extension
(eventually official in the future) for Gnolang.

#### ViM Support

Add to your `.vimrc` file:

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

#### Emacs Support

1. Install [go-mode.el](https://github.com/dominikh/go-mode.el).
2. Add to your emacs configuration file:

```lisp
(add-to-list 'auto-mode-alist '("\\.gno\\'" . go-mode))
```

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
`make test`

This will execute the full test suite, that includes tests for `.gno` files, as well as project `.go` tests.

#### Running test workflows

Using a tool like [act](https://github.com/nektos/act) can enable you to run specific repository workflow files locally,
using Docker. The workflow configurations contain different `go` versions, so it might be worth running if you’re
worried about compatibility.

To run the entire test suite through workflow files, run the following command:
`act -v -j go-test`

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

To let the core team know about _what_ your PR does, and _why_ it does it, you should take a second to fill out one of
two PR templates provided on the repo.

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

#### Base PR template

The [base PR template](https://github.com/gnolang/gno/blob/master/.github/pull_request_template.md) is by default a
simple template meant to illustrate quickly what’s the context of the PR. Here, you should describe your PR in detail,
and leave a remark as to how your changes have been tested.

If you've run a manual test, please provide the exact steps taken.

#### Detailed PR template

The [detailed PR template](https://github.com/gnolang/gno/blob/master/.github/PULL_REQUEST_TEMPLATE/detailed_pr_template.md)
is used when you want to convey a bit more detail and context for your PRs. In contrast to the default PR template, it
follows a checklist approach that is meant to be filled out by the PR creator, with as much information as possible.
These detailed PR descriptions serve an important purpose down the line in the future, when it’s important to understand
the context under which a PR has been created, and for what reason.

You can utilize the detailed PR template by appending the following query param to the URL on the new PR page:

`template=detailed_pr_template.md`

### How do I report a bug?

Found something funky while using gno? You can report it to the team using GitHub issues.

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
Commit standard, and be short and precise.

A general rule of thumb:

- Use Conventional Commits for PR titles
- Never favor rewriting history in PRs (rebases have very few exceptions, like implementation rewrites)

### Go Style Guide

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

#### Labels for Issues

| Label Name      | Description                                                     |
|-----------------|-----------------------------------------------------------------|
| breaking change | PR introduces backwards incompatible functionality              |
| bug fix         | PR introduces a fix for a known bug                             |
| dependencies    | PR introduces a package version bump                            |
| documentation   | PR introduces an improvement or addition to the docs            |
| don’t merge     | PR contains unstable functionality that shouldn't be merged yet |
| feature         | PR adds new functionality to Gno                                |
| hotfix          | PR applies a hotfix that should be merged ASAP                  |

#### Labels for Pull Requests

| Label Name       | Description                                       |
|------------------|---------------------------------------------------|
| bug              | Issue confirms a bug is present in Gnoland        |
| duplicate        | Issue already exists, or has been posted before   |
| good first issue | Issue is a great introduction for newcomers       |
| help wanted      | Issue requires extra attention from the community |
| info needed      | Issue is lacking information needed for resolving |
| investigating    | Issue is still being investigated by the team     |
| question         | Issue starts a discussion or raises a question    |

### Notable Contributions

Notable contributions of fixes/features/refactors:

* [https://github.com/gnolang/gno/pull/208](#208) - @anarcher, gnodev test with testing.T
* [https://github.com/gnolang/gno/pull/167](#167) - @loicttn, website: Add syntax highlighting + security practices
* [https://github.com/gnolang/gno/pull/136](#136) - @moul, foo20, a grc20 example smart contract
* [https://github.com/gnolang/gno/pull/126](#126) - @moul, feat: use the new Precompile in gnodev and in the addpkg/execution flow (2/2)
* [https://github.com/gnolang/gno/pull/119](#119) - @moul, add a Gno2Go precompiler (1/2)
* [https://github.com/gnolang/gno/pull/112](#112) - @moul, feat: add 'gnokey maketx --broadcast' option
* [https://github.com/gnolang/gno/pull/110](#110), [https://github.com/gnolang/gno/pull/109](#109), [https://github.com/gnolang/gno/pull/108](#108), [https://github.com/gnolang/gno/pull/106](#106), [https://github.com/gnolang/gno/pull/103](#103), [https://github.com/gnolang/gno/pull/102](#102), [https://github.com/gnolang/gno/pull/101](#101) - @moul, various chores.
