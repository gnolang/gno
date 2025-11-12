# Custom Resource Demo - [custom_resource](/r/samcrew/daodemo/custom_resource)

DAO with custom blog post actions.

- Creates custom action: `NewPost` for blog posts
- Members vote on proposed blog content (60% majority needed)
- Automatically publish approved posts

## Core Functions

- `Vote(proposalID, vote)` - Cast your vote on a proposal
- `Execute(proposalID)` - Execute an approved proposal
- `Propose(proposalRequest)` - Create a new proposal (requires MsgRun)
- `Render(path)` - Display DAO state and blog posts

## Utility Functions

- `AddMember(address, roles)` - Directly add member (admin only)
- `ProposeAddMember(address, roles)` - Create proposal to add member
- `ProposeNewPost(title, content)` - Create proposal for new blog post

## Running Scripts

Use the transaction script to create proposals:

```bash
# Create a proposal for a new blog post
gnokey maketx run \
  --gas-fee 100000ugnot \
  --gas-wanted 10000000 \
  --broadcast \
  MYKEYNAME \
  ./tx_script/create_proposal.gno
```

## Files Overview

- `custom_resource.gno` - Custom action definition
- `blog.gno` - Blog post storage and management
- `simple_dao.gno` - Main DAO implementation
- `utils.gno` - Helper functions for testing
- `tx_script/create_proposal.gno` - Example transaction script
- `custom_resource_test.gno` - Unit tests
