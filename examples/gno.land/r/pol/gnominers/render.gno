package gnominers

import (
	"time"
	"gno.land/p/pol/queryparsing"
	"gno.land/p/pol/notrandom"
)

var randomMachine *notrandom.Machine

const (
	gnominersPath = "/r/pol/gnominers"
	gnominersShopPath = gnominersPath + ":shop"
	gnominersDashboard = gnominersPath + ":dashboard"
	gnominersAchievements = gnominersPath + ":achievements"
)

var redirectionsPath = map[string]func(params map[string]string) string {
	"shop": renderShop,
	"dashboard": renderDashboard,
	"achievements": renderAchievements,
}

// the rendering

func init() {
	randomMachine = notrandom.NewMachine(time.Now().Unix())
}

func Render(fullpath string) string {
	// RANDOM_MACHINE = notrandom.NewMachine(time.Now().Unix())
	path, params := queryparsing.Parse(fullpath)
	if redirect, ok := redirectionsPath[path]; ok {
		return redirect(params)
	}
	// if _, ok := params["user"]; ok {
	// 	return renderDashboard(params)
	// }
	return renderUserPage(params)
}
