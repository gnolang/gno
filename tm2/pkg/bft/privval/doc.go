/*
* PrivValidator
*
* Package privval implements the BFT validator interface defined in tm2/pkg/bft/types.PrivValidator.
* The validator role is to sign votes and proposals for the consensus protocol, ensuring that it never
* double-signs, even in the case of a crash during the signing process or a malicious attack.
*
* To achieve this, the PrivValidator relies on two components:
* - a signer that generates cryptographic signatures for arbitrary byte slices without any checks.
* - a state that both stores and verifies the last signature and signed data to prevent double-signing.
*
*
* Signer
*
* The signer implements the BFT signer interface defined in tm2/pkg/bft/types.Signer. Two implementations
* are provided in this package:
* - a local signer that uses a keypair encoded with amino and persisted to disk (default for gnoland nodes).
* - a remote signer that uses a client sending signing requests to a remote signer server.
*
* Both the remote signer client and server are provided in tm2/pkg/bft/privval/signer/remote. The current
* implementation supports TCP and UNIX socket connections.
*
* TCP connections are secured using the cryptographic handshake defined in tm2/pkg/p2p/conn.MakeSecretConnection
* which is an implementation of the STS protocol described in this whitepaper:
* https://github.com/tendermint/tendermint/blob/0.1/docs/sts-final.pdf
* TCP connections can optionally be mutually authenticated using a whitelist of authorized public keys for both
* the client and the server.
*
* By default, the remote signer client will indefinitely try to connect to the remote signer server for each
* request it sends. Consequently, a node using a private validator with a remote signer will not fail due to
* temporary network issues or a crash of the remote signer server.
*
* The remote signer server provided by this package is a generic bridge that take any types.Signer as a
* parameter and proxies the client requests to it. Additionally, gnokms is a CLI tool available in
* contribs/gnokms that aims to provide a remote signer server along with a set of backend signers, including
* one based on gnokey.
*
*
* State
*
* The state manager defined in tm2/pkg/bft/privval/state does not implement any interface. It basically keeps
* track of the last signature and signed data to prevent double-signing. The state is persisted to disk in a
* file encoded with amino and all checks are performed locally.
 */
package privval
