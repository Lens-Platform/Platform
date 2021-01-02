package graphql_api

import (
	"context"
	"fmt"

	core_logging "github.com/BlackspaceInc/BlackspacePlatform/src/libraries/core/core-logging/json"
	core_metrics "github.com/BlackspaceInc/BlackspacePlatform/src/libraries/core/core-metrics"
	core_tracing "github.com/BlackspaceInc/BlackspacePlatform/src/libraries/core/core-tracing"

	middleware "github.com/BlackspaceInc/BlackspacePlatform/src/libraries/core/core-middleware"

	"github.com/BlackspaceInc/BlackspacePlatform/src/services/business_account_service/pkg/database"
	svcErrors "github.com/BlackspaceInc/BlackspacePlatform/src/services/business_account_service/pkg/errors"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.
type SagaRefType func(ctx context.Context) error

type Resolver struct {
	Db      *database.Db
	Logger  core_logging.ILog
	Tracer  *core_tracing.TracingEngine
	Metrics *core_metrics.CoreMetricsEngine
}

func (r *mutationResolver) IsRequestAuthorized(ctx context.Context) error {
	if !middleware.IsAuthenticated(ctx) {
		r.Db.Logger.For(ctx).Info(fmt.Sprintf("unauthorized request"))
		return svcErrors.ErrUnauthorizedRequest
	}
	return nil
}
