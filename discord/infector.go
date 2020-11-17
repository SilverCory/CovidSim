package discord

import (
	"fmt"
	"github.com/SilverCory/CovidSim"
	"github.com/bwmarrin/discordgo"
)

// infectionHandler is the message handler that will actually infect users.
func (b *Bot) infectionHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID || m.Author.Bot {
		return
	}

	var content = m.Message.Content
	if isCommand(content) {
		return // I guess we want to ignore commands?
	}

	// check whether the user is wearing a mask, if so they're 'at low risk' (immune).
	wearingMask, err := b.cache.HasMaskCache(m.Author.ID, b.wearingMask)
	if err != nil {
		fmt.Printf("unable to check whether user %s is wearing mask: %v\n", m.Author.ID, err)
		return
	} else if wearingMask {
		return // Don't take it off..!
	}

	// Keep a log of the last user, if they have 'rona, we need to roll the dice.
	var lastUser, ok = b.lastUserByChannel[m.ChannelID]
	if !ok {
		// Load the last user from message history.
		var st, err = s.ChannelMessages(m.ChannelID, 1, m.ID, "", "")
		if err != nil {
			fmt.Printf("Unable to load channel messages for %s before msg %s: %v\n", m.ChannelID, m.ID, err)
			return
		}
		if len(st) == 1 {
			lastUser, ok, err = b.storage.LoadUser(st[0].Author.ID)
			if err != nil {
				fmt.Printf("Unable to load last user %s: %v\n", st[0].Author.ID, err)
				return
			}
			if !ok {
				return // user doesn't exist...
			}
		}
	}

	author, ok, err := b.storage.LoadUser(m.Author.ID)
	if err != nil {
		fmt.Printf("Unable to load author user %s: %v\n", m.Author.ID, err)
		return
	}

	if !ok {
		author, err = CovidSim.NewCovidUser(m.Author)
		if err != nil {
			fmt.Printf("Unable to create author covid user %s: %v\n", m.Author.ID, err)
			return
		}
	}

	var contractedCovid bool
	contractedCovid, err = author.ContractCovid(lastUser, m.GuildID, m.ChannelID, m.ID)
	if err != nil {
		fmt.Printf("Unable to contract covid for user %s from %s: %v\n", author.ID, lastUser.ID, err)
		return
	}

	if contractedCovid {
		go b.infectionHook(m, author)
	}

	if err := b.storage.SaveUser(author); err != nil {
		fmt.Printf("Unable to save covid user %s: %v", author.ID, err)
		return
	}

	b.lastUserByChannel[m.ChannelID] = author
}
