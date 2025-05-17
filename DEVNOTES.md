#### How to define pkgpath when we have no gno.mod

Manfred wants package loading to work without gno.mod but in this case we can't determine package path
a lot of tools rely on pkgpaths being correctly present, for example the genesis tool.

I could derive a pkgpath from the current directory and consider the package as a realm but it's a lot of asumptions and
it could not work when we expect a particular namespace

For example `./somepkg` could become `gno.land/r/dev/somepkg` automatically

I think requiring a gno.mod is not that much friction compared to the expliciteness it brings
Golang devs are already acustomed to do that as a first step

### Use in txtar?

no, it would require too much change and the loader is meant for tools that accept patterns, the txtars do not