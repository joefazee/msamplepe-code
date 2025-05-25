// EXAMPLE SINGLE STEP FORM RESPONSE

const response = `{
    "message": "Form retrieved successfully",
    "data": {
        "form_definition": {
            "id": "a18baade-a2af-4ff2-b150-31623bcce4d6",
            "name": "Contact Us Form",
            "slug": "about-us-form",
            "description": "Simple contact form for customer inquiries",
            "form_type": "about",
            "version": 1,
            "is_active": true,
            "is_multi_step": false,
            "requires_approval": false,
            "is_editable_after_submission": false,
            "approval_workflow": {},
            "created_at": "2025-05-25T12:50:14.074339+01:00",
            "updated_at": "2025-05-25T12:50:14.074339+01:00",
            "created_by": "37b3f0a3-4059-49c8-8fe0-9da3f3e02585"
        },
        "steps": [
            {
                "id": "b3f857ff-9dba-40fe-8087-79d33853a08b",
                "form_definition_id": "a18baade-a2af-4ff2-b150-31623bcce4d6",
                "step_number": 1,
                "name": "Contact Us Form",
                "description": "Simple contact form for customer inquiries",
                "is_optional": false,
                "created_at": "2025-05-25T12:50:14.074339+01:00"
            }
        ],
        "fields": [
            {
                "id": "c0b89579-ae7c-4a1d-884c-74368f04b2be",
                "form_definition_id": "a18baade-a2af-4ff2-b150-31623bcce4d6",
                "form_step_id": "b3f857ff-9dba-40fe-8087-79d33853a08b",
                "field_name": "name",
                "field_type": "text",
                "label": {
                    "en": "Your Name",
                    "fr": "Votre nom"
                },
                "placeholder": {
                    "en": "Enter your full name",
                    "fr": "Entrez votre nom complet"
                },
                "help_text": {},
                "validation_rules": {
                    "required": true,
                    "max_length": 100,
                    "min_length": 2
                },
                "options": {},
                "display_order": 1,
                "is_required": true,
                "is_readonly": false,
                "default_value": "",
                "conditional_logic": {},
                "file_config": {},
                "created_at": "2025-05-25T12:50:14.074339+01:00",
                "updated_at": "2025-05-25T12:50:14.074339+01:00"
            },
            {
                "id": "b982ef31-d345-4f7a-aea9-c6e0b98c47ee",
                "form_definition_id": "a18baade-a2af-4ff2-b150-31623bcce4d6",
                "form_step_id": "b3f857ff-9dba-40fe-8087-79d33853a08b",
                "field_name": "email",
                "field_type": "email",
                "label": {
                    "en": "Email Address",
                    "fr": "Adresse e-mail"
                },
                "placeholder": {
                    "en": "your@email.com",
                    "fr": "votre@email.com"
                },
                "help_text": {},
                "validation_rules": {
                    "email": true,
                    "required": true
                },
                "options": {},
                "display_order": 2,
                "is_required": true,
                "is_readonly": false,
                "default_value": "",
                "conditional_logic": {},
                "file_config": {},
                "created_at": "2025-05-25T12:50:14.074339+01:00",
                "updated_at": "2025-05-25T12:50:14.074339+01:00"
            },
            {
                "id": "3b25ba47-9138-422d-ae4a-914beea3be13",
                "form_definition_id": "a18baade-a2af-4ff2-b150-31623bcce4d6",
                "form_step_id": "b3f857ff-9dba-40fe-8087-79d33853a08b",
                "field_name": "subject",
                "field_type": "select",
                "label": {
                    "en": "Subject",
                    "fr": "Sujet"
                },
                "placeholder": {},
                "help_text": {},
                "validation_rules": {
                    "required": true
                },
                "options": {
                    "type": "static",
                    "static": [
                        {
                            "label": {
                                "en": "General Inquiry",
                                "fr": "Demande générale"
                            },
                            "value": "general"
                        },
                        {
                            "label": {
                                "en": "Technical Support",
                                "fr": "Support technique"
                            },
                            "value": "support"
                        },
                        {
                            "label": {
                                "en": "Billing Question",
                                "fr": "Question de facturation"
                            },
                            "value": "billing"
                        },
                        {
                            "label": {
                                "en": "Feedback",
                                "fr": "Commentaires"
                            },
                            "value": "feedback"
                        }
                    ]
                },
                "display_order": 3,
                "is_required": true,
                "is_readonly": false,
                "default_value": "",
                "conditional_logic": {},
                "file_config": {},
                "created_at": "2025-05-25T12:50:14.074339+01:00",
                "updated_at": "2025-05-25T12:50:14.074339+01:00"
            },
            {
                "id": "5d63f912-11f6-4436-8855-75b755fcb778",
                "form_definition_id": "a18baade-a2af-4ff2-b150-31623bcce4d6",
                "form_step_id": "b3f857ff-9dba-40fe-8087-79d33853a08b",
                "field_name": "message",
                "field_type": "textarea",
                "label": {
                    "en": "Message",
                    "fr": "Message"
                },
                "placeholder": {
                    "en": "Type your message here...",
                    "fr": "Tapez votre message ici..."
                },
                "help_text": {
                    "en": "Please provide as much detail as possible",
                    "fr": "Veuillez fournir autant de détails que possible"
                },
                "validation_rules": {
                    "required": true,
                    "max_length": 1000,
                    "min_length": 10
                },
                "options": {},
                "display_order": 4,
                "is_required": true,
                "is_readonly": false,
                "default_value": "",
                "conditional_logic": {},
                "file_config": {},
                "created_at": "2025-05-25T12:50:14.074339+01:00",
                "updated_at": "2025-05-25T12:50:14.074339+01:00"
            },
            {
                "id": "6cf4d3ae-06b1-428b-9fae-2995d4fe94ab",
                "form_definition_id": "a18baade-a2af-4ff2-b150-31623bcce4d6",
                "form_step_id": "b3f857ff-9dba-40fe-8087-79d33853a08b",
                "field_name": "attachment",
                "field_type": "file",
                "label": {
                    "en": "Attachment (optional)",
                    "fr": "Pièce jointe (facultatif)"
                },
                "placeholder": {},
                "help_text": {
                    "en": "You can attach a file up to 5MB",
                    "fr": "Vous pouvez joindre un fichier jusqu'à 5 Mo"
                },
                "validation_rules": {},
                "options": {},
                "display_order": 5,
                "is_required": false,
                "is_readonly": false,
                "default_value": "",
                "conditional_logic": {},
                "file_config": {
                    "max_size": 5242880,
                    "allowed_types": [
                        "image/jpeg",
                        "image/png",
                        "application/pdf"
                    ]
                },
                "created_at": "2025-05-25T12:50:14.074339+01:00",
                "updated_at": "2025-05-25T12:50:14.074339+01:00"
            }
        ],
        "existing_data": {},
        "submission_id": null,
        "current_step": 0,
        "completion_percentage": 0
    },
    "timestamp": "Sunday, 25-May-25 18:10:33 CEST",
    "status": "OK"
}`
