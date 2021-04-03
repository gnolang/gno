package v036

import (
	"github.com/tendermint/go-amino-x"

	v034auth "github.com/tendermint/classic/sdk/x/auth/legacy/v0_34"
	v036auth "github.com/tendermint/classic/sdk/x/auth/legacy/v0_36"
	v034distr "github.com/tendermint/classic/sdk/x/distribution/legacy/v0_34"
	v036distr "github.com/tendermint/classic/sdk/x/distribution/legacy/v0_36"
	v034genAccounts "github.com/tendermint/classic/sdk/x/genaccounts/legacy/v0_34"
	v036genAccounts "github.com/tendermint/classic/sdk/x/genaccounts/legacy/v0_36"
	"github.com/tendermint/classic/sdk/x/genutil"
	v034gov "github.com/tendermint/classic/sdk/x/gov/legacy/v0_34"
	v036gov "github.com/tendermint/classic/sdk/x/gov/legacy/v0_36"
	v034staking "github.com/tendermint/classic/sdk/x/staking/legacy/v0_34"
	v036staking "github.com/tendermint/classic/sdk/x/staking/legacy/v0_36"
	v036supply "github.com/tendermint/classic/sdk/x/supply/legacy/v0_36"
)

// Migrate migrates exported state from v0.34 to a v0.36 genesis state.
func Migrate(appState genutil.AppMap) genutil.AppMap {

	// migrate genesis accounts state
	if appState[v034genAccounts.ModuleName] != nil {
		var genAccs v034genAccounts.GenesisState
		amino.MustUnmarshalJSON(appState[v034genAccounts.ModuleName], &genAccs)

		var authGenState v034auth.GenesisState
		amino.MustUnmarshalJSON(appState[v034auth.ModuleName], &authGenState)

		var govGenState v034gov.GenesisState
		amino.MustUnmarshalJSON(appState[v034gov.ModuleName], &govGenState)

		var distrGenState v034distr.GenesisState
		amino.MustUnmarshalJSON(appState[v034distr.ModuleName], &distrGenState)

		var stakingGenState v034staking.GenesisState
		amino.MustUnmarshalJSON(appState[v034staking.ModuleName], &stakingGenState)

		delete(appState, v034genAccounts.ModuleName) // delete old key in case the name changed
		appState[v036genAccounts.ModuleName] = amino.MustMarshalJSON(
			v036genAccounts.Migrate(
				genAccs, authGenState.CollectedFees, distrGenState.FeePool.CommunityPool, govGenState.Deposits,
				stakingGenState.Validators, stakingGenState.UnbondingDelegations, distrGenState.OutstandingRewards,
				stakingGenState.Params.BondDenom, v036distr.ModuleName, v036gov.ModuleName,
			),
		)
	}

	// migrate auth state
	if appState[v034auth.ModuleName] != nil {
		var authGenState v034auth.GenesisState
		amino.MustUnmarshalJSON(appState[v034auth.ModuleName], &authGenState)

		delete(appState, v034auth.ModuleName) // delete old key in case the name changed
		appState[v036auth.ModuleName] = amino.MustMarshalJSON(v036auth.Migrate(authGenState))
	}

	// migrate gov state
	if appState[v034gov.ModuleName] != nil {
		var govGenState v034gov.GenesisState
		amino.MustUnmarshalJSON(appState[v034gov.ModuleName], &govGenState)

		delete(appState, v034gov.ModuleName) // delete old key in case the name changed
		appState[v036gov.ModuleName] = amino.MustMarshalJSON(v036gov.Migrate(govGenState))
	}

	// migrate distribution state
	if appState[v034distr.ModuleName] != nil {
		var slashingGenState v034distr.GenesisState
		amino.MustUnmarshalJSON(appState[v034distr.ModuleName], &slashingGenState)

		delete(appState, v034distr.ModuleName) // delete old key in case the name changed
		appState[v036distr.ModuleName] = amino.MustMarshalJSON(v036distr.Migrate(slashingGenState))
	}

	// migrate staking state
	if appState[v034staking.ModuleName] != nil {
		var stakingGenState v034staking.GenesisState
		amino.MustUnmarshalJSON(appState[v034staking.ModuleName], &stakingGenState)

		delete(appState, v034staking.ModuleName) // delete old key in case the name changed
		appState[v036staking.ModuleName] = amino.MustMarshalJSON(v036staking.Migrate(stakingGenState))
	}

	// migrate supply state
	appState[v036supply.ModuleName] = amino.MustMarshalJSON(v036supply.EmptyGenesisState())

	return appState
}
