package helpers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const location string = "england-and-wales"

type locationSpecificHolidays struct {
	Division string
	Events   []event
}

type event struct {
	Title   string
	Date    string
	Notes   string
	Bunting bool
}

var holidaysThisYear = map[string]string{}

func init() {
	resp, err := http.Get("https://www.gov.uk/bank-holidays.json")

	if err != nil {
		log.Fatalf("Unable to obtain holidays list. Quitting")
	}

	defer func() {
		_ = resp.Body.Close()
	}()
	body, _ := ioutil.ReadAll(resp.Body)

	placeHolder := map[string]locationSpecificHolidays{}
	currentYear := strconv.Itoa(time.Now().Year())

	if er := json.Unmarshal(body, &placeHolder); er == nil {
		for _, ev := range placeHolder[location].Events {
			if strings.Contains(ev.Date, currentYear+"-") {
				splitDate := strings.Split(ev.Date, "-")
				formattedSplitDate := []string{splitDate[2], splitDate[1], splitDate[0]}
				holidaysThisYear[strings.Join(formattedSplitDate, "-")] = ev.Title
			}
		}
	} else {
		fmt.Println(er)
	}
}

func IsTodayHoliday() (bool, string) {
	today := time.Now()
	dateToday := today.Format("02-01-2006")

	if today.Weekday() == time.Saturday || today.Weekday() == time.Sunday {
		return true, "Weekend"
	}

	val, ok := holidaysThisYear[dateToday]
	return ok, val
}
