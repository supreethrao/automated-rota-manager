package rota

import (
	"log"
	"math"
	"sort"
	"time"
)

type IndividualSupportHistory struct {
	Name             string
	DaysSupported    uint16
	LatestSupportDay string
}

type TeamSupportHistory []IndividualSupportHistory

func (history TeamSupportHistory) Len() int {
	return len(history)
}

func (history TeamSupportHistory) Swap(i, j int) {
	history[i], history[j] = history[j], history[i]
}

func (history TeamSupportHistory) Less(i, j int) bool {
	return history[i].DaysSupported < history[j].DaysSupported
}

func differenceBetweenDays(ddmmyyyyStr1, ddmmyyyystr2 string) float64 {
	firstDay, e1 := time.Parse("02-01-2006", ddmmyyyyStr1)
	if e1 != nil {
		log.Panicf("Unable to parse date string %s - %v", ddmmyyyyStr1, e1)
	}
	secondDay, e2 := time.Parse("02-01-2006", ddmmyyyystr2)
	if e2 != nil {
		log.Panicf("Unable to parse date string %s - %v", ddmmyyyystr2, e2)
	}
	return math.Round(math.Abs(secondDay.Sub(firstDay).Hours() / 24))
}

func today() string {
	return time.Now().Format("02-01-2006")
}

func Next(t Team) string {
	teamSupportHistory := OrderedList(t.SupportHistoryForTeam())

	if teamSupportHistory.Len() < 1 {
		return "UNKNOWN"
	}

	for _, individual := range teamSupportHistory {
		if differenceBetweenDays(individual.LatestSupportDay, today()) > 2 {
			return individual.Name
		}
	}

	return teamSupportHistory[0].Name
}

func OrderedList(teamSupportHistory TeamSupportHistory) TeamSupportHistory {
	sort.Sort(teamSupportHistory)
	return teamSupportHistory
}
