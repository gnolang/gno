# GnoDAOs Specs for v0.1

## Concept

The goal of `GnoDAOs` is to support the creation and maintenance of DAOs on Gnoland as in Aragon on Ethereum or DAODAO on Juno.

### Disclaimer

This targets test3 avl and does not work on latest master

### DAO v0.1 :

Initial version focuses on building minimum functionalities for DAOs management, and more features will be built after on.

### Moderation DAO v0.1:

Must allow community to moderate /boards or a feed in a decentralized way.

### Target #1:

1. Create a module called Moderation DAO v0.1 on Gno.land
2. Test the Moderation DAO v0.1
3. Integrate Moderation DAO v0.1
4. Open PR for changes & updates on gnoland/moderationdao

### Others permanent needs:

- Participate to Gnoland Dev Calls
- Redact a final article explaining Moderation DAO
- Redact a tutorial about Moderation DAO
- Redact full test documentation about Moderation DAO
- Integrate Gnoland in Teritori dApp
- Inform regularly all contributors about current works

---

## DAO v0.1 info

```go
type DAO struct{
    id uint64
    uri string // DAO homepage link
    metadata string // DAO metadata reference link
    funds Coins // DAO managing funds
    depositHistory []Deposit // deposit history - reserved for later use
    spendHistory []Spend // spend history - reserved for later use
    permissions []string // permissions managed on DAO - reserved for later use
    permMap map[string]map[string]bool // permission map - reserved for later use
	votingPowers     map[string]uint64
	totalVotingPower uint64
    voteQuorum uint64
    threshold uint64
    vetoThreshold uint64
}
```

```go
type Proposal struct{
    daoId uint64 // dao id of the proposal
    id uint64 // unique id assigned for each proposal
    title string // proposal title
    summary string // proposal summary
    submitTime uint64 // proposal submission time
    voteEndTime uint64 // vote end time for the proposal
    status uint64 // StatusNil | StatusVotingPeriod | StatusPassed | StatusRejected | StatusFailed
	votes map[string]Vote // votes on the proposal
}
```

## Proposal types

- Text proposal
- DAO fund spend proposal
- Update Voting power proposal
- Update URI proposal
- Update metadata proposal

## Vote options

- `Yes`: Indicates approval of the proposal in its current form.
- `No`: Indicates disapproval of the proposal in its current form.
- `NoWithVeto`: Indicates stronger opposition to the proposal than simply voting No. Not available for SuperMajority-typed proposals as a simple No of 1/3 out of total votes would result in the same outcome.
- `Abstain`: Indicates that the voter is impartial to the outcome of the proposal. Although Abstain votes are counted towards the quorum, they're excluded when calculating the ratio of other voting options above.

## Entrypoints

### Txs

- CreateDAO
- DepositIntoDAO
- SubmitDAOProposal
- VoteDAOProposal
- TallyAndExecuteDAOProposal

### Queries

- QueryDAO
- QueryDAOProposal
- QueryDAOProposals
