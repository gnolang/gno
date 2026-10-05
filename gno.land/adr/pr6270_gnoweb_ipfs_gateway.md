# ADR: gnoweb resolves IPFS URLs through a configurable gateway

## Context

gnoweb's CSP allowed images from two IPFS gateways, `https://ipfs.io` and
`https://cloudflare-ipfs.com`. Both are gone:

- `cloudflare-ipfs.com` (and `cf-ipfs.com`) no longer resolve.
- `ipfs.io` and `dweb.link` stopped serving hotlinked content on 2026-09-21
  (IPFS blog, "IPFS is moving beyond the sponsored gateways", 2026-08-25).
  Plain requests get `429` with a `Sunset` header; browser subresource
  requests get a Cloudflare challenge. `w3s.link` and `nftstorage.link`
  redirect to them.

Every IPFS image on gno.land broke. Realm code can be redeployed, but realm
state holds such URLs too (user-supplied profile images, NFT metadata), and
nothing short of a render-time fix repairs those.

`ipfs://` URLs did not work either: links pointed at a scheme most browsers
cannot open, and images were blocked by the CSP.

## Decision

- New setting `AppConfig.IPFSGateway` (`-ipfs-gateway` on gnoweb,
  `-web-ipfs-gateway` on gnodev), default `https://ipfs.filebase.io`. Empty
  disables everything below.
- A markdown AST transformer (`markdown/ext_ipfs.go`, priority 400, so it runs
  before link classification and image validation at 500) rewrites links,
  autolinks, reference links and images, including inside `<gno-foreign>`:
  - `ipfs://<cid>/<path>` and `ipns://<name>/<path>` become
    `<gateway>/ipfs/...` and `<gateway>/ipns/...`;
  - URLs on the retired gateways above (plus `gateway.ipfs.io`), in path form,
    protocol-relative form (`//host/...`) or subdomain form
    (`<cid>.ipfs.<host>`, `<label>.ipns.<host>` with DNSLink labels decoded),
    become the same path on the gateway.
  - Only an alphanumeric CID, or an IPNS key or DNS name, is rewritten. Other
    hosts, userinfo, ports and anything unparsable are left as written.
  - So is a path with a `.` or `..` segment, also percent-encoded or split by
    `\`. On the gateway the CID is a path segment rather than the host, so the
    browser would resolve `ipfs://<a>/../<b>` to `<b>`, other content.
- The rewritten destination is escaped for markdown (`\` and `&`), because the
  renderers resolve escapes and character references in a destination once
  more. A test renders each URL form (`ipfs://`, `ipns://`, entity-encoded,
  path gateway, subdomain gateway, inlined DNSLink) next to the same gateway
  URL written by hand and requires identical output.
- The CSP drops the dead hosts and allows the configured gateway origin.
- The gateway must be `https` (`http` only on loopback), with no credentials,
  path, query or fragment. Its host must be a DNS name or an IPv4 address,
  with an optional port in 1..65535 (the CSP has no IPv6 host-source), and
  not on the gnoweb domain or a subdomain of it, where a path gateway would
  serve any author's HTML from the same site.

The default was measured on 2026-09-23 over 16 CIDs (example-realm images,
public-gateway-checker fixtures, NFT metadata directories, xkcd, the Wikipedia
mirror), querying both gateways concurrently:

| Gateway | Served | Typical latency | `/ipns/` |
|---|---|---|---|
| `ipfs.filebase.io` | 14/16 | 0.2 to 1.9 s | yes |
| `gateway.pinata.cloud` | 13/16 | 5 to 13 s | 403 |

Filebase documents its public gateway as meant "for testing and light usage",
at 200 requests per minute. Images are fetched by each visitor's browser, so
that limit applies per visitor, not per gnoweb instance.

## Alternatives considered

- **Swap the two CSP hosts for working gateways.** Leaves every URL that
  already points at ipfs.io broken, and repeats the hardcoding that just broke.
- **Self-hosted gateway (Rainbow, Kubo) or a paid dedicated gateway.** The most
  reliable option and what the IPFS docs now recommend, but it needs someone to
  run or pay for it. With this setting, moving to one is a config change.
- **In-browser verified retrieval (`@helia/verified-fetch`, the drop-in service
  worker).** Verifies content client-side, but ships a large JS bundle and needs
  a broad `connect-src`, loosening gnoweb's CSP.
- **Multi-gateway fallback in frontend JS.** Gateways miss different content, so
  retrying the next one on image error would raise the hit rate. Left as a
  follow-up.
- **Allow `ipfs://` in `p/nt/markdown/sanitize/v0`.** It is deployed on mainnet,
  and deployed v0 packages are not changed in place. Needs a new version.

## Consequences

- IPFS images render again, and `ipfs://` becomes the recommended way to
  reference IPFS content: the gateway can change without touching realms.
- Visitors' browsers contact the configured gateway, as they contacted ipfs.io
  before. A free public gateway has no SLA; operators can point
  `-ipfs-gateway` elsewhere.
- A gateway only serves content some IPFS node provides. One sampled CID had no
  provider on any gateway.
- Not rewritten: the documentation renderer, the site banner, and raw HTML under
  `-html`. Autolinks keep their original label while the href points at the
  gateway.
- `md.Link` and `md.Image` (through `sanitize`) still drop `ipfs://` URLs.
  `docs/users/explore-with-gnoweb.md` says to write the markdown directly only
  for CIDs a realm controls, to validate any user-supplied CID (ASCII
  letters and digits only) before building `ipfs://<cid>` from it, and to
  build alt text and link labels with `sanitize.InlineText`.
