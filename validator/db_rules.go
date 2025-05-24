package validator

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	db "github.com/timchuks/monieverse/internal/db/sqlc"
)

// EmailExists checks if an email exists.
func (v *Validator) EmailExists(email string) {
	_, err := v.store.GetUserByEmail(v.ctx, email)
	v.Check(err != nil, "email", "email not valid on this platform")
}

// EmailExistsForOthers checks if an email exists for other users.
func (v *Validator) EmailExistsForOthers(email string, userID uuid.UUID) {
	_, err := v.store.GetOtherUserByEmail(v.ctx, db.GetOtherUserByEmailParams{
		Email: email,
		ID:    userID,
	})
	v.Check(err != nil, "email", "this email is not available")
}

// PhoneExists checks if a phone number exists.
func (v *Validator) PhoneExists(phone string) {
	_, err := v.store.GetUserByPhone(v.ctx, phone)
	v.Check(err != nil, "phone", "phone not valid on this platform")
}

// PhoneExistsForOthers checks if a phone number exists for other users.
func (v *Validator) PhoneExistsForOthers(phone string, userID uuid.UUID) {
	_, err := v.store.GetOtherUserByPhone(v.ctx, db.GetOtherUserByPhoneParams{
		Phone: phone,
		ID:    userID,
	})
	v.Check(err != nil, "phone", "this phone is not available")
}

func (v *Validator) UserIDExists(userID uuid.UUID) {
	_, err := v.store.GetUser(v.ctx, userID)
	v.Check(err == nil, "user_id", "user_id doesn't exist")
}

func (v *Validator) DocumentIDExists(id uuid.UUID) {
	_, err := v.store.GetOneDocumentByID(v.ctx, id)
	v.Check(err == nil, "id", "id doesn't exist")
}

// HasPermission checks if a user has a permission
func (v *Validator) HasPermission(user db.User, perm string) bool {
	return v.store.HasPermission(context.Background(), user, perm)
}

func (v *Validator) UserExists(userID uuid.UUID) *db.User {
	u, err := v.store.GetUser(v.ctx, userID)
	v.Check(err == nil, "user", "user doesn't exist")
	return &u
}

func (v *Validator) KYCRequirementExists(target, targetID string) (*db.KycRequirement, error) {
	kyc, err := v.store.GetActiveKYCRequirement(v.ctx, db.GetActiveKYCRequirementParams{
		Target:   target,
		TargetID: targetID,
	})

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err

	}

	return &kyc, nil

}

func (v *Validator) KYCRequirementByID(ID uuid.UUID) (*db.KycRequirement, error) {
	kyc, err := v.store.GetKYCRequirementByID(v.ctx, ID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err

	}
	return &kyc, nil
}

// PermissionExistForUser checks if the a permission has already been assigned to a user
func (v *Validator) PermissionExistForUser(permissionID int64, userID uuid.UUID) {
	args := db.GetPermissionForUserParams{
		UserID:       userID,
		PermissionID: permissionID,
	}
	_, err := v.store.GetPermissionForUser(v.ctx, args)

	v.Check(err != nil, fmt.Sprintf("permission_id-%v-user", permissionID), fmt.Sprintf("permission with id %v has already been assigned to user", permissionID))
}

// CanCurrencyHaveWallet checks if the currency can have wallet
func (v *Validator) CanCurrencyHaveWallet(id int32) db.Currency {

	c := v.CurrencyExists(id)

	v.Check(c.CanHaveWallet, "currency", "currency cannot have wallet")
	return c
}

// CanSwapFromCurrency checks if the currency can be swapped from
func (v *Validator) CanSwapFromCurrency(id int32, label string) db.Currency {
	c := v.CurrencyExists(id)
	v.Check(c.CanSwapFrom, label, fmt.Sprintf("swapping from %s is not supported", c.Code))
	return c
}

// CanSwapToCurrency checks if the currency can be swapped to
func (v *Validator) CanSwapToCurrency(id int32, label string) db.Currency {
	c := v.CurrencyExists(id)
	v.Check(c.CanSwapTo, label, fmt.Sprintf("swapping to %s is not supported", c.Code))
	return c
}

// CurrencyExists checks if a currency exists
func (v *Validator) CurrencyExists(id int32) db.Currency {
	c, err := v.store.GetCurrency(v.ctx, id)
	if err != nil {
		v.Check(err != nil, "currency", "currency does not exist")
		return db.Currency{}
	}
	v.Check(c.Active, "currency", fmt.Sprintf("currency %s is not active", c.Code))
	return c
}

func (v *Validator) CurrencyExistsByCode(code string) db.Currency {
	c, err := v.store.GetCurrencyByCode(v.ctx, code)
	if err != nil {
		v.Check(err != nil, "currency", "currency does not exist")
		return db.Currency{}
	}
	v.Check(c.Active, "currency", fmt.Sprintf("currency %s is not active", c.Code))
	return c
}

func (v *Validator) ValidateExchangeRateExistence(baseCurrencyID int32, quoteCurrencyID int32, rateType string) (bool, error) {

	count, err := v.store.ValidateExchangeRateExistence(v.ctx, db.ValidateExchangeRateExistenceParams{
		BaseCurrencyID:  quoteCurrencyID,
		QuoteCurrencyID: baseCurrencyID,
		Type:            rateType,
	})
	if err != nil {
		return false, err
	}

	if count > 0 {
		return true, nil
	}
	return false, nil
}

func (v *Validator) ValidateAccountLevelExchangeRateExistence(userID uuid.UUID, baseCurrencyID int32, quoteCurrencyID int32, rateType string) (bool, error) {

	args := db.ValidateAccountLevelExchangeRateExistenceParams{
		BaseCurrencyID:  quoteCurrencyID,
		QuoteCurrencyID: baseCurrencyID,
		Type:            rateType,
		UserID:          userID,
	}
	count, err := v.store.ValidateAccountLevelExchangeRateExistence(v.ctx, args)

	if err != nil {
		return false, err
	}

	if count > 0 {
		return true, nil
	}
	return false, nil
}

// CanTokenBeResent checks if a user can resend a token
func (v *Validator) CanTokenBeResent(email, scope string) {
	_, err := v.store.GetUserByEmail(v.ctx, email)
	v.Check(err == nil, "email", "email doesn't exist")

	token, err := v.store.FindExistingUserToken(v.ctx, db.FindExistingUserTokenParams{
		Email: email,
		Scope: scope,
	})
	v.Check(err == sql.ErrNoRows, "token", fmt.Sprintf("user has an existing valid token which will expire in %v mins", time.Until(token.Expiry).Minutes()))
}

// WalletExistsByCurrency checks if the currency can be swapped from
func (v *Validator) WalletExistsByCurrency(userID uuid.UUID, currencyID int32) *db.Wallet {

	wallet, err := v.store.GetUserWalletByCurrency(v.ctx, db.GetUserWalletByCurrencyParams{
		UserID:     userID,
		CurrencyID: currencyID,
	})
	if err != nil {
		v.Check(err != nil, "wallet", "wallet does not exist")
		return nil
	}

	return &wallet
}

// WalletExists checks if a wallet exists
func (v *Validator) WalletExists(walletID uuid.UUID) *db.Wallet {

	wallet, err := v.store.GetWallet(v.ctx, walletID)
	if err != nil {
		v.Check(false, "wallet", "wallet does not exist")
		return nil
	}

	return &wallet
}

// UserRecipientExists checks if a recipient exists
func (v *Validator) UserRecipientExists(recipientID uuid.UUID, userID uuid.UUID) *db.Recipient {

	recipient, err := v.store.GetUserRecipient(v.ctx, db.GetUserRecipientParams{
		ID:     recipientID,
		UserID: userID,
	})
	if err != nil {
		v.Check(false, "recipient", "recipient does not exist")
		return nil
	}

	return &recipient
}

func (v *Validator) FindCustomerByUserID(userID uuid.UUID, ownerID uuid.UUID) *db.Customer {
	c, err := v.store.FindCustomerById(v.ctx, db.FindCustomerByIdParams{
		ID:      userID,
		OwnerID: ownerID,
	})
	if err != nil {
		v.Check(false, "customer", "customer does not exist")
		return &db.Customer{}
	}
	return &c
}

// GetPaymentFeesConfig checks if a payment fee config exists
func (v *Validator) GetPaymentFeesConfig(scheme string) *db.SchemaPaymentFeesConfig {
	pfc, err := v.store.GetSchemaPaymentFeeConfig(v.ctx, scheme)
	if err != nil {
		return &db.SchemaPaymentFeesConfig{}
	}
	return &pfc
}

// TransactionExists checks if a transaction exists
func (v *Validator) TransactionExists(id uuid.UUID) db.Transaction {
	tx, err := v.store.GetTransaction(v.ctx, id)
	if err != nil {
		v.Check(err != nil, "transaction", "transaction does not exist: "+err.Error())
		return db.Transaction{}
	}
	return tx
}

// DealerExists checks if a dealer exists
func (v *Validator) DealerExists(id uuid.UUID) db.Dealer {
	dl, err := v.store.GetDealer(v.ctx, id)
	if err != nil {
		v.Check(err != nil, "dealer", "dealer does not exist: "+err.Error())
		return db.Dealer{}
	}
	return dl
}

// BankExistsByCode check if the bank exists by code
func (v *Validator) BankExistsByCode(code string) db.Bank {
	tx, err := v.store.GetBankByCode(v.ctx, code)
	if err != nil {
		v.Check(false, "bank", "bank does not exist: "+err.Error())
		return db.Bank{}
	}
	return tx
}

// ReferralPhoneExists checks if a phone number exists.
func (v *Validator) ReferralPhoneExists(phone string) {
	_, err := v.store.GetUserByPhone(v.ctx, phone)
	v.Check(err != nil, "phone", "phone already exists as a user of monieverse")

	_, err = v.store.GetReferralByPhone(v.ctx, phone)
	v.Check(err != nil, "phone", "phone already is referral on monieverse")
}

// CompanyExists checks if a company with short name exists.
func (v *Validator) CompanyExists(name string) {
	_, err := v.store.GetCompanyByShortName(v.ctx, name)
	v.Check(err != nil, "short_name", "company with short name already exists")
}

// HasIncompleteRegistrationViaEmail checks if a user has an incomplete registration
func (v *Validator) HasIncompleteRegistrationViaEmail(email string) *db.User {
	user, err := v.store.GetUserByEmail(v.ctx, email)
	if err == nil && !user.IsVerified() {
		utx, _ := v.store.GetUserWallets(v.ctx, user.ID)
		if len(utx) == 0 {
			return &user
		}
	}
	return nil
}

// HasIncompleteRegistrationViaPhone checks if a user has an incomplete registration
func (v *Validator) HasIncompleteRegistrationViaPhone(phone string) *db.User {
	user, err := v.store.GetUserByPhone(v.ctx, phone)
	if err == nil && !user.IsVerified() {
		utx, _ := v.store.GetUserWallets(v.ctx, user.ID)
		if len(utx) == 0 {
			return &user
		}
	}
	return nil
}

func (v *Validator) ValidCountryCode(code string) *db.Country {
	c, err := v.store.GetCountry(v.ctx, code)
	if err != nil {
		v.Check(false, "country", "country not valid")
		return nil
	}

	if !c.Enabled {
		v.Check(false, "country", "country not valid")
		return nil
	}

	return c
}

func (v *Validator) BusinessNameAlreadyExists(name string) {
	_, err := v.store.GetBusinessByName(v.ctx, name)
	v.Check(err != nil, "name", "name already exist")
}

func (v *Validator) BusinessNameShouldExists(name string) {
	business, _ := v.store.GetBusinessByName(v.ctx, name)
	v.Check(business.Name == name, "name", "business name cannot be changed")
}

func (v *Validator) BusinessExists(id uuid.UUID) {
	_, err := v.store.GetBusinessByID(v.ctx, id)
	if err != nil {
		v.Check(false, "business", "business does not exist")
	}
}

func (v *Validator) BusinessOwnerExists(ownerID, businessID uuid.UUID) {
	_, err := v.store.GetBusinessOwnerByID(v.ctx, db.GetBusinessOwnerByIDParams{
		ID:         ownerID,
		BusinessID: businessID,
	})
	if err != nil {
		v.Check(false, "business", "business does not exist")
	}
}

func (v *Validator) IsBusinessCreatedByRequestMaker(userID, businessID uuid.UUID) {
	_, err := v.store.GetBusinessCreatedByUserByBusinessID(v.ctx, db.GetBusinessCreatedByUserByBusinessIDParams{
		CreatedBy: userID,
		ID:        businessID,
	})
	if err != nil {
		v.Check(false, "business", "only business created by you can be updated")
	}
}
