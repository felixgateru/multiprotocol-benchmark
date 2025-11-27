package pkg

import (
	"context"
	"fmt"

	"github.com/0x6flab/namegenerator"
	"github.com/absmach/supermq/pkg/sdk"
)

type Channels struct {
	sdk     sdk.SDK
	nameGen namegenerator.NameGenerator
}

func NewChannelsSDK(s sdk.SDK) *Channels {
	return &Channels{
		sdk:     s,
		nameGen: namegenerator.NewGenerator(),
	}
}

func (ct *Channels) CreateChannel(ctx context.Context, domainID, token string) (sdk.Channel, error) {
	ch := sdk.Channel{
		Name: ct.nameGen.Generate(),
	}
	createdCh, err := ct.sdk.CreateChannel(ctx, ch, domainID, token)
	if err != nil {
		return sdk.Channel{}, fmt.Errorf("failed to create channel : %v", err)
	}
	if createdCh.ID == "" {
		return sdk.Channel{}, fmt.Errorf("created channel id is empty")
	}

	return createdCh, nil
}

func (ct *Channels) CreateChannels(ctx context.Context, domainID, token string, n int) ([]sdk.Channel, error) {
	chs := []sdk.Channel{}
	for range n {
		ch := sdk.Channel{
			Name: ct.nameGen.Generate(),
		}
		chs = append(chs, ch)
	}

	createdChs, err := ct.sdk.CreateChannels(ctx, chs, domainID, token)
	if err != nil {
		return nil, fmt.Errorf("failed to create Channels : %v", err)
	}
	if len(createdChs) != n {
		return nil, fmt.Errorf("created Channels count mismatch: got %d, want %d", len(createdChs), n)
	}

	return createdChs, nil
}

func (ct *Channels) ListChannels(ctx context.Context, domainID, token string, noExpected int) error {
	chp, err := ct.sdk.Channels(ctx, sdk.PageMetadata{}, domainID, token)
	if err != nil {
		return fmt.Errorf("failed to list Channels : %v", err)
	}
	if len(chp.Channels) < noExpected {
		return fmt.Errorf("listed Channels count less than expected: got %d, want at least %d", len(chp.Channels), noExpected)
	}

	return nil
}

func (ct *Channels) ConnectClientsToChannels(ctx context.Context, channelIDs []string, clientIDs []string, domainID, token string) error {
	conn := sdk.Connection{
		ChannelIDs: channelIDs,
		ClientIDs:  clientIDs,
		Types:      []string{"Publish", "Subscribe"},
	}

	err := ct.sdk.Connect(ctx, conn, domainID, token)
	if err != nil {
		return fmt.Errorf("failed to connect clients to Channels : %v", err)
	}

	return nil
}

func (ct *Channels) DeleteChannel(ctx context.Context, channelID, domainID, token string) error {
	err := ct.sdk.DeleteChannel(ctx, channelID, domainID, token)
	if err != nil {
		return fmt.Errorf("failed to delete channel : %v", err)
	}
	return nil
}
