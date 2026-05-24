// forms demonstrates retrieving a form's schema and submitting a response
// via the Paubox Forms API. No API credentials are required.
//
// Usage:
//
//	PAUBOX_FORM_ID=your-form-uuid go run main.go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/paubox/paubox-go"
)

func main() {
	formID := requireEnv("PAUBOX_FORM_ID")

	// FormsClient requires no API key — the Forms API is public.
	fc, err := paubox.NewFormsClient()
	if err != nil {
		log.Fatalf("NewFormsClient: %v", err)
	}

	ctx := context.Background()

	// Retrieve the form schema to inspect its fields.
	form, err := fc.GetForm(ctx, formID)
	if err != nil {
		log.Fatalf("GetForm: %v", err)
	}

	fmt.Printf("Form: %s\n", form.Title)
	if form.Description != nil {
		fmt.Printf("Description: %s\n", *form.Description)
	}
	fmt.Printf("Active: %v  Signable: %v\n\n", form.Active, form.Signable)

	if form.FormJSON != nil {
		fmt.Printf("Fields (%d):\n", len(form.FormJSON.Body))
		for _, field := range form.FormJSON.Body {
			props, _ := json.Marshal(field.Properties)
			fmt.Printf("  [%s] %s  properties: %s\n", field.Type, field.ID, props)
		}
	}

	fmt.Println()

	// Submit a response. Keys in FormData should match the form's field names.
	resp, err := fc.SubmitForm(ctx, formID, paubox.FormSubmission{
		FormData: map[string]any{
			"name":  "Alice Example",
			"email": "alice@example.com",
		},
	})
	if err != nil {
		log.Fatalf("SubmitForm: %v", err)
	}

	_ = resp
	fmt.Println("Submission accepted.")
}

func requireEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("environment variable %s is required", key)
	}
	return v
}
