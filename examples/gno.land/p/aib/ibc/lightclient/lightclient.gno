package lightclient

import "gno.land/p/aib/ibc/types"

// Status represents the status of a client
type Status = string

const (
	// Solomachine is used to indicate that the light client is a solo machine.
	Solomachine string = "06-solomachine"

	// Tendermint is used to indicate that the client uses the Tendermint Consensus Algorithm.
	Tendermint string = "07-tendermint"

	// Localhost is the client type for the localhost client.
	Localhost string = "09-localhost"

	// LocalhostClientID is the sentinel client ID for the localhost client.
	LocalhostClientID string = Localhost

	// Active is a status type of a client. An active client is allowed to be used.
	Active Status = "Active"

	// Frozen is a status type of a client. A frozen client is not allowed to be used.
	Frozen Status = "Frozen"

	// Expired is a status type of a client. An expired client is not allowed to be used.
	Expired Status = "Expired"

	// Unknown indicates there was an error in determining the status of a client.
	Unknown Status = "Unknown"

	// Unauthorized indicates that the client type is not registered as an allowed client type.
	Unauthorized Status = "Unauthorized"
)

// Interface is an interface which core IBC uses to interact with light
// client modules.  Light client modules must implement this interface to
// integrate with core IBC.
type Interface interface {
	// Initialize is called upon client creation, it allows the client to perform
	// validation on the client state and initial consensus state.  The light
	// client module is responsible for setting any client-specific data in the
	// store. This includes the client state, initial consensus state and any
	// associated metadata.
	Initialize(ClientState, ConsensusState) error

	// VerifyClientMessage must verify a ClientMessage. A ClientMessage could be
	// a Header, Misbehaviour, or batch update.  It must handle each type of
	// ClientMessage appropriately. Calls to CheckForMisbehaviour, UpdateState,
	// and UpdateStateOnMisbehaviour will assume that the content of the
	// ClientMessage has been verified and can be trusted. An error should be
	// returned if the ClientMessage fails to verify.
	VerifyClientMessage(ClientMessage) error

	// Checks for evidence of a misbehaviour in Header or Misbehaviour type. It
	// assumes the ClientMessage has already been verified.
	CheckForMisbehaviour(ClientMessage) bool

	// UpdateStateOnMisbehaviour should perform appropriate state changes on a
	// client state given that misbehaviour has been detected and verified.
	UpdateStateOnMisbehaviour(ClientMessage)

	// UpdateState updates and stores as necessary any associated information for
	// an IBC client, such as the ClientState and corresponding ConsensusState.
	// Upon successful update, a list of consensus heights is returned. It
	// assumes the ClientMessage has already been verified.
	UpdateState(ClientMessage) []types.Height

	// VerifyMembership is a generic proof verification method which verifies a
	// proof of the existence of a value at a given CommitmentPath at the
	// specified height.  The caller is expected to construct the full
	// CommitmentPath from a CommitmentPrefix and a standardized path (as defined
	// in ICS 24).
	VerifyMembership(height types.Height, delayTimePeriod uint64,
		delayBlockPeriod uint64, proof []byte, path string, value []byte) error

	// VerifyNonMembership is a generic proof verification method which verifies
	// the absence of a given CommitmentPath at a specified height.  The caller
	// is expected to construct the full CommitmentPath from a CommitmentPrefix
	// and a standardized path (as defined in ICS 24).
	VerifyNonMembership(height types.Height, delayTimePeriod uint64,
		delayBlockPeriod uint64, proof []byte, path string) error

	// Status must return the status of the client. Only Active clients are
	// allowed to process packets.
	Status() Status

	// LatestHeight returns the latest height of the client. If no client is
	// present for the provided client identifier a zero value height may be
	// returned.
	LatestHeight() types.Height

	// TimestampAtHeight must return the timestamp for the consensus state
	// associated with the provided height.
	TimestampAtHeight(types.Height) (uint64, error)

	// RecoverClient must verify that the provided substitute may be used to
	// update the subject client.  The light client module must set the updated
	// client and consensus states within the clientStore for the subject client.
	RecoverClient(substituteClientID string) error

	// Upgrade functions NOTE: proof heights are not included as upgrade to a new
	// revision is expected to pass only on the last height committed by the
	// current revision. Clients are responsible for ensuring that the planned
	// last height of the current revision is somehow encoded in the proof
	// verification process.  This is to ensure that no premature upgrades occur,
	// since upgrade plans committed to by the counterparty may be cancelled or
	// modified before the last planned height.  If the upgrade is verified, the
	// upgraded client and consensus states must be set in the client store.
	VerifyUpgradeAndUpdateState(newClient, newConsState, upgradeClientProof,
		upgradeConsensusStateProof []byte) error
}

// ClientState defines the required common functions for light clients.
type ClientState interface {
	ClientType() string
	Validate() error
}

// ConsensusState is the state of the consensus process
type ConsensusState interface {
	ClientType() string
	Validate() error
}

// ClientMessage is an interface used to update an IBC client.
// The update may be done by a single header, a batch of headers, misbehaviour,
// or any type which when verified produces a change to state of the IBC client
type ClientMessage interface {
	ClientType() string
	Validate() error
}
