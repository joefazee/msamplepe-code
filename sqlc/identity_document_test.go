package db

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIdentityDocumentQueries(t *testing.T) {
	user := createRandomUser(t, "Personal")

	arg := CreateIdentityDocumentParams{
		UserID:         user.ID,
		DocumentType:   "pdf",
		DocumentNumber: "2",
		DocumentPath:   "./document-path",
		Bucket:         "aws-bucket",
		Storage:        "test-storage",
	}
	userIdentityDocument, err := testQueries.CreateIdentityDocument(context.Background(), arg)
	assert.NoError(t, err)
	assert.NotEmpty(t, userIdentityDocument)

	//Test GetUserIdentityDocument
	getUserIdentityDocument, err := testQueries.GetUserIdentityDocument(context.Background(), arg.UserID)
	assert.NoError(t, err)
	assert.NotEmpty(t, getUserIdentityDocument)

	// Test DeleteIdentityDocument
	err = testQueries.DeleteIdentityDocument(context.Background(), userIdentityDocument.ID)
	assert.NoError(t, err)

}
