// src/components/FormRenderer/types.ts

export interface FormDefinition {
  id: string;
  name: string;
  description: string;
  form_type: string;
  is_multi_step: boolean;
  requires_approval: boolean;
  is_editable_after_submission: boolean;
}

export interface FormStep {
  id: string;
  step_number: number;
  name: string;
  description: string;
  is_optional: boolean;
}

export interface FormField {
  id: string;
  form_step_id: string;
  field_name: string;
  field_type: string;
  label: Record<string, string>;
  placeholder?: Record<string, string>;
  help_text?: Record<string, string>;
  validation_rules: Record<string, any>;
  options?: Record<string, any>;
  display_order: number;
  is_required: boolean;
  is_readonly: boolean;
  default_value: string;
  conditional_logic?: Record<string, any>;
  file_config?: Record<string, any>;
}

export interface FormData {
  form_definition: FormDefinition;
  steps: FormStep[];
  fields: FormField[];
  existing_data?: Record<string, any>;
  submission_id?: string;
  current_step: number;
  completion_percentage: number;
}

export interface FormRendererProps {
  formData: FormData;
  onSubmit: (data: Record<string, any>, isDraft?: boolean) => void;
  onStepSubmit?: (stepNumber: number, data: Record<string, any>) => void;
  isLoading?: boolean;
}

export interface FieldProps {
  field: FormField;
  value: any;
  onChange: (value: any) => void;
  errors: string[];
  formValues?: Record<string, any>;
}

// Additional utility types that might be needed
export type ValidationRule = {
  required?: boolean;
  min_length?: number;
  max_length?: number;
  pattern?: string;
  min?: number;
  max?: number;
  email?: boolean;
  min_items?: number;
  max_items?: number;
  all_required?: boolean;
  custom_rules?: Record<string, any>;
};

export type OptionValue = {
  value: string;
  label: Record<string, string>;
};

export type FileConfig = {
  max_size?: number;
  allowed_types?: string[];
  max_files?: number;
};

export type ConditionalLogic = {
  action: "show" | "hide";
  logic: "all" | "any";
  conditions: Array<{
    field: string;
    operator: "equals" | "not_equals" | "in" | "not_in";
    value: any;
  }>;
};
