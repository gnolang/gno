func Render(path string) string {
	var b bytes.Buffer

	b.WriteString("# Polls!\n\n")

	if polls.Size() == 0 {
		b.WriteString("### No active polls currently!")
		return b.String()
	}
	polls.Iterate("", "", func(key string, value interface{}) bool {

		// cast raw data from tree into Poll struct
		p := value.(*poll.Poll)
		ddl := p.Deadline()

		yay, nay := p.VoteCount()
		yayPercent := 0
		nayPercent := 0

		if yay+nay != 0 {
			yayPercent = yay * 100 / (yay + nay)
			nayPercent = nay * 100 / (yay + nay)
		}

		b.WriteString(
			ufmt.Sprintf(
				"## Poll #%s: %s\n",
				key, // poll ID
				p.Title(),
			),
		)

		dropdown := "<details>\n<summary>Poll details</summary><br>"

		b.WriteString(dropdown + "Description: " + p.Description())

		b.WriteString(
			ufmt.Sprintf("<br>Voting until block: %d<br>Current vote count: %d",
				p.Deadline(),
				p.Voters().Size()),
		)

		b.WriteString(
			ufmt.Sprintf("<br>YAY votes: %d (%d%%)", yay, yayPercent),
		)
		b.WriteString(
			ufmt.Sprintf("<br>NAY votes: %d (%d%%)</details>", nay, nayPercent),
		)

		dropdown = "<br><details>\n<summary>Vote details</summary>"
		b.WriteString(dropdown)

		p.Voters().Iterate("", "", func(key string, value interface{}) bool {

			voter := key
			vote := value.(bool)

			if vote == true {
				b.WriteString(
					ufmt.Sprintf("<br>%s voted YAY!", voter),
				)
			} else {
				b.WriteString(
					ufmt.Sprintf("<br>%s voted NAY!", voter),
				)
			}
			return false
		})

		b.WriteString("</details>\n\n")
		return false
	})
	return b.String()
}
