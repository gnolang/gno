# Peace! or, Everyone is Invited to Gno.land, if you want!

I've never been put in such a difficult position, of having information that I
cannot reveal. And if you know me, you know that I like to speak my mind.  But
I cannot say the things that I would rather say, because you get a lot of flack
for saying anything bad about a public chain.

So I have been sitting on this issue, losing sleep about it for years, because
it leads me to worry about the safety of the hub. From an external person's
point of view, the solution is obvious -- reveal the information for the
betterment of everyone, no matter the consequences, because that is the right
thing to do. As a stakeholder, and I agree with the majority of the community
that peace and silence is better, with exceptions.

So without turning this into a war of accusations bringing back past drama,
let's just do this: dear core contributors, Ethan Buchman, Zaki Manian, Jack
Zampolin, and everyone, here is my peace plan.

----------------------------------------

## On Prop 69

Prop 69 is about adding CosmWASM to the hub.  I have repeatedly talked about
the dangers of adding CosmWASM to the hub, including a document shared two
years ago.

https://github.com/jaekwon/cosmos_roadmap/tree/master/shape_of_cosmos#smart-contracts

Even before prop 69, I had declared publicly that stakers voting yes to adding
WASM on the hub would not receive airdrops.  Primarily, because it increases
the surface area for attack by an order of magnitude.  CosmWASM adds two layers
of new complexity to the hub. WASM itself, as well as CosmWASM.  WASM as a spec
and its implementations are still maturing, and though available on browsers,
and some blockchains, it still hasn't gone through the gauntlet of time. All
new complex technologies like WASM, like Java, Linux, and even Go, in hindsight
have numerous bugs that could have or were used maliciously. The same will be
true of any WASM integration with the hub, and this potential for exploits
combined with the massive potential rewards (especially of pegged PoW tokens)
makes such exploits an inevitability.

In Juno recently there was a bug that halted the chain for three days.  Worse
can happen on the Cosmos Hub. The very identity of the Cosmos Hub (it's most
valuable asset is specifically a schelling point brand, of being a "common IBC
hub") is threatened if a bug were to result in the theft or loss of coins. On
platforms like Ethereum or Polkadot, perhaps they would have a better time
rolling back the chain to undo a hack as in the DAO hack. The major difference
with an *IBC hub* is that it cannot simply reverse the transactions of other
chains.

We have yet to experience such a bug in any of our zones on a major scale, and
have yet to learn how to coordinate in the case of such in an interconnected
web of zones. Where are the planning documents for disaster scenarios? Between
PoS chains with good governance, we will learn how to roll back transactions
across connections, if need be in exceptional circumstances, but we aren't
there yet. This option isn't even available with pegged PoW coins.

Yes, the contracts that are approved to run will be governance gated, but this
is not enough. For one, even with perfect governance, there are two new pieces
of complexity that will see more zero day bugs in the future for exploitation.
In terms of governance, the contracts are probably going to be written in Rust,
and so suddenly the validators that joined the project by inspecting the Go
code is now required to also audit Rust code. But also, we are now truly
opening the doors to all kinds of contracts to be run, because while governance
does sometimes reject proposals, it is generally accommodating to new features
especially endorsed by core contributors.

I know of three alternatives:

(1) we can use IBC to offload features to other zones. For liquid staking
(which should not be the focus of the hub) the hub could allow validators to
restrict the destination of unbonded ATOMs, and smart contracts running on
other zones can distribute those ATOMs according to the logic of whatever
liquid staking contract. This ensures separation of concerns, and a minimal
hub.

(2) we can use Go plugins to extend the functionality of the chain.

(3) we can do nothing. if liquid staking is such a big deal, something is wrong
about priorities for a cosmic "hub". If the liquid staking market is larger
than the base non-liquid staking market, the system is open for manipulation
and is insecure.  The focus should not be on self-limiting use-cases, but the
infinite market of running validators with replicated security, perhaps running
a simple dex, and most of all innovating on and offering interchain security,
the business of judging validation faults as related to Tendermint, and perhaps
the interpretation and enforcement of self-enforced customs (law) of a
blockchain as defined by its shareholders who defer validation (and perhaps
judicial services) to the Cosmos Hub because it has a reputation for being the
longest ever running proof of stake hub that has never gone down, even as
compared to the upcoming Ethereum2.0.

And note, I'm not proposing that the ATOM stakers forgo the benefits of
supporting contracts with CosmWASM. I support Juno and Tardigrade and Ethan
Frey’s work, but I also support the Hub running shared security, especially
simple replicated shared security where the validators also validate other
chains. I think this, and interchain staking, are the only profit models needed
for the hub (besides being a hub). NOTE: But those "consumer chains" ought to
be provided with full disclosures that the Cosmos Hub validators do not
maintain their respective software (as it would be impossible to audit all
zones that would benefit from the hub's security) but only offering validation
services as-is. This would force the hub validators to solve process isolation
(and I would much prefer building the protocol to NOT require particular
solutions like Docker, but allows validator choice), or else they would quickly
get slashed from malware (and that would be good to prune those validators from
the hub).

So many options that don't require putting WASM on the Cosmos Hub.

------------------------------------

## On Incentivized Votes

In corporations, you can buy shares to influence the outcome of governance
votes.  In democracy, this is not allowed because the vote could be bought to
infringe upon the rights of other people.

What do you do when the chain's own core contributors proposes a proposal that
you judge damages the integrity of the system? I think that's a good time to
create a fork of the hub's ATOM distribution led by a new development team.
Sometimes this option is the only option because of safety concerns, and this
is the case for me here.

### Why is the snapshot date 5/19/2022?

A snapshot in the past is more vulnerable to insider gaming, because there is
an imbalance of information--only the coordinator knows, and so can game the
premine.

It is good to give many people the advantage of participating in a snapshot.
Excluding anyone who would have been an ally of a chain, in turn creates
animosity that would rather see another project succeed where they are
included.

Even before the proposal I had pre-declared that anyone who votes for WASM on
the hub would not receive a gno.land airdrop. The proposer probably knew this
when the proposal was submitted.

The snapshot date would have been 7/4/2022, because that is Independence Day in
the United States. I originally chose Independence Day because of the general
original mission of Tendermint, Cosmos, Bitcoin, and the crypto spirit; and
because the United States (as flawed as it is) is the best historic ideal of
human liberty we've had since before the days of Rome.

Then prop 69 was submitted. I had said previously that we would exclude those
who vote in favor of WASM on the hub, but we don't have the tools yet to tally
the movement of tainted ATOMs after the unbonding period for the hub. So I
decided to move the snapshot date to 5/19/2022.

Now with prop 69, I see that to me, 21 days after the beginning of proposal
\#69, 5/20/2022 (but 5/19/2022 PDT) is a chance to create a new community within
the Cosmos ecosystem that champions safety with a zero tolerance policy and a
mission to develop social coordination tools like the GNO smart contract VM, to
create even better governing bodies than the one we have today.

### Gno.land and Cosmos Hub

Now, I feel compelled to exit should prop \#69 pass. But as it is now, 16.57%
are voting YES, while NO and NO WITH VETO have 70.73% and 8.38% of the votes
with turnout at 30%. If the proposal does not pass, I would feel no need to
exit. For as long as the Cosmos Hub remains minimal and secure, we will favor
it as the dominant or only token hub connected to gno.land via the current IBC
implementations for the purpose of interchain token transfers. It's a job that
we'd rather not solve, as specialization is what will get us to the finish line
before other platforms do, and also I'm quite hooked on gnolang programming and
just want to make gnolang apps. Not everybody wants to build a DTCC, but many
would prefer to use it.

### Airdrop distribution

When I was asked on Cryptocito what I would have changed if I were to do it all
again, well, I would put the ICF in the hands of the chain. So in gno.land, the
ICF's portion of $GNOT will go to DAOs on gno.land. As for me, I have a
significant amount of ATOMs that voted for NO WITH VETO, but most of my tokens
by far are with the company that I previously founded, then called All in Bits,
Inc. AIB will not receive any $GNOT except by completing negotiations with me,
which is taking a lot longer than is reasonable--or not.

For reference, for the genesis of the Cosmos Hub, the total distribution for
both entities was 20% of all ATOMs, and today it is still significant. The
total premine that I control directly or indirectly will not exceed 1/3 of the
total $GNOT distribution, but I am considering 20% again.

Some more guidelines, which may change, so don't take anything here as
financial advice:

 * NO with VETO is slightly better than NO.
 * NO is better than ABSTAIN.
 * ABSTAIN is better than not voting at all.
 * Delegators inherit the votes of the validators (unless delegators override).
 * If you vote YES on \#69, you will not receive gno.land $GNOTs.

Regarding ATOMs locked in IBC channels, those will count as not voting,
which is fine, and corresponding $GNOT will be released once the respective
zone communities create a provable audited distribution given the snapsot date,
which is defined to be exactly when those who voted on #69 can unbond and move their
tokens, which is 21 days from the start of the proposal. Soon we will be more precise
about what that is, but communities please prepare accordingly. If you aren't sure,
leave ATOMs on the hub.

NOTE: If you don't like my airdrop rules, you are free to make your own, and if
you're nice you can even run gno.land contracts if you so want there, or you
can just run a fork of gaia.

If you have a better ideal for such an exit-drop by tweaking the governance
module, I'd love to hear your feedback, or generally how you think I could have
done this better. Some say that they don't want to see more of this kind of
forking, but I think we ought to celebrate it instead.

----------------------------------------

## Conclusion

Here's a peace offering.

Just change your vote from YES to NO, and I will not intervene upon the second
submission of the proposal (and I would even fund its deposit if need be). But
if you instead feel strongly about signaling in favor of CosmWASM, here you can
express it, and I celebrate you, for being different than I, and wish you the
best of luck. That is equivalent to a no-confidence vote on gno.land, and is a
proper way to diss me. Again, I salute you.

If you can reconsider your vote to be a NO, or even better, a NO WITH VETO, I
welcome you to gno.land. Happy 5/19/2022 (5/20/2022 Europe) Gno.land
Independence Day!
