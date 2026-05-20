// send_templated demonstrates sending an email rendered from a dynamic
// Handlebars template stored in your Paubox account.
//
// TemplateValues is a plain Go map — the SDK handles serialising it to the
// JSON string the API requires. Never pre-encode it yourself.
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

	resp, err := client.SendTemplatedMessage(context.Background(), &paubox.SendTemplatedMessageRequest{
		TemplateName: "appointment-confirmation",
		// Pass values as a plain map — the SDK encodes this correctly.
		TemplateValues: map[string]any{
			"first_name": "Jane",
			"date":       "2024-03-15",
			"time":       "2:00 PM",
		},
		Message: paubox.TemplatedMessage{
			Recipients: []string{"jane.doe@example.com"},
			Headers: paubox.MessageHeaders{
				From:    "appointments@yourclinic.com",
				Subject: "Your appointment is confirmed",
			},
		},
	})
	if err != nil {
		log.Fatalf("SendTemplatedMessage: %v", err)
	}

	fmt.Printf("Sent! Tracking ID: %s\n", resp.SourceTrackingID)
}

func requireEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("environment variable %s is required", key)
	}
	return v
}
