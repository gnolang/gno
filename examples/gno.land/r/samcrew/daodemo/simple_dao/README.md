# Simple DAO Demo - [simple_dao](/r/samcrew/daodemo/simple_dao)

Basic elementary DAO example with roles and voting.

- Creates a DAO with 2 roles: `public-relationships` and `finance-officer`
- Members can vote on proposals (60% majority needed)
- Add/remove members and assign roles

## Core Functions

- `Vote(proposalID, vote)` - Cast your vote on a proposal
- `Execute(proposalID)` - Execute an approved proposal
- `Propose(proposalRequest)` - Create a new proposal (requires MsgRun)
- `Render(path)` - Display DAO state and UI

## Utility Functions

- `AddMember(address, roles)` - Directly add member (admin only)
- `ProposeAddMember(address, roles)` - Create proposal to add member

## Running Scripts

Use the transaction script to create proposals:

```bash
# Create a proposal to add a new member
gnokey maketx run \
  --gas-fee 100000ugnot \
  --gas-wanted 10000000 \
  --broadcast \
  MYKEYNAME \
  ./tx_script/create_proposal.gno
```

## Files Overview

- `simple_dao.gno` - Main DAO implementation
- `utils.gno` - Helper functions for testing
- `tx_script/create_proposal.gno` - Example transaction script
- `simple_dao_test.gno` - Unit tests
