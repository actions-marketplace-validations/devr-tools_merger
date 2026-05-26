package github_test

import (
	"context"
	"net/http"
	"strings"
	"testing"
)

func TestClientGetPullRequest(t *testing.T) {
	client := newTestClient(t, roundTripFunc(func(r *http.Request) (*http.Response, error) {
		switch r.URL.Path {
		case "/app/installations/42/access_tokens":
			assertMethod(t, r, http.MethodPost)
			return jsonResponse(t, map[string]any{
				"token":      "installation-token",
				"expires_at": "2099-01-01T00:00:00Z",
			}), nil
		case "/repos/acme/merger/pulls/7":
			assertMethod(t, r, http.MethodGet)
			if got := r.Header.Get("Accept"); got != "application/vnd.github+json" {
				t.Fatalf("unexpected accept header: %s", got)
			}
			return jsonResponse(t, map[string]any{
				"html_url": "https://github.com/acme/merger/pull/7",
				"title":    "Refactor pipeline",
				"body":     "Split handlers",
				"head":     map[string]string{"sha": "head-sha"},
				"base":     map[string]string{"sha": "base-sha"},
				"user":     map[string]string{"login": "merge-bot"},
			}), nil
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
			return nil, nil
		}
	}))
	pr, err := client.GetPullRequest(context.Background(), "acme", "merger", 7)
	if err != nil {
		t.Fatalf("get pull request: %v", err)
	}

	if pr.Title != "Refactor pipeline" || pr.Author != "merge-bot" || pr.HeadSHA != "head-sha" {
		t.Fatalf("unexpected pull request payload: %+v", pr)
	}
}

func TestClientGetPullRequestDiff(t *testing.T) {
	client := newTestClient(t, roundTripFunc(func(r *http.Request) (*http.Response, error) {
		switch r.URL.Path {
		case "/app/installations/42/access_tokens":
			return jsonResponse(t, map[string]any{
				"token":      "installation-token",
				"expires_at": "2099-01-01T00:00:00Z",
			}), nil
		case "/repos/acme/merger/pulls/9":
			if got := r.Header.Get("Accept"); got != "application/vnd.github.diff" {
				t.Fatalf("unexpected accept header: %s", got)
			}
			return textResponse("diff --git a/a.go b/a.go"), nil
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
			return nil, nil
		}
	}))
	diff, err := client.GetPullRequestDiff(context.Background(), "acme", "merger", 9)
	if err != nil {
		t.Fatalf("get pull request diff: %v", err)
	}

	if !strings.Contains(diff, "diff --git") {
		t.Fatalf("unexpected diff: %s", diff)
	}
}
