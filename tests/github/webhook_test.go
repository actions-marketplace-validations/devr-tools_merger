package github_test

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http/httptest"
	"testing"

	"github.com/devr-tools/merger/internal/github"
)

func TestWebhookDecoderRejectsInvalidSignature(t *testing.T) {
	decoder := github.NewWebhookDecoder("secret")
	req := httptest.NewRequest("POST", "/webhooks/github", bytes.NewBufferString(`{"action":"opened"}`))
	req.Header.Set("X-GitHub-Event", "pull_request")
	req.Header.Set("X-Hub-Signature-256", "sha256=bad")

	if _, err := decoder.Decode(req); err == nil {
		t.Fatal("expected invalid signature error")
	}
}

func TestWebhookDecoderAcceptsValidSignature(t *testing.T) {
	body := []byte(`{"action":"opened","repository":{"name":"repo","full_name":"acme/repo","owner":{"login":"acme"}},"pull_request":{"number":1,"title":"x","body":"y","html_url":"https://example.com","head":{"sha":"head"},"base":{"sha":"base"},"user":{"login":"bot","type":"Bot"}}}`)
	mac := hmac.New(sha256.New, []byte("secret"))
	mac.Write(body)
	signature := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	decoder := github.NewWebhookDecoder("secret")
	req := httptest.NewRequest("POST", "/webhooks/github", bytes.NewBuffer(body))
	req.Header.Set("X-GitHub-Event", "pull_request")
	req.Header.Set("X-Hub-Signature-256", signature)

	hook, err := decoder.Decode(req)
	if err != nil {
		t.Fatalf("decode webhook: %v", err)
	}
	if hook.Payload.Action != "opened" {
		t.Fatalf("expected opened action, got %s", hook.Payload.Action)
	}
}
