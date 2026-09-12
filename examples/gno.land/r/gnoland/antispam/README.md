# r/gnoland/antispam

Shared spam scoring for Gno realms. Pre-trained corpus, 400 keywords, 21 scam patterns, reputation tracking.

## Usage

```gno
import antispamr "gno.land/r/gnoland/antispam"
import engine "gno.land/p/gnoland/antispam"

rep := engine.ReputationData{
    AccountAgeDays: 30,
    Balance:        5000000000,
    HasUsername:    true,
}
result := antispamr.Score(author, content, rate, rep, nil, nil, nil, nil, engine.EarlyExitDisabled)

if result.Total >= engine.ThresholdReject {
    // reject (>= 8)
} else if result.Total >= engine.ThresholdHide {
    // hide (>= 5)
}

if result.Total < engine.ThresholdHide {
    antispamr.RecordAccepted(cross, author)
}
```

## API

Public (anyone can call):

```gno
Score(author, content, rate, rep, corpus, fps, dict, bl, earlyExitAt) SpamScore
GetReputation(addr) ReputationData
ReputationCount() int
TrustedCallerCount() int
```

Score() auto-populates FlaggedCount, BanCount, TotalAccepted from internal state. You only provide chain data (AccountAgeDays, Balance, HasUsername). Pass nil for state params to use shared instances. Read-only, doesn't update reputation.

Reputation recording (registered realms only):

```gno
RecordAccepted(realm, addr)
RecordFlag(realm, addr)
RecordBan(realm, addr)
```

Score() doesn't call these automatically. You record decisions manually:

```gno
if result.Total < engine.ThresholdHide {
    antispamr.RecordAccepted(cross, author)
} else if moderatorFlagged {
    antispamr.RecordFlag(cross, author)
}

if adminBanned {
    antispamr.RecordBan(cross, author)
}
```

Why manual? Full control, prevents false positives from poisoning data, allows human review.

Admin functions:

Setup:
- AdminLoadDefaults(realm) - load keywords and patterns post-deploy
- AdminSetAdmin(realm, newAdmin)

Reputation:
- AdminResetReputation(realm, addr)
- AdminRegisterCaller(realm, callerAddr) - allow realm to record reputation
- AdminRemoveCaller(realm, callerAddr)

Blocklist:
- AdminBlockAddress(realm, addr)
- AdminUnblockAddress(realm, addr)
- AdminAllowAddress(realm, addr) - bypass scoring
- AdminRemoveAllow(realm, addr)
- AdminAddPattern(realm, pattern)
- AdminRemovePattern(realm, pattern)

Keywords:
- AdminAddKeyword(realm, word, weight) - weight 1-3
- AdminRemoveKeyword(realm, word)
- AdminBulkAddKeywords(realm, data) - newline-separated "word:weight"

Training:
- AdminTrain(realm, content, isSpam)
- AdminAddSpamFingerprint(realm, content)

## Trust model

- Score/query: anyone (read-only)
- Record reputation: registered realms only
- Modify state: admin only
- Admin ops: admin only (uses OriginCaller, not PreviousRealm)

Security: reputation recording requires registration to prevent poisoning. Blocklisted addresses get 99 immediately. Allowlisted addresses skip all checks.

## State

The realm maintains:
- Bayesian corpus (pre-trained)
- Keyword dict (400 keywords, weights 1-3)
- Blocklist (21 scam patterns)
- Fingerprint store (MinHash)
- Reputation counters (TotalAccepted, FlaggedCount, BanCount)

State starts empty. Admin calls AdminLoadDefaults() post-deploy to populate keywords and patterns.

## Reputation counters

Per-address (no content stored):
- TotalAccepted: content accepted
- FlaggedCount: content flagged
- BanCount: times banned (permanent)

Used by:
- NEW_ACCOUNT (2 pts): under 1 day with TotalAccepted == 0
- BAD_REPUTATION (3 pts): FlaggedCount / TotalAccepted > 30%
- BANNED_BEFORE (1-3 pts): +1 per ban

## Gas optimization

Early exit:
```gno
result := antispamr.Score(author, content, rate, rep,
    nil, nil, nil, nil, engine.ThresholdReject)
```

User tiering:
```gno
rep := antispamr.GetReputation(author)

if rep.TotalAccepted > 10 && rep.FlaggedCount == 0 && chainData.AccountAgeDays > 30 {
    result = antispamr.Score(author, content, rate, chainData,
        nil, nil, nil, nil, engine.ThresholdReject)
} else {
    result = antispamr.Score(author, content, rate, chainData,
        nil, nil, nil, nil, engine.EarlyExitDisabled)
}
```

Custom state:
```gno
myBlocklist := engine.NewBlocklist()
myBlocklist.AddPattern(`my-pattern`)

result := antispamr.Score(author, content, rate, rep,
    nil, nil, nil, myBlocklist, engine.EarlyExitDisabled)
```

## Deployment

Realm starts empty. Post-deploy:

```gno
antispamr.AdminLoadDefaults(cross)
antispamr.AdminRegisterCaller(cross, boardsRealmAddr)
antispamr.AdminTrain(cross, "spam example", true)
antispamr.AdminTrain(cross, "ham example", false)
```

## Examples

Filetests:
- z1_score_demo: basic scoring + reputation recording
- z2_reputation_lifecycle: how reputation affects scores

See p/gnoland/antispam/README.md for architecture and rules.
