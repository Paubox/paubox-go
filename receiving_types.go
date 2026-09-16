package paubox

import "encoding/json"

// ---------------------------------------------------------------------------
// Receiving domains
// ---------------------------------------------------------------------------

// ReceivingDomain is a domain configured for inbound email receiving.
type ReceivingDomain struct {
	ID        int    `json:"id"`
	Slug      string `json:"slug"`
	Domain    string `json:"domain"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// CreateReceivingDomainRequest is the request for [Client.CreateReceivingDomain].
type CreateReceivingDomainRequest struct {
	Slug string `json:"slug,omitempty"`
}

type createReceivingDomainWire struct {
	Slug string `json:"slug,omitempty"`
}

// ---------------------------------------------------------------------------
// Mailboxes
// ---------------------------------------------------------------------------

// Mailbox is a mailbox on a receiving domain.
type Mailbox struct {
	ID         int    `json:"id"`
	DomainID   int    `json:"domain_id"`
	Name       string `json:"name"`
	Email      string `json:"email"`
	QuotaBytes *int64 `json:"quota_bytes,omitempty"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

// CreateMailboxRequest is the request for [Client.CreateMailbox].
type CreateMailboxRequest struct {
	Name       string `json:"name"`
	Password   string `json:"password"`
	QuotaBytes *int64 `json:"quota_bytes,omitempty"`
}

type createMailboxWire struct {
	Name       string `json:"name"`
	Password   string `json:"password"`
	QuotaBytes *int64 `json:"quota_bytes,omitempty"`
}

// ---------------------------------------------------------------------------
// Received emails
// ---------------------------------------------------------------------------

// ReceivedEmail is an inbound email message.
type ReceivedEmail struct {
	ID          string               `json:"id"`
	From        string               `json:"from"`
	To          []string             `json:"to"`
	CC          []string             `json:"cc,omitempty"`
	Subject     string               `json:"subject"`
	Body        *ReceivedBody        `json:"body,omitempty"`
	Attachments []ReceivedAttachment `json:"attachments,omitempty"`
	ReceivedAt  string               `json:"received_at"`
}

// ReceivedBody holds the body of a received email.
type ReceivedBody struct {
	PlainText *string `json:"text/plain,omitempty"`
	HTML      *string `json:"text/html,omitempty"`
}

// ReceivedAttachment is metadata for an attachment on a received email.
type ReceivedAttachment struct {
	BlobID      string `json:"blob_id"`
	FileName    string `json:"file_name"`
	ContentType string `json:"content_type"`
	Size        int64  `json:"size"`
}

// ListReceivedEmailsRequest is the request for [Client.ListReceivedEmails].
type ListReceivedEmailsRequest struct {
	Limit  *int    `json:"-"`
	After  *string `json:"-"`
	Before *string `json:"-"`
}

// ListReceivedEmailsResponse is the response from [Client.ListReceivedEmails].
type ListReceivedEmailsResponse struct {
	Emails []ReceivedEmail `json:"emails"`
}

// ---------------------------------------------------------------------------
// Attachment download
// ---------------------------------------------------------------------------

// AttachmentDownload is the response from [Client.DownloadAttachment].
type AttachmentDownload struct {
	BlobID      string          `json:"blob_id"`
	FileName    string          `json:"file_name"`
	ContentType string          `json:"content_type"`
	Data        json.RawMessage `json:"data"`
}
