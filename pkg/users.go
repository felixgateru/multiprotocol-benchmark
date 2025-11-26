package pkg

import (
	"context"
	"fmt"

	"github.com/absmach/supermq/pkg/sdk"
)

type users struct {
	sdk sdk.SDK
}

func NewUsersSDK(s sdk.SDK) *users {
	return &users{
		sdk: s,
	}
}

func (ut *users) ViewUser(ctx context.Context, id, token string) error {
	user, err := ut.sdk.User(ctx, id, token)
	if err != nil {
		return fmt.Errorf("failed to view user : %v", err)
	}
	if user.ID != id {
		return fmt.Errorf("viewed user id mismatch: got %s, want %s", user.ID, id)
	}

	return nil
}

func (ut *users) UserProfile(ctx context.Context, token string) error {
	user, err := ut.sdk.UserProfile(ctx, token)
	if err != nil {
		return fmt.Errorf("failed to get user profile : %v", err)
	}
	if user.ID == "" {
		return fmt.Errorf("user profile id is empty")
	}

	return nil
}

func (ut *users) CreateToken(ctx context.Context, username, password string) (sdk.Token, error) {
	login := sdk.Login{
		Username: username,
		Password: password,
	}
	token, err := ut.sdk.CreateToken(ctx, login)
	if err != nil {
		return sdk.Token{}, fmt.Errorf("failed to create token : %v", err)
	}
	if token.AccessToken == "" {
		return sdk.Token{}, fmt.Errorf("created token is empty")
	}

	return token, nil
}
