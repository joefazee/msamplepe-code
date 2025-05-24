package db

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

func (q *Queries) CreateKYCRequirementResultEntry(ctx context.Context, kycRequirementID uuid.UUID, values map[string]string) error {

	if len(values) == 0 {
		return nil
	}
	var valueStrings []string
	var valueArgs []interface{}

	i := 0
	for field, value := range values {
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d)", i+1, i+2, i+3))
		valueArgs = append(valueArgs, kycRequirementID, field, value)
		i += 3
	}

	stmt := fmt.Sprintf("INSERT INTO dynamic_kyc_results (kyc_requirement_id, field, value) VALUES %s", strings.Join(valueStrings, ","))

	_, err := q.db.ExecContext(ctx, stmt, valueArgs...)

	return err
}
