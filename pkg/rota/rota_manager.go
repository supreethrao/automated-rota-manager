package rota

import (
	"context"
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/supreethrao/automated-rota-manager/pkg/slackhandler"
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
	history, err := t.RotaHistory()
	if err != nil {
		logrus.Errorf("unable to obtain rota history: %v", err)
		return nil
	}
	return orderedList(history)
}

func (t Team) PickNextPerson(_ context.Context, slackMessager *slackhandler.Messager, ingressURL string) {
	nextPersonOnRota, err := t.Next()
	if err != nil {
		logrus.Errorf("picking next person errored with error: %v", err)
		return
	}

	message := fmt.Sprintf("The person picked for today is: %s. \n "+
		"To confirm, all you have to do is to click: %s/rota/confirm/%s/%s \n\n \n"+
		"To select a different person, click the below ordered link: \n\n %s", nextPersonOnRota, ingressURL, nextPersonOnRota, time.Now().Format("02-01-2006"), t.orderedRotaMessage(ingressURL))

	if err := slackMessager.SendMessage(message); err != nil {
		logrus.Errorf("unable to send slack message with error: %v", err)
	}
}

func (t Team) PersonPickedOnTheDay(date time.Time) string {
	personPicked, err := t.db.Read(t.PersonPickedOnDayKey(date))
	if err != nil {
		log.Printf("Unable to retrieve person picked on %v. error: %v", date, err)
		return "UNKNOWN"
	}
	return string(personPicked)
}

func (t Team) SetPersonPickedForToday(memberName string) error {
	rotaKeys := make(map[string][]byte)

	personAssignedForTheDay := t.PersonPickedOnTheDay(time.Now())

	if personAssignedForTheDay != "UNKNOWN" {
		return fmt.Errorf("%s is already assigned for the day", personAssignedForTheDay)
	}

	currentlyAccruedDays, _ := t.db.Read(t.AccruedDaysCounterKey(memberName))
	newAccruedDays := uintToBytes(bytesToUint(currentlyAccruedDays) + 1)

	rotaKeys[t.AccruedDaysCounterKey(memberName)] = newAccruedDays
	rotaKeys[t.LatestDayPickedKey(memberName)] = []byte(today())
	rotaKeys[t.PersonPickedOnDayKey(time.Now())] = []byte(memberName)
	rotaKeys[t.LatestCronRunKey()] = []byte(today())

	log.Printf("Confirming the selection for the day %q and updating the db with the new accrued number details", memberName)
	err := t.db.MultiWrite(rotaKeys)
	if err != nil {
		log.Printf("error writing to db: %v", err)
	}
	return err
}

func (t Team) OverridePersonPickedForToday(memberName string) error {
	rotaKeys := make(map[string][]byte)

	personAssignedForTheDay := t.PersonPickedOnTheDay(time.Now())

	if personAssignedForTheDay == "UNKNOWN" {
		return t.SetPersonPickedForToday(memberName)
	}

	pickedDaysAsBytes, _ := t.db.Read(t.AccruedDaysCounterKey(personAssignedForTheDay))
	adjustedPickedDays := uint16(0)
	pickedDays := bytesToUint(pickedDaysAsBytes)
	if pickedDays > 0 {
		adjustedPickedDays = pickedDays - 1
	}

	rotaKeys[t.AccruedDaysCounterKey(personAssignedForTheDay)] = uintToBytes(adjustedPickedDays)
	// This is incorrect - need to traverse through the history and get the date this person was previously picked
	rotaKeys[t.LatestDayPickedKey(personAssignedForTheDay)] = []byte("31-12-2006")

	currentlyAccruedDays, _ := t.db.Read(t.AccruedDaysCounterKey(memberName))
	newAccruedDays := uintToBytes(bytesToUint(currentlyAccruedDays) + 1)

	rotaKeys[t.AccruedDaysCounterKey(memberName)] = newAccruedDays
	rotaKeys[t.LatestDayPickedKey(memberName)] = []byte(today())
	rotaKeys[t.PersonPickedOnDayKey(time.Now())] = []byte(memberName)
	rotaKeys[t.LatestCronRunKey()] = []byte(today())

	return t.db.MultiWrite(rotaKeys)
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
