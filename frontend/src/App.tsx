// src/App.tsx - Enhanced API Testing Interface

import React, { useState, useEffect, useCallback } from "react";
import { FormRenderer } from "./components/FormRenderer";
import type { FormData } from "./components/FormRenderer";
import {
  RefreshCw,
  Send,
  Save,
  Edit,
  List,
  Play,
  AlertCircle,
  CheckCircle,
  X,
} from "lucide-react";
import "./App.css";

interface ApiConfig {
  baseUrl: string;
  jwt: string;
}

interface Submission {
  id: string;
  form_definition_id: string;
  status: string;
  created_at: string;
  updated_at: string;
  completion_percentage?: number;
  current_step_number?: number;
}

function App() {
  // API Configuration
  const [apiConfig, setApiConfig] = useState<ApiConfig>({
    baseUrl: "http://localhost:3000/api/v1",
    jwt: "",
  });

  // Form State
  const [formType, setFormType] = useState("");
  const [formData, setFormData] = useState<FormData | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  // Submissions State
  const [submissions, setSubmissions] = useState<Submission[]>([]);
  const [selectedSubmission, setSelectedSubmission] = useState<string | null>(
    null
  );

  // Testing Mode
  const [testingMode, setTestingMode] = useState<"new" | "edit" | "continue">(
    "new"
  );

  // Mobile responsive state
  const [isMobile, setIsMobile] = useState(false);

  // Check for mobile screen size
  useEffect(() => {
    const checkMobile = () => {
      setIsMobile(window.innerWidth < 768);
    };

    checkMobile();
    window.addEventListener("resize", checkMobile);

    return () => window.removeEventListener("resize", checkMobile);
  }, []);

  // Clear messages after timeout
  useEffect(() => {
    if (error || success) {
      const timer = setTimeout(() => {
        setError(null);
        setSuccess(null);
      }, 5000);

      return () => clearTimeout(timer);
    }
  }, [error, success]);

  // API Helper Functions
  const makeApiCall = useCallback(
    async (endpoint: string, options: RequestInit = {}) => {
      const url = `${apiConfig.baseUrl}${endpoint}`;
      const headers = {
        "Content-Type": "application/json",
        Authorization: `Bearer ${apiConfig.jwt}`,
        ...options.headers,
      };

      // Remove Content-Type for FormData requests
      if (options.body instanceof FormData) {
        delete headers["Content-Type"];
      }

      try {
        const response = await fetch(url, {
          ...options,
          headers,
        });

        if (!response.ok) {
          const errorData = await response.text();
          throw new Error(`HTTP ${response.status}: ${errorData}`);
        }

        const data = await response.json();
        return data;
      } catch (err) {
        console.error("API Call failed:", err);
        throw err;
      }
    },
    [apiConfig]
  );

  // Load New Form
  const loadNewForm = useCallback(async () => {
    if (!formType.trim()) {
      setError("Please enter a form type");
      return;
    }

    setLoading(true);
    setError(null);
    setSuccess(null);

    try {
      const response = await makeApiCall(
        `/forms/?type=${encodeURIComponent(formType)}`
      );
      setFormData(response.data);
      setSuccess(`Form "${formType}" loaded successfully`);
      setTestingMode("new");
      setSelectedSubmission(null);
    } catch (err) {
      setError(
        `Failed to load form: ${
          err instanceof Error ? err.message : "Unknown error"
        }`
      );
    } finally {
      setLoading(false);
    }
  }, [formType, makeApiCall]);

  // Load Form with Progress
  const loadFormWithProgress = useCallback(async () => {
    if (!formType.trim()) {
      setError("Please enter a form type");
      return;
    }

    setLoading(true);
    setError(null);
    setSuccess(null);

    try {
      const response = await makeApiCall(
        `/forms/progress?type=${encodeURIComponent(formType)}`
      );
      setFormData(response.data);
      setSuccess(`Form with progress "${formType}" loaded successfully`);
      setTestingMode("continue");
    } catch (err) {
      setError(
        `Failed to load form with progress: ${
          err instanceof Error ? err.message : "Unknown error"
        }`
      );
    } finally {
      setLoading(false);
    }
  }, [formType, makeApiCall]);

  // Load User Submissions
  const loadSubmissions = useCallback(async () => {
    if (!apiConfig.jwt) return;

    setLoading(true);
    setError(null);

    try {
      const response = await makeApiCall("/forms/submissions");
      setSubmissions(response.data || []);
      setSuccess(`Loaded ${response.data?.length || 0} submissions`);
    } catch (err) {
      setError(
        `Failed to load submissions: ${
          err instanceof Error ? err.message : "Unknown error"
        }`
      );
    } finally {
      setLoading(false);
    }
  }, [makeApiCall, apiConfig.jwt]);

  // Load Form for Editing
  const loadFormForEdit = useCallback(
    async (submissionId: string) => {
      setLoading(true);
      setError(null);
      setSuccess(null);

      try {
        const response = await makeApiCall(
          `/forms/submissions/${submissionId}/edit`
        );
        setFormData(response.data);
        setSelectedSubmission(submissionId);
        setTestingMode("edit");
        setSuccess(`Form loaded for editing`);
      } catch (err) {
        setError(
          `Failed to load form for editing: ${
            err instanceof Error ? err.message : "Unknown error"
          }`
        );
      } finally {
        setLoading(false);
      }
    },
    [makeApiCall]
  );

  // Handle Form Submission - Enhanced with better error handling
  const handleFormSubmit = useCallback(
    async (data: Record<string, any>, isDraft = false) => {
      setLoading(true);
      setError(null);
      setSuccess(null);

      try {
        let response: Response;

        if (testingMode === "edit" && selectedSubmission) {
          // Update existing submission
          const formData = new FormData();

          // Add form fields
          Object.entries(data).forEach(([key, value]) => {
            if (value instanceof FileList) {
              Array.from(value).forEach((file) => {
                formData.append(key, file);
              });
            } else if (Array.isArray(value)) {
              value.forEach((v) => formData.append(key, String(v)));
            } else if (value !== null && value !== undefined) {
              formData.append(key, String(value));
            }
          });

          // Add meta fields
          formData.append("_status", isDraft ? "draft" : "submitted");
          formData.append("_partial", isDraft ? "true" : "false");

          response = await fetch(
            `${apiConfig.baseUrl}/forms/submissions/${selectedSubmission}`,
            {
              method: "PUT",
              headers: {
                Authorization: `Bearer ${apiConfig.jwt}`,
              },
              body: formData,
            }
          );
        } else if (testingMode === "new" && formData) {
          // Create new submission
          const formDataObj = new FormData();

          // Add form fields
          Object.entries(data).forEach(([key, value]) => {
            if (value instanceof FileList) {
              Array.from(value).forEach((file) => {
                formDataObj.append(key, file);
              });
            } else if (Array.isArray(value)) {
              value.forEach((v) => formDataObj.append(key, String(v)));
            } else if (value !== null && value !== undefined) {
              formDataObj.append(key, String(value));
            }
          });

          // Add meta fields
          formDataObj.append("_status", isDraft ? "draft" : "submitted");

          const endpoint = isDraft
            ? `/forms/${formData.form_definition.id}/draft`
            : `/forms/${formData.form_definition.id}/submit`;

          response = await fetch(`${apiConfig.baseUrl}${endpoint}`, {
            method: "POST",
            headers: {
              Authorization: `Bearer ${apiConfig.jwt}`,
            },
            body: formDataObj,
          });
        } else {
          throw new Error("Invalid form state");
        }

        if (!response.ok) {
          const errorData = await response.text();
          throw new Error(`HTTP ${response.status}: ${errorData}`);
        }

        const result = await response.json();

        setSuccess(
          isDraft
            ? `Form saved as draft successfully! ${
                testingMode === "new"
                  ? "New submission created."
                  : "Existing submission updated."
              }`
            : `Form submitted successfully! ${
                testingMode === "new"
                  ? "New submission created."
                  : "Existing submission updated."
              }`
        );

        console.log("Submission result:", result);

        // Reload submissions list
        await loadSubmissions();
      } catch (err) {
        setError(
          `Failed to ${isDraft ? "save draft" : "submit form"}: ${
            err instanceof Error ? err.message : "Unknown error"
          }`
        );
        console.error("Submission error:", err);
      } finally {
        setLoading(false);
      }
    },
    [testingMode, selectedSubmission, formData, apiConfig, loadSubmissions]
  );

  // Handle Step Submission - Enhanced with better error handling
  const handleStepSubmit = useCallback(
    async (stepNumber: number, data: Record<string, any>) => {
      if (!formData || !selectedSubmission) return;

      try {
        const formDataObj = new FormData();

        // Add step data
        Object.entries(data).forEach(([key, value]) => {
          if (value instanceof FileList) {
            Array.from(value).forEach((file) => {
              formDataObj.append(key, file);
            });
          } else if (Array.isArray(value)) {
            value.forEach((v) => formDataObj.append(key, String(v)));
          } else if (value !== null && value !== undefined) {
            formDataObj.append(key, String(value));
          }
        });

        formDataObj.append("_status", "in_progress");

        const response = await fetch(
          `${apiConfig.baseUrl}/forms/submissions/${selectedSubmission}/steps/${stepNumber}`,
          {
            method: "PUT",
            headers: {
              Authorization: `Bearer ${apiConfig.jwt}`,
            },
            body: formDataObj,
          }
        );

        if (!response.ok) {
          const errorData = await response.text();
          throw new Error(`HTTP ${response.status}: ${errorData}`);
        }

        const result = await response.json();
        console.log(`Step ${stepNumber} saved:`, result);

        setSuccess(`Step ${stepNumber} progress saved successfully`);
      } catch (err) {
        setError(
          `Failed to save step progress: ${
            err instanceof Error ? err.message : "Unknown error"
          }`
        );
        console.error("Step submission error:", err);
      }
    },
    [formData, selectedSubmission, apiConfig]
  );

  // Auto-load submissions on component mount
  useEffect(() => {
    if (apiConfig.jwt) {
      loadSubmissions();
    }
  }, [apiConfig.jwt, loadSubmissions]);

  // Message component for better mobile display
  const MessageAlert: React.FC<{
    type: "error" | "success";
    message: string;
    onClose: () => void;
  }> = ({ type, message, onClose }) => (
    <div
      className={`rounded-md p-4 mb-6 ${
        type === "error"
          ? "bg-red-50 border border-red-200"
          : "bg-green-50 border border-green-200"
      }`}
    >
      <div className="flex">
        <div className="flex-shrink-0">
          {type === "error" ? (
            <AlertCircle className="h-5 w-5 text-red-400" />
          ) : (
            <CheckCircle className="h-5 w-5 text-green-400" />
          )}
        </div>
        <div className="ml-3 flex-1">
          <h3
            className={`text-sm font-medium ${
              type === "error" ? "text-red-800" : "text-green-800"
            }`}
          >
            {type === "error" ? "Error" : "Success"}
          </h3>
          <div
            className={`mt-2 text-sm ${
              type === "error" ? "text-red-700" : "text-green-700"
            }`}
          >
            <p>{message}</p>
          </div>
        </div>
        <div className="ml-auto pl-3">
          <button
            type="button"
            onClick={onClose}
            className={`inline-flex rounded-md p-1.5 text-sm font-medium focus:outline-none focus:ring-2 focus:ring-offset-2 ${
              type === "error"
                ? "text-red-500 hover:bg-red-100 focus:ring-red-500"
                : "text-green-500 hover:bg-green-100 focus:ring-green-500"
            }`}
          >
            <X className="h-4 w-4" />
          </button>
        </div>
      </div>
    </div>
  );

  return (
    <div className="min-h-screen bg-gray-50 py-4 md:py-8 prevent-scroll">
      <div className="max-w-6xl mx-auto px-4">
        {/* Header */}
        <div className="mb-6 md:mb-8 text-center">
          <h1 className="text-3xl md:text-4xl font-bold text-gray-900 mb-4">
            Monieverse
          </h1>
          <p className="text-base md:text-lg text-gray-600">
            Simple is better than complex - Tim Peter
          </p>
        </div>

        {/* API Configuration */}
        <div className="bg-white rounded-lg shadow-lg p-4 md:p-6 mb-6">
          <h2 className="text-xl md:text-2xl font-semibold text-gray-800 mb-4">
            API Configuration
          </h2>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Base URL
              </label>
              <input
                type="text"
                value={apiConfig.baseUrl}
                onChange={(e) =>
                  setApiConfig((prev) => ({ ...prev, baseUrl: e.target.value }))
                }
                className="w-full mobile-text"
                placeholder="http://localhost:3000/api/v1"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                JWT Token
              </label>
              <input
                type="password"
                value={apiConfig.jwt}
                onChange={(e) =>
                  setApiConfig((prev) => ({ ...prev, jwt: e.target.value }))
                }
                className="w-full mobile-text"
                placeholder="Your JWT token"
              />
            </div>
          </div>
        </div>

        {/* Form Type and Actions */}
        <div className="bg-white rounded-lg shadow-lg p-4 md:p-6 mb-6">
          <h2 className="text-xl md:text-2xl font-semibold text-gray-800 mb-4">
            Form Testing
          </h2>

          <div className="mb-4">
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Form Type
            </label>
            <input
              type="text"
              value={formType}
              onChange={(e) => setFormType(e.target.value)}
              className="w-full mobile-text"
              placeholder="kyc, kyb, contact, etc."
            />
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3 mb-4">
            <button
              onClick={loadNewForm}
              disabled={loading || !apiConfig.jwt || !formType.trim()}
              className="touch-target bg-blue-600 text-white hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <RefreshCw className="w-4 h-4 mr-2" />
              Load New Form
            </button>

            <button
              onClick={loadFormWithProgress}
              disabled={loading || !apiConfig.jwt || !formType.trim()}
              className="touch-target bg-green-600 text-white hover:bg-green-700 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <Play className="w-4 h-4 mr-2" />
              Continue Form
            </button>

            <button
              onClick={loadSubmissions}
              disabled={loading || !apiConfig.jwt}
              className="touch-target bg-gray-600 text-white hover:bg-gray-700 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <List className="w-4 h-4 mr-2" />
              Load Submissions
            </button>
          </div>

          {/* Testing Mode Indicator */}
          {formData && (
            <div className="mb-4">
              <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800">
                Mode:{" "}
                {testingMode.charAt(0).toUpperCase() + testingMode.slice(1)}
                {selectedSubmission &&
                  ` (ID: ${selectedSubmission.slice(0, 8)}...)`}
              </span>
            </div>
          )}
        </div>

        {/* Status Messages */}
        {error && (
          <MessageAlert
            type="error"
            message={error}
            onClose={() => setError(null)}
          />
        )}

        {success && (
          <MessageAlert
            type="success"
            message={success}
            onClose={() => setSuccess(null)}
          />
        )}

        {/* Submissions List */}
        {submissions.length > 0 && (
          <div className="bg-white rounded-lg shadow-lg p-4 md:p-6 mb-6">
            <h2 className="text-xl md:text-2xl font-semibold text-gray-800 mb-4">
              Your Submissions
            </h2>

            {/* Mobile-friendly submissions display */}
            {isMobile ? (
              <div className="space-y-4">
                {submissions.map((submission) => (
                  <div
                    key={submission.id}
                    className="border border-gray-200 rounded-lg p-4"
                  >
                    <div className="flex items-center justify-between mb-2">
                      <span className="text-sm font-medium text-gray-900">
                        {submission.id.slice(0, 8)}...
                      </span>
                      <span
                        className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${
                          submission.status === "submitted"
                            ? "bg-green-100 text-green-800"
                            : submission.status === "draft"
                            ? "bg-yellow-100 text-yellow-800"
                            : "bg-gray-100 text-gray-800"
                        }`}
                      >
                        {submission.status}
                      </span>
                    </div>
                    <div className="text-sm text-gray-500 mb-2">
                      Progress: {submission.completion_percentage || 0}%
                      {submission.current_step_number &&
                        ` (Step ${submission.current_step_number})`}
                    </div>
                    <div className="text-sm text-gray-500 mb-3">
                      Updated:{" "}
                      {new Date(submission.updated_at).toLocaleDateString()}
                    </div>
                    <button
                      onClick={() => loadFormForEdit(submission.id)}
                      disabled={loading}
                      className="w-full touch-target bg-blue-600 text-white hover:bg-blue-700 disabled:opacity-50"
                    >
                      <Edit className="w-4 h-4 mr-1" />
                      Edit
                    </button>
                  </div>
                ))}
              </div>
            ) : (
              <div className="overflow-x-auto">
                <table className="min-w-full divide-y divide-gray-200">
                  <thead className="bg-gray-50">
                    <tr>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        ID
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Status
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Progress
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Updated
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Actions
                      </th>
                    </tr>
                  </thead>
                  <tbody className="bg-white divide-y divide-gray-200">
                    {submissions.map((submission) => (
                      <tr key={submission.id}>
                        <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                          {submission.id.slice(0, 8)}...
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap">
                          <span
                            className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${
                              submission.status === "submitted"
                                ? "bg-green-100 text-green-800"
                                : submission.status === "draft"
                                ? "bg-yellow-100 text-yellow-800"
                                : "bg-gray-100 text-gray-800"
                            }`}
                          >
                            {submission.status}
                          </span>
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                          {submission.completion_percentage || 0}%
                          {submission.current_step_number &&
                            ` (Step ${submission.current_step_number})`}
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                          {new Date(submission.updated_at).toLocaleDateString()}
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-sm font-medium">
                          <button
                            onClick={() => loadFormForEdit(submission.id)}
                            disabled={loading}
                            className="text-blue-600 hover:text-blue-900 disabled:opacity-50"
                          >
                            <Edit className="w-4 h-4 inline mr-1" />
                            Edit
                          </button>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </div>
        )}

        {/* Loading Indicator */}
        {loading && (
          <div className="flex justify-center items-center py-8">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
            <span className="ml-2 text-gray-600">Loading...</span>
          </div>
        )}

        {/* Form Renderer */}
        {formData && !loading && (
          <FormRenderer
            formData={formData}
            onSubmit={handleFormSubmit}
            onStepSubmit={handleStepSubmit}
            isLoading={loading}
          />
        )}

        {/* API Documentation */}
        <div className="mt-8 md:mt-12 bg-white rounded-lg shadow-lg p-4 md:p-6">
          <h2 className="text-xl md:text-2xl font-bold text-gray-900 mb-4">
            Testing Scenarios
          </h2>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div>
              <h3 className="font-semibold text-gray-800 mb-2">
                Available Actions
              </h3>
              <ul className="text-sm text-gray-600 space-y-1">
                <li>
                  • <strong>Load New Form:</strong> Fetch a fresh form by type
                </li>
                <li>
                  • <strong>Continue Form:</strong> Load form with existing
                  progress
                </li>
                <li>
                  • <strong>Load Submissions:</strong> View all your form
                  submissions
                </li>
                <li>
                  • <strong>Edit:</strong> Load existing submission for
                  modification
                </li>
                <li>
                  • <strong>Save Draft:</strong> Save form progress without
                  submitting
                </li>
                <li>
                  • <strong>Submit Form:</strong> Complete form submission
                </li>
              </ul>
            </div>
            <div>
              <h3 className="font-semibold text-gray-800 mb-2">
                API Endpoints Used
              </h3>
              <ul className="text-sm text-gray-600 space-y-1">
                <li>
                  • <code>GET /forms?type={"{type}"}</code>
                </li>
                <li>
                  • <code>GET /forms/progress?type={"{type}"}</code>
                </li>
                <li>
                  • <code>GET /forms/submissions</code>
                </li>
                <li>
                  • <code>GET /forms/submissions/{"{id}"}/edit</code>
                </li>
                <li>
                  • <code>POST /forms/{"{id}"}/draft</code>
                </li>
                <li>
                  • <code>POST /forms/{"{id}"}/submit</code>
                </li>
                <li>
                  • <code>PUT /forms/submissions/{"{id}"}</code>
                </li>
                <li>
                  •{" "}
                  <code>
                    PUT /forms/submissions/{"{id}"}/steps/{"{step}"}
                  </code>
                </li>
              </ul>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

export default App;
