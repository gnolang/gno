package restore

import (
	"context"
	"io"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	_ "github.com/gnolang/gno/gno.land/pkg/sdk/vm" // this is needed to load amino types
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/contribs/tx-archive/log/noop"
)

func TestRestore_ExecuteRestore(t *testing.T) {
	t.Parallel()

	var (
		exampleTxCount = 10
		exampleTxGiven = 0

		exampleTx = &std.Tx{
			Memo: "example tx",
		}

		sentTxs = make([]*std.Tx, 0)

		mockClient = &mockClient{
			sendTransactionFn: func(_ context.Context, tx *std.Tx) error {
				sentTxs = append(sentTxs, tx)

				return nil
			},
		}
		mockSource = &mockSource{
			nextFn: func(_ context.Context) (*std.Tx, error) {
				if exampleTxGiven == exampleTxCount {
					return nil, io.EOF
				}

				exampleTxGiven++

				return exampleTx, nil
			},
		}
	)

	s := NewService(mockClient, mockSource, WithLogger(noop.New()))

	// Execute the restore
	assert.NoError(
		t,
		s.ExecuteRestore(context.Background(), false),
	)

	// Verify the restore was correct
	assert.Len(t, sentTxs, exampleTxCount)

	for _, tx := range sentTxs {
		assert.Equal(t, exampleTx, tx)
	}
}

func TestRestore_ExecuteRestore_Watch(t *testing.T) {
	t.Parallel()

	ctx, cancelFn := context.WithCancel(context.Background())
	defer cancelFn()

	var (
		exampleTxCount = 20
		exampleTxGiven = 0

		simulateEOF atomic.Bool

		exampleTx = &std.Tx{
			Memo: "example tx",
		}

		sentTxs = make([]*std.Tx, 0)

		mockClient = &mockClient{
			sendTransactionFn: func(_ context.Context, tx *std.Tx) error {
				sentTxs = append(sentTxs, tx)

				return nil
			},
		}
		mockSource = &mockSource{
			nextFn: func(_ context.Context) (*std.Tx, error) {
				if simulateEOF.Load() {
					return nil, io.EOF
				}

				// ~ the half mark, cut off the tx stream
				// by simulating the end of the stream (temporarily)
				if exampleTxGiven == exampleTxCount/2 {
					// Simulate EOF, but after some time
					// make sure the Next call returns an actual transaction
					simulateEOF.Store(true)

					time.AfterFunc(
						50*time.Millisecond,
						func() {
							simulateEOF.Store(false)
						},
					)

					exampleTxGiven++

					return exampleTx, nil
				}

				if exampleTxGiven == exampleTxCount {
					// All transactions parsed, simulate
					// the user cancelling the context
					cancelFn()

					return nil, io.EOF
				}

				exampleTxGiven++

				return exampleTx, nil
			},
		}
	)

	s := NewService(mockClient, mockSource, WithLogger(noop.New()))
	s.watchInterval = 10 * time.Millisecond // make the interval almost instant for the test

	// Execute the restore
	assert.NoError(
		t,
		s.ExecuteRestore(
			ctx,
			true, // Enable watch
		),
	)

	// Verify the restore was correct
	assert.Len(t, sentTxs, exampleTxCount)

	for _, tx := range sentTxs {
		assert.Equal(t, exampleTx, tx)
	}
}

func TestRestore_BackwardCompatible(t *testing.T) {
	t.Parallel()

	oldTx := `{"tx":{"msg":[{"@type":"/vm.m_call","caller":"g1ngywvql2ql7t8uzl63w60eqcejkwg4rm4lxdw9",
	"send":"","pkg_path":"gno.land/r/demo/wugnot","func":"Approve","args":
	["g126swhfaq2vyvvjywevhgw7lv9hg8qan93dasu8","18446744073709551615"]},{"@type":"/vm.m_call","caller":
	"g1ngywvql2ql7t8uzl63w60eqcejkwg4rm4lxdw9","send":"","pkg_path":"gno.land/r/gnoswap/v2/gns","func":
	"Approve","args":["g126swhfaq2vyvvjywevhgw7lv9hg8qan93dasu8","18446744073709551615"]},
	{"@type":"/vm.m_call","caller":"g1ngywvql2ql7t8uzl63w60eqcejkwg4rm4lxdw9","send":"","pkg_path":
	"gno.land/r/demo/wugnot","func":"Approve","args":["g14fclvfqynndp0l6kpyxkpgn4sljw9rr96hz46l",
	"18446744073709551615"]},{"@type":"/vm.m_call","caller":
	"g1ngywvql2ql7t8uzl63w60eqcejkwg4rm4lxdw9","send":"","pkg_path":"gno.land/r/gnoswap/v2/position",
	"func":"CollectFee","args":["26"]},{"@type":"/vm.m_call","caller":"g1ngywvql2ql7t8uzl63w60eqcejkwg4rm4lxdw9",
	"send":"","pkg_path":"gno.land/r/gnoswap/v2/staker","func":"CollectReward","args":["26","true"]},
	{"@type":"/vm.m_call","caller":"g1ngywvql2ql7t8uzl63w60eqcejkwg4rm4lxdw9","send":"","pkg_path":
	"gno.land/r/demo/wugnot","func":"Approve","args":["g126swhfaq2vyvvjywevhgw7lv9hg8qan93dasu8",
	"18446744073709551615"]},{"@type":"/vm.m_call","caller":"g1ngywvql2ql7t8uzl63w60eqcejkwg4rm4lxdw9",
	"send":"","pkg_path":"gno.land/r/gnoswap/v2/gns","func":"Approve","args":["g126swhfaq2vyvvjywevhgw7lv9hg8qan93dasu8",
	"18446744073709551615"]},{"@type":"/vm.m_call","caller":"g1ngywvql2ql7t8uzl63w60eqcejkwg4rm4lxdw9","send":"",
	"pkg_path":"gno.land/r/gnoswap/v2/position","func":"CollectFee","args":["146"]}],"fee":{"gas_wanted":"100000000",
	"gas_fee":"1ugnot"},"signatures":[{"pub_key":{"@type":"/tm.PubKeySecp256k1",
	"value":"Atgv/+TCwlR+jzjx94p4Ik0IuGET4J/q2q9ciaL4UOQh"}, 
	"signature":"iVfxsF37nRtgqyq9tMRMhyFLxp5RVdpI1r0mSHLmdg5aly0w82/in0ECey2PSpRk2UQ/fCtMpyOzaqIXiVKC4Q=="}],
	"memo":""}}`

	var out gnoland.TxWithMetadata
	err := amino.UnmarshalJSON([]byte(oldTx), &out)
	require.NoError(t, err)

	require.Nil(t, out.Metadata)
	require.Len(t, out.Tx.Msgs, 8)
}
