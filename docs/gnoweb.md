---
state: improvements-needed # The final PoC is working well, but it must evolve performance-wise.
---

# Gnoweb

Gnoweb is an application that can call a special method on a [realm](./realm.md) that returns markdown as the output. 
This can be used to create any kind of web interface, like social networks, blog systems, forums, or standard status pages.

The method that it calls is `Render(string) string` where the input is a path (to discriminate different requests, and potentially handle them through a router), and the return value is normally markdown to be rendered, but it can be any `string` representation.