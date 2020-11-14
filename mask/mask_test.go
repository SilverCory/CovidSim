package mask

import (
	"github.com/kettek/apng"
	"image"
	"image/gif"
	_ "image/jpeg"
	"image/png"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestAddMask(t *testing.T) {
	resp, err := http.Get("https://cdn.discordapp.com/avatars/303391020622544909/afd206b822db9bc9081f558d8e9a7637.png?size=1024")
	if err != nil {
		t.Fatalf("get avatar: %v", err)
		return
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		t.Fatalf("Status code was not 2xx, instead got %d", resp.StatusCode)
		return
	}

	img, _, err := image.Decode(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		t.Fatalf("Unable to decode image: %v", err)
		return
	}

	filePath := os.TempDir() + time.Now().Format(time.RFC3339) + ".png"
	t.Logf("Saving file to: %s\n", filePath)
	img = AddMask(img)
	f, err := os.Create(filePath)
	if err != nil {
		t.Fatalf("Creating file failed: %v", err)
		return
	}

	if err := png.Encode(f, img); err != nil {
		t.Fatalf("Encoding image failed: %v", err)
		return
	}

	_ = f.Close()
}

func TestAddMaskGIF(t *testing.T) {
	resp, err := http.Get("https://cdn.discordapp.com/avatars/318765117674225665/a_d583d54029fb0c7ed130e9a15fb34d19.gif?size=1024")
	if err != nil {
		t.Fatalf("get avatar: %v", err)
		return
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		t.Fatalf("Status code was not 2xx, instead got %d", resp.StatusCode)
		return
	}

	img, err := gif.DecodeAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		t.Fatalf("Unable to decode image: %v", err)
		return
	}

	filePath := os.TempDir() + time.Now().Format(time.RFC3339) + ".gif"
	t.Logf("Saving file to: %s\n", filePath)
	img = AddMaskGIF(img)
	f, err := os.Create(filePath)
	if err != nil {
		t.Fatalf("Creating file failed: %v", err)
		return
	}

	if err := gif.EncodeAll(f, img); err != nil {
		t.Fatalf("Encoding image failed: %v", err)
		return
	}

	_ = f.Close()
}
func TestAddMaskAPNG(t *testing.T) {
	resp, err := http.Get("https://cdn.discordapp.com/avatars/159767754960928768/a_87c818a999682a3e34471f4b7d178a14.gif?size=1024")
	if err != nil {
		t.Fatalf("get avatar: %v", err)
		return
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		t.Fatalf("Status code was not 2xx, instead got %d", resp.StatusCode)
		return
	}

	img, err := apng.DecodeAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		t.Fatalf("Unable to decode image: %v", err)
		return
	}

	filePath := os.TempDir() + time.Now().Format(time.RFC3339) + ".gif"
	t.Logf("Saving file to: %s\n", filePath)
	img = AddMaskAPNG(img)
	f, err := os.Create(filePath)
	if err != nil {
		t.Fatalf("Creating file failed: %v", err)
		return
	}

	if err := apng.Encode(f, img); err != nil {
		t.Fatalf("Encoding image failed: %v", err)
		return
	}

	_ = f.Close()
}

func TestWearingMask(t *testing.T) {
	t.Run("PNG No Mask", func(t *testing.T) {
		resp, err := http.Get("https://cdn.discordapp.com/avatars/303391020622544909/afd206b822db9bc9081f558d8e9a7637.png?size=1024")
		if err != nil {
			t.Fatalf("get avatar: %v", err)
			return
		}

		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			t.Fatalf("Status code was not 2xx, instead got %d", resp.StatusCode)
			return
		}

		img, _, err := image.Decode(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			t.Fatalf("Unable to decode image: %v", err)
			return
		}

		hasMask := WearingMask(img)
		if hasMask {
			t.Fail()
			t.Errorf("hasMask was true, expected false.")
		}

		t.Run("With mask", func(t *testing.T) {
			hasMask := WearingMask(AddMask(img))
			if !hasMask {
				t.Fail()
				t.Errorf("hasMask was false, expected true.")
			}
		})

	})

	t.Run("GIF No Mask", func(t *testing.T) {
		resp, err := http.Get("https://cdn.discordapp.com/avatars/318765117674225665/a_d583d54029fb0c7ed130e9a15fb34d19.gif?size=1024")
		if err != nil {
			t.Fatalf("get avatar: %v", err)
			return
		}

		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			t.Fatalf("Status code was not 2xx, instead got %d", resp.StatusCode)
			return
		}

		img, err := gif.DecodeAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			t.Fatalf("Unable to decode image: %v", err)
			return
		}

		hasMask := WearingMask(img.Image[0])
		if hasMask {
			t.Fail()
			t.Errorf("hasMask was true, expected false.")
		}

		t.Run("With mask", func(t *testing.T) {
			hasMask := WearingMask(AddMaskGIF(img).Image[0])
			if !hasMask {
				t.Fail()
				t.Errorf("hasMask was false, expected true.")
			}
		})
	})
}
