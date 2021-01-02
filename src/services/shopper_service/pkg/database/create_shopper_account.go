package database

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	svcErrors "github.com/BlackspaceInc/BlackspacePlatform/src/services/shopper_service/pkg/errors"
	"github.com/BlackspaceInc/BlackspacePlatform/src/services/shopper_service/pkg/graphql_api/model"
)

// CreateShopperAccount creates a shopper account record in the backend database
func (db *Db) CreateShopperAccount(ctx context.Context, account *model.ShopperAccount, authnId uint32) (*model.ShopperAccount, error){
	db.Logger.For(ctx).Info(fmt.Sprintf("creating business account - authnId: %d", authnId))
	ctx, span := db.startRootSpan(ctx, "create_account_db_op")
	defer span.Finish()

	tx := func(ctx context.Context, tx *gorm.DB) (interface{}, error){
		db.Logger.For(ctx).Info("starting create account operation - authnId: %d", authnId)
		childSpan := db.TracingEngine.CreateChildSpan(ctx, "create_account_db_tx")
		defer childSpan.Finish()

		if InputParamsAreInvalid(authnId, account) {
			db.Logger.ErrorM(svcErrors.ErrInvalidInputArguments, svcErrors.ErrInvalidInputArguments.Error())
			return nil, svcErrors.ErrInvalidInputArguments
		}

		// attempt to obtain the account from the backend database by email
		if existingAccount := db.GetShopperAccountByEmail(ctx, account.Email); existingAccount != nil {
			db.Logger.Error(svcErrors.ErrAccountAlreadyExist, svcErrors.ErrAccountAlreadyExist.Error())
			return nil, svcErrors.ErrAccountAlreadyExist
		}

		// if account does not exist we attempt to create it
		shopperAccount, err := account.ToORM(ctx)
		if err != nil {
			db.Logger.ErrorM(svcErrors.ErrFailedToConvertToOrmType, err.Error())
			return nil, err
		}

		// hash password
		if shopperAccount.Password, err = db.Conn.ValidateAndHashPassword(shopperAccount.Password); err != nil {
			db.Logger.For(ctx).Error(svcErrors.ErrFailedToHashPassword, err.Error())
			return nil, err
		}

		// activate account and authentication ID
		shopperAccount.IsActive = true
		shopperAccount.AuthnId = authnId

		// return account creation operation response
		return db.CreateAndSaveAccount(ctx, tx, shopperAccount, err)
	}

	result, err := db.Conn.PerformComplexTransaction(ctx, tx)
	if err != nil {
		return nil, err
	}

	createdAccount := result.(*model.ShopperAccount)
	return createdAccount, nil
}

// CreateAndSaveAccount creates and saves an account record in the backend database an performs some transformations from ORM type to regular
// account object
func (db *Db) CreateAndSaveAccount(ctx context.Context, tx *gorm.DB, shopperAccount model.ShopperAccountORM, err error) (*model.ShopperAccount, error) {
	// save the account record in the database
	if err := tx.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&shopperAccount).Error; err != nil {
		db.Logger.For(ctx).Error(svcErrors.ErrFailedToCreateAccount, err.Error())
		return nil, err
	}


	// convert from orm to shopper account
	createdAccount, err := shopperAccount.ToPB(ctx)
	if err != nil {
		db.Logger.For(ctx).Error(svcErrors.ErrFailedToConvertFromOrmType, err.Error())
		return nil, err
	}

	return &createdAccount, nil
}

// InputParamsAreInvalid asserts that the specific set of params are invalid
func InputParamsAreInvalid (id uint32, account *model.ShopperAccount) bool {
	return id == 0 || account == nil || account.Email == "" || account.Password == "" || account.Username == "" || account.LastName == ""
}
