package rota_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/support-bot/localdb"
	"github.com/sky-uk/support-bot/rota"
	"github.com/sky-uk/support-bot/rota_test/helper"
	"sort"
	"testing"
)

func TestPicker(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Test suite for picking logic")
}

var _ = Describe("Test suite for logic of picking next", func() {

	var myTeam = rota.NewTeam("team-picker-test")

	BeforeEach(func() {
		myTeam = rota.NewTeam("test_team")
		Expect(localdb.Remove(myTeam.TeamKey())).To(Succeed())
		for _, member := range helper.TestTeamMembers {
			Expect(localdb.Remove(myTeam.SupportDaysCounterKey(member))).To(Succeed())
			Expect(localdb.Remove(myTeam.LatestDayOnSupportKey(member))).To(Succeed())
		}
		Expect(localdb.Write(myTeam.TeamKey(), helper.TestTeamMembersListYaml))
	})

	Context("Test sorting logic", func() {
		It("Should be sorted based on the number of supported days", func() {
			teamHistory := rota.TeamSupportHistory{
				{"person1", 5, helper.Yesterday()},
				{"person2", 3, helper.Yesterday()},
				{"person3", 7, helper.Yesterday()},
				{"person4", 2, helper.Yesterday()},
			}

			expectedTeamHistory := rota.TeamSupportHistory{
				{"person4", 2, helper.Yesterday()},
				{"person2", 3, helper.Yesterday()},
				{"person1", 5, helper.Yesterday()},
				{"person3", 7, helper.Yesterday()},
			}

			sort.Sort(teamHistory)
			Expect(teamHistory).To(Equal(expectedTeamHistory))
		})
	})

	Context("Test picking based on fair rotation", func() {
		It("Next person is the person who has been on fewer support days", func() {
			Expect(localdb.Write(myTeam.SupportDaysCounterKey("person1"), helper.Uint16ToBytes(4))).To(Succeed())
			Expect(localdb.Write(myTeam.SupportDaysCounterKey("person2"), helper.Uint16ToBytes(6))).To(Succeed())
			Expect(localdb.Write(myTeam.SupportDaysCounterKey("third person"), helper.Uint16ToBytes(3))).To(Succeed())

			Expect(localdb.Write(myTeam.LatestDayOnSupportKey("person1"), []byte(helper.DaysBeforeToday(3)))).To(Succeed())
			Expect(localdb.Write(myTeam.LatestDayOnSupportKey("person2"), []byte(helper.DaysBeforeToday(4)))).To(Succeed())
			Expect(localdb.Write(myTeam.LatestDayOnSupportKey("third person"), []byte(helper.DaysBeforeToday(5)))).To(Succeed())


			nextSupportPerson := rota.Next(myTeam)
			Expect(nextSupportPerson).To(Equal("third person"))
		})

		It("Should have a couple of days breather regardless of number of support days", func() {
			Expect(localdb.Write(myTeam.SupportDaysCounterKey("person1"), helper.Uint16ToBytes(4))).To(Succeed())
			Expect(localdb.Write(myTeam.SupportDaysCounterKey("person2"), helper.Uint16ToBytes(6))).To(Succeed())
			Expect(localdb.Write(myTeam.SupportDaysCounterKey("third person"), helper.Uint16ToBytes(3))).To(Succeed())

			Expect(localdb.Write(myTeam.LatestDayOnSupportKey("person1"), []byte(helper.Yesterday()))).To(Succeed())
			Expect(localdb.Write(myTeam.LatestDayOnSupportKey("person2"), []byte(helper.DaysBeforeToday(3)))).To(Succeed())
			Expect(localdb.Write(myTeam.LatestDayOnSupportKey("third person"), []byte(helper.DayBeforeYesterday()))).To(Succeed())

			nextSupportPerson := rota.Next(myTeam)
			Expect(nextSupportPerson).To(Equal("person2"))
		})
	})
})
