package db

import (
	"context"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestGenerateWalletHash(t *testing.T) {
	user := createRandomUser(t, "Personal")
	currency := createRandomCurrency(t)

	arg := CreateWalletParams{
		UserID:     user.ID,
		CurrencyID: currency.ID,
	}
	wallet, err := testQueries.CreateWallet(context.Background(), arg)
	assert.NoError(t, err)
	assert.NotNil(t, wallet)

	secretKey := []byte("test_secret_key")
	hash := GenerateWalletHash(&wallet, secretKey)

	assert.NotEmpty(t, hash)
}

func TestVerifyWallet(t *testing.T) {
	user := createRandomUser(t, "Personal")
	currency := createRandomCurrency(t)

	arg := CreateWalletParams{
		UserID:     user.ID,
		CurrencyID: currency.ID,
	}
	wallet, err := testQueries.CreateWallet(context.Background(), arg)
	assert.NoError(t, err)
	assert.NotNil(t, wallet)

	secretKey := []byte("test_secret_key")
	hash := GenerateWalletHash(&wallet, secretKey)

	wallet, err = testQueries.UpdateWalletHash(context.Background(), UpdateWalletHashParams{
		Hash: hash,
		ID:   wallet.ID,
	})
	assert.NoError(t, err)

	updatedWallet, err := testQueries.GetWallet(context.Background(), wallet.ID)
	assert.NoError(t, err)
	assert.NotNil(t, updatedWallet)

	isVerified := VerifyWallet(&updatedWallet, secretKey)
	assert.True(t, isVerified)

	wrongKey := []byte("wrong_secret_key")
	isVerified = VerifyWallet(&updatedWallet, wrongKey)
	assert.False(t, isVerified)
}

func TestPerformTransaction(t *testing.T) {
	store := NewStore(testDB, nil)

	user := createRandomUser(t, "Personal")
	currency := createRandomCurrency(t)

	wallet := createRandomWallet(t, user.ID, currency.ID)
	assert.NotNil(t, wallet)

	secretKey := []byte("test_secret_key")
	hash := GenerateWalletHash(wallet, secretKey)

	walletNew, err := testQueries.UpdateWalletHash(context.Background(), UpdateWalletHashParams{
		Hash: hash,
		ID:   wallet.ID,
	})
	assert.NoError(t, err)
	assert.NotNil(t, walletNew)
	wallet = &walletNew

	// Perform a credit transaction
	amount, _ := decimal.NewFromString("100.00")
	arg := CreateTransactionParams{
		WalletID:   wallet.ID,
		Amount:     amount,
		Type:       TransactionTypeCredit,
		UserID:     user.ID,
		CurrencyID: currency.ID,
		Payload:    []byte("{}"),
	}

	transaction, err := store.PerformTransaction(context.Background(), wallet, arg, secretKey)
	assert.NoError(t, err)
	assert.NotNil(t, transaction)
	assert.Equal(t, arg.Type, transaction.Type)

	// Retrieve the updated wallet and compare the balance
	updatedWallet := getWalletByID(t, wallet.ID)
	assert.True(t, amount.Equal(updatedWallet.Balance), "amount mismatch: expected %s, got %s", amount.String(), updatedWallet.Balance.String())

	// Perform a debit transaction
	arg.Type = TransactionTypeDebit
	transaction, err = store.PerformTransaction(context.Background(), updatedWallet, arg, secretKey)
	assert.NoError(t, err)
	assert.NotNil(t, transaction)
	assert.Equal(t, arg.Type, transaction.Type)

	// Retrieve the updated wallet again and compare the balance
	updatedWallet2 := getWalletByID(t, wallet.ID)
	assert.True(t, updatedWallet2.Balance.IsZero())

	// Perform an invalid transaction
	arg.Type = "invalid"
	transaction, err = store.PerformTransaction(context.Background(), updatedWallet2, arg, secretKey)
	assert.Error(t, err)
	assert.Nil(t, transaction)

	arg.Type = TransactionTypeCredit
	updatedWallet2.Hash = "invalid_hash"
	transaction, err = store.PerformTransaction(context.Background(), updatedWallet2, arg, secretKey)
	assert.Error(t, err)
	assert.Equal(t, "wallet integrity check failed", err.Error())
	assert.Nil(t, transaction)

}

func TestPerformTransactionInTx_Rollbacks(t *testing.T) {
	ctx := context.Background()
	transactionKey := []byte("secret_key")

	user := createRandomUser(t, "Personal")
	currency := createRandomCurrency(t)

	wallet := createRandomWallet(t, user.ID, currency.ID)
	assert.NotNil(t, wallet)

	wallet.Hash = GenerateWalletHash(wallet, transactionKey)

	store := SQLStore{
		db:      testDB,
		Queries: testQueries,
	}

	invalidWallet := wallet
	invalidWallet.UserID = uuid.New()

	amount, _ := decimal.NewFromString("100.00")
	arg := CreateTransactionParams{
		WalletID:   wallet.ID,
		Amount:     amount,
		Type:       TransactionTypeCredit,
		CurrencyID: currency.ID,
		Payload:    []byte("{"), // trigger an error
	}

	_, err := store.performTransactionInTx(ctx, invalidWallet, arg, transactionKey)
	assert.Error(t, err)

	//// Check if the wallet balance is still the same after the error
	walletAfterError := getWalletByID(t, wallet.ID)
	assert.True(t, walletAfterError.Balance.IsZero())

	arg.Type = "invalid"
	_, err = store.performTransactionInTx(ctx, wallet, arg, transactionKey)
	assert.Error(t, err)
	assert.Equal(t, `invalid transaction type`, err.Error())
}

func TestPerformTransaction_ConcurrentOperations(t *testing.T) {
	ctx := context.Background()
	transactionKey := []byte("secret_key")

	user := createRandomUser(t, "Personal")
	currency := createRandomCurrency(t)

	wallet := createRandomWallet(t, user.ID, currency.ID)
	assert.NotNil(t, wallet)

	store := SQLStore{
		db:      testDB,
		Queries: testQueries,
	}

	numGoroutines := 10
	amount, _ := decimal.NewFromString("1.00")
	arg := CreateTransactionParams{
		WalletID:   wallet.ID,
		Amount:     amount,
		CurrencyID: currency.ID,
		Payload:    []byte("{}"),
	}

	initialBalance := wallet.Balance
	expectedFinalBalance := initialBalance

	var wg sync.WaitGroup
	wg.Add(numGoroutines * 2)

	// Mutex to synchronize access to the wallet
	var mutex sync.Mutex

	// Perform concurrent debit transactions
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()

			mutex.Lock()
			wallet = getWalletByID(t, wallet.ID) // Reload the wallet before retrying
			arg.Type = TransactionTypeDebit
			_, err := store.performTransactionInTx(ctx, wallet, arg, transactionKey)
			assert.NoError(t, err)
			mutex.Unlock()
		}()
	}

	// Perform concurrent credit transactions
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()

			mutex.Lock()
			wallet = getWalletByID(t, wallet.ID) // Reload the wallet before retrying
			arg.Type = TransactionTypeCredit
			_, err := store.performTransactionInTx(ctx, wallet, arg, transactionKey)
			assert.NoError(t, err)
			mutex.Unlock()
		}()
	}

	wg.Wait()

	walletAfterOperations := getWalletByID(t, wallet.ID)
	assert.True(t, expectedFinalBalance.Equal(walletAfterOperations.Balance), "wallet balance mismatch")
}

func TestPerformTransactionInTx(t *testing.T) {
	ctx := context.Background()
	transactionKey := []byte("secret_key")

	user := createRandomUser(t, "Personal")
	currency := createRandomCurrency(t)

	wallet := createRandomWallet(t, user.ID, currency.ID)
	assert.NotNil(t, wallet)

	store := SQLStore{
		db:      testDB,
		Queries: testQueries,
	}

	hash := GenerateWalletHash(wallet, transactionKey)

	walletNew, err := testQueries.UpdateWalletHash(context.Background(), UpdateWalletHashParams{
		Hash: hash,
		ID:   wallet.ID,
	})
	assert.NoError(t, err)
	assert.NotNil(t, walletNew)

	wallet = &walletNew

	t.Run("wallet version mismatch", func(t *testing.T) {
		// Create a wallet with an outdated version
		outdatedWallet := wallet
		outdatedWallet.Version--

		amount, _ := decimal.NewFromString("10")
		arg := CreateTransactionParams{
			WalletID:   wallet.ID,
			Amount:     amount,
			Type:       TransactionTypeDebit,
			CurrencyID: currency.ID,
			Payload:    []byte("{}"),
		}

		_, err := store.performTransactionInTx(ctx, outdatedWallet, arg, transactionKey)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "wallet version mismatch")
	})

}

func createRandomWallet(t *testing.T, userID uuid.UUID, currencyID int32) *Wallet {
	arg := CreateWalletParams{
		UserID:     userID,
		CurrencyID: currencyID,
	}
	wallet, err := testQueries.CreateWallet(context.Background(), arg)
	assert.NoError(t, err)
	assert.NotNil(t, wallet)

	return &wallet
}

func getWalletByID(t *testing.T, id uuid.UUID) *Wallet {
	wallet, err := testQueries.GetWallet(context.Background(), id)
	assert.NoError(t, err)
	assert.NotNil(t, wallet)
	return &wallet
}
