package paubox

import (
	"context"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// ListReceivingDomains returns all receiving domains.
//
// API: GET /receiving/domains
func (c *Client) ListReceivingDomains(ctx context.Context) ([]ReceivingDomain, error) {
	var env receivingDomainListEnvelope
	if err := c.doJSON(ctx, http.MethodGet, "/receiving/domains", nil, &env); err != nil {
		return nil, err
	}
	return env.Data, nil
}

// CreateReceivingDomain creates a new receiving domain. The slug is optional.
//
// API: POST /receiving/domains
func (c *Client) CreateReceivingDomain(ctx context.Context, req *CreateReceivingDomainRequest) (*ReceivingDomain, error) {
	var wire any
	if req != nil && req.Slug != "" {
		wire = createReceivingDomainWire{Slug: req.Slug}
	}

	var env receivingDomainDataEnvelope
	if err := c.doJSON(ctx, http.MethodPost, "/receiving/domains", wire, &env); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

// GetReceivingDomain retrieves a single receiving domain by ID.
//
// API: GET /receiving/domains/{id}
func (c *Client) GetReceivingDomain(ctx context.Context, domainID int) (*ReceivingDomain, error) {
	if domainID <= 0 {
		return nil, fmt.Errorf("paubox: GetReceivingDomain: domainID must be positive")
	}

	var env receivingDomainDataEnvelope
	if err := c.doJSON(ctx, http.MethodGet, "/receiving/domains/"+strconv.Itoa(domainID), nil, &env); err != nil {
		return nil, err
	}
	return &env.Data, nil
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

	var env mailboxListEnvelope
	if err := c.doJSON(ctx, http.MethodGet, "/receiving/domains/"+strconv.Itoa(domainID)+"/mailboxes", nil, &env); err != nil {
		return nil, err
	}
	return env.Data, nil
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

	var env mailboxDataEnvelope
	if err := c.doJSON(ctx, http.MethodPost, "/receiving/domains/"+strconv.Itoa(domainID)+"/mailboxes", wire, &env); err != nil {
		return nil, err
	}
	return &env.Data, nil
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

	var env mailboxDataEnvelope
	if err := c.doJSON(ctx, http.MethodGet, "/receiving/domains/"+strconv.Itoa(domainID)+"/mailboxes/"+strconv.Itoa(mailboxID), nil, &env); err != nil {
		return nil, err
	}
	return &env.Data, nil
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

// ListReceivedEmails lists received emails with optional pagination, search
// and sorting.
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
		if req.Search != nil {
			q.Set("search", *req.Search)
		}
		if req.Sort != nil {
			q.Set("sort", *req.Sort)
		}
		if req.Ascending != nil {
			q.Set("ascending", strconv.FormatBool(*req.Ascending))
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

// GetReceivedEmail retrieves a single received email by its Paubox UUID.
//
// API: GET /receiving/{email_id}
func (c *Client) GetReceivedEmail(ctx context.Context, emailID string) (*ReceivedEmail, error) {
	if strings.TrimSpace(emailID) == "" {
		return nil, fmt.Errorf("paubox: GetReceivedEmail: emailID must not be empty")
	}

	var env receivedEmailDataEnvelope
	if err := c.doJSON(ctx, http.MethodGet, "/receiving/"+emailID, nil, &env); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

// DownloadAttachment downloads an attachment from a received email and
// returns its raw bytes. emailID and attachmentID are Paubox UUIDs; the
// attachment ID is [ReceivedAttachment.ID].
//
// The returned bytes may contain PHI; handle and store them accordingly.
//
// API: GET /receiving/{email_id}/attachments/{attachment_id}
func (c *Client) DownloadAttachment(ctx context.Context, emailID, attachmentID string) (*AttachmentDownload, error) {
	if strings.TrimSpace(emailID) == "" {
		return nil, fmt.Errorf("paubox: DownloadAttachment: emailID must not be empty")
	}
	if strings.TrimSpace(attachmentID) == "" {
		return nil, fmt.Errorf("paubox: DownloadAttachment: attachmentID must not be empty")
	}

	resp, err := c.do(ctx, http.MethodGet, "/receiving/"+emailID+"/attachments/"+attachmentID, nil, "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close() //nolint:errcheck // close-on-defer; read errors already reported by ReadAll above

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("paubox: reading response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, parseAPIError(resp.StatusCode, resp.Header.Get("X-Request-Id"), raw)
	}

	dl := &AttachmentDownload{
		ContentType: resp.Header.Get("Content-Type"),
		Content:     raw,
	}
	if _, params, err := mime.ParseMediaType(resp.Header.Get("Content-Disposition")); err == nil {
		dl.FileName = params["filename"]
	}
	return dl, nil
}
