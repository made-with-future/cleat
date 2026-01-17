package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/madewithfuture/cleat/internal/config"
	"github.com/madewithfuture/cleat/internal/strategy"
	"github.com/spf13/cobra"
)

var runner = runCommand

func buildProject(cfg *config.Config) error {
	strategies := []strategy.Strategy{
		&strategy.NpmStrategy{},
		&strategy.DjangoStrategy{},
		&strategy.DockerStrategy{},
	}

	for _, s := range strategies {
		if err := s.Run(cfg, strategy.Runner(runner)); err != nil {
			return err
		}
	}

	if !cfg.Django && !cfg.Docker && len(cfg.Npm.Scripts) == 0 {
		fmt.Println("No build steps defined for this project type in cleat.yaml")
	} else {
		fmt.Println("==> Build completed successfully")
	}

	return nil
}

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build the project based on cleat.yaml",
	Long:  `Executes build steps based on the project configuration in cleat.yaml. Supports Docker and Django project types.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig("cleat.yaml")
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("no cleat.yaml found in current directory")
			}
			return fmt.Errorf("error loading config: %w", err)
		}

		return buildProject(cfg)
	},
}

func runCommand(name string, args ...string) error {
	fmt.Printf("Executing: %s %s\n", name, strings.Join(args, " "))
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func init() {
	rootCmd.AddCommand(buildCmd)
}
