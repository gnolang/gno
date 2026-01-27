# Gno.land Constitution

This Constitution is still a draft and until $GNOT is transferrable any
Constitutional Amendments shall require the pre-approval of NT,LLC.

When there are any ambiguities or conflicts within this document NT,LLC
reserves the right to clarify until $GNOT launch; thereafter ambiguities can
only be resolved by Supermajority Decision of GovDAO followed by a
Constitutional Amendment.

## Terms

 * AIB,Inc: All in Bits, Inc.

 * NT,LLC: NewTendermint, LLC.

 * Constitution: This Gno.land constitution document including modifications by
   a Constitutional Amendments. This Constitution is considered independent of
   the Atom.One Constitution although some portions of the Atom.One
   Constitution may be referenced here for clarity.

 * Constitutional Amendment: A modification to this Constitution by a
   Constitutional Majority Decision of GovDAO and any other requirements as
   defined in this Constitution.

 * Majority, Supermajority, and Constitutional Majority Decision: See "GovDAO
   Decisions"

 * Oversight Body: Initially NT,LLC as represented by a multisig account, but
   eventually represented by the Oversight DAO on Gno.land.

 * Core DAO: GovDAO, or any subDAO created by GovDAO or Core DAOs, The
   Oversight Body (DAO) and its subDAOs are not Core DAOs as they are meant to
   be somewhat independent of GovDAO.

 * XXX

## Genesis Allocation

At Gno.land Genesis there will be one billion $GNOT tokens.

 * Airdrop1:    35.0% - from partial Cosmos governance snapshot 3 years ago
 * Airdrop2:    23.1% - from recent AtomOne snapshot prior to launch
 * GovDAO:      11.9% - for prior and future Gno.land ecosystem contributors
 * Investors:    7.0% - reserved for current and future investments
 * NT,LLC:      23.0% - of which a significant portion is allocated for prior loans

$GNOT will not be transferrable initially except for whitelisted addresses.
Whitelisted addresses include "GovDAO" and "Investors" funds and any additional
addresses dedicated for faucets.

The 7% (qualfied, private) investors allocation will be held by NT,LLC in a
segregated account. Proceeds of sales of these tokens will go to NT,LLC for
past or future development of Gno.land, Gno, Tendermin2, other Core Software,
and ecosystem development.

GovDAO is responsible for distributing $GNOT to prior and future Gno.land
ecosystem contributors (as well as those contributing to the blockchain stack,
including external contributors to Tendermint2, GnoVM, Gno.land server and
tooling, GnoWeb) with the exclusion of any current GovDAO members.

Up to one-third of the 11.9% genesis allocation wil be distributed to prior
contributors. Prior and current NT,LLC and AIB,Inc employees may be rewarded
with the approval of NT,LLC and GovDAO except the primary owners and executives
of NT,LLC or AIB,Inc.

NT,LLC owes around $20M worth of $GNOT to All in Bits,Inc at the "fair market
value" of $GNOT at a 25% discount for a "fair market value" which is yet to be
determined as of the date of this writing.

## Gno.land Tokenomics

> $ATONE  : like a Bitcoin miner\
> $PHOTON : like a BTC token\
> $GNOT   : storage deposit token

First a primer on Atom.One tokenomics. $ATONE and $PHOTON are tightly coupled:

 * $ATONE is an inflationary staking token capped at 20% inflation.
 * $PHOTON is a deflationary transaction gas-fee (spam-prevention) token.
 * If all $ATONE burned there would be one billion $PHOTON.
 * Burning is one way; $PHOTON cannot be burned back to $ATONE.
 * $ATONE targets 2/3 to be staked, thus inflationary staking "rewards" are not
   income.
   * Any tax authority that says otherwise is robbing you.

While $PHOTON has no voting rights and cannot be used for staking on Atom.One,
it nevertheless has unique advantages enforced by the Atom.One constitution:

 * All transaction fee payments on Atom.One must be paid with $PHOTON.
 * Atom.One VaaS chains (aka ICS shards) including Gno.land after migration
   must accept $PHOTON as the gas-fee token.
   * Each may also accept other gas-fee tokens but ultimately the chain must
     pay $PHOTON to Atom.One for security.
 * Atom.One VaaS sovereign (consumer) chains may exit (change validator-sets)
   Atom.One and become self-validating or become hosted by another VaaS hub.

Gno.land will launch independently of Atom.One so Gno.land will initialy
require $GNOT for transaction fee payment. Before and after Gno.land migrates
to be hosted by Atom.One VaaS $GNOT will function as a byte storage deposit
token.

 * Every transaction that ends up increasing the amount of persistent state
   will require a bond deposit of $GNOT.
 * Every transaction that ends up freeing up persistent state space will
   receive a refund of $GNOT.
 * One billion $GNOT corresponds to 10TB of persistent state space.
 * The $GNOT inflation schedule will never change, thus the total created $GNOT
   will never exceed 1.333~ billion $GNOT.
 * The $GNOT storage deposit rate (per byte) will never increase.
 * The $GNOT storage deposit rate will be such that the total remaining $GNOT.
   and all future inflationary $GNOT never exceeds 20TB of state (double the
   original 10TB). This is to keep the blockchain state at an accessible level
   especailly accounting for any future potential economic collapses.
 * $GNOT is only considered "burned" and according to approved burn mechanisms
   or transactions (like $ATONE "burn" to $PHOTON except $GNOT burning does not
   yield any new tokens or rights on Gno.land). The automatic burning of $GNOT
   in a every Realms' SDDCAs (Storage Deposit Discoungt Credit Accounts) are
   the only automatic burn mechanism of $GNOT but the Gno.land Constitution may
   also approve of other official burn mechanisms that are reasonable,
   VOLUNTARY, and EXPLICIT. Such burning mechanisms may be used to yield
   storage deposit tokens on other chains.

Gno.land is obligated to migrate to be hosted/secured by Atom.One VaaS
("Validator-as-a-Service") when it is deemed ready by both Atom.One and GovDAO
by Supermajority Decision.

After migratiion to Atom.One VaaS hosting Gno.land should pay the Atom.One
chain in $PHOTONs underneath the hood as the Atom.One constitution requires; an
automated-market-maker exchange module should exist on the Gno.land
shard/instance to facilitate the internal exchange of collected $GNOT to
$PHOTON needed to pay Atom.One for its VaaS services.

Gno.land will initially launch as its own blockchain so the $GNOT token will
function both as the spam-prevention gas-payment token as well as byte-storage
deposit token. Gno.land after launch will merge with Atom.One and be hosted as
an Atom.One ICS chain that is secured by the same validator-set as Atom.One.

Once Gno.land migrates over to Atom.One after the Gno.land <> Atom.One IBC
connection is complete and Atom.One Simple-Replicated ICS MVP is implemented,
$ATONE will be the staking-token (but with limited voting rights for Gno.land
itself), $PHOTON will be the CPU gas-token, and $GNOT the dedicated
byte-storage deposit token. Thus Gno.land will become a key VaaS hosted
application on Atom.One, but other VaaS chains hosted by Atom.One may not offer
the Gno VM or Gno programmability; and even if such chains were a simple fork
of Gno.land they may operate independently of the $GNOT token. 

The $GNOT token should not be used as a storage deposit token or for any other
chain except the original Gno.land chain, nor be used as a staking/bonding
token as this limits the utility (accessible storage capacity) of Gno.land
itself. It may make sense to burn $GNOT via approved burn mechisms to acquire
secondary tokens that can be used for storage deposit on other GnoVM (and
non-Gno) chains hosted on Atom.One.

There will be many more chains hosted with Atom.One ICS that are powered by the
GnoVM or compete with the Gno.land chain itself, but these chains will need to
give Gno.land strong-attribution by the Gno Network GPL copyleft license (a
fork of AGPL3.0 to allow for strong-attribution in a decentralized blockchain
ecosystem with many independent actors), and Gno.land will be the first such
Gno-based chain, so Gno.land and $GNOT will benefit from first-mover advantage
and network effect even if other chains do not use the $GNOT token at all.

Competing smart-contract platforms that are not based on the GnoVM, or those
that are based on other languages will also be supported in Atom.One, as
Atom.One will support quasi-permissionless hosting of many blockchain
application platforms; but Gno and GnoVM will serve as a foundation for future
multi-user general-purpose language innovation.

## Pre-Atom.One Migration Validators

Until Atom.One VaaS/ICS is ready Gno.land the validator selection mechanism is
determined by GovDAO Supermajority Decision.

Priority should be given to Atom.One validators and Gno.land core developers.

The number of validators prior to migration shall not exceed 50.

No validators may operate on any cloud hosting provider, but must run their own
hardware where they have physical access to the server at all times (such as in
a co-location provider with 24/hour access).

Atom.One migration is not contingent on all of its validators running their
their own hardware as above, but GovDAO may impose a requirement by Majority
Decision for Atom.One to have a completed roadmap specified to get there.

Atom.One VaaS/ICS shall not be deemed suitable unless the Gno.land chain
remains whole (not part of any "mega-block" system where the consensus engine
process is shared with other applications) and Gno.land may migrate away from
Atom.One by on-chain transactions.

If Atom.One validators do not largely (> 90% by voting power) run on their own
hardware where validators have physical access to their server at all times 2
years after migration or 4 years after Gno.land launch whichever is sooner,
GovDAO with a Supermajority Decision may choose to fund the development of a
fork Atom.One in the likeness of the Cosmos -> Atom.One fork with a relevant
governance proposal on Atom.One (with preference to the class of voters who
voted in favor of this requirement); provided that GovDAO first submits
reasonable proposals to Atom.One that get rejected by Atom.One.

Likewise if the Atom.One staking incentive system is not such that the top
validator has at most three times the voting power of the lowest validator 2
years after migration or 4 years after Gno.land launch whichever is sooner,
GovDAO with a Supermajority Decision may choose to fund the development of a
fork Atom.One in the likeness of the Cosmos -> Atom.One fork with a relevant
governance proposal on Atom.One (with preference to the class of voters who
voted in favor of this requirement); provided that GovDAO first submits
reasonabe proposals to Atom.One that get rejected by Atom.One.

In the case of an Atom.One fork initiated by GovDAO by the above, the fork need
not run any chain except those that run Gno. The constitution of the fork shall
be determined by Supermajority Decision of GovDAO and respect the rights of
NT,LLC with respect to the "Gno" word mark.

## $GNOT (Deflationary) Inflation

From the date of launch every year 33.33*(0.9^Y) million $GNOT 3.333% of the
Gno.land Genesis $GNOT supply will be inflated continuously as follows (where Y
is the year from launch starting from 0). Any inflationary $GNOT not
transferred will accrue until the Gno.land software is updated to transfer such
funds.

 * After   3 years this represents  90.32 million $GNOT.
 * After  10 years this represents 217.09 million $GNOT.
 * After 100 years this represents 333.29 million $GNOT.

At most there will be 1.333... billion $GNOT, representing a total inflation of
one third of the genesis $GNOT distribution. This amount cannot change even
with amendments to the Gno.land Constitution. This makes $GNOT a deflationary
token similar to Bitcoin.

The inflated tokens will be distributed as follows:

 * One third will go to NewTendermint,LLC -- all if it will go toward Gno.land,
   Gno, and Tendermint ecosystem development and maintenance.  and be held to
   account on chain.
 * One third will go to GovDAO -- all of which must go toward external
   ecosystem contributors according the Gno.land Constitution, its Laws, and
   any applicable Mandates and Bylaws (altogether Governing Documents).
 * One third will go to PayTreasury, ValTreasury, ServicesTreasury, and
   RecompenseTreasury as determined by GovDAO and NewTendermint,LLC. In the
   absence of any existing agreement they will be split equally among these
   treasuries.

The Gno.land treasuries are as follows:

 * GovDAOTreasury is the treasury for GovDAO.
 * PayTreasury can only be used to pay GovDAO members.
 * ValTreasury can only be used to pay validators or Atom.One VaaS hosting fees
   after conversion to $PHOTON.
 * ServicesTreasury can only be used to pay for the creation and
   maintenance of services for Gno.land such as blockchain explorers, indexers,
   wallets, relayers, full nodes, and file hosting servers.
 * RecompenseTreasury can only be used to recompense victims.

Each of the above treasuries are dictated by the Gno.land Governing Documents.
GovDAOTreasury must be managed algorithmically by code that adheres to the
aforementioned documents and be approved by both GovDAO and NewTendermint,LLC.
The other treasuries are not intiially managed by any DAO or committee unless
specified in the Governing Documents.

Each of the above treasuries may accrue unspent inflated $GNOT. Each of these
treasuries may adopt a diversification strategy as approved by the Governing
Documents (some of which is described later).

Funds that are assigned to any of these treasuries may not be clawed back or
transferred to any other DAO or treasury without a Constitutional Amendment.

## Gno.land Revenue

Transaction fees collected in $GNOT or $PHOTON on Gno.land are called Revenue.
The Revenue is distributed according to the following rules in order:

 1. ValTreasury is funded with 75% of Revenue if ValTreasury has less than 1 year of runway, or
 2. ValTreasury is funded with 50% of Revenue if ValTreasury has less than 4 years of runway, or
 3. ValTreasury is funded with 25% of Revenue if ValTreasury has less than 7 years of runway, or
 4. ValTreasury is funded with 10% of Revenue if ValTreasury has more than 7 years of runway.

The portion of Revenue that is not allocated to ValTreasury is called Remaining Revenue 1.

 5. PayTreasury is funded with 75% of Remaining Revenue 1 if PayTreasury has less than 1 year of runway, or
 6. PayTreasury is funded with 50% of Remaining Revenue 1 if PayTreasury has less than 4 years of runway, or
 7. PayTreasury is funded with 25% of Remaining Revenue 1 if PayTreasury has less than 7 years of runway, or
 8. PayTreasury is funded with 10% of Remaining Revenue 1 if PayTreasury has more than 7 years of runway.

PayTreasury has strictly lower priority than ValTreasury because GovDAO members
can theoretically still function without pay, whereas the blockchain cannot
function securely without validators. The PayTreasury runway must take into
account future projected growth of the paid T1 and T2 members.

The portion of Remaining Revenue 1 that is not allocated to PayTreasury is
called Remaining Revenue 2.

 9. ServicesTreasury is funded with min(Remaining Revenue 2, 25% of Revenue).
 10. ReserveTreasury is funded with the remainder from above.

ServicesTreasury has strictly lower priority than ValTreasury or PayTreasury
because there is some elasticity to the variety of services that can be
offered, much of which should become mature and finalized; and some of the
essential services could be migrated to be offered by all of the validators
instead (especially with the help of GovDAO members); and services should
generally pay for themselves with micropayments or subscriptions or by other
means.

The ReserveTreasury has no governing body of its own, and any transfer of funds
from the ReserveTreasury requires a new Constitutional Amendment and must go
toward a specialized DAO with its own mandate and bylaws, and may or may not be
controlled by GovDAO.

## Excess $GNOT Deposits

When the $GNOT storage deposit rate per byte is decreased this results in
excess $GNOT deposits per realm. This is not considered part of Revenue.

When the $GNOT storage deposit rate decreases (not automatically by the $GNOT
burn mechanism described below, but by the decision of GovDAO to lower the
rate) 25% of the excess $GNOT goes into the AuditTreasury, 25% of the excess
$GNOT goes into the RecompenseTreasury, and the remainer goes to the
seggregated virtual account per realm ($GNOT of the realm's SDDCA) to be used
only for future discounts on transactions for that specific realm. These SDDCA
$GNOT tokens may be transferred to other SDDCAs by rules defined in the
Governing Documents but their $GNOT may never be withdrawn even upon freeing
storage, and transfers must be initiated by the present authority of the realm
or containing organization. 

DDCA $GNOTs may be burned automatically by a rate set by an amendment to the
Gno.land Constitution not to exceed 10% a year. This is to prevent stagnant
$GNOT from limiting the allocated storage capacity of Gno.land and thereby
reducing its utility. $GNOT burned in this way will also reduce the $GNOT
storage deposit rate automatically but not trigger any of the the mechanisms
described here.

The AuditTreasury may only be used to fund of the auditing of code deployed to
Gno.land by Qualified Auditors, or to assess and reward Bonded Auditors who
submit valid Bonded Vulnerability Reports.

The RecompenseTreasury may only be used to recompense victims of exploits and
fraud. They may be disbursed only if approved by GovDAO by Supermajority
Decision with members with any conflict of interests abstaining. There is no
obligation to compensate anyone. Recompensation decisions must be preceded by a
thorough analysis of the problem and tasking of a task-force to recover any
ill-gotten gains and at least two weeks for all relevant parties to review the
analysis and task-force for approval.

To prevent abuse of these non-GovDAO treasury funds no Constitutional Amendment
may be passed to use them except to burn tokens from the AuditTreasury or
RecompenseTreasury at a rate not exceeding 10% a year.

## GovDAO Structure

There are three tiers: T1, T2, and T3.

T1 is the highest tier, T3 the lowest.

 * T1: self-selecting "core" with supermajority vote from T1.
 * T2: selected by GovDAO w/ T3 abstaining, with simple majority vote.
 * T3: permissionless invitation from T1 and T2 according to a delegation
   mechanism.

The maximum age for any member is 70; after is automatically withdrawn.

T1, T2, and T3 membership can only be withdrawn by a GovDAO Supermajority
Decision WITH CAUSE and Oversight Body pre-approval. Such members are
considered "ejected". Ejected members are no longer eligible for membership in
GovDAO.

Members may resign at any time with a signed transaction. Resigned members may
be reinstated with a Supermajority Devision of GovDAO.

T1 members may not be actively contributing to the project but they must still
be active in voting or their membership shall be suspended if they miss 3
GovDAO proposal votes in a row until they return to activity with a simple
declaration. The activity status of T1 or their suspension shall not be deemed
cause for removal from T1. In this way T1 is like "tenure", and generally the
T1 members are expected to have made significant contributions to the project.
The exception to this rule is described in "US Preference Option".

## GovDAO Voting Power.

In general,
 * a T1 member gets 3 votes.
 * a T2 member gets 2 votes.
 * a T3 member gets 1 vote.

However,
 * T2 class is capped at 2/3 the voting power of T1 class.
 * T3 class is capped at 1/3 the voting power of T1 class.
 * --> T1 1/2, T2 1/3, T3 1/6 (unless...)

Example 1:
 * T1 100 members --> 300 VP, 3 votes per member
 * T2 100 members --> 200 VP, 2 votes per member
 * T3 100 members --> 100 VP, 1 votes per member

Example 2:
 * T1 100 members --> 300 VP, 3 votes per member
 * T2  50 members --> 100 VP, 2 votes per member *
 * T3  10 members -->  10 VP, 1 votes per member *

Example 3:
 * T1 100 members --> 300 VP, 3 votes per member
 * T2 200 members --> 200 VP, 1 votes per member *
 * T3 100 members --> 100 VP, 1 votes per member

Example 4:
 * T1 100 members  --> 300 VP,   3 votes per member
 * T2 200 members  --> 200 VP,   1 votes per member *
 * T3 1000 members --> 100 VP, 0.1 votes per member *

### T1 Membership Size

Target is minimum 70 members after 7 years.

If the minimum is not reached,
  2 members SHOULD be added every quarter,
  but 1 new member is tolerated.

If the minimum is not reached,
  AND 2 years has passed,
  AND no members are added for the quarter,
  AND there are Qualified Candidates,
  1 Qualified Candidate may be elected by GovDAO, but with T1 abstaining.

If the minimum is not reached,
  AND 2 years has passed,
  AND no members are added for the quarter STILL,
  AND there are Qualified Candidates,
  AND GnotDAO exists and is approved by GovDAO,
  1 Qualified Candidate may be elected from GnotDAO.

### T2 Membership Size

The T2 membership maximum target is 2 x size(T1).

While size(T2) can be greater than 2 x size(T1),
  no more members can be added to T2.

There is no minimum size of T2,
  but the desired minimum size is at least floor(size(T1)/4).

### T3 Membership Size

T1 members get 3 invitation points.
T2 members get 2 invitation points.
T3 members get 1 invitation points.

2 invitation points from 2 members must be delegated for T3 membership.
Delegation/invitation can be withdrawn at any time.

### Payment to GovDAO Members

T1 and T2 members may get paid equally, if they are actively working.

Payment for a full time working T1 or T2 member may not exceed the 90th
percentile of senior software architect roles in the second highest paid city
globally.

Members already employed by another company will not receive any compensation
unless they disclose their compensation to GovDAO T1 and T2 members; then they
may be compensated up to 50% of the usual amount to top up their net payment to
the usual limit.

The number of members who get paid T1T2PaySize is min(70, T1T2PayCapacity).

T1T2PayCapacity is determined by the size of the PayTreasury and is defined
asthe number of members who can be paid for 7 years.

If the net value of PayTreasury shrinks, T1T2PaySize may also shrink. Seniority
is used to determine pay priority regardless of T1 or T2 membership.

All T1 and T2 members must quarterly sign and circulate among each other a
conflict of interest disclosure document with NT,LLC on behalf of Gno.land or a
designated legal entity approved by NT,LLC.

## GovDAO

GovDAO is the primary governing body of Gno.land with limitations by the
Oversight Body and other terms and conditions of this Constitution.

### GovDAO Decisions

A majority decision of GovDAO requires more than 1/2 of voting power of T1, T2,
and T3 members according to the scoring mechanism defined in this Constitution.

A supermajority decision of GovDAO requires more than 2/3 of voting power of
T1, T2, and T3 members according to the scoring mechanism defined in this
Constitution.

A constitutional majority decision of GovDAO requires more than 9/10 of voting
power of T1, T2, and T3 members according to the scoring mechanism defined in
this Constitution.

Some decisions may require additional conditions to be satisifed:

 * When a pre-approval is required the proposal must attach a valid
   cryptographically signed signature of pre-approval of the exact proposal
   body bytes. If no such automated mechanism is implemented on chain the
   proposal must be rejected by all members unless the aforementioned signature
   exists and is independently verified by each voting member. For example, a
   Constitutional Amendment requires the pre-approval of the Oversight Body.

 * When a post-approval is required the proposal must be appropriately
   constructed with a field set that prevents any automated action from taking
   place upon the passage of the proposal until the relevant post-approval
   conditions are met. If no such automated mechanism is implemented on chain
   the post-approval condition must be treated as a pre-approval condition. If
   all relevant conditions are not satisifed the proposal must be rejected by
   all members.

### GovDAO Membership Criteria

#### General Membership Criteria

These GovDAO Membership General Criteria apply for all GovDAO member
candidates:

 * No candidate is eligible who is or had been associated with any entity or
   organization that has confidential membership and requires members to
   protect each other (e.g. Freemasonry). If an existing member joins such they
   must declare a "spiritual conflict of interest" and resign immediately.

 * No candidate is eligible who has or has had any association with any
   government intelligence agency or program, any defense contractor, nor any
   law enforcement agency. Candidates may have served in a military but they
   may not be actively serving.

 * No candidate is eligible who has or has had any affiliation with a communist
   or fascist or Nazi or Neo Nazi nor Zionist party or organization; nor any
   gang.

 * If an existing member gains any association that violates the above general
   requirements they must cease all GovDAO related activity and immediately
   declare their association within 24 hours and resign.

 * No candidate is eligible who has been convicted of a violent, property,
   white-collar, organized, or cyber crime.

   * Any existing member(s) who gets convicted of these crimes must declare so
     and resign or provide the details of the charges and the judgement.
     Thereafter GovDAO minus these members must vote on whether to eject the
     member with a majority decision.

The general requirements for T1, T2, and T3 membership are as follows:

 * All members must be publicly identifiable persons and declare their country
   of residence and citizenshipa and provide timely updates when it changes.

 * T1 members must meet T1,T2,T3 criteria.

 * T2 members must meet T2,T3 criteria.

 * T3 members must meet T3 criteria.

 * T1 criteria includes expertise in categories, significant contributions, and
   demonstration of value alignment, and when active must make public
   appearances promoting Gno.land or publications related to their
   contributions to Gno.land.
   
 * T2 criteria includes expertise in categories and continuing contributions
   incluing public appearances promoting Gno.land or publications related to
   their contributions to Gno.land.

 * T3 criteria includes significant experience in categories and continuing
   contributions.

 * T1 and T2 members are added via individual proposals, each with a markdown
   resume/portfolio application document.

 * Any members who have been proven to have lied or not disclosed material
   information must be demoted or ejected with a majority decision with a
   pre-approval from the Oversight Body contingent on proof and evaluation.

 * Ejected members are never again eligible for membership in T1, T2, or T3.

#### Diversification Requirements

The Diversification Requirements exists to prevent any single country or union
from dominating GovDAO thereby ensuring a globally decentralized governance
body.

 * No new candidates for T1 (or T2) may be proposed if their country or union
   of citizenship represents more than one third of T1 (or T2).

 * No new candidates for T1 (or T2) may be proposed if their country or union
   of residence represents more than one third of T1 (or T2).

#### US Preference Option

The US Preference Option exists as an exception to the Diversification
Requirements and any other restrictions on membership size limitations or rate
of election, but must satisfy the General Membership Criteria as any other
candidate.

 * NT,LLC has the option to appoint as many US citizen residents T1 (or T2)
   members as there are citizens of T1 (or T2) members of any country or union.

 * Dual citizens shall be counted as fractions where the denominator is the
   number of citizenships and the numerator is one.

 * Appointment may happen at any time and is not limited by the Diversification
   Requirements.

 * The US Preference Option may not be changed even by a Constitutional
   Amendment. This ensures that the will of the founder to ensure US
   representation in GovDAO is preserved at all times.

 * NT,LLC may decide at its option to delegate the US Preference Option to the
   subset of current T1 and T2 members who are US citizen residents.

### GovDAO Treasury Spending

Spending of funds from the GovDAO Treasury may be performed by GovDAO directly
by Supermajority Decision or transferred to another Core DAO ultimately managed
by GovDAO Supermajority Decision such that the funds of any Core DAO may be
frozen or returned to the GovDAO Treasury at any time with a GovDAO Majority
Decision. This does not apply to any funds in other treasuries defined by the
Constitution--not all treasuries are managed by GovDAO decisions except that
they may be managed by Constitutional Amendments.

No spending may occur unless an accountability framework is first defined by a
Constitutional Amendment pre-approved by the Oversight Body such that all
spending is reviewed by an independent committees or entities before and after
the spending decision; all spending is preceded by the on-chain approval by the
relevant governance body of a transparent spending proposal; and all spending
is limited to prevent the abuse of funds and conflicts of interests.

All spending for software must be for open-source software.

GovDAO members and Oversight Body members who have a conflict of interest with
any spending proposals must declare their conflict of interest and abstain from
voting on the spending proposal. Violations as determined by the Oversight Body
will result in punitive measures as determined by the Oversight Body including
suspension or demotion of membership or the return of funds.

## Oversight Body

The Oversight Body is initially NT,LLC as represented by a multisig account
approved by NT,LLC. NT,LLC multisig signers must have an agreement with NT,LLC
and only sign transactions explicitly pre-approved by NT,LLC. In the case of a
breach of this requirement its signature shall have no effect or its effect
must be reversed promptly by GovDAO before any other decision.

Within 2 years after launch NT,LLC must delegate its Oversight Body role to a
DAO on Gno.land implemented in the spirit of Atom.One's CommonDAO Spec, or by
by declaring its members of present or former T1 or T2 members in accordance
with the Oversight DAO Diversity Rule.

If after 2 years after launch NT,LLC does not determine the Oversight Body DAO
or its members, GovDAO shall remind NT,LLC to determine the Oversight Body DAO
members. If after a month such members are not decided by NT,LLC, the Oversight
Body DAO shall be determined by a Supermajority Decision of GovDAO, and no
further decisions may be made by GovDAO until an Oversight Body DAO is
determined that satisfies the requirements.

The Oversight Body DAO Requirements are as follows:

 * There shall be 3 T1 members in the Oversight DAO. 
 * Two or more members shall not reside in the same country at the time of appointment.
 * Two or more members shall not be citizens of the same country or union.

The Oversight Body DAO members may be replaced by GovDAO at any time with a
Supermajority Decision.

## GnotDAO

After 3 years after the launch of Gno.land GovDAO must have implemented
or chosen an implementation of GnotDAO where:

 * GnotDAO is writtenin Gno.
 * GnotDAO runs on a non-upgradeable (immutable) or system realm.
 * Free trading $GNOT may be bonded by users.
 * Users may use bonded $GNOT to vote on any number of valid proposals.
 * Proposals are made from a list of templates as defined by this Constitution
   and are pre-approved by at least two existing GovDAO members.
 * Proposals run with the same rules and duration as Atom.One proposals
   including any self-adjusting deposit limits to reduce spam.
 * At the end of all proposal voting periods of proposals voted on by bonded
   users the $GNOT may be returned to the user immediately.

After the initial establishment of an Oversight Body DAO any member of the DAO
committee may be replaced by the following:

 * a Supermajority Decision of GovDAO, or
 * a GnotDAO proposal by supermajority decision with post-approval by a Majority Decision of GovDAO.

Any replacement or election of the Oversight Body DAO members must meet the
Oversight Body DAO Requirements or improve its condition, and furthermore:

 * An election by GnotDAO must be a US citizen resident and a lawyer in good
   standing.
 * An election by GnotDAO may only happen at most once a year.
 * An election by GnotDAO must replace an existing Oversight DAO member
   previously elected by GnotDAO if one already exists.
 * An election by GnotDAO must be of a member that meets the GovDAO Membership
   General Criteria and must declare all conflicts of interests.
 * An election by GnotDAO may be rejected by GovDAO members for any reason by
   denying the post-approval required of a Majority Decision of GovDAO.

The Oversight Body DAO may not self-mutate.

With the exception of any Oversight Body DAO election proposals, the Oversight
Body or its subDAOs shall have the authority to (and must block) any decision
by GovDAO or Core DAOs if such decisions are determined to be:

 * in violation of the spirit of the Constitution
 * in violation of any laws passed by GovDAO
 * in violation of any bylaws or mandates of violates any relevant bylaws or mandates of any Core DAOs
 * in violation of proper prioritization of funding requirements with priority
   given to those listed in this Constitution so as to ensure the completion of
   those items.

The Oversight Body does NOT have the sole authority to transfer, freeze, or
spend any funds.

Neither GovDAO nor the Oversight Body have the authority to transfer, freeze,
or spend any funds already assigned to a treasury defined in the Constitution
except by a Constitutional Amendment.

No Constitutional Amendments may inflate the $GNOT supply above what is already
declared even with Constitutional Amendments.

## Role and Rights of NT,LLC

NT,LLC retains the excusive rights to word mark and brand "Gno", and "\*Gno\*".
NT,LLC grants GovDAO the right to use the "Gno.land" brand, but NT,LLC is
responsible for managing all domain and subdomains that include the "Gno" word
mark unless otherwise delegated as decided by NT,LLC. In the case of the
"gno.land" domain and subdomains NT,LLC must ensure that it points to the
Gno.land chain, unless the Gno.land identity ceases to exist as defined in this
Constitution.

For a period of 4 years after launch NT,LLC has the sole authority to determine
reasonable guidelines for the content of the main entry pages ("gno.land" and
"www.gno.land"). Thereafter the content of these main entry pages are
determined by this Constitution and applicable laws and overseen by the
Oversight Body. All Qualified Forks of Gno.land shall be listed on these main
pages at all times.

Gno.land and its Qualified Forks may only use the "Gno" word mark and brand and
the "Gno.land" identity and domain and "gnoland\*" chain ID for as long as Gno
is the only contracting language on the chain as determined solely by NT,LLC;
and for as long as GnoWeb is used to render the contents; or unless allowed by
NT,LLC approval on chain by cryptographic signatures by its multisig. Only
NT,LLC may determine the identity of the Gno language for any version and may
authorize releases of the GnovM and GnoWeb. The implementation of Gno.land may
not require the GnoVM if after GnoVM is finalized the alternative
implementation is identical with the GnoVM as determined by NT,LLC and each
release of the alternative implementation is approved by NT,LLC on chain by
cryptographic signatures by its multisig.

Nothing prohibites Gno.land from changing its identity and adopting a different
name that does not include the "Gno" wordmark and brand. GovDAO may choose to
do so by a Constitutional Amendment. However in this case NT,LLC reserves the
right to determine a fork of the Gno.land chain with any modifications to the
constitution and with a new governance body. Such a fork shall be considered a
Qualified Fork even if its governance member set is completely independent of
the orignal GovDAO members.

No Constitutional Amendment shall be valid that restricts the rights of NT,LLC
or reduces its powers or authority as declared in this Constitution nor alter
the US Preference Option without the express permission of NT,LLC. This
includes any modifications to the structure and voting rules of GovDAO.

For a period of 10 years after launch any modifications to the structure and
voting rules of GovDAO shall require the pre-approval of NT,LLC.

All GovDAO members agree to help enforce these rights of NT,LLC.

## Amendments to the Constitution 

All amendments (modifications) to this Constitution incuding changes to any
definitions shall require a Constitutional Amendment pre-approved by the
Oversight Body and passed by a Constitutional Majority Decision of GovDAO
following all the rules of this Constitution.

Amendments to this Constitution must belong to one of severeal categories:

 1. Rewording of portions of the Constitution for clarity or refinement while
    maintaining the structure of the Constitution with no other additions or
    deletions.
 2. Restructuring to move portions of the Constitution for legibility without
    any changes to wording except for section headings or titles.
 3. Additions and deletions of portions of the Constitution without any other
    restructuring or rewordings.

Furthermore the Constitution shall be one single markdown file and the latest
constitution shall be present in the repository under docs/CONSTITUTION.md.

Each amendment shall be composed of up to three diff patches, one for each of
the three categories in the order as declared above.

## User Rights and Limitations

User rights, protections, and limitations will be determined by future
Constitutional Amendments provided that freedom of speech will be protected
with the following exceptions:

Pornographic material must be defined by a constitutional amendment but in all
cases shall include: 

 * photographs or animated or AI rendered depictions of humans or animals or
   humanoid figures engaged in sexual activity or depicted in an arousing way
   including games or movies that include such material; except of classic art
   hand-painted prior to the year 1985 or classic sculptures hand-made prior to
   the year 1985.

 * photographs or animated or AI rendered depictuions of naked minors under the
   age of 18 or humanoid figures that could reasonably construed to be a minor
   including games or movies that include such material; except of classic
   sculptures hand-made prior to the year 1985.

Gore material must be defined by a constitutional amendment but shall only
include realistic depictions of gore whether photographic, video, animated, or
AI rendered, and whether embedded in a movie or in a game; and shall exclude
material that is submitted for educational purposes especially those related to
political conflicts or abuses of power.

Harmful harmful (material harmful for users) is defined as material that is
determined to result in hacks or exploits of user devices unbeknownst to the
user. This excludes material that is submitted for educational purposes and
includes relevant disclaimers to make clear that the material is harmful and is
unlikely to result in accidental hacks or exploits of user devices.

Pornographic material, gore material, and harmful harmful as determined by
GovDAO or as determined by NT,LLC may be filtered by GnoWeb or the blockchain
node in API responses in such a way that anyone who runs GnoWeb or the
blockchain node can also replicate the filter without pay or conditions and
without relying on any external service. This shall not include hatespeech or
misinformation: neither GovDAO nor the Oversight Body nor any Core DAOs may
determine what is hatespeech or misinformation. This is to ensure freedom of
speech.

Blockchain nodes that have the full transaction history will always be able to
sync to the latest state hash from genesis using the latest released node
software and possibly also including previous node software releases. This is
to ensure the integrity of the chain. If validators and nodes are not allowed
to store any offending transaction data they may prune past transactions as
long as they maintain a recent qualfiied snapshot as determined by this
Constitution or GovDAO and all transactions since the snapshot such that the
latest blockchain state may be derived from the snapshot.

Only pornographic material, gore material, and harmful material may be purged
from the blockchain state by marking code packages or state objects (including
entire realms) as such with a State Purge Transaction. Transactions that depend
on state purged by State Purge Transactions shall fail with a special
transaction response code. To preserve integrity of the chain the Merkle-tree
root hash shall be derivable as if the data was there even after purge by State
Purge Transactions.

State Purge Transactions must be constructed by a fully deterministic and
accountable procedure made available to anyone to run freely without depending
on any external services or APIs. State Purge Transactions shall be signable by
parties determined by this Constitution.

The procedure for the submission of State Purge Transactions may not
necessarily be part of the usual logic of the node software as it may require
more resources than should be made to all nodes and may involve prioritization
based on off-chain tips such as by flagging which may not require any gas fees.
However the node software must be made such that node operators may optionally
automatically check the integrity of state purge transactions and detect
false-positives (due to non-determinism bugs or due to the wrongful signing of
State Purge Transactions). False Positive Purge Detection Transactions may be
signed and submitted by active validators as well as any GovDAO delegated
authority; but whether such transactions have any immediate effect or whether
node operators must check the integrity of the transaction shall be determined
by future Constitutional Amendments. False Positive Detection Transactions that
are shown to be valid must be addressed by GovDAO.

All AI models used for the purpose of automated moderation shall be registered
on chain by the hash of its bytes and shall be static and not automatically
trained and updated with new transactions such that anyone can check the
integrity of State Purge Transactions easily with the exct AI model registered
on chain at the time of purge. AI models may be replaced with newer models or
be trained with blockchain data and periodically be registered with the chain.
All AI models used for the purpose of automated moderation shall be made
available for anyone on demand with any reasonable fees for the transmission of
its bytes.

Code packages and realms that are primarily designed or used to bypass these
checks may also be frozen with State Freeze Transactions or purged from the
blockchain state with State Purge Transactions either signed by a manual
process defined by this Constitution.

Any $GNOT byte deposit tokens of state purged from the blockchain state with
State Purge Transactions (after a proving period of 6 months without being
marked on-chain as a false-positive purge by any False Positive Purge Detection
Transaction nor determined to be a false-positive by this Constitution or the
Oversight Body) may only be used for the purpose of developing or acquiring
self-contained open-source AI models and tools for automated or semi-automatic
purging of state; or be used to pay a moderation team as determined by this
Constitution. The confiscation of deposited $GNOT for such purposes shall
require a special Purge Deposit Confiscation Transaction signed by any parties
as determined by this Constitution, and shall fail if the purge occured less
than 6 months ago or if the state was marked since by a False Positive Purge
Detection Transaction.

No more than half of such confiscated $GNOT deposits may be used to fund for
any manual moderation--half or more must be used for the development of
automated systems. The Oversight Body may redirect a third of these funds to
fund for the development or acquisition of alternative moderation systems that
conform to all requirements of this Constitution if it determines that the
existing moderation system is not effective, or has an unacceptable amount of
false-positives; but the adoption of any alternative system shall require a
GovDAO majority decision contingent on there being no more false-positive
purges and no more false-negatives as determined by this Constitution and the
GovDAO voting members. The Oversight Body may also block the submission of any
or all Purge Deposit Confiscation Transactions at any time if it determines
that it is being misued; after which GovDAO is responsible for the continued
funding for moderation from its own treasury until the Oversight Body resumes
the submission Purge Deposit Confiscation Transactions.

User or realm tokens or property that are not $GNOT deposit tokens of purged
state whether native tokens or Gno application derived may not be confiscated
by any means, though they may be frozen by a procedure defined by future
Constitutional Amendments.

User or realm tokens or property that are determined to be derived from any
hacks, thefts, or exploits and are not determined to be owned by users
unrelated to any hacks, thefts, or exploits may be frozen by a procedure
defined by future Constitutional Amendments.

Until 6 months after $GNOT is transferrable and until any future Constitutional
Amendments to the contrary GovDAO by Supermajority Decision with the
pre-approval by the Oversight Body may do anything deemed necessary including
chain upgrades with transaction filtering to reverse the damage as best judged
by GovDAO and the Oversight Body in the spirit of this Constitution.

Future Constitutional Amendments are expected to define conventions and
protocols for realm and library logic to support the freezing of user accounts
or tokens or property in such a way to minimize impact for unrelated parties;
such as by the standardization of "wrapper contracts" that manage the
throttling of deposit to and withdraws from any exchange logic.

## Software Audits

GovDAO is responsible for ensuring that GnoWeb and Gno.land track audit reports
of all code uploaded to Gno as well as the code of the blockchain and VM
itself.

All auditing entity must be qualified; they must be vetted and whitelisted by
GovDAO Supermajority Decision on a Gno.land realm managed directly by GovDAO
(the Auditor and Audits Registry). The auditing entity must be vetted with
sufficient due dilligence and already have a good reputation and track record.
This is to prevent malicious audit reports from confusing and exploiting users.
The auditing reports by such Qualified Auditors are called Qualified Audit
Reports.

Each auditing entity must be registered with general information such as the
jusisdiction of the auditing entity and the public identity of its chief
executives; and significant owners (more than 10%) of the auditing entity must
be disclosed to GovDAO T1 members. The auditing entity may not be a DAO nor be
fully automated based on AI. This is to ensure that there is a clear legal
jurisdiction and accountability in case of abuse by any auditing entity.

The Auditor and Audits Registry must allow any sufficiently $GNOT bonded user
(Bonded Auditor) from submitting newly discovered vulnerabilities missed by any
audit report (whether qualified or not) and with some reasonable deposit per
Bonded Vulnerability Report to be used for assessment of the validity of the
submission. The submission should include a hash of the vulnerability report
but not the vulnerability itself. The procedure to determine these reports will
be defined by future Constitutional Amendments.

The Auditor and Audits Registry realm must track the scope of every audit
report and vulnerability report and record all known relevant information such
that any third party may independently derive a score for each auditing entity.

GnoWeb and all official alternatives (GnoWeb+Alts) must prominently link to any
relevant Qualified Audit Reports and hosted on the Auditor and Audits Registry
realm for all Gno hosted Gno code; but must not convey to the user any
warrantees.

Audit reports from non-qualified auditors may also be displayed on GnoWeb+Alts,
but they must be folded away and require at least one more click or user action
to show along with a prominent warning that those auditors have not been vetted
nor unqualified.

GnoWeb+Alts may display with any Gno code any overall audit status, badge, or
similar (Audit Status) derived from all Qualified Audits and any standing
Bonded Vulnerability Reports; however the Audit Status may not be numeric based
on any line-count or file-count or similar so as to be misleading to the user.
Code coverage of audits may be displayed as a percentage or numeric value(s)
but must be clearly separate from the Audit Status.

In no case may GnoWeb+Alts display the best Audit Status or a "green" or
"approved" status or badge or anything that can be construed as such for any
code unless the entire scope of the code and its dependencies have been
reviewed by at least one Qualified Audit Report and all Qualified Audit Reports
are free of any critical or major vulnerabilities or any issue that may result
in the loss of control or value by any user, and all Qualified Auditors are
still in good standing, and no Bonded Vulnerability Reports have been submitted
either that has not been dismissed explicitly as a false-positive for the code
and all of its dependencies. Furthermore such a status or badge or anything
that can be construed as such may only be displayed if manually reviewed by at
least three GovDAO members two of whom are T1 or T2 members; and they shall
each be responsible for ensuring that all the conditions are satisifed; however
they shall not be responsible for the correctness of each Qualified Audit
Report.

## Realm Upgrading

Realms must be clearly shown in GnoWeb to be upgradeable or not upgradeable
(immutable).

 * An immutable realm may not revert to being upgradeable.

 * No types declared in upgradeable realms may be persisted in immutable realms.

 * No upgradeable realms may be imported by immutable realms.

This creates a two-class system where immutable realms only depend on other
immutable realms. The state of immutable realms may be mutated by any logic as
allowed, but the upgrading of upgreadeable realms can not otherwise affect any
immutable realms.

Library packages declared in /p/ may not be upgraded.

Gnoweb must make clear to users of upgradeable realms that any funds depend on
the trust of the realm controller.

The only exception to the above rules may occur in response to a hack or
exploit. GovDAO with supermajority decision and post-approval by the Oversight
Body may patch any realm package or library package if it is deemed to be in
the best interest of the Gno.land users and $GNOT economy.

TODO Rest to be determined after launch by a Constitutional Amendment.

## Name Registration

Name registration will not intially be implemented, but shall be one of the
priorities to figure out.

There shall be one name registration for Gno.land, though realms may define
their own for any purpose not managed by Gno.land.

The name registration system shall satisfy the following:

 * Only lowercase alphabetic ASCII characters are allowed with underscores not
   as the first character.

 * The length is between 5 and 25 characters.

 * The name shall be canonicalized such that 'l' and 'i' are the same, and the
   canonicalized name shall be used to enforce uniqueness.

 * The canonicalized name shall not start with any of:
   * "gl" (and neither "gi", so as to be distinct from "g1..." addresses)
   * "gno"
   * "cosmos"
   * "atom"
   * "admin" (and neither "admln")
   * "support"
   * and more prefixes as determined by GovDAO.

 * Notwithstanding the above NT,LLC may reserve names that start with "gno".

 * The canonicalized name shall not include (as prefix or suffix or otherwise)
   the canonicalized form of any reserved names as defined by
   https://github.com/handshake-org/hsd/pull/819 without an on-chain
   transparent process approved by GovDAO to respect trademarks with priority
   given to US, then the EU; or only if the inclusion is deemed to be
   sufficiently unique by a reasonable algorithm implemented on chain but still
   preventing collisions in the canonicalized form; or on a case-by-case basis.

 * The registration of the name shall be permanent. Names may be disabled by a
   procedure defined by a Constitutional Amendment that defines the criteria
   for such disablement, but such names cannot be used again by anyone else.

 * The name registration takes learnings from the Handeshake protocol including
   https://github.com/handshake-org/hsd/issues/649

## Atom.One ICS Migration

Gno.land is initially separate from the pre-existing Atom.One. Gno.land
initially has its own transaction fee token ($GNOT instead of $PHOTON).
Gno.land is obligated to migrate to Atom.One when declared ready by GovDAO
simple majority decision.

In the case of a fork of Atom.One the target chain may be changed by a
Constitutional Amendment and pre-approval by the Oversight Body to be one of
the forks of Atom.One. Any GovDAO members who are involved in the creation or
development of such a fork or have any additional relative ownership of staking
tokens or fee tokens of the fork are considered to have a conflict of interest
and must disclose so and abstain from voting. If a quorum cannot be reached the
quorum requirement may be waived by NT,LLC or the Oversight Body.

## Treasury diversification

Each of the above mentioned treasuries may be separately diversified by
supermajority vote of one proposal per treasury by GovDAO in accordance with
the Constitution--but all GovDAO members with any conflicts of tokens (except
$GNOT) above 3 months of salary for a typical senior software developer MUST
abstain from voting. If a quorum cannot be reached due to conflicts and GovDAO
approves by a supermajority vote, NT,LLC may waive the quorum requirement.

Diversification for each treasury must adhere to an Treasury Diversification
Template listed explicitly in this Constitution by a future amendment that
declares the desired target ratio of Approved Treasury Diversification Tokens,
and the following:

 * Besides $GNOT or the top two dominant Bitcoin fork tokens (presently $BTC
   and $BCH but may change in the future) which shall always be approved, all
   other tokens of the Approved Treasury Diversifiction Tokens must be
   stablecoins FULLY backed by collateral.
   * Coins that are FULLY backed by silver via decentralized and publicly
     auditable depositories approved by GovDAO AND NewTendermint,LLC are also
     considered stablecoins.
   * Tether and all stablecoins backed by any amount of Tether are never
     permitted.
   * Tokens that lose their approved status must be replaced (sold) for
     approved tokens as soon as reasonable.
   * These subclauses and the parent clause cannot be removed by any
     constitutional amendment.

 * $PHOTON is not a stablecoin but it may be allowed by up to 20%.

 * $ATONE tokens or any other staking tokens are not allowed.

 * In no case shall the amount of $GNOT sold per treasury per month for
   diversification purposes exceed 1% of $GNOT of that treasury as measured at
   the beginning of each month.

 * In no case shall the total amount of $GNOT sold for all treasuries per month
   for diversification purposes exceed 50% of the average of ($GNOT inflation
   rate, and the past month's transaction fee revenue), with priority given to
   ValTreasury, then to PayTreasury, then to ServicesTreasury.

 * All trades for diversification purposes must be performed by audited and
   approved Gno logic for Gno AMM contracts running on Gno.land. No person or
   council may directly manage the diversification of treasury tokens except
   that GovDAO may delegate an oversight DAO to halt any diversification
   exchange transactions.

## Forking Gno.land

These are the procedures to create a Qualified Fork of Gno.land.

First, a fork proposal is voted on by all GovDAO members. The fork proposal
must be voted YES by more than 1/3 of the total voting power of GovDAO
(abstains do not count) by the end of the voting period of three weeks. There
is no commitment yet for voting YES on this initial fork proposal.

If such a fork proposal reaches 1/3 of voting power, a second fork commitment
proposal must be initiated with the same terms as the original within a week
after the first fork proposal's voting period has ended. Again all GovDAO
members vote on this second proposal. The second fork proposal must also be
voted YES by more than 1/3 of all the total voting power of GovDAO by the end
of the voting period of three weeks. Abstaining is equivalent to voting NO.
Those who vote YES here are committing to join the others who voted YES and
split from those who voted NO or abstained if it also reaches 1/3 of voting
power--members may only serve on one chain.

If the second fork commitment proposal also reaches 1/3 of voting power, two
concurrent proposals must be proposed to determine which fork retains the
original Gno.land identity: one for the current chain, and another for the
proposed fork. Only T1 members may vote on these proposals; and they may vote
YES or NO on one or both proposals.

 * If the current chain wins a 2/3 supermajority of votes of the original T1
   members (but not the proposed fork), everyone who voted YES on the fork
   commitment proposal are ejected from the GovDAO membership set, and the
   proposed fork is called a Qualified Fork.

 * If the proposed fork wins a 2/3 supermajority of votes of the original T1
   members but not the current chain, everyone who did not vote YES on the fork
   commitment proposal are ejected from the GovDAO membership set to join
   their Qualified Fork, and the proposed fork retains the identity of the
   Gno.land chain.

 * If both proposals win a 2/3 supermajority, NewTendermint,LLC decides which
   gets to retain the identity of the chain (and the other is ejected from the
   GovDAO membership set to join their Qualified Fork).

 * If neither fork wins a 2/3 supermajority of votes, NewTendermint,LLC decides
   which fork gets to keep the identity of the chain, and whether the existing
   chain can retain the "Gno.land" name.
   * If neither can keep the "Gno.land" name, the "Gno.land" name is deprecated
     and cannot be used by any chain ever again; and the exiting chain may be
     required to change its chain ID; and the Gno.land domain will be managed
     by NewTendermint,LLC to point to both chains; and both chains shall be
     deemed to be Qualified Forks. Any royalty payments due to the original
     chain may be directed to either fork or split between the two forks as
     determined by NewTendermint,LLC; or even nullified but only if
     NewTendermint,LLC does not have any conflict of interest.

In all cases any new forks that include "Gno" in the name or chain ID must get
approved by NewTendermint,LLC.

All Qualified Forks may copy and use all of the state (including transaction
data) and code as from before the fork.

Non-qualified forks are not hereby probited from copying the state or code of
realms and packages of all forks of Gno.land. That is, everyone who submits
code or transactions to Gno.land or any of its forks are agreeing to allow (and
attesting to having the unencumbered rightsto allow) the code and resulting
blockchain state to be used freely as per the Gno Network GPL license (a
copyleft license fork of the AGPL3.0 but with modifications to allow for
"strong attribution"). However this is necessarily complicated when users
submit Gno code that is not owned by them, so no rights are explicitly granted
here by Gno.land or NewTendermint,LLC.

As with most of the software for the Gno.land stack including the GnoVM (except
for Tendermint2 which is Apache2.0) and derived works, all users of the Gno
code and blockchain state derived from Gno.land must abide by the same strong
attribution terms as for the Gno.land and GnoVM software at the time the code
is submitted--for example a non-qualified fork of Gno.land must give strong
attribution to all of its users to Gno.land or another Qualified Fork as
determined by the policy set forth by NewTendermint,LLC. NewTendermint,LLC may
determine that giving attribution to a Qualified Fork is sufficient (as opposed
to giving attribution to Gno.land), but a non-qualified fork will never benefit
from strong attribution in this way. The policies may be updated by
NewTendermint,LLC from time to time; and the goal is to finalize portions of
the policy such that they may be embedded into the Gno.land Constitution.

These terms shall be made clear in GnoWeb and gnocli for anyone submitting code
or any transaction. Specifically for anyone uploading realm or library packages
to Gno.land using gnocli they must first sign an approved CLA compatible with
the above and include a hash of the CLA with each transaction.

## Software Security

The following are defined as Core Software:

 * Gno.land node.
 * GnoWeb+Alts.
 * Tendermint2.
 * GnoVM.
 * Atom.One IBC and ICS.
 * Official webg wallet software.
 * Primary hardware wallet software.

All Core Software must be developed under the Gno Network GPL License with the
appropriate strong attribution clauses as determined by the owner of the
software, except for Tendermint2 which may remain Apache2.0 or as determined by
NT,LLC. All new Core Software must have its strong attribution clause approved
by GovDAO Supermajority Decision.

All Core Software funded by Gno.land must be fully audited before any release
when funded by Gno.land.

GovDAO must contribute to Atom.One until Atom.One IBC and ICS are complete.

Concurrently GovDAO must contribute to Atom.One such that the validators are
incentivized and over time required to run on their own dedicated machines.

### Software Finalization

All funding for Core Software must be toward the finalization of the software
for the major version number of that software.

 * Funding for Core Software must include an estimate for the future time of
   finalization of the software which may be any number of years into the
   future.

 * Software development that exceeds 7 years of development without
   finalization for the software for the major version number may be halted by
   the Oversight Body unless an extension of up to 2 years is granted by a
   Supermajority Decision of GovDAO. Extensions may be granted at any time, and
   may be to another development party.

 * After the finalization of the software the only funding allowed is for the
   maintenance of that software 

### Private Key Security

At no point shall a GovDAO or Gno.land fund for the development of any software
or team or anything else that allows or encourages the user to enter their
private key or mnemonic on any online computer or mobile network-capable
device. The only exception shall be for existing hardware wallets that allow
for bluetooth, as some of these devices already support bluetooth, as long as
bluetooth can be disabled.

Only 24 word mnemonics may be supported by Core Software; not 12 or 18.

Ephemeral private keys with limited capabilities and default reasonable
limitations on losses (in the case of theft or exploit) may be generated on
online computers or mobile network-capable devices, but no ephemeral keys may
be imported manually by the user nor converted into a mnemonic so as to prevent
any confusion. The term "mnemonic" shall only refer to their master private
key.

GovDAO must maintain a simple guide for users to harden their security including:

 * Discouragement from entering their private key or mnemonic on any online
   computer or mobile internet-capable device, even if the mobile device is
   marketed as being secure for key storage; with education about the ability
   for state actors to add back-doors to such systems.

 * How to obtain an offline computer from a list of approved devices and an
   audited image for the operating system with images committed on chain by
   their hash with bounties for demonstrating a vulnerability in the offline
   computer signing context.

   * Only devices that don't already have any network (including bluetooth)
     capabilities may be included in the list of approved devices.  Software
     disabling of network of bluetooth capabilities is not sufficient.

 * Encouragement to use a 52 deck of cards or 42 rolls of 20-sided dice or the
   equivalent rolls of 6-sided dice to generate custom entropy; with education
   that hardware wallets may be compromised to generate insecure private keys.

Gno.land nor GovDAO nor any entity receiving funding from Gno.land or GovDAO
may not sell any hardware devices except by approved retailers of the
manufactuer.

### Quantum Proof Cryptography

As new post-quantum cryptography, the development of quantum computers, and
advancements in algorithms for breaking elliptic curve cryptography and hash
functions are ongoing; and it is generally agreed that hash functions are more
secure against quantum computers than elliptic curve cryptography, we will
first explore Merkle hash based commitments for user and validator accounts
that can be used to recover after "D-day"; and also prohibit user accounts from
the re-use of public keys (whether elliptic curve based or new quantum proof
algorithm based) so as to protect against future attacks by shielding the
public key from exposure via a hash function.

The following is the roadmap for quantum proofing Gno.land. It may be modified
with a Constitutional Amendment by GovDAO and also the pre-approval of NT,LLC.

   * Within 1 year after launch:
     * Determine a protocol to register each account with two Merkle tree hash
       commits of at least 64 distinct Lamport signature public keys with one
       commit using only the sha256 hash function and the other commit using a
       different hash function family such as sha3 (Quantum Recovery Commit)
       that can be used to securely register/sign-for at least 64 new arbitrary
       (e.g. quantum-proof) public key bytes of any length determined in the
       future (e.g. after "D-day") for each commit.
   * Within 2 years after launch:
     * Deploy the aforementioned Quantum Recovery Commit registration system
       and require its usage for all validators and also new user accounts
       before any new transactions are accepted.
     * Determine node, client, and hardware wallet software changes to prevent
       the re-use of public keys for user transactions without relying on new
       or unproven cryptography. Each account is associated with the next
       deterministic 20-byte or larger address committed by each transaction
       based on the strictly sequential account sequence number.
     * Begin work with reputable cryptographers with recent relavant
       publications to determine the number of bytes (currently 20) needed and
       ideal hash function for account addresses as determined by the best
       theoretical attacks, and assuming the most optimistic/advanced estimates
       and projections for quantum computers developed in secret by nation
       state actors.
   * Within 3 years after launch:
     * Deploy the aforementioned node, client, and hardware wallet software
       changes to prevent the re-use of public keys for user transactions.
   * Within 4 years after launch:
     * Disable the re-use of public keys for all user transactions.
     * Freeze all accounts that do not yet have a Quantum Recovery Commit. This
       is to prevent future mass hacking of accounts by quantum computers which
       can destroy the economy of the chain.
   * Users may continue to use secp256k1 hash addresses as long as those
     addresses are not re-used, the address has sufficient byte length given
     the best known theoretical attacks, and attacks for a given address has
     not yet been provably demonstrated within the unbonding period of
     Atom.One. This is to prevent requiring users to adopt new (e.g. new
     quantum-proof) cryptography algorithms that are yet unproven or may prove
     to be insecure in the future.
   * New quantum-proof public key systems must be strictly additive to the
     usual public key system until proven by industry for at least 5 years.

Validator signing and verification have different requirements than user
accounts (e.g. they can assume larger memory and compute capacity than hardware
wallets; and they must remain secure even after public key exposure for
light-client/IBC-bridges and for double-signing slashing/jailing mechanics) so
require a different path for post-quantum readiness.

   * Within 2 years after launch:
     * Investigate changes to Tendermint2 with NT,LLC to support pluggable
       quantum proof cryptography.
   * Within 3 years after launch:
     * Complete an initial draft modification of Tendermint2 approved by NT,LLC
       to support pluggable quantum proof cryptography.
   * Within 4 years after launch:
     * Complete final release of Tendermint2 by NT,LLC to support pluggable
       quantum proof cryptography.
   * Within 5 years after launch:
     * Require validators to run Tendermint2 with quantum proof cryptography.
       If Atom.One does not allow post-quantum Tendermint2 by this time and
       Gno.land runs on Atom.One ICS, Gno.land must migrate away from Atom.One
       ICS.

No development or endorsement of quantum-proof modifications to Gno.land Core
Software or any consensus engine shall take place without the pre-approval of
NT,LLC (or its delegate) of its design specification. GovDAO may use a
reasonable amount of funds (under the typical salary for a software architect
for a year) to collect proposals for such specifications to be approved by
NT,LLC (or its delegate) before funding for its development. 

### New Cryptographic Primitives or Implementations

Any new cryptographic primities or implementations must require NT,LLC
pre-approval unless NT,LLC designates this role to a Cryptography Committee
also approved by GovDAO with at least 3 T1 or T2 members who are respected and
have well-cited publications in the field of cryptography and at least a
doctorates degree in cryptography or relevant fields in mathematics, and at
least 1 T1 or T2 members who are equally expert in the field of quantum
computing.

### Formal Verification of Go/Gno.

Once the market cap of $GNOT exceeds $5B GovDAO must pass a Constitutional
Amendment to include provisions for the formal verification of Go or Gno
programs with the approval of a reputable researcher with publications about
formal verification using open source tools already used by industry for formal
verification of programs.

If meaningful progress of such cannot be made during any 2 years of funding
another qualfiied researcher must be consulted for development with a new team.

## Hardware Security

### Development of Open Hardware

Once the market cap of $GNOT exceeds $5B GovDAO must pass a Constitutional
Amendment to include provisions for the development of 100% open hardware (and
100% open firmware/software) devices for the following categories in order:

 * A device that sits between an online computer or mobile network-capable
   device and hardware signer such that all communication between the
   aforementioned devices can be inspected (the Hardware Packet Inspector).
   This reduces the chances that a backdoor in the hardware wallet may result
   in theft or loss. If no suitable 100% open hardware CPU chip is available a
   Risk-V architecture based chip may be used instead until one is avaialble.

 * Hardware signer that is separate from but compatible the Hardware Packet
   Inspector. This may use a single chip that is proprietary to hold the
   private key and logic for unlocking based on pin, but it must be a module
   contained within completely open hardware and open firmware/source such that
   a vulnerability or backdoor of the single chip can be limited.

 * Validator hardware signer to preserve the private key of the validator and
   prevent double-signing.

 * CPU chip based on the Risk-V architecture (or related reputable open source
   forks) for use in any of the above devices.

 * A device for the hardening of chips such that keys cannot be extracted
   easily from these chips by forensic analysis.

The open hardware and open software/firmware must be made available primarily
by a copyleft license compatible with the latest Gno Network GPL license or a
Gnu copyleft license with the appropriate attribution terms for hardware such
that the product is free of patent restrictions and all modifications to the
funded source must also be made available with no proprietary modifications
allowed except for personal use, and any derivative works and any manufacture
must give prominent credit to the original designer(s) of the hardware.

The funding may go toward investment in an existing reputable hardware
developer or contract with one but the resulting work must be made available
such that after its release anyone may build upon the product with appropriate
attribution.

All open hardware must be formally verified by standard formal verification
software before release.

No one shall be required to use the these products.

### Forensic Analysis of Common Hardware

Furthermore once the market cap of $GNOT exceeds $5B GovDAO must pass a
Constitutional Amendment to include provisions for funding an in-house team to
accountabily verify the integrity of commonly used hardware products of
Gno.land users, especially the forensic analysis of the chips of those hardware
products. 

 * The devices must be acquired externally by a variety of means so as to
   prevent the prediction by manufacturers or retailers. If a device cannot be
   acquired by the user from a physical store without identification, such a
   device cannot be trusted at all. For example, certain AT&T devices can only
   be purchased by AT&T through the mail, and they are tampered with in transit
   when being delivered to targetted individuals. These devices will be
   published in a blacklist curated by GovDAO or a delegate commitee. These
   devices will not be tested.

 * Users may also provide hardware to be inspected by this team for pay.

 * The analysis must be done with a video and photographic recording so as to
   ensure that the analysis is correct. The video and photographs must be of
   sufficient quality so as to be correlated with each other and the
   photographs of sufficient quality that they can be inspected by the public.

 * The shipment of devices to the forensic analysis team must be done in such a
   way that the origin of the devices can be verified during the recording and
   photography.

All of the procedures, software, and materials used by the forensic analysis
team shall be made public and free such that any team may use them to offer
their own forensic analysis.
