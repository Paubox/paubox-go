package paubox

import (
	"context"
	"encoding/json"
	"errors"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// templateFixture returns a JSON-encoded Template for use in test handlers.
func templateFixture(id, name, body string) string {
	t := Template{
		ID:        id,
		Name:      name,
		Body:      body,
		CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
	}
	b, _ := json.Marshal(t)
	return string(b)
}

// readMultipartForm parses a multipart body from an incoming request.
func readMultipartForm(t *testing.T, r *http.Request) *multipart.Form {
	t.Helper()
	ct := r.Header.Get("Content-Type")
	_, params, err := mime.ParseMediaType(ct)
	if err != nil {
		t.Fatalf("ParseMediaType: %v", err)
	}
	form, err := multipart.NewReader(r.Body, params["boundary"]).ReadForm(1 << 20)
	if err != nil {
		t.Fatalf("ReadForm: %v", err)
	}
	return form
}

// ---------------------------------------------------------------------------
// ListTemplates
// ---------------------------------------------------------------------------

func TestListTemplates_HappyPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/testuser/dynamic_templates" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		respondJSON(w, 200, `{"templates":[`+templateFixture("t1", "welcome", "Hi {{name}}")+`]}`)
	}))
	defer srv.Close()

	resp, err := newTestClient(t, srv).ListTemplates(context.Background())
	if err != nil {
		t.Fatalf("ListTemplates() error: %v", err)
	}
	if len(resp.Templates) != 1 {
		t.Fatalf("len = %d, want 1", len(resp.Templates))
	}
	if resp.Templates[0].ID != "t1" || resp.Templates[0].Name != "welcome" {
		t.Errorf("unexpected template: %+v", resp.Templates[0])
	}
}

func TestListTemplates_Empty(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		respondJSON(w, 200, `{"templates":[]}`)
	}))
	defer srv.Close()

	resp, err := newTestClient(t, srv).ListTemplates(context.Background())
	if err != nil {
		t.Fatalf("ListTemplates() error: %v", err)
	}
	if len(resp.Templates) != 0 {
		t.Errorf("expected 0 templates, got %d", len(resp.Templates))
	}
}

func TestListTemplates_Unauthorized(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		respondJSON(w, 401, `{"errors":[{"code":401,"title":"Unauthorized","details":"bad key"}]}`)
	}))
	defer srv.Close()

	_, err := newTestClient(t, srv).ListTemplates(context.Background())
	if !errors.Is(err, ErrUnauthorized) {
		t.Errorf("expected ErrUnauthorized, got %v", err)
	}
}

// ---------------------------------------------------------------------------
// GetTemplate
// ---------------------------------------------------------------------------

func TestGetTemplate_HappyPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/testuser/dynamic_templates/tmpl-1" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		respondJSON(w, 200, templateFixture("tmpl-1", "onboard", "Hello {{name}}"))
	}))
	defer srv.Close()

	tmpl, err := newTestClient(t, srv).GetTemplate(context.Background(), "tmpl-1")
	if err != nil {
		t.Fatalf("GetTemplate() error: %v", err)
	}
	if tmpl.ID != "tmpl-1" {
		t.Errorf("ID = %q, want tmpl-1", tmpl.ID)
	}
}

func TestGetTemplate_EmptyID(t *testing.T) {
	c, _ := New("k", "u")
	_, err := c.GetTemplate(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty id")
	}
}

func TestGetTemplate_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		respondJSON(w, 404, `{"errors":[{"code":404,"title":"Not Found","details":"template not found"}]}`)
	}))
	defer srv.Close()

	_, err := newTestClient(t, srv).GetTemplate(context.Background(), "ghost")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

// ---------------------------------------------------------------------------
// CreateTemplate
// ---------------------------------------------------------------------------

func TestCreateTemplate_HappyPath(t *testing.T) {
	var formName, formBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		form := readMultipartForm(t, r)
		if v := form.Value["data[name]"]; len(v) > 0 {
			formName = v[0]
		}
		if files := form.File["data[body]"]; len(files) > 0 {
			f, _ := files[0].Open()
			var sb strings.Builder
			buf := make([]byte, 512)
			for {
				n, err := f.Read(buf)
				if n > 0 {
					sb.Write(buf[:n])
				}
				if err != nil {
					break
				}
			}
			formBody = sb.String()
		}
		respondJSON(w, 200, templateFixture("new-1", formName, formBody))
	}))
	defer srv.Close()

	tmpl, err := newTestClient(t, srv).CreateTemplate(context.Background(), &CreateTemplateRequest{
		Name: "my-tmpl",
		Body: []byte("Hello {{first_name}}"),
	})
	if err != nil {
		t.Fatalf("CreateTemplate() error: %v", err)
	}
	if tmpl.ID != "new-1" {
		t.Errorf("ID = %q, want new-1", tmpl.ID)
	}
	if formName != "my-tmpl" {
		t.Errorf("form data[name] = %q, want my-tmpl", formName)
	}
	if formBody != "Hello {{first_name}}" {
		t.Errorf("form data[body] = %q", formBody)
	}
}

func TestCreateTemplate_Validation(t *testing.T) {
	tests := []struct {
		name    string
		req     *CreateTemplateRequest
		wantErr string
	}{
		{"nil", nil, "nil"},
		{"no name", &CreateTemplateRequest{Body: []byte("x")}, "Name"},
		{"no body", &CreateTemplateRequest{Name: "n"}, "Body"},
	}
	c, _ := New("k", "u")
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := c.CreateTemplate(context.Background(), tc.req)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("error %q does not contain %q", err.Error(), tc.wantErr)
			}
		})
	}
}

func TestCreateTemplate_400(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		respondJSON(w, 400, `{"errors":[{"code":400,"title":"Bad Request","details":"invalid template"}]}`)
	}))
	defer srv.Close()

	_, err := newTestClient(t, srv).CreateTemplate(context.Background(), &CreateTemplateRequest{
		Name: "bad", Body: []byte("{{"),
	})
	if !errors.Is(err, ErrBadRequest) {
		t.Errorf("expected ErrBadRequest, got %v", err)
	}
}

// ---------------------------------------------------------------------------
// UpdateTemplate
// ---------------------------------------------------------------------------

func TestUpdateTemplate_HappyPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		respondJSON(w, 200, templateFixture("t1", "updated-name", "new body"))
	}))
	defer srv.Close()

	tmpl, err := newTestClient(t, srv).UpdateTemplate(context.Background(), "t1", &UpdateTemplateRequest{
		Name: "updated-name",
		Body: []byte("new body"),
	})
	if err != nil {
		t.Fatalf("UpdateTemplate() error: %v", err)
	}
	if tmpl.Name != "updated-name" {
		t.Errorf("Name = %q, want updated-name", tmpl.Name)
	}
}

func TestUpdateTemplate_NameOnly(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		respondJSON(w, 200, templateFixture("t1", "new-name", "body"))
	}))
	defer srv.Close()

	_, err := newTestClient(t, srv).UpdateTemplate(context.Background(), "t1", &UpdateTemplateRequest{Name: "new-name"})
	if err != nil {
		t.Fatalf("UpdateTemplate() error: %v", err)
	}
}

func TestUpdateTemplate_Validation(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		req     *UpdateTemplateRequest
		wantErr string
	}{
		{"empty id", "", &UpdateTemplateRequest{Name: "n"}, "id"},
		{"nil req", "x", nil, "nil"},
		{"no fields", "x", &UpdateTemplateRequest{}, "at least one"},
	}
	c, _ := New("k", "u")
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := c.UpdateTemplate(context.Background(), tc.id, tc.req)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("error %q does not contain %q", err.Error(), tc.wantErr)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// DeleteTemplate
// ---------------------------------------------------------------------------

func TestDeleteTemplate_HappyPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/testuser/dynamic_templates/del-1" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		respondJSON(w, 200, `{"message":"Template deleted successfully"}`)
	}))
	defer srv.Close()

	resp, err := newTestClient(t, srv).DeleteTemplate(context.Background(), "del-1")
	if err != nil {
		t.Fatalf("DeleteTemplate() error: %v", err)
	}
	if resp.Message != "Template deleted successfully" {
		t.Errorf("Message = %q", resp.Message)
	}
}

func TestDeleteTemplate_EmptyID(t *testing.T) {
	c, _ := New("k", "u")
	_, err := c.DeleteTemplate(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty id")
	}
}

func TestDeleteTemplate_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		respondJSON(w, 404, `{"errors":[{"code":404,"title":"Not Found","details":"template not found"}]}`)
	}))
	defer srv.Close()

	_, err := newTestClient(t, srv).DeleteTemplate(context.Background(), "ghost")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

// ---------------------------------------------------------------------------
// SendTemplatedMessage
// ---------------------------------------------------------------------------

func TestSendTemplatedMessage_HappyPath(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		respondJSON(w, 200, `{"sourceTrackingId":"tmpl-tid","data":"Service OK"}`)
	}))
	defer srv.Close()

	resp, err := newTestClient(t, srv).SendTemplatedMessage(context.Background(), &SendTemplatedMessageRequest{
		TemplateName:   "welcome",
		TemplateValues: map[string]any{"name": "Alice", "score": 99},
		Message: TemplatedMessage{
			Recipients: []string{"alice@example.com"},
			Headers:    MessageHeaders{From: "s@example.com", Subject: "Welcome Alice"},
		},
	})
	if err != nil {
		t.Fatalf("SendTemplatedMessage() error: %v", err)
	}
	if resp.SourceTrackingID != "tmpl-tid" {
		t.Errorf("SourceTrackingID = %q", resp.SourceTrackingID)
	}

	// Verify template_values is a JSON string (not a nested object).
	data := gotBody["data"].(map[string]any)
	tvRaw, ok := data["template_values"].(string)
	if !ok {
		t.Fatalf("template_values type = %T, want string", data["template_values"])
	}
	var tv map[string]any
	if err := json.Unmarshal([]byte(tvRaw), &tv); err != nil {
		t.Fatalf("template_values is not valid JSON string: %v", err)
	}
	if tv["name"] != "Alice" {
		t.Errorf("template_values[name] = %v, want Alice", tv["name"])
	}
}

func TestSendTemplatedMessage_NilTemplateValues(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		respondJSON(w, 200, `{"sourceTrackingId":"x","data":"Service OK"}`)
	}))
	defer srv.Close()

	// Nil TemplateValues must not panic; marshals to "null".
	_, err := newTestClient(t, srv).SendTemplatedMessage(context.Background(), &SendTemplatedMessageRequest{
		TemplateName: "t",
		Message: TemplatedMessage{
			Recipients: []string{"r@x.com"},
			Headers:    MessageHeaders{From: "f@x.com", Subject: "s"},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSendTemplatedMessage_SendsToCorrectPath(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		respondJSON(w, 200, `{"sourceTrackingId":"x","data":"Service OK"}`)
	}))
	defer srv.Close()

	_, _ = newTestClient(t, srv).SendTemplatedMessage(context.Background(), &SendTemplatedMessageRequest{
		TemplateName: "t",
		Message: TemplatedMessage{
			Recipients: []string{"r@x.com"},
			Headers:    MessageHeaders{From: "f@x.com", Subject: "s"},
		},
	})
	if gotPath != "/testuser/templated_messages" {
		t.Errorf("path = %q, want /testuser/templated_messages", gotPath)
	}
}

func TestSendTemplatedMessage_Validation(t *testing.T) {
	tests := []struct {
		name    string
		req     *SendTemplatedMessageRequest
		wantErr string
	}{
		{"nil", nil, "nil"},
		{
			"empty template name",
			&SendTemplatedMessageRequest{
				Message: TemplatedMessage{
					Recipients: []string{"r@x.com"},
					Headers:    MessageHeaders{From: "f@x.com", Subject: "s"},
				},
			},
			"TemplateName",
		},
		{
			"no recipients",
			&SendTemplatedMessageRequest{
				TemplateName: "t",
				Message:      TemplatedMessage{Headers: MessageHeaders{From: "f@x.com", Subject: "s"}},
			},
			"recipients",
		},
		{
			"no from",
			&SendTemplatedMessageRequest{
				TemplateName: "t",
				Message: TemplatedMessage{
					Recipients: []string{"r@x.com"},
					Headers:    MessageHeaders{Subject: "s"},
				},
			},
			"from",
		},
		{
			"no subject",
			&SendTemplatedMessageRequest{
				TemplateName: "t",
				Message: TemplatedMessage{
					Recipients: []string{"r@x.com"},
					Headers:    MessageHeaders{From: "f@x.com"},
				},
			},
			"subject",
		},
	}

	c, _ := New("k", "u")
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := c.SendTemplatedMessage(context.Background(), tc.req)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("error %q does not contain %q", err.Error(), tc.wantErr)
			}
		})
	}
}

func TestSendTemplatedMessage_ErrorResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		respondJSON(w, 400, `{"errors":[{"code":400,"title":"Bad Request","details":"template not found"}]}`)
	}))
	defer srv.Close()

	_, err := newTestClient(t, srv).SendTemplatedMessage(context.Background(), &SendTemplatedMessageRequest{
		TemplateName: "missing",
		Message: TemplatedMessage{
			Recipients: []string{"r@x.com"},
			Headers:    MessageHeaders{From: "f@x.com", Subject: "s"},
		},
	})
	if !errors.Is(err, ErrBadRequest) {
		t.Errorf("expected ErrBadRequest, got %v", err)
	}
}
