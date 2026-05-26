package github

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"path"
	"strings"
)

func (c *Client) GetFileContent(ctx context.Context, owner, repo, filePath, ref string) ([]byte, error) {
	token, err := c.authenticator.Token(ctx, c.installationID)
	if err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("/repos/%s/%s/contents/%s?ref=%s", owner, repo, path.Clean(filePath), url.QueryEscape(ref))
	body, err := c.getWithBearer(ctx, endpoint, token, "application/vnd.github+json")
	if err != nil {
		return nil, err
	}

	var response struct {
		Content  string `json:"content"`
		Encoding string `json:"encoding"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	if !strings.EqualFold(response.Encoding, "base64") {
		return nil, fmt.Errorf("unsupported GitHub content encoding: %s", response.Encoding)
	}

	normalized := strings.ReplaceAll(response.Content, "\n", "")
	return base64.StdEncoding.DecodeString(normalized)
}
