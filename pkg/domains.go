package pkg

import (
	"context"
	"fmt"

	"github.com/0x6flab/namegenerator"
	"github.com/absmach/supermq"
	"github.com/absmach/supermq/pkg/sdk"
)

const routePrefix string = "a"

type Domains struct {
	sdk     sdk.SDK
	nameGen namegenerator.NameGenerator
	idp     supermq.IDProvider
}

func NewDomainsSDK(s sdk.SDK, idp supermq.IDProvider) *Domains {
	return &Domains{
		sdk:     s,
		nameGen: namegenerator.NewGenerator(),
		idp:     idp,
	}
}

func (dt *Domains) CreateDomain(ctx context.Context, token string) (sdk.Domain, error) {
	route, err := dt.idp.ID()
	if err != nil {
		return sdk.Domain{}, fmt.Errorf("failed to generate domain route : %v", err)
	}
	dom := sdk.Domain{
		Name:  dt.nameGen.Generate(),
		Route: routePrefix + route,
	}
	createdDom, err := dt.sdk.CreateDomain(ctx, dom, token)
	if err != nil {
		return sdk.Domain{}, fmt.Errorf("failed to create domain : %v", err)
	}
	if createdDom.ID == "" {
		return sdk.Domain{}, fmt.Errorf("created domain id is empty")
	}

	return createdDom, nil
}

func (dt *Domains) ListDomains(ctx context.Context, token string, noExpected int) error {
	domp, err := dt.sdk.Domains(ctx, sdk.PageMetadata{}, token)
	if err != nil {
		return fmt.Errorf("failed to list Domains : %v", err)
	}
	if len(domp.Domains) < noExpected {
		return fmt.Errorf("listed Domains count less than expected: got %d, want at least %d", len(domp.Domains), noExpected)
	}

	return nil
}
