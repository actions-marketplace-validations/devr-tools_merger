package ingest

import (
	"fmt"

	"github.com/mergerhq/merger/internal/github"
)

func bindInstallation(service github.Service, installationID int64) github.Service {
	if binder, ok := service.(github.InstallationBinder); ok && installationID != 0 {
		return binder.ForInstallation(installationID)
	}
	return service
}

func (p *Processor) DescribeWebhook(payload github.PullRequestWebhookPayload) string {
	return fmt.Sprintf("%s#%d action=%s", payload.Repository.FullName, payload.PullRequest.Number, payload.Action)
}
