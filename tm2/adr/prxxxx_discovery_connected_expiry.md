# ADR: A Connection Keeps An Address In The Discovery Book For A Day

## Context

The discovery reactor keeps up to 1000 peer addresses in its address book,
saves them to disk, and queues every one for dialing when the node starts. The
book used to evict whatever a peer had shared longest ago, so a peer sharing
1,020 fresh addresses replaced the whole book.

The fix keeps an address the node dialed and connected to ahead of every
address it was only told about. A mark that lasts for good has two costs. A
peer earns it for addresses it controls by completing a handshake on each and
hanging up, and those addresses then never leave. A book whose every address
is marked admits no new one, so marked addresses that later die, after a key
rotation or a relaunched test network, fill it until an operator deletes the
file.

## Decision

A connection keeps an address ahead of shared ones for `connectedExpiry`, 24
hours, after the node last held it. Past that, the address is evicted like one
the node never connected to, oldest shared first.

The expiry is set against how often a live peer's mark is renewed. Every
discovery round, 3 seconds by default, the reactor renews the mark of each
outbound peer the node holds, so a live peer it stays connected to never
lapses. A peer it no longer holds is dialed again when someone shares it while
one of the 10 outbound slots is free, when the node restarts, or when the
operator lists its address as a persistent peer or a seed. The expiry is
therefore how long an address stays ahead once the node has dropped it or
stopped. A day keeps every mark across a stop shorter than a day, since the
book is written on stop, and is how long an address an attacker had the node
dial once, or a dead one, keeps its place.

A renewal asks for the mark to be written to disk only once half the expiry
has passed since it last asked, rather than every round, which would rewrite
the file every 3 seconds. After a crash, a saved mark is at most about 12
hours older than the last connection.

## Alternatives considered

- *Mark from a consensus signal*: CometBFT promotes an address once its peer
  has relayed enough votes or block parts. tm2 has that hook commented out in
  its consensus reactor, so it is a change to a second package; the expiry
  bounds the handshake mark without it.
- *Reserve part of the book for addresses not yet connected to*, as CometBFT's
  new and old buckets do: the book keeps learning, but a mark earned by a
  handshake still never ends.
- *An hour*: an attacker's or a dead address leaves sooner, but a node stopped
  for more than an hour restarts with no address ahead of a flood.
- *A week*: a longer stop keeps its marks, and an attacker's or a dead address
  keeps its place seven times longer.

## Consequences

- A peer holding an outbound connection keeps its mark while it does. That is
  at most one per outbound slot, plus the persistent peers, which the slot
  count leaves out.
- An attacker renews a lapsed mark only by being dialed again, which needs a
  free outbound slot unless the operator listed its address.
- For a day after the last connection, a book full of marked dead addresses
  still refuses new ones.
