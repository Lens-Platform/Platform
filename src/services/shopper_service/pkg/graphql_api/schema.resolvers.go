package graphql_api

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"github.com/BlackspaceInc/BlackspacePlatform/src/services/shopper_service/pkg/graphql_api/generated"
	"github.com/BlackspaceInc/BlackspacePlatform/src/services/shopper_service/pkg/graphql_api/model"
	"github.com/BlackspaceInc/BlackspacePlatform/src/services/shopper_service/pkg/graphql_api/proto"
)

func (r *shopperAccountResolver) ID(ctx context.Context, obj *model.ShopperAccount) (*int, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *shopperAccountResolver) Addresses(ctx context.Context, obj *model.ShopperAccount) ([]*proto.Address, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *shopperAccountResolver) Causes(ctx context.Context, obj *model.ShopperAccount) ([]*proto.SubscribedTopic, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *shopperAccountResolver) CreditCard(ctx context.Context, obj *model.ShopperAccount) (*proto.CreditCard, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *shopperAccountResolver) SubscribedTopics(ctx context.Context, obj *model.ShopperAccount) ([]*proto.SubscribedTopic, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *shopperAccountResolver) AuthnID(ctx context.Context, obj *model.ShopperAccount) (*int, error) {
	panic(fmt.Errorf("not implemented"))
}

// ShopperAccount returns generated.ShopperAccountResolver implementation.
func (r *Resolver) ShopperAccount() generated.ShopperAccountResolver {
	return &shopperAccountResolver{r}
}

type shopperAccountResolver struct{ *Resolver }

// !!! WARNING !!!
// The code below was going to be deleted when updating resolvers. It has been copied here so you have
// one last chance to move it out of harms way if you want. There are two reasons this happens:
//  - When renaming or deleting a resolver the old code will be put in here. You can safely delete
//    it when you're done.
//  - You have helper methods in this file. Move them out to keep these resolver files clean.
func (r *shopperAccountResolver) DefaultAddress(ctx context.Context, obj *model.ShopperAccount) (*proto.Address, error) {
	panic(fmt.Errorf("not implemented"))
}
