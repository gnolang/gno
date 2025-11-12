# Custom Condition Demo - [custom_condition](/r/samcrew/daodemo/custom_condition)

DAO with custom voting rules (only roleless members vote).

- Creates custom condition: `NoRole` - only members without roles can vote
- Members vote on proposals using special voting rules
- Demonstrates advanced governance patterns

## Core Functions

- `Vote(proposalID, vote)` - Cast your vote on a proposal
- `Execute(proposalID)` - Execute an approved proposal
- `Propose(proposalRequest)` - Create a new proposal (requires MsgRun)
- `Render(path)` - Display DAO state and voting rules

## Utility Functions

- `AddMember(address, roles)` - Directly add member (admin only)
- `ProposeAddMember(address, roles)` - Create proposal to add member
- `ProposeNewPost(title, content)` - Create proposal using custom condition

## Running Scripts

Use the transaction script to create proposals:

```bash
# Create a proposal with custom voting condition
gnokey maketx run \
  --gas-fee 100000ugnot \
  --gas-wanted 10000000 \
  --broadcast \
  MYKEYNAME \
  ./tx_script/create_proposal.gno
```

## Files Overview

- `custom_condition.gno` - Custom voting condition implementation
- `simple_dao.gno` - Main DAO implementation
- `utils.gno` - Helper functions for testing
- `tx_script/create_proposal.gno` - Example transaction script
- `custom_condition_test.gno` - Unit tests
