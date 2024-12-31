package payrolls

import (
	"net/url"
	"path"
	"std"
	"strings"
	"time"

	"gno.land/p/demo/mux"
	"gno.land/p/demo/ufmt"
	"gno.land/p/demo/users"
	rusers "gno.land/r/demo/users"
)

// TODO: pagination
const listMaxSize = 50

func Render(renderPath string) string {
	router := mux.NewRouter()

	router.HandleFunc("/u/*", func(w *mux.ResponseWriter, r *mux.Request) {
		w.Write("# Payrolls beneficiary\n")
		aon := users.AddressOrName(path.Base(r.Path))
		user := rusers.GetUserByAddressOrName(aon)
		addr := std.Address(aon)
		name := string(aon)
		if user != nil {
			addr = user.Address
			name = "@" + user.Name
			w.Write(ufmt.Sprintf("- 👤 User: @%s\n", user.Name))
		} else if !addr.IsValid() {
			w.Write(ufmt.Sprintf("❓ User %q not found\n", addr.String()))
			return
		}
		w.Write(ufmt.Sprintf("- 🆔 Address: %s\n", addr.String()))

		sb := strings.Builder{}
		tree, ok := payrollsByBeneficiary[addr]
		if !ok {
			return
		}
		estimatedDay := uint64(0)
		estimatedFree := uint64(0)
		estimatedClaimable := uint64(0)
		shouldEstimate := tree.Size() <= listMaxSize
		tree.ReverseIterateByOffset(0, listMaxSize, func(key string, value interface{}) bool {
			p := value.(*payroll)
			sb.WriteString(p.render("beneficiary"))

			if shouldEstimate {
				free := p.freeAt(time.Now())
				estimatedFree += free
				vault := getVaultAmount(p.vaultID)
				if vault > free {
					vault = free
				}
				estimatedClaimable += vault
				estimatedDay += p.freeAt(time.Now().Add(time.Hour*24)) - free
			}

			return false
		})

		if shouldEstimate {
			w.Write(renderEstimatedMonthly("", estimatedDay, denom))
			w.Write(renderClaimable("", estimatedClaimable, estimatedFree, denom, true))
		}

		actions := []string{}

		if !shouldEstimate || estimatedClaimable > 0 {
			claimAllButton := ufmt.Sprintf("[[Claim all]]($help&func=ClaimAll&destination=%s)", addr.String())
			actions = append(actions, claimAllButton)
		}

		label := url.QueryEscape(ufmt.Sprintf("%s Salary Monthly", name))
		newPayrollButton := ufmt.Sprintf("[[New payroll]]($help&func=CreateMonthly&namespace=&label=%s&beneficiary=%s&amount=10000&totalAmount=-1)", label, addr.String())
		actions = append(actions, newPayrollButton)

		w.Write(ufmt.Sprintf("- ✋ Actions: %s\n", strings.Join(actions, " ")))

		w.Write(sb.String())
	})

	router.HandleFunc("/v/*", func(w *mux.ResponseWriter, r *mux.Request) {
		w.Write("# Payrolls vault\n")

		name := path.Base(r.Path)
		vaultID, err := parseVaultID(name)
		if err != nil {
			w.Write(ufmt.Sprintf("❓ Vault %q not found", name))
			return
		}

		if name != vaultID {
			w.Write(ufmt.Sprintf("- 🎩 Name: %s\n", name))
		}
		w.Write(ufmt.Sprintf("- 🆔 ID: %s\n", vaultID))

		payrollsBuffer := strings.Builder{}
		estimatedDay := uint64(0)
		estimatedDebt := uint64(0)
		tree, ok := payrollsByVault[vaultID]
		if !ok {
			return
		}
		shouldEstimate := tree.Size() <= listMaxSize
		tree.ReverseIterateByOffset(0, listMaxSize, func(key string, value interface{}) bool {
			p := value.(*payroll)
			if shouldEstimate {
				estimatedDebt += p.available()
				estimatedDay += p.freeAt(time.Now().Add(time.Hour*24)) - p.freeAt(time.Now())
			}
			payrollsBuffer.WriteString(p.render("vault"))
			return false
		})

		funds := getVaultAmount(vaultID)

		if shouldEstimate {
			w.Write(renderEstimatedMonthly("", estimatedDay, denom))

			balance := int64(funds) - int64(estimatedDebt)
			prefix := "- 🏦 Funds: 🟢 "
			if balance < 0 {
				prefix = "- 🏦 Debt 🔴 "
			}
			w.Write(ufmt.Sprintf("%s%d%s\n", prefix, balance, denom))

			if balance > 0 {
				color := "🔴"
				runWayMonths := funds / (estimatedDay * 30)
				if runWayMonths >= 12 {
					color = "🟢"
				} else if runWayMonths >= 6 {
					color = "🟠"
				}
				w.Write(ufmt.Sprintf("- 🛬 Estimated runway: %s %d months\n", color, runWayMonths))
			}
		} else {
			w.Write(ufmt.Sprintf("- 🏦 Funds: %d%s (does not account for unclaimed amounts)\n", funds, denom))
		}

		label := url.QueryEscape("Salary Monthly")
		idParts := strings.Split(vaultID, ".")
		slug := ""
		if len(idParts) > 1 {
			slug = idParts[1]
		}
		depositFundsAction := ufmt.Sprintf("[[Deposit funds]]($help&func=Fund&addr=%s&slug=%s)", idParts[0], slug)
		newPayrollAction := ufmt.Sprintf("[[New payroll]]($help&func=CreateMonthly&namespace=&label=%s&beneficiary=&amount=10000&totalAmount=-1)", label)
		w.Write(ufmt.Sprintf("- ✋ Actions: %s %s\n", depositFundsAction, newPayrollAction))

		w.Write(payrollsBuffer.String())

	})

	router.HandleFunc("/d/*", func(w *mux.ResponseWriter, r *mux.Request) {
		w.Write("# Payroll details\n")
		idStr := path.Base(r.Path)
		p, ok := payrolls.Get(idStr)
		if !ok {
			w.Write(ufmt.Sprintf("❓ Payroll %q not found", idStr))
			return
		}
		w.Write(p.(*payroll).render(""))
	})

	router.HandleFunc("*", func(w *mux.ResponseWriter, r *mux.Request) {
		w.Write("# Latest payrolls\n")
		payrolls.ReverseIterateByOffset(0, listMaxSize, func(key string, value interface{}) bool {
			w.Write(value.(*payroll).render(""))
			return false
		})

		w.Write("\n\nTODO\n")
		w.Write("- [ ] Stop / Archive\n")
		w.Write("- [ ] Home actions")
		w.Write("- [ ] Pause / Unpause\n")
		w.Write("- [ ] Details: Events history\n")
		w.Write("- [ ] PoV Tags\n")
		w.Write("- [ ] Edit label / PoV labels\n")
		w.Write("- [ ] Stop litigation\n")
		w.Write("- [ ] Fix anchors\n")
	})

	return router.Render(renderPath)
}

func (p *payroll) render(pov string) string {
	sb := strings.Builder{}

	sb.WriteString(ufmt.Sprintf("### [%s](/r/demo/payrolls:/d/%s)\n", p.id.String(), p.id.String()))

	sb.WriteString(ufmt.Sprintf("  - ✍️ Label: %s\n", p.label))

	if p.paused {
		sb.WriteString(ufmt.Sprintf("  - ⏸️ Paused since: %s\n", p.pausedAt.Format(time.DateTime)))
	}

	if pov != "vault" {
		pvid := prettyVaultID(p.vaultID)
		sb.WriteString(ufmt.Sprintf("  - 🎩 Vault: [%s](/r/demo/payrolls:/v/%s)\n", pvid, pvid))
	}

	if pov != "beneficiary" {
		beneficiaryUser := rusers.GetUserByAddress(p.beneficiary)
		var userStr string
		if beneficiaryUser == nil {
			userStr = p.beneficiary.String()
		} else {
			userStr = "@" + beneficiaryUser.Name
		}
		sb.WriteString(ufmt.Sprintf("  - 👤 Beneficiary: [%s](/r/demo/payrolls:/u/%s)\n", userStr, userStr))
	}

	free := p.freeAt(time.Now())
	vaultAmount := getVaultAmount(p.vaultID)
	claimable := free
	if free > vaultAmount {
		claimable = vaultAmount
	}

	if free > 0 {
		sb.WriteString(renderClaimable("  ", claimable, free, denom, false))
	}

	estimatedDay := p.freeAt(time.Now().Add(time.Hour*24)) - p.freeAt(time.Now())
	sb.WriteString(renderEstimatedMonthly("  ", estimatedDay, denom))

	sb.WriteString("  - 🕙 Remaining: ")
	if p.amount >= 0 {
		amountRemaining := uint64(p.amount) - p.totalWithdrawn
		sb.WriteString(ufmt.Sprintf("%s (%d%s)", ((time.Duration(amountRemaining) * 24) / time.Duration(estimatedDay) / time.Hour).String(), amountRemaining, denom))
	} else {
		sb.WriteRune('∞')
	}
	sb.WriteRune('\n')

	actions := []string{}

	if pov != "vault" && claimable > 0 {
		claim := ufmt.Sprintf("[[Claim]]($help&func=Claim&id=%d&requestedAmount=-1&destination=%s)", uint64(p.id), p.beneficiary.String())
		actions = append(actions, claim)
	}

	if len(actions) > 0 {
		sb.WriteString(ufmt.Sprintf("- ✋ Actions: %s\n", strings.Join(actions, " ")))
	}

	return sb.String()
}

func renderClaimable(prefix string, claimable uint64, free uint64, denom string, plural bool) string {
	if free == 0 {
		return ""
	}
	sb := strings.Builder{}
	sb.WriteString(prefix)
	sb.WriteString("- 💵 Claimable: ")
	if claimable < free {
		pls := ""
		if plural {
			pls = "(s)"
		}
		color := "🟠"
		if claimable == 0 {
			color = "🔴"
		}
		sb.WriteString(ufmt.Sprintf("%s %d%s/%d%s (not enough funds in vault%s!)\n", color, claimable, denom, free, denom, pls))
	} else {
		sb.WriteString(ufmt.Sprintf("🟢 %d%s\n", free, denom))
	}
	return sb.String()
}

func renderEstimatedMonthly(prefix string, estimatedDay uint64, denom string) string {
	return ufmt.Sprintf("%s- 💸 Estimated monthly: %d%s\n", prefix, estimatedDay*30, denom)
}

func prettyVaultID(id string) string {
	parts := strings.Split(id, ".")
	if len(parts) == 0 {
		return id
	}

	user := rusers.GetUserByAddress(std.Address(parts[0]))
	if user == nil {
		return id
	}

	parts[0] = "@" + user.Name
	return strings.Join(parts, ".")
}
