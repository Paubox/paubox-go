package paubox

import (
	"encoding/json"
	"time"
)

// Form is a Paubox form as returned by the Forms API.
type Form struct {
	// ID is the form's UUID.
	ID string `json:"id"`

	// Title is the form's display title.
	Title string `json:"title"`

	// Description is the optional form description.
	Description *string `json:"description"`

	// FormHTML is the rendered HTML for the form, when present.
	FormHTML *string `json:"form_html"`

	// FormJSON is the arbitrary form-builder JSON definition.
	FormJSON json.RawMessage `json:"form_json"`

	// FormCSS is the custom CSS for the form, when present.
	FormCSS *string `json:"form_css"`

	// VanityURL is the optional vanity URL slug for the form.
	VanityURL *string `json:"vanity_url"`

	// Version is the form definition version.
	Version int `json:"version"`

	// Active reports whether the form accepts submissions.
	Active bool `json:"active"`

	// CustomerID is the owning Paubox customer ID.
	CustomerID int `json:"customer_id"`

	// OldFormID is the legacy form ID, when the form was migrated.
	OldFormID *int `json:"old_form_id"`

	// CreatedAt is when the form was created.
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when the form was last updated.
	UpdatedAt time.Time `json:"updated_at"`

	// Recipient is a comma-separated list of notification recipient email
	// addresses, when configured.
	Recipient *string `json:"recipient"`

	// Signable reports whether the form supports signatures.
	Signable bool `json:"signable"`

	// SignatureConfirmationLabel is the label shown next to the signature
	// confirmation checkbox, when configured.
	SignatureConfirmationLabel *string `json:"signature_confirmation_label"`

	// SubmissionCount is the number of submissions received.
	SubmissionCount int `json:"submission_count"`

	// Type is the form type (e.g. "marketing_form"), when set.
	Type *string `json:"type"`

	// SubscriptionListID is the marketing subscription list linked to the
	// form, when set.
	SubscriptionListID *string `json:"subscription_list_id"`

	// Deleted reports whether the form has been soft-deleted.
	Deleted bool `json:"deleted"`

	// Archived reports whether the form has been archived.
	Archived bool `json:"archived"`
}

// FormPageInfo describes pagination metadata for a [ListFormsResponse].
type FormPageInfo struct {
	// Count is the total number of matching forms.
	Count int64 `json:"count"`

	// Pages is the total number of pages.
	Pages int `json:"pages"`

	// Page is the current (1-based) page number.
	Page int `json:"page"`

	// Items is the page size used.
	Items int `json:"items"`
}

// ListFormsResponse is the response from FormsClient.ListForms.
type ListFormsResponse struct {
	// Results is the current page of forms.
	Results []Form `json:"results"`

	// PageInfo describes the pagination state.
	PageInfo FormPageInfo `json:"page_info"`
}

// ListFormsParams are the query params for FormsClient.ListForms. Zero
// values are omitted from the query string.
type ListFormsParams struct {
	// CustomerID filters by customer (query param customer_id).
	CustomerID int

	// FormID filters to a single form (query param form_id).
	FormID string

	// Search matches against title/description with a LIKE query
	// (query param search).
	Search string

	// Order is the sort direction: "asc" or "desc" (server default desc).
	Order string

	// OrderBy is the sort column, allowlisted server-side: title,
	// updated_at, submission_count; default created_at.
	OrderBy string

	// Archived filters by archived state (query param archived).
	Archived *bool

	// Active filters by active state (query param active).
	Active *bool

	// Page is the 1-based page number (query param page).
	Page int

	// Items is the page size (query param items). The server caps this at
	// 100 and defaults to 50.
	Items int
}

// CreateFormRequest is the request body for FormsClient.CreateForm.
type CreateFormRequest struct {
	// Title is the form title. Required.
	Title string `json:"title"`

	// CustomerID is the owning customer ID. Required.
	CustomerID int `json:"customer_id"`

	// FormJSON is the form-builder JSON definition. Required.
	FormJSON json.RawMessage `json:"form_json"`

	// Version is required by the API; the SDK defaults 0 to 1.
	Version int `json:"version"`

	// Description is the optional form description.
	Description string `json:"description,omitempty"`

	// FormHTML is the optional rendered HTML for the form.
	FormHTML string `json:"form_html,omitempty"`

	// FormCSS is the optional custom CSS for the form.
	FormCSS string `json:"form_css,omitempty"`

	// Recipient is a comma-separated list of notification recipients.
	Recipient string `json:"recipient,omitempty"`

	// Signable enables signature support on the form.
	Signable bool `json:"signable"`

	// SignatureConfirmationLabel is the label shown next to the signature
	// confirmation checkbox.
	SignatureConfirmationLabel string `json:"signature_confirmation_label,omitempty"`

	// SubscriptionListID links the form to a marketing subscription list.
	SubscriptionListID string `json:"subscription_list_id,omitempty"`

	// Type is the form type (e.g. "marketing_form").
	Type string `json:"type,omitempty"`

	// Active controls whether the form accepts submissions.
	Active bool `json:"active"`

	// SubmissionCount seeds the form's submission counter.
	SubmissionCount int `json:"submission_count"`
}

// CreateFormResponse is the response from FormsClient.CreateForm.
type CreateFormResponse struct {
	// ID is the new form's UUID.
	ID string `json:"id"`
}

// UpdateFormRequest is the request body for FormsClient.UpdateForm.
//
// PATCH-style semantics: a nil pointer field (or nil FormJSON) is omitted
// from the request and left unchanged on the server. Use [Ptr] to set
// pointer fields inline.
type UpdateFormRequest struct {
	// Title replaces the form title.
	Title *string `json:"title,omitempty"`

	// Description replaces the form description.
	Description *string `json:"description,omitempty"`

	// FormJSON replaces the form-builder JSON definition.
	FormJSON json.RawMessage `json:"form_json,omitempty"`

	// VanityURL replaces the vanity URL slug.
	VanityURL *string `json:"vanity_url,omitempty"`

	// Recipient replaces the comma-separated notification recipient list.
	Recipient *string `json:"recipient,omitempty"`

	// Active toggles whether the form accepts submissions.
	Active *bool `json:"active,omitempty"`

	// SubscriptionListID replaces the linked subscription list.
	SubscriptionListID *string `json:"subscription_list_id,omitempty"`
}

// UpdateFormResponse is the response from FormsClient.UpdateForm.
type UpdateFormResponse struct {
	// Detail is the human-readable result message.
	Detail string `json:"detail"`

	// FormID is the UUID of the updated form.
	FormID string `json:"form_id"`
}

// FormActionResponse is the response from FormsClient.ArchiveForm and
// FormsClient.UnarchiveForm.
type FormActionResponse struct {
	// Detail is the human-readable result message.
	Detail string `json:"detail"`
}

// CopyFormRequest is the request body for FormsClient.CopyForm.
type CopyFormRequest struct {
	// FormID is the UUID of the form to copy. Required.
	FormID string `json:"form_id"`

	// Title is the title for the copy. Required.
	Title string `json:"title"`
}

// FormStatsParams are the query params for FormsClient.GetFormStats.
type FormStatsParams struct {
	// CustomerID is optional; the server defaults to the API key's
	// customer (query param customer_id).
	CustomerID int
}

// FormStats is the response from FormsClient.GetFormStats.
type FormStats struct {
	// ActiveFormCount is the number of active forms.
	ActiveFormCount int64 `json:"active_form_count"`

	// TotalSubmissionCount is the all-time submission count.
	TotalSubmissionCount int64 `json:"total_submission_count"`

	// SubmissionsLast7Days is the submission count over the last 7 days.
	SubmissionsLast7Days int64 `json:"submissions_last_7_days"`
}

// FormSubmission is a single form submission as returned by the Forms API.
type FormSubmission struct {
	// ID is the submission's UUID.
	ID string `json:"id"`

	// FormID is the UUID of the form that was submitted.
	FormID string `json:"form_id"`

	// FormData is a JSON-encoded string of the submitted fields. The server
	// re-serializes the submission payload, so it arrives as a string, not
	// an object.
	FormData string `json:"form_data"`

	// StorageType identifies where the submission payload is stored.
	StorageType string `json:"storage_type"`

	// StorageURL is the storage location URL, when present.
	StorageURL *string `json:"storage_url"`

	// SubmitterEmail is the submitter's email address, when captured.
	SubmitterEmail *string `json:"submitter_email"`

	// Recipients is the comma-separated notification recipient list used
	// for this submission, when present.
	Recipients *string `json:"recipients"`

	// Attachment is the stored attachment reference, when present.
	Attachment *string `json:"attachment"`

	// AttachmentName is the attachment's filename, when present.
	AttachmentName *string `json:"attachment_name"`

	// AttachmentURL is the attachment's download URL, when present.
	AttachmentURL *string `json:"attachment_url"`

	// AttachmentType is the attachment's content type, when present.
	AttachmentType *string `json:"attachment_type"`

	// CreatedAt is when the submission was received.
	CreatedAt time.Time `json:"created_at"`
}

// ListFormSubmissionsParams are the query params for
// FormsClient.ListFormSubmissions. Zero values are omitted from the query
// string.
type ListFormSubmissionsParams struct {
	// SubmissionID filters to one submission (query param submission_id).
	SubmissionID string

	// OrderBy is the sort column, allowlisted server-side:
	// submitter_email; default created_at.
	OrderBy string

	// Order is the sort direction: "asc" or "desc".
	Order string

	// Page is the 1-based page number.
	Page int

	// Items is the page size. The server caps this at 100.
	Items int
}

// ListFormSubmissionsResponse is the response from
// FormsClient.ListFormSubmissions.
type ListFormSubmissionsResponse struct {
	// Data is the current page of submissions.
	Data []FormSubmission `json:"data"`

	// Total is the total number of matching submissions.
	Total int64 `json:"total"`

	// Page is the current (1-based) page number.
	Page int `json:"page"`

	// Items is the page size used.
	Items int `json:"items"`
}

// FormSubmissionAttachment is a file attached to a form submission.
type FormSubmissionAttachment struct {
	// Name is the attachment's filename.
	Name string

	// Content is the raw file bytes; the SDK encodes them on the wire as
	// unpadded standard base64 (base64.RawStdEncoding — the Forms service
	// rejects '=' padding).
	Content []byte
}

// SubmitFormRequest is the request for FormsClient.SubmitForm.
type SubmitFormRequest struct {
	// FormData holds the submitted field values, keyed by the form's
	// slugified field names. Required. Unlike the Email API's
	// template_values, it is sent as a real JSON object.
	FormData map[string]any

	// Attachments are optional file attachments for the submission.
	Attachments []FormSubmissionAttachment
}
