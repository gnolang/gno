package types

// IBC client events
const (
	AttributeKeyClientID          = "client_id"
	AttributeKeySubjectClientID   = "subject_client_id"
	AttributeKeyClientType        = "client_type"
	AttributeKeyConsensusHeights  = "consensus_heights"
	AttributeKeyUpgradeStore      = "upgrade_store"
	AttributeKeyUpgradePlanHeight = "upgrade_plan_height"
	AttributeKeyUpgradePlanTitle  = "title"
)

// IBC client events vars
var (
	EventTypeCreateClient               = "create_client"
	EventTypeUpdateClient               = "update_client"
	EventTypeUpgradeClient              = "upgrade_client"
	EventTypeSubmitMisbehaviour         = "client_misbehaviour"
	EventTypeRecoverClient              = "recover_client"
	EventTypeScheduleIBCSoftwareUpgrade = "schedule_ibc_software_upgrade"
	EventTypeUpgradeChain               = "upgrade_chain"
)
