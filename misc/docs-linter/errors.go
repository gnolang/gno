package main

import "errors"

var (
	errEmptyPath             = errors.New("you need to pass in a path to scan")
	err404Link               = errors.New("link returned a 404")
	errFound404Links         = errors.New("found links resulting in a 404 response status")
	errFoundUnescapedJSXTags = errors.New("found unescaped JSX tags")
	errFoundLintItems        = errors.New("found items that need linting")
)
