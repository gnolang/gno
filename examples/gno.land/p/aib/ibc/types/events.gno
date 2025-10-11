package types

// IBC client events
const (
	EventTypeCreateClient               = "create_client"
	EventTypeUpdateClient               = "update_client"
	EventTypeUpgradeClient              = "upgrade_client"
	EventTypeSubmitMisbehaviour         = "client_misbehaviour"
	EventTypeRecoverClient              = "recover_client"
	EventTypeScheduleIBCSoftwareUpgrade = "schedule_ibc_software_upgrade"
	EventTypeUpgradeChain               = "upgrade_chain"

	AttributeKeyClientID          = "client_id"
	AttributeKeySubjectClientID   = "subject_client_id"
	AttributeKeyClientType        = "client_type"
	AttributeKeyConsensusHeights  = "consensus_heights"
	AttributeKeyUpgradeStore      = "upgrade_store"
	AttributeKeyUpgradePlanHeight = "upgrade_plan_height"
	AttributeKeyUpgradePlanTitle  = "title"
)

// IBC core events
const (
	EventTypeSendPacket        = "send_packet"
	EventTypeRecvPacket        = "recv_packet"
	EventTypeTimeoutPacket     = "timeout_packet"
	EventTypeAcknowledgePacket = "acknowledge_packet"
	EventTypeWriteAck          = "write_acknowledgement"

	AttributeKeySrcClient        = "packet_source_client"
	AttributeKeyDstClient        = "packet_dest_client"
	AttributeKeySequence         = "packet_sequence"
	AttributeKeyTimeoutTimestamp = "packet_timeout_timestamp"
	AttributeKeyEncodedPacketHex = "encoded_packet_hex"
	AttributeKeyEncodedAckHex    = "encoded_acknowledgement_hex"
)
