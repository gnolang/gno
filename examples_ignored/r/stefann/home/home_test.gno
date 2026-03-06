package home

import (
	"std"
	"strings"
	"testing"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/testutils"
)

func TestUpdateAboutMe(t *testing.T) {
	owner := address("g1sd5ezmxt4rwpy52u6wl3l3y085n8x0p6nllxm8")
	testing.SetOriginCaller(owner)

	profile.aboutMe = []string{}

	UpdateAboutMe("This is my new bio.|I love coding!")

	expected := []string{"This is my new bio.", "I love coding!"}

	if len(profile.aboutMe) != len(expected) {
		t.Fatalf("expected aboutMe to have length %d, got %d", len(expected), len(profile.aboutMe))
	}

	for i := range profile.aboutMe {
		if profile.aboutMe[i] != expected[i] {
			t.Fatalf("expected aboutMe[%d] to be %s, got %s", i, expected[i], profile.aboutMe[i])
		}
	}
}

func TestUpdateCities(t *testing.T) {
	owner := address("g1sd5ezmxt4rwpy52u6wl3l3y085n8x0p6nllxm8")
	testing.SetOriginCaller(owner)

	travel.cities = []City{}

	newCities := []City{
		{Name: "Berlin", URL: "https://example.com/berlin.jpg"},
		{Name: "Vienna", URL: "https://example.com/vienna.jpg"},
	}

	UpdateCities(newCities)

	if len(travel.cities) != 2 {
		t.Fatalf("expected 2 cities, got %d", len(travel.cities))
	}

	if travel.cities[0].Name != "Berlin" || travel.cities[1].Name != "Vienna" {
		t.Fatalf("expected cities to be updated to Berlin and Vienna, got %+v", travel.cities)
	}
}

func TestUpdateJarLink(t *testing.T) {
	owner := address("g1sd5ezmxt4rwpy52u6wl3l3y085n8x0p6nllxm8")
	testing.SetOriginCaller(owner)

	travel.jarLink = ""

	UpdateJarLink("https://example.com/jar")

	if travel.jarLink != "https://example.com/jar" {
		t.Fatalf("expected jarLink to be https://example.com/jar, got %s", travel.jarLink)
	}
}

func TestUpdateMaxSponsors(t *testing.T) {
	owner := address("g1sd5ezmxt4rwpy52u6wl3l3y085n8x0p6nllxm8")
	testing.SetOriginCaller(owner)

	sponsorship.maxSponsors = 0

	UpdateMaxSponsors(10)

	if sponsorship.maxSponsors != 10 {
		t.Fatalf("expected maxSponsors to be 10, got %d", sponsorship.maxSponsors)
	}

	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic for setting maxSponsors to 0")
		}
	}()
	UpdateMaxSponsors(0)
}

func TestAddCities(t *testing.T) {
	owner := address("g1sd5ezmxt4rwpy52u6wl3l3y085n8x0p6nllxm8")
	testing.SetOriginCaller(owner)

	travel.cities = []City{}

	AddCities(City{Name: "Berlin", URL: "https://example.com/berlin.jpg"})

	if len(travel.cities) != 1 {
		t.Fatalf("expected 1 city, got %d", len(travel.cities))
	}
	if travel.cities[0].Name != "Berlin" || travel.cities[0].URL != "https://example.com/berlin.jpg" {
		t.Fatalf("expected city to be Berlin, got %+v", travel.cities[0])
	}

	AddCities(
		City{Name: "Paris", URL: "https://example.com/paris.jpg"},
		City{Name: "Tokyo", URL: "https://example.com/tokyo.jpg"},
	)

	if len(travel.cities) != 3 {
		t.Fatalf("expected 3 cities, got %d", len(travel.cities))
	}
	if travel.cities[1].Name != "Paris" || travel.cities[2].Name != "Tokyo" {
		t.Fatalf("expected cities to be Paris and Tokyo, got %+v", travel.cities[1:])
	}
}

func TestAddAboutMeRows(t *testing.T) {
	owner := address("g1sd5ezmxt4rwpy52u6wl3l3y085n8x0p6nllxm8")
	testing.SetOriginCaller(owner)

	profile.aboutMe = []string{}

	AddAboutMeRows("I love exploring new technologies!")

	if len(profile.aboutMe) != 1 {
		t.Fatalf("expected 1 aboutMe row, got %d", len(profile.aboutMe))
	}
	if profile.aboutMe[0] != "I love exploring new technologies!" {
		t.Fatalf("expected first aboutMe row to be 'I love exploring new technologies!', got %s", profile.aboutMe[0])
	}

	AddAboutMeRows("Travel is my passion!", "Always learning.")

	if len(profile.aboutMe) != 3 {
		t.Fatalf("expected 3 aboutMe rows, got %d", len(profile.aboutMe))
	}
	if profile.aboutMe[1] != "Travel is my passion!" || profile.aboutMe[2] != "Always learning." {
		t.Fatalf("expected aboutMe rows to be 'Travel is my passion!' and 'Always learning.', got %+v", profile.aboutMe[1:])
	}
}

func TestDonate(t *testing.T) {
	user := testutils.TestAddress("user")
	testing.SetOriginCaller(user)

	sponsorship.sponsors = avl.NewTree()
	sponsorship.DonationsCount = 0
	sponsorship.sponsorsCount = 0
	travel.currentCityIndex = 0

	coinsSent := std.NewCoins(std.NewCoin("ugnot", 500))
	testing.SetOriginSend(coinsSent)
	Donate()

	existingAmount, exists := sponsorship.sponsors.Get(string(user))
	if !exists {
		t.Fatalf("expected sponsor to be added, but it was not found")
	}

	if existingAmount.(std.Coins).AmountOf("ugnot") != 500 {
		t.Fatalf("expected donation amount to be 500ugnot, got %d", existingAmount.(std.Coins).AmountOf("ugnot"))
	}

	if sponsorship.DonationsCount != 1 {
		t.Fatalf("expected DonationsCount to be 1, got %d", sponsorship.DonationsCount)
	}

	if sponsorship.sponsorsCount != 1 {
		t.Fatalf("expected sponsorsCount to be 1, got %d", sponsorship.sponsorsCount)
	}

	if travel.currentCityIndex != 1 {
		t.Fatalf("expected currentCityIndex to be 1, got %d", travel.currentCityIndex)
	}

	coinsSent = std.NewCoins(std.NewCoin("ugnot", 300))
	testing.SetOriginSend(coinsSent)
	Donate()

	existingAmount, exists = sponsorship.sponsors.Get(string(user))
	if !exists {
		t.Fatalf("expected sponsor to exist after second donation, but it was not found")
	}

	if existingAmount.(std.Coins).AmountOf("ugnot") != 800 {
		t.Fatalf("expected total donation amount to be 800ugnot, got %d", existingAmount.(std.Coins).AmountOf("ugnot"))
	}

	if sponsorship.DonationsCount != 2 {
		t.Fatalf("expected DonationsCount to be 2 after second donation, got %d", sponsorship.DonationsCount)
	}

	if travel.currentCityIndex != 2 {
		t.Fatalf("expected currentCityIndex to be 2 after second donation, got %d", travel.currentCityIndex)
	}
}

func TestGetTopSponsors(t *testing.T) {
	user := testutils.TestAddress("user")
	testing.SetOriginCaller(user)

	sponsorship.sponsors = avl.NewTree()
	sponsorship.sponsorsCount = 0

	sponsorship.sponsors.Set("g1address1", std.NewCoins(std.NewCoin("ugnot", 300)))
	sponsorship.sponsors.Set("g1address2", std.NewCoins(std.NewCoin("ugnot", 500)))
	sponsorship.sponsors.Set("g1address3", std.NewCoins(std.NewCoin("ugnot", 200)))
	sponsorship.sponsorsCount = 3

	topSponsors := GetTopSponsors()

	if len(topSponsors) != 3 {
		t.Fatalf("expected 3 sponsors, got %d", len(topSponsors))
	}

	if topSponsors[0].Address.String() != "g1address2" || topSponsors[0].Amount.AmountOf("ugnot") != 500 {
		t.Fatalf("expected top sponsor to be g1address2 with 500ugnot, got %s with %dugnot", topSponsors[0].Address.String(), topSponsors[0].Amount.AmountOf("ugnot"))
	}

	if topSponsors[1].Address.String() != "g1address1" || topSponsors[1].Amount.AmountOf("ugnot") != 300 {
		t.Fatalf("expected second sponsor to be g1address1 with 300ugnot, got %s with %dugnot", topSponsors[1].Address.String(), topSponsors[1].Amount.AmountOf("ugnot"))
	}

	if topSponsors[2].Address.String() != "g1address3" || topSponsors[2].Amount.AmountOf("ugnot") != 200 {
		t.Fatalf("expected third sponsor to be g1address3 with 200ugnot, got %s with %dugnot", topSponsors[2].Address.String(), topSponsors[2].Amount.AmountOf("ugnot"))
	}
}

func TestGetTotalDonations(t *testing.T) {
	user := testutils.TestAddress("user")
	testing.SetOriginCaller(user)

	sponsorship.sponsors = avl.NewTree()
	sponsorship.sponsorsCount = 0

	sponsorship.sponsors.Set("g1address1", std.NewCoins(std.NewCoin("ugnot", 300)))
	sponsorship.sponsors.Set("g1address2", std.NewCoins(std.NewCoin("ugnot", 500)))
	sponsorship.sponsors.Set("g1address3", std.NewCoins(std.NewCoin("ugnot", 200)))
	sponsorship.sponsorsCount = 3

	totalDonations := GetTotalDonations()

	if totalDonations != 1000 {
		t.Fatalf("expected total donations to be 1000ugnot, got %dugnot", totalDonations)
	}
}

func TestRender(t *testing.T) {
	travel.currentCityIndex = 0
	travel.cities = []City{
		{Name: "Venice", URL: "https://example.com/venice.jpg"},
		{Name: "Paris", URL: "https://example.com/paris.jpg"},
	}

	output := Render("")

	expectedCity := "Venice"
	if !strings.Contains(output, expectedCity) {
		t.Fatalf("expected output to contain city name '%s', got %s", expectedCity, output)
	}

	expectedURL := "https://example.com/venice.jpg"
	if !strings.Contains(output, expectedURL) {
		t.Fatalf("expected output to contain city URL '%s', got %s", expectedURL, output)
	}

	travel.currentCityIndex = 1
	output = Render("")

	expectedCity = "Paris"
	if !strings.Contains(output, expectedCity) {
		t.Fatalf("expected output to contain city name '%s', got %s", expectedCity, output)
	}

	expectedURL = "https://example.com/paris.jpg"
	if !strings.Contains(output, expectedURL) {
		t.Fatalf("expected output to contain city URL '%s', got %s", expectedURL, output)
	}
}
