# Seahorse

Seahorse is the distillation of Notional's knowledge about validation, relaying, networking, and systems into a linux distro that includes a blockchain.  It is specifically being launched because of the many, many potential validators who contacted us after we began to work on improving Osmosis epoch times. 


## Blockchain Software stack

After a great deal of thought, the seashorse blockchain software stack is looking like:

* Cosmos SDK v0.45.0
  * Because it's good
* CosmWasm
  * To give juno a friend to play with and allow feature development without chain upgrades
* Tendermint v0.34.15
  * the latest and greatest


This is a quite vanilla, but highly extensible stack.  Should run just about anywhere after arm issues are fixed for cosmwasm.




## You, the validators, make the plan

Typically a blockchain team awards themselves a good deal of stake when making a chain, and typically, I reckon this is the right pattern.  That said, a while back I became enamored with the idea that there needs to be a test of what happens when the valset begins with completely equal VotePower and the chain evolves over time using governance.  

**In short, this means**
* Validators run the chain
* Validators determine how the chain is placed in the market
* Validators brand the chain
* Validators build the chain

I am not going to dictate the strategy with Seahorse.  Its validators will.  Beyond getting the software into an inital, known-good state, I also don't intend to write a lot of software for it.  My expectastion is that governance, initially run entirely by the valset, with equal VotePower will choose the eventual direction of seahorse. Because design and branding are important, I do have someone in mind for that, and he/she/it will be validating, but they'd need to be hired by governance after launch.  


## Hardware

At Notional, we've had the opportunity to learn in depth what kinds of gear validate effectively.  Our gaia validator has even run on a Raspberry Pi (thanks so much to our delegators who encourage this sort of testing) and for about eight months we've done extensive real world tests of "what works".

What works?

Well, any sufficiently fast processor with an nvme 1.4 pcie4x4 hard drive seems to really do a fantastic job.  It's my personal hope to bootstrap this long-overdue hardware from Seahorse, as well.  But bits ship far faster than atoms, so the hardware will happen sometime down the line. 




# Why strap Linux to it?

Peer to peer applications tend to involve a complex stack and are therefore difficult to begin developing. SOS provides you with a complete development environment, As well as patterns and ideas to follow.

This image represents an opinionated approach to the construction of distributed and p2p applications.

### Variants

* Full
  * The full image is a development environment with P2P development tooling ready to go:
    * Starport
    * HNS or HNSD
    * IPFS
    * Gox
    * Docker
    * Go
    * Node.js
    * Zerotier
* Lite
  * Docker
  * Docker-compose

### Supported Devices & Platforms
SOS targets ARM64 and AMD64 processors. 

* Raspberry Pi 3
* Raspberry Pi 4
* Odroid C2
* Minimus

[Arch Linux](archlinux.org) was a very deliberate choice: In contrast to other distributions, arch packages are always up-to-date. Additionally, the arch user repository offers a wide variety of easy to install packages contributed by the community.


### Vital Information:

- designed to be consumed by your favorite CI system and used in the production of ready made system images.

  - defaults to GitHub Actions

- No binaries are used in the build process. All source code is copied to /spos so that users can easily rebuild the operating system. The Raspberry Pi 4 64 bit kernel is currently built elsewhere to save time, but we use a fully-open implementation. If you have spare time, you can build it from [source](https://aur.archlinux.org/packages/linux-raspberrypi4-aarch64/). It is blob-free.

- FAST
  - Full builds take ~30 minutes.
  - SPOS can produce a fully-cached image on a hetzner A61x in about 2 minutes.
  - Docker pull cann be used to load spos into your docker cache.

- one OS for every platform:
  - Cloud
    - AMI (AMD64 & ARM64)
  - Mobile (PinePhone, PineTab)
  - Router
    - Dawn
  - Laptop
    - Samsung
      - Chromebook Plus
    - Acer
      - Chromebook Flip
      - Chromebook R13
  - SBC
    - ~~Raspberry Pi 3 & 4~~
    - Odroid 
      - ~~C2~~
      - N2
    - Dragonboard 410C
    - Pine64
    - Rock64
