package v3_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	v3 "github.com/cosmos/cosmos-sdk/x/staking/migrations/v3"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

func TestMigrateJSON(t *testing.T) {
	encodingConfig := moduletestutil.MakeTestEncodingConfig()
	clientCtx := client.Context{}.
		WithInterfaceRegistry(encodingConfig.InterfaceRegistry).
		WithTxConfig(encodingConfig.TxConfig).
		WithCodec(encodingConfig.Codec)

	oldState := types.DefaultGenesisState()

	newState, err := v3.MigrateJSON(*oldState)
	require.NoError(t, err)

	bz, err := clientCtx.Codec.MarshalJSON(&newState)
	require.NoError(t, err)

	// Indent the JSON bz correctly.
	var jsonObj map[string]interface{}
	err = json.Unmarshal(bz, &jsonObj)
	require.NoError(t, err)
	indentedBz, err := json.MarshalIndent(jsonObj, "", "\t")
	require.NoError(t, err)

	// Make sure about new param MinCommissionRate.
	expected := `{
	"delegations": [],
	"exported": false,
	"last_tokenize_share_record_id": "0",
	"last_total_power": "0",
	"last_validator_powers": [],
	"params": {
		"bond_denom": "stake",
		"global_liquid_staking_cap": "1.000000000000000000",
		"historical_entries": 10000,
		"max_entries": 7,
		"max_validators": 100,
		"min_commission_rate": "0.000000000000000000",
		"unbonding_time": "1814400s",
		"validator_bond_factor": "-1.000000000000000000",
		"validator_liquid_staking_cap": "1.000000000000000000"
	},
	"redelegations": [],
	"tokenize_share_locks": [],
	"tokenize_share_records": [],
	"total_liquid_staked_tokens": "0",
	"unbonding_delegations": [],
	"validators": []
}`

	require.Equal(t, expected, string(indentedBz))
}
