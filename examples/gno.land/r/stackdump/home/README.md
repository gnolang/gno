# Coffee Shop Petri Net Variants

This document outlines the different structural and semantic variations of the **Coffee Shop** Petri net model, with explanations of their purpose and how each transformation aids different forms of analysis.

---

## 📦 Base Model: `coffeeShop()`

This is a full Petri net model representing the operational steps of a coffee shop, including:

- Boiling water
- Grinding beans
- Brewing coffee
- Pouring into a cup
- Sending the order
- Processing payment

### Features:
- Places represent ingredients and resources (e.g., `Water`, `Filter`, `Cup`)
- Transitions represent actions (e.g., `BoilWater`, `PourCoffee`)
- Includes an **inhibitor arc** from `PourCoffee` to `Payment`, encoding a "guard condition"

Used as a base for all transformations.

---

## 🔁 Structural Variants

### 1. `ForgetPlaces()` – Transition Adjacency Graph
**Name:** `Coffee Shop No Places`

- Projects the model into a transition-only graph.
- Highlights potential sequencing of activities.
- Drops token dynamics.

**Use Case:** Causal flow analysis.

### 2. `ForgetTransitions()` – Place Flow Graph
**Name:** `Coffee Shop No Transitions`

- Connects places that are causally related via transitions.
- Useful for visualizing token flow potential.

**Use Case:** Layout planning or data dependency.

---

## 🧭 Workflow Net Variant: `coffeeShopWorkflowNet()`

Enhances the base model to follow **workflow net** structure:

- Adds a **`Start`** place with initial token
- Adds an **`Exit`** place with final target
- Ensures all transitions lie on a path from `Start` to `Exit`

### Purpose:
- To make the model suitable for **soundness analysis** and **workflow validation**
- Enforces clear entry/exit points

---

## 🔁 Workflow Variant Projections

### 3. `ForgetPlaces()` on workflow
**Name:** `Coffee Shop Workflow Net No Places`

- Transition adjacency only
- Emphasizes execution order
- Suitable for transition sequencing analysis

### 4. `ForgetTransitions()` on workflow
**Name:** `Coffee Shop Workflow Net No Transitions`

- Place adjacency only
- Shows state-to-state flows
- Simplifies dependency extraction

---

## ⚠️ Inhibitor Arcs

The base model includes a **single inhibitor arc**:

```go
{Source: "PourCoffee", Target: "Payment", Inhibit: true}
```

### Semantics:
- Prevents payment from processing if coffee hasn't been poured.
- Models a **guard condition** using negation logic.

### Effect on Projections:
- `ForgetPlaces` and `ForgetTransitions` **skip inhibitor arcs** to preserve monotonicity.
- This avoids introducing **spurious transitive edges** due to anti-monotonic behavior.

### Optional Refactoring:
- Inhibitor can be modeled structurally by adding a control token or making the logic implicit in the workflow net design.

---

## 🔧 Implementation Notes

Each model is registered with a description and keyword set using `register()` calls.

This enables structured discovery of the following:
- Full model
- Workflow variant
- Structure-only graphs (Places-only / Transitions-only)

Keywords like `no_places`, `workflow_net`, `payment`, etc., allow fine-grained search or filtering.

---

## 🧩 Summary Table

| Model Name                              | Description                                | Use Case                            |
|----------------------------------------|--------------------------------------------|-------------------------------------|
| Coffee Shop                            | Full Petri net with tokens and inhibitors  | Operational modeling                |
| Coffee Shop No Places                  | Transition-only projection                 | Causal analysis                     |
| Coffee Shop No Transitions             | Place-only projection                      | Token state mapping                 |
| Coffee Shop Workflow Net               | Workflow net with Start/Exit               | Soundness + structured analysis     |
| Coffee Shop Workflow Net No Places     | Transition flow of workflow                | Order-of-execution focus            |
| Coffee Shop Workflow Net No Transitions| Place-level state flow                     | Data/state traceability             |

---

This set of models serves both as an educational scaffold and a testbed for projection, transformation, and analysis techniques across discrete and continuous Petri net systems.

