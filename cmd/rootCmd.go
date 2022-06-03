package cmd

import (
	"context"
	"github.com/spf13/cobra"
	"github.com/supreethrao/automated-rota-manager/pkg/helpers"
	"github.com/supreethrao/automated-rota-manager/pkg/httpserver"
	"github.com/supreethrao/automated-rota-manager/pkg/rota"
	"github.com/supreethrao/automated-rota-manager/pkg/scheduler"
	"golang.org/x/sync/errgroup"
	"log"
)

var teamName string

var rootCmd = &cobra.Command {
	Short: "Manages fair rotation and allocation of next person in the rota",
	RunE: runRotaManager,
}

func init() {
	rootCmd.Flags().StringVarP(&teamName, "team-name", "t", "", "team name for which the rota is managed for")
	_ = rootCmd.MarkFlagRequired("team-name")
}


func runRotaManager(_ *cobra.Command, _ []string) error {
	myTeam := rota.NewTeam(teamName)

	initContext := context.Background()
	cancelContext, cancelFunc := context.WithCancel(initContext)
	defer cancelFunc()

	synGroup, synContext := errgroup.WithContext(cancelContext)

	synGroup.Go(func() error {
		return httpserver.Start(synContext, myTeam)
	})

	synGroup.Go(func() error {
		cronSchedule := helpers.Getenv("CRON_SCHEDULE", "0 0 10 * * 1-5")
		scheduledRotaPicker := scheduler.NewSchedule(cronSchedule, func() {
			if isHoliday, whichOne := helpers.IsTodayHoliday(); isHoliday {
				log.Printf("Today is %s and hence skipping the rota pick \n", whichOne)
			} else {
				myTeam.PickNextPerson(synContext)
			}
		})
		return scheduledRotaPicker.Schedule()
	})

	return synGroup.Wait()
}

func Exec() error {
	return rootCmd.Execute()
}
