package paubox

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// ListReceivingDomains returns all receiving domains.
//
// API: GET /receiving/domains
func (c *Client) ListReceivingDomains(ctx context.Context) ([]ReceivingDomain, error) {
	var domains []ReceivingDomain
	if err := c.doJSON(ctx, http.MethodGet, "/receiving/domains", nil, &domains); err != nil {
		return nil, err
	}
	return domains, nil
}

// CreateReceivingDomain creates a new receiving domain. The slug is optional.
//
// API: POST /receiving/domains
func (c *Client) CreateReceivingDomain(ctx context.Context, req *CreateReceivingDomainRequest) (*ReceivingDomain, error) {
	var wire any
	if req != nil && req.Slug != "" {
		wire = createReceivingDomainWire{Slug: req.Slug}
	}

	var domain ReceivingDomain
	if err := c.doJSON(ctx, http.MethodPost, "/receiving/domains", wire, &domain); err != nil {
		return nil, err
	}
	return &domain, nil
}

// GetReceivingDomain retrieves a single receiving domain by ID.
//
// API: GET /receiving/domains/{id}
func (c *Client) GetReceivingDomain(ctx context.Context, domainID int) (*ReceivingDomain, error) {
	if domainID <= 0 {
		return nil, fmt.Errorf("paubox: GetReceivingDomain: domainID must be positive")
	}

	var domain ReceivingDomain
	if err := c.doJSON(ctx, http.MethodGet, "/receiving/domains/"+strconv.Itoa(domainID), nil, &domain); err != nil {
		return nil, err
	}
	return &domain, nil
}

// DeleteReceivingDomain deletes a receiving domain by ID.
//
// API: DELETE /receiving/domains/{id}
func (c *Client) DeleteReceivingDomain(ctx context.Context, domainID int) error {
	if domainID <= 0 {
		return fmt.Errorf("paubox: DeleteReceivingDomain: domainID must be positive")
	}

	return c.doJSON(ctx, http.MethodDelete, "/receiving/domains/"+strconv.Itoa(domainID), nil, nil)
}

// ListMailboxes returns all mailboxes for the given receiving domain.
//
// API: GET /receiving/domains/{domain_id}/mailboxes
func (c *Client) ListMailboxes(ctx context.Context, domainID int) ([]Mailbox, error) {
	if domainID <= 0 {
		return nil, fmt.Errorf("paubox: ListMailboxes: domainID must be positive")
	}

	var mailboxes []Mailbox
	if err := c.doJSON(ctx, http.MethodGet, "/receiving/domains/"+strconv.Itoa(domainID)+"/mailboxes", nil, &mailboxes); err != nil {
		return nil, err
	}
	return mailboxes, nil
}

// CreateMailbox creates a new mailbox on the given receiving domain.
//
// API: POST /receiving/domains/{domain_id}/mailboxes
func (c *Client) CreateMailbox(ctx context.Context, domainID int, req *CreateMailboxRequest) (*Mailbox, error) {
	if domainID <= 0 {
		return nil, fmt.Errorf("paubox: CreateMailbox: domainID must be positive")
	}
	if req == nil {
		return nil, fmt.Errorf("paubox: CreateMailbox: request must not be nil")
	}
	if strings.TrimSpace(req.Name) == "" {
		return nil, fmt.Errorf("paubox: CreateMailbox: name must not be empty")
	}
	if req.Password == "" {
		return nil, fmt.Errorf("paubox: CreateMailbox: password must not be empty")
	}

	wire := createMailboxWire{
		Name:       req.Name,
		Password:   req.Password,
		QuotaBytes: req.QuotaBytes,
	}

	var mailbox Mailbox
	if err := c.doJSON(ctx, http.MethodPost, "/receiving/domains/"+strconv.Itoa(domainID)+"/mailboxes", wire, &mailbox); err != nil {
		return nil, err
	}
	return &mailbox, nil
}

// GetMailbox retrieves a single mailbox by domain and mailbox ID.
//
// API: GET /receiving/domains/{domain_id}/mailboxes/{id}
func (c *Client) GetMailbox(ctx context.Context, domainID, mailboxID int) (*Mailbox, error) {
	if domainID <= 0 {
		return nil, fmt.Errorf("paubox: GetMailbox: domainID must be positive")
	}
	if mailboxID <= 0 {
		return nil, fmt.Errorf("paubox: GetMailbox: mailboxID must be positive")
	}

	var mailbox Mailbox
	if err := c.doJSON(ctx, http.MethodGet, "/receiving/domains/"+strconv.Itoa(domainID)+"/mailboxes/"+strconv.Itoa(mailboxID), nil, &mailbox); err != nil {
		return nil, err
	}
	return &mailbox, nil
}

// DeleteMailbox deletes a mailbox by domain and mailbox ID.
//
// API: DELETE /receiving/domains/{domain_id}/mailboxes/{id}
func (c *Client) DeleteMailbox(ctx context.Context, domainID, mailboxID int) error {
	if domainID <= 0 {
		return fmt.Errorf("paubox: DeleteMailbox: domainID must be positive")
	}
	if mailboxID <= 0 {
		return fmt.Errorf("paubox: DeleteMailbox: mailboxID must be positive")
	}

	return c.doJSON(ctx, http.MethodDelete, "/receiving/domains/"+strconv.Itoa(domainID)+"/mailboxes/"+strconv.Itoa(mailboxID), nil, nil)
}

// ListReceivedEmails lists received emails with optional pagination.
//
// API: GET /receiving
func (c *Client) ListReceivedEmails(ctx context.Context, req *ListReceivedEmailsRequest) (*ListReceivedEmailsResponse, error) {
	path := "/receiving"
	if req != nil {
		q := url.Values{}
		if req.Limit != nil {
			q.Set("limit", strconv.Itoa(*req.Limit))
		}
		if req.After != nil {
			q.Set("after", *req.After)
		}
		if req.Before != nil {
			q.Set("before", *req.Before)
		}
		if encoded := q.Encode(); encoded != "" {
			path += "?" + encoded
		}
	}

	var resp ListReceivedEmailsResponse
	if err := c.doJSON(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetReceivedEmail retrieves a single received email by ID.
//
// API: GET /receiving/{email_id}
func (c *Client) GetReceivedEmail(ctx context.Context, emailID string) (*ReceivedEmail, error) {
	if strings.TrimSpace(emailID) == "" {
		return nil, fmt.Errorf("paubox: GetReceivedEmail: emailID must not be empty")
	}

	var email ReceivedEmail
	if err := c.doJSON(ctx, http.MethodGet, "/receiving/"+emailID, nil, &email); err != nil {
		return nil, err
	}
	return &email, nil
}

// DownloadAttachment downloads an attachment from a received email.
//
// API: GET /receiving/{email_id}/attachments/{blob_id}
func (c *Client) DownloadAttachment(ctx context.Context, emailID, blobID string) (*AttachmentDownload, error) {
	if strings.TrimSpace(emailID) == "" {
		return nil, fmt.Errorf("paubox: DownloadAttachment: emailID must not be empty")
	}
	if strings.TrimSpace(blobID) == "" {
		return nil, fmt.Errorf("paubox: DownloadAttachment: blobID must not be empty")
	}

	var dl AttachmentDownload
	if err := c.doJSON(ctx, http.MethodGet, "/receiving/"+emailID+"/attachments/"+blobID, nil, &dl); err != nil {
		return nil, err
	}
	return &dl, nil
}
