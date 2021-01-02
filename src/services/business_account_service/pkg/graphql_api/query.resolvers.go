package graphql_api

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	svcErrors "github.com/BlackspaceInc/BlackspacePlatform/src/services/business_account_service/pkg/errors"
	"github.com/BlackspaceInc/BlackspacePlatform/src/services/business_account_service/pkg/graphql_api/generated"
	"github.com/BlackspaceInc/BlackspacePlatform/src/services/business_account_service/pkg/graphql_api/model"
	"github.com/BlackspaceInc/BlackspacePlatform/src/services/business_account_service/pkg/graphql_api/proto"
	opentracing "github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
)

func (r *queryResolver) GetBusinessAccount(ctx context.Context, input proto.GetBusinessAccountRequest) (*model.BusinessAccount, error) {
	r.Db.Logger.For(ctx).Info(fmt.Sprintf("get business account api op"))
	sp, ctx := opentracing.StartSpanFromContext(ctx, "get_business_account_api_op")
	defer sp.Finish()

	// ensure input is not nil or misconfigured
	if &input == nil || input.ID == nil {
		r.Db.Logger.For(ctx).Error(svcErrors.ErrInvalidInputArguments, svcErrors.ErrInvalidInputArguments.Error())
		return nil, svcErrors.ErrInvalidInputArguments
	}

	account, err := r.Db.GetBusinessAccount(ctx, uint32(*input.ID))
	if err != nil {
		r.Db.Logger.For(ctx).ErrorM(err, err.Error())
		return nil, err
	}

	r.Db.Logger.For(ctx).Info(fmt.Sprintf("successfully obtain business account - id: %s", fmt.Sprint(*input.ID)), zap.String("company name",
		account.CompanyName))

	return account, nil
}

func (r *queryResolver) GetBusinessAccounts(ctx context.Context, limit proto.GetBusinessAccountsRequest) ([]*model.BusinessAccount, error) {
	r.Db.Logger.For(ctx).Info(fmt.Sprintf("get paginated business accounts api op"))
	sp, ctx := opentracing.StartSpanFromContext(ctx, "get_paginated_business_account_api_op")
	defer sp.Finish()

	if &limit == nil || limit.Limit == nil {
		r.Db.Logger.For(ctx).Error(svcErrors.ErrInvalidInputArguments, svcErrors.ErrInvalidInputArguments.Error())
		return nil, svcErrors.ErrInvalidInputArguments
	}

	accounts, err := r.Db.GetPaginatedBusinessAccounts(ctx, int64(*limit.Limit))
	if err != nil {
		r.Db.Logger.For(ctx).Error(err, err.Error())
		return nil, err
	}

	r.Db.Logger.For(ctx).Info(fmt.Sprintf("successfully obtain %d business accounts", limit.Limit))
	return accounts, nil
}

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type queryResolver struct{ *Resolver }
