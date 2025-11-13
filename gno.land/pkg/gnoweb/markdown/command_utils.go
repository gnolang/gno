package markdown

import (
	"fmt"
	"io"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
)

// CommandBlockData holds the data needed to render a command block
type CommandBlockData struct {
	FuncName     string
	FuncSig      string
	Params       []vm.NamedType
	PkgPath      string
	ChainId      string
	Remote       string
	SelectedSend string
	Prefix       string // For form vs action differentiation
}

// RenderCommandBlock renders the command block HTML to the given writer
func RenderCommandBlock(w io.Writer, data CommandBlockData) error {
	// Use form-specific prefix for data attributes to avoid conflicts
	prefix := "action-" + data.Prefix

	fmt.Fprintf(w, `<div class="b-code">
    <button data-controller="copy" data-action="click->copy#copy"
      data-copy-remote-value="%s-%s" data-copy-clean-value class="btn-copy"
      aria-label="Copy Command">
      <svg class="c-icon">
	      <use href="#ico-copy" data-copy-target="icon"></use>
	      <use href="#ico-check" class="u-hidden u-color-valid" data-copy-target="icon"></use>
      </svg>
    </button>
    <pre><code><span data-%s-target="mode" data-%s-mode-value="fast" class="u-hidden" data-copy-target="%s-%s"># WARNING: This command is running in an INSECURE mode.
# It is strongly recommended to use a hardware device for signing
# and avoid trusting any computer connected to the internet,
# as your private keys could be exposed.

gnokey maketx call -pkgpath "%s" -func "%s"`,
		prefix, HTMLEscapeString(data.FuncName),
		prefix, prefix, prefix, HTMLEscapeString(data.FuncName),
		HTMLEscapeString(data.PkgPath),
		HTMLEscapeString(data.FuncName))

	// Add parameters for fast mode
	for _, param := range data.Params {
		fmt.Fprintf(w, ` -args "<span data-%s-target="arg" data-%s-arg-value="%s"></span>"`,
			prefix, prefix, HTMLEscapeString(param.Name))
	}

	fmt.Fprintf(w, ` -gas-fee 1000000ugnot -gas-wanted 5000000 -send "<span data-%s-target="send-code"></span>" -broadcast -chainid "%s" -remote "%s" <span data-%s-target="address">ADDRESS</span></span><span data-%s-target="mode" data-%s-mode-value="secure" data-copy-target="%s-%s" class="u-inline">gnokey query -remote "%s" auth/accounts/<span data-%s-target="address">ADDRESS</span>
gnokey maketx call -pkgpath "%s" -func "%s"`,
		prefix, HTMLEscapeString(data.ChainId), HTMLEscapeString(data.Remote), prefix,
		prefix, prefix, prefix, HTMLEscapeString(data.FuncName),
		HTMLEscapeString(data.Remote), prefix,
		HTMLEscapeString(data.PkgPath),
		HTMLEscapeString(data.FuncName))

	// Add parameters for secure mode
	for _, param := range data.Params {
		fmt.Fprintf(w, ` -args "<span data-%s-target="arg" data-%s-arg-value="%s"></span>"`,
			prefix, prefix, HTMLEscapeString(param.Name))
	}

	fmt.Fprintf(w, ` -gas-fee 1000000ugnot -gas-wanted 5000000 -send "<span data-%s-target="send-code"></span>" <span data-%s-target="address">ADDRESS</span> > call.tx
gnokey sign -tx-path call.tx -chainid "%s" -account-number ACCOUNTNUMBER -account-sequence SEQUENCENUMBER <span data-%s-target="address">ADDRESS</span>
gnokey broadcast -remote "%s" call.tx</span></code></pre>
  </div>
`, prefix, prefix, HTMLEscapeString(data.ChainId), prefix, HTMLEscapeString(data.Remote))

	return nil
}
