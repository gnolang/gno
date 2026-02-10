# `Contribution Evaluation`

## Abstract

This document describes general ideas regarding contributions evaluation. The principles laid out are intended to be part of the Evaluation DAO.

## Contents

- [Concepts](#concepts)

  - [Committee](#committee)
  - [Evaluation](#evaluation)
  - [Contribution](#contribution)
  - [Pull Request](#pull-request)
  - [Vote](#vote)

- [Future Improvements](#future-improvements)

- [Implementation](#implementation)

## Concepts

### General Ideas

Contributors DAO will designate members of a committee. In the beginning, the evaluation committee members will be the core development team members or any other trusted entity.
A committee will be given the mandate to evaluate a certain set of contributions.
For example, the first committee will evaluate code contributions inside Gno central repository.
A contribution will be associated with a pull request managed in Git.
A Committee as a trusted entity can decide on a category and its corresponding evaluation criteria.
A member can propose to add a category and its corresponding evaluation criteria.
A member can propose a contribution for evaluation. However, the pull request category must be from the list of approved categories.
At the time of writing, a member can vote based on as set of options either "YES" or "NO", all members need to approve a category or a contribution.

### Committee

A group of designated members who are given a mandate to act as an evaluation authority.
A DAO may elect a committee and designate its members based on contributions or merits of the members.
A committee member can propose a contribution to avoid spam and confirm viable contributions will be evaluated.

### Evaluation

A logical entity to group a certain types of contributions.

#### Category

A group of contributions that should be evaluated based on the same principles and guide lines.
An example of a category is a bounty, a chore, a defect, or a document.

### Contribution

A contribution is associated with a pull request.
A contribution has an evaluation life cycle.
A submission time is set when a contribution is added.
A last evaluation time is set when a contribution is evaluated and approved by a member.
An approval time is set when a contribution is approved by all members (or when a future threshold is reached)

#### Submission

Any committee member can submit a contribution.

#### Status

When a contribution is submitted its status is set to "proposed", its status will change to "approved" once approved by the committee or to "declined" otherwise.
Intermediate status options such as "negotiation", "discussion", "evaluation" are TBD.
A further discussion around the idea of deleting a contribution is required as it raises questions regarding record keeping, double evaluations, and the motive.

#### Approval

A contribution is approved once it reaches a certain threshold.

### Pull Request

A pull request from a source control tool, namely GitHub.

### Vote

#### Voters

Voters are committee members, all committee members have the right and obligation to vote on a contribution.

#### Voting Options

The voting options available to a voter.
A committee may set voting options for its evaluation categories.
The initial option set includes the following options:

- `YES`
- `NO`

#### Voting Period

Voting period is set by the committee, all committee members are obligated to vote within the voting period.

#### Threshold

Threshold is the minimum percentage of `YES` votes from the total votes.

#### Tally Votes

## Future Improvements

The current documentation describes the basic ideas as expressed in the code.
Future improvements listed below will be decided based on future discussions and peer reviews.

- Committee negotiates contributions
FIXME Next line is unfinished:
- A committee may set voting options for its categories and evaluated contributions, otherwise; the Contributors DAO may set a global
- A committee may set a threshold required for a category or a contribution to be approved, otherwise; the Contributors DAO may set a global threshold and quorum.
- A committee sets evaluation criteria scoring range (1-10), scoring a contribution is essential when there are competing contributions (Game of Realm). Otherwise, the evaluation is a binary decision. Moreover, scoring should be translated to rewards of any sort, or become discussion points durning negotiation about the viability of a contribution.
- Committee members assess contributions based on the evaluation criteria and vote accordingly.

## Implementation

The implementation written is to express the ideas described above using code. Not all ideas have been implemented.
