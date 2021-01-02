package graphql_api

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"github.com/BlackspaceInc/BlackspacePlatform/src/services/shopper_service/pkg/graphql_api/generated"
	"github.com/BlackspaceInc/BlackspacePlatform/src/services/shopper_service/pkg/graphql_api/proto"
)

func (r *queryResolver) GetShopperAccount(ctx context.Context, id *int) (*proto.ShopperAccountResponse, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) GetShopperAccounts(ctx context.Context, limit *int) (*proto.ShopperAccountsResponse, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) GetShopperAccountsSubscribedToTopic(ctx context.Context, topic *proto.SubscribedTopicInput) (*proto.ShopperAccountsResponse, error) {
	panic(fmt.Errorf("not implemented"))
}

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type queryResolver struct{ *Resolver }
