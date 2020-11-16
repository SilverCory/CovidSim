package CovidSim

import (
	"image"
	"io"
)

type Cache interface {
	io.Closer

	// InvalidateUser will clear all caching for the specified user.
	InvalidateUser(userID string) error

	// AvatarCache is a cache of a user's avatar data.
	AvatarCache(userID string, getAvatarFunc func(userID string) (image.Image, error)) (image.Image, error)
	// HasMaskCache is a cache of whether the user has a mask.
	HasMaskCache(userID string, hasMaskFunc func(userID string) (bool, error)) (bool, error)

	// GenerationCooldown will enable a cooldown for userID.
	GenerationCooldown(userID string) error
	// GetGenerationCooldown will get whether there is a cooldown for this user.
	GetGenerationCooldown(userID string) (bool, error)
}
