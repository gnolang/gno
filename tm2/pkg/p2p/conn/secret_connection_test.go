package conn

import (
	"bufio"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/async"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	osm "github.com/gnolang/gno/tm2/pkg/os"
	"github.com/gnolang/gno/tm2/pkg/random"
)

type kvstoreConn struct {
	*io.PipeReader
	*io.PipeWriter
}

func (drw kvstoreConn) Close() (err error) {
	err2 := drw.PipeWriter.CloseWithError(io.EOF)
	err1 := drw.PipeReader.Close()
	if err2 != nil {
		return err
	}
	return err1
}

// Each returned ReadWriteCloser is akin to a net.Connection
func makeKVStoreConnPair() (fooConn, barConn kvstoreConn) {
	barReader, fooWriter := io.Pipe()
	fooReader, barWriter := io.Pipe()
	return kvstoreConn{fooReader, fooWriter}, kvstoreConn{barReader, barWriter}
}

func makeSecretConnPair(tb testing.TB) (fooSecConn, barSecConn *SecretConnection) {
	tb.Helper()

	fooConn, barConn := makeKVStoreConnPair()
	fooPrvKey := ed25519.GenPrivKey()
	fooPubKey := fooPrvKey.PubKey()
	barPrvKey := ed25519.GenPrivKey()
	barPubKey := barPrvKey.PubKey()

	// Make connections from both sides in parallel.
	trs, ok := async.Parallel(
		func(_ int) (val any, err error, abort bool) {
			fooSecConn, err = MakeSecretConnection(fooConn, fooPrvKey)
			if err != nil {
				tb.Errorf("Failed to establish SecretConnection for foo: %v", err)
				return nil, err, true
			}
			remotePubBytes := fooSecConn.RemotePubKey()
			if !remotePubBytes.Equals(barPubKey) {
				err = fmt.Errorf("Unexpected fooSecConn.RemotePubKey.  Expected %v, got %v",
					barPubKey, fooSecConn.RemotePubKey())
				tb.Error(err)
				return nil, err, false
			}
			return nil, nil, false
		},
		func(_ int) (val any, err error, abort bool) {
			barSecConn, err = MakeSecretConnection(barConn, barPrvKey)
			if barSecConn == nil {
				tb.Errorf("Failed to establish SecretConnection for bar: %v", err)
				return nil, err, true
			}
			remotePubBytes := barSecConn.RemotePubKey()
			if !remotePubBytes.Equals(fooPubKey) {
				err = fmt.Errorf("Unexpected barSecConn.RemotePubKey.  Expected %v, got %v",
					fooPubKey, barSecConn.RemotePubKey())
				tb.Error(err)
				return nil, nil, false
			}
			return nil, nil, false
		},
	)

	require.Nil(tb, trs.FirstError())
	require.True(tb, ok, "Unexpected task abortion")

	return fooSecConn, barSecConn
}

func TestSecretConnectionHandshake(t *testing.T) {
	t.Parallel()

	fooSecConn, barSecConn := makeSecretConnPair(t)
	if err := fooSecConn.Close(); err != nil {
		t.Error(err)
	}
	if err := barSecConn.Close(); err != nil {
		t.Error(err)
	}
}

// Test that shareEphPubKey rejects lower order public keys based on an
// (incomplete) blacklist.
func TestShareLowOrderPubkey(t *testing.T) {
	t.Parallel()

	fooConn, barConn := makeKVStoreConnPair()
	defer fooConn.Close()
	defer barConn.Close()
	locEphPub, _ := genEphKeys()

	// all blacklisted low order points:
	for _, remLowOrderPubKey := range blacklist {
		remLowOrderPubKey := remLowOrderPubKey
		_, _ = async.Parallel(
			func(_ int) (val any, err error, abort bool) {
				_, err = shareEphPubKey(fooConn, locEphPub)

				require.Error(t, err)
				require.Equal(t, err, ErrSmallOrderRemotePubKey)

				return nil, nil, false
			},
			func(_ int) (val any, err error, abort bool) {
				readRemKey, err := shareEphPubKey(barConn, &remLowOrderPubKey)

				require.NoError(t, err)
				require.Equal(t, locEphPub, readRemKey)

				return nil, nil, false
			})
	}
}

const lowOrderPointError = `crypto/ecdh: bad X25519 remote ECDH input: low order point`

// Test that additionally that the Diffie-Hellman shared secret is non-zero.
// The shared secret would be zero for lower order pub-keys (but tested against the blacklist only).
func TestComputeDHFailsOnLowOrder(t *testing.T) {
	t.Parallel()

	_, locPrivKey := genEphKeys()
	for _, remLowOrderPubKey := range blacklist {
		remLowOrderPubKey := remLowOrderPubKey
		shared, err := computeDHSecret(&remLowOrderPubKey, locPrivKey)
		_ = assert.Error(t, err) &&
			assert.Equal(t, lowOrderPointError, err.Error())
		assert.Empty(t, shared)
	}
}

func TestConcurrentWrite(t *testing.T) {
	t.Parallel()

	fooSecConn, barSecConn := makeSecretConnPair(t)
	fooWriteText := random.RandStr(dataMaxSize)

	// write from two routines.
	// should be safe from race according to net.Conn:
	// https://golang.org/pkg/net/#Conn
	n := 100
	wg := new(sync.WaitGroup)
	wg.Add(3)
	go writeLots(t, wg, fooSecConn, fooWriteText, n)
	go writeLots(t, wg, fooSecConn, fooWriteText, n)

	// Consume reads from bar's reader
	readLots(t, wg, barSecConn, n*2)
	wg.Wait()

	if err := fooSecConn.Close(); err != nil {
		t.Error(err)
	}
}

func TestConcurrentRead(t *testing.T) {
	t.Parallel()

	fooSecConn, barSecConn := makeSecretConnPair(t)
	fooWriteText := random.RandStr(dataMaxSize)
	n := 100

	// read from two routines.
	// should be safe from race according to net.Conn:
	// https://golang.org/pkg/net/#Conn
	wg := new(sync.WaitGroup)
	wg.Add(3)
	go readLots(t, wg, fooSecConn, n/2)
	go readLots(t, wg, fooSecConn, n/2)

	// write to bar
	writeLots(t, wg, barSecConn, fooWriteText, n)
	wg.Wait()

	if err := fooSecConn.Close(); err != nil {
		t.Error(err)
	}
}

func writeLots(t *testing.T, wg *sync.WaitGroup, conn net.Conn, txt string, n int) {
	t.Helper()

	defer wg.Done()
	for range n {
		_, err := conn.Write([]byte(txt))
		if err != nil {
			t.Errorf("Failed to write to fooSecConn: %v", err)
			return
		}
	}
}

func readLots(t *testing.T, wg *sync.WaitGroup, conn net.Conn, n int) {
	t.Helper()

	readBuffer := make([]byte, dataMaxSize)
	for range n {
		_, err := conn.Read(readBuffer)
		assert.NoError(t, err)
	}
	wg.Done()
}

func TestSecretConnectionReadWrite(t *testing.T) {
	t.Parallel()

	fooConn, barConn := makeKVStoreConnPair()
	fooWrites, barWrites := []string{}, []string{}
	fooReads, barReads := []string{}, []string{}

	// Pre-generate the things to write (for foo & bar)
	for range 100 {
		fooWrites = append(fooWrites, random.RandStr((random.RandInt()%(dataMaxSize*5))+1))
		barWrites = append(barWrites, random.RandStr((random.RandInt()%(dataMaxSize*5))+1))
	}

	// A helper that will run with (fooConn, fooWrites, fooReads) and vice versa
	genNodeRunner := func(_ string, nodeConn kvstoreConn, nodeWrites []string, nodeReads *[]string) async.Task {
		return func(_ int) (any, error, bool) {
			// Initiate cryptographic private key and secret connection through nodeConn.
			nodePrvKey := ed25519.GenPrivKey()
			nodeSecretConn, err := MakeSecretConnection(nodeConn, nodePrvKey)
			if err != nil {
				t.Errorf("Failed to establish SecretConnection for node: %v", err)
				return nil, err, true
			}
			// In parallel, handle some reads and writes.
			trs, ok := async.Parallel(
				func(_ int) (any, error, bool) {
					// Node writes:
					for _, nodeWrite := range nodeWrites {
						n, err := nodeSecretConn.Write([]byte(nodeWrite))
						if err != nil {
							t.Errorf("Failed to write to nodeSecretConn: %v", err)
							return nil, err, true
						}
						if n != len(nodeWrite) {
							err = fmt.Errorf("Failed to write all bytes. Expected %v, wrote %v", len(nodeWrite), n)
							t.Error(err)
							return nil, err, true
						}
					}
					if err := nodeConn.PipeWriter.Close(); err != nil {
						t.Error(err)
						return nil, err, true
					}
					return nil, nil, false
				},
				func(_ int) (any, error, bool) {
					// Node reads:
					readBuffer := make([]byte, dataMaxSize)
					for {
						n, err := nodeSecretConn.Read(readBuffer)
						if errors.Is(err, io.EOF) {
							if err := nodeConn.PipeReader.Close(); err != nil {
								t.Error(err)
								return nil, err, true
							}
							return nil, nil, false
						} else if err != nil {
							t.Errorf("Failed to read from nodeSecretConn: %v", err)
							return nil, err, true
						}
						*nodeReads = append(*nodeReads, string(readBuffer[:n]))
					}
				},
			)
			assert.True(t, ok, "Unexpected task abortion")

			// If error:
			if trs.FirstError() != nil {
				return nil, trs.FirstError(), true
			}

			// Otherwise:
			return nil, nil, false
		}
	}

	// Run foo & bar in parallel
	trs, ok := async.Parallel(
		genNodeRunner("foo", fooConn, fooWrites, &fooReads),
		genNodeRunner("bar", barConn, barWrites, &barReads),
	)
	require.Nil(t, trs.FirstError())
	require.True(t, ok, "unexpected task abortion")

	// A helper to ensure that the writes and reads match.
	// Additionally, small writes (<= dataMaxSize) must be atomically read.
	compareWritesReads := func(writes []string, reads []string) {
		for {
			// Pop next write & corresponding reads
			read, write := "", writes[0]
			readCount := 0
			for _, readChunk := range reads {
				read += readChunk
				readCount++
				if len(write) <= len(read) {
					break
				}
				if len(write) <= dataMaxSize {
					break // atomicity of small writes
				}
			}
			// Compare
			if write != read {
				t.Errorf("Expected to read %X, got %X", write, read)
			}
			// Iterate
			writes = writes[1:]
			reads = reads[readCount:]
			if len(writes) == 0 {
				break
			}
		}
	}

	compareWritesReads(fooWrites, barReads)
	compareWritesReads(barWrites, fooReads)
}

// Run go test -update from within this module
// to update the golden test vector file
var update = flag.Bool("update", false, "update .golden files")

func TestDeriveSecretsAndChallengeGolden(t *testing.T) {
	t.Parallel()

	goldenFilepath := filepath.Join("testdata", t.Name()+".golden")
	if *update {
		t.Logf("Updating golden test vector file %s", goldenFilepath)
		data := createGoldenTestVectors()
		osm.WriteFile(goldenFilepath, []byte(data), 0o644)
	}
	f, err := os.Open(goldenFilepath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		params := strings.Split(line, ",")
		randSecretVector, err := hex.DecodeString(params[0])
		require.Nil(t, err)
		randSecret := new([32]byte)
		copy((*randSecret)[:], randSecretVector)
		locIsLeast, err := strconv.ParseBool(params[1])
		require.Nil(t, err)
		expectedRecvSecret, err := hex.DecodeString(params[2])
		require.Nil(t, err)
		expectedSendSecret, err := hex.DecodeString(params[3])
		require.Nil(t, err)
		expectedChallenge, err := hex.DecodeString(params[4])
		require.Nil(t, err)

		recvSecret, sendSecret, challenge := deriveSecretAndChallenge(randSecret, locIsLeast)
		require.Equal(t, expectedRecvSecret, (*recvSecret)[:], "Recv Secrets aren't equal")
		require.Equal(t, expectedSendSecret, (*sendSecret)[:], "Send Secrets aren't equal")
		require.Equal(t, expectedChallenge, (*challenge)[:], "challenges aren't equal")
	}
}

// Creates the data for a test vector file.
// The file format is:
// Hex(diffie_hellman_secret), loc_is_least, Hex(recvSecret), Hex(sendSecret), Hex(challenge)
func createGoldenTestVectors() string {
	data := ""
	for range 32 {
		randSecretVector := random.RandBytes(32)
		randSecret := new([32]byte)
		copy((*randSecret)[:], randSecretVector)
		data += hex.EncodeToString((*randSecret)[:]) + ","
		locIsLeast := random.RandBool()
		data += strconv.FormatBool(locIsLeast) + ","
		recvSecret, sendSecret, challenge := deriveSecretAndChallenge(randSecret, locIsLeast)
		data += hex.EncodeToString((*recvSecret)[:]) + ","
		data += hex.EncodeToString((*sendSecret)[:]) + ","
		data += hex.EncodeToString((*challenge)[:]) + "\n"
	}
	return data
}

func BenchmarkWriteSecretConnection(b *testing.B) {
	b.StopTimer()
	b.ReportAllocs()
	fooSecConn, barSecConn := makeSecretConnPair(b)
	randomMsgSizes := []int{
		dataMaxSize / 10,
		dataMaxSize / 3,
		dataMaxSize / 2,
		dataMaxSize,
		dataMaxSize * 3 / 2,
		dataMaxSize * 2,
		dataMaxSize * 7 / 2,
	}
	fooWriteBytes := make([][]byte, 0, len(randomMsgSizes))
	for _, size := range randomMsgSizes {
		fooWriteBytes = append(fooWriteBytes, random.RandBytes(size))
	}
	// Consume reads from bar's reader
	go func() {
		readBuffer := make([]byte, dataMaxSize)
		for {
			_, err := barSecConn.Read(readBuffer)
			if errors.Is(err, io.EOF) {
				return
			} else if err != nil {
				b.Errorf("Failed to read from barSecConn: %v", err)
				return
			}
		}
	}()

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		idx := random.RandIntn(len(fooWriteBytes))
		_, err := fooSecConn.Write(fooWriteBytes[idx])
		if err != nil {
			b.Errorf("Failed to write to fooSecConn: %v", err)
			return
		}
	}
	b.StopTimer()

	if err := fooSecConn.Close(); err != nil {
		b.Error(err)
	}
	// barSecConn.Close() race condition
}

func BenchmarkReadSecretConnection(b *testing.B) {
	b.StopTimer()
	b.ReportAllocs()
	fooSecConn, barSecConn := makeSecretConnPair(b)
	randomMsgSizes := []int{
		dataMaxSize / 10,
		dataMaxSize / 3,
		dataMaxSize / 2,
		dataMaxSize,
		dataMaxSize * 3 / 2,
		dataMaxSize * 2,
		dataMaxSize * 7 / 2,
	}
	fooWriteBytes := make([][]byte, 0, len(randomMsgSizes))
	for _, size := range randomMsgSizes {
		fooWriteBytes = append(fooWriteBytes, random.RandBytes(size))
	}
	go func() {
		for i := 0; i < b.N; i++ {
			idx := random.RandIntn(len(fooWriteBytes))
			_, err := fooSecConn.Write(fooWriteBytes[idx])
			if err != nil {
				b.Errorf("Failed to write to fooSecConn: %v, %v,%v", err, i, b.N)
				return
			}
		}
	}()

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		readBuffer := make([]byte, dataMaxSize)
		_, err := barSecConn.Read(readBuffer)

		if errors.Is(err, io.EOF) {
			return
		} else if err != nil {
			b.Fatalf("Failed to read from barSecConn: %v", err)
		}
	}
	b.StopTimer()
}
