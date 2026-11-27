package vesting

import (
	"chain/banker"
	"strconv"
	"strings"

	"gno.land/p/nt/markdown/sanitize/v0"
	vst "gno.land/p/nt/vesting/v0"

	"gno.land/r/nt/grc20reg/v0"
)

// Render serves the factory or one grant.
// Paths: "" and "<grantID>".
func Render(path string) string {
	if len(path) > 64 {
		return "# Vesting vault\n\nUnknown grant.\n"
	}
	path = strings.Trim(path, "/")
	if path == "" {
		return renderHome()
	}
	if strings.Contains(path, "/") {
		return "# Vesting vault\n\nUnknown grant.\n"
	}
	id, err := strconv.ParseInt(path, 10, 64)
	if err != nil || id <= 0 || id > int64(len(grants)) {
		return "# Vesting vault\n\nUnknown grant.\n"
	}
	return renderGrant(id)
}

func renderHome() string {
	b := new(strings.Builder)
	b.WriteString("# Vesting vault\n\n")
	b.WriteString("Anyone can lock a native coin or a registered GRC20 on a cliff and a vesting curve. Funds live at the grant address, not at this factory.\n\n")
	b.WriteString("This is unaudited v0.\n\n")
	b.WriteString("Grants: " + strconv.Itoa(len(grants)) + "\n\n")
	if len(grants) == 0 {
		b.WriteString("No grants yet.\n")
		return b.String()
	}
	start := 0
	if len(grants) > 20 {
		start = len(grants) - 20
		b.WriteString("Showing the latest 20.\n\n")
	}
	for i := start; i < len(grants); i++ {
		g := grants[i]
		id := strconv.FormatInt(g.ID(), 10)
		name := g.Label()
		if name == "" {
			name = "grant " + id
		}
		b.WriteString("- [" + sanitize.InlineText(name) + "](" + rel() + ":" + id + ")")
		b.WriteString(" — " + g.Asset().Kind.String())
		b.WriteString(" — `" + sanitize.InlineText(g.Beneficiary().String()) + "`\n")
	}
	return b.String()
}

func renderGrant(id int64) string {
	g := grants[id-1]
	now := nowUnix()
	view, err := g.View(now)
	if err != nil {
		return "# Vesting vault\n\nUnknown grant.\n"
	}
	b := new(strings.Builder)
	title := view.Label
	if title == "" {
		title = "Grant " + strconv.FormatInt(view.ID, 10)
	}
	b.WriteString("# " + sanitize.InlineText(title) + "\n\n")
	b.WriteString("Address: `" + sanitize.InlineText(view.Address.String()) + "`\n\n")
	b.WriteString("Admin: `" + sanitize.InlineText(view.Admin.String()) + "`\n\n")
	b.WriteString("Beneficiary: `" + sanitize.InlineText(view.Beneficiary.String()) + "`\n\n")
	b.WriteString("Asset: " + view.Asset.Kind.String() + " `" + sanitize.InlineText(assetRef(view)) + "`\n\n")
	b.WriteString("Curve: " + view.Terms.Curve.String() + "\n\n")
	b.WriteString("Start: " + strconv.FormatInt(view.Terms.Start, 10) + "\n\n")
	b.WriteString("Cliff: " + strconv.FormatInt(view.Terms.Cliff, 10) + " seconds\n\n")
	b.WriteString("Duration: " + strconv.FormatInt(view.Terms.Duration, 10) + " seconds\n\n")
	if view.Terms.Curve == vst.CurveStep {
		b.WriteString("Period: " + strconv.FormatInt(view.Terms.Period, 10) + " seconds\n\n")
	}
	b.WriteString("Total: " + strconv.FormatInt(view.Terms.Total, 10) + "\n\n")
	b.WriteString("Vested: " + strconv.FormatInt(view.Vested, 10) + "\n\n")
	b.WriteString("Claimable: " + strconv.FormatInt(view.Releasable, 10) + "\n\n")
	b.WriteString("Claimed: " + strconv.FormatInt(view.Claimed, 10) + "\n\n")
	b.WriteString("Returned: " + strconv.FormatInt(view.Returned, 10) + "\n\n")
	b.WriteString("Balance: " + strconv.FormatInt(balanceView(view), 10) + "\n\n")
	switch {
	case view.Revoked:
		b.WriteString("Status: revoked at " + strconv.FormatInt(view.RevokedAt, 10) + ". Unvested tokens went back to the admin. Vested tokens stay claimable.\n\n")
	case view.Vested == view.Terms.Total:
		b.WriteString("Status: fully vested.\n\n")
	case view.Vested == 0:
		b.WriteString("Status: nothing vested yet.\n\n")
	default:
		b.WriteString("Status: vesting.\n\n")
	}
	switch {
	case view.Revoked:
		b.WriteString("Revoke: already revoked.\n\n")
	case view.Revocable:
		b.WriteString("Revoke: the admin may still revoke.\n\n")
	default:
		b.WriteString("Revoke: irrevocable.\n\n")
	}
	b.WriteString("Send the asset to the address above only if you mean to fund this grant. A direct send is not recorded in the schedule.\n")
	return b.String()
}

func assetRef(view vst.GrantView) string {
	// The stored denom or registry key. Render does not call the token
	// for its symbol: that string is chosen by the token's issuer.
	if view.Asset.Kind == vst.AssetGRC20 {
		return view.Asset.Token
	}
	return view.Asset.Denom
}

func balanceView(view vst.GrantView) int64 {
	if view.Asset.Kind == vst.AssetGRC20 {
		tok := grc20reg.Get(view.Asset.Token)
		if tok == nil {
			return 0
		}
		return tok.BalanceOf(view.Address)
	}
	return bankerGet(view.Address, view.Asset.Denom)
}

func bankerGet(addr address, denom string) int64 {
	return banker.NewReadonlyBanker().GetCoin(addr, denom)
}

func rel() string {
	const prefix = "gno.land"
	if strings.HasPrefix(realmPath, prefix) {
		return strings.TrimPrefix(realmPath, prefix)
	}
	return realmPath
}
