package database

import (
	"context"
	"fmt"

	"github.com/BlackspaceInc/BlackspacePlatform/src/services/business_account_service/pkg/errors"
	"gorm.io/gorm"

	svcErrors "github.com/BlackspaceInc/BlackspacePlatform/src/services/shopper_service/pkg/errors"
	"github.com/BlackspaceInc/BlackspacePlatform/src/services/shopper_service/pkg/graphql_api/model"
)

// UpdateShopperAccount updates an existing shopper account record in the backend database
func (db *Db) UpdateShopperAccount(ctx context.Context, id uint32, account *model.ShopperAccount) (*model.ShopperAccount, error) {
	db.Logger.For(ctx).Info(fmt.Sprintf("updating business account - id: %d", id))
	ctx, span := db.startRootSpan(ctx, "update_account_db_op")
	defer span.Finish()

	tx := func(ctx context.Context, tx *gorm.DB) (interface{}, error) {
		db.Logger.For(ctx).Info("starting db transactions")
		childSpan := db.TracingEngine.CreateChildSpan(ctx, "update_business_accounts_db_tx")
		defer childSpan.Finish()

		if InputParamsAreInvalid(id, account) {
			db.Logger.ErrorM(svcErrors.ErrInvalidInputArguments, svcErrors.ErrInvalidInputArguments.Error())
			return nil, svcErrors.ErrInvalidInputArguments
		}

		if _, err := db.AssertAccountExistsAndPasswordUnchanged(ctx, account.Email, account); err != nil {
			return nil, err
		}

		return db.ProcessAccountAndSave(ctx, tx, account)
	}

	res, err := db.Conn.PerformComplexTransaction(ctx, tx)
	if err != nil {
		return nil, err
	}


	return res.(*model.ShopperAccount), nil
}

// ProcessAccountAndSave performs the proper account conversions and saves the record in the backend
// database
func (db *Db) ProcessAccountAndSave(ctx context.Context, tx *gorm.DB, account *model.ShopperAccount) (*model.ShopperAccount, error) {
	// convert account to orm type
	shopperAccountOrm, err := account.ToORM(ctx)
	if err != nil {
		db.Logger.Error(errors.ErrFailedToConvertToOrmType, err.Error())
		return nil, err
	}

	// save the account in the backend database
	if err := tx.Session(&gorm.Session{FullSaveAssociations: true}).Save(&shopperAccountOrm).Error; err != nil {
		db.Logger.Error(errors.ErrFailedToSaveUpdatedAccountRecord, err.Error())
		return nil, err
	}

	shopperAccount, err := shopperAccountOrm.ToPB(ctx)
	if err != nil {
		db.Logger.Error(errors.ErrFailedToConvertFromOrmType, err.Error())
		return nil, err
	}

	return &shopperAccount, nil
}

// AssertAccountExistsAndPasswordUnchanged ensures the account to update does not have an updated password field as well as that the account exists
// in the backend database
func (db *Db) AssertAccountExistsAndPasswordUnchanged(ctx context.Context, email string, account *model.ShopperAccount) (interface{}, error) {
	// attempt to see if account exist
	storedAccount := db.GetShopperAccountByEmail(ctx, email)
	if storedAccount == nil {
		db.Logger.ErrorM(errors.ErrAccountDoesNotExist, errors.ErrAccountDoesNotExist.Error())
		return nil, errors.ErrAccountDoesNotExist
	}

	// TODO compare the passwords and if they differ update the field through /password auth handler service call
	// As of now we do not allow users the ability to update their passwords through this call
	if !db.Conn.ComparePasswords(storedAccount.Password, []byte(account.Password)) {
		db.Logger.ErrorM(errors.ErrCannotUpdatePassword, errors.ErrCannotUpdatePassword.Error())
		return nil, errors.ErrCannotUpdatePassword
	}
	return nil, nil
}
