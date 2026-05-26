package github

import (
	"encoding/json"
	"errors"
	"net/http"
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

func DecodeWebhookRequest(r *http.Request) (Webhook, error) {
	event := r.Header.Get("X-GitHub-Event")
	if event == "" {
		return Webhook{}, errors.New("missing GitHub event header")
	}

	var payload PullRequestWebhookPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		return Webhook{}, err
	}

	return Webhook{
		Event:        event,
		DeliveryID:   r.Header.Get("X-GitHub-Delivery"),
		Signature256: r.Header.Get("X-Hub-Signature-256"),
		Payload:      payload,
	}, nil
}
