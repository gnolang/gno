WIP
---

#### POC for alt-hosted gno.land web API

BACKLOG
-------
- [ ] deploy a compatible interface to gno.and and add a registry

- [ ] try out gno functions (call out to realm to render template) - as MD extensions
- [ ] remove pflow icon and name - replace with generic placeholders
- 
- [ ] build a template mechanism that depends on functions deployed to gno.land
- [ ] support image hosting use case from realm /r/stackdump/www /r/stackdump/bmp:filename.jpg

- [ ] make About page configurable from gno.land code
- [ ] replace pflow and/or add another example widget
 
ICEBOX
------
- [ ] try 250*250 bmp grid - png rendering
- [ ] could we depend on gnoweb and/or gnodev in a better way?

- [ ] consider implementing a frontend-only solution to template rendering
      we could still lint the json body on the server side and then render it on the client side
      https://developer.mozilla.org/en-US/docs/Web/API/Web_components/Using_custom_elements#implementing_a_custom_element

- [ ] refine dependencies on gnodev - eventually make a first-class api to host 3rd party plugins
- [ ] obey TTL set in gno.land code - and/or just have a default for rendering new blocks that involve calls to gnoland
- [ ] try out "HyperRealm" approach - get-w/-content-in-header as render plugin - Why not do a post??
      could have a 'hosted' version of the goldmark plugin so others may depend on this node for rendering
