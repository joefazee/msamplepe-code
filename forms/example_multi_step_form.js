// EXAMPLE OF MULTI STEP FORM RESPONSE

const response = `{
    "message": "Form retrieved successfully",
    "data": {
    "form_definition": {
        "id": "97879a9d-818f-4f74-a953-f2696d7ead9a",
            "name": "Business Verification (KYB)",
            "slug": "kyb-verification",
            "description": "Complete business verification process",
            "form_type": "kyb",
            "version": 1,
            "is_active": true,
            "is_multi_step": true,
            "requires_approval": true,
            "is_editable_after_submission": false,
            "approval_workflow": {},
        "created_at": "2025-05-25T12:58:29.350979+01:00",
            "updated_at": "2025-05-25T12:58:29.350979+01:00",
            "created_by": "37b3f0a3-4059-49c8-8fe0-9da3f3e02585"
    },
    "steps": [
        {
            "id": "b3c926a5-fef8-4b81-8630-b4bb2442f7dd",
            "form_definition_id": "97879a9d-818f-4f74-a953-f2696d7ead9a",
            "step_number": 1,
            "name": "Business Information",
            "description": "Basic business details and structure",
            "is_optional": false,
            "created_at": "2025-05-25T12:58:29.350979+01:00"
        },
        {
            "id": "fcfbe6ab-ba25-4a0a-acf3-5675fb791bae",
            "form_definition_id": "97879a9d-818f-4f74-a953-f2696d7ead9a",
            "step_number": 2,
            "name": "Business Documents",
            "description": "Upload incorporation and licensing documents",
            "is_optional": false,
            "created_at": "2025-05-25T12:58:29.350979+01:00"
        },
        {
            "id": "a610e5ae-aa65-4ee0-b5c9-9a03d0c5fa0f",
            "form_definition_id": "97879a9d-818f-4f74-a953-f2696d7ead9a",
            "step_number": 3,
            "name": "Ownership Structure",
            "description": "Details about business owners and beneficiaries",
            "is_optional": false,
            "created_at": "2025-05-25T12:58:29.350979+01:00"
        },
        {
            "id": "4e5c1a7e-eb35-47ae-b1b7-5592ac31b32a",
            "form_definition_id": "97879a9d-818f-4f74-a953-f2696d7ead9a",
            "step_number": 4,
            "name": "Financial Information",
            "description": "Banking and financial details",
            "is_optional": false,
            "created_at": "2025-05-25T12:58:29.350979+01:00"
        },
        {
            "id": "50be3b87-a4b6-4b60-ab5b-9d76424d55a7",
            "form_definition_id": "97879a9d-818f-4f74-a953-f2696d7ead9a",
            "step_number": 5,
            "name": "Compliance",
            "description": "Regulatory and compliance information",
            "is_optional": false,
            "created_at": "2025-05-25T12:58:29.350979+01:00"
        }
    ],
        "fields": [
        {
            "id": "6ba238f7-647e-44a7-bb63-a99c319e3fdd",
            "form_definition_id": "97879a9d-818f-4f74-a953-f2696d7ead9a",
            "form_step_id": "b3c926a5-fef8-4b81-8630-b4bb2442f7dd",
            "field_name": "legal_business_name",
            "field_type": "text",
            "label": {
                "en": "Legal Business Name"
            },
            "placeholder": {},
            "help_text": {},
            "validation_rules": {
                "required": true,
                "max_length": 200,
                "min_length": 3
            },
            "options": {},
            "display_order": 1,
            "is_required": true,
            "is_readonly": false,
            "default_value": "",
            "conditional_logic": {},
            "file_config": {},
            "created_at": "2025-05-25T12:58:29.350979+01:00",
            "updated_at": "2025-05-25T12:58:29.350979+01:00"
        },
        {
            "id": "5d117a13-ad77-44b9-ae2b-564837d70cfb",
            "form_definition_id": "97879a9d-818f-4f74-a953-f2696d7ead9a",
            "form_step_id": "b3c926a5-fef8-4b81-8630-b4bb2442f7dd",
            "field_name": "trading_name",
            "field_type": "text",
            "label": {
                "en": "Trading Name (DBA)"
            },
            "placeholder": {},
            "help_text": {
                "en": "Leave blank if same as legal name"
            },
            "validation_rules": {},
            "options": {},
            "display_order": 2,
            "is_required": false,
            "is_readonly": false,
            "default_value": "",
            "conditional_logic": {},
            "file_config": {},
            "created_at": "2025-05-25T12:58:29.350979+01:00",
            "updated_at": "2025-05-25T12:58:29.350979+01:00"
        },
        {
            "id": "ff780002-c905-4d21-813a-d874a4d13483",
            "form_definition_id": "97879a9d-818f-4f74-a953-f2696d7ead9a",
            "form_step_id": "b3c926a5-fef8-4b81-8630-b4bb2442f7dd",
            "field_name": "business_type",
            "field_type": "select",
            "label": {
                "en": "Business Type"
            },
            "placeholder": {},
            "help_text": {},
            "validation_rules": {},
            "options": {
                "type": "static",
                "static": [
                    {
                        "label": {
                            "en": "Corporation"
                        },
                        "value": "corporation"
                    },
                    {
                        "label": {
                            "en": "Limited Liability Company (LLC)"
                        },
                        "value": "llc"
                    },
                    {
                        "label": {
                            "en": "Partnership"
                        },
                        "value": "partnership"
                    },
                    {
                        "label": {
                            "en": "Sole Proprietorship"
                        },
                        "value": "sole_proprietorship"
                    },
                    {
                        "label": {
                            "en": "Non-Profit Organization"
                        },
                        "value": "non_profit"
                    },
                    {
                        "label": {
                            "en": "Trust"
                        },
                        "value": "trust"
                    }
                ]
            },
            "display_order": 3,
            "is_required": true,
            "is_readonly": false,
            "default_value": "",
            "conditional_logic": {},
            "file_config": {},
            "created_at": "2025-05-25T12:58:29.350979+01:00",
            "updated_at": "2025-05-25T12:58:29.350979+01:00"
        },
        {
            "id": "7e65bba3-cd04-4620-9498-dc1e7969fbb2",
            "form_definition_id": "97879a9d-818f-4f74-a953-f2696d7ead9a",
            "form_step_id": "b3c926a5-fef8-4b81-8630-b4bb2442f7dd",
            "field_name": "incorporation_country",
            "field_type": "select",
            "label": {
                "en": "Country of Incorporation"
            },
            "placeholder": {},
            "help_text": {},
            "validation_rules": {},
            "options": {
                "type": "dynamic",
                "dynamic": {
                    "source_name": "countries"
                }
            },
            "display_order": 4,
            "is_required": true,
            "is_readonly": false,
            "default_value": "",
            "conditional_logic": {},
            "file_config": {},
            "created_at": "2025-05-25T12:58:29.350979+01:00",
            "updated_at": "2025-05-25T12:58:29.350979+01:00"
        },
        {
            "id": "5c5d8762-8ca0-45af-aff7-8da7621dc6ff",
            "form_definition_id": "97879a9d-818f-4f74-a953-f2696d7ead9a",
            "form_step_id": "b3c926a5-fef8-4b81-8630-b4bb2442f7dd",
            "field_name": "incorporation_state",
            "field_type": "select",
            "label": {
                "en": "State/Province of Incorporation"
            },
            "placeholder": {},
            "help_text": {},
            "validation_rules": {},
            "options": {
                "type": "dynamic",
                "dynamic": {
                    "source_name": "states",
                    "filter_params": {
                        "country": "{{incorporation_country}}"
                    }
                }
            },
            "display_order": 5,
            "is_required": true,
            "is_readonly": false,
            "default_value": "",
            "conditional_logic": {
                "logic": "all",
                "action": "show",
                "conditions": [
                    {
                        "field": "incorporation_country",
                        "value": ["US", "CA"],
                        "operator": "in"
                    }
                ]
            },
            "file_config": {},
            "created_at": "2025-05-25T12:58:29.350979+01:00",
            "updated_at": "2025-05-25T12:58:29.350979+01:00"
        },
        {
            "id": "f20ecd54-8204-41c9-a868-c73ef17a8c36",
            "form_definition_id": "97879a9d-818f-4f74-a953-f2696d7ead9a",
            "form_step_id": "b3c926a5-fef8-4b81-8630-b4bb2442f7dd",
            "field_name": "registration_number",
            "field_type": "text",
            "label": {
                "en": "Business Registration Number"
            },
            "placeholder": {},
            "help_text": {
                "en": "EIN for US, Business Number for Canada, etc."
            },
            "validation_rules": {
                "required": true,
                "custom_rules": {
                    "business_number": {
                        "country": "{{incorporation_country}}"
                    }
                }
            },
            "options": {},
            "display_order": 6,
            "is_required": true,
            "is_readonly": false,
            "default_value": "",
            "conditional_logic": {},
            "file_config": {},
            "created_at": "2025-05-25T12:58:29.350979+01:00",
            "updated_at": "2025-05-25T12:58:29.350979+01:00"
        },
        {
            "id": "c1337cba-4932-4338-a1fe-39ce535a8d1b",
            "form_definition_id": "97879a9d-818f-4f74-a953-f2696d7ead9a",
            "form_step_id": "b3c926a5-fef8-4b81-8630-b4bb2442f7dd",
            "field_name": "incorporation_date",
            "field_type": "date",
            "label": {
                "en": "Date of Incorporation"
            },
            "placeholder": {},
            "help_text": {},
            "validation_rules": {},
            "options": {},
            "display_order": 7,
            "is_required": true,
            "is_readonly": false,
            "default_value": "",
            "conditional_logic": {},
            "file_config": {},
            "created_at": "2025-05-25T12:58:29.350979+01:00",
            "updated_at": "2025-05-25T12:58:29.350979+01:00"
        },
        {
            "id": "b08f29f3-604e-4825-a9ce-3a60da592fbd",
            "form_definition_id": "97879a9d-818f-4f74-a953-f2696d7ead9a",
            "form_step_id": "b3c926a5-fef8-4b81-8630-b4bb2442f7dd",
            "field_name": "business_industry",
            "field_type": "select",
            "label": {
                "en": "Primary Industry"
            },
            "placeholder": {},
            "help_text": {},
            "validation_rules": {},
            "options": {
                "type": "static",
                "static": [
                    {
                        "label": {
                            "en": "Technology"
                        },
                        "value": "technology"
                    },
                    {
                        "label": {
                            "en": "Finance & Banking"
                        },
                        "value": "finance"
                    },
                    {
                        "label": {
                            "en": "Retail & E-commerce"
                        },
                        "value": "retail"
                    },
                    {
                        "label": {
                            "en": "Manufacturing"
                        },
                        "value": "manufacturing"
                    },
                    {
                        "label": {
                            "en": "Healthcare"
                        },
                        "value": "healthcare"
                    },
                    {
                        "label": {
                            "en": "Real Estate"
                        },
                        "value": "real_estate"
                    },
                    {
                        "label": {
                            "en": "Professional Services"
                        },
                        "value": "professional_services"
                    },
                    {
                        "label": {
                            "en": "Other"
                        },
                        "value": "other"
                    }
                ]
            },
            "display_order": 8,
            "is_required": true,
            "is_readonly": false,
            "default_value": "",
            "conditional_logic": {},
            "file_config": {},
            "created_at": "2025-05-25T12:58:29.350979+01:00",
            "updated_at": "2025-05-25T12:58:29.350979+01:00"
        },
        {
            "id": "a6643e7c-377c-4ff0-9e46-f410ecbad696",
            "form_definition_id": "97879a9d-818f-4f74-a953-f2696d7ead9a",
            "form_step_id": "b3c926a5-fef8-4b81-8630-b4bb2442f7dd",
            "field_name": "business_description",
            "field_type": "textarea",
            "label": {
                "en": "Business Description"
            },
            "placeholder": {
                "en": "Describe your business activities..."
            },
            "help_text": {},
            "validation_rules": {
                "required": true,
                "max_length": 1000,
                "min_length": 50
            },
            "options": {},
            "display_order": 9,
            "is_required": true,
            "is_readonly": false,
            "default_value": "",
            "conditional_logic": {},
            "file_config": {},
            "created_at": "2025-05-25T12:58:29.350979+01:00",
            "updated_at": "2025-05-25T12:58:29.350979+01:00"
        },
        {
            "id": "24a3dfec-50d0-44f7-ace8-dbc5d37bfacc",
            "form_definition_id": "97879a9d-818f-4f74-a953-f2696d7ead9a",
            "form_step_id": "fcfbe6ab-ba25-4a0a-acf3-5675fb791bae",
            "field_name": "certificate_of_incorporation",
            "field_type": "file",
            "label": {
                "en": "Certificate of Incorporation"
            },
            "placeholder": {},
            "help_text": {},
            "validation_rules": {},
            "options": {},
            "display_order": 10,
            "is_required": true,
            "is_readonly": false,
            "default_value": "",
            "conditional_logic": {},
            "file_config": {
                "max_size": 20971520,
                "allowed_types": ["application/pdf", "image/jpeg", "image/png"]
            },
            "created_at": "2025-05-25T12:58:29.350979+01:00",
            "updated_at": "2025-05-25T12:58:29.350979+01:00"
        },
        {
            "id": "1d2a2ea3-2e98-4c6d-be46-900e7f7cf6a3",
            "form_definition_id": "97879a9d-818f-4f74-a953-f2696d7ead9a",
            "form_step_id": "fcfbe6ab-ba25-4a0a-acf3-5675fb791bae",
            "field_name": "articles_of_association",
            "field_type": "file",
            "label": {
                "en": "Articles of Association / Operating Agreement"
            },
            "placeholder": {},
            "help_text": {},
            "validation_rules": {},
            "options": {},
            "display_order": 11,
            "is_required": true,
            "is_readonly": false,
            "default_value": "",
            "conditional_logic": {},
            "file_config": {
                "max_size": 20971520,
                "allowed_types": ["application/pdf"]
            },
            "created_at": "2025-05-25T12:58:29.350979+01:00",
            "updated_at": "2025-05-25T12:58:29.350979+01:00"
        },
        {
            "id": "e50ff207-0394-40f2-a9ae-411b8504ac8b",
            "form_definition_id": "97879a9d-818f-4f74-a953-f2696d7ead9a",
            "form_step_id": "fcfbe6ab-ba25-4a0a-acf3-5675fb791bae",
            "field_name": "business_licenses",
            "field_type": "files",
            "label": {
                "en": "Business Licenses"
            },
            "placeholder": {},
            "help_text": {
                "en": "Upload all relevant business licenses"
            },
            "validation_rules": {},
            "options": {},
            "display_order": 12,
            "is_required": false,
            "is_readonly": false,
            "default_value": "",
            "conditional_logic": {},
            "file_config": {
                "max_size": 10485760,
                "max_files": 5,
                "allowed_types": ["application/pdf", "image/jpeg", "image/png"]
            },
            "created_at": "2025-05-25T12:58:29.350979+01:00",
            "updated_at": "2025-05-25T12:58:29.350979+01:00"
        },
        {
            "id": "ac65e8c7-eafe-40ca-a446-7c91b39dec69",
            "form_definition_id": "97879a9d-818f-4f74-a953-f2696d7ead9a",
            "form_step_id": "fcfbe6ab-ba25-4a0a-acf3-5675fb791bae",
            "field_name": "bank_statements",
            "field_type": "files",
            "label": {
                "en": "Bank Statements (Last 3 months)"
            },
            "placeholder": {},
            "help_text": {},
            "validation_rules": {},
            "options": {},
            "display_order": 13,
            "is_required": true,
            "is_readonly": false,
            "default_value": "",
            "conditional_logic": {},
            "file_config": {
                "max_size": 10485760,
                "max_files": 3,
                "allowed_types": ["application/pdf"]
            },
            "created_at": "2025-05-25T12:58:29.350979+01:00",
            "updated_at": "2025-05-25T12:58:29.350979+01:00"
        },
        {
            "id": "de9b1cea-9156-4204-869a-fe2431fb1d8d",
            "form_definition_id": "97879a9d-818f-4f74-a953-f2696d7ead9a",
            "form_step_id": "a610e5ae-aa65-4ee0-b5c9-9a03d0c5fa0f",
            "field_name": "number_of_owners",
            "field_type": "number",
            "label": {
                "en": "Number of Owners/Shareholders"
            },
            "placeholder": {},
            "help_text": {},
            "validation_rules": {
                "max": 20,
                "min": 1,
                "required": true
            },
            "options": {},
            "display_order": 14,
            "is_required": true,
            "is_readonly": false,
            "default_value": "",
            "conditional_logic": {},
            "file_config": {},
            "created_at": "2025-05-25T12:58:29.350979+01:00",
            "updated_at": "2025-05-25T12:58:29.350979+01:00"
        },
        {
            "id": "6dd2133f-e58b-4354-a9cf-68a2d441b807",
            "form_definition_id": "97879a9d-818f-4f74-a953-f2696d7ead9a",
            "form_step_id": "a610e5ae-aa65-4ee0-b5c9-9a03d0c5fa0f",
            "field_name": "owner_1_name",
            "field_type": "text",
            "label": {
                "en": "Owner 1 - Full Name"
            },
            "placeholder": {},
            "help_text": {},
            "validation_rules": {},
            "options": {},
            "display_order": 15,
            "is_required": true,
            "is_readonly": false,
            "default_value": "",
            "conditional_logic": {},
            "file_config": {},
            "created_at": "2025-05-25T12:58:29.350979+01:00",
            "updated_at": "2025-05-25T12:58:29.350979+01:00"
        },
        {
            "id": "7a72d005-ef63-419b-969a-ff7a70bc603b",
            "form_definition_id": "97879a9d-818f-4f74-a953-f2696d7ead9a",
            "form_step_id": "a610e5ae-aa65-4ee0-b5c9-9a03d0c5fa0f",
            "field_name": "owner_1_ownership_percentage",
            "field_type": "number",
            "label": {
                "en": "Owner 1 - Ownership Percentage"
            },
            "placeholder": {},
            "help_text": {},
            "validation_rules": {
                "max": 100,
                "min": 0,
                "required": true
            },
            "options": {},
            "display_order": 16,
            "is_required": true,
            "is_readonly": false,
            "default_value": "",
            "conditional_logic": {},
            "file_config": {},
            "created_at": "2025-05-25T12:58:29.350979+01:00",
            "updated_at": "2025-05-25T12:58:29.350979+01:00"
        },
        {
            "id": "ab6ab6da-70cb-46ce-9c1d-969f068600dc",
            "form_definition_id": "97879a9d-818f-4f74-a953-f2696d7ead9a",
            "form_step_id": "a610e5ae-aa65-4ee0-b5c9-9a03d0c5fa0f",
            "field_name": "owner_1_is_ubo",
            "field_type": "checkbox",
            "label": {
                "en": "Owner 1 - Ultimate Beneficial Owner (25%+ ownership)"
            },
            "placeholder": {},
            "help_text": {},
            "validation_rules": {},
            "options": {},
            "display_order": 17,
            "is_required": false,
            "is_readonly": false,
            "default_value": "",
            "conditional_logic": {},
            "file_config": {},
            "created_at": "2025-05-25T12:58:29.350979+01:00",
            "updated_at": "2025-05-25T12:58:29.350979+01:00"
        },
        {
            "id": "daeffa4b-f754-428f-8f50-8b4981abf6e8",
            "form_definition_id": "97879a9d-818f-4f74-a953-f2696d7ead9a",
            "form_step_id": "a610e5ae-aa65-4ee0-b5c9-9a03d0c5fa0f",
            "field_name": "owner_1_id_document",
            "field_type": "file",
            "label": {
                "en": "Owner 1 - ID Document"
            },
            "placeholder": {},
            "help_text": {},
            "validation_rules": {},
            "options": {},
            "display_order": 18,
            "is_required": true,
            "is_readonly": false,
            "default_value": "",
            "conditional_logic": {},
            "file_config": {
                "max_size": 10485760,
                "allowed_types": ["application/pdf", "image/jpeg", "image/png"]
            },
            "created_at": "2025-05-25T12:58:29.350979+01:00",
            "updated_at": "2025-05-25T12:58:29.350979+01:00"
        },
        {
            "id": "34668bcf-406e-4a7b-880d-33752d3ce163",
            "form_definition_id": "97879a9d-818f-4f74-a953-f2696d7ead9a",
            "form_step_id": "4e5c1a7e-eb35-47ae-b1b7-5592ac31b32a",
            "field_name": "primary_bank",
            "field_type": "select",
            "label": {
                "en": "Primary Bank"
            },
            "placeholder": {},
            "help_text": {},
            "validation_rules": {},
            "options": {
                "type": "dynamic",
                "dynamic": {
                    "source_name": "banks",
                    "filter_params": {
                        "country": "{{incorporation_country}}"
                    }
                }
            },
            "display_order": 19,
            "is_required": true,
            "is_readonly": false,
            "default_value": "",
            "conditional_logic": {},
            "file_config": {},
            "created_at": "2025-05-25T12:58:29.350979+01:00",
            "updated_at": "2025-05-25T12:58:29.350979+01:00"
        },
        {
            "id": "990b1217-c9a8-4b0e-ae57-e76a7f44dcf0",
            "form_definition_id": "97879a9d-818f-4f74-a953-f2696d7ead9a",
            "form_step_id": "4e5c1a7e-eb35-47ae-b1b7-5592ac31b32a",
            "field_name": "bank_account_number",
            "field_type": "text",
            "label": {
                "en": "Bank Account Number"
            },
            "placeholder": {},
            "help_text": {},
            "validation_rules": {
                "required": true,
                "custom_rules": {
                    "bank_account": {
                        "bank": "{{primary_bank}}"
                    }
                }
            },
            "options": {},
            "display_order": 20,
            "is_required": true,
            "is_readonly": false,
            "default_value": "",
            "conditional_logic": {},
            "file_config": {},
            "created_at": "2025-05-25T12:58:29.350979+01:00",
            "updated_at": "2025-05-25T12:58:29.350979+01:00"
        },
        {
            "id": "8451ce9d-34f0-4471-b86a-aaa9b002f821",
            "form_definition_id": "97879a9d-818f-4f74-a953-f2696d7ead9a",
            "form_step_id": "4e5c1a7e-eb35-47ae-b1b7-5592ac31b32a",
            "field_name": "annual_revenue",
            "field_type": "currency",
            "label": {
                "en": "Annual Revenue (Last Year)"
            },
            "placeholder": {},
            "help_text": {},
            "validation_rules": {
                "max": 1000000000,
                "min": 0,
                "required": true
            },
            "options": {},
            "display_order": 21,
            "is_required": true,
            "is_readonly": false,
            "default_value": "",
            "conditional_logic": {},
            "file_config": {},
            "created_at": "2025-05-25T12:58:29.350979+01:00",
            "updated_at": "2025-05-25T12:58:29.350979+01:00"
        },
        {
            "id": "addce105-9d66-458c-8245-bd6ecb7d6ab2",
            "form_definition_id": "97879a9d-818f-4f74-a953-f2696d7ead9a",
            "form_step_id": "4e5c1a7e-eb35-47ae-b1b7-5592ac31b32a",
            "field_name": "expected_monthly_volume",
            "field_type": "currency",
            "label": {
                "en": "Expected Monthly Transaction Volume"
            },
            "placeholder": {},
            "help_text": {},
            "validation_rules": {},
            "options": {},
            "display_order": 22,
            "is_required": true,
            "is_readonly": false,
            "default_value": "",
            "conditional_logic": {},
            "file_config": {},
            "created_at": "2025-05-25T12:58:29.350979+01:00",
            "updated_at": "2025-05-25T12:58:29.350979+01:00"
        },
        {
            "id": "62dbe394-04b7-4361-9b79-530c4855c38f",
            "form_definition_id": "97879a9d-818f-4f74-a953-f2696d7ead9a",
            "form_step_id": "4e5c1a7e-eb35-47ae-b1b7-5592ac31b32a",
            "field_name": "source_of_funds",
            "field_type": "checkbox",
            "label": {
                "en": "Primary Sources of Funds"
            },
            "placeholder": {},
            "help_text": {},
            "validation_rules": {},
            "options": {
                "type": "static",
                "static": [
                    {
                        "label": {
                            "en": "Sales Revenue"
                        },
                        "value": "sales_revenue"
                    },
                    {
                        "label": {
                            "en": "Investments"
                        },
                        "value": "investments"
                    },
                    {
                        "label": {
                            "en": "Business Loans"
                        },
                        "value": "loans"
                    },
                    {
                        "label": {
                            "en": "Grants"
                        },
                        "value": "grants"
                    },
                    {
                        "label": {
                            "en": "Other"
                        },
                        "value": "other"
                    }
                ]
            },
            "display_order": 23,
            "is_required": true,
            "is_readonly": false,
            "default_value": "",
            "conditional_logic": {},
            "file_config": {},
            "created_at": "2025-05-25T12:58:29.350979+01:00",
            "updated_at": "2025-05-25T12:58:29.350979+01:00"
        },
        {
            "id": "0a0dec6a-3153-4867-8342-cfda8b0fbcc0",
            "form_definition_id": "97879a9d-818f-4f74-a953-f2696d7ead9a",
            "form_step_id": "50be3b87-a4b6-4b60-ab5b-9d76424d55a7",
            "field_name": "has_compliance_officer",
            "field_type": "radio",
            "label": {
                "en": "Do you have a designated compliance officer?"
            },
            "placeholder": {},
            "help_text": {},
            "validation_rules": {},
            "options": {
                "type": "static",
                "static": [
                    {
                        "label": {
                            "en": "Yes"
                        },
                        "value": "yes"
                    },
                    {
                        "label": {
                            "en": "No"
                        },
                        "value": "no"
                    }
                ]
            },
            "display_order": 24,
            "is_required": true,
            "is_readonly": false,
            "default_value": "",
            "conditional_logic": {},
            "file_config": {},
            "created_at": "2025-05-25T12:58:29.350979+01:00",
            "updated_at": "2025-05-25T12:58:29.350979+01:00"
        },
        {
            "id": "5a0b91fd-fa30-428d-bcd8-42bb994fe9f0",
            "form_definition_id": "97879a9d-818f-4f74-a953-f2696d7ead9a",
            "form_step_id": "50be3b87-a4b6-4b60-ab5b-9d76424d55a7",
            "field_name": "compliance_officer_name",
            "field_type": "text",
            "label": {
                "en": "Compliance Officer Name"
            },
            "placeholder": {},
            "help_text": {},
            "validation_rules": {},
            "options": {},
            "display_order": 25,
            "is_required": true,
            "is_readonly": false,
            "default_value": "",
            "conditional_logic": {
                "logic": "all",
                "action": "show",
                "conditions": [
                    {
                        "field": "has_compliance_officer",
                        "value": "yes",
                        "operator": "equals"
                    }
                ]
            },
            "file_config": {},
            "created_at": "2025-05-25T12:58:29.350979+01:00",
            "updated_at": "2025-05-25T12:58:29.350979+01:00"
        },
        {
            "id": "0774f4dd-c0d9-4ec5-86cd-ea1acf1e8697",
            "form_definition_id": "97879a9d-818f-4f74-a953-f2696d7ead9a",
            "form_step_id": "50be3b87-a4b6-4b60-ab5b-9d76424d55a7",
            "field_name": "aml_policy",
            "field_type": "file",
            "label": {
                "en": "Anti-Money Laundering (AML) Policy"
            },
            "placeholder": {},
            "help_text": {
                "en": "Upload your AML policy document if available"
            },
            "validation_rules": {},
            "options": {},
            "display_order": 26,
            "is_required": false,
            "is_readonly": false,
            "default_value": "",
            "conditional_logic": {},
            "file_config": {
                "max_size": 20971520,
                "allowed_types": [
                    "application/pdf",
                    "application/msword",
                    "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
                ]
            },
            "created_at": "2025-05-25T12:58:29.350979+01:00",
            "updated_at": "2025-05-25T12:58:29.350979+01:00"
        },
        {
            "id": "b7e8ec69-0e65-4ff1-aec8-c859be32dc35",
            "form_definition_id": "97879a9d-818f-4f74-a953-f2696d7ead9a",
            "form_step_id": "50be3b87-a4b6-4b60-ab5b-9d76424d55a7",
            "field_name": "certifications",
            "field_type": "checkbox",
            "label": {
                "en": "Certifications and Declarations"
            },
            "placeholder": {},
            "help_text": {},
            "validation_rules": {
                "required": true,
                "all_required": true
            },
            "options": {
                "type": "static",
                "static": [
                    {
                        "label": {
                            "en": "I certify that all information provided is accurate and complete"
                        },
                        "value": "info_accurate"
                    },
                    {
                        "label": {
                            "en": "I am authorized to provide this information on behalf of the business"
                        },
                        "value": "authorized"
                    },
                    {
                        "label": {
                            "en": "I agree to cooperate with any verification requests"
                        },
                        "value": "agree_verification"
                    },
                    {
                        "label": {
                            "en": "I agree to the Terms of Service and Privacy Policy"
                        },
                        "value": "agree_terms"
                    }
                ]
            },
            "display_order": 27,
            "is_required": true,
            "is_readonly": false,
            "default_value": "",
            "conditional_logic": {},
            "file_config": {},
            "created_at": "2025-05-25T12:58:29.350979+01:00",
            "updated_at": "2025-05-25T12:58:29.350979+01:00"
        }
    ],
        "submission_id": null,
        "current_step": 0,
        "completion_percentage": 0
},
    "timestamp": "Sunday, 25-May-25 18:01:20 CEST",
    "status": "OK"
}`
