package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/mergerhq/merger/internal/cli"
)

func main() {
	err := cli.Run(context.Background(), os.Args[1:])
	if err == nil {
		return
	}

	var exit cli.ExitError
	if errors.As(err, &exit) {
		if exit.Message != "" {
			fmt.Fprintln(os.Stderr, "merger: "+exit.Message)
		}
		os.Exit(exit.Code)
	}

	fmt.Fprintln(os.Stderr, "merger: "+err.Error())
	os.Exit(1)
}
