// src/components/FormRenderer/FormRenderer.tsx (Updated)

import React, {
  useState,
  useEffect,
  useCallback,
  useMemo,
  useRef,
  FormEvent,
} from "react";
import {
  ChevronLeft,
  ChevronRight,
  Upload,
  X,
  AlertCircle,
  CheckCircle,
} from "lucide-react";
import type { FormRendererProps, FormField, FormStep, FormData } from "./types";

// ==================== Utility Functions ====================
const getLocalizedText = (
  textObj: Record<string, string> | undefined,
  locale = "en"
): string => {
  if (!textObj) return "";
  return textObj[locale] || textObj["en"] || Object.values(textObj)[0] || "";
};

const validateField = (field: FormField, value: any): string[] => {
  const errors: string[] = [];
  const rules = field.validation_rules || {};

  // Required validation
  if (
    field.is_required &&
    (!value || (typeof value === "string" && value.trim() === ""))
  ) {
    errors.push(`${getLocalizedText(field.label)} is required`);
    return errors;
  }

  if (!value && !field.is_required) return errors;

  // Type-specific validation
  switch (field.field_type) {
    case "email":
      if (value && !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value)) {
        errors.push("Invalid email format");
      }
      break;

    case "text":
    case "textarea":
      if (typeof value === "string") {
        if (rules.min_length && value.length < rules.min_length) {
          errors.push(`Must be at least ${rules.min_length} characters`);
        }
        if (rules.max_length && value.length > rules.max_length) {
          errors.push(`Must be at most ${rules.max_length} characters`);
        }
        if (rules.pattern && !new RegExp(rules.pattern).test(value)) {
          errors.push("Invalid format");
        }
      }
      break;

    case "number":
    case "currency":
      const num = Number(value);
      if (isNaN(num)) {
        errors.push("Must be a valid number");
      } else {
        if (rules.min !== undefined && num < rules.min) {
          errors.push(`Must be at least ${rules.min}`);
        }
        if (rules.max !== undefined && num > rules.max) {
          errors.push(`Must be at most ${rules.max}`);
        }
      }
      break;

    case "checkbox":
      if (
        rules.min_items &&
        Array.isArray(value) &&
        value.length < rules.min_items
      ) {
        errors.push(`Select at least ${rules.min_items} options`);
      }
      if (rules.all_required && field.options?.static) {
        const requiredCount = field.options.static.length;
        if (!Array.isArray(value) || value.length !== requiredCount) {
          errors.push("All options must be selected");
        }
      }
      break;
  }

  return errors;
};

const checkConditionalLogic = (
  field: FormField,
  formValues: Record<string, any>
): boolean => {
  const logic = field.conditional_logic;
  if (!logic || !logic.conditions) return true;

  const { conditions, logic: logicType = "all", action = "show" } = logic;

  const results = conditions.map((condition: any) => {
    const fieldValue = formValues[condition.field];
    const conditionValue = condition.value;

    switch (condition.operator) {
      case "equals":
        return fieldValue === conditionValue;
      case "not_equals":
        return fieldValue !== conditionValue;
      case "in":
        return (
          Array.isArray(conditionValue) && conditionValue.includes(fieldValue)
        );
      case "not_in":
        return (
          Array.isArray(conditionValue) && !conditionValue.includes(fieldValue)
        );
      default:
        return true;
    }
  });

  const shouldShow =
    logicType === "all" ? results.every(Boolean) : results.some(Boolean);
  return action === "show" ? shouldShow : !shouldShow;
};

// ==================== Field Components ====================
const TextInput: React.FC<{
  field: FormField;
  value: string;
  onChange: (value: string) => void;
  errors: string[];
}> = ({ field, value, onChange, errors }) => {
  const hasError = errors.length > 0;

  return (
    <div className="mb-4">
      <label className="block text-sm font-medium text-gray-700 mb-1">
        {getLocalizedText(field.label)}
        {field.is_required && <span className="text-red-500 ml-1">*</span>}
      </label>
      <input
        type={field.field_type === "email" ? "email" : "text"}
        name={field.field_name}
        value={value || ""}
        onChange={(e) => onChange(e.target.value)}
        placeholder={getLocalizedText(field.placeholder)}
        disabled={field.is_readonly}
        className={`w-full px-3 py-2 border rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 ${
          hasError ? "border-red-500" : "border-gray-300"
        } ${field.is_readonly ? "bg-gray-100" : ""}`}
      />
      {getLocalizedText(field.help_text) && (
        <p className="mt-1 text-sm text-gray-500">
          {getLocalizedText(field.help_text)}
        </p>
      )}
      {errors.map((error, index) => (
        <p key={index} className="mt-1 text-sm text-red-600 flex items-center">
          <AlertCircle className="w-4 h-4 mr-1" />
          {error}
        </p>
      ))}
    </div>
  );
};

const TextArea: React.FC<{
  field: FormField;
  value: string;
  onChange: (value: string) => void;
  errors: string[];
}> = ({ field, value, onChange, errors }) => {
  const hasError = errors.length > 0;

  return (
    <div className="mb-4">
      <label className="block text-sm font-medium text-gray-700 mb-1">
        {getLocalizedText(field.label)}
        {field.is_required && <span className="text-red-500 ml-1">*</span>}
      </label>
      <textarea
        name={field.field_name}
        value={value || ""}
        onChange={(e) => onChange(e.target.value)}
        placeholder={getLocalizedText(field.placeholder)}
        disabled={field.is_readonly}
        rows={4}
        className={`w-full px-3 py-2 border rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 ${
          hasError ? "border-red-500" : "border-gray-300"
        } ${field.is_readonly ? "bg-gray-100" : ""}`}
      />
      {getLocalizedText(field.help_text) && (
        <p className="mt-1 text-sm text-gray-500">
          {getLocalizedText(field.help_text)}
        </p>
      )}
      {errors.map((error, index) => (
        <p key={index} className="mt-1 text-sm text-red-600 flex items-center">
          <AlertCircle className="w-4 h-4 mr-1" />
          {error}
        </p>
      ))}
    </div>
  );
};

const SelectInput: React.FC<{
  field: FormField;
  value: string;
  onChange: (value: string) => void;
  errors: string[];
}> = ({ field, value, onChange, errors }) => {
  const hasError = errors.length > 0;
  const options = field.options?.static || [];

  return (
    <div className="mb-4">
      <label className="block text-sm font-medium text-gray-700 mb-1">
        {getLocalizedText(field.label)}
        {field.is_required && <span className="text-red-500 ml-1">*</span>}
      </label>
      <select
        name={field.field_name}
        value={value || ""}
        onChange={(e) => onChange(e.target.value)}
        disabled={field.is_readonly}
        className={`w-full px-3 py-2 border rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 ${
          hasError ? "border-red-500" : "border-gray-300"
        } ${field.is_readonly ? "bg-gray-100" : ""}`}
      >
        <option value="">Select an option...</option>
        {options.map((option: any, index: number) => (
          <option key={index} value={option.value}>
            {getLocalizedText(option.label)}
          </option>
        ))}
      </select>
      {getLocalizedText(field.help_text) && (
        <p className="mt-1 text-sm text-gray-500">
          {getLocalizedText(field.help_text)}
        </p>
      )}
      {errors.map((error, index) => (
        <p key={index} className="mt-1 text-sm text-red-600 flex items-center">
          <AlertCircle className="w-4 h-4 mr-1" />
          {error}
        </p>
      ))}
    </div>
  );
};

const RadioInput: React.FC<{
  field: FormField;
  value: string;
  onChange: (value: string) => void;
  errors: string[];
}> = ({ field, value, onChange, errors }) => {
  const options = field.options?.static || [];

  return (
    <div className="mb-4">
      <label className="block text-sm font-medium text-gray-700 mb-2">
        {getLocalizedText(field.label)}
        {field.is_required && <span className="text-red-500 ml-1">*</span>}
      </label>
      <div className="space-y-2">
        {options.map((option: any, index: number) => (
          <label key={index} className="flex items-center">
            <input
              type="radio"
              name={field.field_name}
              value={option.value}
              checked={value === option.value}
              onChange={(e) => onChange(e.target.value)}
              disabled={field.is_readonly}
              className="mr-2 text-blue-600 focus:ring-blue-500"
            />
            <span className="text-sm text-gray-700">
              {getLocalizedText(option.label)}
            </span>
          </label>
        ))}
      </div>
      {getLocalizedText(field.help_text) && (
        <p className="mt-1 text-sm text-gray-500">
          {getLocalizedText(field.help_text)}
        </p>
      )}
      {errors.map((error, index) => (
        <p key={index} className="mt-1 text-sm text-red-600 flex items-center">
          <AlertCircle className="w-4 h-4 mr-1" />
          {error}
        </p>
      ))}
    </div>
  );
};

const CheckboxInput: React.FC<{
  field: FormField;
  value: string[];
  onChange: (value: string[]) => void;
  errors: string[];
}> = ({ field, value = [], onChange, errors }) => {
  const options = field.options?.static || [];

  const handleChange = (optionValue: string, checked: boolean) => {
    if (checked) {
      onChange([...value, optionValue]);
    } else {
      onChange(value.filter((v) => v !== optionValue));
    }
  };

  return (
    <div className="mb-4">
      <label className="block text-sm font-medium text-gray-700 mb-2">
        {getLocalizedText(field.label)}
        {field.is_required && <span className="text-red-500 ml-1">*</span>}
      </label>
      <div className="space-y-2">
        {options.map((option: any, index: number) => (
          <label key={index} className="flex items-center">
            <input
              type="checkbox"
              name={field.field_name}
              value={option.value}
              checked={value.includes(option.value)}
              onChange={(e) => handleChange(option.value, e.target.checked)}
              disabled={field.is_readonly}
              className="mr-2 text-blue-600 focus:ring-blue-500 rounded"
            />
            <span className="text-sm text-gray-700">
              {getLocalizedText(option.label)}
            </span>
          </label>
        ))}
      </div>
      {getLocalizedText(field.help_text) && (
        <p className="mt-1 text-sm text-gray-500">
          {getLocalizedText(field.help_text)}
        </p>
      )}
      {errors.map((error, index) => (
        <p key={index} className="mt-1 text-sm text-red-600 flex items-center">
          <AlertCircle className="w-4 h-4 mr-1" />
          {error}
        </p>
      ))}
    </div>
  );
};

const FileInput: React.FC<{
  field: FormField;
  value: FileList | null;
  onChange: (value: FileList | null) => void;
  errors: string[];
}> = ({ field, value, onChange, errors }) => {
  const [dragActive, setDragActive] = useState(false);
  const fileConfig = field.file_config || {};
  const maxSize = fileConfig.max_size || 10 * 1024 * 1024; // 10MB default
  const allowedTypes = fileConfig.allowed_types || [];
  const maxFiles = field.field_type === "files" ? fileConfig.max_files || 5 : 1;

  const handleDrag = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    if (e.type === "dragenter" || e.type === "dragover") {
      setDragActive(true);
    } else if (e.type === "dragleave") {
      setDragActive(false);
    }
  };

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setDragActive(false);

    if (e.dataTransfer.files && e.dataTransfer.files[0]) {
      onChange(e.dataTransfer.files);
    }
  };

  const formatFileSize = (bytes: number) => {
    if (bytes === 0) return "0 Bytes";
    const k = 1024;
    const sizes = ["Bytes", "KB", "MB", "GB"];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
  };

  return (
    <div className="mb-4">
      <label className="block text-sm font-medium text-gray-700 mb-1">
        {getLocalizedText(field.label)}
        {field.is_required && <span className="text-red-500 ml-1">*</span>}
      </label>

      <div
        className={`border-2 border-dashed rounded-lg p-6 transition-colors ${
          dragActive
            ? "border-blue-500 bg-blue-50"
            : "border-gray-300 hover:border-gray-400"
        }`}
        onDragEnter={handleDrag}
        onDragLeave={handleDrag}
        onDragOver={handleDrag}
        onDrop={handleDrop}
      >
        <div className="text-center">
          <Upload className="mx-auto h-12 w-12 text-gray-400" />
          <div className="mt-4">
            <label
              htmlFor={`file-${field.field_name}`}
              className="cursor-pointer rounded-md bg-white font-medium text-blue-600 focus-within:outline-none focus-within:ring-2 focus-within:ring-blue-500 focus-within:ring-offset-2 hover:text-blue-500"
            >
              <span>Upload a file</span>
              <input
                id={`file-${field.field_name}`}
                name={field.field_name}
                type="file"
                className="sr-only"
                multiple={field.field_type === "files"}
                accept={allowedTypes.join(",")}
                onChange={(e) => onChange(e.target.files)}
                disabled={field.is_readonly}
              />
            </label>
            <p className="pl-1 inline">or drag and drop</p>
          </div>
          <p className="text-xs text-gray-500 mt-2">
            {allowedTypes.length > 0 && `${allowedTypes.join(", ")} `}
            up to {formatFileSize(maxSize)}
            {maxFiles > 1 && ` (max ${maxFiles} files)`}
          </p>
        </div>
      </div>

      {value && value.length > 0 && (
        <div className="mt-2 space-y-2">
          {Array.from(value).map((file, index) => (
            <div
              key={index}
              className="flex items-center justify-between p-2 bg-gray-50 rounded"
            >
              <div className="flex items-center">
                <CheckCircle className="w-4 h-4 text-green-500 mr-2" />
                <span className="text-sm text-gray-700">{file.name}</span>
                <span className="text-xs text-gray-500 ml-2">
                  ({formatFileSize(file.size)})
                </span>
              </div>
              <button
                type="button"
                onClick={() => onChange(null)}
                className="text-red-500 hover:text-red-700"
              >
                <X className="w-4 h-4" />
              </button>
            </div>
          ))}
        </div>
      )}

      {getLocalizedText(field.help_text) && (
        <p className="mt-1 text-sm text-gray-500">
          {getLocalizedText(field.help_text)}
        </p>
      )}
      {errors.map((error, index) => (
        <p key={index} className="mt-1 text-sm text-red-600 flex items-center">
          <AlertCircle className="w-4 h-4 mr-1" />
          {error}
        </p>
      ))}
    </div>
  );
};

const NumberInput: React.FC<{
  field: FormField;
  value: number | string;
  onChange: (value: number | string) => void;
  errors: string[];
}> = ({ field, value, onChange, errors }) => {
  const hasError = errors.length > 0;
  const isCurrency = field.field_type === "currency";

  return (
    <div className="mb-4">
      <label className="block text-sm font-medium text-gray-700 mb-1">
        {getLocalizedText(field.label)}
        {field.is_required && <span className="text-red-500 ml-1">*</span>}
      </label>
      <div className="relative">
        {isCurrency && (
          <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
            <span className="text-gray-500 sm:text-sm">$</span>
          </div>
        )}
        <input
          type="number"
          name={field.field_name}
          value={value || ""}
          onChange={(e) =>
            onChange(e.target.value ? Number(e.target.value) : "")
          }
          placeholder={getLocalizedText(field.placeholder)}
          disabled={field.is_readonly}
          className={`w-full px-3 py-2 border rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 ${
            hasError ? "border-red-500" : "border-gray-300"
          } ${field.is_readonly ? "bg-gray-100" : ""} ${
            isCurrency ? "pl-7" : ""
          }`}
        />
      </div>
      {getLocalizedText(field.help_text) && (
        <p className="mt-1 text-sm text-gray-500">
          {getLocalizedText(field.help_text)}
        </p>
      )}
      {errors.map((error, index) => (
        <p key={index} className="mt-1 text-sm text-red-600 flex items-center">
          <AlertCircle className="w-4 h-4 mr-1" />
          {error}
        </p>
      ))}
    </div>
  );
};

const DateInput: React.FC<{
  field: FormField;
  value: string;
  onChange: (value: string) => void;
  errors: string[];
}> = ({ field, value, onChange, errors }) => {
  const hasError = errors.length > 0;

  return (
    <div className="mb-4">
      <label className="block text-sm font-medium text-gray-700 mb-1">
        {getLocalizedText(field.label)}
        {field.is_required && <span className="text-red-500 ml-1">*</span>}
      </label>
      <input
        type="date"
        name={field.field_name}
        value={value || ""}
        onChange={(e) => onChange(e.target.value)}
        disabled={field.is_readonly}
        className={`w-full px-3 py-2 border rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 ${
          hasError ? "border-red-500" : "border-gray-300"
        } ${field.is_readonly ? "bg-gray-100" : ""}`}
      />
      {getLocalizedText(field.help_text) && (
        <p className="mt-1 text-sm text-gray-500">
          {getLocalizedText(field.help_text)}
        </p>
      )}
      {errors.map((error, index) => (
        <p key={index} className="mt-1 text-sm text-red-600 flex items-center">
          <AlertCircle className="w-4 h-4 mr-1" />
          {error}
        </p>
      ))}
    </div>
  );
};

// ==================== Field Renderer ====================
const FieldRenderer: React.FC<{
  field: FormField;
  value: any;
  onChange: (value: any) => void;
  errors: string[];
  formValues: Record<string, any>;
}> = ({ field, value, onChange, errors, formValues }) => {
  // Check conditional logic
  const shouldShow = checkConditionalLogic(field, formValues);

  if (!shouldShow) return null;

  switch (field.field_type) {
    case "text":
    case "email":
      return (
        <TextInput
          field={field}
          value={value}
          onChange={onChange}
          errors={errors}
        />
      );

    case "textarea":
      return (
        <TextArea
          field={field}
          value={value}
          onChange={onChange}
          errors={errors}
        />
      );

    case "select":
      return (
        <SelectInput
          field={field}
          value={value}
          onChange={onChange}
          errors={errors}
        />
      );

    case "radio":
      return (
        <RadioInput
          field={field}
          value={value}
          onChange={onChange}
          errors={errors}
        />
      );

    case "checkbox":
      return (
        <CheckboxInput
          field={field}
          value={value}
          onChange={onChange}
          errors={errors}
        />
      );

    case "file":
    case "files":
      return (
        <FileInput
          field={field}
          value={value}
          onChange={onChange}
          errors={errors}
        />
      );

    case "number":
    case "currency":
      return (
        <NumberInput
          field={field}
          value={value}
          onChange={onChange}
          errors={errors}
        />
      );

    case "date":
      return (
        <DateInput
          field={field}
          value={value}
          onChange={onChange}
          errors={errors}
        />
      );

    default:
      return (
        <TextInput
          field={field}
          value={value}
          onChange={onChange}
          errors={errors}
        />
      );
  }
};

// ==================== Progress Bar ====================
const ProgressBar: React.FC<{ percentage: number }> = ({ percentage }) => (
  <div className="w-full bg-gray-200 rounded-full h-2 mb-6">
    <div
      className="bg-blue-600 h-2 rounded-full transition-all duration-300"
      style={{ width: `${Math.min(percentage, 100)}%` }}
    ></div>
  </div>
);

// ==================== Step Navigation ====================
const StepNavigation: React.FC<{
  steps: FormStep[];
  currentStep: number;
  onStepChange: (step: number) => void;
  completedSteps: Set<number>;
}> = ({ steps, currentStep, onStepChange, completedSteps }) => (
  <div className="flex items-center justify-between mb-8">
    {steps.map((step, index) => {
      const stepNumber = step.step_number;
      const isActive = stepNumber === currentStep;
      const isCompleted = completedSteps.has(stepNumber);
      const isClickable = isCompleted || stepNumber <= currentStep;

      return (
        <React.Fragment key={step.id}>
          <div className="flex flex-col items-center">
            <button
              type="button"
              onClick={() => isClickable && onStepChange(stepNumber)}
              disabled={!isClickable}
              className={`w-10 h-10 rounded-full flex items-center justify-center text-sm font-medium border-2 transition-colors ${
                isActive
                  ? "bg-blue-600 border-blue-600 text-white"
                  : isCompleted
                  ? "bg-green-600 border-green-600 text-white"
                  : isClickable
                  ? "border-gray-300 text-gray-500 hover:border-blue-600"
                  : "border-gray-200 text-gray-300 cursor-not-allowed"
              }`}
            >
              {isCompleted ? <CheckCircle className="w-5 h-5" /> : stepNumber}
            </button>
            <span
              className={`mt-2 text-xs text-center ${
                isActive ? "text-blue-600 font-medium" : "text-gray-500"
              }`}
            >
              {step.name}
            </span>
          </div>
          {index < steps.length - 1 && (
            <div
              className={`flex-1 h-0.5 mx-4 ${
                isCompleted ? "bg-green-600" : "bg-gray-200"
              }`}
            />
          )}
        </React.Fragment>
      );
    })}
  </div>
);

// ==================== Main Form Renderer ====================
const FormRenderer: React.FC<FormRendererProps> = ({
  formData,
  onSubmit,
  onStepSubmit,
  isLoading = false,
}) => {
  const {
    form_definition,
    steps,
    fields,
    existing_data,
    current_step,
    completion_percentage,
  } = formData;

  // Form reference
  const formRef = useRef<HTMLFormElement>(null);

  // Initialize form values from existing data
  const [formValues, setFormValues] = useState<Record<string, any>>(() => {
    const values: Record<string, any> = {};

    // Set default values
    fields.forEach((field) => {
      if (field.default_value) {
        try {
          // Try to parse JSON for complex defaults
          values[field.field_name] = JSON.parse(field.default_value);
        } catch {
          values[field.field_name] = field.default_value;
        }
      }
    });

    // Merge with existing data
    if (existing_data) {
      Object.assign(values, existing_data);
    }

    return values;
  });

  const [currentStepNumber, setCurrentStepNumber] = useState(current_step || 1);
  const [errors, setErrors] = useState<Record<string, string[]>>({});
  const [completedSteps, setCompletedSteps] = useState<Set<number>>(new Set());

  // Get current step fields
  const currentStepFields = useMemo(() => {
    const currentStepData = steps.find(
      (step) => step.step_number === currentStepNumber
    );
    if (!currentStepData) return [];

    return fields
      .filter((field) => field.form_step_id === currentStepData.id)
      .sort((a, b) => a.display_order - b.display_order);
  }, [fields, steps, currentStepNumber]);

  // Validate current step
  const validateCurrentStep = useCallback(() => {
    const stepErrors: Record<string, string[]> = {};
    let hasErrors = false;

    currentStepFields.forEach((field) => {
      if (!checkConditionalLogic(field, formValues)) return;

      const fieldErrors = validateField(field, formValues[field.field_name]);
      if (fieldErrors.length > 0) {
        stepErrors[field.field_name] = fieldErrors;
        hasErrors = true;
      }
    });

    setErrors(stepErrors);
    return !hasErrors;
  }, [currentStepFields, formValues]);

  // Handle field value change
  const handleFieldChange = useCallback((fieldName: string, value: any) => {
    setFormValues((prev) => ({ ...prev, [fieldName]: value }));

    // Clear errors for this field
    setErrors((prev) => {
      const newErrors = { ...prev };
      delete newErrors[fieldName];
      return newErrors;
    });
  }, []);

  // Handle step navigation
  const handleStepChange = useCallback(
    (stepNumber: number) => {
      if (stepNumber === currentStepNumber) return;

      // Validate current step before moving
      if (stepNumber > currentStepNumber && !validateCurrentStep()) {
        return;
      }

      setCurrentStepNumber(stepNumber);
    },
    [currentStepNumber, validateCurrentStep]
  );

  // Handle next step
  const handleNextStep = useCallback(() => {
    if (!validateCurrentStep()) return;

    // Mark current step as completed
    setCompletedSteps((prev) => new Set([...prev, currentStepNumber]));

    // Save step progress if callback provided
    if (onStepSubmit) {
      const stepData = currentStepFields.reduce((acc, field) => {
        if (formValues[field.field_name] !== undefined) {
          acc[field.field_name] = formValues[field.field_name];
        }
        return acc;
      }, {} as Record<string, any>);

      onStepSubmit(currentStepNumber, stepData);
    }

    // Move to next step
    const nextStep = currentStepNumber + 1;
    if (nextStep <= steps.length) {
      setCurrentStepNumber(nextStep);
    }
  }, [
    validateCurrentStep,
    currentStepNumber,
    currentStepFields,
    formValues,
    onStepSubmit,
    steps.length,
  ]);

  // Handle previous step
  const handlePreviousStep = useCallback(() => {
    if (currentStepNumber > 1) {
      setCurrentStepNumber(currentStepNumber - 1);
    }
  }, [currentStepNumber]);

  // Handle form submission with preventDefault
  const handleFormSubmit = useCallback(
    (event: FormEvent<HTMLFormElement>) => {
      event.preventDefault(); // Prevent page reload

      if (!validateCurrentStep()) return;

      // For multi-step forms, ensure all required steps are completed
      if (form_definition.is_multi_step) {
        const requiredSteps = steps.filter((step) => !step.is_optional);
        const allRequiredCompleted = requiredSteps.every(
          (step) =>
            completedSteps.has(step.step_number) ||
            step.step_number === currentStepNumber
        );

        if (!allRequiredCompleted) {
          alert("Please complete all required steps before submitting.");
          return;
        }
      }

      onSubmit(formValues, false); // false = not a draft
    },
    [
      validateCurrentStep,
      form_definition.is_multi_step,
      steps,
      completedSteps,
      currentStepNumber,
      formValues,
      onSubmit,
    ]
  );

  // Handle draft save
  const handleSaveDraft = useCallback(
    (event: React.MouseEvent) => {
      event.preventDefault(); // Prevent any default behavior
      onSubmit(formValues, true); // true = save as draft
    },
    [formValues, onSubmit]
  );

  const isFirstStep = currentStepNumber === 1;
  const isLastStep = currentStepNumber === steps.length;
  const currentStepData = steps.find(
    (step) => step.step_number === currentStepNumber
  );

  return (
    <div className="max-w-4xl mx-auto p-6 bg-white rounded-lg shadow-lg">
      {/* Form Header */}
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-gray-900 mb-2">
          {form_definition.name}
        </h1>
        {form_definition.description && (
          <p className="text-gray-600">{form_definition.description}</p>
        )}
      </div>

      {/* Progress Bar for Multi-step Forms */}
      {form_definition.is_multi_step && (
        <div className="mb-8">
          <ProgressBar percentage={completion_percentage} />
          <StepNavigation
            steps={steps}
            currentStep={currentStepNumber}
            onStepChange={handleStepChange}
            completedSteps={completedSteps}
          />
        </div>
      )}

      {/* Current Step Content */}
      {currentStepData && (
        <div className="mb-8">
          <h2 className="text-2xl font-semibold text-gray-800 mb-2">
            {currentStepData.name}
          </h2>
          {currentStepData.description && (
            <p className="text-gray-600 mb-6">{currentStepData.description}</p>
          )}
        </div>
      )}

      {/* Main Form with preventDefault */}
      <form ref={formRef} onSubmit={handleFormSubmit} noValidate>
        {/* Form Fields */}
        <div className="space-y-6 mb-8">
          {currentStepFields.map((field) => (
            <FieldRenderer
              key={field.id}
              field={field}
              value={formValues[field.field_name]}
              onChange={(value) => handleFieldChange(field.field_name, value)}
              errors={errors[field.field_name] || []}
              formValues={formValues}
            />
          ))}
        </div>

        {/* Form Actions */}
        <div className="flex justify-between items-center pt-6 border-t border-gray-200">
          <div>
            {!isFirstStep && (
              <button
                type="button" // Important: type="button" to prevent form submission
                onClick={handlePreviousStep}
                className="inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
                disabled={isLoading}
              >
                <ChevronLeft className="w-4 h-4 mr-1" />
                Previous
              </button>
            )}
          </div>

          <div className="flex space-x-3">
            {/* Draft Save Button */}
            <button
              type="button" // Important: type="button" to prevent form submission
              onClick={handleSaveDraft}
              className="inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
              disabled={isLoading}
            >
              Save Draft
            </button>

            {/* Next/Submit Button */}
            {!isLastStep ? (
              <button
                type="button" // Important: type="button" for next step
                onClick={handleNextStep}
                className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
                disabled={isLoading}
              >
                Next
                <ChevronRight className="w-4 h-4 ml-1" />
              </button>
            ) : (
              <button
                type="submit" // This will trigger form submission
                className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md text-white bg-green-600 hover:bg-green-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-green-500"
                disabled={isLoading}
              >
                {isLoading ? "Submitting..." : "Submit Form"}
              </button>
            )}
          </div>
        </div>
      </form>
    </div>
  );
};

export default FormRenderer;
