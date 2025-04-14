package gnominers

import (
	"errors"
	"math"
	"strconv"
	"strings"

	"gno.land/p/demo/ufmt"
	"gno.land/p/moul/md"
	"gno.land/p/moul/mdtable"
)

func redoLink(params map[string]string, count int, from, to gemType) string {
	str := gnominersDAORequest + "?user=" + params["user"]

	// count
	if count >= 0 {
		str += "&count=" + strconv.Itoa(count)
	} else if c, ok := params["count"]; ok {
		str += "&count=" + c
	}

	// from
	if from != _gemEnd {
		str += "&from=" + strconv.Itoa(int(from))
	} else if c, ok := params["from"]; ok {
		str += "&from=" + c
	}

	// to
	if to != _gemEnd {
		str += "&to=" + strconv.Itoa(int(to))
	} else if c, ok := params["to"]; ok {
		str += "&to=" + c
	}
	return str
}

func displayCount(params map[string]string) string {
	// Usefull
	var content string
	countString, ok := params["count"]
	count := 10.
	if ok {
		x, err := strconv.Atoi(countString)
		if err == nil {
			count = float64(x)
		}
	}

	// Buttons
	content += md.H3("How much to convert : " + strconv.Itoa(int(count)))
	content += md.Link("+1", redoLink(params, int(count)+1, _gemEnd, _gemEnd)) + " /"
	content += md.Link("+10", redoLink(params, int(count)+10, _gemEnd, _gemEnd)) + " / "
	content += md.Link("+100", redoLink(params, int(count)+100, _gemEnd, _gemEnd)) + br

	content += md.Link("-1", redoLink(params, int(math.Max(1, count-1)), _gemEnd, _gemEnd)) + " / "
	content += md.Link("-10", redoLink(params, int(math.Max(1, count-10)), _gemEnd, _gemEnd)) + " / "
	content += md.Link("-100", redoLink(params, int(math.Max(1, count-100)), _gemEnd, _gemEnd)) + br
	return content
}

func paramToGem(params map[string]string, name string) (gemType, error) {
	x, ok := params[name]
	if !ok {
		return _gemEnd, errors.New("param doesnt exists")
	}
	xint, err := strconv.Atoi(x)
	if err != nil {
		return _gemEnd, err
	}
	return gemType(xint), nil
}

func generateReq(params map[string]string) []string {
	count, ok := params["count"]
	if !ok {
		count = "10"
	}
	from, err := paramToGem(params, "from")
	if err != nil {
		from = gemRock
	}
	to, err := paramToGem(params, "to")
	if err != nil {
		to = gemQuartz
	}
	return []string{count, strconv.Itoa(int(from)), strconv.Itoa(int(to))}
}

func displayFinal(req []string) string {
	return md.H2(ufmt.Sprintf("Convertion: %s %s \u2192 %s\n", req[0], gemNames[mustAtoi(req[1])], gemNames[mustAtoi(req[2])]))
}

func displayTable(params map[string]string, req []string) string {
	var content, a, b string

	t := mdtable.Table{Headers: []string{"From", "To"}}
	content += md.H3("Gem Selecion")
	for gem, gemName := range gemNames {
		switch gem {
		case mustAtoi(req[2]):
			a = md.Link(gemName, redoLink(params, -1, gemType(mustAtoi(req[2])), gemType(mustAtoi(req[1]))))
		case mustAtoi(req[1]):
			a = md.Bold(gemName)
		default:
			a = md.Link(gemName, redoLink(params, -1, gemType(gem), _gemEnd))
		}

		switch gem {
		case mustAtoi(req[1]):
			b = md.Link(gemName, redoLink(params, -1, gemType(mustAtoi(req[2])), gemType(mustAtoi(req[1]))))
		case mustAtoi(req[2]):
			b = md.Bold(gemName)
		default:
			b = md.Link(gemName, redoLink(params, -1, _gemEnd, gemType(gem)))
		}
		t.Append([]string{a, b})
	}
	content += t.String()
	return content
}

func renderDaoRequestGenerator(params map[string]string) string {
	// General checks
	user, ok := params["user"]
	if !ok {
		return "user is not specified. [Go back to home](" + gnominersPath + ") to select a user."
	}
	_, ok = userSessions[user]
	if !ok {
		return "this user is not registered. [Go back to home](" + gnominersPath + ") to register."
	}

	// Content creator
	var content string
	req := generateReq(params)
	content += md.H1("Generate a request")
	content += "Go back to " + md.Link("DashBoard", gnominersDashboard + "?user=" + user) + br
	content += md.H2(md.Link("Click here when you are done to Generate", gnominersDAO+"?user="+user+"&request="+strings.Join(req, ",")))
	content += displayFinal(req) + "\n"
	content += md.HorizontalRule() + br
	content += "<gno-columns>" + br
	content += displayCount(params)
	content += "|||" + br
	content += displayTable(params, req)
	content += "</gno-columns>" + br
	return content
}
