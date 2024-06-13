package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

func SendSlackMessage(message string, url string) {
	webhookURL := url

	payload := map[string]string{"text": message}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("Error marshalling Slack payload: %v\n", err)
		return
	}

	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		fmt.Printf("Error sending Slack message: %v\n", err)
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Fatalf("Unable to read slack body: %v\n", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Error response from Slack: %s\n", string(body))
	}
}
