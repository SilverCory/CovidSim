package discord

import (
	"bytes"
	"fmt"
	"github.com/SilverCory/CovidSim"
	"github.com/SilverCory/CovidSim/mask"
	"github.com/bwmarrin/discordgo"
	"github.com/kettek/apng"
	"github.com/rs/zerolog"
	"image/gif"
	"image/png"
	"strings"
	"time"
)

type Bot struct {
	l                 zerolog.Logger
	session           *discordgo.Session
	storage           CovidSim.Storage
	cache             CovidSim.Cache
	hookID            string
	hookToken         string
	lastUserByChannel map[string]CovidSim.CovidUser
}

func NewBot(l zerolog.Logger, token string, storage CovidSim.Storage, cache CovidSim.Cache, hookID, hookToken string) (Bot, error) {
	var ret = Bot{
		l:                 l,
		lastUserByChannel: make(map[string]CovidSim.CovidUser),
		storage:           storage,
		cache:             cache,
		hookID:            hookID,
		hookToken:         hookToken,
	}

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		return Bot{}, fmt.Errorf("NewBot: error creating Discord session: %w", err)
	}
	ret.session = dg

	dg.AddHandler(ret.infectionHandler)
	dg.AddHandler(ret.messageCommandHandler)
	dg.AddHandler(ret.userUpdate)
	dg.AddHandler(ret.ready)
	dg.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuildMessages | discordgo.IntentsDirectMessages)

	err = dg.Open()
	if err != nil {
		return Bot{}, fmt.Errorf("NewBot: error opening connection: %w", err)
	}

	ret.infectThy() // F

	return ret, nil
}

func (b *Bot) Close() error {
	if err := b.session.Close(); err != nil {
		return fmt.Errorf("Close: discord bot Close error: %w", err)
	}
	return nil
}

func (b *Bot) userUpdate(s *discordgo.Session, u *discordgo.UserUpdate) {
	if err := b.cache.InvalidateUser(u.ID); err != nil {
		b.l.Err(err).Str("user_id", u.ID).Msg("userUpdate failed")
	}
}

func (b *Bot) ready(s *discordgo.Session, _ *discordgo.Ready) {
	_ = s.UpdateStatus(0, "!wear-a-mask | !invite")
}

func (b *Bot) messageCommandHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID || m.Author.Bot {
		return
	}
	var l = b.l.With().
		Str("user_id", m.ID).
		Str("channel_id", m.ChannelID).
		Str("guild_id", m.GuildID).
		Logger()

	var id = m.Author.ID
	var msg = strings.ToLower(strings.TrimSpace(m.Message.Content))
	if strings.HasPrefix(msg, "!shtatus") {
		b.doShtatus(s, m)
		return
	}

	if strings.HasPrefix(msg, "!invite") {
		b.doInvite(s, m)
		return
	}

	if !strings.HasPrefix(msg, "!wear-a-mask") && !strings.HasPrefix(msg, "!wearamask") {
		return
	}

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

func (b *Bot) infectionHook(m *discordgo.MessageCreate, infectedUser CovidSim.CovidUser) {
	_, _ = b.session.WebhookExecute(b.hookID, b.hookToken, false, &discordgo.WebhookParams{
		Content: fmt.Sprintf(
			"**NEW INFECTION!**\n"+
				"\tID:\t\t%s\n"+
				"\tUser:\t\t%s\n"+
				"\tContractedBy:\t\t%s",
			m.Author.ID,
			m.Author.Username,
			infectedUser.ContractedBy,
		),
	})
}

func isCommand(c string) bool {
	c = strings.TrimSpace(c)
	return strings.HasPrefix(c, "!") ||
		strings.HasPrefix(c, "/") ||
		strings.HasPrefix(c, ".") ||
		strings.HasPrefix(c, ",") ||
		strings.HasPrefix(c, "$") ||
		strings.HasPrefix(c, "%") ||
		strings.HasPrefix(c, "|")
}
