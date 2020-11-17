package discord

import (
	"bytes"
	"fmt"
	"github.com/SilverCory/CovidSim/mask"
	"github.com/bwmarrin/discordgo"
	"github.com/kettek/apng"
	"image"
	"image/gif"
	"strings"
)

// getAvatar will get an avatar for the user.
func (b *Bot) getAvatar(u *discordgo.User) (string, image.Image, *apng.APNG, *gif.GIF, error) {
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

// wearingMask will analyse the users current PFP to check whether they're wearing a mask.
func (b *Bot) wearingMask(userID string) (bool, error) {
	// We can use UserAvatar here because we don't need to analyse the entire animated PFP.
	img, err := b.session.UserAvatar(userID)
	if err != nil {
		return false, fmt.Errorf("wearingMask: get user avatar: %w", err)
	}

	return mask.WearingMask(img), nil
}
