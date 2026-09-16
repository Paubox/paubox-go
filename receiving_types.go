package paubox

import "encoding/json"

// ---------------------------------------------------------------------------
// Receiving domains
// ---------------------------------------------------------------------------

// ReceivingDomain is a domain configured for inbound email receiving.
type ReceivingDomain struct {
	ID             int     `json:"id"`
	Slug           string  `json:"slug"`
	Domain         string  `json:"domain"`
	State          string  `json:"state"`
	MXVerified     bool    `json:"mx_verified"`
	DNSZoneFile    string  `json:"dns_zone_file,omitempty"`
	DKIMPublicKey  *string `json:"dkim_public_key,omitempty"`
	ServerDomainID string  `json:"server_domain_id,omitempty"`
	CustomerID     int     `json:"customer_id,omitempty"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      string  `json:"updated_at"`
}

// CreateReceivingDomainRequest is the request for [Client.CreateReceivingDomain].
type CreateReceivingDomainRequest struct {
	Slug string `json:"slug,omitempty"`
}

type createReceivingDomainWire struct {
	Slug string `json:"slug,omitempty"`
}

type receivingDomainDataEnvelope struct {
	Data ReceivingDomain `json:"data"`
}

type receivingDomainListEnvelope struct {
	Data []ReceivingDomain `json:"data"`
}

// ---------------------------------------------------------------------------
// Mailboxes
// ---------------------------------------------------------------------------

// Mailbox is a mailbox on a receiving domain.
type Mailbox struct {
	ID              int    `json:"id"`
	DomainID        int    `json:"domain_id"`
	Name            string `json:"name"`
	EmailAddress    string `json:"email_address"`
	QuotaBytes      *int64 `json:"quota_bytes,omitempty"`
	ServerAccountID string `json:"server_account_id,omitempty"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
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

type mailboxDataEnvelope struct {
	Data Mailbox `json:"data"`
}

type mailboxListEnvelope struct {
	Data []Mailbox `json:"data"`
}

// ---------------------------------------------------------------------------
// Received emails
// ---------------------------------------------------------------------------

// EmailAddress is a structured email address with optional display name.
type EmailAddress struct {
	Address string  `json:"address"`
	Name    *string `json:"name,omitempty"`
}

// ReceivedEmail is an inbound email message.
type ReceivedEmail struct {
	EmailID       string               `json:"email_id"`
	AccountID     string               `json:"account_id,omitempty"`
	Domain        string               `json:"domain,omitempty"`
	From          []EmailAddress       `json:"from"`
	To            []EmailAddress       `json:"to"`
	CC            []EmailAddress       `json:"cc,omitempty"`
	Subject       string               `json:"subject"`
	TextBody      string               `json:"text_body,omitempty"`
	HTMLBody      string               `json:"html_body,omitempty"`
	HasAttachment bool                 `json:"has_attachment"`
	Attachments   []ReceivedAttachment `json:"attachments,omitempty"`
	Spam          bool                 `json:"spam"`
	SpamScore     *float64             `json:"spam_score,omitempty"`
	Size          int64                `json:"size,omitempty"`
	ReceivedAt    string               `json:"received_at"`
	Date          string               `json:"date,omitempty"`
	MessageID     []string             `json:"message_id,omitempty"`
}

type receivedEmailDataEnvelope struct {
	Data ReceivedEmail `json:"data"`
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
	Emails  []ReceivedEmail `json:"data"`
	HasMore bool            `json:"has_more"`
	Object  string          `json:"object"`
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
