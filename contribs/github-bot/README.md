# GitHub Bot

## Overview

The GitHub Bot is designed to automate and streamline the process of managing pull requests. It can automate certain tasks such as requesting reviews, assigning users or applying labels, but it also ensures that certain requirements are satisfied before allowing a pull request to be merged. Interaction with the bot occurs through a comment on the pull request, providing all the information to the user and allowing them to check boxes for the manual validation of certain rules.

## How It Works

### Configuration

The bot operates by defining a set of rules that are evaluated against each pull request passed as parameter. These rules are categorized into automatic and manual checks:

- **Automatic Checks**: These are rules that the bot evaluates automatically. If a pull request meets the conditions specified in the rule, then the corresponding requirements are exectued. For example, ensuring that changes to specific directories are reviewed by specific team members.
- **Manual Checks**: These require human intervention. If a pull request meets the conditions specified in the rule, then a checkbox that can be checked only by specified teams is displayed on the bot comment. For example, determining if infrastructure needs to be updated based on changes in specific files.

The bot configuration is defined in Go and is located in the file [config.go](./config.go).

### Conditions

// TODO

### Requirements

// TODO

### GitHub Token

// TODO

## Usage

```bash
> go install github.com/gnolang/gno/contribs/github-bot@latest
// (go: downloading ...)

> github-bot --help
This tool checks if the requirements for a PR to be merged are satisfied (defined in config.go) and displays PR status checks accordingly.
A valid GitHub Token must be provided by setting the GITHUB_TOKEN environment variable.

  -dry-run
    	print if pull request requirements are satisfied without updating anything on GitHub
  -owner string
    	owner of the repo to process, if empty, will be retrieved from GitHub Actions context
  -pr-all
    	process all opened pull requests
  -pr-numbers value
    	pull request(s) to process, must be a comma separated list of PR numbers, e.g '42,1337,7890'. If empty, will be retrieved from GitHub Actions context
  -repo string
    	repo to process, if empty, will be retrieved from GitHub Actions context
  -timeout uint
    	timeout in milliseconds
  -verbose
    	set logging level to debug
```
