# Justice DAO

The goal of Justice DAO is to select random members from the DAO members to resolve the conflict on Escrow service.

Those random members selected all have same voting power and those members do on-chain chat and resolve the conflict between two parties.

We consider the conflict proposal as a specific proposal called `Justice DAO Proposal`, while normal proposal is for internal DAO management.

Justice DAO uses `VRF` realm to determine random members.

## Data structure

Justice DAO has 3 main sections.

- `DAO` to manage overall members and DAO itself

```go
type DAO struct {
	uri              string    // DAO homepage link
	metadata         string    // DAO metadata reference link
	funds            uint64    // DAO managing funds
	depositHistory   []string  // deposit history - reserved for later use
	spendHistory     []string  // spend history - reserved for later use
	permissions      []string  // permissions managed on DAO - reserved for later use
	permMap          *avl.Tree // permission map - reserved for later use
	votingPowers     *avl.Tree
	totalVotingPower uint64
	votingPeriod     uint64
	voteQuorum       uint64
	threshold        uint64
	vetoThreshold    uint64
	numJusticeDAO    uint64 // number of justice DAO members on justice proposal
}
```

- `Proposal` to manage interal DAO related proposals

```go

type Vote struct {
	address   std.Address // address of the voter
	timestamp uint64      // block timestamp of the vote
	option    VoteOption  // vote option
}


type VotingPower struct {
	address string
	power   uint64
}

type Proposal struct {
	id           uint64         // unique id assigned for each proposal
	title        string         // proposal title
	summary      string         // proposal summary
	spendAmount  uint64         // amount of tokens to spend as part the proposal
	spender      std.Address    // address to receive spending tokens
	vpUpdates    []VotingPower  // updates on voting power - optional
	newMetadata  string         // new metadata for the DAO - optional
	newURI       string         // new URI for the DAO - optional
	submitTime   uint64         // proposal submission time
	voteEndTime  uint64         // vote end time for the proposal
	status       ProposalStatus // StatusNil | StatusVotingPeriod | StatusPassed | StatusRejected | StatusFailed
	votes        *avl.Tree      // votes on the proposal
	votingPowers []uint64       // voting power sum per voting option
}

```

- `JusticeProposal` for external requests to resolve.

```go
type JusticeProposal struct {
	id           uint64         // unique id assigned for each proposal
	title        string         // proposal title
	summary      string         // proposal summary
	vrfId        uint64         // the vrf request id being used to determine governers
	governers    []string       // the governers of the proposal
	contractId   uint64         // the escrow contract id to resolve
	sellerAmount uint64         // the seller amount determined by Justice DAO
	solution     string         // proposed result of justice DAO proposal
	submitTime   uint64         // solution submission time
	voteEndTime  uint64         // vote end time for the proposal
	status       ProposalStatus // StatusNil | StatusVotingPeriod | StatusPassed | StatusRejected | StatusFailed
	votes        []Vote
}
```

## Realm configuration process

Create DAO by usin `CreateDAO` endpoint

## DAO internal proposals flow

- `CreateProposal` to create an internal proposal
- `VoteProposal` to vote an internal proposal
- `TallyAndExecute` to execute the finally voted proposal

## DAO justice proposals flow

- `CreateJusticeProposal` to create a justice DAO proposal, this requests random number to `VRF` and it will be needed to wait until the required number of random word feeders to feed the words
- `DetermineJusticeDAOMembers` to determine random members from Justice DAO
- `ProposeJusticeDAOSolution` to propose Justice DAO solution by one of elected Justice DAO member
- `VoteJusticeSolutionProposal` to vote on Justice DAO solution
- `TallyAndExecuteJusticeSolution` to execute the finally voted justice DAO solution
