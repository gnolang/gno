/*
* PrivValidator
*
* Package privval implements the BFT validator interface defined in tm2/pkg/bft/types.PrivValidator.
* The BFT validator role is to sign votes and proposals for the consensus protocol and to ensure that
* it never double signs, even in the case of a crash during the signing process or a malicious attack.
*
* To achieve this, the PrivValidator relies on two components:
* - a signer which generates cryptographic signatures for arbitrary byte slice without any checks.
* - a state which both stores and checks the last signature and signed data to prevent double signing.
*
*
* Signer:
*
* The signer implements the BFT signer interface defined in tm2/pkg/bft/types.Signer. Two implementations
* are provided in this package:
* - a local signer which uses a keypair encoded with amino and persisted to disk (default for Gnoland nodes).
* - a remote signer which uses a client that sends signing requests to a remote signer server.
*
* Both the remote signer client and server are provided in tm2/pkg/bft/privval/signer/remote. The current
* implementation of these remote signer client/server supports TCP and UNIX socket connections.
*
* TCP connections are secured using the cryptographic handshake defined in tm2/pkg/p2p/conn.MakeSecretConnection
* which is an implementation of the STS protocol described in this whitepaper:
* https://github.com/tendermint/tendermint/blob/0.1/docs/sts-final.pdf
* TCP connections can optionally be mutually authenticated using a whitelist of allowed public keys for both the
* client and the server.
*
* By default, the remote signer client will indefinitely try to connect to the remote signer server for each
* request it sends. A node using a private validator with a remote signer will therefore not fail in case of
* a temporary network failure or a crash of the remote signer server.
*
* The remote signer server provided by this package is just a generic bridge that take any types.Signer as
* parameter and proxy the client requests to it. gnokms is a cli tool available in contribs/gnokms that aims to
* provide a remote signer server along with a set of backend signers among which a gnokey based signer.
*
*
* State:
*
* The state manager defined in tm2/pkg/bft/privval/state does not implement any interface. It basically keep
* track of the last signature and signed data to prevent double signing. The state is persisted to disk in a
* file encoded with amino and all the checks are done locally.
 */
package privval
