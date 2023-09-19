package dao_proposal_single

import (
	"encoding/binary"
	"std"
	"strings"

	"gno.land/p/demo/teritori/daodao/interfaces"
	"gno.land/p/demo/ujson"
)

// TODO: convert to json

type UpdateSettingsMessage struct {
	dao_interfaces.ExecutableMessage

	Threshold Threshold
}

func (usm *UpdateSettingsMessage) Type() string {
	return "gno.land/p/demo/teritori/daodao/proposal_single.UpdateSettings"
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

func (usm *UpdateSettingsMessage) ToJSON() string {
	return ujson.FormatObject([]ujson.FormatKV{
		{Key: "threshold", Value: usm.Threshold},
	})
}

func (usm *UpdateSettingsMessage) FromJSON(ast *ujson.JSONASTNode) {
	ast.ParseObject([]*ujson.ParseKV{
		{Key: "threshold", CustomParser: func(node *ujson.JSONASTNode) {
			usm.Threshold = ThresholdFromJSON(node)
		}},
	})
}

func NewUpdateSettingsHandler(mod *DAOProposalSingle) dao_interfaces.MessageHandler {
	return &updateSettingsHandler{mod: mod}
}

type updateSettingsHandler struct {
	dao_interfaces.MessageHandler

	mod *DAOProposalSingle
}

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

func (h *updateSettingsHandler) Type() string {
	return UpdateSettingsMessage{}.Type()
}

func (h *updateSettingsHandler) MessageFromJSON(ast *ujson.JSONASTNode) dao_interfaces.ExecutableMessage {
	var usm UpdateSettingsMessage
	ast.ParseAny(&usm)
	return &usm
}
