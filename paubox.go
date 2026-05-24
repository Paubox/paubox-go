// Package paubox provides a Go client for Paubox APIs.
//
// Paubox delivers HIPAA-compliant transactional email and secure web forms.
// This SDK covers the Email API (sending messages, delivery dispositions,
// dynamic Handlebars templates) and the public Forms API (retrieving form
// schemas and submitting responses).
//
// The SDK has zero external runtime dependencies.
//
// # Quick start
//
//	client, err := paubox.New("YOUR_API_KEY", "YOUR_USERNAME")
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
// The Paubox API uses a non-standard authorization header format:
//
//	Authorization: Token token=<API_KEY>
//
// The client sets this header on every request automatically. You never
// need to construct it yourself.
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
package paubox
