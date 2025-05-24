package db

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
)

const (
	TransactionTypeDebit      = "debit"
	TransactionTypeCredit     = "credit"
	TransactionSourceSwap     = "swap"
	TransactionSourceAdmin    = "admin-topup"
	TransactionSourceWallet   = "wallet"
	TransactionSourcePaystack = "paystack"
	TransactionSourceBudPay   = "budpay"

	TransactionStatusPending      = "pending"
	TransactionStatusSwapApproved = "swap-approved"
	TransactionStatusCompleted    = "completed"
	TransactionStatusFailed       = "failed"
	TransactionStatusCanceled     = "canceled"
	TransactionStatusProcessing   = "processing"
	TransactionStatusIssue        = "issue"

	TransactionPaymentMethodBankTransfer = "bank_transfer"
	TransactionPaymentMethodManual       = "manual"

	TransactionActionSwap                      = "swap"
	TransactionActionSwapRefund                = "refund"
	TransactionActionExternalTransfer          = "ext-transfer"
	TransactionActionAffiliateMerchantTransfer = "merchant-transfer"
	TransactionActionBankTransfer              = "bank-transfer"
	TransactionActionTransferRefund            = "transfer-refund"
	TransactionActionFundAccount               = "fund_account"

	SettlementStatusNew       = "new"
	SettlementStatusCompleted = "completed"
	SettlementStatusDelayed   = "delayed"
	SettlementStatusNone      = "none"
)

var ValidTransactionActions = []string{
	TransactionActionSwap,
	TransactionActionSwapRefund,
	TransactionActionExternalTransfer,
	TransactionActionBankTransfer,
	TransactionActionTransferRefund,
	TransactionActionFundAccount,
}

func (store *SQLStore) PerformTransaction(
	ctx context.Context,
	wallet *Wallet,
	args CreateTransactionParams,
	transactionKey []byte,
) (*Transaction, error) {

	if !VerifyWallet(wallet, transactionKey) {
		return nil, fmt.Errorf("wallet integrity check failed")
	}

	transaction, err := store.performTransactionInTx(ctx, wallet, args, transactionKey)
	if err != nil {
		return nil, err
	}

	return transaction, nil
}

func (store *SQLStore) performTransactionInTx(
	ctx context.Context,
	wallet *Wallet,
	args CreateTransactionParams,
	transactionKey []byte,
) (*Transaction, error) {

	var transaction Transaction
	switch args.Type {
	case TransactionTypeDebit:
		wallet.Balance = wallet.Balance.Sub(args.Amount)
	case TransactionTypeCredit:
		wallet.Balance = wallet.Balance.Add(args.Amount)
	default:
		return nil, fmt.Errorf("invalid transaction type")
	}

	err := store.execTx(ctx, func(q *Queries) error {

		currentWallet, err := q.GetWallet(ctx, wallet.ID)
		if err != nil {
			return err
		}
		if wallet.Version != currentWallet.Version {
			return fmt.Errorf("wallet version mismatch, retry transaction")
		}
		currentWallet, err = q.UpdateWalletBalance(ctx, UpdateWalletBalanceParams{ID: wallet.ID, Balance: wallet.Balance})

		if err != nil {
			return err
		}
		newHash := GenerateWalletHash(&currentWallet, transactionKey)

		walletNew, err := q.UpdateWalletHash(ctx, UpdateWalletHashParams{
			Hash: newHash,
			ID:   wallet.ID,
		})
		if err != nil {
			return err
		}
		wallet = &walletNew

		args.WalletID = wallet.ID
		args.UserID = wallet.UserID
		transaction, err = q.CreateTransaction(ctx, args)
		if err != nil {
			return err
		}

		if !VerifyWallet(wallet, transactionKey) {
			return fmt.Errorf("wallet integrity check failed: %s", wallet.ID.String())
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	err = store.CreateTransactionHistory(ctx, CreateTransactionHistoryParams{
		TransactionID: transaction.ID,
		UserID:        args.UserID,
		Reason:        args.Tag,
		Amount:        args.Amount,
		OldStatus:     transaction.Status,
		NewStatus:     transaction.Status,
		Payload:       []byte("{}"),
	})

	return &transaction, err
}

// GenerateWalletHash generates a hash for the wallet
func GenerateWalletHash(wallet *Wallet, key []byte) string {
	h := hmac.New(sha256.New, key)

	data := fmt.Sprintf("%s %s", wallet.ID, wallet.UserID) +
		fmt.Sprintf("%d", wallet.CurrencyID) + wallet.Balance.String() +
		wallet.CreatedAt.String() + strconv.FormatBool(wallet.Locked)

	h.Write([]byte(data))

	return hex.EncodeToString(h.Sum(nil))
}

// VerifyWallet verifies the integrity of the wallet
func VerifyWallet(wallet *Wallet, key []byte) bool {
	expectedHash := GenerateWalletHash(wallet, key)
	return hmac.Equal([]byte(wallet.Hash), []byte(expectedHash))
}
