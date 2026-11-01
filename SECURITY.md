# Security Policy

## Reporting a Vulnerability

If you've found a vulnerability, please **DO NOT** open a public issue or pull
request. Instead, email **gno-security@newtendermint.org**, with one
vulnerability per email.

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
   your email - generally a few minutes.
2. An automated system tries to reproduce your report against the current
   `master` branch.
3. It replies with the outcome, and the reply says that the outcome came from
   automated verification.
4. The team reviews verified findings and follows up with you.

We will respond to every report within 3 business days. The verification outcome
follows when the reproduction finishes, usually within a few days.

## Automated Processing

The contents of your report, including attachments, are processed by automated
tooling. This includes large language models hosted by a third-party provider.

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

Misuse of this feature will have us block your email address; DO NOT use it as
a privileged way to contact the team about anything that is not an active
security issue.

## Rewards

Only verified findings of High or Critical severity, as defined by CVSS (a score
of 7.0 or above), can be eligible for a reward. Amounts and the reward process
are not fixed, and a report does not guarantee payment.

Thank you for helping to keep our ecosystem safe!
