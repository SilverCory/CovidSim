package mask

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/fogleman/gg"
	"github.com/nfnt/resize"
	"github.com/vitali-fedulov/images"
	"image"
	"image/gif"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"strings"
)

var maskImage image.Image

func init() {
	var buf = base64.NewDecoder(base64.StdEncoding, strings.NewReader(MASK_IMAGE))
	var err error
	maskImage, _, err = image.Decode(buf)
	if err != nil {
		panic(fmt.Errorf("mask Image init failed: %w", err))
	}
}

func WearingMask(source image.Image) bool {
	hashA, sizeA := images.Hash(source)
	hashB, sizeB := images.Hash(AddMask(source))

	return images.Similar(hashA, hashB, sizeA, sizeB)
}

func AddMask(source image.Image) image.Image {
	var ctx = gg.NewContextForImage(source)
	ctx.DrawImage(
		resize.Resize(
			uint(ctx.Width()),
			uint(ctx.Height()),
			maskImage,
			resize.Bicubic,
		),
		0,
		0,
	)

	return ctx.Image()
}

func AddMaskGIF(source *gif.GIF) *gif.GIF {
	var buf = new(bytes.Buffer)
	var opts = new(gif.Options)
	for i, img := range source.Image {
		err := gif.Encode(buf, AddMask(img), opts)
		if err != nil {
			// TODO Blah
		}

		out, err := gif.Decode(buf)
		if err != nil {
			// TODO blah
		}

		source.Image[i] = out.(*image.Paletted)
	}

	return source
}
