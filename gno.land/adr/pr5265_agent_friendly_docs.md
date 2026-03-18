# ADR: Agent-Friendly Documentation

## Context

AI coding agents (Claude, Copilot, Cursor, etc.) are increasingly used to
contribute to this codebase. These agents need structured context to work
effectively — they don't have the institutional knowledge that human
contributors build up over time.

Key problems agents face in this repo:

1. **Architecture discovery.** The repo has three pillars (gnovm, gno.land,
   tm2) with distinct roles. Without a guide, agents waste context window
   exploring the wrong directories.
2. **Convention gaps.** Agents don't know about ADRs, conventional commits, or
   the testing patterns (filetests, `.txtar` integration tests) unless told.
3. **No accountability trail.** When agents make architectural decisions,
   there's no record of *why* — making review and future maintenance harder.

## Decision

Introduce a layered documentation system for AI agents:

### AGENTS.md (single source of truth)

One comprehensive file at repo root covering:
- Project overview (Gno = language, gno.land = blockchain)
- Architecture map with directory descriptions
- Build/test commands
- Critical Gnolang vs Go differences
- Conventions split into **Gno-specific**, **Go-specific**, and **Universal**
- ADR requirements and placement rules
- Common tasks, navigation tips, glossary
- Agent-specific convention sections (Claude, Copilot) for tool-specific
  guidance

### CONTRIBUTING.md additions

New sections for human operators of AI agents:
- A human is always responsible for AI-assisted work
- ADR requirement for architectural changes
- Disclose AI usage / human owner in PR descriptions

### ADR requirement

Every non-trivial architectural PR must include an ADR in the appropriate
folder:
- `gnovm/adr/` for VM/interpreter/type-checker changes
- `gno.land/adr/` for node/SDK/keeper/RPC changes
- `tm2/adr/` for consensus/p2p/mempool/crypto changes

This is especially important for AI-assisted work, where the reasoning behind
decisions might otherwise be lost.

## Alternatives Considered

- **Multiple files (CLAUDE.md, copilot-instructions.md, AGENTS.md):** Tried
  initially, but created duplication and maintenance burden. Consolidated to
  single AGENTS.md with agent-specific subsections.
- **No ADR requirement:** Considered letting agents just write code, but
  without ADRs there's no way for reviewers to understand AI reasoning or for
  future agents to learn from past decisions.
- **Embedding all context in CLAUDE.md:** Claude Code auto-loads CLAUDE.md,
  but this locks other agents out. AGENTS.md is tool-agnostic.

## Consequences

- **Positive:** Agents produce higher-quality PRs with less human correction.
  Reviewers can check ADRs to understand AI reasoning. Future agents benefit
  from accumulated ADRs.
- **Positive:** Clear Go/Gno/Universal convention split prevents the most
  common agent mistake (treating .gno as Go).
- **Trade-off:** AGENTS.md must be maintained as the codebase evolves. An
  "Improving This Document" section encourages both humans and agents to keep
  it current.
- **Trade-off:** ADR requirement adds friction to PRs, but the review and
  historical value justifies it for non-trivial changes.
