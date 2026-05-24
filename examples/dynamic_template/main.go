// dynamic_template demonstrates the full lifecycle of a dynamic Handlebars
// template: create → list → get → update → delete.
//
// Usage:
//
//	PAUBOX_API_KEY=your-key PAUBOX_USERNAME=your-username go run main.go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/paubox/paubox-go"
)

func main() {
	client, err := paubox.New(requireEnv("PAUBOX_API_KEY"), requireEnv("PAUBOX_USERNAME"))
	if err != nil {
		log.Fatalf("paubox.New: %v", err)
	}

	ctx := context.Background()

	// --- Create ---
	// Template bodies use Handlebars syntax. Variables: {{variableName}}.
	tmpl, err := client.CreateTemplate(ctx, &paubox.CreateTemplateRequest{
		Name: "appointment-confirmation",
		Body: []byte(`<p>Hello {{first_name}}, your appointment is on {{date}} at {{time}}.</p>`),
	})
	if err != nil {
		log.Fatalf("CreateTemplate: %v", err)
	}
	fmt.Printf("Created: %s (ID: %s)\n", tmpl.Name, tmpl.ID)

	// --- List ---
	list, err := client.ListTemplates(ctx)
	if err != nil {
		log.Fatalf("ListTemplates: %v", err)
	}
	fmt.Printf("Total templates in account: %d\n", len(list.Templates))

	// --- Get ---
	fetched, err := client.GetTemplate(ctx, tmpl.ID)
	if err != nil {
		log.Fatalf("GetTemplate: %v", err)
	}
	fmt.Printf("Fetched: %s\n", fetched.Name)

	// --- Update (name only) ---
	updated, err := client.UpdateTemplate(ctx, tmpl.ID, &paubox.UpdateTemplateRequest{
		Name: "appointment-confirmation-v2",
	})
	if err != nil {
		log.Fatalf("UpdateTemplate: %v", err)
	}
	fmt.Printf("Updated name: %s\n", updated.Name)

	// --- Delete ---
	del, err := client.DeleteTemplate(ctx, tmpl.ID)
	if err != nil {
		log.Fatalf("DeleteTemplate: %v", err)
	}
	fmt.Printf("Deleted: %s\n", del.Message)
}

func requireEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("environment variable %s is required", key)
	}
	return v
}
