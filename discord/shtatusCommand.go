package discord

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"strings"
	"time"
)

// doShtatus is a command handler for the !shtatus command.
func (b *Bot) doShtatus(s *discordgo.Session, m *discordgo.MessageCreate) {
	var statUser *discordgo.User

	parts := strings.SplitN(m.Content, " ", 2)
	if len(parts) != 2 {
		if len(parts) == 1 {
			statUser = m.Author
		} else {
			_, _ = s.ChannelMessage(m.ChannelID, "!shtatus @someone")
			return
		}
	} else {
		id := parts[1]
		if strings.HasPrefix(id, "<@") {
			id = strings.TrimPrefix("<@", strings.TrimSuffix(id, ">"))
		}

		var err error
		statUser, err = s.User(id)
		if err != nil {
			s.ChannelMessage(m.ChannelID, err.Error())
			return
		}
	}

	u, _, err := b.storage.LoadUser(statUser.ID)
	if err != nil {
		s.ChannelMessage(m.ChannelID, err.Error())
		return
	}

	wearingMask, err := b.wearingMask(statUser.ID)
	if err != nil {
		s.ChannelMessage(m.ChannelID, err.Error())
		return
	}

	msgContents := `
UserID: %s
WearingMask: %t
HasCorona: %t
Contracted By (id): %s
Contracted On: %s`
	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf(
		msgContents,
		statUser.ID,
		wearingMask,
		!u.ContractionTime.IsZero(),
		u.ContractedBy,
		u.ContractionTime.Format(time.RFC3339),
	))
}
