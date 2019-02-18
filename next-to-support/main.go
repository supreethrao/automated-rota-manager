package main

import (
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sky-uk/support-bot/helpers"
	"github.com/sky-uk/support-bot/localdb"
	"github.com/sky-uk/support-bot/rota"
	"github.com/sky-uk/support-bot/rota/slackhandler"
	"github.com/sky-uk/support-bot/scheduler"
	"log"
	"net/http"
	"time"
)

var myTeam = rota.NewTeam("core-infrastructure")

var dailySupportPicker = scheduler.NewSchedule("0 0 10 * * 1-5", func() {
	if isHoliday, whichOne := helpers.IsTodayHoliday(); isHoliday {
		log.Printf("Today is %s and hence skipping the support pick \n", whichOne)
	} else {
		pickNextSupportPerson()
	}
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

	router.POST("/outofoffice/:name/:from/:to", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		fromDate, errFrom := time.Parse("02-01-2006", params.ByName("from"))
		toDate, errTo := time.Parse("02-01-2006", params.ByName("to"))

		if errFrom != nil || errTo != nil {
			writer.WriteHeader(http.StatusBadRequest)
			writer.Write([]byte("Invalid date format. From and To date should be in the format DD-MM-YYYY \n"))
			return
		}

		dateToday, _ := time.Parse("02-01-2006", time.Now().Format("02-01-2006"))
		if fromDate.After(toDate) || toDate.Before(dateToday) {
			writer.WriteHeader(http.StatusBadRequest)
			writer.Write([]byte("Invalid date. From date cannot be greater than To date and also To date cannot be in the past"))
			return
		}

		if setError := myTeam.SetOutOfOffice(params.ByName("name"), fromDate, toDate); setError != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			// Need to do error mapping here
			fmt.Fprintln(writer, setError)
		} else {
			writer.WriteHeader(http.StatusCreated)
		}
	})

	router.GET("/outofoffice", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		writer.Write(myTeam.GetTeamOutOfOffice())
	})

	router.GET("/outofoffice/:name", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		writer.Write(myTeam.GetOutOfOffice(params.ByName("name")))
	})

	router.GET("/support/next", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		fmt.Fprintf(writer, "The person chosen to be on support today is: %s. \n", rota.Next(myTeam))
	})

	router.GET("/support/confirm/:name/:date", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		supportPerson := params.ByName("name")

		if time.Now().Format("02-01-2006") != params.ByName("date") {
			writer.Write([]byte("Illegal confirmation. Date has to be today"))
			return
		}

		if isHoliday, whichOne := helpers.IsTodayHoliday(); isHoliday {
			writer.WriteHeader(http.StatusForbidden)
			writer.Write([]byte(fmt.Sprintf("Cheeky attempt to set support for a holiday. Not happening as today is %s \n", whichOne)))
			return
		}

		if err := myTeam.SetPersonOnSupportForToday(supportPerson); err == nil {
			slackhandler.SendMessage(fmt.Sprintf("The person on support for today is confirmed to be: %s \n", supportPerson))
			writer.WriteHeader(http.StatusAccepted)
		} else {
			fmt.Fprintln(writer, err)
		}
	})

	router.GET("/support/override/:name", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		supportPerson := params.ByName("name")

		if isHoliday, whichOne := helpers.IsTodayHoliday(); isHoliday {
			writer.WriteHeader(http.StatusForbidden)
			writer.Write([]byte(fmt.Sprintf("Cheeky attempt to set support for a holiday. Not happening as today is %s \n", whichOne)))
			return
		}

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

func pickNextSupportPerson() {
	nextToSupport := rota.Next(myTeam)

	message := fmt.Sprintf("The person chosen to be on support today is: %s. \n "+
		"To confirm, all you have to do is to click: http://support-bot.dev.cosmic.sky/support/confirm/%s/%s \n\n \n"+
		"To select a different person, click the below ordered link: \n\n %s", nextToSupport, nextToSupport, time.Now().Format("02-01-2006"), orderedRotaMessage())

	if err := slackhandler.SendMessage(message); err != nil {
		log.Panic("Unable to send slack message", err)
	}
}

func orderedRotaMessage() string {
	orderedRota := ""
	today := time.Now().Format("02-01-2006")

	for ind, member := range rota.OrderedRota(myTeam) {
		orderedRota += fmt.Sprintf("%d. http://support-bot.dev.cosmic.sky/support/confirm/%s/%s \n", ind+1, member.Name, today)
	}

	return orderedRota
}

func main() {
	defer localdb.Close()
	dailySupportPicker.Schedule()
	serve()
}
