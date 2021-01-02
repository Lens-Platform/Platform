package graphql_api

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"github.com/BlackspaceInc/BlackspacePlatform/src/libraries/core/core-middleware"
	svcErrors "github.com/BlackspaceInc/BlackspacePlatform/src/services/business_account_service/pkg/errors"
	"github.com/BlackspaceInc/BlackspacePlatform/src/services/business_account_service/pkg/graphql_api/generated"
	"github.com/BlackspaceInc/BlackspacePlatform/src/services/business_account_service/pkg/graphql_api/model"
	"github.com/BlackspaceInc/BlackspacePlatform/src/services/business_account_service/pkg/graphql_api/proto"
	"github.com/itimofeev/go-saga"
	opentracing "github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
)

func (r *mutationResolver) CreateBusinessAccount(ctx context.Context, input proto.CreateBusinessAccountRequest) (*model.BusinessAccount, error) {
	if err := r.IsRequestAuthorized(ctx); err != nil {
		r.Db.Logger.ErrorM(svcErrors.ErrUnauthorizedRequest, svcErrors.ErrUnauthorizedRequest.Error())
		return nil, svcErrors.ErrUnauthorizedRequest
	}

	jwtToken, err := middleware.GetTokenFromCtx(ctx)
	if err != nil {
		r.Db.Logger.ErrorM(svcErrors.ErrUnauthorizedRequest, svcErrors.ErrUnauthorizedRequest.Error())
		return nil, svcErrors.ErrUnauthorizedRequest
	}

	r.Db.Logger.For(ctx).InfoM(fmt.Sprintf("create business accounts api op"), zap.Any("business", input.BusinessAccount), zap.Any("authnId", input.AuthnID))
	sp, ctx := opentracing.StartSpanFromContext(ctx, "create_business_account_api_op")
	defer sp.Finish()

	if &input == nil || input.BusinessAccount == nil || input.BusinessAccount.Validate() != nil {
		r.Db.Logger.For(ctx).Error(svcErrors.ErrInvalidInputArguments, svcErrors.ErrInvalidInputArguments.Error())
		return nil, svcErrors.ErrInvalidInputArguments
	}

	// attempt to obtain business account form backend based on email
	account := r.Db.GetBusinessByEmail(ctx, input.BusinessAccount.Email)
	if account != nil {
		// attempt to see if account is inactive
		if !account.IsActive {
			sagaSteps := make([]*saga.Step, 0)
			// we reactivate the account via a saga
			distributedUnlockOpStep := &saga.Step{
				Name: "unlock_business_account_distributed_tx",
				// initial operation to unlock account
				Func: func(ctx context.Context) error {
					return r.Db.DistributedTxUnlockAccount(ctx, account.AuthnId, sp, jwtToken)
				},
				// compensating function to lock account
				CompensateFunc: func(ctx context.Context) error {
					return r.Db.DistributedTxLockAccount(ctx, account.AuthnId, sp, jwtToken)
				},
			}

			sagaSteps = append(sagaSteps, distributedUnlockOpStep)

			// activate account activity status
			reactivateAccountOpStep := &saga.Step{
				Name: "reactivate_business_account",
				// initial operation to activate account
				Func: func(ctx context.Context) error {
					return r.Db.SetBusinessAccountStatusAndSave(ctx, account, true)
				},
				// compensating operation to deactivate account
				CompensateFunc: func(ctx context.Context) error {
					return r.Db.SetBusinessAccountStatusAndSave(ctx, account, false)
				},
			}

			sagaSteps = append(sagaSteps, reactivateAccountOpStep)

			if err := r.Db.Saga.RunSaga(ctx, "unlock_business_account", sagaSteps...); err != nil {
				r.Db.Logger.For(ctx).Error(err, err.Error())
				return nil, err
			}

			savedAccount := r.Db.GetBusinessById(ctx, account.Id)
			return savedAccount, nil
		} else {
			r.Db.Logger.For(ctx).Error(svcErrors.ErrAccountAlreadyExist, svcErrors.ErrAccountAlreadyExist.Error())
			return nil, svcErrors.ErrAccountAlreadyExist
		}
	}

	// now we attempt to create the account from the current context of this service since for this api to be invoked the api
	// gateway must have already created an entry in the authentication handler service referencing this account record
	// perform this operation as a retryable one
	account, err = r.Db.CreateBusinessAccount(ctx, input.BusinessAccount, uint32(*input.AuthnID))
	if err != nil {
		r.Db.Logger.For(ctx).Error(err, err.Error())
		return nil, err
	}

	return account, nil
}

func (r *mutationResolver) UpdateBusinessAccount(ctx context.Context, input proto.UpdateBusinessAccountRequest) (*model.BusinessAccount, error) {
	if err := r.IsRequestAuthorized(ctx); err != nil {
		return nil, err
	}

	jwtToken, err := middleware.GetTokenFromCtx(ctx)
	if err != nil {
		r.Db.Logger.ErrorM(svcErrors.ErrUnauthorizedRequest, svcErrors.ErrUnauthorizedRequest.Error())
		return nil, svcErrors.ErrUnauthorizedRequest
	}

	r.Db.Logger.For(ctx).Info(fmt.Sprintf("update business account api op"))
	sp, ctx := opentracing.StartSpanFromContext(ctx, "update_business_account_api_op")
	defer sp.Finish()

	// validate the input
	if &input == nil ||
		input.BusinessAccount == nil ||
		input.BusinessAccount.Validate() != nil ||
		input.ID == nil {
		r.Db.Logger.For(ctx).Error(svcErrors.ErrInvalidInputArguments, svcErrors.ErrInvalidInputArguments.Error())
		return nil, svcErrors.ErrInvalidInputArguments
	}

	var newBusinessAccount = input.BusinessAccount
	var businessAccountId = uint32(*input.ID)

	// attempt obtain the business account stored in the backend db first
	oldBusinessAccount := r.Db.GetBusinessById(ctx, businessAccountId)
	if oldBusinessAccount == nil {
		r.Db.Logger.For(ctx).ErrorM(svcErrors.ErrAccountDoesNotExist, fmt.Sprintf("business account with id %d does not exist", businessAccountId))
		return nil, svcErrors.ErrAccountDoesNotExist
	}

	var transactionalSteps = make([]*saga.Step, 0)
	var updatedAccount = make(chan *model.BusinessAccount, 1)
	// TODO: handle password updates via authentication handler service in the future - Need to implement this too
	// TODO: send out an email to the account owner that the email or password has been changed
	if !r.Db.Conn.ComparePasswords(oldBusinessAccount.Password, []byte(newBusinessAccount.Password)) {
		r.Db.Logger.For(ctx).ErrorM(svcErrors.ErrCannotUpdatePassword, svcErrors.ErrCannotUpdatePassword.Error())
		return nil, svcErrors.ErrCannotUpdatePassword
	}

	// define a saga step tailored to saving the new business account record in our backend db
	updateAndSaveAccountStep := saga.Step{
		Name: "update_business_account",
		// initial operation to update business account
		Func: func(ctx context.Context, output chan<- *model.BusinessAccount) SagaRefType {
			return func(ctx context.Context) error {
				acc, err := r.Db.UpdateBusinessAccount(ctx, businessAccountId, newBusinessAccount)
				output <- acc
				return err
			}
		}(ctx, updatedAccount),
		// no compensating operation for the update. Just return an error if the initial call failed
		CompensateFunc: func(ctx context.Context) error { // no compensating function just return an error if this operation fails
			return svcErrors.ErrFailedToSaveUpdatedAccountRecord
		},
	}
	transactionalSteps = append(transactionalSteps, &updateAndSaveAccountStep)

	// check if the email is updated
	if oldBusinessAccount.Email != newBusinessAccount.Email {
		// perform distributed tx via saga
		var authnId = oldBusinessAccount.AuthnId
		var newEmail = newBusinessAccount.Email

		// update the email account from the context of the authentication handler service
		updateEmailInDtxStep := saga.Step{
			Name: "update_business_account_email_distributed_tx",
			// update account initial operation via distributed transaction
			Func: func(ctx context.Context) error {
				return r.Db.DistributedTxUpdateAccountEmail(ctx, authnId, newEmail, sp, jwtToken)
			},
			// compensating function to reset the account to its inital state if the distributed update call failed
			CompensateFunc: func(ctx context.Context, output chan<- *model.BusinessAccount) SagaRefType {
				return func(ctx context.Context) error {
					acc, err := r.Db.UpdateBusinessAccount(ctx, businessAccountId, oldBusinessAccount) // reset business account to original state
					if err != nil {
						return err
					}

					output <- acc
					return nil
				}
			}(ctx, updatedAccount),
		}

		transactionalSteps = append(transactionalSteps, &updateEmailInDtxStep)
	}

	// run the saga
	if err := r.Db.Saga.RunSaga(ctx, "update_business_account", transactionalSteps...); err != nil {
		r.Db.Logger.For(ctx).Error(err, err.Error())
		return nil, err
	}

	return <-updatedAccount, nil
}

func (r *mutationResolver) DeleteBusinessAccount(ctx context.Context, id proto.DeleteBusinessAccountRequest) (*proto.DeleteBusinessAccountResponse, error) {
	trueResp := true
	if err := r.IsRequestAuthorized(ctx); err != nil {
		return nil, err
	}

	jwtToken, err := middleware.GetTokenFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	r.Db.Logger.For(ctx).Info(fmt.Sprintf("delete business account api op"))
	sp, ctx := opentracing.StartSpanFromContext(ctx, "delete_business_account_api_op")
	defer sp.Finish()

	// validate the input
	if id.ID == nil {
		r.Db.Logger.For(ctx).Error(svcErrors.ErrInvalidInputArguments, svcErrors.ErrInvalidInputArguments.Error())
		return nil, svcErrors.ErrInvalidInputArguments
	}

	accountId := uint32(*id.ID)
	account := r.Db.GetBusinessById(ctx, uint32(*id.ID))
	if account == nil {
		r.Db.Logger.For(ctx).Error(svcErrors.ErrAccountDoesNotExist, svcErrors.ErrAccountDoesNotExist.Error())
		return nil, svcErrors.ErrAccountDoesNotExist
	}

	var (
		transactionalSteps                = make([]*saga.Step, 0)
		dtxLockOpStep, archiveAccountStep saga.Step
	)

	// we perform this operation as a distributed transaction
	// since we never truly delete the account from our backend we set the record to inactive
	// while also ensuring from the context of the authentication handler service the account is locked
	// define saga

	// first operation is to perform a distributed transaction and lock the account if possible
	dtxLockOpStep = saga.Step{
		Name: "distributed lock operation",
		// Perform lock operation as a distributed transaction
		Func: func(ctx context.Context) error {
			return r.Db.DistributedTxLockAccount(ctx, account.AuthnId, sp, jwtToken)
		},
		// Perform compensating distributed unlock operation if the initial event failed
		CompensateFunc: func(ctx context.Context) error {
			return r.Db.DistributedTxUnlockAccount(ctx, account.AuthnId, sp, jwtToken)
		},
	}
	transactionalSteps = append(transactionalSteps, &dtxLockOpStep)

	// second operation is to update the state of the account and save to database
	archiveAccountStep = saga.Step{
		Name: "archive business account operation",
		// Archive the business account (perform account deactivation)
		Func: func(ctx context.Context) error {
			return r.Db.ArchiveBusinessAccount(ctx, accountId)
		},
		// Perform business account activation
		CompensateFunc: func(ctx context.Context) error {
			return r.Db.SetBusinessAccountStatusAndSave(ctx, account, true)
		}, // activate account
		Options: nil,
	}
	transactionalSteps = append(transactionalSteps, &archiveAccountStep)

	// run the saga
	if err := r.Db.Saga.RunSaga(ctx, "archive_business_account", transactionalSteps...); err != nil {
		r.Db.Logger.For(ctx).Error(err, err.Error())
		return nil, err
	}

	return &proto.DeleteBusinessAccountResponse{Result: &trueResp}, nil
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

type mutationResolver struct{ *Resolver }
