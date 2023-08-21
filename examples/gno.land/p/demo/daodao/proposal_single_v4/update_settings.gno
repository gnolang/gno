package dao_proposal_single

import (
	"encoding/binary"
	"strings"

	"gno.land/p/demo/daodao/interfaces_v3"
)

type UpdateSettingsMessage struct {
	dao_interfaces.ExecutableMessage

	Threshold *dao_interfaces.Threshold
}

func (usm *UpdateSettingsMessage) Type() string {
	return "UpdateSettings"
}

func (usm *UpdateSettingsMessage) String() string {
	ss := []string{usm.Type()}
	if usm.Threshold != nil {
		ss = append(ss, "Threshold type: ThresholdQuorum\nThreshold: "+usm.Threshold.ThresholdQuorum.Threshold.String()+"\nQuorum: "+usm.Threshold.ThresholdQuorum.Quorum.String())
	}
	return strings.Join(ss, "\n--\n")
}

func (usm *UpdateSettingsMessage) Binary() []byte {
	b := []byte{}

	t := usm.Type()
	b = binary.BigEndian.AppendUint16(b, uint16(len(t)))
	b = append(b, []byte(t)...)

	if usm.Threshold != nil {
		b = append(b, 1)
		b = binary.BigEndian.AppendUint16(b, uint16(usm.Threshold.ThresholdQuorum.Threshold.Percent))
		b = binary.BigEndian.AppendUint16(b, uint16(usm.Threshold.ThresholdQuorum.Quorum.Percent))
	} else {
		b = append(b, 0)
	}

	return b
}

func UpdateSettingsMessageFromBinary(b []byte) *UpdateSettingsMessage {
	usm := UpdateSettingsMessage{}

	if len(b) < 2 {
		panic("invalid length - less than 2")
	}
	l := binary.BigEndian.Uint16(b[:2])
	b = b[2:]
	if len(b) < int(l) {
		panic("invalid length - less than expected")
	}
	t := string(b[:l])
	if t != usm.Type() {
		panic("invalid type")
	}
	b = b[l:]

	hasThreshold := b[0] == 1
	b = b[1:]
	if hasThreshold {
		if len(b) < 4 {
			panic("invalid length - less than 4")
		}
		threshold := binary.BigEndian.Uint16(b[:2])
		b = b[2:]
		quorum := binary.BigEndian.Uint16(b[:2])
		b = b[2:]

		// TODO: validate threshold and quorum

		pt := dao_interfaces.Percent(threshold)
		pq := dao_interfaces.Percent(quorum)

		usm.Threshold = &dao_interfaces.Threshold{
			ThresholdQuorum: &dao_interfaces.ThresholdQuorum{
				Threshold: dao_interfaces.PercentageThreshold{Percent: &pt},
				Quorum:    dao_interfaces.PercentageThreshold{Percent: &pq},
			},
		}
	}

	return &usm
}

func NewUpdateSettingsHandler(mod *daoProposalSingle) dao_interfaces.MessageHandler {
	return &updateSettingsHandler{mod: mod}
}

type updateSettingsHandler struct {
	dao_interfaces.MessageHandler

	mod *daoProposalSingle
}

func (h *updateSettingsHandler) Execute(message dao_interfaces.ExecutableMessage) {
	usm := message.(*UpdateSettingsMessage)

	if usm.Threshold != nil {
		h.mod.opts.Threshold = *usm.Threshold
	}
}

func (h *updateSettingsHandler) Type() string {
	return UpdateSettingsMessage{}.Type()
}

func (h *updateSettingsHandler) FromBinary(b []byte) dao_interfaces.ExecutableMessage {
	return UpdateSettingsMessageFromBinary(b)
}
