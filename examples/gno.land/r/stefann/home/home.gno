package home

import (
	"gno.land/p/demo/ufmt"

	"gno.land/r/stefann/config"
)

type City struct {
	Name string
	URL  string
}

var (
	pfp              string
	cities           []City
	currentCityIndex int
	aboutMe          [2]string
	jarLink          string
)

func init() {
	pfp = "https://i.ibb.co/Bc5YNCx/DSC-0095a.jpg"
	cities = []City{
		{Name: "Venice", URL: "https://i.ibb.co/1mcZ7b1/venice.jpg"},
		{Name: "Tokyo", URL: "https://i.ibb.co/wNDJv3H/tokyo.jpg"},
		{Name: "São Paulo", URL: "https://i.ibb.co/yWMq2Sn/sao-paulo.jpg"},
		{Name: "Toronto", URL: "https://i.ibb.co/pb95HJB/toronto.jpg"},
		{Name: "Bangkok", URL: "https://i.ibb.co/pQy3w2g/bangkok.jpg"},
		{Name: "New York", URL: "https://i.ibb.co/6JWLm0h/new-york.jpg"},
		{Name: "Paris", URL: "https://i.ibb.co/q9vf6Hs/paris.jpg"},
		{Name: "Kandersteg", URL: "https://i.ibb.co/60DzywD/kandersteg.jpg"},
		{Name: "Rothenburg", URL: "https://i.ibb.co/cr8d2rQ/rothenburg.jpg"},
		{Name: "Capetown", URL: "https://i.ibb.co/bPGn0v3/capetown.jpg"},
		{Name: "Sydney", URL: "https://i.ibb.co/TBNzqfy/sydney.jpg"},
		{Name: "Oeschinen Lake", URL: "https://i.ibb.co/QJQwp2y/oeschinen-lake.jpg"},
		{Name: "Barra Grande", URL: "https://i.ibb.co/z4RXKc1/barra-grande.jpg"},
		{Name: "London", URL: "https://i.ibb.co/CPGtvgr/london.jpg"},
	}
	currentCityIndex = 0
	jarLink = "https://TODO" // TODO
	aboutMe = [2]string{
		`<h3 style="font-size: 1.4em;">About Me</h3>
<p style="font-size: 1.1em;">Hey there! I’m Stefan, a student of computer science. I’m all about exploring and adventure — whether it’s diving into the latest tech or discovering a new city, I’m always up for the challenge!</p>`,
		`<h3 style="font-size: 1.4em;">Contributions</h3>
<p style="font-size: 1.1em;">I'm just getting started, but you can follow my journey through Gno.land right <a href="https://github.com/gnolang/TODO" target="_blank">here 🔗</a>.</p>`, // TODO
	}
}

func UpdateCities(newCities []City) {
	config.AssertAuthorized()
	cities = newCities
}

func UpdateJarLink(newLink string) {
	config.AssertAuthorized()
	jarLink = newLink
}

func UpdatePFP(url, caption string) {
	config.AssertAuthorized()
	pfp = url
}

func UpdateAboutMe(col1, col2 string) {
	config.AssertAuthorized()
	aboutMe[0] = col1
	aboutMe[1] = col2
}

func Render(path string) string {
	out := ufmt.Sprintf("# Exploring %s!\n\n", cities[currentCityIndex].Name)

	out += renderAboutMe()
	out += "\n\n"
	out += renderTips()

	return out
}

func renderAboutMe() string {
	out := "<div class='rows-3'>"

	out += "<div style='position: relative; text-align: center;'>\n\n"

	out += ufmt.Sprintf("<div style='background-image: url(%s); background-size: cover; background-position: center; width: 100%%; height: 600px; position: relative; border-radius: 15px; overflow: hidden;'>\n\n", cities[currentCityIndex].URL)

	out += ufmt.Sprintf("<img src='%s' alt='my profile pic' style='width: 250px; height: auto; aspect-ratio: 1 / 1; object-fit: cover; border-radius: 50%%; border: 3px solid #1e1e1e; position: absolute; top: 75%%; left: 50%%; transform: translate(-50%%, -50%%);'>\n\n", pfp)

	out += "</div>\n\n"

	out += "<div>\n\n"
	out += aboutMe[0] + "\n\n"
	out += "</div>\n\n"

	out += "<div>\n\n"
	out += aboutMe[1] + "\n\n"
	out += "</div>\n\n"

	out += "</div><!-- /rows-3 -->\n\n"

	return out
}

func renderTips() string {
	out := `<div class="jumbotron" style="display: flex; flex-direction: column; justify-content: flex-start; align-items: center; padding-top: 40px; padding-bottom: 50px; text-align: center;">` + "\n\n"

	out += `<div class="rows-2" style="max-width: 500px; width: 100%; display: flex; flex-direction: column; justify-content: center; align-items: center;">` + "\n"

	out += `<h1 style="margin-bottom: 50px;">Help Me Travel The World</h1>` + "\n\n"

	out += renderTipsJar() + "\n"

	out += `<p style="margin-top: 20px;">`
	out += `<strong style="font-size: 1.2em;">Every donation takes me to a new destination!</strong><br>`
	out += `<span style="font-size: 1.2em; font-style: italic;">Click the jar, tip in GNOT tokens, and watch my background change as I head to a new adventure!</span></p>` + "\n\n"

	out += renderSponsors()

	out += `</div><!-- /rows-2 -->` + "\n\n"

	out += `</div><!-- /jumbotron -->` + "\n"

	return out
}

func renderSponsors() string {
	out := `<h3>Sponsor Leaderboard</h3>` + "\n"
	out += `<ul style="list-style-type: none; padding: 0; border: 1px solid #ddd; border-radius: 8px; width: 100%; max-width: 300px; margin: 0 auto;">` + "\n"
	out += `<li style="padding: 10px; border-bottom: 1px solid #ddd;"><strong>1. Address 1:</strong> <span style="float: right;">50 GNOT</span></li>` + "\n"
	out += `<li style="padding: 10px; border-bottom: 1px solid #ddd;"><strong>2. Address 2:</strong> <span style="float: right;">45 GNOT</span></li>` + "\n"
	out += `<li style="padding: 10px; border-bottom: 1px solid #ddd;"><strong>3. Address 3:</strong> <span style="float: right;">40 GNOT</span></li>` + "\n"
	out += `<li style="padding: 10px; border-bottom: 1px solid #ddd;"><strong>4. Address 4:</strong> <span style="float: right;">35 GNOT</span></li>` + "\n"
	out += `<li style="padding: 10px;"><strong>5. Address 5:</strong> <span style="float: right;">30 GNOT</span></li>` + "\n"

	out += `</ul>` + "\n"

	return out
}

func renderTipsJar() string {
	out := `<style>
        .tip-jar {
            width: 200px;
            height: 300px;
            background-color: #d1e7dd;
            border-radius: 50% 50% 30% 30%;
            border: 5px solid #a2a2a2;
            position: relative;
            box-shadow: 0 4px 8px rgba(0, 0, 0, 0.2);
            text-align: center;
            overflow: visible;
        }
        .lid {
            width: 180px;
            height: 60px;
            background-color: #d9534f;
            border-radius: 50% 50% 0 0;
            position: absolute;
            top: -10px;
            left: 7px;
            border: 3px solid #b52e2e;
            box-shadow: inset 0 -8px 0 #b52e2e, 0 4px 8px rgba(0, 0, 0, 0.2);
            z-index: 3;
        }
        .label {
            width: 150px;
            height: 70px;
            background-color: #fff;
            border: 3px solid #6c757d;
            position: absolute;
            bottom: 25%;
            left: 50%;
            transform: translateX(-50%);
            border-radius: 10px;
            padding: 10px;
            display: flex;
            justify-content: center;
            align-items: center;
            z-index: 2;
        }
        .label h3 {
            margin: 0;
            font-size: 20px;
            color: #495057;
            font-family: "Comic Sans MS", cursive, sans-serif;
            border-bottom: 2px solid #6c757d;
            display: inline-block;
            padding-bottom: 5px;
        }
        .coins {
            position: absolute;
            bottom: 5%;
            left: 50%;
            transform: translateX(-50%);
            width: 160px;
            height: 100px;
            z-index: 1;
        }
        .coin {
            width: 25px;
            height: 25px;
            background-color: gold;
            border-radius: 50%;
            border: 2px solid #d4af37;
            position: absolute;
            box-shadow: 0 2px 5px rgba(0, 0, 0, 0.2);
        }
        .coin::after {
            content: "GNOT";
            font-size: 7px;
            color: rgba(0, 0, 0, 0.4);
            position: absolute;
            top: 50%;
            left: 49%;
            transform: translate(-50%, -50%);
            pointer-events: none;
        }
        .coin:nth-child(1) {
            left: 10px;
            bottom: 5px;
        }
        .coin:nth-child(2) {
            left: 50px;
            bottom: 0px;
        }
        .coin:nth-child(3) {
            left: 90px;
            bottom: 10px;
        }
        .coin:nth-child(4) {
            left: 20px;
            bottom: 25px;
        }
        .coin:nth-child(5) {
            left: 70px;
            bottom: 20px;
        }
        .coin:nth-child(6) {
            left: 110px;
            bottom: 30px;
        }
        .coin:nth-child(7) {
            left: 40px;
            bottom: 40px;
        }
        .coin:nth-child(8) {
            left: 80px;
            bottom: 35px;
        }
        .coin:nth-child(9) {
            left: 100px;
            bottom: 45px;
        }
    </style>`

	out += ufmt.Sprintf(`<a href="%s" target="_blank" style="display: block; text-decoration: none;">`, jarLink) + "\n"

	out += `<div class="tip-jar">` + "\n"
	out += `<div class="lid"></div>` + "\n"
	out += `<div class="label">` + "\n"
	out += `<h3>Travel Tips</h3>` + "\n"
	out += `</div>` + "\n"

	out += `<div class="coins">` + "\n"
	out += `<div class="coin"></div>` + "\n"
	out += `<div class="coin"></div>` + "\n"
	out += `<div class="coin"></div>` + "\n"
	out += `<div class="coin"></div>` + "\n"
	out += `<div class="coin"></div>` + "\n"
	out += `<div class="coin"></div>` + "\n"
	out += `<div class="coin"></div>` + "\n"
	out += `<div class="coin"></div>` + "\n"
	out += `<div class="coin"></div>` + "\n"
	out += `</div>` + "\n"
	out += `</div>` + "\n"

	out += `</a>` + "\n"

	return out
}
