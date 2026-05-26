package ingest

import (
	"path/filepath"
	"strings"

	"github.com/mergerhq/merger/internal/domain"
	"github.com/mergerhq/merger/pkg/diff"
)

func mapChangedFiles(files []diff.File) []domain.ChangedFile {
	mapped := make([]domain.ChangedFile, 0, len(files))
	for _, file := range files {
		mapped = append(mapped, domain.ChangedFile{
			Path:         file.Path,
			PreviousPath: file.PreviousPath,
			Status:       domain.FileStatus(file.Status),
			Language:     languageFromPath(file.Path),
			Additions:    file.Additions,
			Deletions:    file.Deletions,
			Changes:      file.Additions + file.Deletions,
			Patch:        file.Patch,
		})
	}
	return mapped
}

func languageFromPath(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".go":
		return "go"
	case ".sql":
		return "sql"
	case ".yaml", ".yml":
		return "yaml"
	case ".proto":
		return "proto"
	default:
		return "unknown"
	}
}
