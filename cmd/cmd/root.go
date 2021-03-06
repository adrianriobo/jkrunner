package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrianriobo/jkrunner/pkg/jkrunner"
	"github.com/adrianriobo/jkrunner/pkg/logging"
	"github.com/adrianriobo/jkrunner/pkg/util"
	"github.com/spf13/cobra"
	"k8s.io/utils/exec"
)

const (
	commandName      = "jkrunner"
	descriptionShort = "jenkins runner to run jobs and get artifacts"
	descriptionLong  = "jenkins runner with the only purpose to run jobs and get artifacts"

	defaultErrorExitCode = 1
)

var rootCmd = &cobra.Command{
	Use:   commandName,
	Short: descriptionShort,
	Long:  descriptionLong,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return runPrerun(cmd)
	},
	Run: func(cmd *cobra.Command, args []string) {
		runRoot()
		_ = cmd.Help()
	},
	SilenceUsage:  true,
	SilenceErrors: true,
}

func runPrerun(cmd *cobra.Command) error {
	logging.InitLogrus(
		logging.LogLevel,
		filepath.Join(util.GetHomeDir(), jkrunner.Home),
		jkrunner.LogFileName)
	return nil
}

func runRoot() {
	fmt.Println("No command given")
}

func Execute() {
	attachMiddleware([]string{}, rootCmd)

	if err := rootCmd.ExecuteContext(context.Background()); err != nil {
		runPostrun()
		_, _ = fmt.Fprintln(os.Stderr, err.Error())
		var e exec.CodeExitError
		if errors.As(err, &e) {
			os.Exit(e.ExitStatus())
		} else {
			os.Exit(defaultErrorExitCode)
		}
	}
	runPostrun()
}

func attachMiddleware(names []string, cmd *cobra.Command) {
	if cmd.HasSubCommands() {
		for _, command := range cmd.Commands() {
			attachMiddleware(append(names, cmd.Name()), command)
		}
	} else if cmd.RunE != nil {
		fullCmd := strings.Join(append(names, cmd.Name()), " ")
		src := cmd.RunE
		cmd.RunE = executeWithLogging(fullCmd, src)
	}
}

func executeWithLogging(fullCmd string, input func(cmd *cobra.Command, args []string) error) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		logging.Debugf("Running '%s'", fullCmd)
		return input(cmd, args)
	}
}

func runPostrun() {
	logging.CloseLogging()
}
