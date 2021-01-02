package database

import (
	"context"

	"go.uber.org/zap"
	"gorm.io/gorm"

	svcErrors "github.com/BlackspaceInc/BlackspacePlatform/src/services/shopper_service/pkg/errors"
	"github.com/BlackspaceInc/BlackspacePlatform/src/services/shopper_service/pkg/graphql_api/model"
)

// GetShopperAccountByEmail gets a shopper account by email parameter from the backend database
// This query is performed against a certain field present in the respective backend database table
func (db *Db) GetShopperAccountByEmail(ctx context.Context, email string) *model.ShopperAccount {
	db.Logger.For(ctx).Info("get shopper account by email")
	ctx, span := db.startRootSpan(ctx, "get_shopper_account_by_email_op")
	defer span.Finish()

	tx := func(ctx context.Context, tx *gorm.DB) (interface{}, error) {
		conn := db.Conn.Engine

		db.Logger.For(ctx).Info("starting db transactions")
		childSpan := db.TracingEngine.CreateChildSpan(ctx, "get_shopper_account_by_email_tx")
		defer childSpan.Finish()

		var shopperAccountORM model.ShopperAccountORM

		// attempt to see if the record already exists
		recordNotFoundErr := conn.Where(&model.ShopperAccountORM{Email: email}).First(&shopperAccountORM).Error
		if recordNotFoundErr != nil {
			db.Logger.For(ctx).Error(svcErrors.ErrAccountDoesNotExist, "account does not exist")
			return nil, svcErrors.ErrAccountDoesNotExist
		}

		// transform orm type to account type
		account, err := shopperAccountORM.ToPB(ctx)
		if err != nil {
			db.Logger.For(ctx).Error(svcErrors.ErrFailedToConvertFromOrmType, err.Error())
			return nil, err
		}

		db.Logger.For(ctx).Info("successfully obtained business account", zap.Any("Email", email))
		return &account, nil
	}

	res, err := db.Conn.PerformComplexTransaction(ctx, tx)
	if err != nil {
		return nil
	}

	return res.(*model.ShopperAccount)
}
