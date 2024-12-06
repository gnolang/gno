# GitHub Bot

## Overview

The GitHub Bot is designed to automate and streamline the process of managing pull requests. It can automate certain tasks such as requesting reviews, assigning users or applying labels, but it also ensures that certain requirements are satisfied before allowing a pull request to be merged. Interaction with the bot occurs through a comment on the pull request, providing all the information to the user and allowing them to check boxes for the manual validation of certain rules.

## How It Works

### Configuration

The bot operates by defining a set of rules that are evaluated against each pull request passed as parameter. These rules are categorized into automatic and manual checks:

- **Automatic Checks**: These are rules that the bot evaluates automatically. If a pull request meets the conditions specified in the rule, then the corresponding requirements are executed. For example, ensuring that changes to specific directories are reviewed by specific team members.
- **Manual Checks**: These require human intervention. If a pull request meets the conditions specified in the rule, then a checkbox that can be checked only by specified teams is displayed on the bot comment. For example, determining if infrastructure needs to be updated based on changes to specific files.

The bot configuration is defined in Go and is located in the file [config.go](./internal/config/config.go).

### GitHub Token

For the bot to make requests to the GitHub API, it needs a Personal Access Token. The fine-grained permissions to assign to the token for the bot to function are:

#### Repository permissions

- `pull_requests` scope to read is the bare minimum to run the bot in dry-run mode
- `pull_requests` scope to write to be able to update bot comment, assign user, apply label and request review
- `contents` scope to read to be able to check if the head branch is up to date with another one
- `commit_statuses` scope to write to be able to update pull request bot status check

#### Organization permissions

- `members` scope to read to be able to list the members of a team

#### Bot account role

For the bot to create a commit status on a repo - and only for this feature at the time of writing this - the GitHub account of the bot must either:

- have the `write` role on the repo
- have the `owner` role in the organization that owns the repo

## Usage

```bash
> github-bot check --help
USAGE
  github-bot check [flags]

This tool checks if the requirements for a pull request to be merged are satisfied (defined in ./internal/config/config.go) and displays PR status checks accordingly.
A valid GitHub Token must be provided by setting the GITHUB_TOKEN environment variable.

FLAGS
  -dry-run=false   print if pull request requirements are satisfied without updating anything on GitHub
  -owner ...       owner of the repo to process, if empty, will be retrieved from GitHub Actions context
  -pr-all=false    process all opened pull requests
  -pr-numbers ...  pull request(s) to process, must be a comma separated list of PR numbers, e.g '42,1337,7890'. If empty, will be retrieved from GitHub Actions context
  -repo ...        repo to process, if empty, will be retrieved from GitHub Actions context
  -timeout 0s      timeout after which the bot execution is interrupted
  -verbose=false   set logging level to debug
```
