package multisig

import (
	"chain/banker"
	"strconv"
	"strings"

	"gno.land/p/nt/markdown/sanitize/v0"
	msig "gno.land/p/nt/multisig/v0"

	"gno.land/r/nt/grc20reg/v0"
)

// Render serves the factory, one safe, or one proposal.
// Paths: "", "<safeID>", "<safeID>/<propID>".
func Render(path string) string {
	path = strings.Trim(path, "/")
	if path == "" {
		return renderHome()
	}
	parts := strings.Split(path, "/")
	safeID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || safeID <= 0 || safeID > int64(len(safes)) || safes[int(safeID-1)] == nil {
		return "# Multisig treasury\n\nUnknown safe.\n"
	}
	if len(parts) == 1 {
		return renderSafe(safeID)
	}
	if len(parts) == 2 {
		propID, perr := strconv.ParseInt(parts[1], 10, 64)
		if perr != nil {
			return "# Multisig treasury\n\nUnknown proposal.\n"
		}
		return renderProposal(safeID, propID)
	}
	return "# Multisig treasury\n\nUnknown page.\n"
}

func renderHome() string {
	b := new(strings.Builder)
	b.WriteString("# Multisig treasury\n\n")
	b.WriteString("Anyone can create an M-of-N safe. Funds live at that safe's address, not at the factory.\n\n")
	b.WriteString("Safes: " + strconv.Itoa(len(safes)) + "\n\n")
	start := 0
	if len(safes) > 20 {
		start = len(safes) - 20
		b.WriteString("Showing the latest 20.\n\n")
	}
	for i := start; i < len(safes); i++ {
		s := safes[i]
		id := strconv.FormatInt(s.ID(), 10)
		b.WriteString("- [" + sanitize.InlineText(s.Name()) + "](" + rel() + ":" + id + ")")
		b.WriteString(" — " + strconv.Itoa(s.Threshold()) + " of " + strconv.Itoa(s.SignerCount()))
		b.WriteString(" — " + sanitize.InlineCode(s.Address().String()) + "\n")
	}
	return b.String()
}

func renderSafe(id int64) string {
	s := safes[int(id-1)]
	b := new(strings.Builder)
	b.WriteString("# " + sanitize.InlineText(s.Name()) + "\n\n")
	b.WriteString("Address: " + sanitize.InlineCode(s.Address().String()) + "\n\n")
	b.WriteString("Threshold: " + strconv.Itoa(s.Threshold()) + " of " + strconv.Itoa(s.SignerCount()) + "\n\n")
	b.WriteString("Timelock for new proposals: " + strconv.FormatInt(s.Delay(), 10) + " seconds\n\n")
	if s.Paused() {
		b.WriteString("Status: paused. Coin, GRC20, and call proposals will not execute.\n\n")
	} else {
		b.WriteString("Status: active\n\n")
	}
	b.WriteString("Executions: " + strconv.FormatInt(s.Nonce(), 10) + "\n\n")
	b.WriteString("## Signers\n\n")
	for i := 0; i < s.SignerCount(); i++ {
		b.WriteString("- " + sanitize.InlineCode(s.SignerAt(i).String()) + "\n")
	}
	b.WriteString("\n## Balances\n\n")
	b.WriteString("- ugnot: " + strconv.FormatInt(bankerGet(s.Address(), "ugnot"), 10) + "\n")
	for _, denom := range s.WatchedDenoms() {
		if denom == "ugnot" || !msig.ValidDenom(denom) {
			continue
		}
		b.WriteString("- " + sanitize.InlineCode(denom) + ": " + strconv.FormatInt(bankerGet(s.Address(), denom), 10) + "\n")
	}
	for _, key := range s.WatchedTokens() {
		tok := grc20reg.Get(key)
		if tok == nil {
			b.WriteString("- " + sanitize.InlineCode(key) + ": not in the registry\n")
			continue
		}
		b.WriteString("- " + sanitize.InlineText(tok.GetSymbol()) + " (" + sanitize.InlineCode(key) + "): " + strconv.FormatInt(tok.BalanceOf(s.Address()), 10) + "\n")
	}
	b.WriteString("\nSend native coins to the address above. Do not attach them to a call on the factory unless you use Deposit.\n\n")
	b.WriteString("## Proposals\n\n")
	views := s.Recent(20)
	if len(views) == 0 {
		b.WriteString("None yet.\n")
		return b.String()
	}
	if s.ProposalCount() > len(views) {
		b.WriteString("Showing the latest " + strconv.Itoa(len(views)) + ".\n\n")
	}
	now := nowUnix()
	for _, view := range views {
		st := msig.DisplayStatus(view.Status, view.ExpiresAt, now)
		pid := strconv.FormatInt(view.ID, 10)
		b.WriteString("- [" + pid + " " + view.Kind.String() + "](" + rel() + ":" + strconv.FormatInt(id, 10) + "/" + pid + ") — " + st.String() + "\n")
	}
	return b.String()
}

func renderProposal(safeID, propID int64) string {
	s := safes[int(safeID-1)]
	view, err := s.View(propID)
	if err != nil {
		return "# Multisig treasury\n\nUnknown proposal.\n"
	}
	now := nowUnix()
	st := msig.DisplayStatus(view.Status, view.ExpiresAt, now)
	b := new(strings.Builder)
	b.WriteString("# Proposal " + strconv.FormatInt(view.ID, 10) + "\n\n")
	b.WriteString("Safe: [" + sanitize.InlineText(s.Name()) + "](" + rel() + ":" + strconv.FormatInt(safeID, 10) + ")\n\n")
	b.WriteString("Kind: " + view.Kind.String() + "\n\n")
	b.WriteString("Status: " + st.String() + "\n\n")
	b.WriteString("Proposer: " + sanitize.InlineCode(view.Proposer.String()) + "\n\n")
	if view.Memo != "" {
		b.WriteString("Memo: " + sanitize.InlineText(view.Memo) + "\n\n")
	}
	b.WriteString("Created: " + strconv.FormatInt(view.CreatedAt, 10) + "\n\n")
	b.WriteString("Expires: " + strconv.FormatInt(view.ExpiresAt, 10) + "\n\n")
	b.WriteString("Timelock: " + strconv.FormatInt(view.Delay, 10) + " seconds")
	if view.Armed {
		b.WriteString(", armed at " + strconv.FormatInt(view.ArmedAt, 10))
	}
	b.WriteString("\n\n")
	writeAction(b, view)
	n, _ := s.ApprovalCount(propID)
	b.WriteString("## Approvals (" + strconv.Itoa(n) + " / " + strconv.Itoa(s.Threshold()) + ")\n\n")
	if len(view.Approvals) == 0 {
		b.WriteString("None.\n\n")
	}
	for _, a := range view.Approvals {
		mark := "former signer"
		if s.IsSigner(a) {
			mark = "signer"
		}
		b.WriteString("- " + sanitize.InlineCode(a.String()) + " (" + mark + ")\n")
	}
	b.WriteString("\n")
	sid := strconv.FormatInt(safeID, 10)
	pid := strconv.FormatInt(propID, 10)
	b.WriteString(help("Approve", sid, pid) + " · " + help("Revoke", sid, pid) + " · " + help("Execute", sid, pid) + " · " + help("Cancel", sid, pid) + "\n")
	return b.String()
}

func writeAction(b *strings.Builder, view msig.ProposalView) {
	switch view.Kind {
	case msig.KindCoin:
		b.WriteString("Send " + strconv.FormatInt(view.Amount, 10) + " " + sanitize.InlineCode(view.Denom))
		b.WriteString(" to " + sanitize.InlineCode(view.To.String()) + "\n\n")
	case msig.KindGRC20:
		b.WriteString("Send " + strconv.FormatInt(view.Amount, 10) + " of " + sanitize.InlineCode(view.TokenKey))
		b.WriteString(" to " + sanitize.InlineCode(view.To.String()) + "\n\n")
	case msig.KindCall:
		b.WriteString("Authorize " + sanitize.InlineCode(view.TargetPkg))
		b.WriteString(" to pull payload: " + sanitize.InlineText(view.Payload) + "\n\n")
		if view.Consumed {
			b.WriteString("Payload already consumed.\n\n")
		}
	case msig.KindAddSigner:
		b.WriteString("Add signer " + sanitize.InlineCode(view.Signer.String()) + "\n\n")
	case msig.KindRemoveSigner:
		b.WriteString("Remove signer " + sanitize.InlineCode(view.Signer.String()) + "\n\n")
	case msig.KindThreshold:
		b.WriteString("Set threshold to " + strconv.Itoa(view.Threshold) + "\n\n")
	case msig.KindDelay:
		b.WriteString("Set timelock to " + strconv.FormatInt(view.NewDelay, 10) + " seconds\n\n")
	case msig.KindPause:
		b.WriteString("Pause spends.\n\n")
	case msig.KindUnpause:
		b.WriteString("Unpause spends.\n\n")
	}
}

func help(fn, safeID, propID string) string {
	return "[" + fn + "](" + rel() + "$help&func=" + fn + "&safeID=" + safeID + "&propID=" + propID + ")"
}

func rel() string {
	return strings.TrimPrefix(realmPath, "gno.land")
}

func bankerGet(addr address, denom string) int64 {
	return banker.NewReadonlyBanker().GetCoin(addr, denom)
}
