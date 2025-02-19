package v4

import (
	"sort"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/exported"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// MigrateStore performs in-place store migrations from v3 to v4.
func MigrateStore(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec, legacySubspace exported.Subspace) error {
	store := ctx.KVStore(storeKey)

	// migrate params
	if err := migrateParams(ctx, store, cdc, legacySubspace); err != nil {
		return err
	}

	// migrate unbonding delegations
	if err := migrateUBDEntries(ctx, store, cdc, legacySubspace); err != nil {
		return err
	}

	if err := migrateAllValidatorBondAndLiquidSharesToZero(ctx, store, cdc); err != nil {
		return err
	}

	if err := migrateAllDelegationValidatorBondsFalse(ctx, store, cdc); err != nil {
		return err
	}

	if err := migrateTotalLiquidStakedTokens(ctx, store, cdc, sdk.ZeroInt()); err != nil {
		return err
	}

	return nil
}

// migrateParams will set the params to store from legacySubspace
func migrateParams(ctx sdk.Context, store storetypes.KVStore, cdc codec.BinaryCodec, legacySubspace exported.Subspace) error {
	var legacyParams types.Params
	legacySubspace.GetParamSet(ctx, &legacyParams)

	// add LSM-related params before validating
	legacyParams.ValidatorBondFactor = types.DefaultValidatorBondFactor
	legacyParams.GlobalLiquidStakingCap = types.DefaultGlobalLiquidStakingCap
	legacyParams.ValidatorLiquidStakingCap = types.DefaultValidatorLiquidStakingCap

	if err := legacyParams.Validate(); err != nil {
		return err
	}

	bz := cdc.MustMarshal(&legacyParams)
	store.Set(types.ParamsKey, bz)
	return nil
}

// migrateUBDEntries will remove the ubdEntries with same creation_height
// and create a new ubdEntry with updated balance and initial_balance
func migrateUBDEntries(ctx sdk.Context, store storetypes.KVStore, cdc codec.BinaryCodec, legacySubspace exported.Subspace) error {
	iterator := sdk.KVStorePrefixIterator(store, types.UnbondingDelegationKey)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		ubd := types.MustUnmarshalUBD(cdc, iterator.Value())

		entriesAtSameCreationHeight := make(map[int64][]types.UnbondingDelegationEntry)
		for _, ubdEntry := range ubd.Entries {
			entriesAtSameCreationHeight[ubdEntry.CreationHeight] = append(entriesAtSameCreationHeight[ubdEntry.CreationHeight], ubdEntry)
		}

		creationHeights := make([]int64, 0, len(entriesAtSameCreationHeight))
		for k := range entriesAtSameCreationHeight {
			creationHeights = append(creationHeights, k)
		}

		sort.Slice(creationHeights, func(i, j int) bool { return creationHeights[i] < creationHeights[j] })

		ubd.Entries = make([]types.UnbondingDelegationEntry, 0, len(creationHeights))

		for _, h := range creationHeights {
			ubdEntry := types.UnbondingDelegationEntry{
				Balance:        sdk.ZeroInt(),
				InitialBalance: sdk.ZeroInt(),
			}
			for _, entry := range entriesAtSameCreationHeight[h] {
				ubdEntry.Balance = ubdEntry.Balance.Add(entry.Balance)
				ubdEntry.InitialBalance = ubdEntry.InitialBalance.Add(entry.InitialBalance)
				ubdEntry.CreationHeight = entry.CreationHeight
				ubdEntry.CompletionTime = entry.CompletionTime
			}
			ubd.Entries = append(ubd.Entries, ubdEntry)
		}

		// set the new ubd to the store
		setUBDToStore(ctx, store, cdc, ubd)
	}
	return nil
}

func setUBDToStore(ctx sdk.Context, store storetypes.KVStore, cdc codec.BinaryCodec, ubd types.UnbondingDelegation) {
	delegatorAddress := sdk.MustAccAddressFromBech32(ubd.DelegatorAddress)

	bz := types.MustMarshalUBD(cdc, ubd)

	addr, err := sdk.ValAddressFromBech32(ubd.ValidatorAddress)
	if err != nil {
		panic(err)
	}

	key := types.GetUBDKey(delegatorAddress, addr)

	store.Set(key, bz)
}

// migrateAllValidatorBondAndLiquidSharesToZero sets each validator's ValidatorBondShares and LiquidShares to zero
func migrateAllValidatorBondAndLiquidSharesToZero(ctx sdk.Context, store storetypes.KVStore, cdc codec.BinaryCodec) error {
	iterator := sdk.KVStorePrefixIterator(store, types.ValidatorsKey)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		validator := types.MustUnmarshalValidator(cdc, iterator.Value())

		validator.ValidatorBondShares = sdk.ZeroDec()
		validator.LiquidShares = sdk.ZeroDec()

		bz := types.MustMarshalValidator(cdc, &validator)
		store.Set(types.GetValidatorKey(validator.GetOperator()), bz)
	}

	return nil
}

// migrateAllDelegationValidatorBondsFalse sets each delegation's ValidatorBond to false
func migrateAllDelegationValidatorBondsFalse(ctx sdk.Context, store storetypes.KVStore, cdc codec.BinaryCodec) error {
	iterator := sdk.KVStorePrefixIterator(store, types.DelegationKey)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		delegation := types.MustUnmarshalDelegation(cdc, iterator.Value())

		delegation.ValidatorBond = false

		bz := types.MustMarshalDelegation(cdc, delegation)
		delegatorAddress := sdk.MustAccAddressFromBech32(delegation.DelegatorAddress)
		store.Set(types.GetDelegationKey(delegatorAddress, delegation.GetValidatorAddr()), bz)
	}

	return nil
}

// migrateTotalLiquidStakedTokens migrates the total outstanding tokens owned by a liquid staking provider
func migrateTotalLiquidStakedTokens(ctx sdk.Context, store storetypes.KVStore, cdc codec.BinaryCodec, tokens sdk.Int) error {
	tokensBz, err := tokens.Marshal()
	if err != nil {
		panic(err)
	}

	store.Set(types.TotalLiquidStakedTokensKey, tokensBz)

	return nil
}
