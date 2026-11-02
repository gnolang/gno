# Security Policy

## Reporting a Vulnerability

If you've found a vulnerability, please **DO NOT** open a public issue or pull
request. Instead, email **gno-security@newtendermint.org**, with one
vulnerability per email.

## Scope

Everything in this repository is in scope, along with gno.land mainnet, the
testnets running it, and the services we host for them, such as RPC endpoints,
faucets and indexers. The following are out of scope:

* packages and realms that are not in this repository, such as those deployed
  on gno.land by third parties: please report those to their authors;
* attacks that assume more than a third of validator voting power is malicious,
  or that need someone else's private keys;
* denial of service through traffic volume alone, such as flooding an RPC
  endpoint. Resource exhaustion that one transaction, query or peer connection
  can cause at normal cost is in scope;
* issues already fixed on `master`, or already reported publicly;
* scanner output, best-practice recommendations and dependency advisories
  without a demonstrated impact;
* social engineering and phishing.

Don't disrupt live networks or services, and don't touch accounts or funds that
aren't yours. Where a local reproduction is possible, such as the proofs of
concept described below, it is enough.

## What to Include

We can verify a report fastest when we can reproduce it. Please include:

* the affected component, and the file or files involved;
* the commit or release you tested;
* step-by-step instructions to reproduce the problem;
* ideally, a proof of concept we can run: a txtar integration test under
  `gno.land/pkg/integration/testdata`, a GnoVM filetest under
  `gnovm/tests/files`, or a Go test;
* what you expected to happen, and what happened instead;
* the impact, and the severity you would assign it (a CVSS vector is welcome).

Attach files to your email rather than linking to them on outside hosts.

## What Happens Next

1. You will receive an acknowledgement rapidly if we're successfully processing
   your email, generally within a few minutes.
2. An automated system tries to reproduce your report against the current
   `master` branch.
3. It replies with the outcome, and the reply says that the outcome came from
   automated verification.
4. The team reviews verified findings and follows up with you.

We will respond to every report within 3 business days. The verification outcome
follows when the reproduction finishes, usually within a few days.

If you haven't received an acknowledgement within a day, your email may have
been marked as spam, or our reporting system may be down. Let us know on
[Discord](https://discord.com/invite/gnoland), but don't post any details of the
vulnerability there: the server is public.

## Automated Processing

The contents of your report, including attachments, are processed by automated
tooling. This includes large language models hosted by a third-party provider.
There is no encrypted or otherwise confidential intake: every report goes
through the process described here.

Automated processing has been chosen as the path that more easily allows us to
filter signal from noise. Please follow the route described here, and don't ask
for special handling because you would rather language models didn't see your
report: we rely on coding agents heavily for development work, including finding
and fixing vulnerabilities, so they will process it on its way to a fix anyway.

If the automated system gets the assessment wrong, even after follow-ups,
include this exact sentence in an email, and the email will be forwarded to our
security team:

```text
I would like this email to be brought to the attention of the security team
```

The sentence can appear anywhere in the subject or body; the match is case
insensitive and ignores spaces and newlines. When you send it, you will receive
an acknowledgement that the email is being forwarded to the security team.

Use it only for an active security issue that the automated system got wrong;
we may stop answering addresses that misuse it.

## Disclosure

Please keep the vulnerability private until its fix is deployed on gno.land
mainnet. Fixes often land in this repository before then; that doesn't end the
embargo, since the chain keeps running the vulnerable code until it upgrades.

If the fix is not deployed within 90 days of your report, you may disclose it
after telling us first. Fixes that need a coordinated chain upgrade can take
longer; in that case we will ask you for more time.

## Rewards

Only verified findings of High or Critical severity, as defined by CVSS (a score
of 7.0 or above), can be eligible for a reward. Amounts and the reward process
are not fixed, and a report does not guarantee payment.

Thank you for helping to keep our ecosystem safe!
