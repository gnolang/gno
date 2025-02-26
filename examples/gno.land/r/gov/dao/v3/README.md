# GOVDAO Specifications

## 1. Overview
**GovDAO** is a governance body with three membership tiers—**T1**, **T2**, and **T3**—along with defined voting power, membership requirements, size targets, and compensation policies. 

---

## 2. Membership Tiers

### 2.1 Tier Definitions
- [X] **T1 (Core Tier)**  
  - [X] Highest tier; self-selecting membership with a *supermajority* vote from T1.
  - [X] Membership can only be withdrawn by supermajority vote *with cause*.
- [X] **T2**  
  - [X] Selected by GovDAO with T3 abstaining, requiring a *simple majority* vote.
  - [X] Membership can be withdrawn for any reason.
- [X] **T3 (General Tier)**  
  - [X] Lowest tier; *permissionless invitation* from T1 and T2 members.
  - [X] Membership can be withdrawn for any reason.

### 2.2 Age Limit
- [ ] **Maximum age of 70** for any member. Once a member reaches 70, membership is automatically withdrawn. (NOT IMPLEMENTED AT THE END)

---

## 3. Voting Power

### 3.1 Per-Member Voting Power
- [X] **T1 member**: 3 votes each.
- [X] **T2 member**: 2 votes each.
- [X] **T3 member**: 1 vote each.

### 3.2 Class Voting Power Caps
- [X] **T2 class** is capped at *2/3* of the total T1 voting power.
- [X] **T3 class** is capped at *1/3* of the total T1 voting power.

In practice, this often results in a voting-power ratio of:  
- [X] **T1**: 1/2 of total votes  
- [X] **T2**: 1/3 of total votes  
- [X] **T3**: 1/6 of total votes  

#### Examples
1. [X] **Example 1**  
   - T1: 100 members → 300 VP (3 votes/member)  
   - T2: 100 members → 200 VP (2 votes/member)  
   - T3: 100 members → 100 VP (1 vote/member)
2. [X] **Example 2**  
   - T1: 100 members → 300 VP (3 votes/member)  
   - T2:  50 members → 100 VP (2 votes/member)  
   - T3:  10 members →  10 VP (1 vote/member)
3. [X] **Example 3**  
   - T1: 100 members → 300 VP (3 votes/member)  
   - T2: 200 members → 200 VP (effectively 1 vote/member)  
   - T3: 100 members → 100 VP (1 vote/member)
4. [X] **Example 4**  
   - T1: 100 members  → 300 VP (3 votes/member)  
   - T2: 200 members  → 200 VP (1 vote/member)  
   - T3: 1000 members → 100 VP (~0.1 vote/member)

---

## 4. Basic Membership Requirements

1. **Known Identities.** All members must be known, identifiable individuals.  
2. **Tier Criteria.**  
   - **T1**: Must meet T1, T2, and T3 criteria.
     - Expertise in relevant categories.
     - Significant contributions.
     - Demonstrated value alignment.  
   - **T2**: Must meet T2 and T3 criteria.
     - Expertise in relevant categories.
     - Ongoing contributions.  
   - **T3**: Must meet T3 criteria.
     - Expertise in relevant categories.
     - Ongoing contributions.
3. **Membership Proposals.**  
   - [X] T1 and T2 members are added via individual proposals.
   - [X] Each proposal should include a Markdown resume/portfolio.

---

## 5. T1 Membership Size

### 5.1 Target Minimum
- **Target**: Minimum of 70 T1 members after 7 years.

### 5.2 Quarterly Additions
- If below 70 members:
  - **Should** add 2 new T1 members per quarter.
  - **Tolerated** to add only 1 new member if circumstances require.

### 5.3 Election by GovDAO
- If below 70 T1 members **and** 2 years have passed **and** no T1 members were added in a given quarter **and** qualified candidates exist:
  - **GovDAO** may elect 1 qualified candidate, with T1 abstaining.

### 5.4 Election from GNOTDao
- If still below 70 T1 members under the same conditions:
  - **GNOTDao** may elect 1 qualified candidate if it exists and is approved by GovDAO.

---

## 6. T2 Membership Size

- [ ] **Maximum Target**: `2 × size(T1)`.  
- [ ] If the T2 membership already exceeds `2 × size(T1)`, *no additional T2 members* can be added.  
- [ ] No *formal minimum* T2 size, but *desired* minimum is at least `floor(size(T1)/4)`.

---

## 7. T3 Membership Size

- **Invitation Points**:
  - [X] Each **T1** member has 3 invitation points.
  - [X] Each **T2** member has 2 invitation points.
  - [X] Each **T3** member has 1 invitation point.
- **Requirement for T3 Membership**: 
  - [ ] A total of 2 invitation points from at least 2 existing members.
  - [ ] Invitations can be withdrawn at any time.

---

## 8. Payment Policy

### 8.1 Eligibility and Amount
- [ ] **T1 and T2** members *may* be paid equally if they are actively contributing.
- [ ] Payment level is capped at the 90th percentile of senior software architect roles in the highest-paid city globally.

### 8.2 Pay Capacity
- [ ] **T1T2PaySize** = The number of members (across T1 and T2) who can receive compensation.
- [ ] **T1T2PaySize** = `min(70, T1T2PayCapacity)`  
- [ ] **T1T2PayCapacity** is determined by the **GovDAOPayTreasury**:
  - [ ] The treasury must be able to fund these members for at least 7 years.
  - [ ] If the treasury shrinks, the pay capacity may shrink accordingly.
  - [ ] *Seniority* determines pay priority.

### 8.3 Treasury Composition
- [ ] **GovDAOPayTreasury**: 
  - [ ] Must hold at least 1/3 in GNOT.
  - [ ] Remaining assets can only be stable tokens or Bitcoin.

### 8.4 Conflict of Interest and Profit Sharing
- [ ] All paid T1/T2 members must agree to a conflict-of-interest and profit-sharing policy.

---

## 9. Validators

- [ ] **GovDAO**’s effective voting power (3, 2, or 1 vote per tier) is also used for delegating to **at least 70 validators**.
- [ ] The exact number of validators is determined by GovDAO.
- [ ] A separate **ValidatorTreasury** exists, following similar rules as the GovDAOPayTreasury.

---

## 10. Forking Gno.land

- [ ] A **+1/3** of total voting power can initiate a fork.
- [ ] All deployed contracts may exist on both forks (limitations may apply).
- [ ] **T1, T2, T3** members must choose which fork to remain exclusive to.
- [ ] Determining the name “gno.land” upon a fork:
  - [ ] If **2/3+ of T1** decide on one fork, that fork keeps the name.
  - [ ] Or if a **majority of GovDAO** decides, that fork keeps the name.
  - [ ] Otherwise, both forks must rename.
  - [ ] Any name including “gno” must also be acceptable to NT, LLC.

---

## 11. GNOTDao

1. [ ] May elect T1 members in scenarios where T1 membership is insufficient (per [Section 5](#5-t1-membership-size)).
2. [ ] Must approve any proposals involving GNOT inflation, which also require quorum and supermajority of GNOT holders.

### 11.1 Token Distribution
- [ ] **70%** airdropped.
- [ ] **25%** to NT, LLC.
- [ ] **5%** reserved for user acquisition.

### 11.2 Additional Inflation
- [ ] If GNOTDao exists within 2 years and does **not veto**, then:
  - [ ] Another **10%** (equal to `GNOT/9`) may be inflated for further user acquisition.

---

## 12. Change Log and Amendments

- [X] No automatically withdrawn by age.