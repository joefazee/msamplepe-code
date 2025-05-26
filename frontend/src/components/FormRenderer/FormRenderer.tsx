// src/components/FormRenderer/FormRenderer.tsx (Updated with fixes)

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
  Download,
  Trash2,
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

// ==================== Existing File Component ====================
const ExistingFileDisplay: React.FC<{
  files: any[];
  onDelete: (fileId: string) => void;
  isLoading?: boolean;
}> = ({ files, onDelete, isLoading = false }) => {
  if (!files || files.length === 0) return null;

  const formatFileSize = (bytes: number) => {
    if (bytes === 0) return "0 Bytes";
    const k = 1024;
    const sizes = ["Bytes", "KB", "MB", "GB"];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
  };

  return (
    <div className="mt-2 space-y-2">
      <p className="text-sm font-medium text-gray-700">Existing Files:</p>
      {files.map((file, index) => (
        <div
          key={file.id || index}
          className="flex items-center justify-between p-3 bg-blue-50 border border-blue-200 rounded-md"
        >
          <div className="flex items-center min-w-0 flex-1">
            <CheckCircle className="w-4 h-4 text-blue-500 mr-2 flex-shrink-0" />
            <div className="min-w-0 flex-1">
              <p className="text-sm font-medium text-blue-900 truncate">
                {file.file_name || file.name}
              </p>
              {file.file_size && (
                <p className="text-xs text-blue-600">
                  {formatFileSize(file.file_size)}
                </p>
              )}
            </div>
          </div>
          <div className="flex items-center space-x-2 ml-2">
            {file.url && (
              <a
                href={file.url}
                target="_blank"
                rel="noopener noreferrer"
                className="p-1 text-blue-600 hover:text-blue-800 hover:bg-blue-100 rounded transition-colors"
                title="Download file"
              >
                <Download className="w-4 h-4" />
              </a>
            )}
            <button
              type="button"
              onClick={() => onDelete(file.id)}
              disabled={isLoading}
              className="p-1 text-red-600 hover:text-red-800 hover:bg-red-100 rounded transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
              title="Delete file"
            >
              <Trash2 className="w-4 h-4" />
            </button>
          </div>
        </div>
      ))}
    </div>
  );
};

// ==================== Enhanced Field Components ====================
const TextInput: React.FC<{
  field: FormField;
  value: string;
  onChange: (value: string) => void;
  errors: string[];
}> = ({ field, value, onChange, errors }) => {
  const hasError = errors.length > 0;

  return (
    <div className="mb-4 w-full">
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
        className={`w-full px-3 py-2 border rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition-colors ${
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
          <AlertCircle className="w-4 h-4 mr-1 flex-shrink-0" />
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
    <div className="mb-4 w-full">
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
        className={`w-full px-3 py-2 border rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 resize-vertical transition-colors ${
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
          <AlertCircle className="w-4 h-4 mr-1 flex-shrink-0" />
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
    <div className="mb-4 w-full">
      <label className="block text-sm font-medium text-gray-700 mb-1">
        {getLocalizedText(field.label)}
        {field.is_required && <span className="text-red-500 ml-1">*</span>}
      </label>
      <select
        name={field.field_name}
        value={value || ""}
        onChange={(e) => onChange(e.target.value)}
        disabled={field.is_readonly}
        className={`w-full px-3 py-2 border rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition-colors ${
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
          <AlertCircle className="w-4 h-4 mr-1 flex-shrink-0" />
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
    <div className="mb-4 w-full">
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
          <AlertCircle className="w-4 h-4 mr-1 flex-shrink-0" />
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
    <div className="mb-4 w-full">
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
          <AlertCircle className="w-4 h-4 mr-1 flex-shrink-0" />
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
  existingFiles?: any[];
  onDeleteExistingFile?: (fileId: string) => void;
  isLoading?: boolean;
}> = ({
  field,
  value,
  onChange,
  errors,
  existingFiles = [],
  onDeleteExistingFile,
  isLoading = false,
}) => {
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

  const handleDeleteExisting = (fileId: string) => {
    if (onDeleteExistingFile) {
      onDeleteExistingFile(fileId);
    }
  };

  return (
    <div className="mb-4 w-full">
      <label className="block text-sm font-medium text-gray-700 mb-1">
        {getLocalizedText(field.label)}
        {field.is_required && <span className="text-red-500 ml-1">*</span>}
      </label>

      {/* Existing Files Display */}
      {existingFiles.length > 0 && (
        <ExistingFileDisplay
          files={existingFiles}
          onDelete={handleDeleteExisting}
          isLoading={isLoading}
        />
      )}

      {/* File Upload Area */}
      <div
        className={`border-2 border-dashed rounded-lg p-4 md:p-6 transition-colors ${
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
          <Upload className="mx-auto h-8 w-8 md:h-12 md:w-12 text-gray-400" />
          <div className="mt-2 md:mt-4">
            <label
              htmlFor={`file-${field.field_name}`}
              className="cursor-pointer rounded-md bg-white font-medium text-blue-600 focus-within:outline-none focus-within:ring-2 focus-within:ring-blue-500 focus-within:ring-offset-2 hover:text-blue-500"
            >
              <span className="text-sm md:text-base">Upload a file</span>
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
            <p className="pl-1 inline text-sm md:text-base">or drag and drop</p>
          </div>
          <p className="text-xs text-gray-500 mt-2">
            {allowedTypes.length > 0 && `${allowedTypes.join(", ")} `}
            up to {formatFileSize(maxSize)}
            {maxFiles > 1 && ` (max ${maxFiles} files)`}
          </p>
        </div>
      </div>

      {/* New Files Display */}
      {value && value.length > 0 && (
        <div className="mt-2 space-y-2">
          {Array.from(value).map((file, index) => (
            <div
              key={index}
              className="flex items-center justify-between p-2 bg-green-50 border border-green-200 rounded"
            >
              <div className="flex items-center min-w-0 flex-1">
                <CheckCircle className="w-4 h-4 text-green-500 mr-2 flex-shrink-0" />
                <div className="min-w-0">
                  <span className="text-sm text-gray-700 truncate block">
                    {file.name}
                  </span>
                  <span className="text-xs text-gray-500">
                    ({formatFileSize(file.size)})
                  </span>
                </div>
              </div>
              <button
                type="button"
                onClick={() => onChange(null)}
                className="ml-2 text-red-500 hover:text-red-700 p-1 rounded hover:bg-red-50 transition-colors"
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
          <AlertCircle className="w-4 h-4 mr-1 flex-shrink-0" />
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
    <div className="mb-4 w-full">
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
          className={`w-full px-3 py-2 border rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition-colors ${
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
          <AlertCircle className="w-4 h-4 mr-1 flex-shrink-0" />
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
    <div className="mb-4 w-full">
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
        className={`w-full px-3 py-2 border rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition-colors ${
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
          <AlertCircle className="w-4 h-4 mr-1 flex-shrink-0" />
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
  existingFiles?: any[];
  onDeleteExistingFile?: (fileId: string) => void;
  isLoading?: boolean;
}> = ({
  field,
  value,
  onChange,
  errors,
  formValues,
  existingFiles,
  onDeleteExistingFile,
  isLoading,
}) => {
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
          existingFiles={existingFiles}
          onDeleteExistingFile={onDeleteExistingFile}
          isLoading={isLoading}
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
  <div className="mb-8">
    {/* Mobile Step Navigation */}
    <div className="block md:hidden">
      <div className="flex items-center space-x-2 mb-4">
        <span className="text-sm font-medium text-gray-700">Step</span>
        <span className="text-sm font-bold text-blue-600">{currentStep}</span>
        <span className="text-sm text-gray-500">of {steps.length}</span>
      </div>
      <div className="text-sm font-medium text-gray-900">
        {steps.find((s) => s.step_number === currentStep)?.name}
      </div>
    </div>

    {/* Desktop Step Navigation */}
    <div className="hidden md:flex items-center justify-between">
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
                className={`mt-2 text-xs text-center max-w-20 ${
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

  // Form reference for preventing default submission
  const formRef = useRef<HTMLFormElement>(null);

  // State for existing files by field name
  const [existingFiles, setExistingFiles] = useState<Record<string, any[]>>({});

  // Initialize form values from existing data
  const [formValues, setFormValues] = useState<Record<string, any>>(() => {
    const values: Record<string, any> = {};

    // Set default values
    fields.forEach((field) => {
      if (field.default_value) {
        try {
          // Try to parse JSON for complex defaults (like existing files)
          const parsed = JSON.parse(field.default_value);

          // Handle file fields differently
          if (
            (field.field_type === "file" || field.field_type === "files") &&
            Array.isArray(parsed)
          ) {
            setExistingFiles((prev) => ({
              ...prev,
              [field.field_name]: parsed,
            }));
          } else {
            values[field.field_name] = parsed;
          }
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

  // Handle existing file deletion
  const handleDeleteExistingFile = useCallback(
    (fieldName: string, fileId: string) => {
      setExistingFiles((prev) => ({
        ...prev,
        [fieldName]:
          prev[fieldName]?.filter((file) => file.id !== fileId) || [],
      }));
      // You might want to call an API endpoint here to actually delete the file
      // For now, we'll just remove it from the UI
    },
    []
  );

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

  // Handle form submission with STRICT preventDefault
  const handleFormSubmit = useCallback(
    (event: FormEvent<HTMLFormElement>) => {
      // CRITICAL: Prevent all default form behavior
      event.preventDefault();
      event.stopPropagation();

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

  // Handle draft save with STRICT preventDefault
  const handleSaveDraft = useCallback(
    (event: React.MouseEvent<HTMLButtonElement>) => {
      // CRITICAL: Prevent all default behavior
      event.preventDefault();
      event.stopPropagation();

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
    <div className="w-full max-w-4xl mx-auto p-4 md:p-6 bg-white rounded-lg shadow-lg">
      {/* Form Header */}
      <div className="mb-6 md:mb-8">
        <h1 className="text-2xl md:text-3xl font-bold text-gray-900 mb-2">
          {form_definition.name}
        </h1>
        {form_definition.description && (
          <p className="text-gray-600 text-sm md:text-base">
            {form_definition.description}
          </p>
        )}
      </div>

      {/* Progress Bar for Multi-step Forms */}
      {form_definition.is_multi_step && (
        <div className="mb-6 md:mb-8">
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
        <div className="mb-6 md:mb-8">
          <h2 className="text-xl md:text-2xl font-semibold text-gray-800 mb-2">
            {currentStepData.name}
          </h2>
          {currentStepData.description && (
            <p className="text-gray-600 text-sm md:text-base mb-4 md:mb-6">
              {currentStepData.description}
            </p>
          )}
        </div>
      )}

      {/* Main Form with STRICT preventDefault */}
      <form
        ref={formRef}
        onSubmit={handleFormSubmit}
        noValidate
        // Add additional safeguards
        onReset={(e) => e.preventDefault()}
      >
        {/* Form Fields */}
        <div className="space-y-4 md:space-y-6 mb-6 md:mb-8">
          {currentStepFields.map((field) => (
            <FieldRenderer
              key={field.id}
              field={field}
              value={formValues[field.field_name]}
              onChange={(value) => handleFieldChange(field.field_name, value)}
              errors={errors[field.field_name] || []}
              formValues={formValues}
              existingFiles={existingFiles[field.field_name]}
              onDeleteExistingFile={(fileId) =>
                handleDeleteExistingFile(field.field_name, fileId)
              }
              isLoading={isLoading}
            />
          ))}
        </div>

        {/* Form Actions */}
        <div className="flex flex-col sm:flex-row sm:justify-between sm:items-center pt-6 border-t border-gray-200 space-y-3 sm:space-y-0">
          <div>
            {!isFirstStep && (
              <button
                type="button" // CRITICAL: Prevent form submission
                onClick={handlePreviousStep}
                className="inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                disabled={isLoading}
              >
                <ChevronLeft className="w-4 h-4 mr-1" />
                Previous
              </button>
            )}
          </div>

          <div className="flex flex-col sm:flex-row space-y-3 sm:space-y-0 sm:space-x-3">
            {/* Draft Save Button */}
            <button
              type="button" // CRITICAL: Prevent form submission
              onClick={handleSaveDraft}
              className="w-full sm:w-auto inline-flex items-center justify-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
              disabled={isLoading}
            >
              {isLoading ? "Saving..." : "Save Draft"}
            </button>

            {/* Next/Submit Button */}
            {!isLastStep ? (
              <button
                type="button" // CRITICAL: Prevent form submission for next step
                onClick={handleNextStep}
                className="w-full sm:w-auto inline-flex items-center justify-center px-4 py-2 border border-transparent text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                disabled={isLoading}
              >
                Next
                <ChevronRight className="w-4 h-4 ml-1" />
              </button>
            ) : (
              <button
                type="submit" // This WILL trigger form submission
                className="w-full sm:w-auto inline-flex items-center justify-center px-4 py-2 border border-transparent text-sm font-medium rounded-md text-white bg-green-600 hover:bg-green-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-green-500 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
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
