package cli

import (
	"flag"
	"fmt"

	"github.com/mergerhq/merger/internal/version"
)

func runVersion(args []string) error {
	fs := flag.NewFlagSet("version", flag.ContinueOnError)
	if err := fs.Parse(args); err != nil {
		return err
	}
	fmt.Printf("merger v%s\n", version.Number)
	return nil
}
