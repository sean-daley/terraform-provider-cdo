package examples

import (
	"context"
	"github.com/CiscoDevnet/terraform-provider-cdo/go-client/internal/http"
)

type CreateInput struct {
	Uid string
}

func NewCreateInput(uid string) CreateInput {
	return CreateInput{
		Uid: uid,
	}
}

type CreateOutput struct {
}

func Create(ctx context.Context, client http.Client, createInp CreateInput) (*CreateOutput, error) {

	// TODO

	return nil, nil
}
