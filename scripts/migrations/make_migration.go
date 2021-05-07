package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func main() {
	viper.Set("log-level", 4)

	log := logrus.New()
	cmd := newDefaultCommand(log)

	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

//newDefaultCommand creates the `wui` command and its nested children.
func newDefaultCommand(log *logrus.Logger) *cobra.Command {
	// Parent command to which all subcommands are added.
	cmd := &cobra.Command{
		Use:   "make_migration",
		Short: "make_migration",
		Long:  `make_migration creates new migration files`,
		Run: func(cmd *cobra.Command, args []string) {
			name, err := cmd.Flags().GetString("name")
			if err != nil {
				log.Error(err)
				return
			}
			service, err := cmd.Flags().GetString("service")
			if err != nil {
				log.Error(err)
				return
			}
			path, _ := cmd.Flags().GetString("path")

			if name == "" {
				log.Error("name argument is required")
				return
			}

			if service == "" {
				log.Error("service argument is required")
				return
			}

			err = makeMigration(name, service, path)
			if err != nil {
				log.WithError(err).Error("failed to create migration")
			}
		},
	}

	cmd.Flags().StringP("name", "n", "", "Name of the migration")
	cmd.Flags().StringP("service", "s", "", "Name of the service")
	cmd.Flags().StringP("path", "p", "", "Override output path (defaults to ./internal/{service}/db")

	return cmd
}

func makeMigration(mName, service, path string) error {
	if path == "" {
		path = fmt.Sprintf("./internal/%s/db", service)
	}

	mName = strings.ReplaceAll(mName, " ", "_")

	now := time.Now().Local()

	data := &migrationData{
		Name:   mName,
		Year:   now.Year(),
		Month:  int(now.Month()),
		Day:    now.Day(),
		Hour:   now.Hour(),
		Minute: now.Minute(),
		Second: now.Second(),
	}

	tmpl, err := getTemplate(data)
	if err != nil {
		return fmt.Errorf("compiling template: %w", err)
	}

	filename := fmt.Sprintf("%s/%d_%d_%d_%d%d%d_%s.go", path, data.Year, data.Month, data.Day, data.Hour, data.Minute, data.Second, mName)

	f, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR, os.FileMode(0664))
	if err != nil {
		return fmt.Errorf("creating migraion file: %w", err)
	}
	defer f.Close()
	_, err = f.WriteString(tmpl)

	return err
}
