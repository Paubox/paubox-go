// send_batch demonstrates sending multiple emails in a single API call.
//
// Responses are returned in the same order as the request messages, so you
// can correlate them by index. Paubox recommends batches of 50 or fewer.
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

	recipients := []string{
		"alice@example.com",
		"bob@example.com",
		"carol@example.com",
	}

	messages := make([]paubox.Message, len(recipients))
	for i, r := range recipients {
		messages[i] = paubox.Message{
			Recipients: []string{r},
			Headers: paubox.MessageHeaders{
				From:    "sender@yourdomain.com",
				Subject: "Your monthly report",
			},
			Content: paubox.MessageContent{
				PlainText: paubox.Ptr("Hi, your report is ready."),
			},
		}
	}

	resp, err := client.SendBatch(context.Background(), &paubox.SendBatchRequest{
		Messages: messages,
	})
	if err != nil {
		log.Fatalf("SendBatch: %v", err)
	}

	for i, msg := range resp.Messages {
		fmt.Printf("[%d] %s → %s (%s)\n", i, recipients[i], msg.SourceTrackingID, msg.Data)
	}
}

func requireEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("environment variable %s is required", key)
	}
	return v
}
