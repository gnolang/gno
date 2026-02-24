# Users and Teams in gno.land

## Namespaces

All addresses can deploy under their own address-based namespace:

```
gno.land/p/g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5/...
gno.land/r/g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5/...
```

The namespace is automatically derived from the deployer's address --
no registration is needed. These address-based namespaces allow for
permissionless deployment.

Username-based namespaces (e.g. `gno.land/r/myusername/myrealm`) are not
currently supported and will be revisited via GovDAO governance.

## Teams and Collaborative Development

There is an ongoing effort to bring team-based development through shared
namespaces. This feature will enable:

1. Multiple addresses with permission to deploy under a team namespace
2. Role-based access control for team members
3. Collaborative development of larger projects

Until full team support is available, collaborative development can be achieved through:

1. **Account sharing** - Multiple developers using the same key (not recommended for security reasons)
2. **Multi-signature wallets** - Using multi-sig wallets to control deployment to a shared namespace
3. **Development on branches** - Developing under individual namespaces and then migrating to a main namespace

## Related Resources

For more information on users and namespaces, refer to:

- [Gno Packages](./gno-packages.md) - Understand how namespaces work within the package system
- [Realms](./realms.md) - Learn about stateful applications that can be deployed under your namespace
- [Deploying Packages](../builders/deploy-packages.md) - Instructions for deploying code under your namespace
