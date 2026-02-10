# Users and Teams in gno.land

## User Registration

In Gno.land, users can register a unique username that:
- Provides a more readable identity than a blockchain address
- Grants the exclusive right to deploy code under a matching namespace
- Can be used in social contexts across the ecosystem

### Registration Process

To register a username:

1. Visit the user registry realm at [`gno.land/r/gnoland/users/v1`](https://gno.land/r/gnoland/users/v1)
2. Check if your desired username is available
3. Register using transaction links & a web extension wallet, or with `gnokey` using the following command:

```bash
gnokey maketx call \
  --pkgpath "gno.land/r/gnoland/users/v1" \
  --func "Register" \
  --args "YOUR_USERNAME" \
  --gas-fee 1000000ugnot \
  --gas-wanted 2000000 \
  --send "1000000ugnot" \
  --remote https://rpc.gno.land:443 \
  --chainid staging \
  YOUR_KEY_NAME
```

The registration costs 1 GNOT, which serves as an anti-spam measure and ensures 
users value their identities.

## Username Ownership

Once registered, a username is permanently linked to the registering address. This address:

- Has exclusive rights to deploy packages under that namespace
- Can manage that username's profile
- and more to come.

### Namespace Access

Your registered username becomes an exclusive namespace where you can deploy both realms and pure packages:

```
gno.land/p/YOUR_USERNAME/...  # Pure packages
gno.land/r/YOUR_USERNAME/...  # Realms
```

For example, a user with the username "alice" could deploy:
- `gno.land/r/alice/blog` - A personal blog realm
- `gno.land/p/alice/utils` - A utility package

Only the address that registered the username can deploy to these paths, ensuring security and preventing impersonation.

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

## Public Anonymous Namespaces

In addition to registered usernames, all addresses can deploy under their own address-based namespace:

```
gno.land/p/g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5/...
gno.land/r/g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5/...
```

This allows for permissionless deployment even without a registered username, though these long namespaces are less user-friendly.

## Related Resources

For more information on users and namespaces, refer to:

- [Gno Packages](./gno-packages.md) - Understand how namespaces work within the package system
- [Realms](./realms.md) - Learn about stateful applications that can be deployed under your namespace
- [Deploying Packages](../builders/deploy-packages.md) - Instructions for deploying code under your namespace

To explore registered users, visit the [User Registry](https://gno.land/r/gnoland/users/v1) on the Staging network.
