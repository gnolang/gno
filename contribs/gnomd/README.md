# `gnomd`: Terminal Markdown Viewer for Gno Documentation

**`gnomd`** is a lightweight command-line tool that renders Markdown files as styled ANSI output in your terminal. It is intended for quickly viewing Gno-related documentation—such as proposals, READMEs, or changelogs—without needing a web browser.

## Features

- Renders GitHub-flavored Markdown to your terminal
- Accepts file paths or reads from standard input
- Preserves heading structure, emphasis, and code blocks

## Usage

### Render one or more Markdown files:

```
gnomd README.md another-doc.md
```

### Pipe Markdown from a remote source or command:

```
curl https://gno.land/r/demo/propdao$source\&file=README.md | gnomd
```

## Installation

```
go install github.com/gnolang/gno/contribs/gnomd@latest
```
