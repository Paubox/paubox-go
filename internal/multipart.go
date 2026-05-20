// Package internal contains unexported implementation helpers.
package internal

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/textproto"
)

// BuildTemplateForm constructs a multipart/form-data body for the Paubox
// dynamic template create and update endpoints.
//
// The API expects:
//   - data[name]  — a plain text field with the template name
//   - data[body]  — a binary file field with the Handlebars template content
//
// Either field may be omitted (empty name or nil body) when only one needs
// updating. Returns the encoded body bytes and the Content-Type header value
// (which includes the multipart boundary).
func BuildTemplateForm(name string, body []byte) ([]byte, string, error) {
	var buf bytes.Buffer
	encoded, ct, err := buildTemplateForm(&buf, name, body)
	if err != nil {
		return nil, "", err
	}
	_ = encoded // buf is already populated via the writer
	return buf.Bytes(), ct, nil
}

// buildTemplateForm writes the multipart form to w and returns the
// Content-Type. Separated from BuildTemplateForm to allow error injection in
// tests.
func buildTemplateForm(w io.Writer, name string, body []byte) ([]byte, string, error) {
	mw := multipart.NewWriter(w)

	if name != "" {
		if err := mw.WriteField("data[name]", name); err != nil {
			return nil, "", fmt.Errorf("internal: multipart: writing name field: %w", err)
		}
	}

	if len(body) > 0 {
		h := make(textproto.MIMEHeader)
		h.Set("Content-Disposition",
			`form-data; name="data[body]"; filename="template.hbs"`)
		h.Set("Content-Type", "application/octet-stream")

		part, err := mw.CreatePart(h)
		if err != nil {
			return nil, "", fmt.Errorf("internal: multipart: creating body part: %w", err)
		}
		if _, err := io.Copy(part, bytes.NewReader(body)); err != nil {
			return nil, "", fmt.Errorf("internal: multipart: writing body part: %w", err)
		}
	}

	if err := mw.Close(); err != nil {
		return nil, "", fmt.Errorf("internal: multipart: closing writer: %w", err)
	}

	return nil, mw.FormDataContentType(), nil
}
