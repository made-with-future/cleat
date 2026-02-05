package main

import (
	"fmt"
	"os"

	"github.com/madewithfuture/cleat/internal/cmd"
	"github.com/madewithfuture/cleat/internal/config"
	"github.com/madewithfuture/cleat/internal/logger"
)

func main() {
	// Initialize logger early to catch issues during startup.
	// We include the project ID so logs from different projects in the same global file can be filtered.
	projectID := config.GetProjectID()
	if err := logger.Init("~/.cleat/cleat.log", "debug", map[string]interface{}{"project": projectID}); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to initialize logger: %v\n", err)
	}

	cmd.Execute()
}
