package helpers

import (
	"context"
	"encoding/json"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
	"github.com/stretchr/testify/require"
)

const (
	ProposalVoteYes        = "yes"
	ProposalVoteNo         = "no"
	ProposalVoteNoWithVeto = "no_with_veto"
	ProposalVoteAbstain    = "abstain"
)

type Tally struct {
	AbstainCount    sdk.Int `json:"abstain"`
	NoCount         sdk.Int `json:"no"`
	NoWithVetoCount sdk.Int `json:"no_with_veto"`
	YesCount        sdk.Int `json:"yes"`
}

// QueryProposalTally gets tally results for a proposal
func QueryProposalTally(t *testing.T, ctx context.Context, chainNode *cosmos.ChainNode, proposalID string) Tally {
	stdout, _, err := chainNode.ExecQuery(ctx, "gov", "tally", proposalID)
	require.NoError(t, err)

	t.Log("q out:", string(stdout))

	var tally Tally
	err = json.Unmarshal([]byte(stdout), &tally)
	require.NoError(t, err)

	return tally
}
