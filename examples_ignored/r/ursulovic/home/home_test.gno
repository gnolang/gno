package home

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
)

func TestUpdateGithubUrl(t *testing.T) {
	caller := address("g1d24j8fwnc0w5q427fauyey4gdd30qgu69k6n0x")
	testing.SetOriginCaller(caller)

	newUrl := "https://github.com/example"

	UpdateGithubUrl(newUrl)

	if githubUrl != newUrl {
		t.Fatalf("GitHub url not updated properly!")
	}
}

func TestUpdateLinkedinUrl(t *testing.T) {
	caller := address("g1d24j8fwnc0w5q427fauyey4gdd30qgu69k6n0x")
	testing.SetOriginCaller(caller)

	newUrl := "https://www.linkedin.com/in/example"

	UpdateGithubUrl(newUrl)

	if githubUrl != newUrl {
		t.Fatalf("LinkedIn url not updated properly!")
	}
}

func TestUpdateAboutMe(t *testing.T) {
	caller := address("g1d24j8fwnc0w5q427fauyey4gdd30qgu69k6n0x")
	testing.SetOriginCaller(caller)

	newAboutMe := "This is new description!"

	UpdateAboutMe(newAboutMe)

	if aboutMe != newAboutMe {
		t.Fatalf("About mew not updated properly!")
	}
}

func TestUpdateSelectedImage(t *testing.T) {
	user := testutils.TestAddress("user")
	testing.SetOriginCaller(user)

	validImageUrl := "https://i.ibb.co/hLtmnX0/beautiful-rain-forest-ang-ka-nature-trail-doi-inthanon-national-park-thailand-36703721.webp"

	coinsSent := std.NewCoins(std.NewCoin("ugnot", 5000000)) // Update to match the price expected by your function
	testing.SetOriginSend(coinsSent)

	UpdateSelectedImage(validImageUrl)

	if selectedImage != validImageUrl {
		t.Fatalf("Valid image URL rejected!")
	}

	invalidImageUrl := "https://ibb.co/Kb3rQNn"

	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("Expected panic for invalid image URL, but got no panic")
		}
	}()

	UpdateSelectedImage(invalidImageUrl)

	invalidCoins := std.NewCoins(std.NewCoin("ugnot", 1000000))
	testing.SetOriginSend(invalidCoins)

	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("Expected panic for incorrect coin denomination or amount, but got no panic")
		}
	}()

	UpdateSelectedImage(validImageUrl)
}

func TestUpdateImagePrice(t *testing.T) {
	caller := address("g1d24j8fwnc0w5q427fauyey4gdd30qgu69k6n0x")
	testing.SetOriginCaller(caller)

	var newImageUpdatePrice int64 = 3000000

	UpdateImagePrice(newImageUpdatePrice)

	if imageUpdatePrice != newImageUpdatePrice {
		t.Fatalf("Image update price not updated properly!")
	}
}
