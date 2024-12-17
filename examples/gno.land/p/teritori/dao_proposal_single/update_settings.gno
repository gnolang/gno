package dao_proposal_single

import (
	"strings"

	"gno.land/p/demo/json"
	"gno.land/p/teritori/dao_interfaces"
)

type UpdateSettingsMessage struct {
	dao_interfaces.ExecutableMessage

	Threshold Threshold
}

var _ dao_interfaces.ExecutableMessage = (*UpdateSettingsMessage)(nil)

func (usm UpdateSettingsMessage) Type() string {
	return "gno.land/p/teritori/dao_proposal_single.UpdateSettings"
}

func (usm *UpdateSettingsMessage) String() string {
	ss := []string{usm.Type()}
	switch usm.Threshold.(type) {
	case *ThresholdThresholdQuorum:
		ss = append(ss, "Threshold type: ThresholdQuorum\n\nThreshold: "+usm.Threshold.(*ThresholdThresholdQuorum).Threshold.String()+"\n\nQuorum: "+usm.Threshold.(*ThresholdThresholdQuorum).Quorum.String())

	default:
		ss = append(ss, "Threshold type: unknown")

	}

	return strings.Join(ss, "\n--\n")
}

func (usm *UpdateSettingsMessage) ToJSON() *json.Node {
	return json.ObjectNode("", map[string]*json.Node{
		"threshold": usm.Threshold.ToJSON(),
	})
}

func (usm *UpdateSettingsMessage) FromJSON(ast *json.Node) {
	obj := ast.MustObject()
	usm.Threshold = ThresholdFromJSON(obj["threshold"])
}

func NewUpdateSettingsHandler(mod *DAOProposalSingle) dao_interfaces.MessageHandler {
	return &updateSettingsHandler{mod: mod}
}

type updateSettingsHandler struct {
	dao_interfaces.MessageHandler

	mod *DAOProposalSingle
}

var _ dao_interfaces.MessageHandler = (*updateSettingsHandler)(nil)

func (h *updateSettingsHandler) Execute(message dao_interfaces.ExecutableMessage) {
	usm := message.(*UpdateSettingsMessage)

	switch usm.Threshold.(type) {
	case *ThresholdThresholdQuorum:
		// FIXME: validate better
		h.mod.opts.Threshold = usm.Threshold.(*ThresholdThresholdQuorum)
		return

	default:
		panic("unsupported threshold type")
	}
}

func (h updateSettingsHandler) Type() string {
	return UpdateSettingsMessage{}.Type()
}

func (h *updateSettingsHandler) Instantiate() dao_interfaces.ExecutableMessage {
	return &UpdateSettingsMessage{}
}
