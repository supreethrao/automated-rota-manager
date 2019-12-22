package rota_test

import (
	"encoding/binary"
	"sort"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/supreethrao/support-bot/localdb"
	"github.com/supreethrao/support-bot/rota"
)

var _ = Describe("Test suite for logic of picking next", func() {

	var myTeam = rota.NewTeam("team-picker-test")

	BeforeEach(func() {
		myTeam = rota.NewTeam("test_team")
		Expect(localdb.Remove(myTeam.TeamKey())).To(Succeed())
		for _, member := range TestTeamMembers {
			Expect(localdb.Remove(myTeam.SupportDaysCounterKey(member))).To(Succeed())
			Expect(localdb.Remove(myTeam.LatestDayOnSupportKey(member))).To(Succeed())
			oooFrom, oooTo := myTeam.OutOfOfficeKey(member)
			Expect(localdb.Remove(oooFrom)).To(Succeed())
			Expect(localdb.Remove(oooTo)).To(Succeed())
		}
		Expect(localdb.Write(myTeam.TeamKey(), TestTeamMembersListYaml))
	})

	Context("Test sorting logic", func() {
		It("Should be sorted based on the number of supported days", func() {
			teamHistory := rota.TeamSupportHistory{
				{"person1", 5, Yesterday()},
				{"person2", 3, Yesterday()},
				{"person3", 7, Yesterday()},
				{"person4", 2, Yesterday()},
			}

			expectedTeamHistory := rota.TeamSupportHistory{
				{"person4", 2, Yesterday()},
				{"person2", 3, Yesterday()},
				{"person1", 5, Yesterday()},
				{"person3", 7, Yesterday()},
			}

			sort.Sort(teamHistory)
			Expect(teamHistory).To(Equal(expectedTeamHistory))
		})
	})

	Context("Test picking based on fair rotation", func() {
		It("Next person is the person who has been on fewer support days", func() {
			Expect(localdb.Write(myTeam.SupportDaysCounterKey("person1"), Uint16ToBytes(4))).To(Succeed())
			Expect(localdb.Write(myTeam.SupportDaysCounterKey("person2"), Uint16ToBytes(6))).To(Succeed())
			Expect(localdb.Write(myTeam.SupportDaysCounterKey("third person"), Uint16ToBytes(3))).To(Succeed())

			Expect(localdb.Write(myTeam.LatestDayOnSupportKey("person1"), []byte(DaysBeforeToday(3)))).To(Succeed())
			Expect(localdb.Write(myTeam.LatestDayOnSupportKey("person2"), []byte(DaysBeforeToday(4)))).To(Succeed())
			Expect(localdb.Write(myTeam.LatestDayOnSupportKey("third person"), []byte(DaysBeforeToday(5)))).To(Succeed())

			nextSupportPerson := rota.Next(myTeam)
			Expect(nextSupportPerson).To(Equal("third person"))
		})

		It("Should have a couple of days breather regardless of number of support days", func() {
			Expect(localdb.Write(myTeam.SupportDaysCounterKey("person1"), Uint16ToBytes(4))).To(Succeed())
			Expect(localdb.Write(myTeam.SupportDaysCounterKey("person2"), Uint16ToBytes(6))).To(Succeed())
			Expect(localdb.Write(myTeam.SupportDaysCounterKey("third person"), Uint16ToBytes(3))).To(Succeed())

			Expect(localdb.Write(myTeam.LatestDayOnSupportKey("person1"), []byte(Yesterday()))).To(Succeed())
			Expect(localdb.Write(myTeam.LatestDayOnSupportKey("person2"), []byte(DaysBeforeToday(3)))).To(Succeed())
			Expect(localdb.Write(myTeam.LatestDayOnSupportKey("third person"), []byte(DayBeforeYesterday()))).To(Succeed())

			nextSupportPerson := rota.Next(myTeam)
			Expect(nextSupportPerson).To(Equal("person2"))
		})
	})

	Context("Skip people who are out of office", func() {
		It("Skip the selected person if they are out of office", func() {
			Expect(localdb.Write(myTeam.SupportDaysCounterKey("person1"), Uint16ToBytes(4))).To(Succeed())
			Expect(localdb.Write(myTeam.SupportDaysCounterKey("person2"), Uint16ToBytes(6))).To(Succeed())
			Expect(localdb.Write(myTeam.SupportDaysCounterKey("third person"), Uint16ToBytes(3))).To(Succeed())

			Expect(localdb.Write(myTeam.LatestDayOnSupportKey("person1"), []byte(DaysBeforeToday(3)))).To(Succeed())
			Expect(localdb.Write(myTeam.LatestDayOnSupportKey("person2"), []byte(DaysBeforeToday(4)))).To(Succeed())
			Expect(localdb.Write(myTeam.LatestDayOnSupportKey("third person"), []byte(DaysBeforeToday(5)))).To(Succeed())

			oooFrom, oooTo := myTeam.OutOfOfficeKey("third person")

			Expect(localdb.Write(oooFrom, timeToBytes(time.Now().Add(-time.Hour*24))))
			Expect(localdb.Write(oooTo, timeToBytes(time.Now().Add(time.Hour*24))))

			Expect(rota.Next(myTeam)).To(Equal("person1"))
		})

		It("Skip the selected person who is off for the day", func() {
			Expect(localdb.Write(myTeam.SupportDaysCounterKey("person1"), Uint16ToBytes(4))).To(Succeed())
			Expect(localdb.Write(myTeam.SupportDaysCounterKey("person2"), Uint16ToBytes(6))).To(Succeed())
			Expect(localdb.Write(myTeam.SupportDaysCounterKey("third person"), Uint16ToBytes(3))).To(Succeed())

			Expect(localdb.Write(myTeam.LatestDayOnSupportKey("person1"), []byte(DaysBeforeToday(3)))).To(Succeed())
			Expect(localdb.Write(myTeam.LatestDayOnSupportKey("person2"), []byte(DaysBeforeToday(4)))).To(Succeed())
			Expect(localdb.Write(myTeam.LatestDayOnSupportKey("third person"), []byte(DaysBeforeToday(5)))).To(Succeed())

			oooFrom, oooTo := myTeam.OutOfOfficeKey("third person")

			Expect(localdb.Write(oooFrom, timeToBytes(time.Now())))
			Expect(localdb.Write(oooTo, timeToBytes(time.Now())))

			Expect(rota.Next(myTeam)).To(Equal("person1"))
		})
	})
})

func Today() string {
	return time.Now().Format("02-01-2006")
}

func Yesterday() string {
	return time.Now().AddDate(0, 0, -1).Format("02-01-2006")
}

func DaysBeforeToday(num int) string {
	return time.Now().AddDate(0, 0, -num).Format("02-01-2006")
}

func DayBeforeYesterday() string {
	return time.Now().AddDate(0, 0, -2).Format("02-01-2006")
}

func timeToBytes(t time.Time) []byte {
	return []byte(t.Format("02-01-2006"))
}

func Uint16ToBytes(intVal uint16) []byte {
	byteVal := make([]byte, 2)
	binary.BigEndian.PutUint16(byteVal, intVal)
	return byteVal
}
