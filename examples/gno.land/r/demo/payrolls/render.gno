package payrolls

import (
	"chain"
	"chain/runtime"
	"math"
	"net/url"
	"strconv"
	"strings"
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/seqid"
	"gno.land/p/demo/ufmt"
	"gno.land/p/demo/users"
	"gno.land/r/demo/grc20reg"
	rusers "gno.land/r/demo/users"
)

const (
	listMaxSize   = 50
	defaultFqname = "gno.land/r/demo/foo20"
	defaultDenom  = "/grc20/" + defaultFqname
)

func Render(renderPath string) string {
	w := strings.Builder{}

	switch {
	case strings.HasPrefix(renderPath, "/u/"):
		w.WriteString("# Payrolls beneficiary\n")

		path := strings.TrimPrefix(renderPath, "/u/")
		aon := users.AddressOrName(path)
		user := rusers.GetUserByAddressOrName(aon)
		addr := address(aon)
		if user != nil {
			addr = user.Address
			w.WriteString(ufmt.Sprintf("- 👤 User: @%s\n", user.Name))
		} else if !addr.IsValid() {
			w.WriteString(ufmt.Sprintf("❓ User %q not found\n", addr.String()))
			break
		}

		w.WriteString(ufmt.Sprintf("- 🆔 Address: %s\n", addr.String()))

		sb := strings.Builder{}
		tree, ok := payrollsByBeneficiary[addr]
		if !ok {
			tree = avl.NewTree()
		}
		estimatedDay := chain.Coins{}
		estimatedAvailable := chain.Coins{}
		estimatedClaimable := chain.Coins{}
		shouldEstimate := tree.Size() <= listMaxSize
		if shouldEstimate {
			sb.WriteString("## All payrolls\n")
		} else {
			sb.WriteString("## Latest payrolls\n")
		}
		tree.ReverseIterateByOffset(0, listMaxSize, func(key string, value interface{}) bool {
			p := value.(*payroll)
			sb.WriteString(renderPayroll(p, "beneficiary"))

			if shouldEstimate {
				available := p.available()
				estimatedAvailable = addCoins(estimatedAvailable, available)

				claimable := clampCoins(available, getVaultCoins(p.vaultID))
				estimatedClaimable = addCoins(estimatedClaimable, claimable)

				estimatedDay = addCoins(estimatedDay, p.estimateDay())
			}

			return false
		})

		actions := []string{}

		if !shouldEstimate || !coinsEmpty(estimatedClaimable) {
			actions = append(actions, renderButton("Claim all", "ClaimAll", map[string]string{
				"destination": addr.String(),
			}))
		}

		actions = append(actions, renderButton("New payroll", "CreateMonthlyContinuous", map[string]string{
			"label":          "Salary Monthly",
			"beneficiary":    addr.String(),
			"amountPerMonth": "10000",
			"denom":          defaultDenom,
		}))

		w.WriteString(renderActions("", actions))

		if shouldEstimate {
			for _, coin := range addCoins(estimatedAvailable, estimatedDay) {
				w.WriteString(ufmt.Sprintf("- %s:\n", renderCoinDenom(coin.Denom)))
				w.WriteString(renderEstimatedMonthly("  ", estimatedDay.AmountOf(coin.Denom), coin.Denom))
				w.WriteString(renderClaimable("  ", estimatedClaimable.AmountOf(coin.Denom), estimatedAvailable.AmountOf(coin.Denom), coin.Denom, true))
			}
		}

		w.WriteString(sb.String())

	case strings.HasPrefix(renderPath, "/v/"):
		w.WriteString("# Payrolls vault\n")

		input := strings.TrimPrefix(renderPath, "/v/")
		inputBase, namespace := splitVaultInput(input)
		base, vaultUser, ok := resolveVaultBase(inputBase)
		if !ok {
			w.WriteString(ufmt.Sprintf("❓ Vault %q not found", input))
			break
		}

		vaultID := vaultIDUnsafe(base, namespace)

		name := vaultID
		if vaultUser != nil {
			name = vaultIDUnsafe("@"+vaultUser.Name, namespace)
		}
		w.WriteString(ufmt.Sprintf("- 🎩 Name: %s\n", name))

		w.WriteString(ufmt.Sprintf("- 🆔 ID: %s\n", vaultID))

		payrollsBuffer := strings.Builder{}
		estimatedDay := chain.Coins{}
		estimatedDebt := chain.Coins{}
		tree, ok := payrollsByVault[vaultID]
		if !ok {
			tree = avl.NewTree()
		}
		shouldEstimate := tree.Size() <= listMaxSize
		if shouldEstimate {
			payrollsBuffer.WriteString("## All payrolls\n")
		} else {
			payrollsBuffer.WriteString("## Latest payrolls\n")
		}
		tree.ReverseIterateByOffset(0, listMaxSize, func(key string, value interface{}) bool {
			p := value.(*payroll)
			if shouldEstimate {
				estimatedDebt = addCoins(estimatedDebt, p.available())
				estimatedDay = addCoins(estimatedDay, p.estimateDay())
			}
			payrollsBuffer.WriteString(renderPayroll(p, "vault"))
			return false
		})

		funds := getVaultCoins(vaultID)
		balance := subCoins(funds, estimatedDebt)

		actions := []string{}
		actions = append(actions, renderButton("Deposit native", "FundNative", map[string]string{
			"base":      base,
			"namespace": namespace,
		}))
		actions = append(actions, renderButton("Deposit GRC20", "FundGRC20Reg", map[string]string{
			"base":      base,
			"namespace": namespace,
			"fqname":    defaultFqname,
			"amount":    "-1",
		}))
		if !shouldEstimate || coinsHasPositive(balance) {
			actions = append(actions, renderButton("Withdraw funds", "WithdrawFunds", map[string]string{
				"namespace":   namespace,
				"amount":      "-1",
				"denom":       defaultDenom,
				"destination": base,
			}))
		}
		actions = append(actions, renderButton("New payroll", "CreateMonthlyContinuous", map[string]string{
			"namespace":      namespace,
			"label":          "Salary Monthly",
			"amountPerMonth": "10000",
			"denom":          defaultDenom,
		}))
		w.WriteString(renderActions("", actions))

		if shouldEstimate {
			for _, coin := range addCoins(balance, estimatedDay) {
				w.WriteString(ufmt.Sprintf("- %s:\n", renderCoinDenom(coin.Denom)))

				color := "🟢"
				coinBalance := balance.AmountOf(coin.Denom)
				if coinBalance < 0 {
					color = "🔴"
				}
				w.WriteString(ufmt.Sprintf("  - 🏦 Funds: %s %s\n", color, renderCoin(chain.NewCoin(coin.Denom, coinBalance))))

				w.WriteString(renderEstimatedMonthly("  ", estimatedDay.AmountOf(coin.Denom), coin.Denom))

				if coinBalance > 0 && estimatedDay.AmountOf(coin.Denom) > 0 {
					color := "🔴"
					runWayMonths := coinBalance / (estimatedDay.AmountOf(coin.Denom) * 30)
					if runWayMonths >= 12 {
						color = "🟢"
					} else if runWayMonths >= 6 {
						color = "🟠"
					}
					w.WriteString(ufmt.Sprintf("  - 🛬 Estimated runway: %s %d months\n", color, runWayMonths))
				}
			}
		} else {
			for _, coin := range funds {
				w.WriteString(ufmt.Sprintf("- 🏦 Funds: %d %s (does not account for unclaimed amounts)\n", renderCoin(coin)))
			}
		}

		w.WriteString(payrollsBuffer.String())

	case strings.HasPrefix(renderPath, "/d/"):
		idStr := strings.TrimPrefix(renderPath, "/d/")

		w.WriteString(ufmt.Sprintf("# Payroll #%s\n", idStr))

		p, ok := payrolls.Get(idStr)
		if !ok {
			return ufmt.Sprintf("❓ Payroll %q not found", idStr)
		}

		w.WriteString(renderPayroll(p.(*payroll), "details"))

	case renderPath == "":
		w.WriteString("# Payrolls\n")

		actions := []string{}
		actions = append(actions, renderButton("New payroll", "CreateMonthlyContinuous", map[string]string{
			"label":          "Salary Monthly",
			"amountPerMonth": "10000",
			"denom":          defaultDenom,
		}))
		w.WriteString(renderActions("", actions))

		if payrolls.Size() <= listMaxSize {
			w.WriteString("## All payrolls\n")
		} else {
			w.WriteString("## Latest payrolls\n")
		}
		payrolls.ReverseIterateByOffset(0, listMaxSize, func(key string, value interface{}) bool {
			w.WriteString(renderPayroll(value.(*payroll), "home"))
			return false
		})

		/*
			w.WriteString("## v2 ideas\n")
			w.WriteString("- Better test coverage\n")
			w.WriteString("- Gnoweb pagination\n")
			w.WriteString("- Show payroll running duration\n")
			w.WriteString("- Archive\n")
			w.WriteString("- Details: Events history\n")
			w.WriteString("- PoV Tags\n")
			w.WriteString("- Edit label / PoV labels\n")
			w.WriteString("- Stop litigation\n")
		*/

	default:
		w.WriteString("# Payrolls\n")
		w.WriteString("❓ 404 Not found")
	}

	return w.String()
}

func renderPayroll(p *payroll, pov string) string {
	sb := strings.Builder{}

	if pov != "details" {
		sb.WriteString(ufmt.Sprintf("### %s\n", p.id.String()))
	}

	sb.WriteString(ufmt.Sprintf("  - ✍️ Label: %s\n", p.label))

	if p.paused {
		sb.WriteString(ufmt.Sprintf("  - ⏸️ Paused since: %s\n", p.pausedAt.Format(time.DateTime)))
	}

	if pov != "vault" {
		base, namespace := splitVaultInput(p.vaultID)
		_, usr, _ := resolveVaultBase(base)
		if usr != nil {
			base = "@" + usr.Name
		}
		pvid := vaultIDUnsafe(base, namespace)
		sb.WriteString(ufmt.Sprintf("  - 🎩 Vault: [%s](%s:/v/%s)\n", pvid, linkPrefix(), pvid))
	}

	if pov != "beneficiary" {
		beneficiaryUser := rusers.GetUserByAddress(p.beneficiary)
		var userStr string
		if beneficiaryUser == nil {
			userStr = p.beneficiary.String()
		} else {
			userStr = "@" + beneficiaryUser.Name
		}
		sb.WriteString(ufmt.Sprintf("  - 👤 Beneficiary: [%s](%s:/u/%s)\n", userStr, linkPrefix(), userStr))
	}

	available := p.available()
	vaultAmount := getVaultCoins(p.vaultID)
	claimable := clampCoins(available, vaultAmount)

	sb.WriteString(ufmt.Sprintf("  - 🏁 Created at: %s\n", p.createdAt.Format(time.DateTime)))

	if p.stopped {
		sb.WriteString(ufmt.Sprintf("  - 🛑 Stopped since: %s\n", p.stoppedAt.Format(time.DateTime)))
	}

	if pov != "details" {
		sb.WriteString(ufmt.Sprintf("  - 🔍 [Details](%s:/d/%s)\n", linkPrefix(), p.id.String()))
	}

	actions := []string{}

	if pov != "vault" && coinsHasPositive(claimable) {
		actions = append(actions, renderButton("Claim", "Claim", map[string]string{
			"id":          strconv.FormatUint(uint64(p.id), 10),
			"destination": p.beneficiary.String(),
		}))
	}

	if pov != "vault" && !p.paused && !p.stopped {
		actions = append(actions, renderButton("Pause", "Pause", map[string]string{
			"id": strconv.FormatUint(uint64(p.id), 10),
		}))
	}

	if pov != "vault" && p.paused {
		actions = append(actions, renderButton("Resume", "Resume", map[string]string{
			"id": strconv.FormatUint(uint64(p.id), 10),
		}))
	}

	if !p.stopped {
		actions = append(actions, renderStopButton(p.id))
	}

	sb.WriteString(renderActions("  ", actions))

	estimatedDay := p.estimateDay()
	allCoins := addCoins(available, estimatedDay)

	var beneficiaryBreakupCoins chain.Coins
	var creatorBreakupCoins chain.Coins
	if pov == "details" {
		allCoins = addCoins(allCoins, beneficiaryBreakupCoins)
		allCoins = addCoins(allCoins, creatorBreakupCoins)
		beneficiaryBreakupCoins = p.getBreakupCoins(CallSourceBeneficiary)
		creatorBreakupCoins = p.getBreakupCoins(CallSourceCreator)
	}
	for _, coin := range allCoins {
		sb.WriteString(ufmt.Sprintf("  - %s:\n", renderCoinDenom(coin.Denom)))
		sb.WriteString(renderEstimatedMonthly("    ", estimatedDay.AmountOf(coin.Denom), coin.Denom))
		sb.WriteString(renderClaimable("    ", claimable.AmountOf(coin.Denom), available.AmountOf(coin.Denom), coin.Denom, true))
		if pov == "details" {
			if bb := beneficiaryBreakupCoins.AmountOf(coin.Denom); bb > 0 {
				sb.WriteString(ufmt.Sprintf("    - Breakup by beneficiary bonus: %s\n", renderCoin(chain.NewCoin(coin.Denom, bb))))
			}
			if cb := creatorBreakupCoins.AmountOf(coin.Denom); cb > 0 {
				sb.WriteString(ufmt.Sprintf("    - Breakup by creator bonus: %s\n", renderCoin(chain.NewCoin(coin.Denom, cb))))
			}
		}
	}

	return sb.String()
}

func renderClaimable(prefix string, claimable int64, available int64, denom string, plural bool) string {
	if available == 0 {
		return ""
	}
	sb := strings.Builder{}
	sb.WriteString(prefix)
	sb.WriteString("- 💵 Claimable: ")
	if claimable < available {
		pls := ""
		if plural {
			pls = "(s)"
		}
		color := "🟠"
		if claimable == 0 {
			color = "🔴"
		}
		sb.WriteString(ufmt.Sprintf("%s %s/%s %s (not enough funds in vault%s!)\n",
			color,
			renderCoinAmount(chain.NewCoin(denom, claimable)),
			renderCoinAmount(chain.NewCoin(denom, available)),
			renderCoinDenom(denom),
			pls,
		))
	} else {
		sb.WriteString(ufmt.Sprintf("🟢 %s\n", renderCoin(chain.NewCoin(denom, available))))
	}
	return sb.String()
}

func renderEstimatedMonthly(prefix string, estimatedDay int64, denom string) string {
	if estimatedDay == 0 {
		return ""
	}
	return ufmt.Sprintf("%s- 💸 Estimated monthly: %s\n", prefix, renderCoin(chain.NewCoin(denom, estimatedDay*30)))
}

func renderStopButton(id seqid.ID) string {
	return renderButton("Stop", "Stop", map[string]string{
		"id": strconv.FormatUint(uint64(id), 10),
	})
}

func renderButton(text, funcName string, args map[string]string) string {
	argsBuf := strings.Builder{}
	for key, arg := range args {
		argsBuf.WriteString("&" + url.QueryEscape(key) + "=" + url.QueryEscape(arg))
	}
	return ufmt.Sprintf("[[%s]](%s$help&func=%s%s)", text, linkPrefix(), funcName, argsBuf.String())
}

func renderActions(prefix string, actions []string) string {
	if len(actions) == 0 {
		return ""
	}
	return prefix + "- ✋ Actions: " + strings.Join(actions, " ") + "\n"
}

func resolveVaultBase(base string) (string, *users.User, bool) {
	if len(base) == 0 {
		return "", nil, false
	}

	if strings.HasPrefix(base, runtime.ChainDomain()+"/r/") {
		return base, nil, true
	}

	usr := rusers.GetUserByAddressOrName(users.AddressOrName(base))
	if usr != nil {
		return usr.Address.String(), usr, true
	}

	if address(base).IsValid() {
		return base, nil, true
	}

	return "", nil, false
}

func linkPrefix() string {
	realmPath := runtime.CurrentRealm().PkgPath()
	slashIdx := strings.Index(realmPath, "/")
	if slashIdx == -1 {
		return realmPath
	}
	return realmPath[slashIdx:]
}

func renderCoin(coin chain.Coin) string {
	return ufmt.Sprintf("%s %s", renderCoinAmount(coin), renderCoinDenom(coin.Denom))
}

func renderCoinAmount(coin chain.Coin) string {
	switch {
	case strings.HasPrefix(coin.Denom, "/grc20/"):
		fqdenom := strings.TrimPrefix(coin.Denom, "/grc20/")
		tokenGetter := grc20reg.Get(fqdenom)
		if tokenGetter == nil {
			break
		}
		token := tokenGetter()
		if token == nil {
			break
		}
		amount := float64(coin.Amount) / float64(math.Pow10(int(token.GetDecimals())))
		return ufmt.Sprintf("%g", amount)
	}

	return ufmt.Sprintf("%d", coin.Amount)
}

func renderCoinDenom(denom string) string {
	switch {
	case strings.HasPrefix(denom, "/native/"):
		nativeDenom := strings.TrimPrefix(denom, "/native/")
		return nativeDenom

	case strings.HasPrefix(denom, "/grc20/"):
		fqdenom := strings.TrimPrefix(denom, "/grc20/")
		tokenGetter := grc20reg.Get(fqdenom)
		if tokenGetter == nil {
			break
		}
		token := tokenGetter()
		if token == nil {
			break
		}
		return ufmt.Sprintf("[$%s](%s)", tokenGetter().GetSymbol(), strings.TrimPrefix(fqdenom, runtime.ChainDomain()))
	}

	return denom
}
