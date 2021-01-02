package graphql_api

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"errors"
	"fmt"

	"github.com/BlackspaceInc/BlackspacePlatform/src/services/shopper_service/pkg/graphql_api/generated"
	"github.com/BlackspaceInc/BlackspacePlatform/src/services/shopper_service/pkg/graphql_api/proto"
)

func (r *mutationResolver) CreateShopperAccount(ctx context.Context, input *proto.ShopperAccountMutation) (*proto.ShopperAccountResponse, error) {
	if input == nil {
		return nil, errors.New("invalid input arguments")
	}

	// convert input account to gorm model
	account, err := input.ShopperAccount.ConvertToModel()
	if err != nil {
		return nil, errors.New("failed to convert from input account type to gorm model")
	}

	// now we perform the proper creation steps

	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) DeleteShopperAccount(ctx context.Context, id *int) (*proto.OperationResponse, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) UpdateShopperAccount(ctx context.Context, id *int, input *proto.ShopperAccountMutation) (*proto.ShopperAccountResponse, error) {
	panic(fmt.Errorf("not implemented"))
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

type mutationResolver struct{ *Resolver }
