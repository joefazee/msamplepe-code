package schemes

import (
	"fmt"

	"github.com/docker/distribution/context"
	db "github.com/timchuks/monieverse/internal/db/sqlc"
	"github.com/timchuks/monieverse/internal/domain"
	"github.com/timchuks/monieverse/internal/fields"
)

const (
	AchLabel  = "ACH"
	WireLabel = "WIRE"

	IbanLabel     = "IBAN"
	SortCodeLabel = "SORTCODE"

	CadLocal = "CAD_LOCAL"

	NGNLabel = "NGN"

	CNYLabelUnionPay = "CNY_UNIONPAY"
	CNYLabelAliPay   = "CNY_ALIPAY"

	SwiftLabel = "SWIFT"

	fieldScopePersonal = "p"
	fieldScopeBusiness = "b"

	fieldScopeFull = "*"

	checkingAccountType = "checking"

	savingAccountType = "saving"
)

var (
	businessName     = fields.NewTextField("business_name", "Business name / organization", fieldScopeBusiness)
	achRoutingNumber = fields.NewTextField("ach_routing_number", "ACH Routing Number", fieldScopeFull)

	accountNumber = fields.NewTextField("account_number", "Account number", fieldScopeBusiness)

	accountType = fields.NewDropdownField("account_type", "Account type", fieldScopeFull, fmt.Sprintf(`["%s","%s"]`, checkingAccountType, savingAccountType))

	swiftCode = fields.NewTextField("swift_code", "SWIFT Code or BIC", fieldScopeFull)

	iban = fields.NewTextField("iban", "IBAN", fieldScopeFull)

	accountHolderFullname = fields.NewTextField("account_holder_fullname", "Full name of the account holder", fieldScopeFull)

	countries = fields.NewDropdownField("countries", "Country", fieldScopeFull, domain.Countries.JSON())

	recipientAddress = fields.NewTextField("recipient_address", "Recipient address", fieldScopeFull)

	postCode = fields.NewTextField("post_code", "Post code", fieldScopeFull)

	city = fields.NewTextField("city", "City", fieldScopeFull)

	bankCity = fields.NewTextField("bank_city", "Bank City", fieldScopeFull)

	bankAddress = fields.NewTextField("bank_address", "Bank Address", fieldScopeFull)

	state = fields.NewTextField("state", "State", fieldScopeFull)

	wireRoutingNumber = fields.NewTextField("wire_routing_number", "Fedwire routing number", fieldScopeBusiness)

	ukSortCode = fields.NewTextField("uk_sort_code", "UK Sort Code", fieldScopeFull)

	institutionNumber = fields.NewTextField("institution_number", "Institution Number", fieldScopeFull)

	transitNumber = fields.NewTextField("transit_number", "Transit Number", fieldScopeFull)

	cardNumber = fields.NewTextField("card_number", "Card Number", fieldScopeFull)

	aliPayID = fields.NewTextField("ali_pay_id", "AliPay ID", fieldScopeFull)

	bankName = fields.NewTextField("bank_name", "Bank Name", fieldScopeFull)

	IBAN = []fields.Field{
		accountHolderFullname,
		iban,

		countries,
		city,
		recipientAddress,
		postCode,
	}

	ACH = []fields.Field{
		accountHolderFullname,
		businessName,
		achRoutingNumber,
		accountNumber,
		accountType,
		bankName,
		bankCity,
		bankAddress,

		countries,
		city,
		state,
		recipientAddress,
		postCode,
	}

	SWIFT = []fields.Field{
		businessName,
		swiftCode,
		accountNumber,
		bankName,
		bankCity,

		countries,
		city,
		state,
		recipientAddress,
		postCode,
	}

	WIRE = []fields.Field{
		businessName,
		wireRoutingNumber,
		accountNumber,
		accountType,
		countries,
		city,
		recipientAddress,
		postCode,
	}

	SORTCODE = []fields.Field{
		businessName,
		ukSortCode,
		accountNumber,

		countries,
		city,
		recipientAddress,
		postCode,
	}

	CADLOCAL = []fields.Field{
		businessName,
		institutionNumber,
		transitNumber,
		accountNumber,
		accountType,

		countries,
		city,
		recipientAddress,
		postCode,
	}

	CNYUNIONPAY = []fields.Field{
		accountHolderFullname,
		cardNumber,

		countries,
		city,
		recipientAddress,
		postCode,
	}

	CNYALIPAY = []fields.Field{
		accountHolderFullname,
		aliPayID,

		countries,
		city,
		recipientAddress,
		postCode,
	}

	paymentSchemes = map[string][]fields.Field{
		SwiftLabel:       SWIFT,
		AchLabel:         ACH,
		IbanLabel:        IBAN,
		SortCodeLabel:    SORTCODE,
		CadLocal:         CADLOCAL,
		CNYLabelUnionPay: CNYUNIONPAY,
		CNYLabelAliPay:   CNYALIPAY,
	}
)

func GetNGNFields(store db.Store) ([]fields.Field, error) {
	var ngnFields = []fields.Field{
		accountHolderFullname,
		accountNumber,
	}

	banks, err := store.GetAllBanks(context.Background())
	if err != nil {
		return nil, err
	}
	bankCodes := make(map[string]string)
	for _, bank := range banks {
		bankCodes[bank.Code] = bank.Name
	}

	ngnFields = append(ngnFields, fields.NewDropdownFieldWithItems("nigeria_bank", "Nigeria Banks", fieldScopeFull, bankCodes))

	return ngnFields, nil
}

func GetSchemes(store db.Store) (map[string][]fields.Field, error) {
	ngnFields, err := GetNGNFields(store)
	if err != nil {
		return nil, err
	}
	paymentSchemes[NGNLabel] = ngnFields
	return paymentSchemes, nil
}
