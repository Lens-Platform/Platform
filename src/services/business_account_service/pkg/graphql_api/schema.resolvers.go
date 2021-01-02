package graphql_api

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"github.com/BlackspaceInc/BlackspacePlatform/src/services/business_account_service/pkg/graphql_api/generated"
	"github.com/BlackspaceInc/BlackspacePlatform/src/services/business_account_service/pkg/graphql_api/model"
	"github.com/BlackspaceInc/BlackspacePlatform/src/services/business_account_service/pkg/graphql_api/proto"
)

func (r *businessAccountResolver) ID(ctx context.Context, obj *model.BusinessAccount) (*int, error) {
	ID := obj.GetId()
	return HandleErrorsIfPresent(ID)
}

func (r *businessAccountResolver) PhoneNumber(ctx context.Context, obj *model.BusinessAccount) (*proto.PhoneNumber, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *businessAccountResolver) Media(ctx context.Context, obj *model.BusinessAccount) (*proto.Media, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *businessAccountResolver) TypeOfBusiness(ctx context.Context, obj *model.BusinessAccount) (*proto.BusinessType, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *businessAccountResolver) MerchantType(ctx context.Context, obj *model.BusinessAccount) (*proto.MerchantType, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *businessAccountResolver) PaymentDetails(ctx context.Context, obj *model.BusinessAccount) (*proto.PaymentProcessingMethods, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *businessAccountResolver) ServicesManagedByBlackspace(ctx context.Context, obj *model.BusinessAccount) (*proto.ServicesManagedByBlackspace, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *businessAccountResolver) FounderAddress(ctx context.Context, obj *model.BusinessAccount) (*proto.Address, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *businessAccountResolver) SubscribedTopics(ctx context.Context, obj *model.BusinessAccount) (*proto.Topics, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *businessAccountResolver) AuthnID(ctx context.Context, obj *model.BusinessAccount) (*int, error) {
	ID := obj.GetAuthnId()
	return HandleErrorsIfPresent(ID)
}

// BusinessAccount returns generated.BusinessAccountResolver implementation.
func (r *Resolver) BusinessAccount() generated.BusinessAccountResolver {
	return &businessAccountResolver{r}
}

type businessAccountResolver struct{ *Resolver }
