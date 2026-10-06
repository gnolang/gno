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

- New setting `RenderConfig.IPFSGateway`, set by `-ipfs-gateway` on gnoweb and
  `-web-ipfs-gateway` on gnodev. Both flags default to
  `https://ipfs.filebase.io` (`DefaultIPFSGateway`). `NewDefaultAppConfig`
  leaves it empty, like the realm notice, so library callers opt in. Empty
  disables everything below. `NewRouter` rejects an invalid value.
  `NewHTMLRenderer`, which does not know the chain domain, ignores a malformed
  one, so no caller renders relative URLs.
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
    `\`. On the gateway the CID is a path segment rather than the host, so
    once rewritten, `ipfs://<a>/../<b>` would resolve to `<gateway>/ipfs/<b>`,
    other content.
  - So is a URL with a `uri` query parameter: gateways built on boxo (Kubo,
    Rainbow) redirect such a request to the content the parameter names.
- Rewritten images get `loading="lazy"`, so an image no IPFS node provides does
  not hold the page's `load` event, which can otherwise take close to a minute.
  It does not spread out a gallery's requests: unsized images sit inside the
  browser's lazy-load margin, so they are all requested at once anyway.
- A rewritten autolink is labelled with the gateway URL. The original text may
  name a retired gateway, and copying it would give a dead URL.
- The rewritten destination is escaped for markdown (`\` and `&`), because the
  renderers resolve escapes and character references in a destination once
  more. A test renders each URL form (`ipfs://`, `ipns://`, entity-encoded,
  path gateway, subdomain gateway, inlined DNSLink) next to the same gateway
  URL written by hand and requires identical output.
- The CSP drops the dead hosts and allows the configured gateway origin.
- The gateway must be `https` (`http` only on loopback), with no credentials,
  path, query or fragment. Its host must be a DNS name or an IPv4 address,
  with an optional port in 1..65535 (the CSP has no IPv6 host-source). It
  must not be on the chain domain (`AppConfig.Domain`, `gno.land`) or a
  subdomain of it: gno.land's own gnoweb is served there, and a path gateway
  on that site would serve any author's HTML from it. gnoweb does not know
  the host it is actually served from, so other deployments must not point
  the gateway at their own site.

The default was measured on 2026-09-23 over 16 CIDs (example-realm images,
public-gateway-checker fixtures, NFT metadata directories, xkcd, the Wikipedia
mirror), querying both gateways concurrently:

| Gateway | Served | Typical latency | `/ipns/` |
|---|---|---|---|
| `ipfs.filebase.io` | 14/16 | 0.2 to 1.9 s | yes |
| `gateway.pinata.cloud` | 13/16 | 5 to 13 s | 403 |

Filebase documents its public gateway as meant "for testing and light usage",
at 200 requests per minute, and it is stricter in practice. During review, a
second burst of 100 image requests from one IP, a few seconds after the
first, got 85 to 90 of them back as `429`: a single client hits the limit
easily. It is a fine default for development, but gno.land itself needs a
dedicated or self-hosted gateway.

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
- The rewrite and the CSP must ship together. On 2026-10-06, gno.land still
  served the old `img-src` (`ipfs.io`, `cloudflare-ipfs.com`,
  `assets.gnoteam.com`). Rewritten images stay blocked there until the
  deployed header allows the gateway, whether gnoweb `-strict` sets that
  header or the proxy in front of it does.
- Not rewritten: the documentation renderer, the site banner, and raw HTML under
  `-html`.
- `md.Link` and `md.Image` (through `sanitize`) still drop `ipfs://` URLs.
  `docs/users/explore-with-gnoweb.md` says to write the markdown directly only
  for CIDs a realm controls, to validate any user-supplied CID (ASCII
  letters and digits only) before building `ipfs://<cid>` from it, and to
  build alt text and link labels with `sanitize.InlineText`.
