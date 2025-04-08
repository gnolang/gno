package gnominers

import (
	"errors"
	"math"
	"strconv"
	"strings"

	"gno.land/p/demo/ufmt"
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
	content += "### How much to convert : " + strconv.Itoa(int(count)) + "\n"
	content += "[-100](" + redoLink(params, int(math.Max(1, count-100)), _gemEnd, _gemEnd) + ") "
	content += "[-10](" + redoLink(params, int(math.Max(1, count-10)), _gemEnd, _gemEnd) + ") "
	content += "[-1](" + redoLink(params, int(math.Max(1, count-1)), _gemEnd, _gemEnd) + ") "
	content += "|"
	content += "[+1](" + redoLink(params, int(count)+1, _gemEnd, _gemEnd) + ") "
	content += "[+10](" + redoLink(params, int(count)+10, _gemEnd, _gemEnd) + ") "
	content += "[+100](" + redoLink(params, int(count)+100, _gemEnd, _gemEnd) + ")\n\n"
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
	return ufmt.Sprintf("## Convertion: %s %s \u2192 %s\n", req[0], gemNames[mustAtoi(req[1])], gemNames[mustAtoi(req[2])])
}

func displayTable(params map[string]string, req []string) string {
	var content string

	content += "### Gem Selecion\n"
	content += "|From|To|\n|-|-|\n"
	for gem, gemName := range gemNames {
		switch gem {
		case mustAtoi(req[2]): content += ufmt.Sprintf("|[%s](%s)", gemName, redoLink(params, -1, gemType(mustAtoi(req[2])), gemType(mustAtoi(req[1]))))
		case mustAtoi(req[1]): content += ufmt.Sprintf("|**%s**", gemName)
		default: content += ufmt.Sprintf("|[%s](%s)", gemName, redoLink(params, -1, gemType(gem), _gemEnd))
		}

		switch gem {
		case mustAtoi(req[1]): content += ufmt.Sprintf("|[%s](%s)|\n", gemName, redoLink(params, -1, gemType(mustAtoi(req[2])), gemType(mustAtoi(req[1]))))
		case mustAtoi(req[2]): content += ufmt.Sprintf("|**%s**|\n", gemName)
		default: content += ufmt.Sprintf("|[%s](%s)|\n", gemName, redoLink(params, -1, _gemEnd, gemType(gem)))
		}
	}
	content += "\n"
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
	content += "# Generate a request\n"
	content += "Go back to [DashBoard](" + gnominersDashboard + "?user=" + user + ").\n\n"
	content += ufmt.Sprintf("## [Click here when you are done to Generate](%s) !\n", gnominersDAO+"?user="+user+"&request="+strings.Join(req, ","))
	content += displayFinal(req) + "\n"
	content += displayCount(params)
	content += displayTable(params, req)
	return content
}
