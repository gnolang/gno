package secp256k1

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"math/rand/v2"
	"os"
	"strings"
	"testing"

	dcrsecp "github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
	"golang.org/x/crypto/sha3"
)

// The corpus in testdata/ was generated from public test keys with
// go-ethereum v1.10.21: its pure-Go recovery, its libsecp256k1 recovery and
// its Berlin ecrecover precompile.
//
//   - vectors.json: 28 digest/signature pairs, including high-S twins, wrong
//     parities and digests, out-of-range scalars, an r that is no curve
//     point, and wrong lengths.
//   - proposed-boundary.json: the outcome all three agree on once the digest
//     is 32 bytes, the signature 65 bytes and the recovery ID 0 or 1, which is
//     what RecoverPubkey enforces.
//   - recovery-ids-pure-go.json: for two signatures (the first vector, and
//     the same with r = 2), which of the 256 recovery bytes the precompile
//     accepts and the address each recovers.

type corpusVector struct {
	Name      string `json:"name"`
	Digest    string `json:"digest_hex"`
	Signature string `json:"signature_hex"`
}

type corpusOutcome struct {
	Status    string `json:"status"`
	PublicKey string `json:"public_key_hex"`
}

func readCorpus(t *testing.T, name string, v any) {
	t.Helper()
	b, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(b, v); err != nil {
		t.Fatal(err)
	}
}

func mustHex(t *testing.T, s string) []byte {
	t.Helper()
	b, err := hex.DecodeString(s)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// ethAddress returns the hex Ethereum address of an uncompressed key.
func ethAddress(pubkey [65]byte) string {
	h := sha3.NewLegacyKeccak256()
	h.Write(pubkey[1:])
	return hex.EncodeToString(h.Sum(nil)[12:])
}

func TestRecoverPubkeyCorpus(t *testing.T) {
	var vectors []corpusVector
	var want map[string]corpusOutcome
	readCorpus(t, "vectors.json", &vectors)
	readCorpus(t, "proposed-boundary.json", &want)
	if len(vectors) != 28 || len(want) != 28 {
		t.Fatalf("corpus has %d vectors and %d outcomes, want 28 each", len(vectors), len(want))
	}
	fixedSize := 0
	for _, v := range vectors {
		d, s := mustHex(t, v.Digest), mustHex(t, v.Signature)
		// Other lengths never reach the native: RecoverPubkey refuses them in
		// Gno (recover_test.gno, and the VM test in gnovm/stdlibs).
		if len(d) != 32 || len(s) != 65 {
			continue
		}
		fixedSize++
		t.Run(v.Name, func(t *testing.T) {
			w, found := want[v.Name]
			if !found {
				t.Fatal("no expected outcome")
			}
			key, ok := X_recoverPubkey([32]byte(d), [65]byte(s))
			switch w.Status {
			case "recovered":
				if !ok || hex.EncodeToString(key[:]) != w.PublicKey {
					t.Errorf("got %x, %v; want %s", key, ok, w.PublicKey)
				}
			case "rejected":
				if ok || key != [65]byte{} {
					t.Errorf("got %x, %v; want a zero key and false", key, ok)
				}
			default:
				t.Fatalf("unknown status %q", w.Status)
			}
		})
	}
	if fixedSize != 19 {
		t.Fatalf("ran %d fixed-size vectors, want 19", fixedSize)
	}
}

func TestRecoverPubkeyRecoveryIDs(t *testing.T) {
	var vectors []corpusVector
	readCorpus(t, "vectors.json", &vectors)
	var ids map[string]map[string]struct {
		Recovered     map[string][]int `json:"recovered_address_to_ids"`
		RejectedCount int              `json:"rejected_count"`
	}
	readCorpus(t, "recovery-ids-pure-go.json", &ids)
	if vectors[0].Name != "signer-0/canonical" {
		t.Fatalf("first vector is %q", vectors[0].Name)
	}
	digest := [32]byte(mustHex(t, vectors[0].Digest))
	for _, seed := range []string{"fixture-signature-0", "small-r-2"} {
		evm := ids[seed]["evm_precompile"]
		want := map[int]string{}
		for address, accepted := range evm.Recovered {
			for _, id := range accepted {
				want[id] = strings.ToLower(strings.TrimPrefix(address, "0x"))
			}
		}
		if len(want)+evm.RejectedCount != 256 {
			t.Fatalf("%s: %d accepted + %d rejected IDs, want 256", seed, len(want), evm.RejectedCount)
		}
		sig := [65]byte(mustHex(t, vectors[0].Signature))
		if seed == "small-r-2" {
			clear(sig[:32])
			sig[31] = 2
		}
		for id := range 256 {
			sig[64] = byte(id)
			key, ok := X_recoverPubkey(digest, sig)
			address, accepted := want[id]
			switch {
			case ok != accepted:
				t.Errorf("%s, ID %d: ok = %v, want %v", seed, id, ok, accepted)
			case !ok && key != [65]byte{}:
				t.Errorf("%s, ID %d: refused with a non-zero key", seed, id)
			case ok && ethAddress(key) != address:
				t.Errorf("%s, ID %d: address %s, want %s", seed, id, ethAddress(key), address)
			}
		}
	}
}

func randomBytes(rng *rand.Rand, b []byte) {
	for i := range b {
		b[i] = byte(rng.Uint32())
	}
}

// signRecoverable signs digest and returns r || s || recovery ID, the form
// RecoverPubkey takes. SignCompact returns a header byte (27 plus the
// recovery ID, for an uncompressed key), then r and s.
func signRecoverable(t *testing.T, priv *dcrsecp.PrivateKey, digest [32]byte) [65]byte {
	t.Helper()
	compact := ecdsa.SignCompact(priv, digest[:], false)
	var sig [65]byte
	copy(sig[:64], compact[1:])
	sig[64] = compact[0] - compactUncompressedHeader
	if sig[64] > 1 {
		t.Fatalf("signature needs recovery ID %d", sig[64])
	}
	return sig
}

// TestRecoverPubkeyRoundTrip signs random digests with random keys and
// recovers each key from the signature and from its high-S twin.
func TestRecoverPubkeyRoundTrip(t *testing.T) {
	rng := rand.New(rand.NewPCG(5688, 1))
	for i := range 256 {
		var secret, digest [32]byte
		randomBytes(rng, secret[:])
		randomBytes(rng, digest[:])
		priv := dcrsecp.PrivKeyFromBytes(secret[:])
		want := priv.PubKey().SerializeUncompressed()
		sig := signRecoverable(t, priv, digest)
		if key, ok := X_recoverPubkey(digest, sig); !ok || !bytes.Equal(key[:], want) {
			t.Fatalf("signature %d: got %x, %v; want %x", i, key, ok, want)
		}

		// (r, n-s) with the other recovery ID recovers the same key.
		var s dcrsecp.ModNScalar
		s.SetByteSlice(sig[32:64])
		negated := s.Negate().Bytes()
		highS := sig
		copy(highS[32:64], negated[:])
		highS[64] ^= 1
		if key, ok := X_recoverPubkey(digest, highS); !ok || !bytes.Equal(key[:], want) {
			t.Fatalf("signature %d, high S: got %x, %v; want %x", i, key, ok, want)
		}

		// The other recovery ID alone recovers some other key.
		wrong := sig
		wrong[64] ^= 1
		if key, ok := X_recoverPubkey(digest, wrong); ok && bytes.Equal(key[:], want) {
			t.Fatalf("signature %d: the wrong recovery ID recovered the signer", i)
		}
	}
}

// TestRecoverPubkeyMatchesVerify checks that Verify accepts the key that
// RecoverPubkey returns, for the same message and signature. Verify goes
// through tm2's verifier, so with -tags libsecp256k1 this checks the pure-Go
// recovery against libsecp256k1.
func TestRecoverPubkeyMatchesVerify(t *testing.T) {
	rng := rand.New(rand.NewPCG(5688, 2))
	for i := range 64 {
		var secret [32]byte
		randomBytes(rng, secret[:])
		msg := make([]byte, rng.IntN(128))
		randomBytes(rng, msg)
		// Verify hashes the message with SHA-256, so recovery gets that hash.
		digest := sha256.Sum256(msg)
		sig := signRecoverable(t, dcrsecp.PrivKeyFromBytes(secret[:]), digest)

		key, ok := X_recoverPubkey(digest, sig)
		if !ok {
			t.Fatalf("message %d: recovery refused", i)
		}
		pub, err := dcrsecp.ParsePubKey(key[:])
		if err != nil {
			t.Fatalf("message %d: recovered key does not parse: %v", i, err)
		}
		if !X_verify(pub.SerializeCompressed(), msg, sig[:64]) {
			t.Fatalf("message %d: Verify refuses the recovered key", i)
		}
	}
}

// TestRecoverPubkeyEdgeCases pins inputs the corpus lacks.
func TestRecoverPubkeyEdgeCases(t *testing.T) {
	const (
		one = "0000000000000000000000000000000000000000000000000000000000000001"
		gx  = "79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798"
		// Signer 0 of the corpus: private key 1, so its key is the generator.
		generator = "0479be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798483ada7726a3c4655da4fbfc0e1108a8fd17b448a68554199c47d08ffb10d4b8"
	)
	for _, c := range []struct {
		name, digest, signature, want string // want "" means refused
	}{
		// Digest 1, r = Gx, s = 1: ID 0 makes the key G - G, the point at
		// infinity, so it is refused after both multiplications; ID 1 takes
		// X = -G and recovers another key, -(2/r)G.
		{"point at infinity", one, gx + one + "00", ""},
		{"same input, other ID", one, gx + one + "01", "0439f29c5586f5542b3c01785850e7a32906e19292b9a01c42ab886179418bcab56290b3832738776e934824a8eaa0df5f9caabf2f3fd7272575d9cb74e2844209"},
		// With a digest of their choosing, anyone can recover any key: here
		// signer 0's key from digest 0, with no private key involved. This is
		// why RecoverPubkey's doc forbids taking the digest from untrusted input.
		{"chosen digest", strings.Repeat("0", 64), "5cbdf0646e5db4eaa398f365f2ea7a0e3d419b7e0330e39ce92bddedcac4f9bce8ad6b7c0fc43e6aa9a822c56bd85a933b0cf820dfd78549a1284c2cf401802e00", generator},
		// s chosen so that u2 = s/r has a dense signed-digit form: the slow
		// input of the calibration benches. It recovers some key.
		{"dense s", "4c358ead6f3ea252278a6c0e69069eecec31c517786eba631a79352dc8f61c56", "a1842bb853abc7b49fe40f5af46b0f75aa6c962219e689f5a7ccea2b81d4f4e8472aebbbdd5aa013207fa01a3a071b7b5890278fd3c1f76823ad13ba4f75fac800", "04236d5484ec1ab3cbde4a534e605538a3e1506c3672ccba420979141027fa17f52315f37838ecdea2f3022a50d8f3e578549fe079cff6bd935117bbd3a90b3be5"},
	} {
		key, ok := X_recoverPubkey([32]byte(mustHex(t, c.digest)), [65]byte(mustHex(t, c.signature)))
		if c.want == "" {
			if ok || key != [65]byte{} {
				t.Errorf("%s: got %x, %v; want a zero key and false", c.name, key, ok)
			}
		} else if !ok || hex.EncodeToString(key[:]) != c.want {
			t.Errorf("%s: got %x, %v; want %s", c.name, key, ok, c.want)
		}
	}
}

// FuzzRecoverPubkey checks that no 97-byte input panics, that a refusal comes
// with a zero key, and that a recovered key is a valid uncompressed point. The
// input's first 32 bytes are the digest and the next 65 the signature; shorter
// inputs are zero-padded. Plain `go test` runs the corpus seeds only.
func FuzzRecoverPubkey(f *testing.F) {
	b, err := os.ReadFile("testdata/vectors.json")
	if err != nil {
		f.Fatal(err)
	}
	var vectors []corpusVector
	if err := json.Unmarshal(b, &vectors); err != nil {
		f.Fatal(err)
	}
	for _, v := range vectors {
		d, err1 := hex.DecodeString(v.Digest)
		s, err2 := hex.DecodeString(v.Signature)
		if err1 == nil && err2 == nil && len(d) == 32 && len(s) == 65 {
			f.Add(append(d, s...))
		}
	}
	f.Fuzz(func(t *testing.T, in []byte) {
		var digest [32]byte
		var sig [65]byte
		copy(digest[:], in)
		if len(in) > 32 {
			copy(sig[:], in[32:])
		}
		key, ok := X_recoverPubkey(digest, sig)
		if !ok {
			if key != [65]byte{} {
				t.Fatalf("refused with a non-zero key %x", key)
			}
			return
		}
		if _, err := dcrsecp.ParsePubKey(key[:]); err != nil || key[0] != 0x04 {
			t.Fatalf("recovered %x, which is not an uncompressed point: %v", key, err)
		}
	})
}
