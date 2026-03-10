# GNO.LAND CONSTITUTION

_This Constitution is still being worked on and is expected to be ratified
after Gno.land genesis by a Supermajority Decision of GovDAO._

## Declaration of Intent

We the gnomes of Gno.land, to bring to life a virtual world of structured
information and logic expressed in a multiuser language based system, for
bringing light to darkness by hosting all that will be known, hereby establish
this Constitution for Gno.land.

## Terms

 * AIB,Inc: All in Bits, Inc.

 * Bonded Auditor: A sufficiently $GNOT bonded user who submits newly
   discovered vulnerabilities to the Auditor and Audits Registry. See
   "Software Audits".

 * Bonded Vulnerability Report: A vulnerability report submitted by a Bonded
   Auditor for a vulnerability missed by any audit report. See "Software
   Audits".

 * Constitution: This Gno.land constitution document including modifications by
   Constitutional Amendments. This Constitution is considered independent of
   the Atom.One Constitution although some portions of the Atom.One
   Constitution may be referenced here for clarity or context.

 * Constitutional Amendment: A modification to this living Constitution by a
   Constitutional Majority Decision of GovDAO and any other requirements as
   defined in this Constitution; composed of diffs.

 * Constitution Only: When something is governed by the "Constitution Only" its
   governing rules must be specified within the living Constitution directly;
   not in any Laws, Bylaws, or Mandates.

 * Core DAO: GovDAO, the Oversight Body DAO, or any subDAO created or managed
   by any Core DAO implemented on Gno.land conforming to the Common DAO Spec.

 * Core Software: The minimal set of reference or standard software needed for
   the Gno.land blockchain and its users. See "Software Security".

 * Fully Audited: See "Software Security".

 * GnotDAO: A $GNOT-bonded on-chain voting DAO for broader community
   participation in certain governance decisions. See "GnotDAO".

 * GovDAO DAO: "All GovDAO DAOs" or "GovDAO DAO Set" refers to GovDAO and all
   descendant subDAOs ultimately created or managed by GovDAO. "A GovDAO DAO"
   (or "relevant GovDAO DAOs") refers to one DAO within the GovDAO DAO Set (or
   any number of DAOs within the GovDAO DAO Set) that has a Mandate that
   matches the given scope.

 * Governing Documents: This Gno.land Constitution, its Laws, and any
   applicable Mandates and Bylaws (altogether Governing Documents).

 * ICS: Inter-Chain-Security, or the general service of validating a chain with
   the validator set of another chain, or a subset thereof.

 * Law Amendment: A modification to the Laws by a Supermajority Decision of
   GovDAO. The procedure is similar to Constitutional Amendments. See
   "Gno.land Laws".

 * NT,LLC: NewTendermint, LLC.

 * Open Source IP: Intellectual Property that is made available freely under
   the Gno Network GPL or any liberal license compatible with the Gno Network
   GPL.

 * Oversight Body: Initially NT,LLC as represented by a multisig account, but
   eventually represented by an Oversight DAO on Gno.land.

 * Qualified Auditor: An auditing entity vetted and whitelisted by GovDAO
   Supermajority Decision on the Auditor and Audits Registry. See "Software
   Audits".

 * Qualified Audit Report: An audit report produced by a Qualified Auditor.
   See "Software Audits".

 * Qualified Candidate: A candidate who meets all GovDAO Membership Criteria
   including General Membership Criteria and Diversification Requirements. See
   "GovDAO Membership Criteria".

 * Qualified Fork: A fork of Gno.land created through the procedures defined
   in "Forking Gno.land".

 * Revenue: Transaction fees collected in $GNOT or $PHOTON on Gno.land. See
   "Gno.land Revenue".

 * SDDCAs (Storage Deposit Discount Credit Accounts): Excess $GNOT deposits
   from the reduction in price of $GNOT/byte deposit ($GNOT Storage Deposit
   Price). See "Excess $GNOT Deposits".

 * Seven Mandates of Gno.land: The following immutable mandates:
   1. Protect the user; preserve user data and intent.
   2. Preserve the spirit of the Genesis Constitution.
   3. Safety over everything else, followed by intuitive simplicity.
   4. Factor (keep minimal) Constitutional and Legislative Amendments.
   5. Do not require a name or identity for any core services.
   6. Aim to be perfect to serve one thousand years.
   7. Expand the light unto the multitude and nations.

 * Simple Majority, Supermajority, and Constitutional Majority Decision: See
   "GovDAO Decisions" and "Common DAO Spec".

 * State Freeze Transaction: A blockchain transaction that freezes a realm,
   package, or account, preventing further mutations without purging state.
   See "User Rights and Limitations".

 * State Purge Transaction: A blockchain transaction for deleting state from a
   Gno.land realm or package. See "State Purge Transactions".

 * Treasury: See "Gno.land Treasuries".

## Genesis Allocation

At Gno.land Genesis there will be one billion $GNOT tokens.

 * Airdrop1:               35.0% - from partial Cosmos governance snapshot 3 years ago
 * Airdrop2:               23.1% - from recent AtomOne snapshot prior to launch
 * Ecosystem Treasury:     11.9% - for prior and future Gno.land ecosystem development
 * Investors:               7.0% - reserved for current and future investments
 * NT,LLC:                 23.0% - of which a significant portion is allocated for prior loans

$GNOT will not be transferrable initially except for whitelisted addresses.
Whitelisted addresses include "Ecosystem" and "Investors" funds and any
additional addresses needed for the operation of the chain.

The 7% (qualified, private) investors allocation will be held by NT,LLC in a
segregated account. Proceeds of sales of these tokens will go toward NT,LLC for
past or future development of Gno.land, Gno, Tendermint2, other Core Software,
and ecosystem development.

GovDAO is responsible for distributing the $GNOT of the Ecosystem Treasury
allocation to prior and future Gno.land ecosystem contributors (as well as
those contributing to the Core Software). For more information see "Gno.land
Treasuries".

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
 * Atom.One ICS Consumer Chains (aka ICS shards)
   including Gno.land after migration must accept $PHOTON as the gas-fee token.
   * Each may also accept other gas-fee tokens but ultimately the chain must
     pay $PHOTON to Atom.One for security.
 * Atom.One ICS sovereign (consumer) chains may exit (change validator-sets)
   Atom.One and become self-validating or become hosted by another ICS hub.

Gno.land will launch independently of Atom.One so Gno.land will initially
require $GNOT for transaction fee payment. Before and after Gno.land migrates
to be hosted by Atom.One ICS $GNOT will function as a byte storage deposit
token.

 * Every transaction that ends up increasing the amount of persistent state
   will require a bond deposit of $GNOT.
 * Every transaction that ends up freeing up persistent state space will
   receive a refund of $GNOT.
 * One billion $GNOT corresponds to 10TB of persistent state space.
 * The $GNOT inflation schedule will never change, thus the total created $GNOT
   will never exceed 1.333~ billion $GNOT.
 * The $GNOT Storage Deposit Price (per byte) will never increase.
 * The $GNOT Storage Deposit Price may decrease by at most 10% a year.
 * The $GNOT Storage Deposit Price may not decrease if the total existing
   storage exceeds the size of a typical consumer internal hard-disk or
   solid-state drive. See also "Gno.land Storage Capacity".

Gno.land is obligated to migrate to be hosted/secured by Atom.One ICS when it
is deemed ready according to this Constitution by GovDAO by Supermajority
Decision.

After migration to Atom.One ICS hosting Gno.land should pay the Atom.One chain
in $PHOTONs underneath the hood as the Atom.One constitution requires; an AMM
exchange module should exist on the Gno.land shard/instance to facilitate the
internal exchange of collected $GNOT to $PHOTON needed to pay Atom.One for its
ICS services.

Once Gno.land migrates over to Atom.One after the Gno.land <> Atom.One IBC
connection is complete and Atom.One ICS MVP is implemented $ATONE will be the
staking and governance token on Atom.One (but with no voting rights for
Gno.land itself), and $PHOTON will be
primarily the CPU-time gas token paid to Atom.One (which in turn pays for all
that is necessary to secure the chain via ICS validation), and $GNOT the
dedicated byte-storage deposit token on Gno.land. Gno.land will become a key
ICS consumer chain on Atom.One especially in the beginning even as Atom.One is
free to offer its ICS services to other applications unrelated to Gno.land and
$GNOT, or even those forked of Gno.land and the GnoVM in the future. 

## $GNOT (Deflationary) Inflation

From the date of launch every year 33.33*(0.9^Y) million $GNOT will be inflated
continuously (where Y is the year from launch starting from 0). This is 3.333%
of the Gno.land Genesis $GNOT supply in the first year, decaying by 10% each
subsequent year. Any inflationary $GNOT not
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

 * One third to the Core Treasury.
 * One third to the Ecosystem Treasury.
 * One third to the Reserve Treasury.

## Gno.land Treasuries

The Gno.land Treasuries are as follows:

 * Core Treasury
 * Ecosystem Treasury
 * Validator Services Treasury (ValTreasury for short)
 * GovDAO Pay Treasury (PayTreasury for short)
 * Security Treasury
 * Reserve Treasury

A Treasury is defined as account or DAO that can hold funds where its type,
structure, governance (if any), and purpose are specified in the Constitution
Only. If a Gno.land Treasury is defined without governance or conditions its
funds can only be used with a Constitutional Amendment. _The Reserve Treasury
in particular is reserved for future uses._

A Treasury with a given purpose cannot be used for any other purpose even if
there are excess funds in the Treasury account (or DAO). If GovDAO passes a
Constitutional Amendment that violates this principle (such as by the deletion
of this term) it is the responsibility of the Oversight Body to reject such
proposals. Constitutional Amendments for clarification of language in the
spirit of the terms prior to any such amendments are allowed for cases where
the original language is ambiguous. _The exception to this rule is as per the
disclaimer at the top of this draft Constitution--the set of treasuries and
their purposes, or even this term, may change under certain conditions until
$GNOT is made transferrable._

No spending proposal should be voted on unless an accountability framework is
first defined by a Constitutional Amendment pre-approved by the Oversight Body
such that all spending is reviewed by independent committees or entities
before and after the spending decision.

No subDAO of a top-level Core DAO may make funding decisions from the Core
Treasury directly unless otherwise specified in the Constitution.

Funds from Gno.land Treasuries may be transferred to another Core subDAO
ultimately managed by GovDAO such that the funds of any Core subDAO may be
frozen or returned to the origin Treasury at any time with a Simple Majority
Decision of the responsible Core DAO.

Funds that are transferred or assigned to any of these treasuries but not yet
used (e.g. by a decision to fund an entity or a DAO) may not be clawed back or
transferred to any other DAO or treasury nor burned without a Constitutional
Amendment.

All spending for software or other IP must be for Open Source IP.

### Core Treasury

These funds are reserved for the development of Core Software and Essential
Services.

All funding decisions from the Core Treasury require a Supermajority Decision
by GovDAO.

The Core Software is the minimal set of reference or standard software needed
for the Gno.land blockchain and its users and includes (and must prioritize):

 * Gno.land node
 * GnoWeb+Alts
 * Tendermint2
 * GnoVM
 * Go
 * Atom.One IBC and ICS
 * Official standard reference browser extension wallet
 * Official standard hardware wallet software
 * All other software mentioned in the Constitution Only

Essential Services are the set of services rendered by machine or man that are
deemed essential for the operation of Gno.land and for users to securely
interact with Gno.land and includes (and must prioritize):

 * Blockchain explorers
 * Blockchain indexers
 * User (customer) support
 * Official community activities

All Core Software must be released under the Gno Network GPL License with the
appropriate strong attribution clauses as determined by the owner of the
software, except for pre-existing software that is already released with a
compatible open source license.

Funding for Tendermint2, the GnoVM, and the gno cli is primarily the
responsibility of NT,LLC. Funding for Atom.One IBC and ICS is primarily the
responsibility of the Atom.One chain. That said, GovDAO should contribute to
these projects (and any other projects of the Core Software owned by external
entities) within reason as needed to complete the objectives mentioned in the
relevant Governing Documents; and thereafter for as long as the software is
deemed Core Software GovDAO must ensure the continuous and timely monitoring of
future improvements and updates by reasonable funding if necessary from the
Core Treasury to keep Gno.land and its users safe.

Funding from the Core Treasury for Essential Services software may only go
toward Open Source IP that is fully audited.

Funding from the Core Treasury may also go toward validation services on an
emergency basis if the Validator Services Treasury is insufficient.

GovDAO is required to use a legal entity such as NT,LLC or any new entities
(Proxy Entity) to keep ownership of IP on behalf of the Gno.land chain. New
Proxy Entities must be proven to be subservient to GovDAO in accordance with
the Constitution such that no transfers of IP nor any changes to license terms
may occur without the prior approval of the relevant GovDAO DAO and Oversight
Body. Each IP repo must have a top-level Markdown file describing the proxy
relationship (except for Tendermint2 and the GnoVM)

Refactoring of projects out of the genesis monorepo should preserve the
history of all prior commits relevant to all the files of each project (via the
git "--follow" flag). The expectation is for the Proxy Entity and GovDAO to
come to agreement about serving as the Proxy Entity for some or most of the
repos already hosted under the organization.

### Ecosystem Treasury

These funds are reserved for rewarding the creators of External Contributions
defined as contributions of Open Source IP mentioned in any of the Governing
Documents and Gno software not already paid for by other means. Recognizing
and rewarding contributors is essential to the long-term health of Gno.land
and its ecosystem.

At most 1/4 of the GovDAO 11.9% genesis allocation may be distributed to prior
contributors by GovDAO Supermajority Decision. Present GovDAO members are not
eligible for any allocation from the Ecosystem Treasury genesis allocation.

All funding decisions from the Ecosystem Treasury require a Supermajority
Decision by GovDAO. 

All funding decisions from the Ecosystem Treasury must be based on well defined
heuristics as defined in the Constitution Only, or based on quality as
assessed by the review of council members of the relevant GovDAO DAO. Naturally
council members must abstain from voting on matters with any conflict of
interests, but the Oversight Body is specifically responsible for (and has the
authority for) ensuring that council members do not collude to reward each
other unfairly in any given proposal nor over multiple proposals.

All funding from the Ecosystem Treasury must go to contributors whose real
human identity is known and recorded in accordance with the Constitution or
relevant Laws of Gno.land.

### Validator Services Treasury

The Validator Services Treasury may only be used to pay for validation
services.

No Core DAO may vote for funding decisions from the Validator Services Treasury
except immediately after genesis until automated payment mechanisms are
implemented.

See also "Gno.land Pre-Migration Validators" and "Core Treasury".

### GovDAO Pay Treasury

These funds are reserved for paying GovDAO members who are actively
contributing to Gno.land. Participation in the governance of GovDAO itself is
not considered active contribution.

All funding decisions from the GovDAO Pay Treasury require a Supermajority
Decision by GovDAO (naturally in accordance with the Constitution).

No payment may go to GovDAO members for any contributions older than 3 months.
That is, retroactive compensation is prohibited.

T1 and T2 members who are actively contributing full-time must get paid
equally. T3 members must get paid strictly less by comparison.

T3 members may not get paid unless there exists in the GovDAO Pay Treasury
sufficient funds to pay all T1 and T2 members for 7 years after launch taking
into account the latest estimated projections including new T1 and T2 members.
That said, T3 members may get paid by other means outside of the GovDAO Pay
Treasury.

Payment for any member from the GovDAO Pay Treasury may not exceed the 90th
percentile of senior software architect roles in the second highest paid city
globally.

Members already employed by another company will not receive any compensation
unless they disclose their compensation to GovDAO T1 members via a process
defined in the Governing Documents; afterwards they may be compensated up to
50% of the usual limit to top up their net payment to the usual limit.

If there are not enough funds in the GovDAO Pay Treasury to pay all GovDAO
members for the next quarter, payment must be reduced equally by up to 10% of
the usual amount for the next quarter. If this reduction is not sufficient then
all T3 members must lose their funding (if any). If this reduction is still not
sufficient then T2 members must lose their funding for the next quarter based
on seniority. If this reduction is still not sufficient then T1 members must
lose their funding for the next quarter based on seniority.

Members who lose their pay due to inadequate funds of the GovDAO Pay Treasury
do not automatically lose their membership; and must not be required to work
more than 25% of full-time to maintain any status regarding activity.

### Security Treasury

The Security Treasury may only be used to fund the auditing of code deployed
to Gno.land by Qualified Auditors, to assess and reward Bonded Auditors who
submit valid Bonded Vulnerability Reports, and to recompense victims of
exploits and fraud.

All funding decisions from the Security Treasury require a Supermajority
Decision by GovDAO (naturally with members with any conflict of interests
abstaining).

There is no obligation to compensate anyone; and furthermore no Core DAO may
make guarantees of recompensation to anyone.

Recompensation decisions must be preceded by a thorough analysis of the problem
and tasking of a task-force to recover any ill-gotten gains and at least two
weeks for all relevant parties to review the analysis and task-force for
approval.

GovDAO by Supermajority Decision may choose to burn tokens from the Security
Treasury at a rate not exceeding 10% a year.

### Reserve Treasury

No funding decisions are allowed from the Reserve Treasury without a GovDAO
Constitutional Majority Decision.

## Gno.land Revenue

Transaction fees collected in $GNOT or $PHOTON on Gno.land are called Revenue.
Revenue is distributed in the following priority order. Each treasury receives
its share from whatever remains after higher-priority treasuries have been
funded.

 1. **ValTreasury** receives a runway-based share of Revenue:
    * 50% if less than 1 year of runway, or
    * 25% if less than 2 years of runway, or
    * 10% if less than 3 years of runway, or
    *  5% if more than 3 years of runway.

 2. **PayTreasury** receives a runway-based share of the remainder:
    * 50% if less than 1 year of runway, or
    * 25% if less than 2 years of runway, or
    * 10% if less than 3 years of runway, or
    *  5% if more than 3 years of runway.

 3. The remainder is split equally:
    * **Core Treasury**: 25%
    * **Security Treasury**: 25%
    * **Ecosystem Treasury**: 25%
    * **Reserve Treasury**: 25%

ValTreasury and PayTreasury have priority because the blockchain cannot
function without validators or governance. The PayTreasury runway must take
into account future projected growth of the paid T1 and T2 members. The
remaining revenue is split equally among the four other treasuries. The Core
Treasury also receives one third of inflation; the equal split ensures that
Security, Ecosystem, and Reserve receive meaningful perpetual funding
independent of decaying inflation.

## Excess $GNOT Deposits

When the $GNOT storage deposit rate per byte is decreased this results in
excess $GNOT deposits per realm. This is not considered part of Revenue.

When the $GNOT storage deposit rate decreases (not automatically by the $GNOT
burn mechanism described below, but by the decision of GovDAO to lower the
rate) 25% of the excess $GNOT goes into the Security Treasury and the remainder
goes to each segregated virtual account per realm ($GNOT of the realm's SDDCA) to
be used only for future discounts on transactions for that specific realm.
These SDDCA $GNOT tokens may be transferred to other SDDCAs by rules defined in
the Governing Documents but their $GNOT may never be withdrawn even upon
freeing storage, and transfers must be initiated by the present authority of
the realm or containing organization. 

DDCA $GNOTs may be burned automatically by a rate set by an amendment to the
Gno.land Constitution not to exceed 10% a year. This is to prevent stagnant
$GNOT from limiting the allocated storage capacity of Gno.land and thereby
reducing its utility. $GNOT burned in this way will also reduce the $GNOT
storage deposit rate automatically but not trigger any of the mechanisms
described here.

See also "Security Treasury".

## Gno.land Pre-Migration Validators 

Until Atom.One ICS is ready, Gno.land's validator selection mechanism is
determined by GovDAO Supermajority Decision.

Priority should be given to Atom.One validators and Gno.land core developers.

The number of validators prior to migration may not exceed 50.

No validators may operate on any cloud hosting provider, but must run their own
hardware where they have physical access to the server at all times (such as in
a co-location provider with 24/hour access).

Atom.One ICS may not be deemed suitable unless the Gno.land chain remains whole
(not part of any "mega-block" system where the consensus engine process is
shared with other applications) and Gno.land may migrate away from Atom.One by
on-chain transactions.

GovDAO by Supermajority Decision may migrate Gno.land away from Atom.One ICS
or fund the development of a fork of Atom.One if Atom.One fails to meet
Gno.land's security or decentralization requirements. Such requirements include
but are not limited to validators running their own hardware and reasonable
limits on validator power concentration.

Pre-migration and post-migration validators are paid from the Validator
Services Treasury. See also "Validator Services Treasury".

## Atom.One ICS Migration

In the case of a fork of Atom.One the migration target chain may be changed by
a Constitutional Amendment. Any GovDAO members who are involved in the creation
or development of such a fork or have any additional relative ownership of
staking tokens or fee tokens of the fork are considered to have a conflict of
interest and must disclose so and abstain from voting. If a quorum cannot be
reached due to this the quorum requirement may be waived by the Oversight Body.

## GovDAO

GovDAO is the primary governing body of Gno.land with limitations by the
Oversight Body and other terms and conditions of this Constitution.

### GovDAO Structure

There are three tiers: T1, T2, and T3.

T1 is the highest tier, T3 the lowest.

 * T1: self-selecting "core" with supermajority vote from T1.
 * T2: selected by GovDAO w/ T3 abstaining, with a Simple Majority Decision.
 * T3: permissionless invitation from T1 and T2 according to a delegation
   mechanism.

The maximum age for any member is 70; after is automatically withdrawn.

T1, T2, and T3 membership can only be withdrawn by a GovDAO Supermajority
Decision WITH CAUSE and Oversight Body pre-approval. Such members are
considered "ejected". Ejected members are no longer eligible for membership in
GovDAO.

Members may resign at any time with a signed transaction. Resigned members may
be reinstated with a Supermajority Decision of GovDAO.

T1 members may not be actively contributing to the project but they must still
be active in voting or their membership may be suspended if they miss 3
GovDAO proposal votes in a row until they return to activity with a simple
declaration. The activity status of T1 or their suspension may not be deemed
cause for removal from T1. In this way T1 is like "tenure", and generally the
T1 members are expected to have made significant contributions to the project.

### GovDAO Voting Power

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

#### T1 Membership Size

The target T1 membership size is a minimum 70 members in 7 years.

If the target minimum is not reached, at most 2 members may be elected by the
usual process of election every quarter.

If the target minimum is not reached AND there are Qualified Candidates AND two
quarters have passed with no new T1 members elected, 1 additional Qualified
Candidate may be elected by GovDAO with T1 abstaining in the following quarter.

If the target minimum is not reached AND there are Qualified Candidates AND
four quarters have passed with no new T1 members elected, 1 additional
Qualified Candidate may be elected by GnotDAO.

#### T2 Membership Size

The T2 membership size target minimum is floor(size(T1)/4). The T2 membership
size target maximum is 2 x size(T1).

While size(T2) can be greater than 2 x size(T1) or less than 2 x size(T1), no
more members may be added to T2 if size(T2) is greater than 2 x size(T1).

#### T3 Membership Size

T3 membership is decided automatically based on a delegation system where:

 * T1 members get 3 invitation points.
 * T2 members get 2 invitation points.
 * T3 members get 1 invitation point.

At least 2 invitation points from at least 2 members must be delegated for T3
membership. Invitation points are whole numbers; they are not divisible.

The T3 membership is determined automatically based on the current GovDAO
membership and the latest delegations. Delegations may be changed at any time;
thus a T3 member may lose their membership immediately as a consequence of
undelegation, and this may cause another T3 member to lose their membership
concurrently.

#### Payment to GovDAO Members

See "GovDAO Pay Treasury".

### GovDAO Decisions

A Simple Majority Decision of GovDAO requires more than 1/2 of voting power of
T1, T2, and T3 members according to the scoring mechanism defined in this
Constitution.

A Supermajority Decision of GovDAO requires more than 2/3 of voting power of
T1, T2, and T3 members according to the scoring mechanism defined in this
Constitution. Most decisions require a Supermajority Decision.

A Constitutional Majority Decision of GovDAO requires more than 9/10 of voting
power of T1, T2, and T3 members according to the scoring mechanism defined in
this Constitution.

Some decisions may require additional conditions to be satisfied:

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
   all relevant conditions are not satisfied the proposal must be rejected by
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

The general requirements for T1, T2, and T3 membership are as follows:

 * All members must be publicly identifiable persons and declare their country
   of residence and citizenship and provide timely updates when it changes.

 * T1 members must meet T1,T2,T3 criteria.

 * T2 members must meet T2,T3 criteria.

 * T3 members must meet T3 criteria.

 * T1 criteria includes expertise in categories, significant contributions, and
   demonstration of value alignment, and when active must make public
   appearances promoting Gno.land or publications related to their
   contributions to Gno.land.
   
 * T2 criteria includes expertise in categories and continuing contributions
   including public appearances promoting Gno.land or publications related to
   their contributions to Gno.land.

 * T3 criteria includes significant experience in categories and continuing
   contributions.

 * T1 and T2 members are added via individual proposals, each with a markdown
   resume/portfolio application document.

 * Any members who have been proven to have lied or not disclosed material
   information must be demoted or ejected with a Simple Majority Decision with
   a pre-approval from the Oversight Body contingent on proof and evaluation.

 * Ejected members are never again eligible for membership in T1, T2, or T3.

#### Diversification Requirements

The Diversification Requirements exists to prevent any single country or union
from dominating GovDAO thereby ensuring a globally decentralized governance
body.

 * No new candidates for T1 (or T2) may be proposed if their country or union
   of citizenship represents more than one third of T1 (or T2).

 * No new candidates for T1 (or T2) may be proposed if their country or union
   of residence represents more than one third of T1 (or T2).

#### Conflict of Interest Policy

All Core DAO members must annually sign and circulate to all T1 members a
conflict of interest disclosure document.

All Core DAO members who have any conflict of interests with any proposals must
declare their conflict of interest and abstain from voting on the spending
proposal.

See also "General Membership Criteria" for additional terms related to conflict
of interests.

Violations of the Conflict of Interest Policy as determined by the Oversight
Body must result in punitive measures as determined by the Governing Documents
including ejection, suspension, or demotion of membership, or the clawback of
funds. That said, any punitive measure by the Oversight Body may be vetoed or
overruled by a GovDAO Supermajority Decision at any time.

## Oversight Body

The Oversight Body is responsible for ensuring that all proposals (except
proposals to replace Oversight Body DAO members) adhere to the Governing
Documents with priority given to the Constitution in case of any conflicts.

The Oversight Body is initially NT,LLC as represented by a multisig account
approved by NT,LLC. NT,LLC multisig signers must have an agreement with NT,LLC
and only sign transactions explicitly pre-approved by NT,LLC. In the case of a
breach of this requirement its signature may have no effect or its effect
must be reversed promptly by GovDAO before any other decision.

Within 2 years after launch NT,LLC must delegate its Oversight Body role to a
DAO (the Oversight Body DAO) on Gno.land, or by declaring its members in
accordance with all requirements.

If after 2 years after launch NT,LLC does not determine the Oversight Body DAO
or its members, GovDAO must remind NT,LLC to determine the Oversight Body DAO
members. If after a month such members are not decided by NT,LLC, the Oversight
Body DAO may be determined by a Supermajority Decision of GovDAO, and no
further decisions may be made by GovDAO until an Oversight Body DAO is
determined that satisfies the requirements.

The Oversight Body DAO committee Requirements are as follows:

 * One Oversight Body DAO member must be from T1 or T2 (First Oversight Member).
 * One Oversight Body DAO member must be from T1 or T2 (Second Oversight Member).
 * One Oversight Body DAO member must not be a GovDAO member but otherwise be
   a Qualified Candidate with no significant conflict of interests such as by
   being invested in a competing crypto project (Third Oversight Member).

Furthermore,

 * Each member gets 1 vote. All decisions require 2 YES votes.
 * Two or more members may not reside in the same country at the time of
   appointment.
 * Two or more members may not be citizens of (or be automatically eligible
   for citizenship by birthright of) the same country or union.
 * All Oversight Body DAO members must fully understand the Constitution and
   Laws as assessed in a live interview test with T1 members.
 * Each member may serve a maximum term of three years, thereafter may not
   be eligible for re-election for another three years.
 * All members must be considered working full time and paid as a typical T1
   GovDAO member (including the Third Oversight Member) from the PayTreasury.

After the initial establishment of an Oversight Body DAO:

 * The First Oversight Member may be replaced at any time by a Supermajority
   Decision of GovDAO.
 * The Second Oversight Member may be replaced at any time by NT,LLC.
 * The Third Oversight Member who has served already for three months may be
   replaced at any time by a Supermajority Decision of GnotDAO (not GovDAO).
 * The Third Oversight Member who has served already for three months may also
   be replaced at any time by a Supermajority Decision of GovDAO AND the
   pre-approval of NT,LLC.

After any election by GnotDAO the candidate must pass the live interview test
with T1 members (as determined by a Simple Majority Decision by GovDAO), and
the test must be recorded and result shared with the public. The Third
Oversight Member candidate who gets rejected in the live interview cannot be
re-elected for another one year.

Any Oversight Member who gets replaced will be deemed to have served their
full term and cannot be re-elected for another three years.

A vacancy (e.g. from resignation, incapacitation, disqualification, ejection
from GovDAO etc) must be replaced before any Constitutional Amendment or Law
gets passed; and furthermore must be replaced within 30 days before further
proposals get passed save for any urgent Node Software Upgrades related to
software bugs or Transaction Replay Forks.

The Oversight Body DAO members committee may not self-mutate except when a
vacancy arises after which the two remaining members may self-elect by
consensus a temporary Qualified Candidate (the Temporary Oversight Member)
without restriction. This Temporary Oversight Member is expected to get
superseded (or be voted in) by the usual election rules, but such a member
will not be deemed to have been "replaced": they may be elected again soon
after. Temporary Oversight Members may vote to elect more Temporary Oversight
Members with any more vacancies.

With the exception of any Oversight Body DAO election proposals, the Oversight
Body or its subDAOs have the authority to (and must block) any decision by
GovDAO or Core DAOs if such decisions are determined to be:

 * in violation of the Constitution or the spirit of the Constitution
 * in violation of any Laws
 * in violation of any Bylaws or Mandates of any Core DAOs

The Oversight Body may block any proposals that have otherwise passed in the
prior month unless otherwise specified or a pre-approval was already granted.
This is especially important for proposals that immediately pass due to a
Supermajority Decision.

The Oversight Body does NOT have the sole authority to transfer, spend, freeze,
or burn any funds or property.

## GnotDAO

After 3 years after the launch of Gno.land GovDAO must have implemented
or chosen an implementation of GnotDAO where:

 * GnotDAO is written in Gno.
 * GnotDAO runs on a non-upgradeable (immutable) or system realm.
 * Free trading $GNOT may be bonded by users.
 * Users may use bonded $GNOT to vote on any number of valid proposals.
 * Proposals are made from a list of templates as defined by this Constitution
   and are pre-approved by at least two existing GovDAO members.
 * Proposals run with the same rules and duration as Atom.One proposals
   including any self-adjusting deposit limits to reduce spam.
 * At the end of all proposal voting periods of proposals voted on by bonded
   users the $GNOT may be returned to the user immediately.

No Constitutional Amendments may inflate the $GNOT supply above what is already
declared.

## Gno.land Laws

While Gno.land Constitutional Amendments require a Constitutional Majority
Decision, Law Amendments only require a Supermajority Decision by GovDAO.
Otherwise the procedure for making Law Amendments are similar to those for
Constitutional Amendments.

Laws may apply to GovDAO, the Oversight Body, all Core DAOs and also all users
and service providers of Gno.land.

If there are any conflicts between the Constitution and Laws the Constitution
takes precedence over the Laws. Likewise if there are any conflicts between
Bylaws of Core DAOs and Laws, the Laws take precedence over Bylaws.

Both GovDAO and the Oversight Body are responsible for ensuring that new Law
Amendments are consistent with the Constitution.

## Role and Rights of NT,LLC

NT,LLC retains the exclusive rights to word mark and brand "Gno", and "\*Gno\*".
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
Oversight Body. All Qualified Forks of Gno.land must be listed on these main
pages at all times.

Gno.land and its Qualified Forks may only use the "Gno" word mark and brand and
the "Gno.land" identity and domain and "gnoland\*" chain ID for as long as Gno
is the only contracting language on the chain as determined solely by NT,LLC;
and for as long as GnoWeb is used to render the contents; or unless allowed by
NT,LLC pre-approval on chain by cryptographic signatures by its multisig. Only
NT,LLC may determine the identity of the Gno language for any version and may
authorize releases of the GnoVM and GnoWeb. The implementation of Gno.land may
not require the GnoVM if after GnoVM is finalized the alternative
implementation is identical with the GnoVM as determined by NT,LLC and each
release of the alternative implementation is approved by NT,LLC on chain by
cryptographic signatures by its multisig.

Nothing prohibits Gno.land from changing its identity and adopting a different
name that does not include the "Gno" wordmark and brand. GovDAO may choose to
do so by a Constitutional Amendment. However in this case NT,LLC reserves the
right to determine a fork of the Gno.land chain with any modifications to the
constitution and with a new governance body. Such a fork will be considered a
Qualified Fork even if its governance member set is completely independent of
the original GovDAO members.


## Principle of the Chain

In all cases the latest released Gno.land node software must be able to run the
transactions from the first Genesis launch until current latest transaction.
This is achieved by the following:

 * A chain upgrade may have a sequential number in the chain ID.
 * Each unique chain ID has its own Genesis State & transactions from index 0.
 * The last transaction for a given chain ID results in the End State.
 * The Genesis State of the next chain ID is composed of {End State, Migration
   Script, Other Genesis Params}.
 * The Migration Script is applied to the End State to form a part of the next
   Genesis State.
 * A folder with subfolders for each chain ID in sequence where each subfolder
   includes the blockchain of blocks of txs and a partial Genesis State that
   omits the End State of the previous chain ID; and also the last node release
   for that chain ID.
 * A simple bash script to derive the latest state from the first Genesis
   by running all txs in sequence from the first chain ID to the last.

Blockchain nodes that have the full transaction history will always be able to
sync to the latest state hash from genesis using the latest released node
software and possibly also including previous node software releases. This is
to preserve integrity. If validators and nodes are not allowed to store any
offending transaction data they may prune past transactions as long as they
maintain a recent qualified snapshot as determined by this Constitution or
GovDAO and all transactions since the snapshot such that the latest blockchain
state may be derived from the snapshot.

### Gno.land Node Upgrades

All upgrades to the Gno.land Node Software must be for Fully Audited Open
Source IP (as any official software release) approved by GovDAO Supermajority
Decision that includes the exact commit hash of the software release. The only
exception to the Fully Audited requirement will be for emergency security
patches as approved by Law.

No node operator may be required to run Gno.land Node Software that cannot be
deterministically built with the source even in the case of security issues.

### Gno.land Storage Capacity

The $GNOT Storage Deposit Price (per byte) will never increase.

The $GNOT Storage Deposit Price may decrease by at most 10% a year.

The $GNOT Storage Deposit Price may not decrease if the total presently used
storage exceeds the size of a typical consumer internal hard-disk or
solid-state drive available for purchase by anyone for a PC; where such a
storage drive is manufactured by independent manufacturers of at least three
independent manufacturers and in three different countries; and such a storage
drive is at least as good as the following in key performance metrics:

 * Western Digital 10TB WD\_Black Performance Internal Hard Drive HDD - 7200
   RPM, SATA 6 Gb/s, 512 MB Cache, 3.5" - WD102FZBX

This is to keep the blockchain state at an accessible level for newcomers,
developers, and hobbyists and also accounts for any future potential economic
collapses.

### State Purge Transactions

State Purge Transactions are blockchain transactions for deleting state from a
Gno.land realm or package. Transactions that depend on state purged by State
Purge Transactions must fail with a special transaction response code.  The
Merkle-tree root hash must be derivable as if the data was there even after
purge by State Purge Transactions. This helps preserve the integrity of the
chain state for valid use-cases and makes it easier to undo when needed.

The procedures for State Purge Transactions including submission, signing,
false-positive detection, and AI model registration are defined in the Laws.

### Addressing Core Software Bugs

See "Transaction Replay Forks".

### Addressing Hacks, Theft, and Exploits.

See "Transaction Replay Forks".

## Amendments to the Constitution 

All amendments (modifications) to this Constitution including changes to any
definitions require a Constitutional Amendment pre-approved by the Oversight
Body and passed by a Constitutional Majority Decision of GovDAO following all
the rules of this Constitution.

Amendments to this Constitution must belong to one of several categories:

 1. Rewording of portions of the Constitution for clarity or refinement while
    maintaining the structure of the Constitution with no other additions or
    deletions.
 2. Restructuring to move portions of the Constitution for legibility without
    any changes to wording except for section headings or titles.
 3. Additions and deletions of portions of the Constitution without any other
    restructuring or rewordings.

Furthermore the Constitution must be one single markdown file and the latest
constitution must be present in the repository under docs/CONSTITUTION.md.

Each amendment must be composed of up to three diff patches, one for each of
the three categories in the order as declared above.

## User Rights and Limitations

User rights, protections, and limitations will be determined by future
Constitutional Amendments provided that freedom of speech will be protected
with the following exceptions:

### Banned Offensive Material

Pornographic material is defined as follows and can be further restricted by
Law:

 * photographs or animated or AI rendered depictions of humans or animals or
   humanoid figures engaged in sexual activity or depicted in an arousing way
   including games or movies that include such material; except of classic art
   hand-painted prior to the year 1985 or classic sculptures hand-made prior to
   the year 1985.

 * photographs or animated or AI rendered depictions of naked minors under the
   age of 18 or humanoid figures that could reasonably be construed to be a minor
   including games or movies that include such material; except of classic
   sculptures hand-made prior to the year 1985.

Gore material is defined as follows:

 * realistic depictions of gore whether photographic, video, animated, or AI
   rendered, and whether embedded in a movie or in a game; and excludes
   material that is submitted for educational purposes especially those related
   to political conflicts or abuses of power.

Harmful material (material harmful for users) is defined as:

 * material that is determined to result in hacks or exploits of user devices
   unbeknownst to the user. This excludes material that is submitted for
   educational purposes and includes relevant disclaimers to make clear that
   the material is harmful and is unlikely to result in accidental hacks or
   exploits of user devices.

Pornographic, Gore, and Harmful material are altogether called Banned Offensive
Material. Banned Offensive Material as determined by the relevant GovDAO DAO or
as determined by the Oversight Body may be filtered by GnoWeb or the blockchain
node in API responses in such a way that anyone who runs GnoWeb or the
blockchain node can also replicate the filter without pay or conditions and
without relying on any external service. This does not include hatespeech or
misinformation: neither GovDAO nor the Oversight Body nor any Core DAOs may
determine what is hatespeech or misinformation. This is to ensure freedom of
speech.

For the avoidance of doubt, any files released to the public as a matter of
public accountability regarding the Epstein files may reside on gno.land but
with complete redaction of the offending materials (e.g. redacted photos are
OK, explicit photos are not). The files are a matter of national and
international security, and as such they belong on Gno.land as do any
discussions. The privacy of individuals mentioned in those files is not yet
protected by this Constitution but should be clarified with an amendment.

Only Banned Offensive Material may be purged
from the blockchain state by marking code packages or state objects (including
entire realms) as such with a [State Purge Transaction](#state-purge-transactions).

The Principle of the Chain must be preserved at all times; that is, the
blockchain of blocks of transactions will not itself be modified (but the state
may be with subsequent transactions). See also "Principle of the Chain".

Code packages and realms that are primarily designed or used to bypass these
checks may also be frozen with a State Freeze Transaction with a future
Constitutional Amendment that defines the procedure and limitations but in no
case may any state be purged except for the Banned Offensive Material. A porn
realm may have all of its pornographic material purged but comments about the
porn do not fall in this category so therefore cannot be purged. This is to
prevent abuse and censorship of legitimate discussions.

### Intellectual Property and Privacy

DMCA or GDPR or similar that are concerned with intellectual property or
privacy will not affect the Principle of the Chain; this is because such laws
can be misused to censor information that the public ought to know. As with the
aforementioned other offending material each validator or node is responsible
for deleting old blockchain transaction history as needed. GovDAO may pass a
Constitutional Amendment to include DMCA or GDPR issues in the class of
material that may be purged from blockchain state with a Purge Transaction but
only after a hierarchical bonded system of manual review first takes place to
filter for a strict subset of DMCA or GDPR or similar requests in accordance
with terms defined in the Constitution under "User Rights and Limitations". In
no case will there be an automated system that purges such state, nor any
guarantees of timeliness of processing such material be offered or required.

Notwithstanding the above, such material may be filtered by GnoWeb.

### $GNOT Deposits for Purged State

Any $GNOT byte deposit tokens of state purged from the blockchain state with
State Purge Transactions (after a proving period of 6 months without being
marked on-chain as a false-positive purge by any False Positive Purge Detection
Transaction nor determined to be a false-positive by this Constitution or the
Oversight Body) may only be used for the purpose of developing or acquiring
self-contained open-source AI models and tools for automated or semi-automatic
purging of state; or be used to pay a moderation team as determined by this
Constitution. The confiscation of deposited $GNOT for such purposes must
require a special Purge Deposit Confiscation Transaction signed by any parties
as determined by this Constitution, and will fail if the purge occurred less
than 6 months ago or if the state was marked since by a False Positive Purge
Detection Transaction.

No more than half of such confiscated $GNOT deposits may be used to fund for
any manual moderation--half or more must be used for the development of
automated systems. The Oversight Body may redirect a third of these funds to
fund for the development or acquisition of alternative moderation systems that
conform to all requirements of this Constitution if it determines that the
existing moderation system is not effective, or has an unacceptable amount of
false-positives; but the adoption of any alternative system must require a
GovDAO Simple Majority Decision contingent on there being no more false-positive
purges and no more false-negatives as determined by this Constitution and the
GovDAO voting members. The Oversight Body may also block the submission of any
or all Purge Deposit Confiscation Transactions at any time if it determines
that it is being misused; after which GovDAO is responsible for the continued
funding for moderation from its own treasury until the Oversight Body resumes
the submission Purge Deposit Confiscation Transactions.

User or realm tokens or property that are not $GNOT deposit tokens of purged
state whether native tokens or Gno application derived may not be confiscated
by any means, though they may be frozen by State Freeze Transactions.

### Hacks, Thefts, and Exploits

User or realm tokens or property that are determined to be derived from any
hacks, thefts, or exploits and are not determined to have already been
transferred to unrelated parties may be frozen by a State Freeze Transaction or
the funds used to recompense associated victims as determined by a future
Constitutional Amendment and governed by the Constitution Only.

Each recompense decision must be independent of another and require a
Supermajority Decision of GovDAO. The public must be provided sufficient
information to judge the soundness of each decision.

### Transaction Replay Forks

At any time GovDAO by Supermajority Decision will have the right and authority
to address bugs in the Core Software including the GnoVM, node software, and
systems Gno code by filtering or annotating transactions or making adjustments
to state (Transaction Replay Forks). This does not apply to general Gno
ecosystem Gno code issues.

Before and until 1 year after $GNOT is transferrable GovDAO by Supermajority
Decision has the right and authority to make Transaction Replay Forks for the
purpose of addressing hacks, thefts, and exploits or other problems in the
spirit of the Constitution.

The extra-blockchain information of a Transaction Replay Fork such as
transaction annotations, filters, and state adjustments is considered a part of
the node software and not a part of the blockchain structure (of blocks of
transactions) itself, so as to preserve the Principle of the Chain.

These terms are expected to be modified before the 1 year mark to clarify or
modify the Transaction Replay Fork rules.

## Software Audits

GovDAO is responsible for ensuring that GnoWeb and Gno.land track audit reports
of all code uploaded to Gno as well as the code of the blockchain and VM
itself.

All auditing entities must be qualified; they must be vetted and whitelisted by
GovDAO Supermajority Decision on a Gno.land realm managed directly by GovDAO
(the Auditor and Audits Registry). The auditing entity must be vetted with
sufficient due diligence and already have a good reputation and track record.
This is to prevent malicious audit reports from confusing and exploiting users.
The auditing reports by such Qualified Auditors are called Qualified Audit
Reports.

Each auditing entity must be registered with general information such as the
jurisdiction of the auditing entity and the public identity of its chief
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
to show along with a prominent warning that those auditors have not been vetted.

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
least three GovDAO members two of whom are T1 or T2 members; and they will
each be responsible for ensuring that all the conditions are satisfied; however
they will not be responsible for the correctness of each Qualified Audit
Report.

## Realm Upgrading

Realms must be clearly shown in GnoWeb to be upgradeable or not upgradeable
(immutable).

 * An immutable realm may not revert to being upgradeable.

 * No types declared in upgradeable realms may be persisted in immutable realms.

 * No upgradeable realms may be imported by immutable realms.

This creates a two-class system where immutable realms only depend on other
immutable realms. The state of immutable realms may be mutated by any logic as
allowed, but the upgrading of upgradeable realms can not otherwise affect any
immutable realms.

Library packages declared in /p/ may not be upgraded.

Gnoweb must make clear to users of upgradeable realms that any funds depend on
the trust of the realm controller.

The only exception to the above rules may occur in response to a hack or
exploit. GovDAO by Supermajority Decision and pre-approval by the Oversight
Body may patch any realm package or library package if it is deemed to be in
the best interest of the Gno.land users and $GNOT economy.

TODO Rest to be determined after launch by a Constitutional Amendment.

## Name Registration

The primary purpose is for providing human readable names for package paths
(e.g. `org_name` in `"gno.land/r/org_name/pkg_name"`).

Name registration will not initially be implemented.

There will be at most one core name registration system for Gno.land, though
anyone may deploy their own for any purpose not managed by Gno.land, and as per
the Seven Mandates of Gno.land no name is required for any core services (such
as MsgAddPackage or MsgExec).

The name registration system must satisfy the following:

 * Only lowercase alphabetic ASCII characters are allowed with underscores not
   as the first character.

 * The length is between 5 and 25 characters.

 * The canonicalized name must not start with any of:
   * "gl" (and neither "gi", so as to be distinct from "g1..." addresses)
   * "gno"
   * "cosmos"
   * "atom"
   * "admin" (and neither "admln")
   * "support"
   * and more prefixes as determined by GovDAO.
   * and blacklists also determined by GovDAO.

 * Notwithstanding the above NT,LLC may reserve names that start with "gno".

 * The canonicalized name must not include (as prefix or suffix or otherwise)
   the canonicalized form of any reserved names as defined by
   https://github.com/handshake-org/hsd/pull/819 without an on-chain
   transparent process approved by GovDAO to respect trademarks with priority
   given to US, then the EU; or only if the inclusion is deemed to be
   sufficiently unique by a reasonable algorithm implemented on chain but still
   preventing collisions in the canonicalized form; or on a case-by-case basis.

 * A registration of the name is permanent. Names may be disabled by a
   procedure defined by a Constitutional Amendment that defines the criteria
   for such disablement, but such names cannot be used again by anyone else.

 * The name registration takes learnings from the Handshake protocol including
   https://github.com/handshake-org/hsd/issues/649

## Treasury diversification

Each of the above mentioned treasuries may be separately diversified by
Supermajority Decision by GovDAO of one proposal per treasury in accordance
with the Constitution--but all GovDAO members with any conflicts of tokens
(except $GNOT) above 3 months of salary for a typical senior software developer
MUST abstain from voting. If a quorum cannot possibly be reached due to
conflict of interests the Oversight Body may waive the quorum requirement.

Diversification for each treasury must adhere to a Treasury Diversification
Template listed explicitly in this Constitution by a future amendment that
declares the desired target ratio of Approved Treasury Diversification Tokens,
and the following:

 * Besides $GNOT or the top two dominant Bitcoin fork tokens (presently $BTC
   and $BCH but may change in the future) which will always be approved, all
   other tokens of the Approved Treasury Diversification Tokens must be
   stablecoins FULLY backed by collateral.
   * Coins that are FULLY backed by silver via decentralized and publicly
     auditable depositories approved by GovDAO are also
     considered stablecoins.
   * Tether and all stablecoins backed by any amount of Tether are never
     permitted.
   * Tokens that lose their approved status must be replaced (sold) for
     approved tokens as soon as reasonable.
   * These subclauses and the parent clause cannot be removed by any
     constitutional amendment.

 * $PHOTON is not a stablecoin but it may be allowed by up to 20%.

 * $ATONE tokens or any other staking tokens are not allowed.

 * In no case may the amount of $GNOT sold per treasury per month for
   diversification purposes exceed 1% of $GNOT of that treasury as measured at
   the beginning of each month.

 * In no case may the total amount of $GNOT sold for all treasuries per month
   for diversification purposes exceed 50% of the average of ($GNOT inflation
   rate, and the past month's transaction fee revenue), with priority given to
   ValTreasury, then to PayTreasury, then to CoreTreasury.

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

If such a fork proposal reaches more than 1/3 of voting power, a second fork
commitment proposal must be initiated with the same terms as the original
within a week after the first fork proposal's voting period has ended. Again
all GovDAO members vote on this second proposal. The second fork proposal must
also be voted YES by more than 1/3 of all the total voting power of GovDAO by
the end of the voting period of three weeks. Abstaining is equivalent to voting
NO.  Those who vote YES here are committing to join the others who voted YES
and split from those who voted NO or abstained if it also reaches 1/3 of voting
power--members may only serve on one chain.

If the second fork commitment proposal also reaches more than 1/3 of voting
power, two concurrent proposals must be proposed to determine which fork
retains the original Gno.land identity: one for the current chain, and another
for the proposed fork. Only T1 members may vote on these proposals; and they
may vote YES or NO on one or both proposals.

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
     by NewTendermint,LLC to point to both chains; and both chains will be
     deemed to be Qualified Forks. Any royalty payments due to the original
     chain may be directed to either fork or split between the two forks as
     determined by NewTendermint,LLC.

In all cases any new forks that include "Gno" in the name or chain ID must get
approved by NewTendermint,LLC.

All Qualified Forks may copy and use all of the state (including transaction
data) and code as from before the fork.

Non-qualified forks are not hereby prohibited from copying the state or code of
realms and packages of all forks of Gno.land. That is, everyone who submits
code or transactions to Gno.land or any of its forks are agreeing to allow (and
attesting to having the unencumbered rights to allow) the code and resulting
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

These terms must be made clear in GnoWeb and gnocli for anyone submitting code
or any transaction. Specifically for anyone uploading realm or library packages
to Gno.land using gnocli they must first sign an approved CLA compatible with
the above and include a hash of the CLA with each transaction.

## Software Security

All software including Core Software funded for by Gno.land must be Fully
Audited before any release when funded by Gno.land. This condition and
guidelines for compliance must be made explicit in any contract or agreement
before any funding decisions are approved by any Core DAO.

Fully Audited means that all dependencies of the software are also audited by a
Qualified Auditor unless exempt by the Constitution or Law. This requirement
may be relaxed for a period of 4 years after launch for existing software
unless otherwise stated by Law.

External dependencies must be regulated such that every update to the external
dependency requires audits as well or manual review of each minimal
security patch. No major or minor version upgrades will be allowed
automatically for releases; that is, development branches may allow such
automated updates but the release process must include a procedure for vetting
any changes to dependencies since the last version.

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

At no point may a GovDAO or Gno.land fund for the development of any software
or team or anything else that allows or encourages the user to enter their
private key or mnemonic on any online computer or mobile network-capable
device. The only exception is for existing hardware wallets that allow
for bluetooth, as some of these devices already support bluetooth, as long as
bluetooth can be disabled.

Only 24 word mnemonics may be supported by Core Software; not 12 or 18.

Ephemeral private keys with limited capabilities and default reasonable
limitations on losses (in the case of theft or exploit) may be generated on
online computers or mobile network-capable devices, but no ephemeral keys may
be imported manually by the user nor converted into a mnemonic so as to prevent
any confusion. The term "mnemonic" must only refer to their master private
key.

GovDAO must maintain a simple guide for users to harden their security as
defined in the Laws.

### Safety Wrapper Contracts

GovDAO must ensure the timely development of Safety Wrapper Contracts for
protecting users from theft or loss. The specifications are defined in the
Laws.

### Quantum Proof Cryptography

Gno.land is committed to quantum-proof readiness for both user accounts and
validators. The re-use of public keys should be prohibited when feasible, and
Merkle hash based commitments should be explored for account recovery after
quantum breakthroughs. The detailed roadmap is defined in the Laws.

### New Cryptographic Primitives or Implementations

New cryptographic primitives or implementations must be approved by a qualified
Cryptography Committee as defined in the Laws.

### Formal Verification of Go/Gno

GovDAO must pursue formal verification of Go or Gno programs when economically
feasible. The conditions and procedures are defined in the Laws.

## Hardware Security

GovDAO must pursue the development of open hardware and forensic analysis of
common hardware when economically feasible. The specifications and procedures
are defined in the Laws.

# Appendix

## Common DAO Spec

This is based off of atomone-hub/genesis/CONSTITUTION.md and simplified.

These common DAO specifications apply for all Core DAOs unless otherwise
specified, Special Purpose DAOs, and all sub-DAOs of these DAOs. Other DAOs
that are not Core DAOs or Special Purpose DAOs or Descendants of these DAOs
need not implement these specifications.

## Common DAO Spec - version "Gno.land Genesis"

All sub-DAOs have parent DAOs. The parent DAO of the Core DAOs and Special
Purpose DAOs is GovDAO itself. Therefore all Core DAOs and Special Purpose DAOs
as well as their sub-DAOs and GovDAO itself altogether form a tree structure.
The parent DAO and the parent DAO’s parent DAO and so on, all the way up to
GovDAO are altogether called the Ancestors of a DAO. The sub-DAOs and their
sub-DAOs and so on are called the Descendants of a DAO.

Every DAO, upon creation, must have a Charter (which is composed of Purpose and
Description), an initial set of Council members (which may be empty) and may
also have Bylaws and Mandates. The Purpose and Description must be plaintext
files. The Bylaws and Mandates must be named plaintext files or folders of
plaintext files, or folders of folders.

A DAO’s Charter, Bylaws, and Mandates may be changed by a Simple Majority vote
from any of the DAO’s ancestors, except from GovDAO which requires a Supermajority vote.

A DAO’s Bylaws, Mandates, and the Bylaws and Mandates of its ancestor DAOs, the
relevant Laws, and this Constitution, altogether are called the Governing
Documents of the DAO. 

A DAO has a Council composed of zero or more members, with no maximum number of
members unless otherwise specified in its Governing Documents.

The Council of a DAO may change the Bylaws of the DAO, and otherwise make
Decisions on behalf of the DAO by passing Proposals.

A DAO may establish any number of sub-DAOs through the DAO Council’s Simple
Majority vote, with their own defined Charters and specific bylaws and
mandates, as necessity may arise and in accordance with the parent DAO Charter
and bylaws. Sub-DAOs are owned by and can be controlled by the parent DAO, and
members are also subject to the ancestor DAOs’ bylaws and mandates.

A Simple Majority in DAO governance is defined to be exactly “more than half”.
A Supermajority in DAO governance is defined to be exactly "two thirds or
more". This is distinct from a Supermajority Decision of GovDAO.

By default, unless specified otherwise in its Governing Documents, the
following rules apply for Council voting:

- each member has equal voting power (no member may occupy multiple
  seats)  
- a Council member may resign and thereby remove themselves from the Council  
- vote options are YES, NO, or ABSTAIN
- the tally denominator is the total number of voters (ABSTAINS do not count).
- voting for proposals are open until they are decided immediately by a
  supermajority of YES votes, or dismissed immediately by a simple majority of
  NO votes, or otherwise the voting period has passed.

By default, unless specified otherwise in its Governing Documents, the
following rules apply for Council membership election:

- the Council may elect one or more new members, and/or remove one or more
  members, by Supermajority vote. (self mutating).   
- the DAO’s Ancestors may modify the Council membership with a Supermajority
  vote.

Each DAO may have an associated crypto address which can hold any number of
tokens.  

DAOs may operate with logic on Gno.land, or, represented as a m-of-n
multisig account on Gno.land where the signers are each members of the
DAO’s council, where m is more than ½ n and also m is 3 or more. In all cases
financial transactions from the DAO’s treasury must follow the passage of
governance proposals on the DAO.
