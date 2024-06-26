package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
)

var whitelistedDomains = []string{
	"hooks.slack.com",
}

func isWhitelistedDomain(requestURL string) bool {
	parsedURL, err := url.Parse(requestURL)
	if err != nil {
		return false
	}
	for _, domain := range whitelistedDomains {
		if parsedURL.Host == domain {
			return true
		}
	}
	return false
}

func isPrivateIP(ip net.IP) bool {
	privateIPBlocks := []*net.IPNet{
		{IP: net.IPv4(10, 0, 0, 0), Mask: net.CIDRMask(8, 32)},
		{IP: net.IPv4(172, 16, 0, 0), Mask: net.CIDRMask(12, 32)},
		{IP: net.IPv4(192, 168, 0, 0), Mask: net.CIDRMask(16, 32)},
		{IP: net.IPv4(127, 0, 0, 0), Mask: net.CIDRMask(8, 32)},
		{IP: net.IPv6loopback, Mask: net.CIDRMask(128, 128)},
	}

	for _, block := range privateIPBlocks {
		if block.Contains(ip) {
			return true
		}
	}
	return false
}

func isSafeURL(requestURL string) bool {
	parsedURL, err := url.Parse(requestURL)
	if err != nil {
		return false
	}

	ips, err := net.LookupIP(parsedURL.Hostname())
	if err != nil {
		return false
	}

	for _, ip := range ips {
		if isPrivateIP(ip) {
			return false
		}
	}
	return true
}

func SendSlackMessage(message string, webhookURL string) {
	if !isWhitelistedDomain(webhookURL) {
		fmt.Println("Error: URL is not whitelisted")
		return
	}

	if !isSafeURL(webhookURL) {
		fmt.Println("Error: URL resolves to a private IP")
		return
	}

	payload := map[string]string{"text": message}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("Error marshalling Slack payload: %v\n", err)
		return
	}

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error sending Slack message: %v\n", err)
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Fatalf("Unable to close response body: %v\n", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("Error reading Slack response body: %v\n", err)
			return
		}
		fmt.Printf("Error response from Slack: %s\n", string(body))
		return
	}

	fmt.Println("Slack message sent successfully")
}
