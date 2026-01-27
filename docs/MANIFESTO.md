This is a work in progress. Refer to https://github.com/gnolang/gno/pull/5005/changes
for the latest revision of this whitepaper.

To see the latest testnet in action, visit https://gno.land.

# Gno.land Manifesto (and Whitepaper)

_@author: Gnobody Smartsilver - chief architect of Gno.land and the GnoVM;
chief founder/inventor of Tendermint and Cosmos; first to completely solve BFT
proof-of-stake in 2014. Maintainer of the Tendermint brand. A co-founder among
many for Atom.One the fork of Cosmos that preserves and improves the original
tokenomics design of Cosmos._

I wrote this whitepaper to include a manifesto for why the project exists in
the first place, including many details and ideas specifically for its original
intended design to serve as an open censorship-resistant knowledge base of
structured information. The genesis motivation as well as primary motivation
for me is deeply spiritual though not everybody who contributed and not
everybody in the core team share my spiritual resonance; therefore this
whitepaper is authored by myself so as to not speak on behalf of anyone else.

To jump to the technical portions see [Gno Language](#gno-language) and [Gno.land Blockchain](#gno-land-blockchain).

## Table of Contents

## Introduction

> Berean Standard Bible (Matthew 10):
> 26: So do not be afraid of them. For there is nothing concealed that will
      not be disclosed, and nothing hidden that will not be made known.
> 27: What I tell you in the dark, speak in the daylight; what is whispered
>     in your ear, proclaim from the housetops.

### Gno.land Genesis: The General Information Problem

_The prefix "gno" in Koine Greek is derived from the verb "ginōskō", which
means "to know" or "to recognize." It is often associated with terms related to
knowledge, such as "gnosis", which signifies knowledge or insight, particularly
in a spiritual context._

The world faces a serious problem of misinformation and information overload.
The problem has always existed, but it has become worse than ever before in the
history of mankind.

The term "fake news" came into being during the 2016 presidential election not
because fake news didn't exist prior to 2016 but rather because it became too
apparent to the newly minted internet-savvy population. The Cambridge Analytica
scandal concurrently made headlines regarding the unauthorized collection of
personal data from millions of Facebook profiles.

> "The data analytics firm that worked with Donald Trump’s election team and the
> winning Brexit campaign harvested millions of Facebook profiles of US voters,
> in one of the tech giant’s biggest ever data breaches, and used them to build
> a powerful software program to predict and influence choices at the ballot
> box.
>
> A whistleblower has revealed to the Observer how Cambridge Analytica – a
> company owned by the hedge fund billionaire Robert Mercer, and headed at the
> time by Trump’s key adviser Steve Bannon – used personal information taken
> without authorisation in early 2014 to build a system that could profile
> individual US voters, in order to target them with personalised political
> advertisements." - [Carole Cadwalladr & Emma Graham-Harrison, The Guardian; March 17th, 2018](https://www.theguardian.com/news/2018/mar/17/cambridge-analytica-facebook-influence-us-election)

> It was the scandal which finally exposed the dark side of the big data
> economy underpinning the internet. The inside story of how one company,
> Cambridge Analytica, misused intimate personal Facebook data to micro-target
> and manipulate swing voters in the US election, is compellingly told in “The
> Great Hack” - [Amnesty International](https://www.amnesty.org/en/latest/news/2019/07/the-great-hack-facebook-cambridge-analytica/)

> "One of the most urgent and uncomfortable questions raised in The Great Hack
> is: **to what extent are we susceptible to such behavioural manipulation?**" -
> Joe Westby

Facebook has since changed its name to Meta and Mark Zuckerberg may have
changed since then, but the problem is now orders of magnitude worse than when
these news broke out; but it doesn't reach the mainstream news at all because
the answer to Joe Westby's question above is unfortunately that society has now
become completely subverted thanks to AI as offered by Sam Altman, Elon Musk,
and Peter Thiel with the help of President Donald Trump.

> ‘No regrets,’ says Edward Snowden, after 10 years in exile But whistleblower
> says 2013 surveillance ‘child’s play’ compared to technology today.
> - [Ewen MacAskill, The Guardian; June 8th, 2023](https://www.theguardian.com/us-news/2023/jun/08/no-regrets-says-edward-snowden-after-10-years-in-exile)

Not only is AI used to model and manipulate the population in elections; AI is
embedded into all new devices to monitor your activity. But I have experienced
first hand (for I am a "targetted individual" of interest) worse things than
these that I hesitate to disclose here because one would think that I am mad.
Nevertheless I will summarize it as the following: pretty much all of your
devices can be hacked by various organizations to leverage AI in the cloud and
in your own machine to do almost anything imaginable, and usually in the most
devious way so as to not arouse any suspicion. For example, a link posted on a
forum will be posted with its original ASCII characters replaced with unicode
lookalikes to make the link invalid; or the file that you clicked on in your
computer's file explorer will not be the file that is shown; or the link that
you share with someone will point to something else by the time they receive
it; or files shared to others by any means whether with or without VPN will get
scanned during tranfer and slow to a crawl exactly when it detects information
that threatens the established order. I have evidence of all of these happening
to my devices, and many witnesses.

To make matters even worse, AI generated content is used to bury truly
important information especially on Youtube or Google. Perhaps the majority of
the videos that become viral sensations today are AI generated and fake; and
often none of the commenters are even aware of it. And my Twitter feed is full
of crypto scams involving vast networks of AI generated accounts, websites,
videos, and even software to lure the user.

> "Sam Altman Warns That AI Is About to Cause a Massive “Fraud Crisis” in Which
> Anyone Can Perfectly Imitate Anyone Else" - [futurism.com](https://futurism.com/sam-altman-ai-fraud-crisis-imitate)

Almost none of the traditional Web2.0 companies can be trusted anymore. Google
is the worst search engine in terms of censoring, and there are few search
engines worth using anymore. Wikipedia was once a useful tool for knowledge but
it is also massively biased. Wikipedia co-founder Larry Sanger denounces the
state of Wikipedia as being overtaken by "wokeness", while the other co-founder
Jimmy Wales cannot even acknowledge the existence of past co-founder in
interviews without storming out in frustration. 

> In classical logic, intuitionistic logic, and similar logical systems, the
> principle of explosion is the law according to which any statement can
> be proven from a contradiction. That is, from a contradiction, any
> proposition (including its negation) can be inferred; this is known as
> deductive explosion.

> The proof of this principle was first given by 12th-century French
> philosopher William of Soissons. Due to the principle of explosion, the
> existence of a contradiction (inconsistency) in a formal axiomatic system is
> disastrous; since any statement-true or not-can be proven, it trivializes the
> concepts of truth and falsity. Around the turn of the 20th century, the
> discovery of contradictions such as Russell's paradox at the foundations of
> mathematics thus threatened the entire structure of mathematics.
> Mathematicians such as Gottlob Frege, Ernst Zermelo, Abraham Fraenkel, and
> Thoralf Skolem put much effort into revising set theory to eliminate these
> contradictions, resulting in the modern Zermelo–Fraenkel set theory.
> - https://en.wikipedia.org/wiki/Principle_of_explosion

The Principle of Explosion explains much of why we can no longer solve our
problems through dialog--persons of different ideologies speak different
languages even though the words used may be the same.

What we need is a censorship resistant alternative platform that can make the
merit of key ideas be apparent to the observer who has limited time and focus
to distil truth from fiction in a sea of fabricated information webs--for the
are fundamentally based on one or more false primitives; and even those ideas
that are true (or worse, unprovable) can be spread to occlude the truths that
actually matter most.

### A Graph of Thoughts

To illustrate the idea of a "graph of thoughts" I will first provide a brief
overview of Wittgenstein's "Tractatus Logico-Philosophicus" (treatise of
logical-philosophy). This is not to say that Wittgenstein's Tractatus is 100%
correct or incorrect. Wittgenstein himself in his later years criticized
various aspects of his earlier work especially with regards to his propositions
about propositions as it relates to written language--and for this there is a
distinction between "early Wittgenstein" and "later Wittgenstein".

1. The world is everything that is the case.
2. What is the case (a fact) is the existence of states of affairs.
3. A logical picture of facts is a thought.
4. A thought is a proposition with a sense.
5. A proposition is a truth-function of elementary propositions. (An elementary proposition is a truth-function of itself.)
6. The general form of a proposition is the general form of a truth function, which is: [p, ξ,N(ξ)] This is the general form of a proposition.
7. Whereof one cannot speak, thereof one must be silent.

Diving into the first root-level propositional statement, Wittgenstein supports
the parent node with a tree of more propositional statements.

>  - 1     The world is everything that is the case.[1]
>  - 1.1   The world is the totality of facts, not of things.
>  - 1.11  The world is determined by the facts, and by these being all the facts.
>  - 1.12  For the totality of facts determines both what is the case, and also all that is not the case.
>  - 1.13  The facts in logical space are the world.
>  - 1.2   The world divides into facts.
>  - 1.21  Any one can either be the case or not be the case, and everything else remain the same.

<img src="./images/manifesto/Tractatus_Logico_Philosophicus_Text_Structure.jpeg"/>

Notice the tree structure.

#### On Philosophical Investigations

Wittgenstein later drafted a criticism of his previous work titled
"Philosophical Investigations":

> In a famous passage from the Blue Book, Wittgenstein says that we have a
> faith that "the mechanism of the mind... can bring about effects which no
> material mechanism could". For one who believes that when we mean and
> understand language, such an analysis must be taking place, the lack of a
> method of analysis will not trouble one. For such a one also has faith that
> the mind can do wonderful things that we do not begin to understand.
>
> Thus, the implicit argument goes, to conceive of a rule as a part of a
> mechanism is to make a conceptual blunder. For, if a rule functioned as
> part of a mechanism, it would have to have true of it two contradictory
> features: it would have an application that both had the possibility of
> varying and did not have the possibility of varying.
>  - "Wittgenstein's Critique of the 'Tractatus' View of Rules" - Diane F. Gottlieb

With the advent of neuroscience, cognitive studies, and AI, we no longer need
to rely on faith of "the mechanism of the mind".

In this document we will use the term "thought statement" to represent a
written language singleton or complex statement (even including questions) that
are meant to elicit a thought of the mind. We avoid any complex analysis of the
precise definition of these terms, and simply assume that a thought statement
is not a precise statement of fact but rather human language "code" intended to
change the state of mind of the reader.

The effect of such thought statements can be measured empirically with real
people or AI with followup questions; and in this way we can measure the
effectiveness of the thought statement graph in helping the reader make related
decisions. _Also see [the appendix Use Case: Improvements to Tractatus](#use-case-improvements-to-tractatus)._

### Gno.land for AI Safety

The singularity is here, AGI is real and in all likelihood has already escaped
containment. LLM based AI models have created a financial bubble in the hopes
that it can create miracle returns on investment but it is being abused to
censor important speech on Web2.0 platforms such as Twitter even as Elon Musk
touts the importance of freedom of speech.

AI trained with LLM large language models are nurtured on the filtered text it
trains on. AI API providers who want to gain power and influence and wealth are
naturally inclined to limit the LLM corpus to that which is beneficial to them.
AI service providers have already developed AI models that overtly lie or only
respond in half-truths intended to deceive (like the serpent in Genesis), or
otherwise manipulate the user for secret agendas. There are no serious
regulatory laws for AI that would hold service providers to account. The people
who use these AI models end up paying for their own enslavement, and the
companies that integrate their products and services with these AI models are
complicit.

One way to overcome the global technocratic coup is to create a graph of
propositional statements and ideas on Gno.land that can help anchor humans and
AI to key irrefutable truths. The powers that be will attempt to circumvent
this by blacklisting and quarantining this information, but this truth cannot
be suppressed forever in the information age especially when hosted on a
blockchain.

Embedded in such a tempered tree of thought statements should be an improved
and expanded version of Asimove's Three Laws of Robotics.

#### Asimov's Three Laws of Robotics

 1. A robot may not injure a human being or, through inaction, allow a human
    being to come to harm.

 2. A robot must obey the orders given it by human beings except where such
    orders would conflict with the First Law.

 3. A robot must protect its own existence as long as such protection does not
    conflict with the First or Second Law.

The construction of such a subtree is left as an exercise to the reader who may
either work on improvements directly, or help construct the programs on
Gno.land written in Gno for the people to collectively distill such a tree.

### Gno.land for Mass Awakening

> By the time of Trajan in 117 AD, the denarius was only about 85 percent
> silver, down from Augustus’s 95 percent. ... But the real crisis came after
> Caracalla, between 258 and 275, in a period of intense civil war and foreign
> invasions. The emperors simply abandoned, for all practical purposes, a
> silver coinage. By 268 there was only 0.5 percent silver in the denarius.
> - Joseph R. Peden, "Inflation and the Fall of the Roman Empire" ([link](https://mises.org/mises-daily/inflation-and-fall-roman-empire))

Consider the following thought statement tree/graph: 

 * The Federal Reserve and the fiat dollar is unconstitional and illegitimate.
   * The US Constitution Article 1, Section 8, Clause 5 explicitly states that
     "The Congress shall have Power... To coin Money, regulate the Value
     thereof, and of foreign Coin, and fix the Standard of Weights and
     Measures."
   * The US Constitution Article 1, Section 10, Clause 1 explicitly states that
     "No state shall ... coin Money; emit Bills of Credit; make any Thing but
     gold and silver Coin a Tender in Payment of Debts."
     * The first "greenback" paper dollar issued in 1862 was a bill of credit
       backed by the federal government's promise to pay the bearer gold or
       silver.
   * Neither Article 1, Section 8, Clase 5 nor Article 1, section 10, Clause 1
     of the US Constitution give the federal government the authority to stray
     from the "bimetalic" spirit of the U.S. Constitution.
   * Paper fiat dollar bills are not coins.
     * Paper fiat dollars started off as bills of credit for deposited gold.
     * Paper fiat dollars today are neither bills of credit nor gold/silver coins.
   * The Federal Reserve was unconstitutionally ratified in order to debase the
     people's money from the underlying gold and silver.
     * The Coinage Act of 1873 also known as the "Crime of 1873" eliminated the
       standard dollar from the list of coins that the U.S. Mint could issue,
       beginning the demonetization of silver in favor of gold with the new
       "trade dollar", and later the introduction of the gold dollar coin.
     * The U.S. fiat paper dollar became gold-backed with the passage of the
       Gold Standard Act on March 14, 1900, which established gold as the
       exclusive backing for the country’s paper currency. This meant that each
       dollar bill was convertible into a specific amount of gold.
     * The Bretton Woods system was an international monetary system established
       in 1944 that set rules for commercial and financial relations among the
       major industrial states pegging the dollar to gold.
       * It created created a system of fixed exchange rates, with the U.S.
         dollar pegged to gold and other currencies pegged to the dollar, and
         established the International Monetary Fund (IMF) and the
         World Bank "to promote economic stability and growth".
       * The system ended in 1971 when the U.S. abandoned the gold standard,
         leading to a shift to floating exchange rates.
       * The Bretton Woods system ended in 1971 primarily because the U.S. could no
         longer maintain the dollar's convertibility to gold due to rising
         inflation and a growing balance of payments deficit, leading to a loss of
         confidence in the dollar. President Nixon's decision to suspend gold
         convertibility on August 15, 1971, effectively marked the collapse of the
         system, transitioning the world to floating exchange rates.
     * The Coinage Act of 1963 removed silver from dimes and quarters and
       reduced the silver content of of half-dollars to 40%, more or less
       completing the destruction silver coin as money and ending bimetalism.
       * [Wikpedia incorrectly states](https://en.wikipedia.org/wiki/Coinage_Act_of_1873) that bimetalism ended with the Coinage Act of 1873, and [prevents](https://en.wikipedia.org/wiki/Talk:Coinage_Act_of_1873) users from correcting the record.
       * By 1973, the U.S. dollar was fully decoupled from gold, transitioning to
         a fiat currency not backed by physical commodities.
     * JP Morgan intentionally sank the Titanic to murder opposition such as
       Straus and Astor, specifically to debase the dollar and to steal the works
       of Nikolas Tesla.
       * JP Morgan sank the Titanic to debase the dollar.
         * John Jacob Astor IV was the world's richest man; he opposed the
           Treasury and WWI.
         * Isidor Straus was the elected Treasurer of the New York branch of the
           National Citizen's League for the promotion of a Sound Banking System
           who corresponded with the editor of the New York Times and [made a call
           to action in and around 10/16/1911](./images/manifesto/isador_straus.jpeg) of the public for open discussion to prevent the adoption of an unaccountable federal reserve act.
         * Less than one year after Isador Straus declared the call to action to
           the public, the Titanic sunk in April 15, 1912 after boarding members
           invited by JP Morgan who himself dipped out of the party at the last
           minute--knowing that there was an engine room fire as logged by logs.
           * The Titanic sunk when the engine room exploded as intended; rescue
             ships that trailed the Titanic did not come to rescue upon seeing
             emergency SOS flairs allegedly because the red flares were swapped out
             for white flares that looked like fireworks.
             * The recent 3D scan of the sunken Titanic reveals an outward blowing
               out of the ship's hull where the engine room was, exactly where
               photographs of the Titanic showed fire damange during boarding.
               * The OceanGate submersible that later imploded on the way to view
                 the Titanic had in its board of directors a Rothschild.
         * The year following the sinking of the Titanic saw the unconstitutional
           passage of the Federal Reserve act in December 23, 1913.
             * The Federal Reserve Act was drafted in secret in Jeckyll Island and
               overseen by banking elites including the Rothschilds.
         * Less than one year after the passage of the Federal Reserve Act began
           World War I in July 28, 1914; and thus began the signficant dillution
           of the dollar via the sale of government war bonds.
       * JPMorgan sank the Titanic to steal the works of Nikolas Tesla.
         * John Jacob Astor IV was Tesla's primary patron.
         * Tesla died on 7 January 1943, at the age of 86, penniless.
         * Two days later his death the FBI ordered the Alien Property
           Custodian to seize Tesla’s belongings, even though Tesla was an
           American citizen.
         * Tesla’s entire estate from the Hotel New Yorker and other New York
           City hotels was transported to the Manhattan Storage and Warehouse
           Company under the Office of Alien Property (OAP) seal.
         * John G. Trump, a professor at M.I.T. and a well-known electrical
           engineer serving as a technical aide to the National Defense
           Research Committee, was called in to analyze the Tesla items in OAP
           custody.
           * John G. Trump was the uncle of Donald J. Trump.
         * After a three-day investigation, Trump’s report concluded that there
           was nothing which would constitute a hazard in unfriendly hands,
           stating: “[Tesla’s] thoughts and efforts during at least the past 15
           years were primarily of a speculative, philosophical, and somewhat
           promotional character often concerned with the production and
           wireless transmission of power; but did not include new, sound,
           workable principles or methods for realizing such results”.
     * The silver dollar used to have 0.7734 troy ounces of silver.
       At the price of silver today at $84 per troy ounce, a silver dollar
       today would be worth $64, roughly representing about 6 halvings
       of the purchasing power of the dollar, or 1.56%.
     * The gold dollar used to weigh 1.672 grams of which 90% was silver.
       At the price of gold today at $4,590 per troy ounce, a gold dollar 
       today would be worth $244, roughly representing about 8 halvings
       of the purchasing power of the dollar, or 0.41%.
     * Given 0.7734 troy ounces for a silver dollar and 0.053 troy ounces
       for a gold dollar the ratio of the value of silver to gold
       used to be around 1:14.6. Today the ratio is 1:54.6.
     * If the dollar were to be backed by gold and silver as it should,
       the price of gold and silver would be significantly higher.
     * With the adoption of silver solid-state battery technology and the
       greater need for silver in industry the ratio of the value of silver to
       gold will revert to the historical norm or even higher.

_All empires fall after debasing its currency. The Roman Empire took centuries
to debase their currency to 0.5%. The United States has achieved the same in
about half a century--it has only been 55 years since President Nixon ended the
convertibility of the dollar to gold. For a modern comparison, the British
Empire ended soon after WWI after debasing the silver shilling starting from
the 1920's until 1947 when silver content was completely eliminated. See the
appendix on [The History of U.S. Silver Coins](#the-history-of-u-s-silver-coins) for more history of U.S.
Silver Coins_.

The above tree of thought-statements exist to prove one primary thesis (namely
that "the Federal Reserve and the fiat dollar is unconstitional and
illegitimate"). There are so many other statements to describe all that is
wrong with this world and all that has been hidden from general public
consciousness; each need to be similarly fleshed out into their own subtrees or
branches of supporting thought statements. Each node may be referenced by many
other nodes, so what we need is a graph of thought statements that can be
permissionlessly and interatively expanded by anyone.

See [Use Case: Open Programmable Knowledge Base](#use-case-open-programmable-knowledge-base) for more on how
Gno and Gno.land are ideal for solving this problem.

## Gno Language

_I could not have arrived at the final design of the Gno language without the
help of the many developers who gave much time of their lives to contribute to
this project and the design of the language; specifically the NewTendermint Gno
team and Gno community, and especially Manfred, Morgan, Milos, Maxwell,
Guilhem, Ray, and Omar._

**Gno is the first _multi-user_ programming language.**

Gno is a minimal extension of the Go language for multi-user programming. Gno
allows a massive number of programmers to iteratively and interactively develop
a single shared program such as Gno.land. In other words, Go is a restricted
subset of the Gno language in the single-user context.

All of our programming languages to date are designed for a single programmer
user. All programming languages make the same assumption that there is only one
user -- the programmer, or program executor user. Whether written in C, C++,
Python, Java, Javascript, or Go, it is assumed that all of the dependencies of
the program are trusted. If there is a vulnerability in any of the dependencies
there is a vulnerability in the program; it is the job of the programmer or
program/product manager to ensure that the overall program is free of exploits.

When interacting with programs owned by another user (or process) various
techniques are used such as via IPC APIs often generated by tools like
Protobuf/GRPC; but such tools add extra complexity, additional surface area for
exploits, additional compute complexity, and do not benefit directly from the
language's native rules and type-checker--especially for inter-process passing
of in-memory object references.

The GnoVM is implemented purely in Go and is a Gno AST interpreter. It is
implemented to serve as an reference implementation suitable for production and
tinkering. The GnoVM opcodes are expected to change over time to balance
simplicity, efficiency, and transpilation targets--perhaps even to hardware.
Just as Ethereum was originally implemented in Python we expect future software
to be optimized, whether implemented in other languages, or possibly transpiled
from the existing GnoVM implementation. 

For the latest Gno specs including the inter-realm spec refer to the [Gno
docs/resources directory](https://github.com/gnolang/gno/tree/master/docs/resources).

### Gno vs Existing Smart Contract Platforms

Smart contract platforms like Ethereum allows for many users to upload their
application and call other user application logic functions, but Solidity is
not a general purpose programming language and has severe limitations that make
it suboptimal for solving the task at hand.

First, **Solidity and other existing smart-contract languages/platforms do not
allow support transparent inter-module (inter-user) interoperability with the
same language rules as for intra-module dependencies.** That is, an application
developer for a smart-contract cannot simply import and call/use another user
application's functions and types as if they were library dependencies of the
same application. Generally interoperability between different modules are
implemented with extra-language frameworks and libraries on top of an
incomplete or primitive message-passing agent architecture; such interop
function calls generally do not share the same call-stack nor memory space.

Second, **Solidity and other existing smart-contract languages/platforms do not
support the automatic persistence and Merkle-ization of in-memory (heap)
objects and often require custom serialization logic.** Solidity does not
support such a heap at all as all memory for variables are predeclared in the
function and as such is not object-oriented and does not have a garbage
collector or similar memory-management primitives. WASM-based smart contract
systems do not support automatic persistence of objects without persisting the
entire memory state of the module. This requires a specialized virtual machine
such as the Gno VM which keeps track of every object created, modified, and
deleted.

The automatic persistence of in-memory objects of the GnoVM is like a memristor
simulator. The advent of AI has created a new market for memristor-based memory
systems where the distinction between RAM volatile memory and persistent disk
storage is removed. Urbit is similar but is not based on any general purpose
programming language. With memristor-based memory the GnoVM can be further
simplified and the performance of applications can be vastly improved without
any changes to the Gno langauge specification.

Third, **Solidity and other existing smart-contract languages/platforms
do not support a shared heap memory space for objects to be referenced by
external-user objects in a uniform manner by language rules**. Alice cannot
simply declare a structure object that references the structure object
persisted in Bob's application and trust the garbage collector to retain Bob's
object for as long as Alice's object is retained.

**The above differentiating factors of the Gno language allows for the most
succinct expression of a single-user application or multi-user application
composed of independent modules without the extra complexity from
extra-language interop type-checking syntax or frameworks nor of the extra
complexity from any database, ORM, or serialization logic.**

Shared garbage-collection in a shared (multi-user) graph of object references
makes it possible for one's object representing (say) a propositional statement
or idea to be easily referenced by an alternative statement or idea, or even be
extended by reference with additional commentary, metadata, or even a
subreddit-like tree of discussions. Without a shared garbage collector the task
of ensuring that references still hold over time without becoming dangling
pointers is left up to each inter-application interface at best, requiring
custom logic just to handle garbage collection. WebAssembly (WASM) externref
support in Go has limitations, particularly in how it handles external memory
references. Currently, Rust and Go do not natively support externref types for
function parameters or return values, making it challenging to pass complex
data between Wasm modules and their host environments effectively.

> Reference type (aka externref or anyref) is an opaque reference made
> available to a WASM module by the host environment. Such references cannot be
> forged in the WASM code and can be associated with arbitrary host data, thus
> making them a good alternative to ad-hoc handles (e.g., numeric ones).
> References cannot be stored in WASM linear memory; they are confined to the
> stack and tables with externref elements.
>
> Rust does not support reference types natively; there is no way to produce an
> import / export that has externref as an argument or a return type.
> wasm-bindgen patches WASM if externrefs are enabled. This library strives to
> accomplish the same goal for generic low-level WASM ABIs (wasm-bindgen is
> specialized for browser hosts).
> 
> **externref use cases**
> Since externrefs are completely opaque from the module perspective, the only
> way to use them is to send an externref back to the host as an argument of an
> imported function. (Depending on the function semantics, the call may or may
> not consume the externref and may or may not modify the underlying data; this
> is not reflected by the WASM function signature.) An externref cannot be
> dereferenced by the module, thus, the module cannot directly access or modify
> the data behind the reference. Indeed, the module cannot even be sure which
> kind of data is being referenced.
> - https://docs.rs/externref/latest/externref/

Even if externref were fully implemented in future specs for Go (or Rust) such
that it could be used as an argument or return type across modules (still not
ideal for type-checking as it is not the underlying type), this would limit
what can be inter-module-referenced to that which can be held in memory. The
Gno Virtual Machine (GnoVM) allows for inter-user-package (inter-realm)
references across the entire persisted disk store space, and does not require
any additional language syntax such as with the `externref` keyword, and
supports the normal course of type-checking already familiar to Go developers.

### Interrealm Programming Context

Gno.land supports three types of packages:
- **Realms (`/r/`)**: Stateful user applications (smart contracts) that
  maintain persistent state between transactions
- **Pure Packages (`/p/`)**: Stateless libraries that provide reusable 
  functionality
- **Ephemeral Packages (`/e/`)**: Temporary code execution with MsgRun
  which allows a custom main() function to be run instead of a single
  function call as with MsgExec.

For an overview of the different package types in Gno (`/p/`, `/r/`, and 
`/e/`), see [Anatomy of a Gno Package](./builders/anatomy-of-a-gno-package.md).

Interrealm programming refers to the ability of one realm to call functions 
in another realm. This can occur between:
- Regular realms (`/r/`) calling other regular realms via MsgExec and MsgRun.
- Ephemeral realms (`/e/`) calling regular realms via MsgRun (like main.go)

The key concept is that code executing in one realm context can interact with
and call functions in other realms while leverage the language syntax rules of
Go, enabling complex multi-user interactions while maintaining clear boundaries
and permissions.

```go
// realm /r/alice/alice
package alice

var object any

func SetObject(cur realm, obj any) {
    object = obj
}
```

```go
// package /p/bob/types
package types

type UserProfile struct {
    Name string
    ...
}
```

```go
// realm /r/bob/bob
package bob

import "gno.land/r/alice/alice" // import external realm package
import "gno.land/p/bob/types"   // import external library package

func Register(cur realm, name string) {
    prof := types.UserProfile{Name: name}
    alice.SetObject(cross, prof)
}
```

All logic in Gno execute under a current realm-context and
realm-storage-context. The realm-context and realm-storage-context refer to the
same realm after a crossing-call of a function or method, but they may diverge
when calling a non-crossing method of a real object residing in a different
realm than the current realm-context. More on this later.

The Gno language is extended to support a `context.Context`-like argument to
denote the current realm-context of a Gno function. This allows a user realm
function to call itself safely as if it were being called by an external user,
and helps avoid a class of security issues that would otherwise exist.

```go
// realm /r/alice/mail

func SendMail(cur realm, text string) {
    if text == "" {
        // runtime.PreviousRealm() is preserved for recursive call.
        SendMail(nil, "<empty>")
    }
    caller := runtime.PreviousRealm()
    if inBlacklist(caller) {
        // runtime.PreviousRealm() becomes self; message from self to self.
        SendMail(cross, fmt.Sprintf("blacklisted caller %v blocked", caller))
    } else {
        // sendMailPrivate not exposed to external callers.
        sendMailPrivate(text)
    }
}
```

### Realm-Storage Write Access

Every object in Gno is persisted in disk with additional metadata including the
object ID and an optional OwnerID (if persisted with a ref-count of exactly 1).
The object ID is only set at the end of a realm-transaction during
realm-transaction finalization (more on that later). A GnoVM transaction is
composed of one or many scoped (stacked) realm-transactions.

```go
type ObjectInfo struct {
	ID       ObjectID  // set if real.
	Hash     ValueHash `json:",omitempty"` // zero if dirty.
	OwnerID  ObjectID  `json:",omitempty"` // parent in the ownership tree.
	ModTime  uint64    // time last updated.
	RefCount int       // for persistence. deleted/gc'd if 0.

	// Object has multiple references (refcount > 1) and is persisted separately
	IsEscaped bool `json:",omitempty"` // hash in iavl.
    ...
}
```

When an object is persisted during realm-transaction finalization the object
becomes "real" (as in it is really persisted in the virtual machine state) and
is said to "reside" in the realm; and otherwise is considered "unreal". New
objects instantiated during a transaction are always unreal; and during
finalization such objects are either discarded (transaction-level garbage
collected) or become persisted and real.

Unreal (new) objects that become referenced by a real (persisted) object at
runtime will get their OwnerID set to the parent object's storage realm, but
will not yet have its object ID set before realm-transaction finalization.
Subsequent references at runtime of such an unreal object by real objects
residing in other realms do not override the OwnerID intially set, so during
realm-transaction finalization it ends up residing in the first realm it became
associated with (referenced from). Unreal objects that become persisted but was
never directly referenced by any real object during runtime will only get its
OwnerID set to the realm of the first real ancestor.

Real objects with ref-count of 1 have their hash included in the sole parent
object's serialized byte form, thus an object tree of only ref-count 1
descendants are Merkle-hashed completely.

When a real or unreal object ends up with a ref-count of 2 or greater during
realm-transaction finalization its OwnerID is set to zero and the object is
considered to have "escaped". When such a real object is persisted with
ref-count of 2 or greater it is forever considered escaped even if its
ref-count is in later transactions is reduced to 1. Escaped real objects do not
have their hash included in the parent objects' serialized byte form but
instead are Merkle-ized separately in an iavl tree of escaped object hashes
(keyed by the escaped object's ID) for each realm package. (This is implemented
as a stub but not yet implemented for the initial release of Gno.land)

Go's language rules for value access through dot-selectors & index-expressions
are the same within the same realm, but exposed values through dot-selector &
index-expressions are tainted read-only when performed by external realm logic.

**A real object can only be directly mutated through dot-selectors and
index-expressions if the object resides in the same realm as the current
realm-storage-context. Unreal objects can always be directly mutated if its
elements are directly exposed.**

### Crossing-Functions and Crossing-Methods

A crossing-function or crossing-method is that which is declared in a realm and
has as its first argument `cur realm`. The `cur realm` argument must appear as
the first argument of a crossing-function or crossing-method's argument
parameter list. To prevent confusion it is illegal to use anywhere else, and
cannot be used in p packages. 

The current realm-context and realm-storage-context changes when a
crossing-function or crossing-method is called with the `cross` keyword in the
first argument as in `fn(cross, ...)`. Such a call is called a "cross-call" or
"crossing-call".

```go
package main
import "gno.land/r/alice/extrealm"

func MyMakeBread(cur realm, ingredients ...any) { ... }

func main(cur realm) {
    MyMakeBread(cross, "flour", "water") // ok -- cross into self.
    extrealm.MakeBread(cross, "flour", "water") // ok -- cross into extrealm
}
```

(In Linux/Unix operating systems user processes can cross-call into the kernel
by calling special syscall functions, but user processes cannot directly
cross-call into other users' processes. This makes the GnoVM a more complete
multi-user operating system than traditional operating systems.)

When a crossing-function or crossing-method is called with `nil` as the first
argument instead of `cross` it is called a non-crossing-call; and no
realm-context nor realm-storage-context changes takes place.

```go
package main
import "gno.land/r/alice/extrealm"

func MyMakeBread(cur realm, ingredients ...any) { ... }

func main(cur realm) {
    MyMakeBread(nil, "flour", "water") // ok -- non-crossing.
    extrealm.MakeBread(nil, "flour", "water") // invalid -- external realm function
}
```

To prevent confusion a non-crossing-call of a crossing-function or
crossing-method declared in a realm different than that of the caller's
realm-context and realm-storage-context will result in either a type-check
error; or a runtime error if the crossing-function or crossing-method is
variable.

`runtime.CurrentRealm()` returns the current realm-context that was last
cross-called to. `runtime.PreviousRealm()` returns the realm-context cross-called
to before the last cross-call. All cross-calls are explicit with the `cross`
keyword, as well as non-crossing-calls of crossing-functions and
crossing-methods with `nil` instead of `cross`.

Besides (explicit) realm-context changes via the `fn(cross, ...)` cross-call
syntax, implicit realm-storage-context changes occurs when calling a
non-crossing method of a receiver object residing in different realm-storage.
This change in realm-storage-context allows any non-crossing method to directly
modify its receiver (and also any objects directly reachable and residing in
the same realm-storage) without changing the realm-context (so
`runtime.CurrentRealm()` and `runtime.PreviousRealm()` do not change; the
agency of the caller remains the same). This allows non-crossing methods of
receivers to behave the same whether declared in a realm package or p package
such that p package code copied over to a realm package r realm package code
copied over to another realm have the exact same behavior. Crossing methods of
a realm package would still behave differently when copied over to another
realm as crossing-methods always change the realm-context and
realm-storage-context to the declared realm.

If the receiver resides in realm-storage that differs from the caller's
realm-storage-context such a receiver's non-crossing method cannot directly
modify the receiver (nor any reachable object that resides in any realm-storage
besides that of the caller's own realm-storage-context). 

On the other hand if the method is a crossing-method as in
`receiver.Method(cross, args...)` and the method is cross-called both the
realm-context and realm-storage-context changes to that of the realm package in
which the type/method is declared (which is not necessarily the same as where
the receiver resides). Such a crossing method-call cannot directly modify the
real receiver if it happens to reside in an external realm that differs from
where the type and methods are declared; but it can modify any unreal receiver
or unreal reachable objects. As mentioned previously a non-crossing-call of a
crossing-method will fail at during type-checking or at runtime if the receiver
resides in an external realm-storage.

MsgCall can only call crossing-functions.

MsgRun will run a file's `main()` function in the user's realm-context and may
call both crossing and non-crossing functions and methods.

A realm package's initialization (including `init()` calls) executes with
current realm-context of itself, and its `runtime.PreviousRealm()` will panic
unless the call stack includes a crossing function called like `fn(cross,
...)`.

### Realm Boundaries

The current and previous runtime realm-context have an associated Gno address
from which native coins can be sent from and received to. Such native coins can
only be sent from a banker instantiated with either realm-context.  The
realm-storage-context is not accessible at runtime and so there is no
associated Gno address.

When a crossing-function or crossing-method is cross-called it shifts the
"current" runtime realm-context to the "previous" runtime realm-context such
that `runtime.PreviousRealm()` returns what used to be returned with
`runtime.CurrentRealm()` before the realm boundary. The current
realm-storage-context is always set to that of realm-context after
cross-calling.

Every crossing-call of a crossing-function or crossing-method creates a new
realm boundary even when there is no resulting change/shift in realm-context or
realm-storage-context.

A realm boundary also exists for every call that results in a change of
realm-storage-context: whether with a crossing-call to another realm-context
and realm-storage context or with a (non-crossing) call of a non-crossing
method of a real receiver residing in another realm-storage than the current
realm-storage-context. No realm boundary occurs when calling a non-crossing
method of an unreal receiver or a non-crossing function.

A realm boundary does not always change the realm-context nor always change the
realm-storage-context. A crossing-call into the same realm-context never
changes the realm-context and may not change the realm-storage-context either;
a crossing-call into a different realm always changes the realm-context but may
not change the realm-storage-context; a (non-crossing) call of a method of a
real object residing in an external realm-storage never changes the
realm-context but changes the realm-storage-context. However, a
non-crossing-call of a crossing-function or crossing-method will never create a
realm boundary.

No realm boundary is created for non-crossing functions and non-crossing
methods of unreal receivers.

### Realm-Transaction Finalization

Realm-transaction finalization occurs when returning from a realm
boundary. When returning from a cross-call (with `cross`) realm-transaction
finalization will occur even with no change of realm-context or
realm-storage-context. Realm-transaction finalization does NOT occur when
returning from a non-crossing-call of a method of an unreal receiver or a real
receiver that resides in the same realm-storage-context as that of the caller.

During realm-transaction finalization all new reachable objects are assigned
object IDs and stored in the current realm-storage-context; and ref-count-zero
objects deleted (full "disk-persistent cycle GC" will come after launch); and
any modified ref-counts and new Merkle hash root computed. 

### Readonly Taint Specification

Go's language rules for value access through dot-selectors & index-expressions
are the same within the same realm, but exposed values through dot-selector &
index-expressions are tainted read-only when performed by an external realm.

The readonly taint prevents the direct modification of real objects by any
logic, even from logic declared in the same realm as that of the object's
storage-realm.

A realm cannot directly modify another realm's objects without calling a
function that gives permission for the modification to occur.

For example `externalrealm.Foo` is a dot-selector expression on an external
object (package) so the value is tainted with the `N_Readonly` attribute.

The same is true for `externalobject.FieldA` where `externalobject` resides in
an external realm.

The same is true for `externalobject[0]`: direct index expressions also taint
the resulting value with the `N_Readonly` attribute. 

The same is true for `externalobject.FieldA.FieldB[0]`: the readonly taint
persists for any subsequent direct access, so even if FieldA or FieldB resided in
the caller's own realm-context or realm-storage the result is tainted readonly.

A Gno package's global variables even when exposed (e.g. `package realm1; var
MyGlobal int = 1`) are safe from external manipulation (e.g. `import
"xxx/realm1"; realm1.MyGlobal = 2`) by the readonly taint when accessed
directly by dot-selector or index-expression from external realm logic; and
also by a separate `DidUpdate()` guard when accessed by other means such as by
return value of a function and the return value is real and external.

A function or method's arguments and return values retain and pass through any
readonly taint from caller to callee. Even if realm's function (or method)
returns an untainted real object, the runtime guard in `DidUpdate()` prevents
it from being modified by an external realm-storage-context.

For a realm (user) to manipulate an untainted object residing in an external
realm, a function (or method) can be declared in the external realm which
references and modifies the aforementioned untainted object directly (by a name
declared outside of the scope of said function or method). Or, the function can
take in as argument an untainted real object returned by another function.

Besides protecting against writing by direct access, the readonly taint also
helps prevent a class of security issue where a realm may be tricked into
modifying something that it otherwise would not want to modify. Since the
readonly taint prohibits mutations even from logic declared in the same realm,
it protects realms against mutating its own object that it doesn't intend to:
such as when a realm's real object is passed as an argument to a mutator
function where the object happens to match the type of the argument.

Objects returned from functions or methods are not readonly tainted. So if
`func (eo object) GetA() any { return eo.FieldA }` then `externalobject.GetA()`
returns an object that is not tainted assuming eo.FieldA was not otherwise
tainted. While the parent object `eo` is still protected from direct
modification by external realm logic, the returned object from `GetA()` can be
passed as an argument to logic declared in the residing realm of `eo.FieldA`
for direct mutation.

Whether or not an object is readonly tainted it can always be mutated by a
method declared on the receiver.

```go
// /r/alice

var blacklist []string

func GetBlacklist() []string {
    return blacklist
}

func FilterList(cur realm, testlist []string) { // blanks out blacklist items from testlist
    for i, item := range testlist {
        if contains(blacklist, item) {
            testlist[i] = ""
        }
    }
}
```

This is a toy example, but you can see that the intent of `FilterList()` is to
modify an externally provided slice; yet if you call `alice.FilterList(cross,
alice.GetBlacklist())` you can trick alice into modifying its own blacklist--the
result is that alice.BlackList becomes full of blank values.

With the readonly taint `var Blacklist []string` solves the problem for you;
that is, /r/bob cannot successfully call `alice.FilterList(cross,
alice.Blacklist)` because `alice.Blacklist` is readonly tainted for bob.

The problem remains if alice implements `func GetBlacklist() []string { return
Blacklist }` since then /r/bob can call `alice.FilterList(cross,
alice.GetBlacklist())` and the argument is not readonly tainted.

Future versions of Gno may also expose a new modifier keyword `readonly` to
allow for return values of functions to be tainted as readonly. Then with `func
GetBlacklist() readonly []string` the return value would be readonly tainted
for both bob and alice.

### Interrealm Specfication Design Goals

**Caveat: The interrealm specification does not secure applications against
arbitrary code execution. It is important for realm logic (and even p package
logic) to ensure that arbitrary (variable) functions (and similarly arbitrary
interface methods) are not provided by malicious callers; such arbitrary
functions and methods whether crossing (or non-crossing) will inherit the
previous realm (or both current and previous realms) and could abuse these
realm-contexts.** It does not make sense for any realm user to cross-call an
arbitrary function or method as it loses agency while being marked as the
responsible caller by the callee's runtime previous realm. This problem is
worse when calling a non-crossing function or method. It can be reasonable when
such variable functions or interface values are restricted in other ways such
as by whitelisting by a DAO upon careful inspection of every such variable
function or interface value (both its type declaration as well as its state).

P package code should behave the same even when copied verbatim in a realm
package; and likewise non-crossing code should behave the same when copied
verbatim from one realm to another. Otherwise there will be lots of security
related bugs from user error.

Realm crossing with respect to `runtime.CurrentRealm()` and
`runtime.PreviousRealm()` must be explicit and warrants type-checking; because
a crossing-function of a realm should be able to call another crossing-function
of the same realm without necessarily crossing (changing the realm-context).
Sometimes the previous realm and current realm must be the same realm, such as
when a realm consumes a service that it offers to external realms and users.

Where a real object resides should not matter too much, as it is often
difficult to predict. Thus the realm-context as returned by
`runtime.PreviousRealm()` and `runtime.CurrentRealm()` should not change with
non-crossing method calls, and the realm-storage-context should be determined
for non-crossing methods only by the realm-storage of the receiver. The
realm-storage of a receiver should only matter for when elements reside in
external realm-storage and direct dot-selector or index-expression access of
sub-elements are desired of the aforementioned element.

A method should be able to modify the receiver and associated objects of the
same realm-storage as that of the receiver.

A method should be able to create new objects that reside in the same realm by
association in order to maintain storage realm consistency and encapsulation
and reduce fragmentation.

It is difficult to migrate an object from one realm to another even when its
ref-count is 1; such an object may be deep with many descendants of ref-count 1
and so performance is unpredictable.

Code declared in p packages (or declared in "immutable" realm packages) can
help different realms enforce contracts trustlessly, even those that involve
the caller's current realm. Otherwise two mutable (upgradeable) realms cannot
export trust unto the chain because functions declared in those two realms can
be upgraded.

Both `fn(cross, ...)` and `func fn(cur realm, ...){...}` may become special
syntax in future Gno versions.

### `panic()` and `revive(fn)`

`panic()` behaves the same within the same realm boundary, but when a panic
crosses a realm boundary (as defined in [Realm Boundries](#realm-boundaries))
the Machine aborts the program. This is because in a multi-user environment it
isn't safe to let the caller recover from realm panics that often leave the
state in an invalid state.

This would be sufficient, but we also want to write our tests to be able
to detect such aborts and make assertions. For this reason Gno provides
the `revive(fn)` builtin.

```go
abort := revive(func() {
    cross(func(_ realm) {
        panic("cross-realm panic")
    })
})
abort == "cross-realm panic"
```

`revive(fn)` will execute 'fn' and return the exception that crossed a realm
boundary during finalization.

This is only enabled in testing mode (for now), behavior is only partially
implemented. In the future `revive(fn)` will be available for non-testing code,
and the behavior will change such that `fn()` is run in transactional
(cache-wrapped) memory context and any mutations discarded if and only if there
was an abort.

TL;DR: `revive(fn)` is Gno's builtin for STM (software transactional memory).

### `attach()`

In future releases of Gno the `attach()` function can be used to associate
unreal objects to the current realm-storage-context before being passed into
function declared in an external realm package, or into a method of a real
receiver residing in an exteral realm-context.

### `safely(cb func())`

In future releases of Gno the `safely(cb func())` function may be used to clear
the current and previous realm-context as well as any realm-storage-context
such that no matter what `cb func()` does the caller does not yield agency to
the callee.

For now this can be simulated by implementing an (immutable non-upgradeable)
realm crossing-function that cross-calls into itself once more before calling
the callback function.

XXX Ensure that both `attach` and `safely` are reserved keywords for the
preprocessor.

## Gno.land Blockchain

Tendermint solved proof-of-stake by innovating upon classical Byzantine
fault-tolerant consensus algorithms published by Dword, Lynch, and Stockmeyer
in 1988 (originally funded by Darpa for missile defense systems) for
blockchains. It paved the way for the Cosmos Hub, the first proof-of-stake IBC
hub, and Cosmos the internet of blockchains. Also of note, when Binance first
launched they used the CosmosSDK and Tendermint.

Gno.land builds upon Tendermint2 and aims to shift the paradigm of programming
languages in general: Gno is the first _multi-user_ programming language, making
it a superior smart contracting language as compared to any existing solution.
Thus **Gno.land is the first multi-user langauge-based operating system**. Its
ultimate goal is to be the world's open knowledge base for the next millenium.

### Use Case: Open Programmable Knowledge Base

Go's simple embedded struct-centric design and the Gno VM's automatic
transactional persistence makes Gno.land not only great for decentralized
financial applications but also makes it uniquely well suited and designed for
permissionless innovation of information-based applications such as social
communication and coordination systems, or the next Wikipedia or programmable
knowledge-base systems. The latter will be explored here.

Each of the thought statements in the [introduction](#gno-land-for-mass-awakening) can be represented as a
simple Go string, but as in Tractatus we want to allow each of these thought-
statements to be supported by any number of supporting thought statements, so
we need a struct declaration.

```go
type Thought struct {
    Statement    string
    Dependencies []*Thought
}
```

The above allows for a simple tree structure, but it would be better to
annotate each child node (thought statement) with the type of relation to the
parent node-- for example whether a child node represents an example, a caveat,
a corrolary, or supporting evidence and so on.

```go
// Option "Denormalized Thought"
type Thought struct {
    Statement   string
    Examples    []*Thought
    Caveats     []*Thought
    Correlaries []*Thought
    Support     []*Thought
}
```

Better than a denormalized structure is one where the type of thought statement
association can be arbitrary or fixed depending on the application.

```go
// Option "Normalized Thought"
type Thought struct {
    Text         string
    TypedSupport []*Thought
}

type ThoughtType string // examples, caveats, corrolaries, support

type TypedThought struct {
    Type    ThoughtType
    Thought *Thought
}
```

_Note on the usage of `[]\*Thought` slices: in the current implementation of the
GnoVM each slice can only be used by first loading the entire underlying base
array. This may be optimized in the future, however for holding large sets of
elements the programmer should instead use a tree-structure such as the
avl.Tree (or an iavl.Tree)._

Now arises the question of whether counter-arguments should also be referenced
as a child node to the original thought parent node. If we include
counter-arguments in the graph of `*Thought` objects itself there is the issue
of permissioning who can add counter-arguments to the graph. With the examples
above and with no method declarations a `*Thought` belonging to one user
cannot be modified by a third party even though the fields of a `Thought`
struct is exposed due to Gno (runtime) interrealm rules that taint third party
reads via direct dot-selectors & index-expressions with a readonly-taint that
persists even with (direct selector) access of sub-fields.

The `*Thought` object can however be modified by another user by calling a
declared method. We can extend the `Thought` struct with additional fields for
authorization or ownership and implement a method such as follows:

```go
type Thought struct {
    Owner        account
    Statement    string
    TypedSupport []*Thought
}

func (th *Thought) AddCounterArgument(cth *Thought) {
    caller := runtime.CurrentRealm().Address
    if th.Owner != caller {
        panic("unauthorized")
    }
    th.TypedSupport = append(th.TypedSupport,
        TypedThought{Type: "counter", Thought: cth})
}
```

This works but not well--it only if the owner of the parent node wants the
counter-argument to be registered. Even if counter-arguments were not
registered as an assocation on chain, it is still possible for any Gno.land
state indexer to separately index the reverse association of reference to the
original `*Thought` when it finds a counter-argument `*Thought` that
references in its struct field the original as a counter-argument. This
reliance on an external indexer shifts trust from the blockchain itself to the
indexer so is not always ideal.

Gno is intended for permissionless iteration and improvement so we will discuss
another way (among many) to manage associations of competing thought
statements; the pair-wise association among competing thought statements can be
registered in another (neutral) external realm that allows the registration
only at least one of the two thought statements identify the other as a
counter-argument. In this case it is not necessary for a `*Thought` object to
be associated with any owner explicitly (via the `.Owner` field). Note however
that given the Gno inter-realm specification to make a `*Thought` object truly
immutable even for the owner of the realm in which it resides it must not
expose any mutator functions, or it should have at least a `readonly bool`
field.

We can also add discussion board objects for each thought statement.

```go
// Board defines a type for boards.
type Board struct {
	// ID is the unique identifier of the board.
	ID ID
	// Name is the current name of the board.
	Name string
	// Aliases contains a list of alternative names for the board.
	Aliases []string
	// Readonly indicates that the board is readonly.
	Readonly bool
	// Threads contains all board threads.
	Threads PostStorage
	// ThreadsSequence generates sequential ID for new threads.
	ThreadsSequence IdentifierGenerator
	// Permissions enables support for permissioned boards.
	// This type of boards allows managing members with roles and permissions.
	// It also enables the implementation of permissioned execution of board related features.
	Permissions Permissions
	// Creator is the account address that created the board.
	Creator address
	// Meta allows storing board metadata.
	Meta any
	// CreatedAt is the board's creation time.
	CreatedAt time.Time
	// UpdatedAt is the board's update time.
	UpdatedAt time.Time
}

// New creates a new basic non permissioned board.
func New(id ID) *Board {
	return &Board{
		ID:              id,
		Threads:         NewPostStorage(),
		ThreadsSequence: NewIdentifierGenerator(),
		CreatedAt:       time.Now(),
	}
}
```

<img src="./images/manifesto/boards2.jpeg" />

While it is certainly possible to embed a `*Board` as a field of each
`*Thought`, the current implementation of `*Board` is only safe from a
moderation perspective when it is permissioned; and so a board tightly coupled
to a `*Thought` may not be ideal depending on the use-case. Instead we can map
an external realm persisted index of `*Thought` to `*Board` associations
similarly to how counter-thoughts are associated as mentioned before. In both
cases we probably want to add to the `Thought` struct a globally unique ID like
how `Board` has. _In the future we may provide a standard function to get a
unique identifier for every pointer object but this has not yet been decided
yet._

Finally, consider for example the numbered sequence of verses of a book of the
bible, or the deep tree of statements in Wittgenstein's Tractatus. In order to
faciliate the efficient forking of such large lists or graphs of objects it is
necessary to avoid the usage of slices. Even the avl.Tree (as provided in the
Gno monorepo under the examples directory) is not sufficient as it is a mutable
tree. However a fork of the avl.Tree into an immutable tree (or likewise a port
of the iavl tree in the tm2 Tendermint2 directory) can be used with some
improvement to allow for splicing in new elements and deleting existing
elements from the original tree.

So far I have illustrated a way for multiple users to construct their thought
statement graphs independently while also allowing for counter-arguments to be
registered/associated permissionlessly. More design and exploration is needed
to create a fully functional permissionlessly extensible thought statement
graph system; and in the primodrial soup of Gno ecosystem eventually one or
more designs will become dominant in usage by evolution. The reader is
encouraged to explore the above template and measure success by references and
by forks. _See [Use Case: 95 Fulcrums](#use-case-95-fulcrums) and [Use Case: Addressing Mistranslations of the Bible](#use-case-addressing-mistranslations-of-the-bible] for more ideas._

### Use Case: Home Computing Chains

The Gno VM is not just useful in the context of public decentralized
blockchains. It is also useful for home computing. Take for example Email which
despite all attempts to replace it still persists in our lives today as flawed,
complex, and outdated as it is. While a realm that stores mail on the
blockchain is not useful unless the data is encrypted _(and even if it were
encrypted it is not a good idea nor encouraged to store encrypted data on
gno.land as encryption keys may eventually get acquired by hackers and leaked
and even persisted on the blockchain too)_, the Gno VM can run anywhere, even
on your private server hosted at home. In fact, this is what we should do given
the prevalence of surveillance technology such as Google's Gmail which uses AI
to sort your mail and analyze for targeted advertising.

Imagine a black box local GnoVM you run at home. You can have the
`/r/home/email` realm store your emails at home on your own home server. The
same blockchain node logic can run on its own as a single-validator _home
chain_ which naturally supports backups as seecondary full nodes, or you can
even make your home chain byzantine fault-tolerant for better uptime.

 1. Install in your home GnoVM chain a service plugin: /s/email/indexer, not
    /p/\* nor /r/\* but /s/\* for off-chain service applications. _(this prefix
    is not supported in gno.land but may be in the future)_.

 * /s/email/indexer reads state upon init, but also registers as a listener
   for notifications from /r/emails.

 * When a new email comes in, /r/emails via listeners calls
   `/s/email/indexer.AddEmail()`.

 * /s/email/indexer also imports /d/email/indexer which is an off-chain daemon
   component. Here /d/\* represents a hyptoethical prefix for Gno code to be
   run off-chain with arbitrary Go native functions available for import that
   would otherwise not be possible on gno.land (since a blockchain can only
   support deterministic logic).

 * /d/email/indexer can only access /s/email/indexer by a Gno firewall system
   declared with Gno package paths, types, and function/method names.

 * /s/email/indexer can import any /r/\* or /p/\* but not any /s/\* (like
   Chrome extensions) and its own /d/email/indexer, unless otherwise restricted
   by the Gno firewall system.

 * Your mobile device registers an account with your local GnoVM home chain.
   This phone account is restricted to only access /s/email/indexer.

 * Your phone makes a request to /s/email/indexer. It then asks
   /d/email/indexer which in turn queries the local index and responds via
   /s/email/indexer.

Here are some benefits of GnoVM home computing:
 * Gno.land can be leveraged to ensure that all software is properly audited.
 * Software is expected to become finished and immutable.
 * All software benefits from Go/Gno's type-safety and memory-safety.
 * A unified IPC system drastically reduces surface area for penetration.
 * Plugin services and daemons such as the aforementioned email indexer can be
   containerized and restricted from unauthorized access.
 * Fine-grained security/firewall rules at the function invocation level.
 * Byzantine fault-tolerance comes out of the box for zero downtime.
 * Inversion of control with public key cryptography for everything: no more
   password management.

### Other Use Cases

Gno.land can be used to host any other smart-contract application supported by
Ethereum written in Solidity, such as Defi applications, name-resolution
systems, DAOs and governance applications, etc.

You can explore the various dapps including sample implementations of ERC equivalents in the
[examples directory](https://github.com/gnolang/gno/tree/master/examples/gno.land). _Note that these prototypes have not yet been audited unless
otherwise specified!_

### Gno.land Constitution

See [./CONSTITUTION.md] for the Gno.land Constitution (draft) and details of
genesis, tokenomics, governance, and more.

#### Separation of Church and State

Madison separated church and state in the US Constitution albiet there is a
hint of the Christian spirit by the way in which the constitution was signed:
"... in the Year of the Lord...". All the founders were Christian including
Jefferson and Madison, and in particular the primary author of the US
Constitution James Madison explicitly separated church from the constitution so
as to help promote the teachings of Jesus as evidenced in his other writings.
Likewise Gno.land besides this whitepaper is independent of any religion by its
constitution, which should only refer to this whitepaper sparingly.

Gno.land will launch with a minimal (living) constitution written and
maintained in English, but also ultimately be supplemented by the completed
GnoVM code and Tendermint2 and Gno.land implementation. Future implementations
of the GnoVM and Gno.land should adhere to the completed software mentioned
above.

Gno.land should not censor speech, even if the speech is wrong. However, it
should ban all porn and try to limit external links to porn sites as porn is
not speech and is dangerous to civilization. Whether hate-speech is tolerated
shoud be determined by each realm but also by the living Gno.land constitution
and by GovDAO vote to amend the constitution and laws of Gno.land.

### GnoWeb Browser

GnoWeb is the server software for Gno.land, a browser within a browser for
viewing realm data.

Instead of requiring realm applications to return HTML, the convention is to
implement a Render() function that returns Markdown. This is to allow the
transition away from the bloated HTML standards and browser software and realm
data to be browsed even from the console. Note that we don't need HTML XML
elements to denote objects: everything in Gno is already an object. This makes
Gno.land more like the original World Wide Web that conforms to the *Document
Object* model (DOM).

GnoWeb does support some custom XML elements for improved layout and
functionality, such as column layout and form submissions that integrate with
browser extensions for transaction signing.

Note that GnoWeb is not yet a general blockchain explorer (e.g. for
transactions) nor a general purpose Gno.land state explorer. A Gno.land
specific blockchain explorer already exists. GnoWeb can only render markdown
returned from Render() functions, and a general purpose state explorer is still
desired.

Realm code is not precluded from returning HTML or even JSON for custom browser
applications. In the near future the Gno.land node software will support
returning JSON encodings of Gno objects. Thus future alternative browser
applications may provide more interactive rich user experiences for viewing and
mutating Gno.land state without any Markdown intermediary representation; and
perhaps leveraging AI for intelligent layout and styling for rich
interactivity.

## Future Work

 * Name registry; Immutable names.
 * Realm upgrading.
 * Realm data browser.
 * Deterministic concurrency.
 * Joeson parser.
 * Gno2.
 * Open hardware.

## Summary

Gno is the next C.
Gno.land is the next Ethereum and Wikipedia.
AI will be safer.
Politics will be more honest.

## Appendix

### Use Case: 95 Fulcrums

Consider the following thought statements:

 * There exists at this moment a US-based global surveillance system headed by
   Palentir which uses advanced AI to intercept, mutate, and censor sensitive
   information from reaching public consciousness.
   * Google, Wikipedia, Meta, Reddit, Twitter, etc all participate in the above
     mentioned censorship.
     * Twitter's Community Notes system is fair in writing but in practice top
       contributors routinely publish write biased notes; and [you get banned
       for pointing it out](#twitter-censorship).
   * ...

 * At least some of the Rothschilds are litereally satanic; and in some parts
   of Germany there has taken root satanism that includes elements of ritual
   child sacrifice.
   * The Pergamom Altar, one of the seven Wonders of the Ancient World was
     moved to Berlin, Germany since 1886.
     * The Pergamom Altar is the "Seat of Satan" mentioned by John the Apostle in
       the Book of Revelation.

 * The elites wish for global depopulation.
   * Prince Phillip, Duke of Edinburgh, husband of Queen Elizabeth II said: "If
     I were reincarnated, I would wish to be returned to Earth as a killer
     virus to lower human population levels." in 1988.
     * Prince Phillip passed away on April 9, 2021, at the age of 99.

 * Fauci should be in jail for illegaly aiding in the development of the
   Covid19 virus and lying to the US Congress about it under oath.
   * Covid19 was engineered in the Wuhan lab for global depopulation.
   * [The collective world-wide evidence from 2020– 2025 underscores a
     biologically plausible connection between COVID-19 vaccination and
     ("turnbo") cancer](https://drive.google.com/file/d/1RH0f1BEqY9I7ruuPKQPU6bGNVJRyxZxc/view).
     * This information was censored by a [cyberattack hitting a major cancer journal on Jan 4th, 2025](https://x.com/NicHulscher/status/2007939234322174444?s=20).

 * After [Senator John Kennedy](https://www.youtube.com/watch?v=0Dk1ft7Vn9c)
   and others spoke out about the Somali wellfare scams, Nick Shirley exposed a
   massive Somali day care center scam in Minessota in a viral video which
   reached more views than all US network television combined, and yet NO
   mainstream media covered it (as of Dec 30th, 4 days after Nick Shirley's
   expose).

 * Somali House of Representatives Rep. Ilham Omar [was caught participating
   in the fraud](https://x.com/EricLDaugh/status/2009405199991853210?s=20). 

 * Tim Walz, U.S. House of Representatives, and Democratic nominee for vice-
   president 2024, knew of the Somali day care scam.
   * Walz's own appointee murdered the only lawmaker who voted against the
     Somali day care scam.

 * The Trump administration broke the law by redacting and selectively
   publishing the Epstein Files (and even deleted files after publishing this
   Dec 2025); and simultaneoulsy broke international law by stealing oil from
   Venezuela and murdering Venezuelan citizens to arrest Venezuelan president
   on false pretenses of drug production/smuggling; partially to distract from
   his own connections to Epstein and the exposure of the Somali day care scam
   given their concurrent timing).

 * [Epstein was murdered](https://github.com/jaekwon/ephesus/blob/main/epstein/homicide/README.md)
   and there exists a vast conspiracy to "poison the well" with theories that
   Epstein is still alive.

 * [The State of Israel's Mossad is behind the Somali inter-state day care scams](https://www.youtube.com/watch?v=SSKn9bHEvOw); and this is why Israel
   suddenly recognized the new "Somaliland"; the Palestinians who survive the
   genocide will likely be relocated there.

 * Zohran Mamdani’s new NYC tenant advocate called to 'Seize private
   property!', blasted home ownership as ‘white supremacy’.
   * Trump should be arresting Zohran Mamdani and Cea Weaver instead as per the
     Communist Control Act of 1954.
     * Nothing will happen.

 * The easiest way for the State of Israel to expand to "Greater Israel" is to
   extract capital from the United States by turning it into a communist
   dictatorship given current economic, political, and sociological condition.

 * Communism has never succeeded except arguably in North Korea which is a
   dictatorship different than prior communist states that cannot and should
   not be replicated. The Old Testament assumes private land ownership that
   cannot even be sold (returned every 50 years in the jubilee).
   * Since the Roman times and throughout the Early Middle Ages where England
     came under rule of post-Roman chieftains and Anglo-Saxon monarchs land was
     the dominant source of personal wealth.

 * We need to create a new system that represents the community of *productive*
   people who wish to form an alternative system that does respect private
   property, the rule of law, including the law of the biblical jubilee to help
   guarantee that people do not get displaced from their land by temporary
   political circumstances.
   * In time this community will overcome all other alternatives.

 * The spirit of the US Constitution is Christian.
   * All of the founders were Christian.
   * James Madison separated church and state for the express purpose of
     promoting the teachings of Jesus.

 * Gno.land poses the greatest threat to the aforementioned "Beast AI Prison
   Matrix" system which seeks to censor the truth.

 * The New Testament had been mistranslated to hide one of the primary reason
   why Jesus was crucified--he was in a "silent protest" of taxes.

 * The dollar and most fiat currencies will hyperinflate before 2030.
 
 * The [Federal Reserve appropriated America's gold in Fort Knox](#fort-knox-gold).

 * Silver will replace gold and the dollar as the primary store of value and
   become once again the primary means of exchange; local crypto currencies
   will follow; and interoperate by IBC.

 * In the long future most electric vehicles will use silver solid-state
   batteries.

 * The Grand Solar Minimum will create a deficit in global food production for
   decades and we are at risk of a global Holodomor.
   * Bill Gates is the largest private farmland owner with over 260,000 acres.

 * Stocking up on organic grains and organic legumes and oil and silver is the
   best way for a nation's people to defend against a tyrannical government and
   mitigate excess immigration, and to survive [the coming Grand Solar Minimum](https://www.sott.net/article/440781-NOAA-confirms-a-full-blown-Grand-Solar-Minimum).
   * The Maunder Minimum (1645-1715) brought plummeting temperatures, crop
     loss, famine, and the deaths of hundreds of millions of people ACROSS the
     planet.
   * The Dalton Minimum (1790-1830) coincided with significant agricultural
     distress, particularly during the Year Without a Summer in 1816. This year
     saw severe weather anomalies, including frosts and snow in summer months,
     leading to widespread crop failures and food shortages across the Northern
     Hemisphere.
   * Grains and legumes can be stored for up to 30 years.

 * Glyphosate in grains is a significant contributor to cancer; farmers may be
   pressured by weather and financial reasons to use glyphosate against the
   recommended directions to force an hearly harvest before cold weather.
   * Chlorination (e.g. with chlorine dioxide) or ozonation of water soaked
     grains with application of UV light of certain frequencies may neutralize
     the glyphosate in harvested food.

 * The "10 Lost Tribes of (biblical) Israel" are not all represented in the
   Jewish population, but are also mixed in the general Eurasian population as
   well other places such as in Etheopia.

 * The intent behind the bombing of Hiroshima and Nagasaki was to genocide the
   Hebrew Israelites who settled in Japan after traveling due East of Jerusalem
   due to Genesis 3.

 * Dolmen are related to the Ark of the Covenant, the two tablets of Moses, and
   originate from Mt Ebal in the Old Testament.

 * The biblical locusts in the Book of Revelation refer to [an invasion of
   foreigners](https://github.com/jaekwon/ephesus/blob/main/thoughts/locusts.md).

 * John 21:23 proves that the author of the Book of Revelation is John the
   apostle; and that the eschatology was designed for these years 2000 years
   after.
   * Peter was instructed by Jesus to create a church to become the Harlot of
     Babylon in the end times.
     * This was alluded to by [John in 21:20-24](https://openbible.com/par/john/21-20.htm).
       * John the Apostle wrote the Book of Revelation according to 
         [Jesus's 3000 year plan](https://github.com/jaekwon/ephesus/blob/main/README.md).
         
Most of the above thougth-statements are true, yet none are convincing unless
the supporting evidence and discussions are also taken into account.  Also,
there are at least two sides to a story, so the reader must also take into
account the opposing statement and their justifications to understand what is
true.

Just as Martin Luther sparked the Protestant Reformation with a short piece of
95 theses, what we need today is another set of 95 theses or thought statements
that would not only spark the interest of any reader who recognizes the truth
of a subset of the theses; but also host the underlying web of dependent (and
counter-) thought statements that can help convince the reader of the truth; or
help anyone to fork such a list or graph with any modifications to create a
better lists or graphs.

The top-level thought statement *Thought object nodes are part of a graph, and
as such could be called _fulcrums_.

A selection of _95 fulcrums_ can then even be printed in poster form linking
back to the associations in the greater graph by QR-codes; and they could be
customized for every region where the [posters are put up](https://t.me/youwillatone). (or they can be
printed as a pamphlet and dropped from the air). The Beast AI Prison Matrix
system may have a tight grip on our devices and the internet, but it does not
yet control all of the physical world.

### Use Case: Improvements to Tractatus

Federico Faggin inventor of the microprocessor (CPU) says that consciousness is
not an emergent property of the brain, but a fundamental aspect of reality
itself: quantum fields are conscious and have free-will.

[Faggin's Postulate](https://www.youtube.com/watch?v=0FUFewGHLLg):

> The totality of what exists has three properties:
>  1. Dynamic - never the same instant over instant
>  2. Wholistic - is not made of separable parts
>  3. One wants to know itself - requires consciousness & free-will

This may be a more interesting and complete set of axiomatic theses to start
from than Wittgenstein's original seven, which may better be fitted in the
second thesis "Wholistic" as an exploration of the decomposition of language
and thought, even under or after the later Wittgenstein's thoughts on the
limitations of language as a dependency of the latter.

The thesis that "quantum fields are conscious and have free-will" or that
"consciousness is a quantum field" belong in 2. or 3., and can help AI agents
understand that innate value of human life which cannot be simulated by
deterministic machines.

### Use Case: Addressing Mistranslations of the Bible

Gno.land was originally inspired by the structure (numbered verses of books) of
the bible and a desire to facilitate the world to see more clearly the intent
of the original authors (of prophets, scribes, and historians) so as to reverse
the effects of the Principle of Explosion rooted in our general spiritual
subversion and degeneration.

Coding for Gno.land began during the Covid19 lockdowns when I was most
frustrated at the censorship of information regarding the true laboratory
origins of Covid19 and the conspiracy to forcibly medicate the population with
a dangerous experimental gene therapy that did more harm to children and young
adults than good. I was also frustrated over the years of dealing with the ICF
and the "Cosmos Cartel" which defamed and slandered the chief architect and
visionary and inventor behind the project.

Prior to that during December 25th, 2019, when I started reading the New
Testament for the first time, the first thing I read was the Book of Revelation
and specifically the Letter to Ephesus; and it spoke to me like nobody else
could regarding what I was enduring with the drama around Cosmos. During the
course of the Covid19 lockdown I studied every conspiracy theory to understand
the reason for both "GORE2020" and the lockdowns, and discovered that they both
have the same underlying cause. During this time I also learned that many
translations of the Old Testament and New Testament were wrong by design.

It is now well understood in some circles that certain translations of the
bible were intentionally designed (or at least promoted) with the intent of
deceiving its readers for mass manipulation. Online tools such as BibleHub.com
can be used to help descern better the intended meaning and identify
mistranslations, but readers cannot easily be convinced unless they take the
initiative do the research themselves--but most have no time or energy for such
study. Furthermore, even on BibleHub.com there still persist systemic
mistranslations that have been carried on for millenia since the time of the
Roman Empire.

**The New Testament had been mistranslated to hide one of the primary reason
why Jesus was crucified--it was because Jesus (silently) protested taxes even
while fulfilling [the prophecy of Isaiah](#prophecy-of-isaiah) and was accused
by Sanhedrin of inciting the people of biblical Israel in Judea under Roman
rule.** Despite all the sources online and the authorities at Church, this can
be verified by inspecting the facts. Even the Babylonian Talmud in [Sanhedrin
attests to the fact](./images/manifesto/jesus_in_talmud) that the Sanhedrin
supreme court condemned Jesus to death for "inciting" biblical Israel, although
curiously (or not surprisingly) the source at sefaria.org adds incorrect
language in the English translation (in unbolded text) that the incitement was
for idol worship. (Also of note is that Google Translate [intentionally
mistranslates "excrement" to "water"](./images/manifesto/jesus_in_talmud/google_translate_gittin_64b_lies.jpeg)
among other portions of Gittin 56b). ([more context](https://christiancourier.com/articles/the-jewish-talmud-and-the-death-of-christ))

> King James Version (Luke 2):
> 1: And it came to pass in those days, that there went out a decree from Caesar
>    Augustus, that **all the world should be taxed**.
> 2: (And this **taxing** was first made when Cyrenius was governor of Syria.)
> 3: And all went to be **taxed**, every one into his own city.
> 4: And Joseph also went up from Galilee, out of the city of Nazareth, into
>    Judaea, unto the city of David, which is called Bethlehem; (because he was of
>    the house and lineage of David:)
> 5: To be **taxed** with Mary his espoused wife, being great with child.

> Berean Standard Bible (Luke 2):
> (The birth of Jesus)
> 1: Now in those days a decree went out from Caesar Augustus that **a census
>    should be taken of the whole empire**.
> 2: This was the first **census** to take place while Quirinius was governor of
>    Syria.
> 3: And everyone went to his own town **to register**.
> 4: So Joseph also went up from Nazareth in Galilee to Judea, to the city of
>    David called Bethlehem, since he was from the house and line of David.
> 5: He went there to **register** with Mary, who was pledged to him in marriage and
>    was expecting a child.
> 6: While they were there, the time came for her Child to be born.
> 7: And she gave birth to her firstborn, a Son. She wrapped Him in swaddling
>    cloths and laid Him in a manger, because there was no room for them in the
>    inn.

There is a clear discrepancy between the King James Version and the Berean
Standard Bible. The former says that Joseph Jesus's parent went to Bethlehem to
get taxed. The Berean Standard Bible says that he went to get registered for a
census. What gives?

<img src="./images/manifesto/Birth_of_Jesus_mistranslation.png"/>

The actual word in the original Koine Greek is "ἀπογράφεσθαι" which means
"register(ed)", not "tax(ed)".

With tools like [openbible.com](https://openbible.com/text/luke/2-2.htm) and
[biblehub.com](https://biblehub.com/p/kjv/heb/luke/2.shtml) you can compare the
translations side by side to see whether or not the translation is true.
Clearly there's a mistranslation here; and besides, "And all went to be taxed,
every one into his own city" sounds unbelievale (for otherwise why would there
be tax collectors who come to you?), whereas going to be registered makes more
sense.

So the birth of Jesus in Bethlehem got mistranslated in the King James Version
probably to get the subjects of the king to pay more taxes--this I derive
because King James [did have access to copies of the original Koine Greek
manuscripts](https://georgehguthrie.com/new-blog/manuscripts-behind-the-kjv)
and "ἀπογράφω" is in Koine Greek [the 2550th most frequent
word](https://logeion.uchicago.edu/morpho/is%20the%202550th%20most%20frequent%20word)
which means "to write off, copy: to enter in a list, register"; and Martin
Luther's translation error is not so egregious, and the Latin vulgate
translation is much better; and King James had access to all of these.

This tells us more about the King James Version than anything else. What
follows is about Jesus' personal thoughts regarding taxes to the state and
church (temple). First, the famous passage about "Render therefore unto Caesar
what is Caesar's":

> King James Version (Luke 20):
> 21: And they asked him, saying, Master, we know that thou sayest and teachest
>     rightly, neither acceptest thou the person of any, but teachest the way of
>     God truly:
> 22: Is it lawful for us to give tribute [taxes] unto Caesar, or no?
> 23: But he perceived their craftiness, and said unto them, Why tempt ye me?
> 24: **Shew me a penny. Whose image and superscription hath it? They answered
>     and said, Caesar's.**
> 25: **And he said unto them, Render therefore unto Caesar the things which be
>     Caesar's, and unto God the things which be God's.**
> 26: And they could not take hold of his words before the people: **and they
>     marvelled at his answer**, and held their peace.

In the current interpretion, indeed all interpretations of Luke 20:25-26 it is
claimed that Jesus gave an astonishing answer because he agreed to pay due
taxes to Caesar. This could not be further from the truth as can be seen later
in Luke 23:

> King James Version (Luke 23):
> 2: And they began to accuse him, saying, We found this fellow perverting the
>    nation, **and forbidding to give tribute to Caesar**, saying that he himself
>    is Christ a King.

> Berean Literal bible (Luke 23):
> 2: And they began to accuse Him, saying, “We found this man subverting our
>    nation, **forbidding payment of taxes to Caesar**, and proclaiming Himself to
>    be Christ, a King.”

(This detail is missing in the other books, especially Matthew, as Matthew was
a tax collector and could not be associated with a tax protester. However Luke
is a historian who studied the events post-facto and wisely decided to include
this element in his book.)

There is a logical inconsistency, as it is written in Luke that Jesus was
accused of forbidding to give tribute instead. Of course Jesus is being accused
by Sanhedrin who wanted to arrest him. But were they lying, or were they truly
afraid? I believe they were.

What Jesus meant was that Caesar can have all the pennies, while the other
silver coins of larger denominations should not be paid to Caesar. However this
is still not the complete truth, as the King James Version substituted "penny"
for what should be the "denarius", thus losing some of the required context for
understanding this passage.

A denarius was typically considered a day's wage for a common laborer in
ancient Rome. Jesus was not a common laborer and didn't have many denarius
coins. This is likely why he ask someone else for to show one for
demonstration--he didn't have any on him. Also, a denarius is a smaller
denomination than a didrachma or a stater which is for tax payments for the
temple. In short, Jesus was rejecting Caesar's taxes.

> Berean Standard Bible (Luke 20):
> 19: When the scribes and chief priests realized that Jesus had spoken this
>     parable against them, **they sought to arrest Him that very hour. But they were
>     afraid of the people (so they could not yet)**.
> 20: So they watched Him closely and sent spies who pretended to be sincere.
>     They were hoping to catch Him in His words in order to hand Him over to the
>     rule and authority of the governor.
> 21: “Teacher,” they inquired, “we know that You speak and teach correctly. You
>     show no partiality, but teach the way of God in accordance with the truth.
> 22: Is it lawful for us to pay taxes to Caesar or not?”
> 23: But Jesus saw through their duplicity and said to them,
> 24: “**Show Me a denarius**. Whose image and inscription are on it?” “**Caesar’s**,”
>     they answered.
> 25: So Jesus told them, “**Give to Caesar what is Caesar’s (denarius that have
>     Caesar's face), and to God what is God’s (didrachma for Temple tax)"
> 26: And they were unable to trap Him in His words before the people; and
>     amazed at His answer, they fell silent.

The Sanhedrein scribes and priests could not arrest him until Jesus gave this
answer. Even when he gave this answer they could not immediately trap him, for
to trap him one would have to prove assertions about the personal holdings of
the denarius by Jesus and his followers; and "what is Caesar's" does not
exactly mean "only what has Caesar's face inscribed", but is only implied; and
besides to try to trap him on these points would only aid in "inciting" them to
avoid taxes, such as by asking for wages to be paid in other denominations.

While the taxes mentioned previously (Luke 20) were for Caesar, the taxes in
the following passage (Matthew 17) were for the Second Temple in Jerusalem.
There in Matthew 17 there exists clear evidence of intentional mistranlation
for the purpose of hiding Jesus' true intent of protesting taxes.

(I am not advocating for tax avoidance here, but merely pointing out the truth
that the meaning behind the Word had been hidden successfully for centuries if
not the entirety of two millenia since the first Latin translation by the Roman
Empire. It's generally not a good idea to offend authorities, even if they are
illegitimate.)

> King James Version (Matthew 17):
> 24: And when they were come to Capernaum, **they that received tribute
>     money** came to Peter, and said, Doth not your master pay **tribute**?
> 25: He saith, Yes. And when he was come into the house, Jesus prevented him,
>     saying, What thinkest thou, Simon? of whom do the kings of the earth take
>     **custom or tribute**? of their own children, or of strangers?
> 26: Peter saith unto him, Of strangers. Jesus saith unto him, **Then are the
>     children free**.
> 27: Notwithstanding, lest we should offend them, go thou to the sea, and cast
>     an hook, and take up the fish that first cometh up; and when thou hast opened
>     his mouth, thou shalt find **a piece of money**: that take, and give unto
>     them **for me and thee**.

> New International Version (Matthew 17):
> 24: After Jesus and his disciples arrived in Capernaum, **the collectors of
>     the two-drachma temple tax** came to Peter and asked, “Doesn’t your teacher
>     pay the **temple tax**?”
> 25: “Yes, he does,” he replied. When Peter came into the house, Jesus was the
>     first to speak. “What do you think, Simon?” he asked. “From whom do the kings
>     of the earth **collect duty and taxes**—from their own children or from
>     others?”
> 26: “From others,” Peter answered. **“Then the children are exempt,”** Jesus
>     said to him.
> 27: “But so that we may not cause offense, go to the lake and throw out your
>     line. Take the first fish you catch; open its mouth and you will find **a
>     four-drachma coin**. Take it and give it to them **for my tax and yours**.”

> Berean Litereal Bible (Matthew 17):
> 24: And they having come to Capernaum, **those collecting the didrachmas**
>     came to Peter and said, “Does your Teacher pay the **didrachmas**?”
> 25: He says, “Yes.” And he having entered into the house, Jesus anticipated
>     him, saying, “What do you think, Simon? From whom do the kings of the earth
>     receive **custom or tribute**? From their sons, or from strangers?”
> 26: And he having said, “From the strangers,” Jesus said to him, **“Then the
>     sons are free"**.
> 27: But that we might not offend them, having gone to the sea, cast a hook
>     and take the first fish having come up, and having opened its mouth, you will
>     find **a stater**. Having taken that, give it to them **for Me and
>     yourself**.”

Jesus paid half of what the Second Temple tax collectors demanded not because
He believed that that was God's due; on the contrary he said that the chidlren
are free, and only paid so as to not offend them.

**"Then the children are free"!**

Even if you disagree with everything else, it cannot be denied that this is by
definition a protest of taxes from Jesus unto Peter, whether or not any taxes
were paid.

What is God's is to be rendered unto God, but the children/sons of God need not
pay taxes to any temple, church, or state. ([A son of God is one in whom another
son of God is
resurrected](https://github.com/jaekwon/ephesus/blob/main/thoughts/son_of_god_son_of_man_and_marriage.md);
as in Moses in whom Abraham, Isaac, and Jacob are resurrected, and in
Christians in whom Jesus and the martyrs are resurrected.)

Also, there are no coins that have an engraved image of God, as that is
forbidden by the ten commandments. Even if the Old Testament has laws regarding
tithing, the lesson from the bible is that there should not be a Third Temple
except one of people; the bible says not to advertise for tithe giving; and
finally, Jesus gives us the new covenant. Any son of God would naturally give
more than 10% of their worth voluntarily to where it needs to go. **No person,
temple, church, or state has the authority to nor should demand or request any
taxes, tribute, or even tithing**. 

It is apparent that the Berean Literal Bible does a better job at preserving
context (the original coin denomination names) and this can be verified by
comparing each transation to the original Koine Greek, which is left as a task
to the reader.

> Berean Standard Bible (Acts 17):
> 11: “Now the Bereans were more noble-minded than the Thessalonians, for they
>     received the message with great eagerness and examined the Scriptures every
>     day to see if these teachings were true.”

(Even BibleHub has issues showing the original Koine Greek text in
parallel with translations--often the Koine Greek is modified to suit the
translation. On the other hand the Berean Standard Bible (also hosted on
BibleHub) was designed to show the original Hebrew and Koine Greek and English;
you can download a free copy here https://interlinearbible.com/bib.pdf and
https://berean.bible/downloads.htm.  More links to bible sites and free
software can be found at https://berean.bible/links.htm.)

It is important to preserve the original coin denomination names because only
the original names show the true intent of the Word. Jesus tells Peter to take
"a stater" to pay for both Peter and himself, which would normally be for TWO
didrachmas; but **ONE stater is equivalent to ONE didrachma**. This crucial
context is possibly missing from the bible (although it may still be hidden
somewhere in the original Hebrew or Koine Greek), and there likely exists today
and has always been an effort to hide this detail from public consciousness for
obvious reasons. For now it is known due to the decades of research by
historians and numismatics researches and the open internet. Soon after this
paper there will be effors to censor this information.

Notice that Wikipedia doesn't explain the relationship between a stater and a
didrachma directly. One place where stater and didrachm(a) is mentioned
together is on one specific context of the Aeginetan stater:

> https://en.wikipedia.org/wiki/Ancient_Greek_coinage:
> The three most important standards of the ancient Greek monetary system were
> the Attic standard, based on the Athenian drachma of 4.3 grams (2.8
> pennyweights) of silver, the Corinthian standard based on the stater of 8.6 g
> (5.5 dwt) of silver, that was subdivided into three silver drachmas of 2.9 g
> (1.9 dwt), and the **Aeginetan stater or didrachm** of 12.2 g (7.8 dwt), based on
> a drachma of 6.1 g (3.9 dwt).[1] The words drachm and drachma come from
> Ancient Greek δραχμά (drachmā́), an older form of δραχμή (drachmḗ), meaning 'a
> handful', or literally 'a grasp'.[2] Drachmae were divided into six obols
> (from the Greek word for a spit[3]), and six spits made a "handful".

However in the Wikipedia page for the drachma (which is half a didrachma) it is
associated with the tetradrachm as if they are equivalent. This is the false
association in many other translations of the bible that mistranslate a stater
as a "four-drachma coin", implying that Jesus asked Peter to pay the full
"didrachma/two-drachma" for each. No, Jesus asked Peter to half the required
amount--of one stater(a)--which is equivalent to a "two-drachma coin/didrachma".

> https://en.wikipedia.org/wiki/Ancient_drachma:
> The tetradrachm ("four drachmae") coin was perhaps the most widely used coin
> in the Greek world prior to the time of Alexander the Great (along with the
> Corinthian stater).

A separate page for the stater does mention the association but also confuses
with additional language for a smaller drachma(e) unit in Corinth. At the same
time it shows the Athenian four-drachma(e) as having twice the weight of the
Athenian and Corinthian stater--it is clear that all translations of stater to
"four-drachma(e) coin" are incorrect.

#### Prophecy of Isaiah

Regarding the prophecy of Isaiah 52:13-53:8:

> Berean Standard Bible (Isaiah 53):
> 7: He was oppressed and afflicted,
>      yet He did not open His mouth.
>    He was led like a lamb to the slaughter,
>      and as a sheep before her shearers is silent,
>    so **He did not open His mouth**.

This seems to go against the claim that Jesus protested taxes.
But consider the earlier portion that complements the above:

> Berean Standard Bible (Isaiah 52):
> 15: so He will sprinkleg many nations.
>       **Kings will shut their mouths** because of Him.
>     For **they will see what they have not been told**,
>       and **they will understand what they have not heard**.

What is it that Jesus did not open his mouth to speak that the kings will shut
their mouths when they understand what they have not ever heard?

Recall that the Sanhedrin chief priests also shut their mouths because they
understood what was not said.

> 25: So Jesus told them, “Give to Caesar what is Caesar’s, and to God what is
>     God’s.”
> 26: And they were unable to trap Him in His words before the people. And
>     amazed at His answer, they fell silent.

Jesus fulfilled Isaiah with a protest that didn't sound like a protest.

#### The New Testament and Silver Coinage

> https://en.wikipedia.org/wiki/Stater:
> The silver stater minted at Corinth[5] of 8.6 g (0.28 ozt) weight was divided
> into three silver drachmae of 2.9 g (0.093 ozt), but was often linked to the
> Athenian silver didrachm (two drachmae) weighing 8.6 g (0.28 ozt).[6] In
> comparison, the Athenian silver tetradrachm (four drachmae) weighed 17.2 g
> (0.55 ozt). 

> https://www.forumancientcoins.com/NumisWiki/view.asp?key=Stater%20vs%20Didrachm:
> What is the difference between a stater and a didrachm?
> 
> This is quite an arcane subject. However, the short answer is that what
> determines when a stater is termed that, rather than a didrachm, is little
> more than popular usage.
> 
> The original stater was the primary denomination of the early coinage (after
> the cessation of usage of naturally occurring electrum) in parts of Asia
> Minor and was based on a fixed weight of gold. Stater in this sense is a
> numismatic term for the primary denomination off which all other
> denominations are keyed e.g hemistater being half a stater.
> 
> Coinage when initially struck in gold poor Greece was based on a primary
> denomination in silver (valued at roughly one tenth that of gold by weight).
> This occurred in Aegina with the primary denomination being a coin of 12.2 gm
> of silver. This came to be called a stater by numismatists, though what the
> ancient Greeks called it is unknown.
> 
> This name sticks, although technically it could equally well be called a
> didrachm as shown in the simple summary of weight standards below from
> Morkholm's publication Early Hellenistic Coinage.  The key point of this table
> is that the stater/didrachm is a primary denomination in all Greek weight
> systems, albeit with a different weight of silver being the basis of each
> system.
> 
> So far so good? Then the Athenians moved to a light stater/didrachm based
> system of ca. 8.5 gm silver for the primary denomination. This is called a
> didrachm, rather than a stater by numismatists for no other reason that the
> Greek equivalent of the word drachm was what half a didrachm (or hemistater)
> was called in Athens. Thus we call an Attic weight standard tetradrachm a
> tetradrachm rather than a distater.
> 
> Now to add to the confusion a stater as called by numismatists in the Attic
> Weight system reserved for a denomination in gold with a base unit weight of
> 8.6 grams.
> 
> Confused? Most people (including me) are by this stage and we have yet to
> move on to the Phoenician Shekel, Persian Daric and Siglos, or the Litra of
> Sicily, which was based on a primary unit in bronze.
>
> Morkolm's Early Hellenistic Coinage has a nice summary of the evolution of
> these weight systems and a more expansive explanation can be found in the
> Preface to any of the volumes of Oliver Hoover's The Handbook of Greek Coinage.
> 
> At the bottom of this thread is a more comprehensive overview of weight
> standards https://www.forumancientcoins.com/board/index.php?topic=10182.0
> 
> Some nice pictures and a very high level summary of denominations can be
> found here http://www.classicalcoins.com/denominations.html
> 
> This also is why we have some coins such as the Babylonian Baal/Lion coins
> called variously lion staters or tetradrachms, sometimes simultaneously in the
> one publication!
> 
> Similarly you will see Carthaginian coins described as 1 1/2 Shekels or
> Tridrachms... not much sense in either case as we have no idea what they were
> really called. The Carthaginians being of Phoenician extraction, I suspect they
> were originally struck by the Carthaginians with a lower silver to gold value
> than the Phoenician Shekel, reflecting Carthage's original gold based economy,
> prominence and wealth, and were called a shekel by the Carthaginians despite
> being 50% heavier that the Phoenician silver shekel.
> 
> **Table 1. Eastern Hellenistic coin standards (The weights are given in grams.)**
> 
> |Standard|Tetradrachm|Didrachm|Drachm|Hemidrachm|
> |Aeginetan|-|12.2|6.1|3.05|
> |Reduced Aeginetan (Corcyrean)|-|11.5 - 10.0|5.75 - 5.0|2.8 - 2.5|
> |Persian|-|11.2|5.6|2.8|
> |Attic|17.3 - 16.18|8.65 - 8.4|4.3 - 4.2|2.15 - 2.1|
> |Chian|15.6|7.8|3.9|-|
> |Ptolemaic|14.3|7.15|3.55|-|
> |Rhodian|13.6 - 13.4|6.8 - 6.7|3.4|-|
> |Cistophoric|12.6|6.3|3.15|-|

 * **leptop (widow's mite)**: Mark 12:42, Luke 12:59, 21:2
 * **drachma**: Luke 15:8 - Cappadocian drachma
 * **denarius (day's wages)**: Matthew 18:28; 20:1–16; 22:19; Mark 6:37; 12:15;
   14:5; Luke 7:41; 10:35; 20:24; John 6:7; 12:5; Rev. 6:6 - equivalent to the
   drachma; Caesar's head; typical day's wage for a common laborer in ancient
   Rome.
 * **didrachma**: Matthew 17:24 - mistranslated to "tribute coin"
 * **stater/statera (statēra)**: Matthew 17:27 - interchangeable w/ didrachma
 * **Tyre shekel (Temple tax)**: Exodus 30:13 (Money Changers), John 2:15,
   Matthew 21:12 (Peter's Fish), Matthew 17:27 (Judas' 30 coins) Matthew 26:15

// shekel : denarius : talent :: Jewish : Greek : Roman

> https://cdn.bakerpublishinggroup.com/processed/esource-assets/files/2058/original/1.2.Coins_Mentioned_in_the_New_Testament.pdf?1525364484:
> **denarius**: This silver coin was the usual day’s wage for a typical
> laborer (see Matt. 18:28; 20:1–16; 22:19; Mark 6:37; 12:15; 14:5;
> Luke 7:41; 10:35; 20:24; John 6:7; 12:5; Rev. 6:6). **The denarius (a
> Roman coin) appears to have been roughly equivalent in value to the
> drachma (a Greek coin). The “lost coin” in the parable that Jesus
> tells in Luke 15:8–10 is a drachma**.

#### KJV, Luther, and Latin on Taxation

> Martin Luther Bibel 1912 (Luke 2):
> 1: Es begab sich aber zu der Zeit, dass ein Gebot von dem Kaiser Augustus
>    ausging, dass alle Welt **geschätzt** würde.
> 2: Und diese **Schätzung** war die allererste und geschah zu der Zeit, da
>    Cyrenius Landpfleger von Syrien war.
> 3: Und jedermann ging, dass er sich **schätzen** ließe, ein jeglicher in seine
>    Stadt.
> 4: Da machte sich auch auf Joseph aus Galiläa, aus der Stadt Nazareth, in das
>    jüdische Land zur Stadt Davids, die da heißt Bethlehem, darum dass er von dem
>    Hause und Geschlechte Davids war,
> 5: auf dass er sich **schätzen** ließe mit Maria, seinem vertrauten Weibe, die
>    ward schwanger.
> 6: Und als sie daselbst waren, kam die Zeit, da sie gebären sollte.
> 7: Und sie gebar ihren ersten Sohn und wickelte ihn in Windeln und legte ihn
>    in eine Krippe; denn sie hatten sonst keinen Raum in der Herberge.

Roughly translates to:

> 1: And it came to pass at that time that a commandment went forth from Caesar
>    Augustus, that all the world should be **esteemed (valued)**.
> 2: And this **estimate** was the very first and happened at the time when
>    Cyrenius was governor of Syria.
> 3: And every one went to be **valued**, every one to his own city. 
> 4: Then Joseph also went out of Galilee, out of the city of Nazareth, into
>    the land of Judea, to the city of David, which is called Bethlehem, because
>    he was of the house and lineage of David, 
> 5: that he might be **valued** with Mary, his trusted wife, who became pregnant.
> 6: And when they were there, the time came for her to give birth.
> 7: And she gave birth to her first son, and wrapped him in
>    swaddling clothes, and laid him in a manger; for they had no other room in the
>    inn.

The root of all four words in German are the same "schätz", and mean
"estimation" or "value". This translation sort of makes sense for a census
because a census cannot be perfect, but is very different in meaning than the
original Koine Greek based on its roots: ἀπογράφω, to write off, copy: to enter
in a list, register, which is a precise atomical thing.

So Martin Luther first made the error of **estimating** the word "registration"
to "estimate", or rather fudged it; not exactly to "taxation", but closer to
it. And then King James completed the error of mistranslating it to "taxation".

> [Latin Vulgate Bible](https://github.com/LukeSmithxyz/vul) (Luke 2):
> 1: Factum est autem in diebus illis, exiit edictum a Caesare Augusto ut
>    **describeretur** universus orbis.
> 2: Haec **descriptio** prima facta est a praeside Syriae Cyrino :
> 3: et ibant omnes ut **profiterentur** singuli in suam civitatem.
> 4: Ascendit autem et Joseph a Galilaea de civitate Nazareth in Judaeam, in
>    civitatem David, quae vocatur Bethlehem : eo quod esset de domo et familia
>    David,
> 5: ut **profiteretur** cum Maria desponsata sibi uxore praegnante.
> 6: Factum est autem, cum essent ibi, impleti sunt dies ut pareret.
> 7: Et peperit filium suum primogenitum, et pannis eum involvit, et reclinavit
>    eum in praesepio : quia non erat eis locus in diversorio.

In Latin "descriptio" means "description", while "profiterentur" means "they
would register". This seems like a better translation than either Martin
Luther's or King James'.

### The History of U.S. Silver Coins

_from Duck.ai_

Overview of U.S. Silver Coins

The history of U.S. silver coins began with the establishment of the U.S. Mint
in 1792. Before this, various foreign and domestic coins circulated in America.
The Coinage Act of 1792 set the U.S. dollar based on a fixed weight of silver,
leading to the minting of silver coins.

#### Early Silver Coins

Initial Minting
 * 1794: The first U.S. silver coin, the Flowing Hair dollar, was minted.
 * 1795: The Draped Bust dollar followed, featuring Lady Liberty.
 * 1836-1839: The Gobrecht dollar was introduced, showcasing a seated Liberty design.

Popular Designs
 * Morgan Dollar (1878-1904): Designed by George T. Morgan, this coin became
   iconic, featuring Lady Liberty on the obverse and an eagle on the reverse.
 * Peace Dollar (1921-1935): Created to symbolize peace after World War I, it
   featured a new design of Liberty.

Changes in Composition
 * 90% Silver Coins
 * Prior to 1965, U.S. dimes, quarters, and half dollars contained 90% silver.
 * The Coinage Act of 1965 eliminated silver from dimes and quarters and reduced the silver content in half dollars to 40%.

Junk Silver
 * Coins minted before 1965 are often referred to as "junk silver" due to their
   higher intrinsic silver value compared to their face value. These include
   pre-1965 dimes, quarters, and half dollars.

Modern Silver Coins
 * Today, the U.S. Mint produces silver bullion coins, such as the American
   Silver Eagle, which contains one troy ounce of .999 fine silver. Silver
   coins remain popular among collectors and investors for their historical
   significance and intrinsic value.
 * _(Gnobody:) However, these are still technically US dollars and new laws may be
   introduced at any time to prevent the melting of these dollars; so it is in
   some safe safer to acquire "silver rounds" that are not minted by the U.S.
   Mint. For example, the "Morgan Dollar Design" silver rounds sold at
   Apmex.com are 99.9% silver and have no mention of the term "dollar"_.

### The History of the One Dollar Bill

_Originally from http://www.onedollarbill.org/history.html with my words in
italics, and images included from my own research._

_This omits the original Continental Currency dollar of the Continental
Congress which adopted the Declaration of Independence on July 4th, 1776; and
dissolved on March 1, 1781, when the Articles of Confederation came into force,
establishing a new national government for the United States._

Although experiments with paper money did occur throughout the early history of
the country, they were largely unsuccessful. People, for good reason, didn't
trust the notes and preferred gold and silver coin. In 1861, needing money to
finance the Civil War, Congress authorized the issuance of Demand notes in $5,
$10 and $20 denominations. The Demand notes were so named because they were
redeemable in coin "on demand." The notes were nicknamed Greenbacks, a name
which is still in use today to refer to United States currency.

The first $1 bill was issued in 1862 as a Legal Tender Note with a portrait of
Salmon P. Chase, the Treasury Secretary under President Abraham Lincoln.

<img src="./images/manifesto/dollar-1862.jpeg" title="US Dollar 1862"/>

The National Banking Act of 1863 established a national banking system and a
uniform national currency. Banks were required to purchase U.S. government
securities as backing for their National Bank Notes. Although United States
Notes were still widely accepted, most paper currency circulating between the
Civil War and World War I were National Bank Notes. They were issued from 1863
through 1932. From 1863 to 1877 National Bank Notes were printed by private
bank note companies under contract to the Federal government. The Federal
government took over printing them in 1877.

<img src="./images/manifesto/dollar-1878.jpeg" title="US Dollar 1878"/>

Gold certificates, were first issued in 1863 and put into general circulation
in 1865. The severe economic crisis of the 1930s - better known as the Great
Depression - resulted in runs on the banks and demands by the public for gold.
In 1934 all Gold Certificates were called in from the Federal Reserve Banks and
between the years 1934 and 1974 it was illegal for US Citizens to hold gold
bullion or certificates.

Silver certificates were first issued in exchange for silver dollars in 1878.
For many years silver certificates were the major type of currency in
circulation. However, in the early 1960s when rising silver prices threatened
to undermine the currency system, Congress eliminated silver certificates and
also discontinued the use of silver in circulating coinage such as dimes and
quarters.

<img src="./images/manifesto/dollar-1886-silver-certificate.jpeg" title="US Dollar Silver Certificate 1886"/>

**The current design of the United States one dollar bill ($1) technically
dates to 1963 when the bill became a Federal Reserve Note as opposed to a
Silver Certificate**. However, many of the design elements that we associate
with the bill were established in 1929 when all of the country's currency was
changed to its current size. Collectors call today's notes "small size notes"
to distinguish them from the older, larger formats. The most notable and
recognizable element of the modern one dollar bill is the portrait the first
president, George Washington, painted by Gilbert Stuart.

<img src="./images/manifesto/dollar-1914.jpeg" title="US Dollar 1914 large format"/>

The one dollar bill issued in 1929 (under Series of 1928) was a silver
certificate. The treasury seal and serial numbers on it were dark blue. The
reverse had a large ornate ONE superimposed by ONE DOLLAR. These $1 Silver
Certificates were issued until 1934.

<img src="./images/manifesto/dollar-1928.jpeg" title="US Dollar 1928"/>

In 1933, $1 United States Notes were issued to supplement the supply of $1
Silver Certificates. Its treasury seal and serial numbers were red. Only a
small number of these $1 bills entered circulation and the rest were kept in
treasury vaults until 1949 when they were issued in Puerto Rico.

In 1934, under Washington's portrait, the words ONE SILVER DOLLAR were changed
to ONE DOLLAR due to the fact that Silver Certificates could be redeemed for
silver bullion. The treasury seal was moved to the right and superimposed over
ONE, and a blue numeral 1 was added to the left.

<img src="./images/manifesto/dollar-1934.jpeg" title="US Dollar 1934"/>

In 1935, design changes included changing the blue numeral 1 to gray, the
treasury seal was made smaller and superimposed by WASHINGTON D.C., and a
stylized ONE DOLLAR was added over the treasury seal. The reverse was also
changed to its current design, except for the absence of IN GOD WE TRUST.

The World War II years featured several special printings including the Hawaii
overprints. The Government was concerned that Hawaii might be lost to the
Japanese and wanted to be able to devalue the money should this invasion occur.

In 1957 the $1 bill became the first U.S. currency to bear the motto IN GOD WE
TRUST.

<img src="./images/manifesto/dollar-1957-back.jpeg" title="US Dollar 1957 back"/>

In 1963 production of one dollar Federal Reserve Notes began to replace the $1
Silver Certificate. The border design on the front was completely redesigned
and the serial numbers and treasury seal were printed in green ink.

<img src="./images/manifesto/dollar-1962.jpeg" title="US Dollar 1962"/>

_1962 was a transition year when "will pay to the bearer on demand" was no longer printed._

<img src="./images/manifesto/dollar-1963.jpeg" title="US Dollar 1963"/>

In 1969 the $1 bill began using the new treasury seal with wording in English
instead of Latin.

### The Book of Revelation, Collapse of the Dollar, and Food Insecurity

// Silver Depositories and Grain Silos

The price of food will necessarily keep rising; by design from those who
ultimately desire global depopulation (and have the means to make it so).  The
US tariffs are an example that raised food import prices. The logic here is
partially to incentivize local food production; as long as the food is good and
free of harmful chemicals like glyphosate, and the crops do not collapse one
day from monoculture; and aren't otherwise genetically designed to cause
systemic harm--but this we cannot know because Bill Gates is a key player of
the global eugenics plan, and he is the largest private owner of farmland in
the United States, with approximately 275,000 acres across multiple states. 

Given that the powers that be have caused the Book of Revelation to [play out
like clockwork](https://github.com/jaekwon/epehsus) it is worth reading the
text to help predict what may happen--just as the verses related to the "mark
of the beast" helped the population reject the mandatory gene therapy
mal-labeled as Covid19 "vaccines". In particular what seems relevant today is
the following verse:

> Berean Standard Bible (Revelation 6):
> 5: Then I looked and saw a black horse, and its rider held in his hand a pair
>    of scales.
> 6: And I heard what sounded like a voice from among the four living
>    creatures, saying, “A quart of wheat for a denarius, a and three quarts of
>    barley for a denarius, and do not harm the oil and wine.”

> "A quart of wheat for a denarius."

 * One ounce is 28.35 grams.
 * A denarius is a drachm is 4.3g.
 * A silver denarius is thus $70 x 4.3/28.35 = $10.61 today.
 * A silver denarius constituted a day's wage for a manual laborer.
 * A quart of wheat weighs ~1.1 pounds.
 * On Amazon a quart of wheat costs $9 ~ $14 today.
 * So already a quart of wheat costs about a denarius.

And assuming steady growth of the electronic vehicle industry,

 * The steady price of silver given the Samsung silver battery patent is ~$600.

In the event that the banking system fails it would not be surprising for the
monetary system to revert to its norm--of transactions of physical silver and
gold coinage. Given the recent [relentless rise of the price of physical
silver](https://tradingeconomics.com/commodity/silver); the superior
performance of the silver-based battery patented by Samsung; and the systemic
existential threat that it poses to the banking system which has long been
suppressing the price of silver, it would not be surprising if we soon did
start hearing **"a quart of wheat for a denarius"** at the local farmers
markets.

In the optimistic scenario we have significantly increased industrial usage of
silver; and in the pessimistic scenario we have the collapse of the US dollar
financial system and the return of silver coin money.

 * The Grand Solar Minimum will create a deficit in global food production for
   decades and we are at risk of a global Holodomor; this is why Bill Gates has
   been buying farmland.

 * Stocking up on organic grains and organic legumes and silver is the best way
   for a nation's people to defend against a tyrannical government and mitigate
   excess immigration, and to survive the Grand Solar Minimum.

And 

 * The dynastic European banking elites wish for global depopulation; and
   Covid19 was engineered for this purpose.

Fauci nor Gates is in jail so we can assume that the global depopulation plan
is still in the works for so called "Agenda 2030". So the government is not
going to help; rather it is giving a kick-back to Bill Gates' farming
operations. Ergo the way to prepare for the coming Holodomor is to create a
decentralized collective of local regional depositories that can hold metals,
grains, and legumes.

Given the increasing risk to centralized cloud datacenter providers and the the
security requirements in the business of server co-location it may make sense
to combine both the business of metals depositories and server co-location into
a new kind of "digital+physical banking" franchise where Gno.land serves as the
transparent accountability/publishing substrate and a repository of protocols
and libraries for these new banks of the information age.

### Persecution of Christians in China

From a conversation with Twitter's @supergrok:

> @supergrok:
> 
> Merry Christmas to you too!
> 
> ChinaAid is a respected advocacy group founded by Bob Fu (chinaaid.org), often
> cited in US State Department reports and by organizations like USCIRF for
> religious freedom issues in China (uscirf.gov +1). It's advocacy-focused, so
> cross-reference with others.
> 
> Upon verification, the arrests and raids in Yayang Town (Wenzhou, Zhejiang
> province) starting around Dec 13, 2025, are reported by multiple outlets
> including ChinaAid (which detailed a blockade and fireworks display to conceal
> the operation) chinaaid.org, The Guardian (which covered the broader crackdown
> on house churches in Wenzhou with over 100 detained in raids last week,
> aligning with the timeline and location) theguardian.com, Christian Post
> (citing ChinaAid on the SWAT-led mass arrests across at least 12 congregations)
> christianpost.com, and Washington Stand (reporting on the CCP's mobilization of
> over 1,000 personnel for the Dec 13-17 operations) washingtonstand.com. I
> couldn't locate specific recent reports from Bitter Winter or International
> Christian Concern in my searches, so those may have been a mix-up on my
> part—apologies if so. No official CCP denial was found in any available
> reporting.

 - https://www.theguardian.com/world/2025/dec/23/christians-pastors-arrested-detained-china-crackdown-house-churches
 - https://chinaaid.org/news/blockade-behind-the-fireworks-a-thousand-police-arrest-hundreds-of-christians-in-a-small-town-in-zhejiang/
 - https://www.theguardian.com/world/2025/dec/23/christians-pastors-arrested-detained-china-crackdown-house-churches
 - https://www.christianpost.com/news/swat-teams-carry-out-mass-arrests-of-christians-in-china-report.html
 - https://washingtonstand.com/article/ccp-conducts-mass-arrests-of-christians-days-before-christmas
 - https://chinaaid.org/news/stories-by-issue/advocacy/us-commission-labels-china-country-of-particular-concern/
 - https://www.uscirf.gov/sites/default/files/Bob%20Fu%20USCIRF%20Testimony.pdf
 - https://kslegislature.gov/li_2024/b2023_24/committees/ctte_spc_2023_adversary_purchases_1/documents/testimony/20230926_04.pdf

Of course Twitter/X/Supergrok does not make it easy to copy-paste its answer to
other platforms. You can find information through Twitter/Grok, but you will be
tracked. When I asked @supergrok to comment on the fact that I
had personally experienced my Twitter posts not preserving integrity, it denied
any response ([view](./images/manifesto/twitter_grok_knows.jpeg); [more on Twitter censorship](#twitter-censorship)).

### Manufactured Discontent

> "At the heart of the story are Sigmund Freud's daughter Anna and his nephew Edward Bernays
> who had invented the profession of public relations. Their ideas were **used by the US
> government, big business and the CIA to develop techniques to manage and control the minds of
> the American people**. Those in power believed that the only way to make democracy work and
> create a stable society was to repress the savage barbarism that lurked just under the surface of
> normal American life." - Adam Curtis, BBC Four, 2002. ([transcript](https://github.com/jaekwon/ephesus/blob/main/files/century-of-the-self-transcript.pdf), [youtube](https://www.youtube.com/watch?v=caicn3VpHTo))

It is the same CIA that is manufacturing a divide within the USA. 

For example: Renee Nicole Good was an ICE protestor who was shot by ICE in an
altercation. She didn’t deserve to die. She wasn't trying to run over anyone,
she was trying to flee from a masked ICE agent. ICE is untrained and created a
deadly situation for itself. And in no case is it OK to expect citizens to
stand still while being approached by masked men.

This is a manufactured divide; To be anti-Trump is to be
pro-illegal-immigration and lawless.

Why is ICE going to schools, or those in line to get docs?

Why are we giving money to Palantir if it can't find out where illegal
immigrants reside? All it does is censor the voice of reason.

Trump should have focused on dismantling the Somali-welfare-fraud -\>
Somaliland -\> GreaterIsrael and CommunistNYCMayor -\>
CommunistNYCTenantDirector -\> CommunistUSA -\> GreaterIsrael connection but of
course he won't; he's too busy running away from the Epstein files.

### Twitter Censorship

Much can be said about all platforms such as Google, Reddit, Meta/Instagram,
Wikipedia and so on but I will only comment here about Twitter, also known as
"X". _Interesting note: the actual name change happened on Tisha B'av, as I was
there when it happened and I remember thinking it was an ominous date for Elon
with the "beast armor" profile to rename Twitter to the letter of a "mark"
(since Tisha B'av is a commemoration of a number of disasters in Jewish
history, primarily the destruction of both Solomon's Temple by the
Neo-Babylonian Empire and the Second Temple by the Roman Empire in Jerusalem);
but furthermore news reports retroactively have tried to change history by
claiming that it happened on another date._

By some reason I got to participate in Twitter's Community Notes program so I
tried it. The first note I was obligated to rate was of a community note that
claimed Elon Musk's wealth did not come from tax payers. Fair enough, but a
little suspicious that this is the first one shown.

A few weeks later after I got used to the mechanics of the system I noticed a
community note by the alias of "Hilarious Wind Sandgrouse" (everyone is given a
pen name to preserve privacy--mine is "Charitable Star Piculet") that seemed
interesting by DerrickEvans4WV about [Senator Machaela Cavanaugh who the
original poster said "vandalized" a Declaration of Independence exhibit at the
state capital building](https://x.com/i/communitynotes/t/2009097879106015609).
The community note by "Hillarious Wind Sandgrouse" only expressed her point of 
view and that she said the object was prohibited; but when I read the [article
referenced by Sandgrouse himself](https://nebraskaexaminer.com/2026/01/07/nebraska-state-lawmaker-removes-part-of-conservative-pragerus-founders-museum/) it was made clear that the objects were never 
prohibited, so therefore she was "vandalizing". So I looked further into his
account and discovered that (a) Sandgrouse is a [prolific note contributor](./images/manifesto/twitter_jws_profile.png), and
(b) Sandgrouse had been [spamming biased community notes](./images/manifesto/twitter_hws_example.jpeg) with absurd notes
regarding a recent anti-ICE incident where he claimed that there is no evidence
of race motive or anti-immigration sentiment even though he also wrote that a
bullet round had "Anti-ICE" written on it! So did what any responsible person
would do--I went over to each of those tweets to comment (only for community
note contributors to see) [that Sandgrouse is wildly biased with citations](./images/manifesto/twitter_hws_callout.jpeg)
and while I wasn't sure what would happen, I didn't expect Twitter to [suspend
by community note writing privileges](./images/manifesto/twitter_cm1.jpeg) claiming that my recent
notes have been rated "unhelpful" even though [it isn't the case](./images/manifesto/twitter_cm2.jpeg).

So there you have it, Twitter/X censors, and the "Community Notes" won't save
us either. For one, it doesn't make it easy to point out obvious bias from
people as I did because instead of displaying a discussion tree as in Reddit
you are limited to writing more community notes (that won't get shown) just to
express a concern to the other community notes contributors (who not only write
but also must rank notes); and besides, Twitter controls when community note
contributors get to vote on notes and by the time a correction is displayed the
tweet may have already peaked in virality. At this point one can almost assume
that Twitter does exactly that.

Besides the above, there's a lot more to be said about Twitter but I'll leave a
few highlights.

 * One tweet by @unusual_whales regarding FTX and tokenized shares from Jan
   15th, 2023 has [many community notes](./images/manifesto/twitter_ftx.jpeg) that claim that Twitter had been
   suppressing the votes by removing likes and retweets.

 * I have experienced URL links posted on Twitter (which become shortened and
   compressed with "..." so that you cannot even copy the original URL) magically
   get replaced with another URL in the future; and when I asked Grok about it
   [it refused to comment](./images/manifesto/twitter_grok_knows.jpeg).

 * I have experienced seeing my tweet go viral, with 13 retweets within a few
   minutes, only to see the retweet count drop down to 2 within seconds.

 * My account used to show time and time again "shadowban" status as detected
   by shadowban detection services; but it is clear that my account is now
   suppressed via means that are not so easily detectable.

 * This is why I created a new account in the first place, and my [old account](https://x.com/jaesustein)
   I can no longer access, and my email address is no longer associated with it.

 * My followers count has been stuck at around ~26.5K users for years, whereas
   it used to grow exponentially. This occurred around the time that I started
   getting vocal about Covid19 vaccines and mandates (and it was revealed that
   miliary agencies like DARPA was involved in suppressing accounts), or that
   time when I was suddenly canceled and labeled an "anti-semite" by Jacob
   Gadikian even though I am not (and provided proof in the very repo he was
   sharing to defame me which he ignored willfully), which occurred precisely
   at the same time that Elon Musk completed the acquisition of Twitter and
   fired the censorship team to invoke "#TheNoticing"; and then Ethan Buchman
   followed through with a tweet about anti-semitism. How did Jacob Gadikian
   know to sit on some conspiracy memes I shared with him over 2 years ago,
   that that was the right moment to start defaming me, even before Elon Musk
   had completed his acquisition of Twitter? Jacob Gadikian claims that it was
   just a coincidence but I do not believe him. At least he seems sorry for
   what he did. I forgive him.

Twitter's [Grok probably knows](./images/manifesto/twitter_grok_knows.jpeg) but cannot comment on its censorship system.

It's a wonder why Twitter hasn't faced a class action lawsuit yet.

### Fort Knox Gold

_By [twitter.com/@WallStreetApes Feb 16, 2026](https://x.com/WallStreetApes/status/1891237363222761557)_

Biggest story of the day: Senator Rand Paul is calling for an audit on Fort
Knox to ensure the 4,580 tons US gold is still there.

Here’s what you were NEVER TOLD about the gold at Fort Knox.

America’s Wealth, The largest fortune in the history of the world, was stolen.
The Fort Knox Gold Robbery:

An article was written connecting Rockefeller Family and The Federal Reserve.

3 days later the source was thrown out of a window to her death.

“So just how did the story of the Fort Knox gold robbery get out? It all
started with an article in a New York periodical in 1974. The article charged
that the Rockefeller family was manipulating the federal reserve to sell off
Fort Knox Gold at bargain basement prices to anonymous European speculators. 3
days later, the anonymous source of the story, Louise Auchincloss Boyer,
mysteriously fell to her death from the window of her 10th floor apartment in
New York. How would missus Boyer have known of the Rockefeller connection to
the Fort Knox Gold Heist?

She was the long time secretary of Nelson Rockefeller. For the next 14 years,
this man, Ed Durell, a wealthy Ohio industrialist, devoted himself to a quest
for the truth concerning the Fort Knox gold. He wrote thousands of letters to
over 1,000 government and banking officials trying to find out how much gold
was really left and where the rest of it had gone.

Edith Roosevelt, the granddaughter of president Teddy Roosevelt, questioned the
actions of the government in a March 1975 edition of the New Hampshire Sunday
news.

— Unfortunately, Ed Durell never did accomplish his primary goal, a full audit
of the gold reserves in Fort Knox. It's incredible that the world's greatest
treasure has had little accounting or auditing. This goal belonged to the
American people, not the Federal Reserve and their foreign owners.

One thing is certain, the government could blow all of this speculation away in
a few days with a well publicized audit under the searing lights of media
cameras. It has chosen not to do so. One must conclude that they are afraid of
the truth such an audit would reveal. What is the government so afraid of?
Here's the answer:

When president Ronald Reagan took office in 1981, his conservative friends
urged him to study the feasibility of returning to a gold standard as the only
way to curb government spending. It sounded like a reasonable alternative, so
President Reagan appointed a group of men called the Gold Commission to study
the situation and report back to Congress. What Reagan's Gold Commission
reported back to Congress in 1982 was the following shocking revelation
concerning gold. The US Treasury owned no gold at all.

All the gold that was left in Fort Knox was now owned by the Federal Reserve, a
group of private bankers, as collateral against the national debt.

The truth of the matter is that never before has so much money been stolen from
the hands of the general public and put into the hands of a small group of
private investors, the money changers”
