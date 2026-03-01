# GRC-000 — Metamodel Token Standard

**Status:** Draft **Version:** v0 **Target chain:** gno.land  
**Focus:** Minimal **GRC-based balance primitives**, **composable-first** via Petri-net semantics.

---

### Document Navigation

1. **[Overview](#1-overview)**
2. **[Primitives](#2-primitives)**
    1. [Wallet](#21-wallet-account-holds-tokens)
    2. [Transfer](#22-transfer-tokens-move-between-accounts)
    3. [Mint/Burn](#23-mintburn-tokens-created-or-destroyed)
    4. [Token](#24-token-combines-wallet-transfer-mintburn)
3. **[Terms](#3-terms)**
4. **[Petri-Net Ports & Composition](#4-petri-net-ports--composition)**
    1. [Port Definitions](#41-port-definitions)
    2. [Port Usage](#42-port-usage)
5. **[Composition Principles & Append-Only Evolution](#5-composition-principles--append-only-evolution)**
    1. [Principles](#51-principles)
    2. [Why Append-Only?](#52-why-append-only)
    3. [Deployment Lifecycle](#53-deployment-lifecycle)
6. **[GRC000 as a Little Language](#6-grc000-as-a-little-language)**
    1. [Primitives as Vocabulary](#61-primitives-as-vocabulary)
    2. [Combinators as Phrases](#62-combinators-as-phrases)
    3. [Model Composition as Nested Syntax](#63-model-composition-as-nested-syntax)
    4. [Rendering as Translation](#64-rendering-as-translation)
7. **[Token Standards: Interface vs. Behavior-First](#7-token-standards-interface-vs-behavior-first)**
    1. [Two paradigms, side-by-side](#71-two-paradigms-side-by-side)
    2. [What you gain with behavior-first](#72-what-you-gain-with-behavior-first)
    3. [What you lose (or must manage)](#73-what-you-lose-or-must-manage)
    4. [Token standards, concretely](#74-token-standards-concretely)
    5. [Migration / adoption pattern](#75-migration--adoption-pattern)
    6. [When to choose which](#76-when-to-choose-which)
8. **[References](#8-references)**

---
## 1. Overview

**GRC-000** specifies a **safe, deterministic, minimal** interface for fungible tokens on `gno.land`, designed for **Petri-net composition**.  

The aim: provide a **core set of primitives**—small, deterministic Petri-net fragments—that can be composed into more complex transaction models.

This follows the **“Smart Objects, Dumb Code”** principle [(Yoneda-inspired)](https://ncatlab.org/nlab/show/Yoneda+lemma): each object’s identity is defined by behavior in all contexts, with internal rules embedded in the object and minimal external orchestration.

These primitives are developed under **append-only** rules: new behaviors are added without mutating existing ones, preserving determinism and backward compatibility.

---

## 2. Primitives

A **primitive** is a minimal Petri-net fragment capturing one atomic behavior.

- Trivial alone, powerful when **composed**.
- Deterministic, self-contained, and reusable.
- Designed for **append-only growth**: the primitive’s logic is fixed, but new primitives can be added to extend functionality.
- These basic elements combine to form DEXs, staking, DAOs, faucets—without redefining basics.
### 2.1 **Wallet:** account holds tokens.
 ```go
func GRC000_Wallet(opts map[string]any) *mm.Model {
    return mm.New(map[string]mm.Place{
        "$wallet": {Initial: initial, Capacity: mm.T(0), X: 160, Y: 180},
    })
}
```
### 2.2 **Transfer:** tokens move between accounts.
```go
func GRC000_Transfer(opts map[string]any) *mm.Model {
    return mm.New(
        map[string]mm.Place{
            "$recipient": {Initial: recipientInitial, X: 360, Y: 180},
        },
        map[string]mm.Transition{
            "transfer": {X: 210, Y: 180},
        },
        []mm.Arrow{
            {Source: "$wallet", Target: "transfer", Weight: transferWeight},
            {Source: "transfer", Target: "$recipient", Weight: transferWeight},
        },
    )
}
```
### 2.3 **Mint/Burn:** tokens created or destroyed.
```go
func GRC000_MintBurn(opts map[string]any) *mm.Model {
    return mm.New(
        map[string]mm.Place{
            "$minter": {Initial: minterInitial, X: 60, Y: 180},
            "$burner": {Initial: burnerInitial, X: 360, Y: 180},
        },
        map[string]mm.Transition{
            "mint": {X: 210, Y: 120},
            "burn": {X: 210, Y: 240},
        },
        []mm.Arrow{
            {Source: "$minter", Target: "mint", Weight: mintWeight},
            {Source: "mint", Target: "$wallet", Weight: mintWeight},
            {Source: "$burner", Target: "burn", Weight: burnWeight},
            {Source: "burn", Target: "$wallet", Weight: burnWeight},
        },
    )
}
```

### 2.4 **Token:** combines wallet, transfer, mint/burn.
```go
func GRC000_Token(opts map[string]any) *mm.Model {
    return mm.New(
        GRC000_Wallet(opts),
        GRC000_Transfer(opts),
        GRC000_MintBurn(opts),
    )
}
```


---

## 3. Terms

| Term                                                     | Meaning                                                                                    |
|----------------------------------------------------------|--------------------------------------------------------------------------------------------|
| **Account**                                              | `gno.land` address string.                                                                 |
| **Amount**                                               | `uint64` (upgrade path to `uint256`).                                                      |
| **Port**                                                 | Named Petri-net boundary place (e.g. `$wallet`, `$recipient`).                             |
| **Primitive**                                            | Atomic Petri-net submodel for composition.                                                 |
| **Append-Only**                                          | A design constraint ensuring primitives evolve by addition only—never removal or mutation. |
| **Deterministic**                                        | Behavior is predictable and consistent across all contexts.                             |
| [**Petri-Net**](https://ncatlab.org/nlab/show/Petri+net) | A mathematical model of computation using places, transitions, and arrows

---

## 4. Petri-Net Ports & Composition

Named **ports** are consistent across primitives, enabling direct integration into larger Petri-nets.

Primitives are **composed** into useful models, and models are bound to code by these ports.
This allows for **deterministic behavior** and **reproducibility** across different contexts

### 4.1 Port Definitions
By design each port name starts with `$` to distinguish them from places in the Petri-net.

- **$wallet**: Holds the token balance.
- **$recipient**: Destination for transfers.
- **$minter**: Source for minting new tokens.

### 4.2 Port Usage
Models use ports as a means for code to interact with the Petri-net.

When executing a model, ports are populated with values from the context (e.g. a transaction).

---

## 5. Composition Principles & Append-Only Evolution

### 5.1 Principles
1. **Stable Ports:** `$wallet`, `$recipient`, `$minter` stay consistent for easy integration.
2. **Append-Only Development:**
    - **Add** new primitives or transitions.
    - **Never** change or remove existing ones.
    - All past states and invariants remain valid.
3. **Merge by Ports:**
    - Unify shared ports.
    - Append transitions/arrows.
    - Preserve determinism.

### 5.2 Why Append-Only?
- **Composable Upgrades:** New features can be layered without breaking integrations.
- **Trust Preservation:** Contracts referencing old primitives remain correct.
- **Reproducibility:** State can be replayed from genesis without divergence.
- **Formal Verification:** Old proofs remain valid.

### 5.3 Deployment Lifecycle
1. **Design:** Draft primitive with fixed ports & logic.
2. **Simulation:** Test in isolation and in compositions.
3. **Freeze:** Lock logic; assign version.
4. **Publish:** Deploy to GRC-000 standard library.
5. **Compose:** Integrate with existing primitives in append-only fashion.

---

## 6. GRC000 as a Little Language

### 6.1 Primitives as Vocabulary
- **Place** → noun (state: `$wallet`, `$recipient`, `$minter`)
- **Transition** → verb (action: `transfer`, `mint`, `burn`)
- **Arrow** → grammar rule (mapping how verbs consume/produce nouns)

These are the alphabet and grammar of the *inner* Petri-net language.

---

### 6.2 Combinators as Phrases
Functions like `GRC000_Wallet`, `GRC000_Transfer`, and `GRC000_MintBurn` are idioms in this inner language — sentences that express atomic token behaviors.

---

### 6.3 Model Composition as Nested Syntax
`GRC000_Token` composes smaller idioms into paragraphs.  
The **outer Go syntax** orchestrates; the **inner Petri-net language** defines behavior.  
Append-only rules ensure new “paragraphs” never rewrite history.

---

### 6.4 Rendering as Translation
`Render(path string)` translates the nested Petri-net model into Markdown + SVG — akin to compiling or translating to another language.  
Because the net is append-only, all renderings are reproducible over time.

---

## 7. Token Standards: Interface vs. Behavior-First

> We standardize **behavior, not just function names**. The token’s canonical definition is its **open Petri net**; ERC-20-style functions are **generated views** for legacy tooling.

- **ERC-20-style** = *nominal interface*: a handful of function names + event shapes. You prove conformance by matching the signature, not the whole behavior.
- **OPetri / typed-Petri approach** = *behavioral specification*: you spell out **every allowed state transition**, and interfaces are just *views* over that model.

---

### 7.1 Two paradigms, side-by-side

| Dimension | Interface-first (ERC-20) | Behavior-first (OPetri / typed Petri) |
|---|---|---|
| Spec unit | Function signatures (`transfer`, `approve`, …) | Places, transitions, guards, arcs (typed) |
| Conformance | “Implements the ABI” | “Implements these transitions and invariants” |
| Guarantees | Minimal (must not revert in obvious cases) | Rich: invariants, conservation laws, liveness, safety |
| Extensibility | Add new funcs / extensions (fragmentation risk) | Compose subnets; refinement preserves proofs |
| Composability | Name/ABI matching; adapters required | Structural: glue along shared `$objects` (ports) |
| Analysis | Hard: need bespoke proofs per impl | Native: reachability, invariants, ODE analysis |
| Concurrency | Not modeled; relies on VM/nonce | First-class: enabling conditions, mutual exclusion |
| Upgrades | “New interface” or proxy patterns | Add/replace subnets; reuse proofs via refinement |
| Testing | Unit/integration by example | Model checking, simulation, parameter sweeps |
| Failure modes | Same ABI, wildly different semantics | Same ports ⇒ bounded semantic variation |

---

### 7.2 What you gain with behavior-first
- **Total semantics**: No “mystery paths.” Every state change is one of your transitions.
- **Machine-checkable properties**:
   - *Conservation*: token mass is preserved except at mint/burn.
   - *Safety*: guards prevent underflow/overdraft.
   - *Liveness*: absence of deadlocks; intended flows always possible.
- **True composability**: DEXes, escrows, staking = subnets that **plug into** `$token` and `$wallet`.
- **Refinement without fragmentation**: Extensions (`permit`, `batch`) = new subnets; existing proofs still valid.

---

### 7.3 What you lose (or must manage)
- **Higher upfront modeling cost**: You design a net, not just an ABI.
- **Learning curve**: Readers must understand places, transitions, guards.
- **Interop optics**: World “speaks ERC-20.” You’ll expose an ERC-20-compatible view.

---

### 7.4 Token standards, concretely

**Model (canonical):**
- Places: `$wallet`, `$recipient`, `$supply`, `$allowance[(owner,spend)]`
- Transitions: `mint`, `burn`, `transfer`, `approve`, `transferFrom`, …
- Guards: balances/allowances ≥ weight; role checks; time locks optional.
- Invariants:
   - `Σ tokens($wallet_i) + tokens($supply) = constant + minted − burned`
   - No negative markings; respect capacity bounds.

**Interface (derived view):**
- `balanceOf(a)` ⇒ marking at `$wallet[a]`
- `totalSupply()` ⇒ marking at `$supply`
- `transfer(to, amt)` ⇒ fire `transfer` if enabled
- `Approval`, `Transfer` events ⇒ emitted on transitions

---

### 7.5 Migration / adoption pattern
1. **Author the net** (OPetri/typed Petri).
2. **Auto-derive ABI** (queries + callable transitions).
3. **Ship both**: net + thin ABI adapter.
4. **Prove**: conservation, safety, allowance monotonicity.
5. **Compose**: plug into DEX/vesting/staking subnets.

---

### 7.6 When to choose which
- **Interface-first**: fastest baseline compatibility.
- **Behavior-first**: for nontrivial policy, assurance, composability.

## 8. References

- [Open Petri Nets](https://arxiv.org/abs/1808.05415) - a foundational paper on Petri nets.
- [Petri Nets](https://ncatlab.org/nlab/show/Petri+net) - all-purpose reference.
- [Yoneda Lemma](https://ncatlab.org/nlab/show/Yoneda+lemma) - foundational concept in category theory, relevant to compositionality.
- [ERC-20](https://eips.ethereum.org/EIPS/eip-20) - the classic token standard.
- [Cross-chain Deals and Adversarial Commerce](https://arxiv.org/abs/1905.09743)
