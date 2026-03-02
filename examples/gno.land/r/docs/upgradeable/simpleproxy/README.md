# SimpleProxy Pattern

This pattern allows a single proxy realm to redirect calls to multiple implementations.
The proxy is the entry pattern to all future implementations. 

Limitations:
- Limited upgradeability defined by an interface at the beginning
- No storage migration mechanism by default - data is stored in each implementation
- 