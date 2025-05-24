package db

import (
	"context"
)

const moneySendrGetPaginatedTransactions = `
SELECT count(*) OVER() AS total_records, id, email, phone, bank_name, bank_code, account_number, account_name, created_at, updated_at, remark, status, exchangeratedata, fromcurrencycode, tocurrencycode, amount_in, amount_out, fees, rate, internal_remark FROM moneysendr_recipients  ORDER BY created_at DESC,status desc LIMIT $1 OFFSET $2
`

func (store *SQLStore) MoneySendrGetPaginatedTransactions(ctx context.Context, filter Filter) ([]MoneysendrRecipient, Metadata, error) {
	rows, err := store.db.QueryContext(ctx, moneySendrGetPaginatedTransactions, filter.Limit(), filter.Offset())
	if err != nil {
		return nil, EmptyMetadata, err
	}
	defer rows.Close()
	items := []MoneysendrRecipient{}
	var totalRecords int
	for rows.Next() {
		var i MoneysendrRecipient
		if err := rows.Scan(
			&totalRecords,
			&i.ID,
			&i.Email,
			&i.Phone,
			&i.BankName,
			&i.BankCode,
			&i.AccountNumber,
			&i.AccountName,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.Remark,
			&i.Status,
			&i.Exchangeratedata,
			&i.Fromcurrencycode,
			&i.Tocurrencycode,
			&i.AmountIn,
			&i.AmountOut,
			&i.Fees,
			&i.Rate,
			&i.InternalRemark,
		); err != nil {
			return nil, EmptyMetadata, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, EmptyMetadata, err
	}
	if err := rows.Err(); err != nil {
		return nil, EmptyMetadata, err
	}

	metadata := CalculateMetadata(totalRecords, filter.Page, filter.Limit())
	return items, metadata, nil
}
