# PR xxxx: Security patterns guide and example

## Context

Audit synthesis found two recurring builder-facing mistakes that are easy to
teach with small, sanitized examples:

- deriving caller identity from `cur.Previous()` without first checking
  `cur.IsCurrent()`;
- returning URL path or user-authored text from `Render` without escaping
  markdown-sensitive characters.

The repository already has long-form security documentation and many example
realms, but there was no compact docs example that put these two patterns next
to executable tests.

## Decision

Add a small `gno.land/r/docs/security_patterns` realm that demonstrates:

- an authenticated mutator that derives the caller only after `cur.IsCurrent()`;
- a `Render(path)` implementation that escapes untrusted markdown text;
- tests for both the admin check and Render escaping.

Update `docs/resources/gno-security-guide.md` with a concise Render anti-pattern
section and checklist item. Link the example from the docs examples index.

## Alternatives Considered

- Updating an older safe-object example in place. This would mix a security
  modernization with existing tutorial behavior and increase review scope.
- Adding only documentation. That would explain the rule but leave no executable
  contract for future agents and maintainers to test.
- Adding a broader security demo app. That would be more complete but less
  reviewable for this first public slice.

## Consequences

Builders get a small, runnable reference for two common audit findings. Future
security-guide slices can add similarly narrow examples without depending on
private audit material.
