package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/vladlosev/node-relabeler/pkg/kube"
	"github.com/vladlosev/node-relabeler/pkg/specs"
)

var relabelOptions []string = nil
var logLevel string

// NewWorkerCommand returns a new command that will keep relabeling nodes
// matching the spec, forever.
func NewWorkerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "worker",
		Short: "Relabel nodes according to spec",
		RunE:  startRelabeler,
	}

	cmd.PersistentFlags().StringArrayVar(
		&relabelOptions,
		"relabel",
		[]string{},
		"Re-labeling specs in the form old/label=value:new/label=newvalue",
	)
	cmd.PersistentFlags().StringVar(
		&logLevel,
		"log-level",
		"info",
		"Log level. One of: error, warn, info, debug",
	)
	return cmd
}

func startRelabeler(cmd *cobra.Command, args []string) error {
	var logrusLevel logrus.Level
	switch logLevel {
	case "error":
		logrusLevel = logrus.ErrorLevel
	case "warn":
		logrusLevel = logrus.WarnLevel
	case "info":
		logrusLevel = logrus.InfoLevel
	case "debug":
		logrusLevel = logrus.DebugLevel
	default:
		return fmt.Errorf("Invalid log level: %s", logLevel)
	}
	logrus.SetLevel(logrusLevel)

	parsedSpecs, err := specs.Parse(relabelOptions)
	if err != nil {
		return err
	}

	client, err := kube.GetKubernetesClient()
	if err != nil {
		return err
	}
	signals := make(chan os.Signal)
	stop := make(chan struct{})

	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-signals
		logrus.WithField("signal", sig).Info("Received signal, exiting")
		close(stop)
	}()

	controller, err := kube.NewController(client, parsedSpecs)
	if err != nil {
		return err
	}
	return controller.Run(stop)
}
