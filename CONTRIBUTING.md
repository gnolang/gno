# Contributing to Gno

Thank you for contributing to Gno! This guide will help you get started.

## Important Resources

- **[Documentation](https://docs.gno.land)** - comprehensive documentation for Gno
- **[Go Package Docs](https://gnolang.github.io/gno/github.com/gnolang/gno.html)** - API reference documentation
- **[Awesome Gno](https://github.com/gnoverse/awesome-gno)** - curated list of Gno resources
- **[Discord](https://discord.gg/YFtMjWwUN7)** - join our community for discussions and support

## Getting Started

### Prerequisites

- Go 1.23+
- Unix environment (Linux/macOS/WSL2)
- `make` command

### Setup

```bash
git clone https://github.com/gnolang/gno.git
cd gno
make install
```

If `gno` and `gnokey` commands are not found, see [Go's official
documentation](https://go.dev/doc/tutorial/compile-install) for configuring your
PATH.

### Testing

Run all tests:
```bash
make test
```

Test specific Gno code:
```bash
gno test ./examples/... -v
```

## Project Structure

See the [README](./README.md) for project structure overview. Most important
directories have their own README explaining their purpose and how to
contribute.

**Tip**: Look at recent commits to understand typical file modifications:
```bash
git log --oneline -10
```

## Making Changes

### Submitting Pull Requests

1. **Open as draft first** - Start with a draft PR to run initial checks
2. **Read bot comments** - Our bots provide helpful guidance
3. **Check CI results** - We have extensive CI to catch issues early
4. **Fill out the template** - Explain what and why
5. **Move to ready** - Once CI passes and you've self-reviewed

Our CI is designed to help both you and maintainers identify potential side
effects of changes. Use it as a guide to improve your PR.

### Git Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):
- `feat(scope): add new feature`
- `fix(scope): fix issue`
- `docs(scope): update documentation`

### Rebasing

Avoid rebasing after opening your PR for review. Maintainers handle the final
squash/merge. Add new commits to address feedback instead of force-pushing.

Using merge commits instead of rebase allows reviewers
to better review changes only since their last review.
To disable rebase when using `git pull` on the repository, run:

	git config pull.rebase false

This will be disabled only for the git repository you're currently on.

### Code Style

- Read [PHILOSOPHY.md](./PHILOSOPHY.md) to understand our approach
- Follow [Effective Go](https://go.dev/doc/effective_go)
- Use existing patterns in the codebase
- Run `make fmt` and `make lint` before committing

### Editor Setup

For editor configuration and LSP support, see
[github.com/gnoverse/gnopls](https://github.com/gnoverse/gnopls).

## Reporting Issues

Before opening an issue:
1. Search existing issues
2. Include reproduction steps, version info, and logs (see `gno bug`)

## Quality Standards

Your contribution should:
- Solve a real problem
- Include tests
- Update documentation if needed
- Follow existing patterns

Avoid superficial changes aimed at farming airdrops or karma.

## Communication

See the [Community section](./README.md#community) in our README for Discord,
GitHub discussions, and other communication channels.

---

For more documentation, see the [docs](./docs/) folder.
