package main

import "time"

func main() {
	const layout = "2006-01-02 15:04:05 -0700 MST"

	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		panic(err)
	}

	// US eastern war time
	t := time.Date(1944, time.August, 15, 0, 0, 0, 0, loc)
	println(t.Format(layout))

	// Us eastern peace time
	t = t.Add(24 * time.Hour * 30 * 13)
	println(t.Format(layout))

	loc, err = time.LoadLocation("America/Chicago")
	if err != nil {
		panic(err)
	}

	// US central time
	t = time.Date(1935, time.April, 5, 11, 11, 11, 0, loc)
	println(t.Format(layout))

	// US eastern time for a bit
	t = t.Add(24 * time.Hour * 365)
	println(t.Format(layout))

	// They didn't like it -- stayed light too late -- back to central
	t = t.Add(24 * time.Hour * 365)
	println(t.Format(layout))

	loc, err = time.LoadLocation("Asia/Kathmandu")
	if err != nil {
		panic(err)
	}

	// Nepalese time -- :30 off the hour
	t = time.Date(1985, time.September, 17, 12, 12, 12, 0, loc)
	println(t.Format(layout))

	// :30 off the hour is too hard so let's change it to on the hour.
	// Wait, no, let's switch to :45 off the hour, for convenience :)
	t = t.Add(24 * time.Hour * 365)
	println(t.Format(layout))

	loc, err = time.LoadLocation("Pacific/Kwajalein")
	if err != nil {
		panic(err)
	}

	// Marshall Islands -- where the world's day ends
	t = time.Date(1993, time.July, 4, 8, 0, 0, 0, loc)
	println(t.Format(layout))

	// They didn't like that. They want to be where the world's day begins.
	t = t.Add(24 * time.Hour * 60)
	println(t.Format(layout))

	loc, err = time.LoadLocation("Pacific/Guam")
	if err != nil {
		panic(err)
	}

	// Guam
	t = time.Date(1999, time.December, 25, 12, 0, 0, 0, loc)
	println(t.Format(layout))

	// Sometimes you want to change your timezone abbreviation for the sake of national identity.
	// A merry Christmas indeed!
	t = t.Add(24 * time.Hour * 365)
	println(t.Format(layout))

	loc, err = time.LoadLocation("Europe/Paris")
	if err != nil {
		panic(err)
	}

	// Paris -- days of yore -- local mean time, determined by longitude
	t = time.Date(1891, time.February, 14, 9, 0, 0, 0, loc)
	println(t.Format(layout))

	// Paris mean time
	t = t.Add(24 * time.Hour * 365)
	println(t.Format(layout))

	// Paris in the present -- CET -- a month earlier than the original date
	// due to 130 years worth of leap days.
	t = t.Add(24 * time.Hour * 365 * 130)
	println(t.Format(layout))

	// Paris in the summer -- CEST
	t = t.Add(24 * time.Hour * 30 * 5)
	println(t.Format(layout))
}

// Output:
// 1944-08-15 00:00:00 -0400 EWT
// 1945-09-09 00:00:00 -0400 EPT
// 1935-04-05 11:11:11 -0600 CST
// 1936-04-04 12:11:11 -0500 EST
// 1937-04-04 11:11:11 -0600 CST
// 1985-09-17 12:12:12 +0530 +0530
// 1986-09-17 12:27:12 +0545 +0545
// 1993-07-04 08:00:00 -1200 -12
// 1993-09-03 08:00:00 +1200 +12
// 1999-12-25 12:00:00 +1000 GST
// 2000-12-24 12:00:00 +1000 ChST
// 1891-02-14 09:00:00 +0009 LMT
// 1892-02-14 09:00:00 +0009 PMT
// 2022-01-13 09:50:39 +0100 CET
// 2022-06-12 10:50:39 +0200 CEST
