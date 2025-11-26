package pkg

import (
	"context"
	"fmt"

	"github.com/0x6flab/namegenerator"
	"github.com/absmach/supermq"
	"github.com/absmach/supermq/pkg/sdk"
)

type clients struct {
	sdk     sdk.SDK
	nameGen namegenerator.NameGenerator
	idp     supermq.IDProvider
}

func NewClientsSDK(s sdk.SDK, idp supermq.IDProvider) *clients {
	return &clients{
		sdk:     s,
		nameGen: namegenerator.NewGenerator(),
		idp:     idp,
	}
}

func (ct *clients) CreateClient(ctx context.Context, domainID, token string) (sdk.Client, error) {
	secret, err := ct.idp.ID()
	if err != nil {
		return sdk.Client{}, fmt.Errorf("failed to generate client secret : %v", err)
	}
	cli := sdk.Client{
		Name: ct.nameGen.Generate(),
		Credentials: sdk.ClientCredentials{
			Secret: secret,
		},
	}
	client, err := ct.sdk.CreateClient(ctx, cli, domainID, token)
	if err != nil {
		return sdk.Client{}, fmt.Errorf("failed to create client : %v", err)
	}
	if client.ID == "" {
		return sdk.Client{}, fmt.Errorf("created client id is empty")
	}

	return client, nil
}

func (ct *clients) CreateClients(ctx context.Context, domainID, token string, n int) ([]sdk.Client, error) {
	clis := []sdk.Client{}
	for range n {
		secret, err := ct.idp.ID()
		if err != nil {
			return nil, fmt.Errorf("failed to generate client secret : %v", err)
		}
		cli := sdk.Client{
			Name: ct.nameGen.Generate(),
			Credentials: sdk.ClientCredentials{
				Secret: secret,
			},
		}
		clis = append(clis, cli)
	}

	createdClients, err := ct.sdk.CreateClients(ctx, clis, domainID, token)
	if err != nil {
		return nil, fmt.Errorf("failed to create clients : %v", err)
	}
	if len(createdClients) != n {
		return nil, fmt.Errorf("number of created clients mismatch: got %d, want %d", len(createdClients), n)
	}
	for _, client := range createdClients {
		if client.ID == "" {
			return nil, fmt.Errorf("one of the created client id is empty")
		}
	}

	return createdClients, nil
}

func (ct *clients) ListClients(ctx context.Context, domain, token string, expectedNo int) error {
	cp, err := ct.sdk.Clients(ctx, sdk.PageMetadata{}, domain, token)
	if err != nil {
		return fmt.Errorf("failed to list clients : %v", err)
	}
	if len(cp.Clients) < expectedNo {
		return fmt.Errorf("number of listed clients less than expected: got %d, want at least %d", len(cp.Clients), expectedNo)
	}

	return nil
}
