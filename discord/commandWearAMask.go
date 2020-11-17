package discord

import (
	"bytes"
	"fmt"
	"github.com/SilverCory/CovidSim/mask"
	"github.com/bwmarrin/discordgo"
	"github.com/kettek/apng"
	"image/gif"
	"image/png"
	"strings"
	"time"
)

func (b *Bot) doWearAMask(s *discordgo.Session, m *discordgo.MessageCreate) {
	var l = b.l.With().
		Str("user_id", m.ID).
		Str("channel_id", m.ChannelID).
		Str("guild_id", m.GuildID).
		Logger()
	var id = m.Author.ID

	ch, err := s.UserChannelCreate(id)
	if err != nil {
		l.Err(err).Msg("messageCommandHandler: channel create failed.")
		return
	}

	u, ok, _ := b.storage.LoadUser(m.Author.ID)
	// If we are going to offer vote testing then we shouldn't do this.
	if ok && !u.ContractionTime.IsZero() {
		_, _ = s.ChannelMessageSend(
			ch.ID,
			"LMAO it's too late for you. You're already infected :joy::frowning:\n"+
				"Here's a mask anyway, not that it's of any use.",
		)
	}

	chill, err := b.cache.GetGenerationCooldown(id)
	if err != nil {
		l.Err(err).Msg("messageCommandHandler: cache get generation cooldown failed.")
		return
	}
	if chill {
		_, _ = s.ChannelMessageSend(
			ch.ID,
			"Chill! You can only do this every 10 mins.\n"+
				"Don't take it off and you won't have to put it back on! :weary:",
		)
		return
	}

	_, _ = s.ChannelMessageSend(ch.ID, "Getting you a mask... :timer:")
	fail := func(err error) {
		_, _ = s.ChannelMessageSend(ch.ID, "Damn looks like I'm out.. :joy: (I broke, try later).\n\n"+err.Error())
	}

	var user = m.Author
	if m.Author.ID == "303391020622544909" { // Cory
		user, err = s.User(strings.SplitN(m.Content, " ", 2)[1])
		if err != nil {
			_, _ = s.ChannelMessageSend(ch.ID, err.Error())
			return
		}
		id = user.ID
	}

	format, img, ap, g, err := b.getAvatar(user)
	if err != nil {
		l.Err(err).Msg("messageCommandHandler: getAvatar failed.")
		fail(err)
		return
	}

	var buf = new(bytes.Buffer)
	if g != nil {
		format = "gif"
		if err := gif.EncodeAll(buf, mask.AddMaskGIF(g)); err != nil {
			l.Err(err).Msg("messageCommandHandler: gif encode failed.")
			fail(err)
			return
		}
	} else if ap != nil {
		format = "gif"
		if err := apng.Encode(buf, mask.AddMaskAPNG(*ap)); err != nil {
			l.Err(err).Msg("messageCommandHandler: apng encode failed.")
			fail(err)
			return
		}
	} else {
		format = "png"
		if err := png.Encode(buf, mask.AddMask(img)); err != nil {
			l.Err(err).Msg("messageCommandHandler: png encode failed.")
			fail(err)
			return
		}
	}

	_, err = s.ChannelMessageSendComplex(ch.ID, &discordgo.MessageSend{
		Content: "Here is your mask! **SET IT AS YOUR AVATAR TO WEAR IT!**",
		Files: []*discordgo.File{{
			Name:        time.Now().Format(time.RFC3339) + "_MaskedUp." + format,
			ContentType: "image/" + format,
			Reader:      buf,
		}},
	})
	if err != nil {
		fmt.Printf("Unable to send complex avatar message: %v\n", err)
		l.Err(err).Msg("messageCommandHandler: channel message send complex failed.")
		fail(err)
		return
	}

	if err = b.cache.GenerationCooldown(id); err != nil {
		// Don't tell the user lol.
		l.Err(err).Str("user_id", id).Msg("messageCommandHandler: store cooldown.")
	}
}
