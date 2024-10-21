package home

import (
        "std"
        "gno.land/p/demo/ufmt"
)

var (
        pfp          string // link to profile picture
        pfpCaption   string // profile picture caption
        abtMe        string
        currentColor string

        redVotes   int64
        blueVotes  int64
        greenVotes int64

        blueLink  string
        redLink   string
        greenLink string
)

func init() {
        pfp = "https://muaythairecords.com/fighters/rodtang-jitmuangnon/ogimage"
        pfpCaption = "My profile picture - Rodtang Jitmuangnon (Muay Thai fighter)"
        abtMe =
                `### About me
Motivated Computer Science student with strong
analytical and problem-solving skills. Proficient in
programming and version control, with a high level of
focus and attention to detail. Eager to apply academic
knowledge to real-world projects and contribute to
innovative technology solutions.

In addition to my academic pursuits, I enjoy traveling and staying active through weightlifting. I have a keen interest in electronic music and often explore various genres. I believe in maintaining a balanced lifestyle that complements my professional development.`
        redVotes = 0
        greenVotes = 0
        blueVotes = 0
        currentColor = `rgb(120, 120, 120)`
        blueLink = "https://www.google.com"
        redLink = "https://www.google.com"
        greenLink = "https://www.google.com"
}

func UpdatePFP(url, caption string) {
        AssertAuthorized()
        pfp = url
        pfpCaption = caption
}

func UpdateAboutMe(col1 string) {
        AssertAuthorized()
        abtMe = col1
}

func maxOfThree(a, b, c int64) int64 {
        max := a
        if b > max {
                max = b
        }
        if c > max {
                max = c
        }
        return max
}

func calculateRGB(red, green, blue int64) (int, int, int) {
        maxValue := maxOfThree(red, green, blue)

        if maxValue > 255 {
                scaleFactor := float64(255) / float64(maxValue)

                red = int64(float64(red) * scaleFactor)
                green = int64(float64(green) * scaleFactor)
                blue = int64(float64(blue) * scaleFactor)
        }

        return int(red), int(green), int(blue)
}

func VoteRed() {

        ugnotAmount := std.GetOrigSend().AmountOf("ugnot")

        redVotes = redVotes + ugnotAmount/750
        red, green, blue := calculateRGB(redVotes, greenVotes, blueVotes)
        currentColor = ufmt.Sprintf(`rgb(%d, %d, %d)`, red, green, blue)
}

func VoteGreen() {
        ugnotAmount := std.GetOrigSend().AmountOf("ugnot")

        greenVotes = greenVotes + ugnotAmount/750
        red, green, blue := calculateRGB(redVotes, greenVotes, blueVotes)
        currentColor = ufmt.Sprintf(`rgb(%d, %d, %d)`, red, green, blue)
}

func VoteBlue() {
        ugnotAmount := std.GetOrigSend().AmountOf("ugnot")

        blueVotes = blueVotes + ugnotAmount/750
        red, green, blue := calculateRGB(redVotes, greenVotes, blueVotes)
        currentColor = ufmt.Sprintf(`rgb(%d, %d, %d)`, red, green, blue)

}

func CollectBalance() {
        AssertAuthorized()

        banker := std.GetBanker(std.BankerTypeRealmSend)
        ownerAddr := Address()

        banker.SendCoins(std.CurrentRealm().Addr(), ownerAddr, banker.GetCoins(std.CurrentRealm().Addr()))
}

func Render(path string) string {

        out := ufmt.Sprintf("<div style='font-family: \"Trebuchet MS\", sans-serif;font-size: 5rem; color: %s;'>\n\n", currentColor)
        out += "# Matija's Homepage \n\n"
        out += "</div>\n\n"
        out += "<div style='font-family: \"Trebuchet MS\", sans-serif; text-align: justify;'>\n\n"
        out += renderAboutMe()
        out += "\n\n"
        out += ufmt.Sprintf("<h3 style='color: %s;'>Vote for the color of the page - the more you pay the more effect it will have on the homepage ;)</h3>\n", currentColor)
        out += renderButtons()
        out += "</div>"
        out += renderGitHubProjects()
        out += "\n\n"

        return out
}

func renderAboutMe() string {
        out := "<div class='columns-2'>"

        out += "<div>\n\n"
        out += ufmt.Sprintf(
                `<img src='%s' alt='Profile Picture' style='border-radius: 50%%; width: 200px; transition: transform 0.3s ease; outline: 5px solid %s;' />`,
                pfp, currentColor)
        out += ufmt.Sprintf("<p style='color: %s'>%s</p>\n\n", currentColor, pfpCaption)
        out += "</div>\n\n"

        out += ufmt.Sprintf("<div style='color: %s'>\n\n", currentColor)
        out += abtMe + "\n\n"
        out += "</div>\n\n"

        out += "</div>\n\n"

        return out
}

func renderGitHubProjects() string {
        out := `<div style="display: flex; justify-content: center; gap: 10px;">`

        // GitHub Button
        out += `<a href="https://github.com/matijamarjanovic" target="_blank" style="text-decoration: none;">`
        out += `<button style="background-color: black; color: white; padding: 8px 16px; border: none; border-radius: 5px; cursor: pointer; display: flex; align-items: center; font-size: 14px;">`
        out += `<img src="https://upload.wikimedia.org/wikipedia/commons/9/91/Octicons-mark-github.svg" alt="GitHub Logo" style="width: 16px; height: 16px; margin-right: 8px;">`
        out += `GitHub`
        out += `</button>`
        out += `</a>`

        // LinkedIn Button
        out += `<a href="https://www.linkedin.com/in/matijamarjanovic" target="_blank" style="text-decoration: none;">`
        out += `<button style="background-color: rgb(30, 70, 220); color: white; padding: 8px 16px; border: none; border-radius: 5px; cursor: pointer; display: flex; align-items: center; font-size: 14px;">`
        out += `<img src="https://upload.wikimedia.org/wikipedia/commons/0/01/LinkedIn_Logo.svg" alt="LinkedIn Logo" style="width: 16px; height: 16px; margin-right: 8px;">`
        out += `LinkedIn`
        out += `</button>`
        out += `</a>`

        out += "</div>\n\n\n\n\n"

        return out
}

func renderButtons() string {
        out := "<div style='display: flex; justify-content: space-around; padding-top: 20px;'>\n\n"

        // Red Button
        out += "<div>\n"
        out += ufmt.Sprintf(
                "<a href='%s' target='_blank' style='text-decoration: none;'>\n<button style='background-color: darkred; color: white; padding: 20px 40px; font-size: 24px; border: none; border-radius: 5px; cursor: pointer;'>Vote Red</button>\n</a>\n",
                redLink)
        out += ufmt.Sprintf("<p style='color: %s; text-align: center;'>Red Votes: %d</p>\n", "darkred", redVotes) // Show red votes
        out += "</div>\n\n"

        // Green Button
        out += "<div>\n"
        out += ufmt.Sprintf(
                "<a href='%s' target='_blank' style='text-decoration: none;'>\n<button style='background-color: teal; color: white; padding: 20px 40px; font-size: 24px; border: none; border-radius: 5px; cursor: pointer;'>Vote Green</button>\n</a>\n",
                greenLink)
        out += ufmt.Sprintf("<p style='color: %s; text-align: center;'>Green Votes: %d</p>\n", "teal", greenVotes) // Show green votes
        out += "</div>\n\n"

        // Blue Button
        out += "<div>\n"
        out += ufmt.Sprintf(
                "<a href='%s' target='_blank' style='text-decoration: none;'>\n<button style='background-color: navy; color: white; padding: 20px 40px; font-size: 24px; border: none; border-radius: 5px; cursor: pointer;'>Vote Blue</button>\n</a>\n",
                blueLink)
        out += ufmt.Sprintf("<p style='color: %s; text-align: center;'>Blue Votes: %d</p>\n", "navy", blueVotes) // Show blue votes
        out += "</div>\n\n"

        out += "</div>\n\n"

        // Current RGB Display
        out += "<div style='text-align: center; margin-top: 20px;'>\n"
        out += ufmt.Sprintf("<p style='color: %s;'>Current RGB: %s</p>\n", currentColor, currentColor)
        out += "</div>\n\n"

        return out
}

func UpdateBlueLink(link string) {
        AssertAuthorized()
        blueLink = link
}

func UpdateRedLink(link string) {
        AssertAuthorized()
        redLink = link
}

func UpdateGreenLink(link string) {
        AssertAuthorized()
        greenLink = link
}