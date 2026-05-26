package github

import (
	"context"
	"fmt"
)

func (c *Client) PublishCheckRun(ctx context.Context, input CheckRunInput) error {
	token, err := c.authenticator.Token(ctx, c.installationID)
	if err != nil {
		return err
	}

	payload := map[string]any{
		"name":       input.Name,
		"head_sha":   input.HeadSHA,
		"status":     input.Status,
		"conclusion": input.Conclusion,
		"output": map[string]string{
			"title":   input.Name,
			"summary": input.Summary,
		},
	}
	if input.DetailsURL != "" {
		payload["details_url"] = input.DetailsURL
	}

	_, err = c.postWithBearer(ctx, fmt.Sprintf("/repos/%s/%s/check-runs", input.RepoOwner, input.RepoName), token, payload)
	return err
}
