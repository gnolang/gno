package upstream

// secret_connection.go: SecretConnection for the tmkms-listener path.
//
// **Direct port of cometbft/p2p/conn/secret_connection.go (CometBFT
// v0.34.34) with tm2 type substitutions.** This is the v0.34
// Merlin-transcript-bound STS handshake that tmkms / tendermint-rs
// speak. tm2's chain-internal SecretConnection in
// tm2/pkg/p2p/conn/secret_connection.go is a pre-Merlin variant that
// is wire-incompatible with tmkms — see Phase 6's
// secret_connection_compat_test.go for the byte-level diff.
//
// We keep tm2's variant unchanged (changing it is a chain wire break)
// and use this adapted upstream-compat copy ONLY on the tmkms
// listener path (socket_listener.go's TCP Accept).
//
// Adaptation log (cometbft v0.34.34 → tm2):
//   - package conn                 → package upstream
//   - libs/sync.Mutex              → sync.Mutex
//   - libs/async                   → tm2/pkg/async (note: Task return order
//                                    is (val, err, abort) here, not the
//                                    cometbft (val, abort, err))
//   - libs/protoio                 → tm2/pkg/bft/privval/upstream.protoio
//   - crypto.{PubKey,PrivKey}      → tm2/pkg/crypto.{PubKey,PrivKey}
//   - crypto/ed25519.{PubKey,PrivKey} → ed25519.{PubKeyEd25519,PrivKeyEd25519}
//   - crypto/encoding.PubKeyToProto/FromProto → upstream.PubKeyToProto/FromProto
//     (existing helpers in translator_pb.go)
//   - PubKey.VerifySignature       → PubKey.VerifyBytes
//     (the only actual method-name drift between cometbft and tm2)
//   - gogo/protobuf/types.BytesValue → wrapperspb.BytesValue
//     (identical wire bytes, different Go type)
//   - tendermint.p2p.AuthSigMessage → upstreampb.AuthSigMessage
//     (added to upstream.proto for this path)
//
// Constants below intentionally duplicate the values in
// tm2/pkg/p2p/conn — local copies keep this package self-contained
// and audit-friendly against cometbft.

import (
	"bytes"
	"crypto/cipher"
	crand "crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"sync"
	"time"

	"github.com/gtank/merlin"
	pool "github.com/libp2p/go-buffer-pool"
	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/curve25519"
	"golang.org/x/crypto/hkdf"
	"golang.org/x/crypto/nacl/box"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/gnolang/gno/tm2/pkg/async"
	"github.com/gnolang/gno/tm2/pkg/bft/privval/upstream/upstreampb"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
)

// 4 + 1024 == 1028 total frame size
const (
	scDataLenSize      = 4
	scDataMaxSize      = 1024
	scTotalFrameSize   = scDataMaxSize + scDataLenSize
	scAEADSizeOverhead = 16 // poly1305 auth tag
	scAEADKeySize      = chacha20poly1305.KeySize
	scAEADNonceSize    = chacha20poly1305.NonceSize
)

var (
	// ErrSmallOrderRemotePubKey is returned when the remote ephemeral
	// pubkey is one of the well-known small-order points (which would
	// collapse the DH shared secret to a known value).
	ErrSmallOrderRemotePubKey = errors.New("upstream: detected low order point from remote peer")

	labelEphemeralLowerPublicKey = []byte("EPHEMERAL_LOWER_PUBLIC_KEY")
	labelEphemeralUpperPublicKey = []byte("EPHEMERAL_UPPER_PUBLIC_KEY")
	labelDHSecret                = []byte("DH_SECRET")
	labelSecretConnectionMac     = []byte("SECRET_CONNECTION_MAC")

	secretConnKeyAndChallengeGen = []byte("TENDERMINT_SECRET_CONNECTION_KEY_AND_CHALLENGE_GEN")
)

// SecretConnection is a net.Conn that performs the upstream Tendermint
// v0.34 SecretConnection STS handshake (Merlin-bound) and then frames
// payloads with ChaCha20-Poly1305.
type SecretConnection struct {
	// immutable
	recvAead cipher.AEAD
	sendAead cipher.AEAD

	remPubKey crypto.PubKey
	conn      io.ReadWriteCloser

	// net.Conn must be safe for concurrent Read and Write — we hold
	// independent mutexes for each direction.
	recvMtx    sync.Mutex
	recvBuffer []byte
	recvNonce  *[scAEADNonceSize]byte

	sendMtx   sync.Mutex
	sendNonce *[scAEADNonceSize]byte
}

// MakeSecretConnection performs the upstream Tendermint v0.34
// SecretConnection handshake over conn, identifying ourselves with
// locPrivKey. Returns an error on any handshake failure; the caller
// is responsible for closing conn in that case.
func MakeSecretConnection(conn io.ReadWriteCloser, locPrivKey crypto.PrivKey) (*SecretConnection, error) {
	locPubKey := locPrivKey.PubKey()

	// Generate ephemeral X25519 keys for perfect forward secrecy.
	locEphPub, locEphPriv := genEphKeys()

	// Exchange ephemeral pubkeys.
	remEphPub, err := shareEphPubKey(conn, locEphPub)
	if err != nil {
		return nil, err
	}

	// Sort lexically; lower/upper labels go into the Merlin transcript
	// in a fixed order so both sides derive the same MAC regardless of
	// who connected.
	loEphPub, hiEphPub := sort32(locEphPub, remEphPub)

	transcript := merlin.NewTranscript("TENDERMINT_SECRET_CONNECTION_TRANSCRIPT_HASH")
	transcript.AppendMessage(labelEphemeralLowerPublicKey, loEphPub[:])
	transcript.AppendMessage(labelEphemeralUpperPublicKey, hiEphPub[:])

	locIsLeast := bytes.Equal(locEphPub[:], loEphPub[:])

	dhSecret, err := computeDHSecret(remEphPub, locEphPriv)
	if err != nil {
		return nil, err
	}

	transcript.AppendMessage(labelDHSecret, dhSecret[:])

	// Derive recv/send keys from dhSecret via HKDF-SHA256 (same as
	// the pre-Merlin handshake). The Merlin transcript is used only
	// for the MAC challenge below.
	recvSecret, sendSecret := deriveSecrets(dhSecret, locIsLeast)

	const challengeSize = 32
	var challenge [challengeSize]byte
	challengeSlice := transcript.ExtractBytes(labelSecretConnectionMac, challengeSize)
	copy(challenge[:], challengeSlice[0:challengeSize])

	sendAead, err := chacha20poly1305.New(sendSecret[:])
	if err != nil {
		return nil, errors.New("upstream: invalid send SecretConnection key")
	}
	recvAead, err := chacha20poly1305.New(recvSecret[:])
	if err != nil {
		return nil, errors.New("upstream: invalid receive SecretConnection key")
	}

	sc := &SecretConnection{
		conn:       conn,
		recvBuffer: nil,
		recvNonce:  new([scAEADNonceSize]byte),
		sendNonce:  new([scAEADNonceSize]byte),
		recvAead:   recvAead,
		sendAead:   sendAead,
	}

	// Sign the Merlin-derived challenge — proves possession of the
	// consensus key bound to *this* handshake (not a replay).
	locSignature, err := locPrivKey.Sign(challenge[:])
	if err != nil {
		return nil, fmt.Errorf("upstream: sign challenge: %w", err)
	}

	authSigMsg, err := shareAuthSignature(sc, locPubKey, locSignature)
	if err != nil {
		return nil, err
	}

	remPubKey, remSignature := authSigMsg.Key, authSigMsg.Sig
	if _, ok := remPubKey.(ed25519.PubKeyEd25519); !ok {
		return nil, fmt.Errorf("upstream: expected ed25519 pubkey, got %T", remPubKey)
	}
	if !remPubKey.VerifyBytes(challenge[:], remSignature) {
		return nil, errors.New("upstream: challenge verification failed")
	}

	sc.remPubKey = remPubKey
	return sc, nil
}

// RemotePubKey returns the authenticated remote pubkey.
func (sc *SecretConnection) RemotePubKey() crypto.PubKey {
	return sc.remPubKey
}

// Write encrypts and frames data, calling the underlying conn's Write
// per ≤dataMaxSize chunk. Atomic for chunks that fit in a single frame.
func (sc *SecretConnection) Write(data []byte) (n int, err error) {
	sc.sendMtx.Lock()
	defer sc.sendMtx.Unlock()

	for 0 < len(data) {
		if err := func() error {
			sealedFrame := pool.Get(scAEADSizeOverhead + scTotalFrameSize)
			frame := pool.Get(scTotalFrameSize)
			defer func() {
				pool.Put(sealedFrame)
				pool.Put(frame)
			}()
			var chunk []byte
			if scDataMaxSize < len(data) {
				chunk = data[:scDataMaxSize]
				data = data[scDataMaxSize:]
			} else {
				chunk = data
				data = nil
			}
			chunkLength := len(chunk)
			binary.LittleEndian.PutUint32(frame, uint32(chunkLength))
			copy(frame[scDataLenSize:], chunk)

			sc.sendAead.Seal(sealedFrame[:0], sc.sendNonce[:], frame, nil)
			incrSecConnNonce(sc.sendNonce)

			_, err = sc.conn.Write(sealedFrame)
			if err != nil {
				return err
			}
			n += len(chunk)
			return nil
		}(); err != nil {
			return n, err
		}
	}
	return n, err
}

// Read pulls one sealed frame off the conn, decrypts it, and copies
// up to len(data) bytes of payload into data. Excess payload is held
// in recvBuffer for the next Read call.
func (sc *SecretConnection) Read(data []byte) (n int, err error) {
	sc.recvMtx.Lock()
	defer sc.recvMtx.Unlock()

	if 0 < len(sc.recvBuffer) {
		n = copy(data, sc.recvBuffer)
		sc.recvBuffer = sc.recvBuffer[n:]
		return
	}

	sealedFrame := pool.Get(scAEADSizeOverhead + scTotalFrameSize)
	defer pool.Put(sealedFrame)
	_, err = io.ReadFull(sc.conn, sealedFrame)
	if err != nil {
		return
	}

	frame := pool.Get(scTotalFrameSize)
	defer pool.Put(frame)
	_, err = sc.recvAead.Open(frame[:0], sc.recvNonce[:], sealedFrame, nil)
	if err != nil {
		return n, fmt.Errorf("upstream: failed to decrypt SecretConnection: %w", err)
	}
	incrSecConnNonce(sc.recvNonce)

	chunkLength := binary.LittleEndian.Uint32(frame)
	if chunkLength > scDataMaxSize {
		return 0, errors.New("upstream: chunkLength exceeds dataMaxSize")
	}
	chunk := frame[scDataLenSize : scDataLenSize+chunkLength]
	n = copy(data, chunk)
	if n < len(chunk) {
		sc.recvBuffer = make([]byte, len(chunk)-n)
		copy(sc.recvBuffer, chunk[n:])
	}
	return n, err
}

// net.Conn passthroughs — same semantics as the underlying conn.
func (sc *SecretConnection) Close() error                  { return sc.conn.Close() }
func (sc *SecretConnection) LocalAddr() net.Addr           { return sc.conn.(net.Conn).LocalAddr() }
func (sc *SecretConnection) RemoteAddr() net.Addr          { return sc.conn.(net.Conn).RemoteAddr() }
func (sc *SecretConnection) SetDeadline(t time.Time) error { return sc.conn.(net.Conn).SetDeadline(t) }
func (sc *SecretConnection) SetReadDeadline(t time.Time) error {
	return sc.conn.(net.Conn).SetReadDeadline(t)
}

func (sc *SecretConnection) SetWriteDeadline(t time.Time) error {
	return sc.conn.(net.Conn).SetWriteDeadline(t)
}

// genEphKeys generates a fresh X25519 keypair via nacl/box (same
// primitive cometbft uses). The private scalar is NOT clamped — the
// upstream Rust impl using x25519-dalek clamps; this divergence has
// been documented as harmless in cometbft for years.
func genEphKeys() (ephPub, ephPriv *[32]byte) {
	var err error
	ephPub, ephPriv, err = box.GenerateKey(crand.Reader)
	if err != nil {
		panic("upstream: could not generate ephemeral key-pair")
	}
	return
}

// shareEphPubKey writes our ephemeral pubkey and reads the peer's, in
// parallel. Each side wraps the 32-byte pubkey in a gogoproto
// BytesValue and uses a varint-length-prefixed framing — wire
// equivalent to wrapperspb.BytesValue used here.
func shareEphPubKey(conn io.ReadWriteCloser, locEphPub *[32]byte) (remEphPub *[32]byte, err error) {
	trs, _ := async.Parallel(
		func(_ int) (val any, err error, abort bool) {
			lc := *locEphPub
			_, err1 := NewDelimitedWriter(conn).WriteMsg(&wrapperspb.BytesValue{Value: lc[:]})
			if err1 != nil {
				return nil, err1, true
			}
			return nil, nil, false
		},
		func(_ int) (val any, err error, abort bool) {
			var bv wrapperspb.BytesValue
			_, err2 := NewDelimitedReader(conn, 1024*1024).ReadMsg(&bv)
			if err2 != nil {
				return nil, err2, true
			}
			if len(bv.Value) != 32 {
				return nil, fmt.Errorf("upstream: remote ephemeral pubkey size %d, want 32", len(bv.Value)), true
			}
			var rep [32]byte
			copy(rep[:], bv.Value)
			if hasSmallOrder(rep) {
				return nil, ErrSmallOrderRemotePubKey, true
			}
			return rep, nil, false
		},
	)

	if trs.FirstError() != nil {
		err = trs.FirstError()
		return
	}
	rep := trs.FirstValue().([32]byte)
	return &rep, nil
}

// blacklist of small-order Curve25519 points (libsodium's set). Same
// as upstream Tendermint's; identical bytes.
var blacklist = [][32]byte{
	// 0 (order 4)
	{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
	// 1 (order 1)
	{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
	// 325606250916557431795983626356110631294008115727848805560023387167927233504 (order 8)
	{0xe0, 0xeb, 0x7a, 0x7c, 0x3b, 0x41, 0xb8, 0xae, 0x16, 0x56, 0xe3,
		0xfa, 0xf1, 0x9f, 0xc4, 0x6a, 0xda, 0x09, 0x8d, 0xeb, 0x9c, 0x32,
		0xb1, 0xfd, 0x86, 0x62, 0x05, 0x16, 0x5f, 0x49, 0xb8, 0x00},
	// 39382357235489614581723060781553021112529911719440698176882885853963445705823 (order 8)
	{0x5f, 0x9c, 0x95, 0xbc, 0xa3, 0x50, 0x8c, 0x24, 0xb1, 0xd0, 0xb1,
		0x55, 0x9c, 0x83, 0xef, 0x5b, 0x04, 0x44, 0x5c, 0xc4, 0x58, 0x1c,
		0x8e, 0x86, 0xd8, 0x22, 0x4e, 0xdd, 0xd0, 0x9f, 0x11, 0x57},
	// p-1 (order 2)
	{0xec, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x7f},
	// p (=0, order 4)
	{0xed, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x7f},
	// p+1 (=1, order 1)
	{0xee, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x7f},
}

func hasSmallOrder(pubKey [32]byte) bool {
	for _, bl := range blacklist {
		if bytes.Equal(pubKey[:], bl[:]) {
			return true
		}
	}
	return false
}

func deriveSecrets(dhSecret *[32]byte, locIsLeast bool) (recvSecret, sendSecret *[scAEADKeySize]byte) {
	hash := sha256.New
	h := hkdf.New(hash, dhSecret[:], nil, secretConnKeyAndChallengeGen)
	res := new([2 * scAEADKeySize]byte)
	if _, err := io.ReadFull(h, res[:]); err != nil {
		panic(fmt.Errorf("upstream: hkdf: %w", err))
	}

	recvSecret = new([scAEADKeySize]byte)
	sendSecret = new([scAEADKeySize]byte)

	// Bytes 0..keysize and keysize..2*keysize are the two AEAD keys;
	// which is recv vs send depends on lex order of ephemeral pubkeys.
	if locIsLeast {
		copy(recvSecret[:], res[0:scAEADKeySize])
		copy(sendSecret[:], res[scAEADKeySize:scAEADKeySize*2])
	} else {
		copy(sendSecret[:], res[0:scAEADKeySize])
		copy(recvSecret[:], res[scAEADKeySize:scAEADKeySize*2])
	}
	return
}

// computeDHSecret derives the X25519 shared secret. Returns an error
// for the trivial all-zero output case (which curve25519.X25519
// already rejects internally).
func computeDHSecret(remPubKey, locPrivKey *[32]byte) (*[32]byte, error) {
	shr, err := curve25519.X25519(locPrivKey[:], remPubKey[:])
	if err != nil {
		return nil, err
	}
	var out [32]byte
	copy(out[:], shr)
	return &out, nil
}

func sort32(foo, bar *[32]byte) (lo, hi *[32]byte) {
	if bytes.Compare(foo[:], bar[:]) < 0 {
		return foo, bar
	}
	return bar, foo
}

type authSigMessage struct {
	Key crypto.PubKey
	Sig []byte
}

// shareAuthSignature exchanges AuthSigMessage in parallel over the now-
// AEAD-protected conn. Each side packages its consensus pubkey + the
// challenge signature as a length-delimited upstreampb.AuthSigMessage.
func shareAuthSignature(sc io.ReadWriter, pubKey crypto.PubKey, signature []byte) (recvMsg authSigMessage, err error) {
	trs, _ := async.Parallel(
		func(_ int) (val any, err error, abort bool) {
			pbpk, perr := PubKeyToProto(pubKey)
			if perr != nil {
				return nil, perr, true
			}
			_, err1 := NewDelimitedWriter(sc).WriteMsg(&upstreampb.AuthSigMessage{PubKey: pbpk, Sig: signature})
			if err1 != nil {
				return nil, err1, true
			}
			return nil, nil, false
		},
		func(_ int) (val any, err error, abort bool) {
			var pba upstreampb.AuthSigMessage
			_, err2 := NewDelimitedReader(sc, 1024*1024).ReadMsg(&pba)
			if err2 != nil {
				return nil, err2, true
			}
			pk, perr := PubKeyFromProto(pba.PubKey)
			if perr != nil {
				return nil, perr, true
			}
			return authSigMessage{Key: pk, Sig: pba.Sig}, nil, false
		},
	)

	if trs.FirstError() != nil {
		err = trs.FirstError()
		return
	}
	return trs.FirstValue().(authSigMessage), nil
}

// incrSecConnNonce increments the 12-byte ChaCha20-Poly1305 nonce
// little-endian by 1, leaving bytes 0..4 reserved (zero) and using
// bytes 4..12 as a 64-bit counter. Panics on overflow rather than
// reusing a nonce.
func incrSecConnNonce(nonce *[scAEADNonceSize]byte) {
	counter := binary.LittleEndian.Uint64(nonce[4:])
	if counter == math.MaxUint64 {
		panic("upstream: SecretConnection nonce counter overflow")
	}
	counter++
	binary.LittleEndian.PutUint64(nonce[4:], counter)
}
