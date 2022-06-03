package rota

import (
	"context"
	"fmt"
	"github.com/supreethrao/automated-rota-manager/pkg/helpers"
	"github.com/supreethrao/automated-rota-manager/pkg/localdb"
	"github.com/supreethrao/automated-rota-manager/pkg/slackhandler"
	"log"
	"sort"
	"time"
)

type IndividualHistory struct {
	Name            string
	DaysAccrued     uint16
	LatestPickedDay string
}

type TeamRotaHistory []IndividualHistory

func (history TeamRotaHistory) Len() int {
	return len(history)
}

func (history TeamRotaHistory) Swap(i, j int) {
	history[i], history[j] = history[j], history[i]
}

func (history TeamRotaHistory) Less(i, j int) bool {
	return history[i].DaysAccrued < history[j].DaysAccrued
}

func (t Team) OrderedRota() []IndividualHistory {
	return orderedList(t.RotaHistory())
}

func (t Team) PickNextPerson(_ context.Context) {
	nextPersonOnRota := t.Next()
	host := helpers.Getenv("AUTOMATED_ROTA_MANAGER_HOST", "UNKNOWN_HOST")

	message := fmt.Sprintf("The person picked for today is: %s. \n "+
		"To confirm, all you have to do is to click: %s/rota/confirm/%s/%s \n\n \n"+
		"To select a different person, click the below ordered link: \n\n %s", nextPersonOnRota, host, nextPersonOnRota, time.Now().Format("02-01-2006"), t.orderedRotaMessage(host))

	if err := slackhandler.SendMessage(message); err != nil {
		log.Panic("Unable to send slack message", err)
	}
}

func (t Team) PersonPickedOnTheDay(date time.Time) string {
	personPicked, err := localdb.Read(t.PersonPickedOnDayKey(date))
	if err == nil {
		return string(personPicked)
	}

	log.Printf("Unable to retrieve person picked on %v", date)
	return "UNKNOWN"
}

func (t Team) SetPersonPickedForToday(memberName string) error {
	rotaKeys := make(map[string][]byte)

	personAssignedForTheDay := t.PersonPickedOnTheDay(time.Now())

	if personAssignedForTheDay != "UNKNOWN" {
		return fmt.Errorf("%s is already assigned for the day", personAssignedForTheDay)
	}

	currentlyAccruedDays, _ := localdb.Read(t.AccruedDaysCounterKey(memberName))
	newAccruedDays := uintToBytes(bytesToUint(currentlyAccruedDays) + 1)

	rotaKeys[t.AccruedDaysCounterKey(memberName)] = newAccruedDays
	rotaKeys[t.LatestDayPickedKey(memberName)] = []byte(today())
	rotaKeys[t.PersonPickedOnDayKey(time.Now())] = []byte(memberName)

	return localdb.MultiWrite(rotaKeys)
}

func (t Team) OverridePersonPickedForToday(memberName string) error {
	rotaKeys := make(map[string][]byte)

	personAssignedForTheDay := t.PersonPickedOnTheDay(time.Now())

	if personAssignedForTheDay == "UNKNOWN" {
		return t.SetPersonPickedForToday(memberName)
	}

	pickedDaysAsBytes, _ := localdb.Read(t.AccruedDaysCounterKey(personAssignedForTheDay))
	adjustedPickedDays := uint16(0)
	pickedDays := bytesToUint(pickedDaysAsBytes)
	if pickedDays > 0 {
		adjustedPickedDays = pickedDays - 1
	}

	rotaKeys[t.AccruedDaysCounterKey(personAssignedForTheDay)] = uintToBytes(adjustedPickedDays)
	// This is incorrect - need to traverse through the history and get the date this person was previously picked
	rotaKeys[t.LatestDayPickedKey(personAssignedForTheDay)] = []byte("31-12-2006")

	currentlyAccruedDays, _ := localdb.Read(t.AccruedDaysCounterKey(memberName))
	newAccruedDays := uintToBytes(bytesToUint(currentlyAccruedDays) + 1)

	rotaKeys[t.AccruedDaysCounterKey(memberName)] = newAccruedDays
	rotaKeys[t.LatestDayPickedKey(memberName)] = []byte(today())
	rotaKeys[t.PersonPickedOnDayKey(time.Now())] = []byte(memberName)

	return localdb.MultiWrite(rotaKeys)
}

func orderedList(teamRotaHistory TeamRotaHistory) TeamRotaHistory {
	sort.Sort(teamRotaHistory)
	return teamRotaHistory
}

func (t Team) orderedRotaMessage(host string) string {
	orderedRota := ""
	today := time.Now().Format("02-01-2006")

	for ind, member := range t.OrderedRota() {
		orderedRota += fmt.Sprintf("%d. %s/rota/confirm/%s/%s \n", ind+1, host, member.Name, today)
	}

	return orderedRota
}
