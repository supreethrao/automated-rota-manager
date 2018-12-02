package main

import (
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sky-uk/support-bot/localdb"
	"github.com/sky-uk/support-bot/rota"
	"log"
	"net/http"
)

var myTeam = rota.NewTeam("core-infrastructure")

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
		fmt.Fprint(writer, rota.Next(myTeam))
	})

	router.GET("/support/confirm/:name", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		if err := myTeam.SetPersonOnSupportForToday(params.ByName("name")); err == nil {
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

	http.Handle("/metrics", promhttp.Handler())

	log.Fatal(http.ListenAndServe(":9090", router))
}

func main() {
	defer localdb.Close()
	serve()
}
