package cmd

import (
	"context"
	"github.com/spf13/cobra"
	"github.com/supreethrao/automated-rota-manager/pkg/config"
	"github.com/supreethrao/automated-rota-manager/pkg/helpers"
	"github.com/supreethrao/automated-rota-manager/pkg/httpserver"
	"github.com/supreethrao/automated-rota-manager/pkg/localdb"
	"github.com/supreethrao/automated-rota-manager/pkg/rota"
	"github.com/supreethrao/automated-rota-manager/pkg/scheduler"
	"github.com/supreethrao/automated-rota-manager/pkg/slackhandler"
	"golang.org/x/sync/errgroup"
	"log"
)

var configFilePath string

var rootCmd = &cobra.Command {
	Short: "Manages fair rotation and allocation of next person in the rota",
	RunE: runRotaManager,
}

func init() {
	rootCmd.Flags().StringVarP(&configFilePath, "config", "f", "/app/config/arm-config.yaml", "config file path")
}

func runRotaManager(_ *cobra.Command, _ []string) error {
	cfg, err := config.New(configFilePath)
	if err != nil {
		return err
	}

	dbHandle, err := localdb.GetHandle()
	if err != nil {
		return err
	}
	defer dbHandle.Close()

	myTeam := rota.NewTeam(cfg.TeamName, dbHandle)

	initContext := context.Background()
	cancelContext, cancelFunc := context.WithCancel(initContext)
	defer cancelFunc()

	synGroup, synContext := errgroup.WithContext(cancelContext)

	slackToken := helpers.Getenv("SLACK_TOKEN", "unknown")
	slackConfig := slackhandler.SlackConfig{
		Token:    slackToken,
		Channel:  cfg.SlackChannel,
		UserName: cfg.SlackUserName,
	}
	slackMessager := slackhandler.NewMessager(slackConfig)

	synGroup.Go(func() error {
		return httpserver.Start(synContext, slackMessager, myTeam)
	})

	synGroup.Go(func() error {
		scheduledRotaPicker := scheduler.NewSchedule(cfg.CronSchedule, func() {
			if isHoliday, whichOne := helpers.IsTodayHoliday(); isHoliday {
				log.Printf("Today is %s and hence skipping the rota pick \n", whichOne)
			} else {
				myTeam.PickNextPerson(synContext, slackMessager, cfg.IngressURL)
			}
		})
		return scheduledRotaPicker.Schedule()
	})

	return synGroup.Wait()
}

func Exec() error {
	return rootCmd.Execute()
}
