# Staker Reward

## Abstract

The **Staker** module handles the distribution of both **internal** (GNS emission) and **external** (user-provided) rewards to stakers:

- **[Internal rewards](https://docs.gnoswap.io/references/warm-up-periods)** (GNS emission) are allocated to “tiered” pools (tiers 1, 2, and 3). First, emission is split across the tiers according to the **TierRatio**. Then, within each tier, the emission is shared evenly among member pools and finally distributed proportionally to each staked position’s in-range liquidity.

- **[External rewards](https://docs.gnoswap.io/references/warm-up-periods)** (user-provided incentives) can be created for specific pools. Each external incentive emits a constant reward per block. Any user with in-range staked liquidity on that pool can claim a share of the reward, proportional to their staked liquidity.

- If, during a given block, no staked liquidity is in range, the internal emission is diverted to the community pool, and any external reward for that block is returned to the incentive creator.

- Every staked position has a designated [warmup schedule](https://docs.gnoswap.io/references/warm-up-periods). As it remains staked, the position progresses through multiple warmup periods. In each warmup period, a certain percentage of the reward is awarded to the position, and the remainder goes either to the community pool (for internal incentives) or is returned to the incentive creator (for external incentives).

## Main Reward Calculation Logic

Below is an example function that computes the rewards for a position. It does the following:

1. Caches any pending per-pool internal incentive rewards up to the current block.  
2. Retrieves or initializes the pool from `param.Pools`.  
3. Accumulates internal and external rewards (and the corresponding penalties) for each warmup period.  
4. Returns a list of rewards, one for each warmup period of the staked position.

```go
func CalcPositionReward(param CalcPositionRewardParam) []Reward {
	// cache per-pool rewards in the internal incentive(tiers)
	param.PoolTier.cacheReward(param.CurrentHeight, param.Pools)

	deposit := param.Deposits.Get(param.TokenId)
	poolPath := deposit.targetPoolPath

	pool, ok := param.Pools.Get(poolPath)
	if !ok {
		pool = NewPool(poolPath, param.CurrentHeight)
		param.Pools.Set(poolPath, pool)
	}

	lastCollectHeight := deposit.lastCollectHeight

	// Initialize arrays to hold reward & penalty data for each warmup
	internalRewards := make([]uint64, len(deposit.warmups))
	internalPenalties := make([]uint64, len(deposit.warmups))
	externalRewards := make([]map[string]uint64, len(deposit.warmups))
	externalPenalties := make([]map[string]uint64, len(deposit.warmups))

	if param.PoolTier.CurrentTier(poolPath) != 0 {
		// Internal incentivized pool
		internalRewards, internalPenalties = pool.RewardStateOf(deposit).CalculateInternalReward(lastCollectHeight, param.CurrentHeight)
	}

	// Retrieve all active external incentives from lastCollectHeight to CurrentHeight
	allIncentives := pool.incentives.GetAllInHeights(lastCollectHeight, param.CurrentHeight)

	for i := range externalRewards {
		externalRewards[i] = make(map[string]uint64)
		externalPenalties[i] = make(map[string]uint64)
	}

	for incentiveId, incentive := range allIncentives {
		// External incentivized pool
		externalReward, externalPenalty := pool.RewardStateOf(deposit).CalculateExternalReward(
			int64(lastCollectHeight),
			int64(param.CurrentHeight),
			incentive,
		)

		for i := range externalReward {
			externalRewards[i][incentiveId] = externalReward[i]
			externalPenalties[i][incentiveId] = externalPenalty[i]
		}
	}

	rewards := make([]Reward, len(internalRewards))
	for i := range internalRewards {
		rewards[i] = Reward{
			Internal:        internalRewards[i],
			InternalPenalty: internalPenalties[i],
			External:        externalRewards[i],
			ExternalPenalty: externalPenalties[i],
		}
	}

	return rewards
}
```

## TickCrossHook

`TickCrossHook` is triggered whenever a swap crosses an initialized tick. If any staked position uses that tick, the hook:

1. Updates the `stakedLiquidity` and `globalRewardRatioAccumulation`.
2. Sets the historical tick (for record-keeping and later calculations).
3. Depending on whether the total staked liquidity is now nonzero or zero, it begins or ends any unclaimable period.
4. Updates the `CurrentOutsideAccumulation` for the tick.

The variable `globalRewardRatioAccumulation` holds the integral of $\(f(h) = 1 \div \text{TotalStakedLiquidity}(h)\)$, but only when $\text{TotalStakedLiquidity}(h)$ is nonzero. Meanwhile, `CurrentOutsideAccumulation` tracks the same integral but over intervals where the pool tick is considered “outside” (i.e., on the opposite side of the current tick). When a tick cross occurs, this “outside” condition may flip, so the hook adjusts `CurrentOutsideAccumulation` by subtracting it from the latest `globalRewardRatioAccumulation`.

## Internal Reward

Internal rewards are distributed across tiers and then among pools. Each pool’s internal reward is determined as:

```math
\text{poolReward}(\mathrm{pool}) 
= \frac{\text{emission} \,\times\, \mathrm{TierRatio}\!\bigl(\mathrm{tier}(\mathrm{pool})\bigr)}
       {\mathrm{Count}\!\bigl(\mathrm{tier}(\mathrm{pool})\bigr)}.
```

The TierRatio is defined piecewise:

```math
\mathrm{TierRatio}(t) \;=\;
\begin{cases}
[1,\,0,\,0]_{\,t-1}, 
& \text{if } \mathrm{Count}(2) = 0 \;\land\; \mathrm{Count}(3) = 0, \\[8pt]
[0.8,\,0,\,0.2]_{\,t-1}, 
& \text{if } \mathrm{Count}(2) = 0, \\[8pt]
[0.7,\,0.3,\,0]_{\,t-1}, 
& \text{if } \mathrm{Count}(3) = 0, \\[8pt]
[0.5,\,0.3,\,0.2]_{\,t-1}, 
& \text{otherwise}.
\end{cases}
```

The total emission used by the staker contract is:

```math
\text{emission} 
= \mathrm{GNSEmissionPerSecond} 
  \times
  \Bigl(\frac{\mathrm{avgMsPerBlock}}{1000}\Bigr)
  \times
  \mathrm{StakerEmissionRatio}.
```

> **Note:**  
> - There is always at least one tier-1 pool.  
> - `GNSEmissionPerSecond` is a constant apart from any halving events (ignored here).  
> - When `avgMsPerBlock` or `StakerEmissionRatio` changes, a callback is triggered to cache rewards up to the current block and then update the emission rate. This also happens when a pool has its tier changed.

### How Internal Rewards Are Cached and Distributed

- **PoolTier.cacheReward** recalculates all reward-related data from the last cache height to the current block.  
  - **Halving blocks:** If there are halving events in this interval, the process “splits” the caching at each halving block, updates the staker emission accordingly, and continues.  
  - **Unclaimable period:** If the pool had no in-range stakers (currently in unclaimable state), it updates the unclaimable accumulation using the old emission rate, then starts a new period immediately so the future accumulation could be done based on the new emission rate.
  - The function finally updates the `GlobalRewardRatioAccumulation`, which is used later to compute each position’s rewards.

- After caching is updated to the current block, **CalculateInternalReward** computes the total claimable internal rewards for a position. Consider the following formulation, which handles multiple warmup intervals:

```math
\begin{aligned}
\mathrm{TotalRewardRatio}(s,e)
&=
  \sum_{i=0}^{m-1}
    \Bigl[
      \Delta\mathrm{Raw}\bigl(\alpha_i,\, \beta_i\bigr)
    \Bigr]
    \times
    r_i,
\\[6pt]
\alpha_i
&=
  \max\!\bigl(s,\, H_{i-1}\bigr),
\quad
\beta_i
=
  \min\!\bigl(e,\, H_{i}\bigr),
\\[6pt]
\Delta\mathrm{Raw}(a, b)
&=
  \mathrm{CalcRaw}(b)
  \;-\;
  \mathrm{CalcRaw}(a),
\\[6pt]
\mathrm{CalcRaw}(h)
&=
  \begin{cases}
    L(h) \;-\; U(h), 
      & \text{if } \mathrm{tick}(h) < \ell, \\[4pt]
    U(h) \;-\; L(h),
      & \text{if } \mathrm{tick}(h) \ge u, \\[4pt]
    G(h) \;-\; \bigl(L(h) + U(h)\bigr), 
      & \text{otherwise}.
  \end{cases}
\end{aligned}
```

where
- Each warmup interval $\(\bigl[H_{i-1},\,H_{i}\bigr]\)$ has a reward ratio $r_i$.
- $\alpha_i = \max(s,\, H_{i-1})$ and $\beta_i = \min(e,\, H_{i})$ slice the interval to fit $[s,e)$. If $\alpha_i \ge \beta_i$, that segment contributes zero.
- $\(L(h)\)$ = `tickLower.OutsideAccumulation(h)`
- $\(U(h)\)$ = `tickUpper.OutsideAccumulation(h)`
- $\(G(h)\)$ = `globalRewardRatioAccumulation(h)`
- $\(\ell\)$ = `tickLower.id`, $\(u\)$ = `tickUpper.id`

The final reward for a position is the sum of each applicable `TotalRewardRatio` multiplied by `poolReward` and `positionLiquidity`:
```math
\begin{aligned}
\text{finalReward}
&=
  \text{TotalRewardRatio} 
  \;\times\;
  \text{poolReward}
  \;\times\;
  \text{positionLiquidity}
\\[6pt]
&=
  \Bigl(
    \int_{s}^{e}
      \frac{1}{\mathrm{TotalStakedLiquidity}(h)}
    \, dh
  \Bigr)
  \;\times\;
  \text{poolReward}
  \;\times\;
  \text{positionLiquidity}
\\[6pt]
&=
  \int_{s}^{e}
    \frac{\text{poolReward}\;\times\;\text{positionLiquidity}}{\mathrm{TotalStakedLiquidity}(h)}
  \, dh.
\end{aligned}
```

## External Reward

External rewards emit a constant **reward per block** for their duration. To calculate the external reward for a specific incentive, we reuse the same approach of computing a `TotalRewardRatio` (similar to the internal reward method), but without any tier-based pooling or variable `poolReward`. Instead, we multiply the `TotalRewardRatio` by `ExternalIncentive.rewardPerBlock` and `positionLiquidity` for the relevant blocks.
