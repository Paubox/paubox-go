// send_single demonstrates sending a single HIPAA-compliant email via Paubox.
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
	"time"

	"github.com/paubox/paubox-go"
)

func main() {
	apiKey := requireEnv("PAUBOX_API_KEY")
	username := requireEnv("PAUBOX_USERNAME")

	// Create a client once and reuse it. It is safe for concurrent use.
	client, err := paubox.New(apiKey, username,
		paubox.WithTimeout(15*time.Second),
	)
	if err != nil {
		log.Fatalf("paubox.New: %v", err)
	}

	ctx := context.Background()

	// Send a single message.
	resp, err := client.SendMessage(ctx, &paubox.SendMessageRequest{
		Message: paubox.Message{
			// Recipients accepts plain addresses or "Display Name <addr>" format.
			Recipients: []string{"recipient@example.com"},
			Headers: paubox.MessageHeaders{
				// From must match a domain verified in your Paubox account.
				From:    "sender@yourdomain.com",
				Subject: "Hello from Paubox",
			},
			Content: paubox.MessageContent{
				PlainText: paubox.Ptr("This is the plain-text body."),
				HTML:      paubox.Ptr("<p>This is the <strong>HTML</strong> body.</p>"),
			},
		},
	})
	if err != nil {
		log.Fatalf("SendMessage: %v", err)
	}
	fmt.Printf("Sent! Tracking ID: %s  Status: %s\n", resp.SourceTrackingID, resp.Data)

	// Poll for delivery status once (use webhooks in production instead).
	time.Sleep(5 * time.Second)

	disp, err := client.GetEmailDisposition(ctx, resp.SourceTrackingID)
	if err != nil {
		log.Fatalf("GetEmailDisposition: %v", err)
	}
	for _, d := range disp.Data.Message.MessageDeliveries {
		fmt.Printf("  %s → %s\n", d.Recipient, d.Status.DeliveryStatus)
	}
}

func requireEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("environment variable %s is required", key)
	}
	return v
}
