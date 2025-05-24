package db

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/timchuks/monieverse/internal/common"
)

// TestCreateAffiliateMerchantWebhookMetaData tests CreateAffiliateMerchantWebhookMetaData function
func TestCreateAffiliateMerchantWebhookMetaData(t *testing.T) {

	id := common.RandomAPIKey()
	newStruct := struct {
		amount   string
		currency string
	}{
		amount:   "merchant",
		currency: "NGN",
	}
	swap, _ := common.JSONEncode(newStruct)
	// Define test parameters
	testParams := CreateAffiliateMerchantWebhookMetaDataParams{
		MerchantID:    id,
		SwapData:      swap,
		MetaData:      swap,
		WebhookStatus: 0,
	}

	// Mock expected AffiliateMerchantWebhookTmp response
	expectedResult := AffiliateMerchantWebhookTmp{
		MerchantID:    testParams.MerchantID,
		SwapData:      testParams.SwapData,
		MetaData:      testParams.MetaData,
		WebhookStatus: testParams.WebhookStatus,
	}

	// Call the function being tested
	result, err := testQueries.CreateAffiliateMerchantWebhookMetaData(context.Background(), testParams)

	// Check for errors
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Compare the returned AffiliateMerchantWebhookTmp with the expected one
	assert.Equal(t, expectedResult.MerchantID, result.MerchantID)
	assert.Equal(t, expectedResult.MetaData, result.MetaData)
	assert.Equal(t, expectedResult.SwapData, result.SwapData)
	assert.Equal(t, expectedResult.WebhookStatus, result.WebhookStatus)
}

// TestUpdateAffiliateMerchantWebhookStatus tests UpdateAffiliateMerchantWebhookStatus function
func TestUpdateAffiliateMerchantWebhookStatus(t *testing.T) {

	id, _ := uuid.NewUUID()
	newStruct := struct {
		amount   string
		currency string
	}{
		amount:   "merchant",
		currency: "NGN",
	}
	swap, _ := common.JSONEncode(newStruct)
	// Define test parameters for valid input
	validParams := CreateAffiliateMerchantWebhookMetaDataParams{
		MerchantID:    id,
		SwapData:      swap,
		MetaData:      swap,
		WebhookStatus: 0,
	}

	// Call the function being tested with valid parameters
	validData, validErr := testQueries.CreateAffiliateMerchantWebhookMetaData(context.Background(), validParams)
	assert.NoError(t, validErr)
	assert.NotNil(t, validData)
	// Define test parameters
	testParams := UpdateAffiliateMerchantWebhookStatusParams{
		WebhookStatus: 1,
		ID:            validData.ID,
	}

	// Mock expected AffiliateMerchant response
	expectedMerchant := AffiliateMerchantWebhookTmp{
		WebhookStatus: testParams.WebhookStatus,
		ID:            testParams.ID,
		MerchantID:    validData.MerchantID,
		SwapData:      validData.SwapData,
		MetaData:      validData.MetaData,
		CreatedAt:     validData.CreatedAt,
		UpdatedAt:     validData.UpdatedAt,
	}

	// Call the function being tested
	updatedMerchant, err := testQueries.UpdateAffiliateMerchantWebhookStatus(context.Background(), testParams)

	// Check for errors
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Compare the returned AffiliateMerchant with the expected one
	assert.NotNil(t, updatedMerchant)
	assert.Equal(t, expectedMerchant, updatedMerchant)
}

func TestGetAffiliateMerchantWebhookMetaDatasByWebhookStatus(t *testing.T) {
	// Define a test case
	testCases := []struct {
		name   string
		status int32
		expect error // Expected error (nil if no error is expected)
	}{
		{
			name:   "Valid status",
			status: 0,
			expect: nil, // No error expected
		},
		{
			name:   "Invalid status",
			status: 4,
			expect: fmt.Errorf("sql: no rows in result set"), // Error expected
		},
	}

	// Iterate through test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Call the function being tested
			_, err := testQueries.GetAffiliateMerchantWebhookMetaDatasByWebhookStatus(context.Background(), tc.status)

			// Check if the error matches the expected error
			if err != nil {
				assert.Equal(t, tc.expect, err)
			}
		})
	}
}
