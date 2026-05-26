package github_test

import (
	"context"
	"encoding/base64"
	"net/http"
	"testing"
)

func TestClientGetFileContent(t *testing.T) {
	client := newTestClient(t, roundTripFunc(func(r *http.Request) (*http.Response, error) {
		switch r.URL.Path {
		case "/app/installations/42/access_tokens":
			return jsonResponse(t, map[string]any{
				"token":      "installation-token",
				"expires_at": "2099-01-01T00:00:00Z",
			}), nil
		case "/repos/acme/merger/contents/config/app.yaml":
			if got := r.URL.Query().Get("ref"); got != "main" {
				t.Fatalf("unexpected ref query: %s", got)
			}
			return jsonResponse(t, map[string]string{
				"encoding": "base64",
				"content":  base64.StdEncoding.EncodeToString([]byte("name: merger\n")),
			}), nil
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
			return nil, nil
		}
	}))
	content, err := client.GetFileContent(context.Background(), "acme", "merger", "config/app.yaml", "main")
	if err != nil {
		t.Fatalf("get file content: %v", err)
	}

	if string(content) != "name: merger\n" {
		t.Fatalf("unexpected content: %q", string(content))
	}
}
