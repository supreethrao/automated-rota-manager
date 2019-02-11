package main

import (
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sky-uk/support-bot/localdb"
	"github.com/sky-uk/support-bot/rota"
	"github.com/sky-uk/support-bot/rota/slackhandler"
	"github.com/sky-uk/support-bot/scheduler"
	"log"
	"net/http"
)

var myTeam = rota.NewTeam("core-infrastructure")

var dailySupportPicker = scheduler.NewSchedule("0 0 10 * * 1-5", func() {
	pickNextSupportPerson()
})

func serve() {
	router := httprouter.New()

	router.GET("/members", func(writer http.ResponseWriter, request *http.Request, _ httprouter.Params) {
		writer.Header().Set("Content-Type", "application/json")
		jsonData, _ := json.Marshal(myTeam.SupportHistoryForTeam())
		writer.Write(jsonData)
	})

	router.DELETE("/members/:name", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		if err := myTeam.Remove(params.ByName("name")); err != nil {
			_, _ = fmt.Fprint(writer, )
		}
	})

	router.POST("/members/:name", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		if err := myTeam.Add(params.ByName("name")); err != nil {
			_, _ = fmt.Fprint(writer, )
		}
	})

	router.GET("/support/next", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		fmt.Fprintf(writer, "The person chosen to be on support today is: %s. \n", rota.Next(myTeam))
	})

	router.GET("/support/confirm/:name", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		supportPerson := params.ByName("name")
		if err := myTeam.SetPersonOnSupportForToday(supportPerson); err == nil {
			slackhandler.SendMessage(fmt.Sprintf("The person on support for today is confirmed to be: %s \n", supportPerson))
			writer.WriteHeader(http.StatusAccepted)
		} else {
			fmt.Fprintln(writer, err)
		}
	})

	router.GET("/support/override/:name", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		supportPerson := params.ByName("name")
		if err := myTeam.OverrideSupportPersonForToday(supportPerson); err == nil {
			slackhandler.SendMessage(fmt.Sprintf("The person support for today was overridden. It's now: %s \n", supportPerson))
			writer.WriteHeader(http.StatusAccepted)
		} else {
			fmt.Fprintln(writer, err)
		}
	})

	http.Handle("/metrics", promhttp.Handler())

	log.Fatal(http.ListenAndServe(":9090", router))
}

func pickNextSupportPerson() error {
	nextToSupport := rota.Next(myTeam)

	message := fmt.Sprintf("The person chosen to be on support today is: %s. \n "+
		"To confirm, all you have to do is to click: http://support-bot.dev.cosmic.sky/support/confirm/%s \n\n \n"+
		"To select a different person, click the below ordered link: \n\n %s", nextToSupport, nextToSupport, orderedRotaMessage())

	return slackhandler.SendMessage(message)
}

func orderedRotaMessage() string {
	orderedRota := ""

	for ind, member := range rota.OrderedRota(myTeam) {
		orderedRota += fmt.Sprintf("%d. http://support-bot.dev.cosmic.sky/support/confirm/%s \n", ind+1, member.Name)
	}

	return orderedRota
}

func main() {
	defer localdb.Close()
	dailySupportPicker.Schedule()
	serve()
}
