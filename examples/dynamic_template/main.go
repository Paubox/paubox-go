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

	const name = "appointment-confirmation"

	// --- Create ---
	// Template bodies use Handlebars syntax. Variables: {{variableName}}.
	// The API confirms creation with a message but does not return the new
	// template's ID — we resolve it from ListTemplates below.
	created, err := client.CreateTemplate(ctx, &paubox.CreateTemplateRequest{
		Name: name,
		Body: []byte(`<p>Hello {{first_name}}, your appointment is on {{date}} at {{time}}.</p>`),
	})
	if err != nil {
		log.Fatalf("CreateTemplate: %v", err)
	}
	fmt.Printf("Created: %s\n", created.Message)

	// --- List (and resolve the new template's ID by name) ---
	list, err := client.ListTemplates(ctx)
	if err != nil {
		log.Fatalf("ListTemplates: %v", err)
	}
	fmt.Printf("Total templates in account: %d\n", len(list.Templates))

	var id int64
	for _, t := range list.Templates {
		if t.Name == name {
			id = t.ID
			break
		}
	}
	if id == 0 {
		log.Fatalf("could not find template %q after creating it", name)
	}

	// --- Get ---
	fetched, err := client.GetTemplate(ctx, id)
	if err != nil {
		log.Fatalf("GetTemplate: %v", err)
	}
	fmt.Printf("Fetched: %s (ID: %d)\n", fetched.Name, fetched.ID)

	// --- Update (name only) ---
	updated, err := client.UpdateTemplate(ctx, id, &paubox.UpdateTemplateRequest{
		Name: "appointment-confirmation-v2",
	})
	if err != nil {
		log.Fatalf("UpdateTemplate: %v", err)
	}
	fmt.Printf("Updated: %s\n", updated.Message)

	// --- Delete ---
	del, err := client.DeleteTemplate(ctx, id)
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
