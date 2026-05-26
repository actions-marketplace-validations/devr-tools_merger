package diff

import (
	"strings"
)

type Status string

const (
	StatusAdded    Status = "added"
	StatusModified Status = "modified"
	StatusRemoved  Status = "removed"
	StatusRenamed  Status = "renamed"
)

type File struct {
	Path         string
	PreviousPath string
	Status       Status
	Additions    int
	Deletions    int
	Patch        string
}

func ParseUnified(input string) ([]File, error) {
	if strings.TrimSpace(input) == "" {
		return nil, nil
	}

	lines := strings.Split(input, "\n")
	files := make([]File, 0)
	var current *File
	var patch []string

	flush := func() {
		if current == nil {
			return
		}
		current.Patch = strings.Join(patch, "\n")
		files = append(files, *current)
		current = nil
		patch = nil
	}

	for _, line := range lines {
		if strings.HasPrefix(line, "diff --git ") {
			flush()
			current = &File{Status: StatusModified}
			parts := strings.Split(line, " ")
			if len(parts) >= 4 {
				current.PreviousPath = strings.TrimPrefix(parts[2], "a/")
				current.Path = strings.TrimPrefix(parts[3], "b/")
			}
			patch = append(patch, line)
			continue
		}

		if current == nil {
			continue
		}

		patch = append(patch, line)

		switch {
		case strings.HasPrefix(line, "new file mode"):
			current.Status = StatusAdded
		case strings.HasPrefix(line, "deleted file mode"):
			current.Status = StatusRemoved
		case strings.HasPrefix(line, "rename from "):
			current.PreviousPath = strings.TrimPrefix(line, "rename from ")
			current.Status = StatusRenamed
		case strings.HasPrefix(line, "rename to "):
			current.Path = strings.TrimPrefix(line, "rename to ")
			current.Status = StatusRenamed
		case strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++"):
			current.Additions++
		case strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---"):
			current.Deletions++
		}
	}

	flush()
	return files, nil
}
