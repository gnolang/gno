# ghverify

This realm is intended to enable off chain gno address to github handle verification.
The steps are as follows:
- A user calls `RequestVerification` and provides a github handle. This creates a new static oracle feed.
- An off-chain agent controlled by the owner of this realm requests current feeds using the `GnorkleEntrypoint` function and provides a message of `"request"`
- The agent receives the task information that includes the github handle and the gno address. It performs the verification step by checking whether this github user has the address in a github repository it controls.
- The agent publishes the result of the verification by calling `GnorkleEntrypoint` with a message structured like: `"ingest,<task id>,<verification status>"`. The verification status is `OK` if verification succeeded and any other value if it failed.
- The oracle feed's ingester processes the verification and the handle to address mapping is written to the avl trees that exist as ghverify realm variables.