package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/timchuks/monieverse/internal/mapper"
	"github.com/timchuks/monieverse/internal/settings"
)

var (
	AnonymousUser = &User{}
)

type Store interface {
	Querier
	CreateUserTx(ctx context.Context, arg CreateUserTxParams, afterCreate AfterCreateUserFunc) (CreateUserTxResult, error)
	GetOneTokenForUser(ctx context.Context, userID uuid.UUID, scope string) (Token, error)
	AddUserPermission(ctx context.Context, userID uuid.UUID, codes ...string) error
	PerformTransaction(ctx context.Context, wallet *Wallet, arg CreateTransactionParams, transactionKey []byte) (*Transaction, error)
	GetPaginatedExchangeRate(ctx context.Context, filter Filter) ([]GetPaginatedExchangeRateRow, Metadata, error)
	GetPaginatedAccountExchangeRate(ctx context.Context, filter Filter, userID uuid.UUID) ([]GetPaginatedAccountExchangeRateRow, Metadata, error)
	GetPaginatedUsersList(ctx context.Context, filter UserListFilter) ([]GetPaginatedUserRow, Metadata, error)
	CreateIdentityDocumentTx(ctx context.Context, before BeforeIDDocCreateFunc) (*UserIdentityDocument, error)
	CreateDocumentTx(ctx context.Context, before BeforeDocumentCreateFunc) (*Document, error)
	UpdateDocumentTx(ctx context.Context, before BeforeDocumentUpdateFunc) (*Document, error)
	GetPaginatedSwapRequests(ctx context.Context, filter *TransactionFilter) ([]SwapRequestTransactionRow, Metadata, error)
	GetPaginatedRecipients(ctx context.Context, filter *RecipientFilter) ([]Recipient, Metadata, error)
	GetUserSettings(ctx context.Context, userID uuid.UUID) (settings.UserSettings, error)
	UpdateTransactionTx(ctx context.Context, arg UpdateTransactionTxParams, afterUpdate AfterTransactionUpdateFunc) (UpdateTransactionTxResult, error)
	GetPaginatedTransactions(ctx context.Context, filter *TransactionFilter) ([]TransactionRow, Metadata, error)
	AdminGetPaginatedTransactionList(ctx context.Context, filter *TransactionFilter) ([]GetPaginatedTransactionRow, Metadata, error)
	GetPaginatedTransfers(ctx context.Context, filter *TransactionFilter) ([]TransactionRow, Metadata, error)
	CreateSettlementTx(ctx context.Context, arg CreateSettlementTxParams, afterCreate AfterCreateSettlementFunc) (CreateSettlementTxResult, error)
	HasPermission(ctx context.Context, user User, perm string) bool
	GetCustomers(ctx context.Context, filter CustomerFilter) ([]Customer, Metadata, error)
	GetPaginatedUserActions(ctx context.Context, filter *UserActionFilter) ([]UserAction, Metadata, error)
	CreateKYCRequirementResultEntry(ctx context.Context, kycRequirementID uuid.UUID, values map[string]string) error
	MoneySendrGetPaginatedTransactions(ctx context.Context, filter Filter) ([]MoneysendrRecipient, Metadata, error)
	GetUserMetas(ctx context.Context, userID uuid.UUID) (UserMeta, error)
	UserMetaExists(ctx context.Context, userID uuid.UUID, key string) (bool, error)
	SetUserMeta(ctx context.Context, input UserMetaCreateParams) error
	GetSystemUser(email string) (*User, error)
	GetCountry(ctx context.Context, identifier interface{}) (*Country, error)
	GetCountries(ctx context.Context, queryFilter CountryQueryFilter, pagination Filter) ([]Country, error)
	AddSupportedCurrency(ctx context.Context, countryID, currencyID int, accountType string) error
	RemoveSupportedCurrency(ctx context.Context, countryID, currencyID int, accountType string) error
	IsCountryCurrencySupported(ctx context.Context, countryIdentifier, currencyIdentifier interface{}, accountType string) (bool, error)
	GetCountrySupportedCurrencies(ctx context.Context, countryIdentifier interface{}, accountType string) ([]Currency, error)
	GetEnabledCountries(ctx context.Context) ([]SimpleCountry, error)
	EasyeuroWebhookCreate(ctx context.Context, arg EasyeuroWebhook) (*EasyeuroWebhook, error)
	EasyeuroWebhookFindOneByURLAndEventType(ctx context.Context, url string, eventType string) (*EasyeuroWebhook, error)
	EasyeuroWebhookFindOne(ctx context.Context, id string) (*EasyeuroWebhook, error)
	GetFailedLogin(ctx context.Context, userIdentity, provider string) (*FailedLogin, error)
	CreateFailedLogin(ctx context.Context, userIdentity, provider string) error
	UpdateFailedLogin(ctx context.Context, failedLogin *FailedLogin) error
	ResetFailedLogin(ctx context.Context, userIdentity, provider string) error
	GetBannedUntil(ctx context.Context, userIdentity, provider string) (*time.Time, error)
	GetTransactionRecipientDataByID(ctx context.Context, id uuid.UUID) (*TransactionRecipientData, error)
	UpdateTransactionRecipientData(ctx context.Context, newData interface{}, id uuid.UUID) error
	GetKYCRequirementsForUser(ctx context.Context, arg GetKYCRequirementsForUserParams) ([]KycRequirement, error)
	SetBusinessApprovalStatus(ctx context.Context, id uuid.UUID, status BusinessStatus, reason string, approvedBy uuid.UUID) error
	SetBusinessOwnersApprovalStatus(ctx context.Context, id uuid.UUID, status BusinessStatus, reason string, approvedBy uuid.UUID) error
	PaginatedBusinesses(ctx context.Context, filter BusinessListFilter) ([]Business, error)
	VerifyWalletHistory(ctx context.Context, id int64) (bool, error)
	UpdateWalletHistoryStatusTx(ctx context.Context, arg UpdateWalletHistoryStatusParams) (*WalletHistory, error)
	GetWalletHistory(ctx context.Context, id int64) (*WalletHistory, error)
	CreateWalletHistoryTx(ctx context.Context, arg CreateWalletHistoryParams) (*WalletHistory, error)
	CreateFormDefinitionTx(ctx context.Context, input *FormDefinitionInput) (*FormDefinition, error)
	ProcessFormSubmissionTx(ctx context.Context, input *FormSubmissionInput) (*FormSubmission, error)
	UpdateFormSubmissionTx(ctx context.Context, input *FormSubmissionUpdateInput) (*FormSubmission, error)
	SaveStepProgressTx(ctx context.Context, input *SaveStepProgressInput) error
}

type SQLStore struct {
	*Queries
	db     *sql.DB
	mapper mapper.ConfigMapper
}

func NewStore(db *sql.DB, mapper mapper.ConfigMapper) Store {
	return &SQLStore{
		db:      db,
		Queries: New(db),
		mapper:  mapper,
	}
}

// execTx executes a function within a database transaction
func (store *SQLStore) execTx(ctx context.Context, fn func(*Queries) error) error {

	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	q := New(tx)

	err = fn(q)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err: %v, rbErr: %v", err, rbErr)
		}

		return err
	}

	return tx.Commit()
}

// IsVerified returns true if the user has been verified
func (u *User) IsVerified() bool {
	return u.Active
}

func (u *User) IsAnonymous() bool {
	return u == nil || u == AnonymousUser || u.ID == uuid.Nil
}

func (u *User) IsBusiness() bool {
	return u != nil && strings.ToLower(u.AccountType) == "business"
}

func (u *User) GetID() string {
	if u.IsAnonymous() {
		return ""
	}

	return u.ID.String()
}
