# Core Gno Packages

This folder contains Gno smart contracts and applications that are essential to
the operation of the `gno.land` chain and are maintained by the core team.

These include:

- Frequently-used crucial pure packages, such as `p/nt/avl`
- System packages like `/r/sys/users` & `/r/sys/names` 
- Chain-related governance packages like `/r/gov/dao`
- Official `gno.land` applications such as `/r/gnoland/blog` and `/r/gnoland/users`
- Other core-maintained packages, such as items under `/r/docs`

All packages here are planned to be included in the genesis block of the chain,
and are subject to higher review, testing, and maintenance standards.

New packages should only be added here if they are core-maintained and critical to the platform.

### Contributing

If you're thinking of adding something to this folder, ask yourself:

"Would this package be considered long-term as part of the chain’s identity or stability?
Is the core team expected to maintain and provide support for it?" 

If yes, place the package here. 
If not, it's place-to-be is somewhere else; `examples-gno/`, Gnoverse org, 
personal repository, etc.