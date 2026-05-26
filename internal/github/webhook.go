package github

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Webhook struct {
	Event        string
	DeliveryID   string
	Signature256 string
	Payload      PullRequestWebhookPayload
}

type PullRequestWebhookPayload struct {
	Action       string `json:"action"`
	Installation struct {
		ID int64 `json:"id"`
	} `json:"installation"`
	Repository struct {
		Name     string `json:"name"`
		FullName string `json:"full_name"`
		Owner    struct {
			Login string `json:"login"`
		} `json:"owner"`
	} `json:"repository"`
	PullRequest struct {
		Number  int    `json:"number"`
		Title   string `json:"title"`
		Body    string `json:"body"`
		HTMLURL string `json:"html_url"`
		Head    struct {
			SHA string `json:"sha"`
		} `json:"head"`
		Base struct {
			SHA string `json:"sha"`
		} `json:"base"`
		User struct {
			Login string `json:"login"`
			Type  string `json:"type"`
		} `json:"user"`
	} `json:"pull_request"`
}

type WebhookDecoder struct {
	secret string
}

func NewWebhookDecoder(secret string) WebhookDecoder {
	return WebhookDecoder{secret: secret}
}

func (d WebhookDecoder) Decode(r *http.Request) (Webhook, error) {
	event := r.Header.Get("X-GitHub-Event")
	if event == "" {
		return Webhook{}, errors.New("missing GitHub event header")
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return Webhook{}, err
	}

	signature := r.Header.Get("X-Hub-Signature-256")
	if err := d.verify(body, signature); err != nil {
		return Webhook{}, err
	}

	var payload PullRequestWebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return Webhook{}, err
	}

	return Webhook{
		Event:        event,
		DeliveryID:   r.Header.Get("X-GitHub-Delivery"),
		Signature256: signature,
		Payload:      payload,
	}, nil
}

func (d WebhookDecoder) verify(payload []byte, signature string) error {
	if d.secret == "" {
		return nil
	}
	if signature == "" {
		return errors.New("missing GitHub signature header")
	}
	if !strings.HasPrefix(signature, "sha256=") {
		return fmt.Errorf("unsupported GitHub signature format")
	}

	expectedMAC := hmac.New(sha256.New, []byte(d.secret))
	expectedMAC.Write(payload)
	expected := "sha256=" + hex.EncodeToString(expectedMAC.Sum(nil))

	if !hmac.Equal([]byte(expected), []byte(signature)) {
		return errors.New("invalid GitHub webhook signature")
	}

	return nil
}
