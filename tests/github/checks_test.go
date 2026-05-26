package github_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/mergerhq/merger/internal/github"
)

func TestClientPublishCheckRun(t *testing.T) {
	client := newTestClient(t, roundTripFunc(func(r *http.Request) (*http.Response, error) {
		switch r.URL.Path {
		case "/app/installations/42/access_tokens":
			return jsonResponse(t, map[string]any{
				"token":      "installation-token",
				"expires_at": "2099-01-01T00:00:00Z",
			}), nil
		case "/repos/acme/merger/check-runs":
			assertMethod(t, r, http.MethodPost)
			var payload map[string]any
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatalf("decode payload: %v", err)
			}
			if payload["name"] != "merger/risk" || payload["head_sha"] != "abc123" {
				t.Fatalf("unexpected payload: %#v", payload)
			}
			if payload["details_url"] != "https://merger.example/checks/1" {
				t.Fatalf("unexpected details url: %#v", payload["details_url"])
			}
			return jsonResponse(t, map[string]string{"id": "1"}), nil
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
			return nil, nil
		}
	}))
	err := client.PublishCheckRun(context.Background(), github.CheckRunInput{
		RepoOwner:  "acme",
		RepoName:   "merger",
		HeadSHA:    "abc123",
		Name:       "merger/risk",
		Status:     "completed",
		Conclusion: "success",
		Summary:    "Low blast radius",
		DetailsURL: "https://merger.example/checks/1",
	})
	if err != nil {
		t.Fatalf("publish check run: %v", err)
	}
}
