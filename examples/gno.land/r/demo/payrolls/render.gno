package payrolls

import (
	"net/url"
	"std"
	"strconv"
	"strings"
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/seqid"
	"gno.land/p/demo/ufmt"
	"gno.land/p/demo/users"
	rusers "gno.land/r/demo/users"
)

const listMaxSize = 50

func Render(renderPath string) string {
	w := strings.Builder{}

	switch {
	case strings.HasPrefix(renderPath, "/u/"):
		w.WriteString("# Payrolls beneficiary\n")

		path := strings.TrimPrefix(renderPath, "/u/")
		aon := users.AddressOrName(path)
		user := rusers.GetUserByAddressOrName(aon)
		addr := std.Address(aon)
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
		estimatedDay := uint64(0)
		estimatedAvailable := uint64(0)
		estimatedClaimable := uint64(0)
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
				estimatedAvailable += available
				vaultAmount := getVaultAmount(p.vaultID)
				if available > vaultAmount {
					available = vaultAmount
				}
				estimatedClaimable += available
				estimatedDay += p.estimateDay()
			}

			return false
		})

		if shouldEstimate {
			w.WriteString(renderEstimatedMonthly("", estimatedDay, denom))
			w.WriteString(renderClaimable("", estimatedClaimable, estimatedAvailable, denom, true))
		}

		actions := []string{}

		if !shouldEstimate || estimatedClaimable > 0 {
			actions = append(actions, renderButton("Claim all", "ClaimAll", map[string]string{
				"destination": addr.String(),
			}))
		}

		actions = append(actions, renderButton("New payroll", "CreateCDI", map[string]string{
			"label":          "Salary Monthly",
			"beneficiary":    addr.String(),
			"amountPerMonth": "10000",
		}))

		w.WriteString(renderActions("", actions))

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
		estimatedDay := uint64(0)
		estimatedDebt := uint64(0)
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
				estimatedDebt += p.available()
				estimatedDay += p.estimateDay()
			}
			payrollsBuffer.WriteString(renderPayroll(p, "vault"))
			return false
		})

		funds := getVaultAmount(vaultID)
		balance := int64(funds) - int64(estimatedDebt)

		if shouldEstimate {
			w.WriteString(renderEstimatedMonthly("", estimatedDay, denom))

			prefix := "Funds: 🟢 "
			if balance < 0 {
				prefix = "Debt: 🔴 "
			}
			w.WriteString(ufmt.Sprintf("- 🏦 %s%d%s\n", prefix, balance, denom))

			if balance > 0 && estimatedDay > 0 {
				color := "🔴"
				runWayMonths := funds / (estimatedDay * 30)
				if runWayMonths >= 12 {
					color = "🟢"
				} else if runWayMonths >= 6 {
					color = "🟠"
				}
				w.WriteString(ufmt.Sprintf("- 🛬 Estimated runway: %s %d months\n", color, runWayMonths))
			}
		} else {
			w.WriteString(ufmt.Sprintf("- 🏦 Funds: %d%s (does not account for unclaimed amounts)\n", funds, denom))
		}

		actions := []string{}
		actions = append(actions, renderButton("Deposit funds", "Fund", map[string]string{
			"base":      base,
			"namespace": namespace,
		}))
		if !shouldEstimate || balance > 0 {
			actions = append(actions, renderButton("Withdraw funds", "WithdrawFunds", map[string]string{
				"namespace": namespace,
				"amount":    "-1",
			}))
		}
		actions = append(actions, renderButton("New payroll", "CreateCDI", map[string]string{
			"namespace":      namespace,
			"label":          "Salary Monthly",
			"amountPerMonth": "10000",
		}))
		w.WriteString(renderActions("", actions))

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
		actions = append(actions, renderButton("New payroll", "CreateCDI", map[string]string{
			"label":          "Salary Monthly",
			"amountPerMonth": "10000",
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

		w.WriteString("## TODO\n")
		w.WriteString("- [ ] Show created at\n")
		w.WriteString("- [ ] Show running duration\n")
		w.WriteString("- [ ] Support any coin\n")
		w.WriteString("- [ ] Render breakup amount\n")
		w.WriteString("- [ ] Unit tests\n")
		w.WriteString("- [ ] txtar\n")
		w.WriteString("- [ ] Write doc\n")

		w.WriteString("## v2\n")
		w.WriteString("- [ ] Gnoweb pagination\n")
		w.WriteString("- [ ] Archive\n")
		w.WriteString("- [ ] Details: Events history\n")
		w.WriteString("- [ ] PoV Tags\n")
		w.WriteString("- [ ] Edit label / PoV labels\n")
		w.WriteString("- [ ] Stop litigation\n")

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
	vaultAmount := getVaultAmount(p.vaultID)
	claimable := available
	if available > vaultAmount {
		claimable = vaultAmount
	}
	if available > 0 {
		sb.WriteString(renderClaimable("  ", claimable, available, denom, false))
	}

	sb.WriteString(ufmt.Sprintf("  - 🏁 Created at: %s\n", p.createdAt.Format(time.DateTime)))

	if !p.stopped {
		estimatedDay := p.estimateDay()
		sb.WriteString(renderEstimatedMonthly("  ", estimatedDay, denom))

		sb.WriteString("  - 🕙 Remaining: ")
		if p.cap >= 0 {
			amountRemaining := uint64(p.cap) - p.totalWithdrawn
			sb.WriteString(ufmt.Sprintf("%s (%d%s)", ((time.Duration(amountRemaining) * 24) / time.Duration(estimatedDay) / time.Hour).String(), amountRemaining, denom))
		} else {
			sb.WriteRune('∞')
		}
		sb.WriteRune('\n')
	} else {
		sb.WriteString(ufmt.Sprintf("  - 🛑 Stopped since: %s\n", p.stoppedAt.Format(time.DateTime)))
	}

	if pov != "details" {
		sb.WriteString(ufmt.Sprintf("  - 🔍 [Details](%s:/d/%s)\n", linkPrefix(), p.id.String()))
	}

	actions := []string{}

	if pov != "vault" && claimable > 0 {
		actions = append(actions, renderButton("Claim", "Claim", map[string]string{
			"id":              strconv.FormatUint(uint64(p.id), 10),
			"requestedAmount": "-1",
			"destination":     p.beneficiary.String(),
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

	return sb.String()
}

func renderClaimable(prefix string, claimable uint64, available uint64, denom string, plural bool) string {
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
		sb.WriteString(ufmt.Sprintf("%s %d%s/%d%s (not enough funds in vault%s!)\n", color, claimable, denom, available, denom, pls))
	} else {
		sb.WriteString(ufmt.Sprintf("🟢 %d%s\n", available, denom))
	}
	return sb.String()
}

func renderEstimatedMonthly(prefix string, estimatedDay uint64, denom string) string {
	return ufmt.Sprintf("%s- 💸 Estimated monthly: %d%s\n", prefix, estimatedDay*30, denom)
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

	if strings.HasPrefix(base, std.GetChainDomain()+"/r/") {
		return base, nil, true
	}

	usr := rusers.GetUserByAddressOrName(users.AddressOrName(base))
	if usr != nil {
		return usr.Address.String(), usr, true
	}

	if std.Address(base).IsValid() {
		return base, nil, true
	}

	return "", nil, false
}

func linkPrefix() string {
	realmPath := std.CurrentRealm().PkgPath()
	slashIdx := strings.Index(realmPath, "/")
	if slashIdx == -1 {
		return realmPath
	}
	return realmPath[slashIdx:]
}
