package CovidSim

import (
	"image"
	"io"
)

type Cache interface {
	io.Closer
	InvalidateUser(userID string) error
	AvatarCache(userID string, getAvatarFunc func(userID string) (image.Image, error)) (image.Image, error)
	HasMaskCache(userID string, hasMaskFunc func(userID string) (bool, error)) (bool, error)
	SetMaskCooldown(userID string) error
	GetMaskCooldown(userID string) (bool, error)
}
