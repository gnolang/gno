# Escrow contract specification

## Feature Description

We’re building a decentralized version of Fiverr.com
So, we need to code an escrow system allowing:

- freelancers to be sure that the client have the money
- the client must be sure that he can trust the freelance
- being able to call for a vote if there is a conflict between them

## Escrow Contract Requirements

1. a client can create an contract by specifying the freelancer's address, escrow amount, offer expire date, contract duration(optional)
2. a client can cancel/withdraw the offer unless the seller accepted
3. a seller can't accepted the expried/canceled offers
4. a clietn can't cancel/withdraw the accepted offers
5. a client can end the contract with succuess status, it will send the escrow token to the seller.
6. a client can mint a nft that represents the feedback of the seller
7. DAO admin can end the contract if there's problem reported by client or seller. admin can decide where to send the escrowed token to send(client or seller or portions to both)

## Completed Contract flows

```sequence
Client->Escrow: Create an offer
Client-->Escrow: Send escrow tokens
Note left of Escrow: Check tokens received
Note left of Escrow: Create a new offer_id
Seller->Escrow: Accept an offer(offer_id)
Note left of Escrow: Check the seller address
Note left of Escrow: Check the expire date
Note left of Escrow: Update the offer status
Client->Escrow: Complete the contract(offer_id, feedback)
Note left of Escrow: Update the offer status
Escrow-->Seller: Send escrowed tokens
Escrow-->Seller: Mint a nft(level: 0-5)


```

```mermaid
flowchart LR
    C(Client) -- 1. Create an offer --> O
    C -. 1. send tokens .-> O
    subgraph Escrow
    O(Contract)
    O1(Contract1)
    O2(Contract2)
    O3(Contract3)
    O4(Contract4)
    end
    S(Selleer) -- 2. Accept an offer --> O
    C -- 3. Complete an offer --> O
    O -. 4. tokens .-> S
    O -. 5. NFT .-> S

```
