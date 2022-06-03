package rota_test

import (
	"encoding/binary"
	"github.com/supreethrao/automated-rota-manager/pkg/rota"
	"sort"
	"time"

	"github.com/supreethrao/automated-rota-manager/pkg/localdb"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Test suite for logic of picking next", func() {

	var myTeam = rota.NewTeam("team-picker-test")

	BeforeEach(func() {
		myTeam = rota.NewTeam("test_team")
		Expect(localdb.Remove(myTeam.TeamKey())).To(Succeed())
		for _, member := range testTeamMembers {
			Expect(localdb.Remove(myTeam.AccruedDaysCounterKey(member))).To(Succeed())
			Expect(localdb.Remove(myTeam.LatestDayPickedKey(member))).To(Succeed())
			oooFrom, oooTo := myTeam.OutOfOfficeKey(member)
			Expect(localdb.Remove(oooFrom)).To(Succeed())
			Expect(localdb.Remove(oooTo)).To(Succeed())
		}
		Expect(localdb.Write(myTeam.TeamKey(), TestTeamMembersListYaml))
	})

	Context("Test sorting logic", func() {
		It("Should be sorted based on the accrued days", func() {
			teamHistory := rota.TeamRotaHistory{
				{"person1", 5, Yesterday()},
				{"person2", 3, Yesterday()},
				{"person3", 7, Yesterday()},
				{"person4", 2, Yesterday()},
			}

			expectedTeamHistory := rota.TeamRotaHistory{
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
		It("Next person is the person who has been fewer accrued days", func() {
			Expect(localdb.Write(myTeam.AccruedDaysCounterKey("person1"), Uint16ToBytes(4))).To(Succeed())
			Expect(localdb.Write(myTeam.AccruedDaysCounterKey("person2"), Uint16ToBytes(6))).To(Succeed())
			Expect(localdb.Write(myTeam.AccruedDaysCounterKey("third person"), Uint16ToBytes(3))).To(Succeed())

			Expect(localdb.Write(myTeam.LatestDayPickedKey("person1"), []byte(DaysBeforeToday(3)))).To(Succeed())
			Expect(localdb.Write(myTeam.LatestDayPickedKey("person2"), []byte(DaysBeforeToday(4)))).To(Succeed())
			Expect(localdb.Write(myTeam.LatestDayPickedKey("third person"), []byte(DaysBeforeToday(5)))).To(Succeed())

			nextPerson := myTeam.Next()
			Expect(nextPerson).To(Equal("third person"))
		})

		It("Should have a couple of days breather regardless of number of accrued days", func() {
			Expect(localdb.Write(myTeam.AccruedDaysCounterKey("person1"), Uint16ToBytes(4))).To(Succeed())
			Expect(localdb.Write(myTeam.AccruedDaysCounterKey("person2"), Uint16ToBytes(6))).To(Succeed())
			Expect(localdb.Write(myTeam.AccruedDaysCounterKey("third person"), Uint16ToBytes(3))).To(Succeed())

			Expect(localdb.Write(myTeam.LatestDayPickedKey("person1"), []byte(Yesterday()))).To(Succeed())
			Expect(localdb.Write(myTeam.LatestDayPickedKey("person2"), []byte(DaysBeforeToday(3)))).To(Succeed())
			Expect(localdb.Write(myTeam.LatestDayPickedKey("third person"), []byte(DayBeforeYesterday()))).To(Succeed())

			nextPerson := myTeam.Next()
			Expect(nextPerson).To(Equal("person2"))
		})
	})

	Context("Skip people who are out of office", func() {
		It("Skip the selected person if they are out of office", func() {
			Expect(localdb.Write(myTeam.AccruedDaysCounterKey("person1"), Uint16ToBytes(4))).To(Succeed())
			Expect(localdb.Write(myTeam.AccruedDaysCounterKey("person2"), Uint16ToBytes(6))).To(Succeed())
			Expect(localdb.Write(myTeam.AccruedDaysCounterKey("third person"), Uint16ToBytes(3))).To(Succeed())

			Expect(localdb.Write(myTeam.LatestDayPickedKey("person1"), []byte(DaysBeforeToday(3)))).To(Succeed())
			Expect(localdb.Write(myTeam.LatestDayPickedKey("person2"), []byte(DaysBeforeToday(4)))).To(Succeed())
			Expect(localdb.Write(myTeam.LatestDayPickedKey("third person"), []byte(DaysBeforeToday(5)))).To(Succeed())

			oooFrom, oooTo := myTeam.OutOfOfficeKey("third person")

			Expect(localdb.Write(oooFrom, timeToBytes(time.Now().Add(-time.Hour*24))))
			Expect(localdb.Write(oooTo, timeToBytes(time.Now().Add(time.Hour*24))))

			Expect(myTeam.Next()).To(Equal("person1"))
		})

		It("Skip the selected person who is off for the day", func() {
			Expect(localdb.Write(myTeam.AccruedDaysCounterKey("person1"), Uint16ToBytes(4))).To(Succeed())
			Expect(localdb.Write(myTeam.AccruedDaysCounterKey("person2"), Uint16ToBytes(6))).To(Succeed())
			Expect(localdb.Write(myTeam.AccruedDaysCounterKey("third person"), Uint16ToBytes(3))).To(Succeed())

			Expect(localdb.Write(myTeam.LatestDayPickedKey("person1"), []byte(DaysBeforeToday(3)))).To(Succeed())
			Expect(localdb.Write(myTeam.LatestDayPickedKey("person2"), []byte(DaysBeforeToday(4)))).To(Succeed())
			Expect(localdb.Write(myTeam.LatestDayPickedKey("third person"), []byte(DaysBeforeToday(5)))).To(Succeed())

			oooFrom, oooTo := myTeam.OutOfOfficeKey("third person")

			Expect(localdb.Write(oooFrom, timeToBytes(time.Now())))
			Expect(localdb.Write(oooTo, timeToBytes(time.Now())))

			Expect(myTeam.Next()).To(Equal("person1"))
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
