# Backup/restore feature

## Initial assignment

This effort covers the implementation of a proper backup / restore functionality for the Gno node.

Currently, there is no block backup / restore functionality, but only transaction exports, and transaction replays.

This can be cumbersome, considering users want to be able to quickly restore Gno chains from their trusted backup without having to sync with other peers.

This functionality should work with direct node access (not over JSON-RPC), for example through special commands that utilize gRPC, as there is no need for users to back up remote node states.

Successful outcome of this effort:

    The Gno node implements an efficient backup / restore functionality that a node operator can use to save chain state

## Logs

### Init

The storage layer of a gnoland node is initialized in the gnoland command so this seems like the correct place for these features so I'm adding `gnoland backup create` and `gnoland backup restore` commands.

The only stable database backend is goleveldb. After searching for `backup` on the goleveldb repo I didn't see any official way to backup the database. There is [an issue asking for it opened in 2016](https://github.com/syndtr/goleveldb/issues/135) but no response.

We could backup every entry in the database to be agnostic of the backend but it would probably be much less efficient than working with the filesystem directly.

The output format will be a single file compressed archive. It would be better to not depend on system tools and use pure a pure go library for that.

The only lib I found that seem serious and to offer this functionality is https://github.com/mholt/archives

After examination of the data needed to be saved, it seems only the "db" and "wal" dirs are needed

Tar + Xz was chosen for the archive format but this could become options