package discord

import (
	"fmt"
	"github.com/SilverCory/CovidSim"
	"time"
)

func (b *Bot) infectThy() {
	covidUser, ok, err := b.storage.LoadUser("105484726235607040")
	if err != nil {
		fmt.Println("Infecting thy load user: ", err)
		return
	}

	if ok && !covidUser.ContractionTime.IsZero() {
		return
	}

	user, err := b.session.User("105484726235607040")
	if err != nil {
		fmt.Printf("Thy not here? %v\n", err)
	} else {
		covUser, _ := CovidSim.NewCovidUser(user)
		covUser.ContractionTime = time.Now()
		covUser.ContractedBy = "1"
		covUser.ContractedChannel = "1"
		covUser.ContractedMessage = "1"
		covUser.CovidEncounter = 1

		if err := b.storage.SaveUser(covUser); err != nil {
			fmt.Printf("Can't save thy's infection :( %v", err)
		}
	}
}
