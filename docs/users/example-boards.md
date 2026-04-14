# Using the Boards Application on Gno.land

The Boards realm is one of the first applications on Gno.land, offering a
decentralized discussion forum where anyone can create and participate in
conversations. This guide will walk you through discovering, exploring, and
interacting with the Boards realm.

## Finding Boards

The main Boards application can be found at
[gno.land/r/gnoland/boards2/v1](https://gno.land/r/gnoland/boards2/v1). When you visit this
URL, you'll see the rendered output of the Boards realm's `Render()` function,
which displays the current state of the forum.

## Exploring Boards

The Boards realm organizes content into:

1. **Boards** - General topic categories (e.g., "Gno", "Random", "Meta")
2. **Threads** - Individual discussion topics within boards
3. **Posts** - Individual messages within threads

You can navigate between these levels by clicking on the links in the rendered
output. Each level presents different information and options for interaction.

For more details on browsing through realms and their content, see
[Exploring with gnoweb](./explore-with-gnoweb.md).

## Interacting with Boards

To interact with the Boards application (creating boards, threads, or posts),
you'll need:

1. A Gno.land account with some GNOT tokens
2. A way to sign and send transactions to the Gno.land network
3. For the time being - an invite to use a board within the Boards2 app. You can
request an invite from the creator on the board page. 

You can interact with Boards through the command line using `gnokey`. For
detailed instructions on sending transactions to realms, see
[Interacting with gnokey](./interact-with-gnokey.md).

## Viewing Your Contributions

After interacting with the Boards realm, you can view your contributions by
navigating to the board where you posted and finding your thread or post in the
list.

## Building Your Own Board

Inspired by the Boards realm? You can create your own version by:

1. Studying the [source code](https://gno.land/r/gnoland/boards2/v1$source&file=public.gno)
2. Deploying a modified version to your own namespace
3. Adding your own features and improvements

The Boards application showcases many of Gno's powerful features including state
persistence, rendered UI, and interactive functionality - making it an excellent
example to learn from.
