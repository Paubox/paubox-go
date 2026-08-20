// Package paubox provides Go clients for the Paubox Email and Forms APIs.
//
// Paubox delivers HIPAA-compliant transactional email and form collection.
// This SDK covers the Email API — sending individual and batch messages,
// retrieving delivery dispositions, and managing dynamic Handlebars
// templates — and the Paubox Forms API: creating and managing forms,
// archiving/copying and statistics, public form retrieval and submission,
// and listing or exporting submissions (CSV/PDF).
//
// The SDK has zero external runtime dependencies.
//
// # Quick start
//
//	client, err := paubox.New("YOUR_API_KEY")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	resp, err := client.SendMessage(ctx, &paubox.SendMessageRequest{
//	    Message: paubox.Message{
//	        Recipients: []string{"recipient@example.com"},
//	        Headers: paubox.MessageHeaders{
//	            From:    "sender@yourdomain.com",
//	            Subject: "Hello from Paubox",
//	        },
//	        Content: paubox.MessageContent{
//	            PlainText: paubox.Ptr("Hello, world!"),
//	        },
//	    },
//	})
//
// # Authentication
//
// The Paubox Email API uses a non-standard authorization header format:
//
//	Authorization: Token token=<API_KEY>
//
// The client sets this header on every request automatically. You never
// need to construct it yourself.
//
// # Paubox Forms
//
// Forms use a dedicated client created with [NewForms]. Protected Forms
// endpoints authenticate with a scoped API key carrying the "forms" scope,
// sent as a standard Bearer header:
//
//	Authorization: Bearer <SCOPED_API_KEY>
//
// Email keys and Forms scoped keys are not interchangeable. Passing an
// empty key to [NewForms] yields a public-only client on which only the
// unauthenticated endpoints — [FormsClient.GetPublicForm] and
// [FormsClient.SubmitForm] — work:
//
//	forms, err := paubox.NewForms(os.Getenv("PAUBOX_FORMS_API_KEY"))
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	created, err := forms.CreateForm(ctx, &paubox.CreateFormRequest{
//	    Title:      "Patient intake",
//	    CustomerID: 1234,
//	    FormJSON:   json.RawMessage(`{"fields":[]}`),
//	})
//
// # Error handling
//
// All API errors are returned as *[PauboxError]. Use [errors.As] to inspect
// them and [errors.Is] to match against the sentinel values:
//
//	var apiErr *paubox.PauboxError
//	if errors.As(err, &apiErr) {
//	    fmt.Printf("HTTP %d: %s\n", apiErr.StatusCode, apiErr.Title)
//	}
//
//	if errors.Is(err, paubox.ErrUnauthorized) {
//	    // rotate your API key
//	}
//
// # HIPAA / PHI note
//
// This SDK never logs request bodies, response bodies, or API credentials.
// Callers must take care not to include Protected Health Information (PHI)
// in log statements, error messages, or telemetry in their own code.
// See SECURITY.md and SECURITY_REVIEW.md for detailed guidance.
//
// # Versioning
//
// This SDK follows Semantic Versioning 2.0.0. The exported identifiers of
// this package are the public API; the internal package and the examples
// directory are not covered. See README.md for the full compatibility
// policy.
package paubox

// Version is the released version of this SDK. It is reported in the
// User-Agent header of every request.
//
// This constant and the newest release heading in CHANGELOG.md are kept in
// lockstep by TestVersionMatchesChangelog.
const Version = "1.0.0"
