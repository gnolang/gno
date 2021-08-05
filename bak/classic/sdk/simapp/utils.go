//nolint
package simapp

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"time"

	"github.com/tendermint/classic/crypto"
	"github.com/tendermint/classic/crypto/secp256k1"
	dbm "github.com/tendermint/classic/db"
	cmn "github.com/tendermint/classic/libs/common"
	"github.com/tendermint/classic/libs/log"
	tmtypes "github.com/tendermint/classic/types"
	"github.com/tendermint/go-amino-x"

	"github.com/tendermint/classic/sdk/baseapp"
	sdk "github.com/tendermint/classic/sdk/types"
	"github.com/tendermint/classic/sdk/x/auth"
	"github.com/tendermint/classic/sdk/x/bank"
	"github.com/tendermint/classic/sdk/x/distribution"
	"github.com/tendermint/classic/sdk/x/genaccounts"
	"github.com/tendermint/classic/sdk/x/gov"
	"github.com/tendermint/classic/sdk/x/mint"
	"github.com/tendermint/classic/sdk/x/simulation"
	"github.com/tendermint/classic/sdk/x/slashing"
	"github.com/tendermint/classic/sdk/x/staking"
	"github.com/tendermint/classic/sdk/x/supply"
)

// List of available flags for the simulator
var (
	genesisFile        string
	paramsFile         string
	exportParamsPath   string
	exportParamsHeight int
	exportStatePath    string
	exportStatsPath    string
	seed               int64
	initialBlockHeight int
	numBlocks          int
	blockSize          int
	enabled            bool
	verbose            bool
	lean               bool
	commit             bool
	period             int
	onOperation        bool // TODO Remove in favor of binary search for invariant violation
	allInvariants      bool
	genesisTime        int64
)

// NewSimAppUNSAFE is used for debugging purposes only.
//
// NOTE: to not use this function with non-test code
func NewSimAppUNSAFE(logger log.Logger, db dbm.DB, traceStore io.Writer, loadLatest bool,
	invCheckPeriod uint, baseAppOptions ...func(*baseapp.BaseApp),
) (gapp *SimApp, keyMain, keyStaking *sdk.KVStoreKey, stakingKeeper staking.Keeper) {

	gapp = NewSimApp(logger, db, traceStore, loadLatest, invCheckPeriod, baseAppOptions...)
	return gapp, gapp.keys[baseapp.MainStoreKey], gapp.keys[staking.StoreKey], gapp.stakingKeeper
}

// AppStateFromGenesisFileFn util function to generate the genesis AppState
// from a genesis.json file
func AppStateFromGenesisFileFn(r *rand.Rand) (tmtypes.GenesisDoc, []simulation.Account) {
	var genesis tmtypes.GenesisDoc

	bytes, err := ioutil.ReadFile(genesisFile)
	if err != nil {
		panic(err)
	}

	amino.MustUnmarshalJSON(bytes, &genesis)

	var appState GenesisState
	amino.MustUnmarshalJSON(genesis.AppState, &appState)

	accounts := genaccounts.GetGenesisStateFromAppState(appState)

	newAccs := make([]simulation.Account, len(accounts))
	for i, acc := range accounts {
		// Pick a random private key, since we don't know the actual key
		// This should be fine as it's only used for mock Tendermint validators
		// and these keys are never actually used to sign by mock Tendermint.
		privkeySeed := make([]byte, 15)
		r.Read(privkeySeed)

		privKey := secp256k1.GenPrivKeySecp256k1(privkeySeed)
		newAccs[i] = simulation.Account{privKey, privKey.PubKey(), acc.Address}
	}

	return genesis, newAccs
}

// GenAuthGenesisState generates a random GenesisState for auth
func GenAuthGenesisState(r *rand.Rand, ap simulation.AppParams, genesisState map[string]json.RawMessage) {
	authGenesis := auth.NewGenesisState(
		auth.NewParams(
			func(r *rand.Rand) uint64 {
				var v uint64
				ap.GetOrGenerate(simulation.MaxMemoChars, &v, r,
					func(r *rand.Rand) {
						v = simulation.ModuleParamSimulator[simulation.MaxMemoChars](r).(uint64)
					})
				return v
			}(r),
			func(r *rand.Rand) uint64 {
				var v uint64
				ap.GetOrGenerate(simulation.TxSigLimit, &v, r,
					func(r *rand.Rand) {
						v = simulation.ModuleParamSimulator[simulation.TxSigLimit](r).(uint64)
					})
				return v
			}(r),
			func(r *rand.Rand) uint64 {
				var v uint64
				ap.GetOrGenerate(simulation.TxSizeCostPerByte, &v, r,
					func(r *rand.Rand) {
						v = simulation.ModuleParamSimulator[simulation.TxSizeCostPerByte](r).(uint64)
					})
				return v
			}(r),
			func(r *rand.Rand) uint64 {
				var v uint64
				ap.GetOrGenerate(simulation.SigVerifyCostED25519, &v, r,
					func(r *rand.Rand) {
						v = simulation.ModuleParamSimulator[simulation.SigVerifyCostED25519](r).(uint64)
					})
				return v
			}(r),
			func(r *rand.Rand) uint64 {
				var v uint64
				ap.GetOrGenerate(simulation.SigVerifyCostSECP256K1, &v, r,
					func(r *rand.Rand) {
						v = simulation.ModuleParamSimulator[simulation.SigVerifyCostSECP256K1](r).(uint64)
					})
				return v
			}(r),
		),
	)

	fmt.Printf("Selected randomly generated auth parameters:\n%s\n", amino.MustMarshalJSONIndent(authGenesis.Params))
	genesisState[auth.ModuleName] = amino.MustMarshalJSON(authGenesis)
}

// GenBankGenesisState generates a random GenesisState for bank
func GenBankGenesisState(r *rand.Rand, ap simulation.AppParams, genesisState map[string]json.RawMessage) {
	bankGenesis := bank.NewGenesisState(
		func(r *rand.Rand) bool {
			var v bool
			ap.GetOrGenerate(simulation.SendEnabled, &v, r,
				func(r *rand.Rand) {
					v = simulation.ModuleParamSimulator[simulation.SendEnabled](r).(bool)
				})
			return v
		}(r),
	)

	fmt.Printf("Selected randomly generated bank parameters:\n%s\n", amino.MustMarshalJSONIndent(bankGenesis))
	genesisState[bank.ModuleName] = amino.MustMarshalJSON(bankGenesis)
}

// GenSupplyGenesisState generates a random GenesisState for supply
func GenSupplyGenesisState(amount, numInitiallyBonded, numAccs int64, genesisState map[string]json.RawMessage) {
	totalSupply := sdk.NewInt(amount * (numAccs + numInitiallyBonded))
	supplyGenesis := supply.NewGenesisState(
		sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, totalSupply)),
	)

	fmt.Printf("Generated supply parameters:\n%s\n", amino.MustMarshalJSONIndent(supplyGenesis))
	genesisState[supply.ModuleName] = amino.MustMarshalJSON(supplyGenesis)
}

// GenGenesisAccounts generates a random GenesisState for the genesis accounts
func GenGenesisAccounts(
	r *rand.Rand, accs []simulation.Account,
	genesisTimestamp time.Time, amount, numInitiallyBonded int64,
	genesisState map[string]json.RawMessage,
) {

	var genesisAccounts []genaccounts.GenesisAccount

	// randomly generate some genesis accounts
	for i, acc := range accs {
		coins := sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(amount))}
		bacc := auth.NewBaseAccountWithAddress(acc.Address)
		bacc.SetCoins(coins)

		var gacc genaccounts.GenesisAccount

		// Only consider making a vesting account once the initial bonded validator
		// set is exhausted due to needing to track DelegatedVesting.
		if int64(i) > numInitiallyBonded && r.Intn(100) < 50 {
			var (
				vacc    auth.VestingAccount
				endTime int64
			)

			startTime := genesisTimestamp.Unix()

			// Allow for some vesting accounts to vest very quickly while others very slowly.
			if r.Intn(100) < 50 {
				endTime = int64(simulation.RandIntBetween(r, int(startTime), int(startTime+(60*60*24*30))))
			} else {
				endTime = int64(simulation.RandIntBetween(r, int(startTime), int(startTime+(60*60*12))))
			}

			if startTime == endTime {
				endTime++
			}

			if r.Intn(100) < 50 {
				vacc = auth.NewContinuousVestingAccount(&bacc, startTime, endTime)
			} else {
				vacc = auth.NewDelayedVestingAccount(&bacc, endTime)
			}

			var err error
			gacc, err = genaccounts.NewGenesisAccountI(vacc)
			if err != nil {
				panic(err)
			}
		} else {
			gacc = genaccounts.NewGenesisAccount(&bacc)
		}

		genesisAccounts = append(genesisAccounts, gacc)
	}

	genesisState[genaccounts.ModuleName] = amino.MustMarshalJSON(genesisAccounts)
}

// GenGovGenesisState generates a random GenesisState for gov
func GenGovGenesisState(r *rand.Rand, ap simulation.AppParams, genesisState map[string]json.RawMessage) {
	var vp time.Duration
	ap.GetOrGenerate(simulation.VotingParamsVotingPeriod, &vp, r,
		func(r *rand.Rand) {
			vp = simulation.ModuleParamSimulator[simulation.VotingParamsVotingPeriod](r).(time.Duration)
		})

	govGenesis := gov.NewGenesisState(
		uint64(r.Intn(100)),
		gov.NewDepositParams(
			func(r *rand.Rand) sdk.Coins {
				var v sdk.Coins
				ap.GetOrGenerate(simulation.DepositParamsMinDeposit, &v, r,
					func(r *rand.Rand) {
						v = simulation.ModuleParamSimulator[simulation.DepositParamsMinDeposit](r).(sdk.Coins)
					})
				return v
			}(r),
			vp,
		),
		gov.NewVotingParams(vp),
		gov.NewTallyParams(
			func(r *rand.Rand) sdk.Dec {
				var v sdk.Dec
				ap.GetOrGenerate(simulation.TallyParamsQuorum, &v, r,
					func(r *rand.Rand) {
						v = simulation.ModuleParamSimulator[simulation.TallyParamsQuorum](r).(sdk.Dec)
					})
				return v
			}(r),
			func(r *rand.Rand) sdk.Dec {
				var v sdk.Dec
				ap.GetOrGenerate(simulation.TallyParamsThreshold, &v, r,
					func(r *rand.Rand) {
						v = simulation.ModuleParamSimulator[simulation.TallyParamsThreshold](r).(sdk.Dec)
					})
				return v
			}(r),
			func(r *rand.Rand) sdk.Dec {
				var v sdk.Dec
				ap.GetOrGenerate(simulation.TallyParamsVeto, &v, r,
					func(r *rand.Rand) {
						v = simulation.ModuleParamSimulator[simulation.TallyParamsVeto](r).(sdk.Dec)
					})
				return v
			}(r),
		),
	)

	fmt.Printf("Selected randomly generated governance parameters:\n%s\n", amino.MustMarshalJSONIndent(govGenesis))
	genesisState[gov.ModuleName] = amino.MustMarshalJSON(govGenesis)
}

// GenMintGenesisState generates a random GenesisState for mint
func GenMintGenesisState(r *rand.Rand, ap simulation.AppParams, genesisState map[string]json.RawMessage) {
	mintGenesis := mint.NewGenesisState(
		mint.InitialMinter(
			func(r *rand.Rand) sdk.Dec {
				var v sdk.Dec
				ap.GetOrGenerate(simulation.Inflation, &v, r,
					func(r *rand.Rand) {
						v = simulation.ModuleParamSimulator[simulation.Inflation](r).(sdk.Dec)
					})
				return v
			}(r),
		),
		mint.NewParams(
			sdk.DefaultBondDenom,
			func(r *rand.Rand) sdk.Dec {
				var v sdk.Dec
				ap.GetOrGenerate(simulation.InflationRateChange, &v, r,
					func(r *rand.Rand) {
						v = simulation.ModuleParamSimulator[simulation.InflationRateChange](r).(sdk.Dec)
					})
				return v
			}(r),
			func(r *rand.Rand) sdk.Dec {
				var v sdk.Dec
				ap.GetOrGenerate(simulation.InflationMax, &v, r,
					func(r *rand.Rand) {
						v = simulation.ModuleParamSimulator[simulation.InflationMax](r).(sdk.Dec)
					})
				return v
			}(r),
			func(r *rand.Rand) sdk.Dec {
				var v sdk.Dec
				ap.GetOrGenerate(simulation.InflationMin, &v, r,
					func(r *rand.Rand) {
						v = simulation.ModuleParamSimulator[simulation.InflationMin](r).(sdk.Dec)
					})
				return v
			}(r),
			func(r *rand.Rand) sdk.Dec {
				var v sdk.Dec
				ap.GetOrGenerate(simulation.GoalBonded, &v, r,
					func(r *rand.Rand) {
						v = simulation.ModuleParamSimulator[simulation.GoalBonded](r).(sdk.Dec)
					})
				return v
			}(r),
			uint64(60*60*8766/5),
		),
	)

	fmt.Printf("Selected randomly generated minting parameters:\n%s\n", amino.MustMarshalJSONIndent(mintGenesis.Params))
	genesisState[mint.ModuleName] = amino.MustMarshalJSON(mintGenesis)
}

// GenDistrGenesisState generates a random GenesisState for distribution
func GenDistrGenesisState(r *rand.Rand, ap simulation.AppParams, genesisState map[string]json.RawMessage) {
	distrGenesis := distribution.GenesisState{
		FeePool: distribution.InitialFeePool(),
		CommunityTax: func(r *rand.Rand) sdk.Dec {
			var v sdk.Dec
			ap.GetOrGenerate(simulation.CommunityTax, &v, r,
				func(r *rand.Rand) {
					v = simulation.ModuleParamSimulator[simulation.CommunityTax](r).(sdk.Dec)
				})
			return v
		}(r),
		BaseProposerReward: func(r *rand.Rand) sdk.Dec {
			var v sdk.Dec
			ap.GetOrGenerate(simulation.BaseProposerReward, &v, r,
				func(r *rand.Rand) {
					v = simulation.ModuleParamSimulator[simulation.BaseProposerReward](r).(sdk.Dec)
				})
			return v
		}(r),
		BonusProposerReward: func(r *rand.Rand) sdk.Dec {
			var v sdk.Dec
			ap.GetOrGenerate(simulation.BonusProposerReward, &v, r,
				func(r *rand.Rand) {
					v = simulation.ModuleParamSimulator[simulation.BonusProposerReward](r).(sdk.Dec)
				})
			return v
		}(r),
	}

	fmt.Printf("Selected randomly generated distribution parameters:\n%s\n", amino.MustMarshalJSONIndent(distrGenesis))
	genesisState[distribution.ModuleName] = amino.MustMarshalJSON(distrGenesis)
}

// GenSlashingGenesisState generates a random GenesisState for slashing
func GenSlashingGenesisState(
	r *rand.Rand, stakingGen staking.GenesisState,
	ap simulation.AppParams, genesisState map[string]json.RawMessage,
) {
	slashingGenesis := slashing.NewGenesisState(
		slashing.NewParams(
			stakingGen.Params.UnbondingTime,
			func(r *rand.Rand) int64 {
				var v int64
				ap.GetOrGenerate(simulation.SignedBlocksWindow, &v, r,
					func(r *rand.Rand) {
						v = simulation.ModuleParamSimulator[simulation.SignedBlocksWindow](r).(int64)
					})
				return v
			}(r),
			func(r *rand.Rand) sdk.Dec {
				var v sdk.Dec
				ap.GetOrGenerate(simulation.MinSignedPerWindow, &v, r,
					func(r *rand.Rand) {
						v = simulation.ModuleParamSimulator[simulation.MinSignedPerWindow](r).(sdk.Dec)
					})
				return v
			}(r),
			func(r *rand.Rand) time.Duration {
				var v time.Duration
				ap.GetOrGenerate(simulation.DowntimeJailDuration, &v, r,
					func(r *rand.Rand) {
						v = simulation.ModuleParamSimulator[simulation.DowntimeJailDuration](r).(time.Duration)
					})
				return v
			}(r),
			func(r *rand.Rand) sdk.Dec {
				var v sdk.Dec
				ap.GetOrGenerate(simulation.SlashFractionDoubleSign, &v, r,
					func(r *rand.Rand) {
						v = simulation.ModuleParamSimulator[simulation.SlashFractionDoubleSign](r).(sdk.Dec)
					})
				return v
			}(r),
			func(r *rand.Rand) sdk.Dec {
				var v sdk.Dec
				ap.GetOrGenerate(simulation.SlashFractionDowntime, &v, r,
					func(r *rand.Rand) {
						v = simulation.ModuleParamSimulator[simulation.SlashFractionDowntime](r).(sdk.Dec)
					})
				return v
			}(r),
		),
		nil,
		nil,
	)

	fmt.Printf("Selected randomly generated slashing parameters:\n%s\n", amino.MustMarshalJSONIndent(slashingGenesis.Params))
	genesisState[slashing.ModuleName] = amino.MustMarshalJSON(slashingGenesis)
}

// GenStakingGenesisState generates a random GenesisState for staking
func GenStakingGenesisState(
	r *rand.Rand, accs []simulation.Account, amount, numAccs, numInitiallyBonded int64,
	ap simulation.AppParams, genesisState map[string]json.RawMessage,
) staking.GenesisState {

	stakingGenesis := staking.NewGenesisState(
		staking.NewParams(
			func(r *rand.Rand) time.Duration {
				var v time.Duration
				ap.GetOrGenerate(simulation.UnbondingTime, &v, r,
					func(r *rand.Rand) {
						v = simulation.ModuleParamSimulator[simulation.UnbondingTime](r).(time.Duration)
					})
				return v
			}(r),
			func(r *rand.Rand) uint16 {
				var v uint16
				ap.GetOrGenerate(simulation.MaxValidators, &v, r,
					func(r *rand.Rand) {
						v = simulation.ModuleParamSimulator[simulation.MaxValidators](r).(uint16)
					})
				return v
			}(r),
			7,
			sdk.DefaultBondDenom,
		),
		nil,
		nil,
	)

	var (
		validators  []staking.Validator
		delegations []staking.Delegation
	)

	valAddrs := make([]sdk.ValAddress, numInitiallyBonded)
	for i := 0; i < int(numInitiallyBonded); i++ {
		valAddr := sdk.ValAddress(accs[i].Address)
		valAddrs[i] = valAddr

		validator := staking.NewValidator(valAddr, accs[i].PubKey, staking.Description{})
		validator.Tokens = sdk.NewInt(amount)
		validator.DelegatorShares = sdk.NewDec(amount)
		delegation := staking.NewDelegation(accs[i].Address, valAddr, sdk.NewDec(amount))
		validators = append(validators, validator)
		delegations = append(delegations, delegation)
	}

	stakingGenesis.Validators = validators
	stakingGenesis.Delegations = delegations

	fmt.Printf("Selected randomly generated staking parameters:\n%s\n", amino.MustMarshalJSONIndent(stakingGenesis.Params))
	genesisState[staking.ModuleName] = amino.MustMarshalJSON(stakingGenesis)

	return stakingGenesis
}

// GetSimulationLog unmarshals the KVPair's Value to the corresponding type based on the
// each's module store key and the prefix bytes of the KVPair's key.
func GetSimulationLog(storeName string, kvA, kvB cmn.KVPair) (log string) {
	log = fmt.Sprintf("store A %X => %X\nstore B %X => %X\n", kvA.Key, kvA.Value, kvB.Key, kvB.Value)

	if len(kvA.Value) == 0 && len(kvB.Value) == 0 {
		return
	}

	switch storeName {
	case auth.StoreKey:
		return DecodeAccountStore(kvA, kvB)
	case mint.StoreKey:
		return DecodeMintStore(kvA, kvB)
	case staking.StoreKey:
		return DecodeStakingStore(kvA, kvB)
	case slashing.StoreKey:
		return DecodeSlashingStore(kvA, kvB)
	case gov.StoreKey:
		return DecodeGovStore(kvA, kvB)
	case distribution.StoreKey:
		return DecodeDistributionStore(kvA, kvB)
	case supply.StoreKey:
		return DecodeSupplyStore(kvA, kvB)
	default:
		return
	}
}

// DecodeAccountStore unmarshals the KVPair's Value to the corresponding auth type
func DecodeAccountStore(kvA, kvB cmn.KVPair) string {
	switch {
	case bytes.Equal(kvA.Key[:1], auth.AddressStoreKeyPrefix):
		var accA, accB auth.Account
		amino.MustUnmarshal(kvA.Value, &accA)
		amino.MustUnmarshal(kvB.Value, &accB)
		return fmt.Sprintf("%v\n%v", accA, accB)
	case bytes.Equal(kvA.Key, auth.GlobalAccountNumberKey):
		var globalAccNumberA, globalAccNumberB uint64
		amino.MustUnmarshalSized(kvA.Value, &globalAccNumberA)
		amino.MustUnmarshalSized(kvB.Value, &globalAccNumberB)
		return fmt.Sprintf("GlobalAccNumberA: %d\nGlobalAccNumberB: %d", globalAccNumberA, globalAccNumberB)
	default:
		panic(fmt.Sprintf("invalid account key %X", kvA.Key))
	}
}

// DecodeMintStore unmarshals the KVPair's Value to the corresponding mint type
func DecodeMintStore(kvA, kvB cmn.KVPair) string {
	switch {
	case bytes.Equal(kvA.Key, mint.MinterKey):
		var minterA, minterB mint.Minter
		amino.MustUnmarshalSized(kvA.Value, &minterA)
		amino.MustUnmarshalSized(kvB.Value, &minterB)
		return fmt.Sprintf("%v\n%v", minterA, minterB)
	default:
		panic(fmt.Sprintf("invalid mint key %X", kvA.Key))
	}
}

// DecodeDistributionStore unmarshals the KVPair's Value to the corresponding distribution type
func DecodeDistributionStore(kvA, kvB cmn.KVPair) string {
	switch {
	case bytes.Equal(kvA.Key[:1], distribution.FeePoolKey):
		var feePoolA, feePoolB distribution.FeePool
		amino.MustUnmarshalSized(kvA.Value, &feePoolA)
		amino.MustUnmarshalSized(kvB.Value, &feePoolB)
		return fmt.Sprintf("%v\n%v", feePoolA, feePoolB)

	case bytes.Equal(kvA.Key[:1], distribution.ProposerKey):
		return fmt.Sprintf("%v\n%v", sdk.ConsAddress(kvA.Value), sdk.ConsAddress(kvB.Value))

	case bytes.Equal(kvA.Key[:1], distribution.ValidatorOutstandingRewardsPrefix):
		var rewardsA, rewardsB distribution.ValidatorOutstandingRewards
		amino.MustUnmarshalSized(kvA.Value, &rewardsA)
		amino.MustUnmarshalSized(kvB.Value, &rewardsB)
		return fmt.Sprintf("%v\n%v", rewardsA, rewardsB)

	case bytes.Equal(kvA.Key[:1], distribution.DelegatorWithdrawAddrPrefix):
		return fmt.Sprintf("%v\n%v", sdk.AccAddress(kvA.Value), sdk.AccAddress(kvB.Value))

	case bytes.Equal(kvA.Key[:1], distribution.DelegatorStartingInfoPrefix):
		var infoA, infoB distribution.DelegatorStartingInfo
		amino.MustUnmarshalSized(kvA.Value, &infoA)
		amino.MustUnmarshalSized(kvB.Value, &infoB)
		return fmt.Sprintf("%v\n%v", infoA, infoB)

	case bytes.Equal(kvA.Key[:1], distribution.ValidatorHistoricalRewardsPrefix):
		var rewardsA, rewardsB distribution.ValidatorHistoricalRewards
		amino.MustUnmarshalSized(kvA.Value, &rewardsA)
		amino.MustUnmarshalSized(kvB.Value, &rewardsB)
		return fmt.Sprintf("%v\n%v", rewardsA, rewardsB)

	case bytes.Equal(kvA.Key[:1], distribution.ValidatorCurrentRewardsPrefix):
		var rewardsA, rewardsB distribution.ValidatorCurrentRewards
		amino.MustUnmarshalSized(kvA.Value, &rewardsA)
		amino.MustUnmarshalSized(kvB.Value, &rewardsB)
		return fmt.Sprintf("%v\n%v", rewardsA, rewardsB)

	case bytes.Equal(kvA.Key[:1], distribution.ValidatorAccumulatedCommissionPrefix):
		var commissionA, commissionB distribution.ValidatorAccumulatedCommission
		amino.MustUnmarshalSized(kvA.Value, &commissionA)
		amino.MustUnmarshalSized(kvB.Value, &commissionB)
		return fmt.Sprintf("%v\n%v", commissionA, commissionB)

	case bytes.Equal(kvA.Key[:1], distribution.ValidatorSlashEventPrefix):
		var eventA, eventB distribution.ValidatorSlashEvent
		amino.MustUnmarshalSized(kvA.Value, &eventA)
		amino.MustUnmarshalSized(kvB.Value, &eventB)
		return fmt.Sprintf("%v\n%v", eventA, eventB)

	default:
		panic(fmt.Sprintf("invalid distribution key prefix %X", kvA.Key[:1]))
	}
}

// DecodeStakingStore unmarshals the KVPair's Value to the corresponding staking type
func DecodeStakingStore(kvA, kvB cmn.KVPair) string {
	switch {
	case bytes.Equal(kvA.Key[:1], staking.LastTotalPowerKey):
		var powerA, powerB sdk.Int
		amino.MustUnmarshalSized(kvA.Value, &powerA)
		amino.MustUnmarshalSized(kvB.Value, &powerB)
		return fmt.Sprintf("%v\n%v", powerA, powerB)

	case bytes.Equal(kvA.Key[:1], staking.ValidatorsKey):
		var validatorA, validatorB staking.Validator
		amino.MustUnmarshalSized(kvA.Value, &validatorA)
		amino.MustUnmarshalSized(kvB.Value, &validatorB)
		return fmt.Sprintf("%v\n%v", validatorA, validatorB)

	case bytes.Equal(kvA.Key[:1], staking.LastValidatorPowerKey),
		bytes.Equal(kvA.Key[:1], staking.ValidatorsByConsAddrKey),
		bytes.Equal(kvA.Key[:1], staking.ValidatorsByPowerIndexKey):
		return fmt.Sprintf("%v\n%v", sdk.ValAddress(kvA.Value), sdk.ValAddress(kvB.Value))

	case bytes.Equal(kvA.Key[:1], staking.DelegationKey):
		var delegationA, delegationB staking.Delegation
		amino.MustUnmarshalSized(kvA.Value, &delegationA)
		amino.MustUnmarshalSized(kvB.Value, &delegationB)
		return fmt.Sprintf("%v\n%v", delegationA, delegationB)

	case bytes.Equal(kvA.Key[:1], staking.UnbondingDelegationKey),
		bytes.Equal(kvA.Key[:1], staking.UnbondingDelegationByValIndexKey):
		var ubdA, ubdB staking.UnbondingDelegation
		amino.MustUnmarshalSized(kvA.Value, &ubdA)
		amino.MustUnmarshalSized(kvB.Value, &ubdB)
		return fmt.Sprintf("%v\n%v", ubdA, ubdB)

	case bytes.Equal(kvA.Key[:1], staking.RedelegationKey),
		bytes.Equal(kvA.Key[:1], staking.RedelegationByValSrcIndexKey):
		var redA, redB staking.Redelegation
		amino.MustUnmarshalSized(kvA.Value, &redA)
		amino.MustUnmarshalSized(kvB.Value, &redB)
		return fmt.Sprintf("%v\n%v", redA, redB)

	default:
		panic(fmt.Sprintf("invalid staking key prefix %X", kvA.Key[:1]))
	}
}

// DecodeSlashingStore unmarshals the KVPair's Value to the corresponding slashing type
func DecodeSlashingStore(kvA, kvB cmn.KVPair) string {
	switch {
	case bytes.Equal(kvA.Key[:1], slashing.ValidatorSigningInfoKey):
		var infoA, infoB slashing.ValidatorSigningInfo
		amino.MustUnmarshalSized(kvA.Value, &infoA)
		amino.MustUnmarshalSized(kvB.Value, &infoB)
		return fmt.Sprintf("%v\n%v", infoA, infoB)

	case bytes.Equal(kvA.Key[:1], slashing.ValidatorMissedBlockBitArrayKey):
		var missedA, missedB bool
		amino.MustUnmarshalSized(kvA.Value, &missedA)
		amino.MustUnmarshalSized(kvB.Value, &missedB)
		return fmt.Sprintf("missedA: %v\nmissedB: %v", missedA, missedB)

	case bytes.Equal(kvA.Key[:1], slashing.AddrPubkeyRelationKey):
		var pubKeyA, pubKeyB crypto.PubKey
		amino.MustUnmarshalSized(kvA.Value, &pubKeyA)
		amino.MustUnmarshalSized(kvB.Value, &pubKeyB)
		bechPKA := sdk.MustBech32ifyAccPub(pubKeyA)
		bechPKB := sdk.MustBech32ifyAccPub(pubKeyB)
		return fmt.Sprintf("PubKeyA: %s\nPubKeyB: %s", bechPKA, bechPKB)

	default:
		panic(fmt.Sprintf("invalid slashing key prefix %X", kvA.Key[:1]))
	}
}

// DecodeGovStore unmarshals the KVPair's Value to the corresponding gov type
func DecodeGovStore(kvA, kvB cmn.KVPair) string {
	switch {
	case bytes.Equal(kvA.Key[:1], gov.ProposalsKeyPrefix):
		var proposalA, proposalB gov.Proposal
		amino.MustUnmarshalSized(kvA.Value, &proposalA)
		amino.MustUnmarshalSized(kvB.Value, &proposalB)
		return fmt.Sprintf("%v\n%v", proposalA, proposalB)

	case bytes.Equal(kvA.Key[:1], gov.ActiveProposalQueuePrefix),
		bytes.Equal(kvA.Key[:1], gov.InactiveProposalQueuePrefix),
		bytes.Equal(kvA.Key[:1], gov.ProposalIDKey):
		proposalIDA := binary.LittleEndian.Uint64(kvA.Value)
		proposalIDB := binary.LittleEndian.Uint64(kvB.Value)
		return fmt.Sprintf("proposalIDA: %d\nProposalIDB: %d", proposalIDA, proposalIDB)

	case bytes.Equal(kvA.Key[:1], gov.DepositsKeyPrefix):
		var depositA, depositB gov.Deposit
		amino.MustUnmarshalSized(kvA.Value, &depositA)
		amino.MustUnmarshalSized(kvB.Value, &depositB)
		return fmt.Sprintf("%v\n%v", depositA, depositB)

	case bytes.Equal(kvA.Key[:1], gov.VotesKeyPrefix):
		var voteA, voteB gov.Vote
		amino.MustUnmarshalSized(kvA.Value, &voteA)
		amino.MustUnmarshalSized(kvB.Value, &voteB)
		return fmt.Sprintf("%v\n%v", voteA, voteB)

	default:
		panic(fmt.Sprintf("invalid governance key prefix %X", kvA.Key[:1]))
	}
}

// DecodeSupplyStore unmarshals the KVPair's Value to the corresponding supply type
func DecodeSupplyStore(kvA, kvB cmn.KVPair) string {
	switch {
	case bytes.Equal(kvA.Key[:1], supply.SupplyKey):
		var supplyA, supplyB supply.Supply
		amino.MustUnmarshalSized(kvA.Value, &supplyA)
		amino.MustUnmarshalSized(kvB.Value, &supplyB)
		return fmt.Sprintf("%v\n%v", supplyB, supplyB)
	default:
		panic(fmt.Sprintf("invalid supply key %X", kvA.Key))
	}
}
