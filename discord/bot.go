package discord

import (
	"bytes"
	"fmt"
	"github.com/SilverCory/CovidSim"
	"github.com/SilverCory/CovidSim/mask"
	"github.com/bwmarrin/discordgo"
	"github.com/kettek/apng"
	"image"
	"image/gif"
	"image/png"
	"strings"
	"time"
)

type Bot struct {
	session           *discordgo.Session
	storage           CovidSim.Storage
	cache             CovidSim.Cache
	hookID            string
	hookToken         string
	lastUserByChannel map[string]CovidSim.CovidUser
}

func NewBot(token string, storage CovidSim.Storage, cache CovidSim.Cache, hookID, hookToken string) (Bot, error) {
	var ret = Bot{
		lastUserByChannel: make(map[string]CovidSim.CovidUser),
		storage:           storage,
		cache:             cache,
		hookID:            hookID,
		hookToken:         hookToken,
	}

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		return Bot{}, fmt.Errorf("error creating Discord session: %w", err)
	}
	ret.session = dg

	// Register the messageCreate func as a callback for MessageCreate events.
	dg.AddHandler(ret.messageCreate)
	dg.AddHandler(ret.messageCreateCmd)
	dg.AddHandler(ret.userUpdate)
	dg.AddHandler(ret.ready)
	dg.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuildMessages | discordgo.IntentsDirectMessages)

	err = dg.Open()
	if err != nil {
		return Bot{}, fmt.Errorf("error opening connection: %w", err)
	}

	ret.infectThy()

	return ret, nil
}

func (b *Bot) Close() error {
	if err := b.session.Close(); err != nil {
		return fmt.Errorf("discord bot Close error: %w", err)
	}
	return nil
}

func (b *Bot) userUpdate(s *discordgo.Session, u *discordgo.UserUpdate) {
	if err := b.cache.InvalidateUser(u.ID); err != nil {
		fmt.Printf("Unable to invalidate a user on update: %v\n", err)
	}
}

func (b *Bot) ready(s *discordgo.Session, _ *discordgo.Ready) {
	_ = s.UpdateStatus(0, "!wear-a-mask | !invite")
}

func (b *Bot) messageCreateCmd(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID || m.Author.Bot {
		return
	}

	var id = m.Author.ID
	var msg = strings.ToLower(strings.TrimSpace(m.Message.Content))
	if strings.HasPrefix(msg, "!shtatus") {
		b.doShtatus(s, m)
		return
	}

	if strings.HasPrefix(msg, "!invite") {
		ch, err := s.UserChannelCreate(id)
		if err != nil {
			fmt.Printf("Unable to create user channel for %s: %w", id)
			return
		}

		_, _ = s.ChannelMessageSendComplex(ch.ID, &discordgo.MessageSend{
			Content: "https://discord.com/oauth2/authorize?client_id=776788476611264583&scope=bot&permissions=76800\nhttps://discord.gg/a6BKEXu7JQ",
			Embed:   &discordgo.MessageEmbed{},
		})
		return
	}

	if !strings.HasPrefix(msg, "!wear-a-mask") && !strings.HasPrefix(msg, "!wearamask") {
		return
	}

	ch, err := s.UserChannelCreate(id)
	if err != nil {
		fmt.Printf("Unable to create user channel for %s: %w", id)
		return
	}

	u, ok, _ := b.storage.LoadUser(m.Author.ID)
	if ok && !u.ContractionTime.IsZero() {
		_, _ = s.ChannelMessageSend(
			ch.ID,
			"LMAO it's too late for you. You're already infected :joy::frowning:\n"+
				"Here's a mask anyway, not that it's of any use.",
		)
	}

	chill, err := b.cache.GetGenerationCooldown(id)
	if err != nil {
		fmt.Printf("getting cooldown failed for %s: %w", id)
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
	fail := func() {
		_, _ = s.ChannelMessageSend(ch.ID, "Damn looks like I'm out.. :joy: (I broke, try later).")
	}

	var user = m.Author
	if m.Author.ID == "303391020622544909" {
		user, err = s.User(strings.SplitN(m.Content, " ", 2)[1])
		if err != nil {
			_, _ = s.ChannelMessageSend(ch.ID, err.Error())
			return
		}
		id = user.ID
	}

	format, img, ap, g, err := b.getAvatar(user)
	if err != nil {
		fmt.Printf("Unable to get avatar for %s: %v\n", id, err)
		fail()
		return
	}

	var buf = new(bytes.Buffer)
	if g != nil {
		format = "gif"
		if err := gif.EncodeAll(buf, mask.AddMaskGIF(g)); err != nil {
			fmt.Printf("Unable to encode gif: %v\n", err)
			fail()
			return
		}
	} else if ap != nil {
		format = "gif"
		if err := apng.Encode(buf, mask.AddMaskAPNG(*ap)); err != nil {
			fmt.Printf("Unable to encode apng: %v\n", err)
			fail()
			return
		}
	} else {
		format = "png"
		if err := png.Encode(buf, mask.AddMask(img)); err != nil {
			fmt.Printf("Unable to encode png: %v\n", err)
			fail()
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
		fail()
		return
	}

	if err = b.cache.GenerationCooldown(id); err != nil {
		fmt.Printf("Unable to store cooldown: %v\n", err)
	}
}

func (b *Bot) messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID || m.Author.Bot {
		return
	}

	var content = m.Message.Content
	if isCommand(content) {
		return // I guess we want to ignore commands?
	}

	wearingMask, err := b.cache.HasMaskCache(m.Author.ID, b.wearingMask)
	if err != nil {
		fmt.Printf("unable to check whether user %s is wearing mask: %v\n", m.Author.ID, err)
		return
	}
	if wearingMask {
		return // Don't take it off..!
	}

	var lastUser, ok = b.lastUserByChannel[m.ChannelID]
	if !ok {
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

func (b *Bot) wearingMask(userID string) (bool, error) {
	img, err := b.session.UserAvatar(userID)
	if err != nil {
		return false, fmt.Errorf("wearingMask get user avatar: %w", err)
	}

	return mask.WearingMask(img), nil
}

func (b *Bot) getAvatar(u *discordgo.User) (string, image.Image, *apng.APNG, *gif.GIF, error) {
	// TODO clean this up.
	body, err := b.session.RequestWithBucketID("GET", u.AvatarURL("1024"), nil, discordgo.EndpointUserAvatar("", ""))
	if err != nil {
		return "", nil, nil, nil, err
	}
	var buf = bytes.NewBuffer(body)

	// Is gif
	if strings.HasPrefix(u.Avatar, "a_") {
		g, err := gif.DecodeAll(buf)
		return "gif", nil, nil, g, err
	}

	img, f, err := image.Decode(buf)
	return f, img, nil, nil, err
}

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
