package CovidSim

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"math/rand"
	"strconv"
	"time"
)

const (
	// ContractionChance is the percent chance we contract covid.
	ContractionChance float32 = 20
)

type CovidUser struct {
	ID                string    `json:"id" gorm:"index:ux_coviduser_id,unique"`
	IntID             int64     `json:"int_id"`
	CovidEncounter    int       `json:"covid_encounter"`
	ContractionTime   time.Time `json:"contraction_time"`
	ContractedBy      string    `json:"contracted_by"`
	ContractedServer  string    `json:"contracted_server"`
	ContractedChannel string    `json:"contracted_channel"`
	ContractedMessage string    `json:"contracted_message"`
}

func NewCovidUser(user *discordgo.User) (CovidUser, error) {
	userID, err := strconv.ParseInt(user.ID, 10, 64)
	if err != nil {
		return CovidUser{}, fmt.Errorf("ContractCovid: invalid target user id %s: %w", user.ID, err)
	}

	var ret = CovidUser{
		ID:    user.ID,
		IntID: userID,
	}
	return ret, nil
}

func (u *CovidUser) ContractCovid(infectedUser CovidUser, serverID, channelID, messageID string) (bool, error) {
	if !u.ContractionTime.IsZero() || infectedUser.ContractionTime.IsZero() || infectedUser.ID == u.ID {
		return false, nil
	}

	var rand = rand.New(rand.NewSource(u.IntID + infectedUser.IntID))
	var roll = float32(rand.Intn(10000)) / 100
	roll -= float32(u.CovidEncounter) * (ContractionChance / 2)
	u.CovidEncounter += 1
	if roll < ContractionChance {
		u.ContractionTime = time.Now()
		u.ContractedBy = infectedUser.ID
		u.ContractedServer = serverID
		u.ContractedChannel = channelID
		u.ContractedMessage = messageID
		return true, nil
	}

	return false, nil
}
