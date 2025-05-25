package forms

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	db "github.com/timchuks/monieverse/internal/db/sqlc"
)

func CreateKYBForm(ctx context.Context, store db.Store, creatorID uuid.UUID) error {
	input := &db.FormDefinitionInput{
		Name:             "Business Verification (KYB)",
		Slug:             "kyb-verification",
		Description:      "Complete business verification process",
		FormType:         "kyb",
		IsMultiStep:      true,
		RequiresApproval: true,
		CreatedBy:        creatorID,
		Steps: []db.StepInput{
			{
				StepNumber:  1,
				Name:        "Business Information",
				Description: "Basic business details and structure",
				IsOptional:  false,
			},
			{
				StepNumber:  2,
				Name:        "Business Documents",
				Description: "Upload incorporation and licensing documents",
				IsOptional:  false,
			},
			{
				StepNumber:  3,
				Name:        "Ownership Structure",
				Description: "Details about business owners and beneficiaries",
				IsOptional:  false,
			},
			{
				StepNumber:  4,
				Name:        "Financial Information",
				Description: "Banking and financial details",
				IsOptional:  false,
			},
			{
				StepNumber:  5,
				Name:        "Compliance",
				Description: "Regulatory and compliance information",
				IsOptional:  false,
			},
		},
		Fields: []db.FieldInput{
			// Step 1: Business Information
			{
				FieldName:    "legal_business_name",
				FieldType:    "text",
				StepNumber:   1,
				Label:        map[string]string{"en": "Legal Business Name"},
				DisplayOrder: 1,
				IsRequired:   true,
				ValidationRules: map[string]interface{}{
					"required":   true,
					"min_length": 3,
					"max_length": 200,
				},
			},
			{
				FieldName:    "trading_name",
				FieldType:    "text",
				StepNumber:   1,
				Label:        map[string]string{"en": "Trading Name (DBA)"},
				HelpText:     map[string]string{"en": "Leave blank if same as legal name"},
				DisplayOrder: 2,
				IsRequired:   false,
			},
			{
				FieldName:    "business_type",
				FieldType:    "select",
				StepNumber:   1,
				Label:        map[string]string{"en": "Business Type"},
				DisplayOrder: 3,
				IsRequired:   true,
				Options: map[string]interface{}{
					"type": "static",
					"static": []map[string]interface{}{
						{"value": "corporation", "label": map[string]string{"en": "Corporation"}},
						{"value": "llc", "label": map[string]string{"en": "Limited Liability Company (LLC)"}},
						{"value": "partnership", "label": map[string]string{"en": "Partnership"}},
						{"value": "sole_proprietorship", "label": map[string]string{"en": "Sole Proprietorship"}},
						{"value": "non_profit", "label": map[string]string{"en": "Non-Profit Organization"}},
						{"value": "trust", "label": map[string]string{"en": "Trust"}},
					},
				},
			},
			{
				FieldName:    "incorporation_country",
				FieldType:    "select",
				StepNumber:   1,
				Label:        map[string]string{"en": "Country of Incorporation"},
				DisplayOrder: 4,
				IsRequired:   true,
				Options: map[string]interface{}{
					"type": "dynamic",
					"dynamic": map[string]interface{}{
						"source_name": "countries",
					},
				},
			},
			{
				FieldName:    "incorporation_state",
				FieldType:    "select",
				StepNumber:   1,
				Label:        map[string]string{"en": "State/Province of Incorporation"},
				DisplayOrder: 5,
				IsRequired:   true,
				Options: map[string]interface{}{
					"type": "dynamic",
					"dynamic": map[string]interface{}{
						"source_name": "states",
						"filter_params": map[string]string{
							"country": "{{incorporation_country}}",
						},
					},
				},
				ConditionalLogic: map[string]interface{}{
					"action": "show",
					"conditions": []map[string]interface{}{
						{
							"field":    "incorporation_country",
							"operator": "in",
							"value":    []string{"US", "CA"},
						},
					},
					"logic": "all",
				},
			},
			{
				FieldName:    "registration_number",
				FieldType:    "text",
				StepNumber:   1,
				Label:        map[string]string{"en": "Business Registration Number"},
				HelpText:     map[string]string{"en": "EIN for US, Business Number for Canada, etc."},
				DisplayOrder: 6,
				IsRequired:   true,
				ValidationRules: map[string]interface{}{
					"required": true,
					"custom_rules": map[string]interface{}{
						"business_number": map[string]interface{}{
							"country": "{{incorporation_country}}",
						},
					},
				},
			},
			{
				FieldName:    "incorporation_date",
				FieldType:    "date",
				StepNumber:   1,
				Label:        map[string]string{"en": "Date of Incorporation"},
				DisplayOrder: 7,
				IsRequired:   true,
			},
			{
				FieldName:    "business_industry",
				FieldType:    "select",
				StepNumber:   1,
				Label:        map[string]string{"en": "Primary Industry"},
				DisplayOrder: 8,
				IsRequired:   true,
				Options: map[string]interface{}{
					"type": "static",
					"static": []map[string]interface{}{
						{"value": "technology", "label": map[string]string{"en": "Technology"}},
						{"value": "finance", "label": map[string]string{"en": "Finance & Banking"}},
						{"value": "retail", "label": map[string]string{"en": "Retail & E-commerce"}},
						{"value": "manufacturing", "label": map[string]string{"en": "Manufacturing"}},
						{"value": "healthcare", "label": map[string]string{"en": "Healthcare"}},
						{"value": "real_estate", "label": map[string]string{"en": "Real Estate"}},
						{"value": "professional_services", "label": map[string]string{"en": "Professional Services"}},
						{"value": "other", "label": map[string]string{"en": "Other"}},
					},
				},
			},
			{
				FieldName:    "business_description",
				FieldType:    "textarea",
				StepNumber:   1,
				Label:        map[string]string{"en": "Business Description"},
				Placeholder:  map[string]string{"en": "Describe your business activities..."},
				DisplayOrder: 9,
				IsRequired:   true,
				ValidationRules: map[string]interface{}{
					"required":   true,
					"min_length": 50,
					"max_length": 1000,
				},
			},

			// Step 2: Business Documents
			{
				FieldName:    "certificate_of_incorporation",
				FieldType:    "file",
				StepNumber:   2,
				Label:        map[string]string{"en": "Certificate of Incorporation"},
				DisplayOrder: 10,
				IsRequired:   true,
				FileConfig: map[string]interface{}{
					"max_size":      20971520, // 20MB
					"allowed_types": []string{"application/pdf", "image/jpeg", "image/png"},
				},
			},
			{
				FieldName:    "articles_of_association",
				FieldType:    "file",
				StepNumber:   2,
				Label:        map[string]string{"en": "Articles of Association / Operating Agreement"},
				DisplayOrder: 11,
				IsRequired:   true,
				FileConfig: map[string]interface{}{
					"max_size":      20971520,
					"allowed_types": []string{"application/pdf"},
				},
			},
			{
				FieldName:    "business_licenses",
				FieldType:    "files",
				StepNumber:   2,
				Label:        map[string]string{"en": "Business Licenses"},
				HelpText:     map[string]string{"en": "Upload all relevant business licenses"},
				DisplayOrder: 12,
				IsRequired:   false,
				FileConfig: map[string]interface{}{
					"max_size":      10485760,
					"allowed_types": []string{"application/pdf", "image/jpeg", "image/png"},
					"max_files":     5,
				},
			},
			{
				FieldName:    "bank_statements",
				FieldType:    "files",
				StepNumber:   2,
				Label:        map[string]string{"en": "Bank Statements (Last 3 months)"},
				DisplayOrder: 13,
				IsRequired:   true,
				FileConfig: map[string]interface{}{
					"max_size":      10485760,
					"allowed_types": []string{"application/pdf"},
					"max_files":     3,
				},
			},

			// Step 3: Ownership Structure (Dynamic section)
			{
				FieldName:    "number_of_owners",
				FieldType:    "number",
				StepNumber:   3,
				Label:        map[string]string{"en": "Number of Owners/Shareholders"},
				DisplayOrder: 14,
				IsRequired:   true,
				ValidationRules: map[string]interface{}{
					"required": true,
					"min":      1,
					"max":      20,
				},
			},
			// Note: In real implementation, you'd dynamically generate owner fields based on number_of_owners
			{
				FieldName:    "owner_1_name",
				FieldType:    "text",
				StepNumber:   3,
				Label:        map[string]string{"en": "Owner 1 - Full Name"},
				DisplayOrder: 15,
				IsRequired:   true,
			},
			{
				FieldName:    "owner_1_ownership_percentage",
				FieldType:    "number",
				StepNumber:   3,
				Label:        map[string]string{"en": "Owner 1 - Ownership Percentage"},
				DisplayOrder: 16,
				IsRequired:   true,
				ValidationRules: map[string]interface{}{
					"required": true,
					"min":      0,
					"max":      100,
				},
			},
			{
				FieldName:    "owner_1_is_ubo",
				FieldType:    "checkbox",
				StepNumber:   3,
				Label:        map[string]string{"en": "Owner 1 - Ultimate Beneficial Owner (25%+ ownership)"},
				DisplayOrder: 17,
				IsRequired:   false,
			},
			{
				FieldName:    "owner_1_id_document",
				FieldType:    "file",
				StepNumber:   3,
				Label:        map[string]string{"en": "Owner 1 - ID Document"},
				DisplayOrder: 18,
				IsRequired:   true,
				FileConfig: map[string]interface{}{
					"max_size":      10485760,
					"allowed_types": []string{"application/pdf", "image/jpeg", "image/png"},
				},
			},

			// Step 4: Financial Information
			{
				FieldName:    "primary_bank",
				FieldType:    "select",
				StepNumber:   4,
				Label:        map[string]string{"en": "Primary Bank"},
				DisplayOrder: 19,
				IsRequired:   true,
				Options: map[string]interface{}{
					"type": "dynamic",
					"dynamic": map[string]interface{}{
						"source_name": "banks",
						"filter_params": map[string]string{
							"country": "{{incorporation_country}}",
						},
					},
				},
			},
			{
				FieldName:    "bank_account_number",
				FieldType:    "text",
				StepNumber:   4,
				Label:        map[string]string{"en": "Bank Account Number"},
				DisplayOrder: 20,
				IsRequired:   true,
				ValidationRules: map[string]interface{}{
					"required": true,
					"custom_rules": map[string]interface{}{
						"bank_account": map[string]interface{}{
							"bank": "{{primary_bank}}",
						},
					},
				},
			},
			{
				FieldName:    "annual_revenue",
				FieldType:    "currency",
				StepNumber:   4,
				Label:        map[string]string{"en": "Annual Revenue (Last Year)"},
				DisplayOrder: 21,
				IsRequired:   true,
				ValidationRules: map[string]interface{}{
					"required": true,
					"min":      0,
					"max":      1000000000,
				},
			},
			{
				FieldName:    "expected_monthly_volume",
				FieldType:    "currency",
				StepNumber:   4,
				Label:        map[string]string{"en": "Expected Monthly Transaction Volume"},
				DisplayOrder: 22,
				IsRequired:   true,
			},
			{
				FieldName:    "source_of_funds",
				FieldType:    "checkbox",
				StepNumber:   4,
				Label:        map[string]string{"en": "Primary Sources of Funds"},
				DisplayOrder: 23,
				IsRequired:   true,
				Options: map[string]interface{}{
					"type": "static",
					"static": []map[string]interface{}{
						{"value": "sales_revenue", "label": map[string]string{"en": "Sales Revenue"}},
						{"value": "investments", "label": map[string]string{"en": "Investments"}},
						{"value": "loans", "label": map[string]string{"en": "Business Loans"}},
						{"value": "grants", "label": map[string]string{"en": "Grants"}},
						{"value": "other", "label": map[string]string{"en": "Other"}},
					},
				},
			},

			// Step 5: Compliance
			{
				FieldName:    "has_compliance_officer",
				FieldType:    "radio",
				StepNumber:   5,
				Label:        map[string]string{"en": "Do you have a designated compliance officer?"},
				DisplayOrder: 24,
				IsRequired:   true,
				Options: map[string]interface{}{
					"type": "static",
					"static": []map[string]interface{}{
						{"value": "yes", "label": map[string]string{"en": "Yes"}},
						{"value": "no", "label": map[string]string{"en": "No"}},
					},
				},
			},
			{
				FieldName:    "compliance_officer_name",
				FieldType:    "text",
				StepNumber:   5,
				Label:        map[string]string{"en": "Compliance Officer Name"},
				DisplayOrder: 25,
				IsRequired:   true,
				ConditionalLogic: map[string]interface{}{
					"action": "show",
					"conditions": []map[string]interface{}{
						{
							"field":    "has_compliance_officer",
							"operator": "equals",
							"value":    "yes",
						},
					},
					"logic": "all",
				},
			},
			{
				FieldName:    "aml_policy",
				FieldType:    "file",
				StepNumber:   5,
				Label:        map[string]string{"en": "Anti-Money Laundering (AML) Policy"},
				HelpText:     map[string]string{"en": "Upload your AML policy document if available"},
				DisplayOrder: 26,
				IsRequired:   false,
				FileConfig: map[string]interface{}{
					"max_size":      20971520,
					"allowed_types": []string{"application/pdf", "application/msword", "application/vnd.openxmlformats-officedocument.wordprocessingml.document"},
				},
			},
			{
				FieldName:    "certifications",
				FieldType:    "checkbox",
				StepNumber:   5,
				Label:        map[string]string{"en": "Certifications and Declarations"},
				DisplayOrder: 27,
				IsRequired:   true,
				Options: map[string]interface{}{
					"type": "static",
					"static": []map[string]interface{}{
						{
							"value": "info_accurate",
							"label": map[string]string{"en": "I certify that all information provided is accurate and complete"},
						},
						{
							"value": "authorized",
							"label": map[string]string{"en": "I am authorized to provide this information on behalf of the business"},
						},
						{
							"value": "agree_verification",
							"label": map[string]string{"en": "I agree to cooperate with any verification requests"},
						},
						{
							"value": "agree_terms",
							"label": map[string]string{"en": "I agree to the Terms of Service and Privacy Policy"},
						},
					},
				},
				ValidationRules: map[string]interface{}{
					"required":     true,
					"all_required": true, // Custom rule to ensure all checkboxes are checked
				},
			},
		},
		PersistenceConfig: &db.PersistenceConfigInput{
			PersistenceMode: "multi_table",
			TargetConfigs: []map[string]interface{}{
				{"table_name": "businesses", "priority": 1},
				{"table_name": "business_owners", "priority": 2},
				{"table_name": "documents", "priority": 3},
				{"table_name": "form_submissions", "priority": 4}, // Also keep full submission
			},
			FieldMappings: map[string]interface{}{
				"legal_business_name": map[string]interface{}{
					"form_field":  "legal_business_name",
					"table_name":  "businesses",
					"column_name": "name",
					"data_type":   "varchar",
				},
				"registration_number": map[string]interface{}{
					"form_field":  "registration_number",
					"table_name":  "businesses",
					"column_name": "registration_number",
					"data_type":   "varchar",
				},
				"business_type": map[string]interface{}{
					"form_field":  "business_type",
					"table_name":  "businesses",
					"column_name": "business_category",
					"data_type":   "varchar",
				},
				"business_description": map[string]interface{}{
					"form_field":  "business_description",
					"table_name":  "businesses",
					"column_name": "product_description",
					"data_type":   "text",
				},
				"incorporation_date": map[string]interface{}{
					"form_field":  "incorporation_date",
					"table_name":  "businesses",
					"column_name": "registration_date",
					"data_type":   "varchar",
					"transform":   "date_to_string",
				},
				// Owner mappings would be dynamic based on number of owners
			},
			ValidationHooks: []map[string]interface{}{
				{
					"type": "webhook",
					"config": map[string]interface{}{
						"url":    "https://compliance-api.example.com/verify-business",
						"method": "POST",
					},
				},
			},
		},
	}

	form, err := store.CreateFormDefinitionTx(ctx, input)
	if err != nil {
		return err
	}

	// Create assignments for business accounts
	_, err = store.CreateFormAssignment(ctx, db.CreateFormAssignmentParams{
		ID:               uuid.New(),
		FormDefinitionID: form.ID,
		AssignmentType:   "user_type",
		AssignmentValue:  "business",
		Priority:         10,
		CreatedBy:        db.NewNullUUID(creatorID),
	})

	if err != nil {
		return fmt.Errorf("create form assignment: %w", err)
	}

	// Special assignment for Delaware businesses
	delawareConditions, _ := json.Marshal(map[string]interface{}{
		"user_type": "business",
		"state":     "DE",
	})

	_, err = store.CreateFormAssignment(ctx, db.CreateFormAssignmentParams{
		ID:               uuid.New(),
		FormDefinitionID: form.ID,
		AssignmentType:   "state",
		AssignmentValue:  "DE",
		Conditions:       db.NullRawMessage(delawareConditions),
		Priority:         20,
		CreatedBy:        db.NewNullUUID(creatorID),
	})

	return err
}
