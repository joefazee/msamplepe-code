package db

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/timchuks/monieverse/internal/common"
)

func TestQueries_CreateAffiliateMerchant(t *testing.T) {
	user := createRandomUser(t, "Business")
	apiKey := common.RandomAPIKey()

	// Define test parameters
	testParams := CreateAffiliateMerchantParams{
		Email:  user.Email,
		ApiKey: apiKey.String(),
	}

	// Mock expected AffiliateMerchant response
	expectedMerchant := AffiliateMerchant{
		Email:         testParams.Email,
		ApiKey:        testParams.ApiKey,
		WebhookUrl:    "",
		WebhookSecret: "",
	}

	// Call the function being tested
	createdMerchant, err := testQueries.CreateAffiliateMerchant(context.Background(), testParams)

	// Check for errors
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Compare the returned AffiliateMerchant with the expected one
	assert.Equal(t, expectedMerchant.ApiKey, createdMerchant.ApiKey)
	assert.Equal(t, expectedMerchant.Email, createdMerchant.Email)
	assert.Equal(t, expectedMerchant.WebhookUrl, createdMerchant.WebhookUrl)
	assert.Equal(t, expectedMerchant.WebhookSecret, createdMerchant.WebhookSecret)

}

func TestQueries_GetAffiliateMerchantByEmail(t *testing.T) {
	user := createRandomUser(t, "Business")
	apiKey := common.RandomAPIKey()

	arg := CreateAffiliateMerchantParams{
		Email:  user.Email,
		ApiKey: apiKey.String(),
	}

	merchant, err := testQueries.CreateAffiliateMerchant(context.Background(), arg)
	assert.NoError(t, err)
	assert.NotNil(t, merchant)

	// Mock expected AffiliateMerchant response
	expectedMerchant := AffiliateMerchant{
		Email:         merchant.Email,
		ApiKey:        merchant.ApiKey,
		WebhookUrl:    merchant.WebhookUrl,
		WebhookSecret: merchant.WebhookSecret,
	}
	// Call the function being tested
	merchantData, err := testQueries.GetAffiliateMerchantByEmail(context.Background(), user.Email)

	// Check for errors
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Compare the returned AffiliateMerchant with the expected one
	assert.Equal(t, expectedMerchant.ApiKey, merchantData.ApiKey)
	assert.Equal(t, expectedMerchant.Email, merchantData.Email)
	assert.Equal(t, expectedMerchant.WebhookUrl, merchantData.WebhookUrl)
	assert.Equal(t, expectedMerchant.WebhookSecret, merchantData.WebhookSecret)
}

func TestQueries_GetAffiliateMerchantByAPIKey(t *testing.T) {
	user := createRandomUser(t, "Business")
	apiKey := common.RandomAPIKey()

	arg := CreateAffiliateMerchantParams{
		Email:  user.Email,
		ApiKey: apiKey.String(),
	}

	merchant, err := testQueries.CreateAffiliateMerchant(context.Background(), arg)
	assert.NoError(t, err)
	assert.NotNil(t, merchant)

	// Mock expected AffiliateMerchant response
	expectedMerchant := AffiliateMerchant{
		Email:         merchant.Email,
		ApiKey:        merchant.ApiKey,
		WebhookUrl:    merchant.WebhookUrl,
		WebhookSecret: merchant.WebhookSecret,
	}

	// Call the function being tested
	merchantData, err := testQueries.GetAffiliateMerchantByAPIKey(context.Background(), apiKey.String())

	// Check for errors
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Compare the returned AffiliateMerchant with the expected one
	assert.Equal(t, expectedMerchant.ApiKey, merchantData.ApiKey)
	assert.Equal(t, expectedMerchant.Email, merchantData.Email)
	assert.Equal(t, expectedMerchant.WebhookUrl, merchantData.WebhookUrl)
	assert.Equal(t, expectedMerchant.WebhookSecret, merchantData.WebhookSecret)
}

func TestQueries_GetAffiliateMerchantByID(t *testing.T) {
	user := createRandomUser(t, "Business")
	apiKey := common.RandomAPIKey()

	arg := CreateAffiliateMerchantParams{
		Email:  user.Email,
		ApiKey: apiKey.String(),
	}

	merchant, err := testQueries.CreateAffiliateMerchant(context.Background(), arg)
	assert.NoError(t, err)
	assert.NotNil(t, merchant)

	// Mock expected AffiliateMerchant response
	expectedMerchant := AffiliateMerchant{
		Email:         merchant.Email,
		ApiKey:        merchant.ApiKey,
		WebhookUrl:    merchant.WebhookUrl,
		WebhookSecret: merchant.WebhookSecret,
	}

	// Call the function being tested
	merchantData, err := testQueries.GetAffiliateMerchantByID(context.Background(), merchant.ID)

	// Check for errors
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Compare the returned AffiliateMerchant with the expected one
	assert.Equal(t, expectedMerchant.ApiKey, merchantData.ApiKey)
	assert.Equal(t, expectedMerchant.Email, merchantData.Email)
	assert.Equal(t, expectedMerchant.WebhookUrl, merchantData.WebhookUrl)
	assert.Equal(t, expectedMerchant.WebhookSecret, merchantData.WebhookSecret)
}

func TestQueries_UpdateAffiliateMerchant(t *testing.T) {
	user := createRandomUser(t, "Business")
	apiKey := common.RandomAPIKey()
	arg := CreateAffiliateMerchantParams{
		Email:  user.Email,
		ApiKey: apiKey.String(),
	}
	merchant, err := testQueries.CreateAffiliateMerchant(context.Background(), arg)
	assert.NoError(t, err)
	assert.NotNil(t, merchant)
	// Define test parameters
	testParams := UpdateAffiliateMerchantAPIKeyParams{
		Email:  merchant.Email,
		ApiKey: common.RandomAPIKey().String(),
	}

	// Mock expected AffiliateMerchant response
	expectedMerchant := AffiliateMerchant{
		Email:         merchant.Email,
		ApiKey:        testParams.ApiKey,
		WebhookUrl:    merchant.WebhookUrl,
		WebhookSecret: merchant.WebhookSecret,
	}

	// Call the function being tested
	updatedMerchant, err := testQueries.UpdateAffiliateMerchantAPIKey(context.Background(), testParams)

	// Check for errors
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Compare the returned AffiliateMerchant with the expected one
	assert.Equal(t, expectedMerchant.ApiKey, updatedMerchant.ApiKey)
	assert.Equal(t, expectedMerchant.Email, updatedMerchant.Email)
	assert.Equal(t, expectedMerchant.WebhookUrl, updatedMerchant.WebhookUrl)
	assert.Equal(t, expectedMerchant.WebhookSecret, updatedMerchant.WebhookSecret)
}

func TestQueries_UpdateAffiliateMerchantWebhookData(t *testing.T) {
	user := createRandomUser(t, "Business")
	webhookURL := common.RandomString(25)
	webhookSecret := common.RandomString(25)
	apiKey := common.RandomAPIKey()
	arg := CreateAffiliateMerchantParams{
		Email:  user.Email,
		ApiKey: apiKey.String(),
	}
	merchant, err := testQueries.CreateAffiliateMerchant(context.Background(), arg)
	assert.NoError(t, err)
	assert.NotNil(t, merchant)

	// Define test parameters
	testParams := UpdateAffiliateMerchantWebhookDataParams{
		Email:         merchant.Email,
		WebhookUrl:    webhookURL,
		WebhookSecret: webhookSecret,
	}

	// Mock expected AffiliateMerchant response
	expectedMerchant := AffiliateMerchant{
		ID:            merchant.ID,
		Email:         testParams.Email,
		ApiKey:        merchant.ApiKey,
		WebhookUrl:    testParams.WebhookUrl,
		WebhookSecret: testParams.WebhookSecret,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Call the function being tested
	updatedMerchant, err := testQueries.UpdateAffiliateMerchantWebhookData(context.Background(), testParams)

	// Check for errors
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Compare the returned AffiliateMerchant with the expected one
	assert.Equal(t, expectedMerchant.ApiKey, updatedMerchant.ApiKey)
	assert.Equal(t, expectedMerchant.Email, updatedMerchant.Email)
	assert.Equal(t, expectedMerchant.WebhookUrl, updatedMerchant.WebhookUrl)
	assert.Equal(t, expectedMerchant.WebhookSecret, updatedMerchant.WebhookSecret)
}
