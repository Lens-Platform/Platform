package database

import (
	"context"
	"fmt"

	saga "github.com/itimofeev/go-saga"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/BlackspaceInc/BlackspacePlatform/src/services/business_account_service/pkg/errors"
	proto "github.com/BlackspaceInc/BlackspacePlatform/src/services/business_account_service/pkg/graphql_api/model"
)

type dbOperationType string

type IDbOperations interface {
	// CreateBusinessAccount creates a business account
	CreateBusinessAccount(ctx context.Context, account *proto.BusinessAccount) (*proto.BusinessAccount, error)
	// UpdateBusinessAccount updates a business account by Id
	UpdateBusinessAccount(ctx context.Context, id uint32, account *proto.BusinessAccount) (*proto.BusinessAccount, error)
	// DeleteBusinessAccount deletes a business account by Id
	ArchiveBusinessAccount(ctx context.Context, id uint32) (bool, error)
	// DeleteBusinessAccounts deletes a set of business accounts by Id
	ArchiveBusinessAccounts(ctx context.Context, ids []uint32) ([]bool, error)
	// GetBusinessAccount gets a business account by id
	GetBusinessAccount(ctx context.Context, id uint32) (*proto.BusinessAccount, error)
	// GetBusinessAccounts gets a set of business accounts by id
	GetBusinessAccounts(ctx context.Context, ids []uint32) ([]*proto.BusinessAccount, error)
}

// UpdateBusinessAccount updates a business account
func (db *Db) UpdateBusinessAccount(ctx context.Context, id uint32, account *proto.BusinessAccount) (*proto.BusinessAccount, error) {
	db.Logger.For(ctx).Info(fmt.Sprintf("updated business account - id : %d", id))
	ctx, span := db.startRootSpan(ctx, "get_business_accounts_db_op")
	defer span.Finish()

	tx := func(ctx context.Context, tx *gorm.DB) (interface{}, error) {
		db.Logger.For(ctx).Info("starting update account operation for account with id :%id", id)
		childSpan := db.TracingEngine.CreateChildSpan(ctx, "update_business_accounts_db_tx")
		defer childSpan.Finish()

		if id == 0 {
			db.Logger.ErrorM(errors.ErrInvalidInputArguments, errors.ErrInvalidInputArguments.Error())
			return nil, errors.ErrInvalidInputArguments
		}

		// attempt to see if account exists
		result := db.GetBusinessById(ctx, id)
		if result == nil {
			db.Logger.ErrorM(errors.ErrAccountDoesNotExist, errors.ErrAccountDoesNotExist.Error())
			return nil, errors.ErrAccountDoesNotExist
		}

		// TODO compare the passwords and if they differ update the field through /password auth handler service call
		// As of now we do not allow users the ability to update their passwords through this call
		if !db.Conn.ComparePasswords(result.Password, []byte(account.Password)) {
			db.Logger.ErrorM(errors.ErrCannotUpdatePassword, errors.ErrCannotUpdatePassword.Error())
			return nil, errors.ErrCannotUpdatePassword
		}

		// convert account record to ORM type
		businessAccountOrm, err := account.ToORM(ctx)
		if err != nil {
			db.Logger.Error(errors.ErrFailedToConvertToOrmType, err.Error())
			return nil, err
		}

		// save the account
		if err := tx.Session(&gorm.Session{FullSaveAssociations: true}).Save(&businessAccountOrm).Error; err != nil {
			db.Logger.Error(errors.ErrFailedToSaveUpdatedAccountRecord, err.Error())
			return nil, err
		}

		updatedAccount, err := businessAccountOrm.ToPB(ctx)
		if err != nil {
			db.Logger.Error(errors.ErrFailedToConvertFromOrmType, err.Error())
			return nil, err
		}

		return &updatedAccount, nil
	}

	res, err := db.Conn.PerformComplexTransaction(ctx, tx)
	if err != nil {
		return nil, err
	}

	return res.(*proto.BusinessAccount), nil
}

// DeleteBusinessAccounts deletes a set of business accounts by ids
func (db *Db) ArchiveBusinessAccounts(ctx context.Context, ids []uint32) ([]bool, error) {
	db.Logger.For(ctx).Info("delete business accounts")
	ctx, span := db.startRootSpan(ctx, "archive_business_accounts_db_op")
	defer span.Finish()

	tx := func(ctx context.Context, tx *gorm.DB) (interface{}, error) {
		// start child span
		db.Logger.For(ctx).Info("starting db transactions")
		childSpan := db.TracingEngine.CreateChildSpan(ctx, "archive_business_accounts_db_tx")
		defer childSpan.Finish()

		var deleteStatus = make([]bool, len(ids))
		for _, id := range ids {
			var status bool = true
			err := db.ArchiveBusinessAccount(ctx, id)
			if err != nil {
				db.Logger.For(ctx).Error(err, fmt.Sprintf("%s - for id %d ", errors.ErrFailedToDeleteBusinessAccount.Error(), id))
				status = false
			}

			deleteStatus = append(deleteStatus, status)
		}

		return deleteStatus, nil
	}

	res, err := db.Conn.PerformComplexTransaction(ctx, tx)
	if err != nil {
		return nil, err
	}

	return res.([]bool), nil
}

// DeleteBusinessAccount archives a business account by id
func (db *Db) ArchiveBusinessAccount(ctx context.Context, id uint32) error {
	db.Logger.For(ctx).Info("delete business account")
	ctx, span := db.startRootSpan(ctx, "archive_business_account_db_op")
	defer span.Finish()

	tx := func(ctx context.Context, tx *gorm.DB) error {
		db.Logger.For(ctx).Info("starting db transactions")
		childSpan := db.TracingEngine.CreateChildSpan(ctx, "archive_business_account_db_tx")
		defer childSpan.Finish()

		if id == 0 {
			db.Logger.ErrorM(errors.ErrInvalidInputArguments, errors.ErrInvalidInputArguments.Error())
			return errors.ErrInvalidInputArguments
		}

		// check if business actually exists
		account := db.GetBusinessById(ctx, id)
		if account == nil {
			db.Logger.For(ctx).ErrorM(errors.ErrAccountDoesNotExist, errors.ErrAccountDoesNotExist.Error())
			return errors.ErrAccountDoesNotExist
		}

		// deactivate account activity status
		deactivateAccountOpStep := saga.Step{
			Name: "deactivate_business_account",
			Func: func(ctx context.Context) error {
				return db.SetBusinessAccountStatusAndSave(ctx, account, false)
			},
			CompensateFunc: func(ctx context.Context) error {
				return db.SetBusinessAccountStatusAndSave(ctx, account, true)
			},
			Options: nil,
		}

		if err := db.Saga.RunSaga(ctx, "deactivate_business_account", &deactivateAccountOpStep); err != nil {
			db.Logger.For(ctx).Error(err, err.Error())
			return err
		}

		return nil
	}

	err := db.Conn.PerformTransaction(ctx, tx)
	if err != nil {
		return err
	}

	return nil
}

// GetBusinessAccount gets a set of business accounts by ids
func (db *Db) GetBusinessAccounts(ctx context.Context, ids []uint32) ([]*proto.BusinessAccount, error) {
	// define initial log entry
	db.Logger.For(ctx).Info("get business accounts")
	ctx, span := db.startRootSpan(ctx, "get_business_accounts_db_op")
	defer span.Finish()

	tx := func(ctx context.Context, tx *gorm.DB) (interface{}, error) {
		// start child span
		db.Logger.For(ctx).Info("starting db transactions")
		childSpan := db.TracingEngine.CreateChildSpan(ctx, "get business accounts - operation")
		defer childSpan.Finish()

		var accounts = make([]*proto.BusinessAccount, len(ids)+1)

		for _, id := range ids {
			account := db.GetBusinessById(ctx, id)
			if account == nil {
				db.Logger.For(ctx).Error(errors.ErrAccountDoesNotExist, fmt.Sprintf("%s - for id %d ", errors.ErrAccountDoesNotExist.Error(), id))
			} else {
				accounts = append(accounts, account)
			}
		}

		return accounts, nil
	}

	res, err := db.Conn.PerformComplexTransaction(ctx, tx)
	if err != nil {
		return nil, err
	}

	return res.([]*proto.BusinessAccount), nil
}

func (db *Db) GetPaginatedBusinessAccounts(ctx context.Context, limit int64) ([]*proto.BusinessAccount, error) {
	// define initial log entry
	db.Logger.For(ctx).Info("get business accounts")
	ctx, span := db.startRootSpan(ctx, "get_paginated_business_accounts_db_op")
	defer span.Finish()

	tx := func(ctx context.Context, tx *gorm.DB) (interface{}, error) {
		// start child span
		db.Logger.For(ctx).Info("starting db transactions")
		childSpan := db.TracingEngine.CreateChildSpan(ctx, "get_paginated_business_account_db_tx")
		defer childSpan.Finish()

		var obtainedAccounts = make([]*proto.BusinessAccount, 0)
		var result []*proto.BusinessAccountORM
		if err := db.Conn.Engine.Limit(int(limit)).Find(&result).Error; err != nil {
			db.Logger.For(ctx).Error(errors.ErrUnableToObtainBusinessAccounts, err.Error())
			return nil, err
		}

		if len(result) == 0 {
			return []*proto.BusinessAccount{}, nil
		}

		for _, account := range result {
			obj, err := account.ToPB(ctx)
			if err != nil {
				db.Logger.For(ctx).Error(errors.ErrFailedToConvertFromOrmType, err.Error())
				return nil, err
			}

			obtainedAccounts = append(obtainedAccounts, &obj)
		}

		return obtainedAccounts, nil
	}

	res, err := db.Conn.PerformComplexTransaction(ctx, tx)
	if err != nil {
		return nil, err
	}

	return res.([]*proto.BusinessAccount), nil
}

// GetBusinessAccount gets a singular business account
func (db *Db) GetBusinessAccount(ctx context.Context, id uint32) (*proto.BusinessAccount, error) {
	// define initial log entry
	db.Logger.For(ctx).Info(fmt.Sprintf("get business account - id : %d", id))
	ctx, span := db.startRootSpan(ctx, "get_business_account_db_op")
	defer span.Finish()

	tx := func(ctx context.Context, tx *gorm.DB) (interface{}, error) {
		// start child span
		db.Logger.For(ctx).Info("starting db transactions")
		childSpan := db.TracingEngine.CreateChildSpan(ctx, "get_business_account_db_tx")
		defer childSpan.Finish()

		account := db.GetBusinessById(ctx, id)
		if account == nil {
			db.Logger.For(ctx).Error(errors.ErrAccountDoesNotExist, errors.ErrAccountDoesNotExist.Error())
			return nil, errors.ErrAccountDoesNotExist
		}

		db.Logger.For(ctx).Info("successfully obtained business account", zap.Any("id", fmt.Sprintf("%d", id)), zap.String("name",
			account.CompanyName))
		return account, nil
	}

	res, err := db.Conn.PerformComplexTransaction(ctx, tx)
	if err != nil {
		return nil, err
	}

	return res.(*proto.BusinessAccount), nil
}

// GetBusinessById gets a business account by id from the backend database
func (db *Db) GetBusinessById(ctx context.Context, id uint32) *proto.BusinessAccount {
	// define initial log entry
	db.Logger.For(ctx).Info(fmt.Sprintf("get business account by id - id : %d", id))
	ctx, span := db.startRootSpan(ctx, "get_business_account_by_id_op")
	defer span.Finish()

	tx := func(ctx context.Context, tx *gorm.DB) (interface{}, error) {
		// start child span
		db.Logger.For(ctx).Info("starting db transactions")
		childSpan := db.TracingEngine.CreateChildSpan(ctx, "get_business_account_by_id_tx")
		defer childSpan.Finish()

		var businessAccountOrm proto.BusinessAccountORM
		// attempt to see if the record already exists
		recordNotFoundErr := db.PreloadTx(tx).
			Where(&proto.BusinessAccountORM{Id: id}).
			First(&businessAccountOrm).Error
		if recordNotFoundErr != nil {
			db.Logger.For(ctx).Error(errors.ErrAccountDoesNotExist, "account does not exist")
			return nil, errors.ErrAccountDoesNotExist
		}

		// transform orm type to account type
		account, err := businessAccountOrm.ToPB(ctx)
		if err != nil {
			db.Logger.For(ctx).Error(errors.ErrFailedToConvertFromOrmType, err.Error())
			return nil, err
		}

		db.Logger.For(ctx).Info("successfully obtained business account", zap.String("id", fmt.Sprintf("%d", id)))
		return &account, nil
	}

	res, err := db.Conn.PerformComplexTransaction(ctx, tx)
	if err != nil {
		return nil
	}

	return res.(*proto.BusinessAccount)
}

// GetBusinessByEmail gets a business account by email from the backend database
func (db *Db) GetBusinessByEmail(ctx context.Context, email string) *proto.BusinessAccount {
	// define initial log entry
	db.Logger.For(ctx).Info("get business account by email")

	ctx, span := db.startRootSpan(ctx, "get_business_account_by_email_op")
	defer span.Finish()

	tx := func(ctx context.Context, tx *gorm.DB) (interface{}, error) {
		// start child span
		db.Logger.For(ctx).Info("starting db transactions")
		childSpan := db.TracingEngine.CreateChildSpan(ctx, "get_business_account_by_email_tx")
		defer childSpan.Finish()

		var businessAccountOrm proto.BusinessAccountORM

		// attempt to see if the record already exists
		recordNotFoundErr := db.PreloadTx(tx).
			Where(&proto.BusinessAccountORM{Email: email}).
			First(&businessAccountOrm).Error
		if recordNotFoundErr != nil {
			db.Logger.For(ctx).Error(errors.ErrAccountDoesNotExist, "account does not exist")
			return nil, errors.ErrAccountDoesNotExist
		}

		// transform orm type to account type
		account, err := businessAccountOrm.ToPB(ctx)
		if err != nil {
			db.Logger.For(ctx).Error(errors.ErrFailedToConvertFromOrmType, err.Error())
			return nil, err
		}

		db.Logger.For(ctx).Info("successfully obtained business account", zap.String("email", email))
		return &account, nil
	}

	res, err := db.Conn.PerformComplexTransaction(ctx, tx)
	if err != nil {
		return nil
	}

	return res.(*proto.BusinessAccount)
}

// CreateBusinessAccount creates a business account and saves it to the database
func (db *Db) CreateBusinessAccount(ctx context.Context, account *proto.BusinessAccount, authnid uint32) (*proto.BusinessAccount, error) {
	db.Logger.For(ctx).InfoM("creating business account")
	ctx, span := db.startRootSpan(ctx, "create_business_account_op")
	defer span.Finish()

	tx := func(ctx context.Context, tx *gorm.DB) (interface{}, error) {
		// start child span
		db.Logger.For(ctx).Info("starting transaction")
		span := db.TracingEngine.CreateChildSpan(ctx, "create_business_account_tx")
		defer span.Finish()

		// validate account object
		if err := account.Validate(); err != nil {
			db.Logger.Error(errors.ErrInvalidAccount, err.Error())
			return nil, err
		}

		// ensure authn id is valid
		if authnid == 0 || account.Email == "" || account.Password == "" || account.CompanyName == "" {
			db.Logger.Error(errors.ErrInvalidInputArguments, fmt.Sprintf("authn id cannot be %d", authnid))
			return nil, errors.ErrInvalidInputArguments
		}

		var businessAccount proto.BusinessAccountORM
		// attempt to see if the record already exists
		// no 2 records in our backend database can have the same email or company name
		recordNotFoundErr := db.PreloadTx(tx).Where(&proto.BusinessAccountORM{Email: account.Email,
			CompanyName: account.CompanyName}).First(&businessAccount).Error

		if recordNotFoundErr == nil {
			// account already exists
			db.Logger.ErrorM(errors.ErrAccountAlreadyExist, errors.ErrAccountAlreadyExist.Error())
			return nil, errors.ErrAccountAlreadyExist
		}

		// if the account does not exist we save it in the db
		// convert it first to orm type
		businessAccount, err := account.ToORM(ctx)
		if err != nil {
			db.Logger.ErrorM(errors.ErrFailedToConvertToOrmType, err.Error())
			return nil, err
		}

		// hash password
		if businessAccount.Password, err = db.Conn.ValidateAndHashPassword(businessAccount.Password); err != nil {
			db.Logger.For(ctx).Error(errors.ErrFailedToHashPassword, err.Error())
			return nil, err
		}

		// activate account and assign authn id relation
		businessAccount.IsActive = true
		businessAccount.AuthnId = authnid

		// save the account record in the db
		if err := tx.Create(&businessAccount).Error; err != nil {
			db.Logger.For(ctx).Error(errors.ErrFailedToCreateAccount, err.Error())
			return nil, err
		}

		createdAccount, err := businessAccount.ToPB(ctx)
		if err != nil {
			db.Logger.For(ctx).Error(errors.ErrFailedToConvertFromOrmType, err.Error())
			return nil, err
		}

		return &createdAccount, nil
	}

	result, err := db.Conn.PerformComplexTransaction(ctx, tx)
	if err != nil {
		return nil, err
	}

	createdAccount := result.(*proto.BusinessAccount)
	return createdAccount, nil
}

// SetBusinessAccountStatusAndSave updates the active status of a business account in the backend database
func (db *Db) SetBusinessAccountStatusAndSave(ctx context.Context, businessAccount *proto.BusinessAccount,
	activateAccount bool) error {

	db.Logger.For(ctx).Info(fmt.Sprintf("updating business account active status to %v", activateAccount))
	ctx, span := db.startRootSpan(ctx, "set_business_account_active_status")
	defer span.Finish()

	tx := func(ctx context.Context, tx *gorm.DB) error {
		// convert to orm type
		account, err := businessAccount.ToORM(ctx)
		if err != nil {
			db.Logger.For(ctx).Error(errors.ErrFailedToConvertToOrmType, err.Error())
			return err
		}

		// set account active status
		if err = tx.Clauses(clause.OnConflict{
			UpdateAll: true,
		}).Model(&proto.BusinessAccountORM{}).Where("email = ?", account.Email).Update("is_active", activateAccount).Error; err != nil {
			db.Logger.For(ctx).Error(errors.ErrFailedToUpdateAccountActiveStatus, err.Error())
			return err
		}

		return nil
	}

	f := func() error {
		return db.Conn.PerformTransaction(ctx, tx)
	}

	// should perform this as a retryable operation in case of errors
	return db.PerformRetryableOperation(f)
}
