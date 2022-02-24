package starportcmd

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"
	campaigntypes "github.com/tendermint/spn/x/campaign/types"

	"github.com/tendermint/starport/starport/services/network"
)

const (
	flagCampaignName        = "name"
	flagCampaignMetadata    = "metadata"
	flagCampaignTotalShares = "total-shares"
	flagCampaignTotalSupply = "total-supply"
)

func NewNetworkCampaignUpdate() *cobra.Command {
	c := &cobra.Command{
		Use:   "update [campaign-id]",
		Short: "Update details fo the campaign of the campaign",
		Args:  cobra.ExactArgs(1),
		RunE:  networkCampaignUpdateHandler,
	}
	c.Flags().String(flagCampaignName, "", "Update the campaign name")
	c.Flags().String(flagCampaignMetadata, "", "Update the campaign metadata")
	c.Flags().String(flagCampaignTotalShares, "", "Update the shares supply for the campaign")
	c.Flags().String(flagCampaignTotalSupply, "", "Update the total of the mainnet of a campaign")
	c.Flags().AddFlagSet(flagNetworkFrom())
	c.Flags().AddFlagSet(flagSetKeyringBackend())
	return c
}

func networkCampaignUpdateHandler(cmd *cobra.Command, args []string) error {
	var (
		campaignName, _        = cmd.Flags().GetString(flagCampaignName)
		metadata, _            = cmd.Flags().GetString(flagCampaignMetadata)
		campaignTotalShares, _ = cmd.Flags().GetString(flagCampaignTotalShares)
		campaignTotalSupply, _ = cmd.Flags().GetString(flagCampaignTotalSupply)
	)
	totalShares, err := campaigntypes.NewShares(campaignTotalShares)
	if err != nil {
		return err
	}
	totalSupply, err := sdk.ParseCoinsNormalized(campaignTotalSupply)
	if err != nil {
		return err
	}

	nb, err := newNetworkBuilder(cmd)
	if err != nil {
		return err
	}
	defer nb.Cleanup()

	// parse campaign ID
	campaignID, err := network.ParseID(args[0])
	if err != nil {
		return err
	}

	if campaignName == "" && metadata == "" &&
		totalShares.Empty() && totalSupply.Empty() {
		return errors.New("at least one of the flags must be provided")
	}

	n, err := nb.Network()
	if err != nil {
		return err
	}

	return n.CampaignEdit(campaignID, campaignName, []byte(metadata), totalShares, totalSupply)
}