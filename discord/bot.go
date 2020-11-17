package discord

import (
	"fmt"
	"github.com/SilverCory/CovidSim"
	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog"
	"strings"
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

func (b *Bot) ready(s *discordgo.Session, _ *discordgo.Ready) {
	_ = s.UpdateStatus(0, "!wear-a-mask | !invite")
}

func (b *Bot) userUpdate(s *discordgo.Session, u *discordgo.UserUpdate) {
	if err := b.cache.InvalidateUser(u.ID); err != nil {
		b.l.Err(err).Str("user_id", u.ID).Msg("userUpdate failed")
	}
}

func (b *Bot) messageCommandHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID || m.Author.Bot {
		return
	}

	var msg = strings.ToLower(strings.TrimSpace(m.Message.Content))
	if strings.HasPrefix(msg, "!shtatus") {
		b.doShtatus(s, m)
		return
	}

	if strings.HasPrefix(msg, "!invite") {
		b.doInvite(s, m)
		return
	}

	if strings.HasPrefix(msg, "!wear-a-mask") || strings.HasPrefix(msg, "!wearamask") {
		b.doWearAMask(s, m)
		return
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

func (b *Bot) Close() error {
	if err := b.session.Close(); err != nil {
		return fmt.Errorf("close: discord bot Close error: %w", err)
	}
	return nil
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
