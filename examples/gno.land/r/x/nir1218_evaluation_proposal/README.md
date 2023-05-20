# `Contribution Evaluation`

## Abstract

This documnet describes general ideas regarding contributions evaluation, the prinicples laid out are intendenant to be part of the Evaluation DAO.

## Contents

- [Concepts](#concepts)

  - [Commity](#commity)
  - [Evaluation](#evaluation)
  - [Contibution](#contribution)
  - [Pull Request](#pull-request)
  - [Vote](#vote)

- [Future Improvements](#future-improvements)

- [Implementation](#implementation)

## Concepts

### General Ideas

Contributors DAO will designate members of a commity, in the begining, the evaluation commity members will be the core development team members or any other trusted entity.
A commity will be given the mandate to evaluate a certain set of contributions.
For example, the first commity will evaluate code contributions inside Gno central repository.
A contribution will be associated with a pull request managed in Git.
A Commity as a trusted entity can decide on a category and its corresponding evaluation criteria.
A member can propose to add a category and its corresponding evaluaiton criteria.
A member can propose a contribution for evaluation. However, the pull request category must be from the list of approved categories.
At the time of writing, a member can vote based on as set of options either "YES" or "NO", all members need to approve a category or a contribution.

### Commity

A group of designated members who are given a manadate to act as an evalution authority

### Evaluation

A logical entity to group a certain types of contributions.

#### Category

A group of contributions that should be evaluated based on the same priciples and guide lines.

### Contribution

A contribution is associated with a pull request.
A contribution is an evaluation life cycle.

#### Submission

Any commity member can submit contributions.

#### Approval

A contribution is approved once it reaches a certain treshold.

### Pull Request

A pull request from source control tool namely GitHub.

### Vote

#### Voters

Voters are commity members, all commity members have the right to vote on a contribution.

#### Voting Options

The voting options available to a voter.
A commity may set voting options for its evaluation categories.
The initial option set includes the following options:

- `YES`
- `NO`

#### Threshold

Threshold is the minimum percentage of `YES` votes from the total votes.

#### Tally Votes

## Future Improvements

The current documentation describes the basic ideas as expressed in the code.
Future improvements listed below will be decided based on future discussions and peer reviews.

- Commity negotiates contributions
- A commity may set voting options for its categories and evaluated contributions, otherwise; the Contributors DAO may set a global
- A commity may set a threshold required for a catgory or a contribution to be approved, otherwise; the Contributors DAO may set a global threshold and quorum.
- A commity sets evalution criteria scoring range (1-10), scoring a contribution is essential when there are competing contributions (Game of Realm). Otherwise, the evaluaiton is a binary decision.
- Commity members asses contributions based on the evaluation criteria and vote accordingly

## Implementation

The implementation written is to express the ideas described above using code.
