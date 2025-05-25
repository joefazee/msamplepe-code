// src/App.tsx - API Testing Interface (Updated)

import React, { useState, useEffect } from "react";
import { FormRenderer } from "./components/FormRenderer";
import type { FormData } from "./components/FormRenderer";
import { RefreshCw, Send, Save, Edit, List, Play } from "lucide-react";
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

  // API Helper Functions
  const makeApiCall = async (endpoint: string, options: RequestInit = {}) => {
    const url = `${apiConfig.baseUrl}${endpoint}`;
    const headers = {
      "Content-Type": "application/json",
      Authorization: `Bearer ${apiConfig.jwt}`,
      ...options.headers,
    };

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
  };

  // Load New Form
  const loadNewForm = async () => {
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
  };

  // Load Form with Progress
  const loadFormWithProgress = async () => {
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
  };

  // Load User Submissions
  const loadSubmissions = async () => {
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
  };

  // Load Form for Editing
  const loadFormForEdit = async (submissionId: string) => {
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
  };

  // Handle Form Submission - Updated to prevent any default behavior
  const handleFormSubmit = async (
    data: Record<string, any>,
    isDraft = false
  ) => {
    // Prevent any potential default behaviors
    setLoading(true);
    setError(null);
    setSuccess(null);

    try {
      let response;

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
            value.forEach((v) => formData.append(key, v));
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
            value.forEach((v) => formDataObj.append(key, v));
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
      }

      if (response && !response.ok) {
        const errorData = await response.text();
        throw new Error(`HTTP ${response.status}: ${errorData}`);
      }

      const result = response ? await response.json() : null;

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
      loadSubmissions();
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
  };

  // Handle Step Submission - Updated to prevent any default behavior
  const handleStepSubmit = async (
    stepNumber: number,
    data: Record<string, any>
  ) => {
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
          value.forEach((v) => formDataObj.append(key, v));
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
  };

  // Auto-load submissions on component mount
  useEffect(() => {
    if (apiConfig.jwt) {
      loadSubmissions();
    }
  }, [apiConfig.jwt]);

  return (
    <div className="min-h-screen bg-gray-50 py-8">
      <div className="max-w-6xl mx-auto px-4">
        {/* Header */}
        <div className="mb-8 text-center">
          <h1 className="text-4xl font-bold text-gray-900 mb-4">Monieverse</h1>
          <p className="text-lg text-gray-600">
            Simple is better than complex - Tim Peter
          </p>
        </div>

        {/* API Configuration */}
        <div className="bg-white rounded-lg shadow-lg p-6 mb-6">
          <h2 className="text-2xl font-semibold text-gray-800 mb-4">
            API Configuration
          </h2>
          <div className="grid md:grid-cols-2 gap-4">
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
                className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
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
                className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                placeholder="Your JWT token"
              />
            </div>
          </div>
        </div>

        {/* Form Type and Actions */}
        <div className="bg-white rounded-lg shadow-lg p-6 mb-6">
          <h2 className="text-2xl font-semibold text-gray-800 mb-4">
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
              className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
              placeholder="kyc, kyb, contact, etc."
            />
          </div>

          <div className="flex flex-wrap gap-3 mb-4">
            <button
              onClick={loadNewForm}
              disabled={loading || !apiConfig.jwt || !formType.trim()}
              className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <RefreshCw className="w-4 h-4 mr-2" />
              Load New Form
            </button>

            <button
              onClick={loadFormWithProgress}
              disabled={loading || !apiConfig.jwt || !formType.trim()}
              className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md text-white bg-green-600 hover:bg-green-700 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <Play className="w-4 h-4 mr-2" />
              Continue Form
            </button>

            <button
              onClick={loadSubmissions}
              disabled={loading || !apiConfig.jwt}
              className="inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
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
          <div className="bg-red-50 border border-red-200 rounded-md p-4 mb-6">
            <div className="flex">
              <div className="flex-shrink-0">
                <svg
                  className="h-5 w-5 text-red-400"
                  viewBox="0 0 20 20"
                  fill="currentColor"
                >
                  <path
                    fillRule="evenodd"
                    d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z"
                    clipRule="evenodd"
                  />
                </svg>
              </div>
              <div className="ml-3">
                <h3 className="text-sm font-medium text-red-800">Error</h3>
                <div className="mt-2 text-sm text-red-700">
                  <p>{error}</p>
                </div>
              </div>
            </div>
          </div>
        )}

        {success && (
          <div className="bg-green-50 border border-green-200 rounded-md p-4 mb-6">
            <div className="flex">
              <div className="flex-shrink-0">
                <svg
                  className="h-5 w-5 text-green-400"
                  viewBox="0 0 20 20"
                  fill="currentColor"
                >
                  <path
                    fillRule="evenodd"
                    d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z"
                    clipRule="evenodd"
                  />
                </svg>
              </div>
              <div className="ml-3">
                <h3 className="text-sm font-medium text-green-800">Success</h3>
                <div className="mt-2 text-sm text-green-700">
                  <p>{success}</p>
                </div>
              </div>
            </div>
          </div>
        )}

        {/* Submissions List */}
        {submissions.length > 0 && (
          <div className="bg-white rounded-lg shadow-lg p-6 mb-6">
            <h2 className="text-2xl font-semibold text-gray-800 mb-4">
              Your Submissions
            </h2>
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
        <div className="mt-12 bg-white rounded-lg shadow-lg p-6">
          <h2 className="text-2xl font-bold text-gray-900 mb-4">
            Testing Scenarios
          </h2>
          <div className="grid md:grid-cols-2 gap-6">
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
