package github

import (
	"context"
	"encoding/json"
	"fmt"
)

func (c *Client) GetPullRequest(ctx context.Context, owner, repo string, number int) (PullRequest, error) {
	token, err := c.authenticator.Token(ctx, c.installationID)
	if err != nil {
		return PullRequest{}, err
	}

	body, err := c.getWithBearer(ctx, fmt.Sprintf("/repos/%s/%s/pulls/%d", owner, repo, number), token, "application/vnd.github+json")
	if err != nil {
		return PullRequest{}, err
	}

	var response struct {
		HTMLURL string `json:"html_url"`
		Title   string `json:"title"`
		Body    string `json:"body"`
		Head    struct {
			SHA string `json:"sha"`
		} `json:"head"`
		Base struct {
			SHA string `json:"sha"`
		} `json:"base"`
		User struct {
			Login string `json:"login"`
		} `json:"user"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return PullRequest{}, err
	}

	return PullRequest{
		Owner:   owner,
		Repo:    repo,
		Number:  number,
		Title:   response.Title,
		Body:    response.Body,
		Author:  response.User.Login,
		URL:     response.HTMLURL,
		HeadSHA: response.Head.SHA,
		BaseSHA: response.Base.SHA,
	}, nil
}

func (c *Client) GetPullRequestDiff(ctx context.Context, owner, repo string, number int) (string, error) {
	token, err := c.authenticator.Token(ctx, c.installationID)
	if err != nil {
		return "", err
	}

	body, err := c.getWithBearer(ctx, fmt.Sprintf("/repos/%s/%s/pulls/%d", owner, repo, number), token, "application/vnd.github.diff")
	if err != nil {
		return "", err
	}

	return string(body), nil
}
