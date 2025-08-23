# Gno Optimistic Oracle (OO)

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

An Optimistic Oracle (OO) built on Gno.land. This system is designed to bring external data onto the blockchain by leveraging game-theoretic incentives. It assumes data is correct unless disputed, hence the term "optimistic."  
This implementation is inspired by the [UMA Optimistic Oracle](https://uma.xyz/) but adapted for the Gno ecosystem.

## Table of Contents
- [Core Concepts](#core-concepts)
- [How It Works: The Lifecycle of a Data Request](#how-it-works-the-lifecycle-of-a-data-request)
- [Architecture](#architecture)
- [User Roles](#user-roles)
- [Usage Example](#usage-example)
- [Developer](#developer)

## Core Concepts

The Gno Optimistic Oracle operates on the principle that data proposed to the oracle is assumed to be true. A bond is required for any new proposition. This proposition enters a "liveness" period where anyone can dispute it by posting an equal bond.

- **Happy Path**: If no one disputes the data within the liveness period, it is considered resolved and accepted as truth. The proposer's bond is returned along with a reward.
- **Unhappy Path (Dispute)**: If the data is disputed, the Gno community is called upon to vote on the correct outcome. This is handled by the `court.gno` contract. Token holders vote, and the outcome is decided by the total token weight backing each value. The winner's bond is returned, and they receive a portion of the loser's slashed bond.

## How It Works: The Lifecycle of a Data Request

The entire process, from requesting data to its final resolution, follows a clear, multi-step path.

### 1. Data Request (`RequestData`)
A user or a contract initiates a request for data by calling `RequestData`.
- **Ancillary Data**: A clear, human-readable question (e.g., "What was the price of ETH/USD at block X?").
- **Type**: The request can be a `Yes/No` question (represented by 0 and 1) or a `Numeric` value.
- **Reward**: The requester must lock a `RequesterReward` in GNOT to incentivize a proposer to provide the data.
- **Deadline**: The requester sets a deadline by which the data must be proposed, otherwise they can retrieve their locked reward.

### 2. Value Proposal (`ProposeValue`)
A **Proposer** provides an answer to the request.
- They call `ProposeValue` with the proposed answer.
- They must post a `Bond` in GNOT, which is held in escrow.
- This action starts the **Resolution Time**, a liveness window during which the proposal can be disputed.

### 3. The Liveness Period
Once a value is proposed, a countdown begins. During this period, anyone can challenge the proposed value.

- **If Undisputed**: If the `ResolutionTime` expires without any disputes, the request is considered final. Anyone can call `ResolveRequest`. The `ProposedValue` becomes the `WinningValue`. The Proposer gets their bond back, plus the `RequesterReward`.
- **If Disputed**: If another user believes the proposed value is incorrect, they can challenge it.

### 4. Dispute (`DisputeData`)
A **Disputer** can challenge the Proposer's value.
- They must call `DisputeData` before the `ResolutionTime` ends.
- They must also post a `Bond` equal to the Proposer's bond.
- This action pauses the request's resolution and initiates a formal dispute, handled by the `court.gno` contract.

### 5. Voting (`VoteOnDispute`)
The dispute is now open for voting by all GNOT holders. The system uses a **commit-reveal scheme** to prevent vote-copying.

- **Commit Phase**: During the `DisputeDuration`, voters submit a hash of their vote (`SHA256(value + salt)`) by calling `VoteOnDispute`. They must also pay a small `VotePrice` fee. // todo
- **Reveal Phase**: After the commit phase ends, the `RevealDuration` begins. Voters must call `RevealVote`, submitting their original `value` and `salt`. The contract verifies that the hash matches the one submitted during the commit phase.

### 6. Dispute Resolution (`ResolveDispute`)
Once the reveal period is over, anyone can call `ResolveDispute`.
- The `resolver.gno` contract tallies the votes. The winning value is the one with the highest cumulative token weight from voters.
- The `WinningValue` is set in the original `DataRequest`.
- **Slashing & Rewards**: The party (Proposer or Disputer) that lost the vote has their bond slashed. The winning party gets their bond back, and the slashed bond is distributed among the voters who voted for the winning outcome.

## Architecture

The oracle is composed of three main contracts:

- `oracle.gno`: Manages the data request lifecycle (request, propose, dispute, resolve). It is the main entry point for users.
- `court.gno`: Handles the entire dispute resolution process, including the commit-reveal voting scheme.
- `resolver.gno`: Contains the business logic for tallying votes and determining the winning value of a dispute. It supports both Yes/No and Numeric resolutions.

## User Roles

- **Requester**: The user or contract that needs external data. They create the request and fund the reward.
- **Proposer**: The user who provides the initial answer to a data request and posts a bond.
- **Disputer**: A user who challenges a proposed value and posts a bond to initiate a vote.
- **Voter**: A GNOT holder who participates in a dispute by voting on the correct outcome.

## Usage Example

Here is a full workflow using `gnokey`.

**1. Request Data**
```bash
# Ask a Yes/No question: "Will ETH be below $4000 ?" (replace DEADLINE_TIMESTAMP with a future unix timestamp more than 24h from now)
gnokey maketx call -pkgpath "gno.land/r/intermarch3/oo" -func "RequestData" -args "Will ETH be below 4000$ ?" -args "true" -args "DEADLINE_TIMESTAMP" --gas-fee 1000000ugnot --gas-wanted 5000000 --send "1000000ugnot" --broadcast true --chainid "dev" --remote "tcp://127.0.0.1:26657" <your-key-name>
```

**2. Propose a Value**
```bash
# Propose "Yes" (value 1) (replace ID with the actual ID returned from the RequestData call)
gnokey maketx call -pkgpath "gno.land/r/intermarch3/oo" -func "ProposeValue" -args "ID" -args "1" --gas-fee 1000000ugnot --gas-wanted 10000000 --send "2000000ugnot" --broadcast true --chainid "dev" --remote "tcp://127.0.0.1:26657" <proposer-key-name>
```

**3. Dispute the Value**
```bash
# Dispute the proposal (replace ID with the actual ID)
gnokey maketx call -pkgpath "gno.land/r/intermarch3/oo" -func "DisputeData" -args "ID" --gas-fee 1000000ugnot --gas-wanted 5000000 --send "2000000ugnot" --broadcast true --chainid "dev" --remote "tcp://127.0.0.1:26657" <disputer-key-name>
```

**4. Vote on the Dispute**
First, generate a hash locally. Let's vote "No" (value 0) with salt "mysecret".
Hash: `sha256("0" + "mysecret")` -> `a96e0beb59a16b085a7d2b3b5ffd6e5971870aa2903c6df86f26fa908ded2e21`
```bash
# Commit the vote (replace ID with the actual ID)
gnokey maketx call -pkgpath "gno.land/r/intermarch3/oo" -func "VoteOnDispute" -args "ID" -args "a96e0beb59a16b085a7d2b3b5ffd6e5971870aa2903c6df86f26fa908ded2e21" --gas-fee 1000000ugnot --gas-wanted 5000000 --send "1000000ugnot" --broadcast true --chainid "dev" --remote "tcp://127.0.0.1:26657" <voter-key-name>
```

**5. Reveal the Vote**
```bash
# Reveal the vote after the voting period ends (replace ID with the actual ID)
gnokey maketx call -pkgpath "gno.land/r/intermarch3/oo" -func "RevealVote" -args "ID" -args "0" -args "mysecret" --gas-fee 1000000ugnot --gas-wanted 10000000 --broadcast true --chainid "dev" --remote "tcp://127.0.0.1:26657" <voter-key-name>
```

**6. Resolve the Dispute**
```bash
# After the reveal period, anyone can trigger the final resolution (replace ID with the actual ID).
gnokey maketx call -pkgpath "gno.land/r/intermarch3/oo" -func "ResolveDispute" -args "ID" --gas-fee 1000000ugnot --gas-wanted 10000000 --broadcast true --chainid "dev" --remote "tcp://127.0.0.1:26657" <any-key-name>
```  

**Warning**: When testing with `gnodev`, ensure to make transactions between waiting periods as `gnodev` does create blocks only when a transaction is made and the oracle relies on block timestamps.

## Developer

| [<img src="https://github.com/intermarch3.png?size=85" width=85><br><sub>Lucas Leclerc</sub>](https://github.com/intermarch3) |
| :---: |