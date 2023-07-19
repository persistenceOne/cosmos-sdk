package interchaintest

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	interchaintest "github.com/strangelove-ventures/interchaintest/v7"
	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v7/ibc"
	"github.com/strangelove-ventures/interchaintest/v7/testreporter"
)

var (
	simdEncoding = cosmos.DefaultEncoding()

	// simdLegacyGenesisKV contains gov params for SDK v45
	simdLegacyGenesisKV = []cosmos.GenesisKV{
		{
			Key:   "app_state.gov.voting_params.voting_period",
			Value: "300s",
		},
		{
			Key:   "app_state.gov.deposit_params.max_deposit_period",
			Value: "5s",
		},
		{
			Key:   "app_state.gov.deposit_params.min_deposit.0.denom",
			Value: "stake",
		},
		{
			Key:   "app_state.gov.deposit_params.min_deposit.0.amount",
			Value: "10",
		},
	}

	simdConfig = ibc.ChainConfig{
		Type:    "cosmos",
		Name:    "simd",
		ChainID: "ictest-simd-1",
		Images: []ibc.DockerImage{{
			Repository: "simd",
			Version:    "local",
		}},

		Bin:                    "simd",
		Bech32Prefix:           "cosmos",
		Denom:                  "stake",
		CoinType:               "118",
		GasPrices:              "0stake",
		GasAdjustment:          1.5,
		TrustingPeriod:         "112h",
		NoHostMount:            false,
		ConfigFileOverrides:    nil,
		EncodingConfig:         &simdEncoding,
		UsingNewGenesisCommand: false, // important for v0.45.16
		ModifyGenesis:          cosmos.ModifyGenesis(simdLegacyGenesisKV),
	}

	genesisWalletAmount = int64(10_000_000)
)

func CreateChain(t *testing.T, ctx context.Context, numVals, numFull int) (*interchaintest.Interchain, *cosmos.CosmosChain) {
	cf := interchaintest.NewBuiltinChainFactory(zaptest.NewLogger(t), []*interchaintest.ChainSpec{
		{
			Name:          "simd",
			ChainName:     "simd",
			Version:       "local",
			ChainConfig:   simdConfig,
			NumValidators: &numVals,
			NumFullNodes:  &numFull,
		},
	})

	chains, err := cf.Chains(t.Name())
	require.NoError(t, err)

	ic := interchaintest.NewInterchain().AddChain(chains[0])
	client, network := interchaintest.DockerSetup(t)

	err = ic.Build(
		ctx,
		testreporter.NewNopReporter().RelayerExecReporter(t),
		interchaintest.InterchainBuildOptions{
			TestName:         t.Name(),
			Client:           client,
			NetworkID:        network,
			SkipPathCreation: true,
		},
	)
	require.NoError(t, err)

	return ic, chains[0].(*cosmos.CosmosChain)
}

func firstUserName(prefix string) string {
	return prefix + "-user1"
}

func secondUserName(prefix string) string {
	return prefix + "-user2"
}
