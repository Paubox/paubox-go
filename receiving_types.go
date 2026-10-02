package paubox

import "encoding/json"

// ---------------------------------------------------------------------------
// Receiving domains
// ---------------------------------------------------------------------------

// ReceivingDomain is a domain configured for inbound email receiving.
type ReceivingDomain struct {
	ID            int     `json:"id"`
	Slug          string  `json:"slug"`
	Domain        string  `json:"domain"`
	State         string  `json:"state"`
	MXVerified    bool    `json:"mx_verified"`
	DNSZoneFile   string  `json:"dns_zone_file,omitempty"`
	DKIMPublicKey *string `json:"dkim_public_key,omitempty"`
	CustomerID    int     `json:"customer_id,omitempty"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`

	// Deprecated: The API no longer returns server_domain_id; this field is
	// always empty.
	ServerDomainID string `json:"server_domain_id,omitempty"`
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
	ID           int    `json:"id"`
	DomainID     int    `json:"domain_id"`
	Name         string `json:"name"`
	EmailAddress string `json:"email_address"`
	QuotaBytes   *int64 `json:"quota_bytes,omitempty"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`

	// Deprecated: The API no longer returns server_account_id; this field is
	// always empty.
	ServerAccountID string `json:"server_account_id,omitempty"`
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

// ReceivedEmail is an inbound email message. EmailID is a Paubox UUID.
//
// [Client.ListReceivedEmails] populates only EmailID, From, To, Subject,
// ReceivedAt, HasAttachment, Spam, Size and Domain; [Client.GetReceivedEmail]
// populates every field except HasAttachment.
type ReceivedEmail struct {
	EmailID        string                       `json:"email_id"`
	Domain         string                       `json:"domain,omitempty"`
	From           []EmailAddress               `json:"from"`
	To             []EmailAddress               `json:"to"`
	CC             []EmailAddress               `json:"cc,omitempty"`
	Subject        string                       `json:"subject"`
	TextBody       string                       `json:"text_body,omitempty"`
	HTMLBody       string                       `json:"html_body,omitempty"`
	HasAttachment  bool                         `json:"has_attachment"`
	Attachments    []ReceivedAttachment         `json:"attachments,omitempty"`
	Spam           bool                         `json:"spam"`
	SpamScore      *float64                     `json:"spam_score,omitempty"`
	Size           int64                        `json:"size,omitempty"`
	ReceivedAt     string                       `json:"received_at"`
	Date           string                       `json:"date,omitempty"`
	MessageID      []string                     `json:"message_id,omitempty"`
	InReplyTo      []string                     `json:"in_reply_to,omitempty"`
	References     []string                     `json:"references,omitempty"`
	Authentication *ReceivedEmailAuthentication `json:"authentication,omitempty"`
	Headers        []ReceivedEmailHeader        `json:"headers,omitempty"`

	// Deprecated: The API no longer returns account_id; this field is always
	// empty.
	AccountID string `json:"account_id,omitempty"`
}

// ReceivedEmailAuthentication holds the SPF, DKIM and DMARC results for a
// received email.
type ReceivedEmailAuthentication struct {
	SPF   string `json:"spf"`
	DKIM  string `json:"dkim"`
	DMARC string `json:"dmarc"`
}

// ReceivedEmailHeader is a single header from a received email.
type ReceivedEmailHeader struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type receivedEmailDataEnvelope struct {
	Data ReceivedEmail `json:"data"`
}

// ReceivedAttachment is metadata for an attachment on a received email. ID is
// a Paubox UUID; pass it to [Client.DownloadAttachment].
type ReceivedAttachment struct {
	ID          string  `json:"id"`
	FileName    string  `json:"filename"`
	ContentType string  `json:"content_type"`
	Size        int64   `json:"size"`
	ContentID   *string `json:"content_id,omitempty"`
	DownloadURL string  `json:"download_url"`

	// Deprecated: The API no longer returns blob_id; this field is always
	// empty. Use ID.
	BlobID string `json:"blob_id,omitempty"`
}

// ListReceivedEmailsRequest is the request for [Client.ListReceivedEmails].
// Limit defaults to 25 and is capped at 100 by the server. After and Before
// take an EmailID from a previous page.
type ListReceivedEmailsRequest struct {
	Limit     *int    `json:"-"`
	After     *string `json:"-"`
	Before    *string `json:"-"`
	Search    *string `json:"-"`
	Sort      *string `json:"-"`
	Ascending *bool   `json:"-"`
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
// Content holds the raw file bytes; ContentType and FileName come from the
// response's Content-Type and Content-Disposition headers. FileName is empty
// when the server does not supply one.
type AttachmentDownload struct {
	FileName    string `json:"file_name"`
	ContentType string `json:"content_type"`
	Content     []byte `json:"content,omitempty"`

	// Deprecated: The API no longer returns blob_id; this field is always
	// empty.
	BlobID string `json:"blob_id,omitempty"`

	// Deprecated: The endpoint returns raw file bytes, not JSON; this field
	// is always nil. Use Content.
	Data json.RawMessage `json:"data,omitempty"`
}
