package discord

import "github.com/bwmarrin/discordgo"

const inviteLinks = `https://discord.com/oauth2/authorize?client_id=776788476611264583&scope=bot&permissions=76800
https://discord.gg/a6BKEXu7JQ`

func (b *Bot) doInvite(s *discordgo.Session, m *discordgo.MessageCreate) {
	ch, err := s.UserChannelCreate(m.ID)
	if err != nil {
		b.l.Err(err).Str("user_id", m.ID).Msg("messageCommandHandler: !invite: channel create failed")
		return
	}

	_, _ = s.ChannelMessageSendComplex(ch.ID, &discordgo.MessageSend{
		Content: inviteLinks,
		Embed:   &discordgo.MessageEmbed{},
	})
	return
}
