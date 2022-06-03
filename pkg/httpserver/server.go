package httpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/supreethrao/automated-rota-manager/pkg/helpers"
	"github.com/supreethrao/automated-rota-manager/pkg/rota"
	"github.com/supreethrao/automated-rota-manager/pkg/slackhandler"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	httpServerPort      = 9090
	shutdownGracePeriod = 15 * time.Second
)

func Start(_ context.Context, myTeam *rota.Team) error {
	router := httprouter.New()

	router.POST("/members/:name", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		if err := myTeam.Add(params.ByName("name")); err != nil {
			_, _ = fmt.Fprint(writer)
		}
	})

	router.GET("/members", func(writer http.ResponseWriter, request *http.Request, _ httprouter.Params) {
		writer.Header().Set("Content-Type", "application/json")
		jsonData, _ := json.Marshal(myTeam.RotaHistory())
		_, _ = writer.Write(jsonData)
	})

	router.DELETE("/members/:name", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		if err := myTeam.Remove(params.ByName("name")); err != nil {
			_, _ = fmt.Fprint(writer)
		}
	})

	router.POST("/outofoffice/:name/:from/:to", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		fromDate, errFrom := time.Parse("02-01-2006", params.ByName("from"))
		toDate, errTo := time.Parse("02-01-2006", params.ByName("to"))

		if errFrom != nil || errTo != nil {
			writer.WriteHeader(http.StatusBadRequest)
			_, _ = writer.Write([]byte("Invalid date format. From and To date should be in the format DD-MM-YYYY \n"))
			return
		}

		dateToday, _ := time.Parse("02-01-2006", time.Now().Format("02-01-2006"))
		if fromDate.After(toDate) || toDate.Before(dateToday) {
			writer.WriteHeader(http.StatusBadRequest)
			_, _ = writer.Write([]byte("Invalid date. From date cannot be greater than To date and also To date cannot be in the past"))
			return
		}

		if setError := myTeam.SetOutOfOffice(params.ByName("name"), fromDate, toDate); setError != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			// Need to do error mapping here
			_, _ = fmt.Fprintln(writer, setError)
		} else {
			writer.WriteHeader(http.StatusCreated)
		}
	})

	router.GET("/outofoffice", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		_, _ = writer.Write(myTeam.GetTeamOutOfOffice())
	})

	router.GET("/outofoffice/:name", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		_, _ = writer.Write(myTeam.GetOutOfOffice(params.ByName("name")))
	})

	router.GET("/rota/next", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		_, _ = fmt.Fprintf(writer, "The person picked today is: %s. \n", myTeam.Next())
	})

	router.GET("/rota/confirm/:name/:date", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		personPickedToday := params.ByName("name")

		if time.Now().Format("02-01-2006") != params.ByName("date") {
			_, _ = writer.Write([]byte("Illegal confirmation. Date has to be today"))
			return
		}

		if isHoliday, whichOne := helpers.IsTodayHoliday(); isHoliday {
			writer.WriteHeader(http.StatusForbidden)
			_, _ = writer.Write([]byte(fmt.Sprintf("Cheeky attempt to pick a person on a holiday. Not happening as today is %s \n", whichOne)))
			return
		}

		if err := myTeam.SetPersonPickedForToday(personPickedToday); err == nil {
			_ = slackhandler.SendMessage(fmt.Sprintf("The person picked today is confirmed to be: %s \n", personPickedToday))
			writer.WriteHeader(http.StatusAccepted)
		} else {
			_, _ = fmt.Fprintln(writer, err)
		}
	})

	router.GET("/rota/override/:name", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		personToOverrideWith := params.ByName("name")

		if isHoliday, whichOne := helpers.IsTodayHoliday(); isHoliday {
			writer.WriteHeader(http.StatusForbidden)
			_, _ = writer.Write([]byte(fmt.Sprintf("Cheeky attempt to override picking a person on a holiday. Not happening as today is %s \n", whichOne)))
			return
		}

		if err := myTeam.OverridePersonPickedForToday(personToOverrideWith); err == nil {
			_ = slackhandler.SendMessage(fmt.Sprintf("The rota pick for today was overridden. It's now: %s \n", personToOverrideWith))
			writer.WriteHeader(http.StatusAccepted)
		} else {
			_, _ = fmt.Fprintln(writer, err)
		}
	})

	router.GET("/metrics", func(writer http.ResponseWriter, request *http.Request, _ httprouter.Params) {
		promhttp.Handler().ServeHTTP(writer, request)
	})

	httpServer := &http.Server{Addr: fmt.Sprintf(":%d", httpServerPort), Handler: router}
	errChan := make(chan error, 1)

	go func() {
		err := httpServer.ListenAndServe()
		errChan <- err
	}()

	return gracefulShutdown(httpServer, errChan)
}

func gracefulShutdown(httpServer *http.Server, serverError chan error) error {
	signalHandler := make(chan os.Signal, 1)
	signal.Notify(signalHandler, syscall.SIGTERM, syscall.SIGINT)

	select {
	case err := <-serverError:
		return err
	case <-signalHandler:
		timeoutContext, cancel := context.WithTimeout(context.Background(), shutdownGracePeriod)
		defer cancel()
		return httpServer.Shutdown(timeoutContext)
	}
}
