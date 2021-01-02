package database

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/opentracing/opentracing-go"

	"github.com/BlackspaceInc/BlackspacePlatform/src/services/business_account_service/pkg/errors"
	"github.com/BlackspaceInc/BlackspacePlatform/src/services/business_account_service/pkg/utils"
)

// DistributedTxUnlockAccount unlocks an account in a distributed transaction
func (db *Db) DistributedTxUnlockAccount(ctx context.Context, id uint32, childSpan opentracing.Span, token string) error {
	f := func() error {
		subSpan, ctx := opentracing.StartSpanFromContext(ctx, "unlock_account_dtx_op", opentracing.ChildOf(childSpan.Context()))
		defer subSpan.Finish()

		// perform call to the authentication handler service
		httpClient := &http.Client{}
		url := db.AuthenticationHandlerServiceBaseEndpoint + "/unlock/" + fmt.Sprint(id)
		httpReq, _ := http.NewRequest("POST", url, nil)
		httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

		// Transmit the span's TraceContext as HTTP headers on our
		// outbound request.
		_ = opentracing.GlobalTracer().Inject(
			childSpan.Context(),
			opentracing.HTTPHeaders,
			opentracing.HTTPHeadersCarrier(httpReq.Header))

		resp, err := httpClient.Do(httpReq)
		if err != nil {
			db.Logger.For(ctx).Error(errors.ErrDistributedTransactionError, err.Error())
			return err
		}

		if resp.StatusCode != http.StatusOK {
			db.Logger.For(ctx).Error(errors.ErrDistributedTransactionError, errors.ErrDistributedTransactionError.Error()+" authentication handler service")
			return errors.ErrDistributedTransactionError
		}
		return nil
	}

	return db.PerformRetryableOperation(f)
}

// DistributedTxLockAccount locks an account in a distributed transaction
func (db *Db) DistributedTxLockAccount(ctx context.Context, id uint32, childSpan opentracing.Span, token string) error {
	f := func() error {
		subSpan, ctx := opentracing.StartSpanFromContext(ctx, "lock_account_dtx_op", opentracing.ChildOf(childSpan.Context()))
		defer subSpan.Finish()

		// perform call to the authentication handler service
		httpClient := &http.Client{}
		url := db.AuthenticationHandlerServiceBaseEndpoint + "/lock/" + fmt.Sprint(id)
		httpReq := db.createRequestAndPropagateTraces(url, subSpan, nil)
		httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

		resp, err := db.performHttpRequest(httpClient, httpReq, ctx)
		if err != nil || db.processResponseStatusCode(resp, ctx) {
			// todo: retry the operation with exponential backoff
			return err
		}

		return nil
	}

	return db.PerformRetryableOperation(f)
}

// DistributedTxUpdateAccountEmail updates the account record's email entry in a distributed transaction
func (db *Db) DistributedTxUpdateAccountEmail(ctx context.Context, id uint32, email string, childSpan opentracing.Span, token string) error {
	f := func() error {
		subSpan, ctx := opentracing.StartSpanFromContext(ctx, "update_account_dtx_op", opentracing.ChildOf(childSpan.Context()))
		defer subSpan.Finish()

		// TODO: extract token and place in request header
		// perform call to the authentication handler service
		httpClient := &http.Client{}
		url := db.AuthenticationHandlerServiceBaseEndpoint + "/update/" + fmt.Sprint(id)

		type updateEmailReq struct {
			Email string `json:"email"`
		}

		reqBody := updateEmailReq{
			Email: email,
		}

		body, err := utils.CreateRequestBody(reqBody)
		if err != nil {
			return err
		}

		httpReq := db.createRequestAndPropagateTraces(url, subSpan, body)
		httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

		resp, err := db.performHttpRequest(httpClient, httpReq, ctx)
		if err != nil || db.processResponseStatusCode(resp, ctx) {
			// todo: retry the operation with exponential backoff
			return err
		}

		return nil
	}

	return db.PerformRetryableOperation(f)
}

// DistributedTxCreateAccount creates the account record in a distributed transaction
func (db *Db) DistributedTxCreateAccount(ctx context.Context, email, password string, childSpan opentracing.Span) (*uint32, error) {
	return func() (*uint32, error) {
		subSpan, ctx := opentracing.StartSpanFromContext(ctx, "create_account_dtx_op", opentracing.ChildOf(childSpan.Context()))
		defer subSpan.Finish()

		// perform call to the authentication handler service
		httpClient := &http.Client{}
		url := db.AuthenticationHandlerServiceBaseEndpoint + "/create"

		// TODO: extract token and place in request header
		type createAccountReq struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		type createAccountRes struct {
			Error string `json:"error"`
			Id    uint32 `json:"id"`
		}

		var result createAccountRes
		reqBody := createAccountReq{
			Email:    email,
			Password: password,
		}

		body, err := utils.CreateRequestBody(reqBody)
		if err != nil {
			return nil, err
		}

		httpReq := db.createRequestAndPropagateTraces(url, subSpan, body)

		resp, err := db.performHttpRequest(httpClient, httpReq, ctx)
		if err != nil || db.processResponseStatusCode(resp, ctx) {
			// todo: retry the operation with exponential backoff
			return nil, err
		}

		if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, err
		}

		// update the id by reference
		return &result.Id, nil
	}()
}
