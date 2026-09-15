> **v0 - Unaudited**
> This is an initial version of this realm that has not yet been formally audited.
> A fully audited version will be published as a subsequent release.
> Use in production at your own risk.

# commondao (realm)

Reference realm implementation of the `gno.land/p/nt/commondao/v0` package for
managing Decentralized Autonomous Organizations per the Common DAO Spec
(`docs/CONSTITUTION.md`, Appendix).

What it hosts:

- **DAOs and sub-DAO trees** with a Charter (purpose + description), a
  council, and per-DAO treasury addresses derived as realm sub-identities
  (`cur.Sub("dao/<id>")`).
- **Proposals** through a per-DAO registry of proposal kinds. Ten default
  kinds are seeded at creation (text, council updates, sub-DAO creation,
  dissolution, treasury spend/clawback/freeze, manage-kinds, amend-bylaws);
  the arbitrary-execution kind is opt-in by governance. The kind set itself
  is governable (`manage-kinds`), and manage-kinds can never be deregistered.
- **Treasuries**: spends from a DAO's own sub-identity; ancestor clawback and
  freeze; dissolution sweeps. Freeze blocks every self-initiated movement,
  including arbitrary execution.
- **Bylaws & Mandates** as named plaintext documents amended by verifiable
  diff patches (`gno.land/p/nt/bylaws/v0`); the `mandates/` folder is
  reserved for the (deferred) ancestor amendment power.
- **Render** pages for DAOs, settings, bylaws, proposals and votes.

Design records live in `gno.land/adr/pr6012_commondao_*.md`; the extension
guide for building your own realm on the `/p/` package is in
`gno.land/p/nt/commondao/v0`'s README.
