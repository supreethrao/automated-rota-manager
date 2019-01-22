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
			slackhandler.SendMessage(fmt.Sprintf("The person to be on support today is confirmed to be: %s \n", supportPerson))
			writer.WriteHeader(http.StatusAccepted)
		} else {
			fmt.Fprintln(writer, err)
		}
	})

	router.GET("/support/override/:name", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		if err := myTeam.OverrideSupportPersonForToday(params.ByName("name")); err == nil {
			writer.WriteHeader(http.StatusAccepted)
		} else {
			fmt.Fprintln(writer, err)
		}
	})

	router.GET("/reset", func(writer http.ResponseWriter, request *http.Request, _ httprouter.Params) {
		myTeam.Reset()
		writer.WriteHeader(http.StatusAccepted)
	})

	router.GET("/bot/next", func(writer http.ResponseWriter, request *http.Request, _ httprouter.Params) {
		pickNextSupportPerson()
		writer.WriteHeader(http.StatusNoContent)
	})

	http.Handle("/metrics", promhttp.Handler())

	log.Fatal(http.ListenAndServe(":9090", router))
}

func pickNextSupportPerson() error {
	nextToSupport := rota.Next(myTeam)
	message := fmt.Sprintf("The person chosen to be on support today is: %s. \n " +
		"To confirm, all you have to do is to click: http://support-bot.dev.cosmic.sky/support/confirm/%s \n\n " +
		"To select a different person, use the link http://support-bot.dev.cosmic.sky/support/confirm/<name> \n\n where: " +
		"<name> is one of: Supreeth, Isaac, Matt, Anthony, Pete, Howard, Yorg or Dom.", nextToSupport, nextToSupport)

	return slackhandler.SendMessage(message)
}

func main() {
	defer localdb.Close()
	dailySupportPicker.Schedule()
	serve()
}
