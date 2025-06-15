package p2p

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/cmap"
	"github.com/gnolang/gno/tm2/pkg/p2p/config"
	"github.com/gnolang/gno/tm2/pkg/p2p/conn"
	"github.com/gnolang/gno/tm2/pkg/p2p/mock"
	"github.com/gnolang/gno/tm2/pkg/p2p/types"
	"github.com/gnolang/gno/tm2/pkg/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPeer_Properties(t *testing.T) {
	t.Parallel()

	t.Run("connection info", func(t *testing.T) {
		t.Parallel()

		t.Run("remote IP", func(t *testing.T) {
			t.Parallel()

			var (
				info = &ConnInfo{
					RemoteIP: net.IP{127, 0, 0, 1},
				}

				p = &peer{
					connInfo: info,
				}
			)

			assert.Equal(t, info.RemoteIP, p.RemoteIP())
		})

		t.Run("remote address", func(t *testing.T) {
			t.Parallel()

			tcpAddr, err := net.ResolveTCPAddr("tcp", "localhost:8080")
			require.NoError(t, err)

			var (
				info = &ConnInfo{
					Conn: &mock.Conn{
						RemoteAddrFn: func() net.Addr {
							return tcpAddr
						},
					},
				}

				p = &peer{
					connInfo: info,
				}
			)

			assert.Equal(t, tcpAddr.String(), p.RemoteAddr().String())
		})

		t.Run("socket address", func(t *testing.T) {
			t.Parallel()

			tcpAddr, err := net.ResolveTCPAddr("tcp", "localhost:8080")
			require.NoError(t, err)

			netAddr, err := types.NewNetAddress(types.GenerateNodeKey().ID(), tcpAddr)
			require.NoError(t, err)

			var (
				info = &ConnInfo{
					SocketAddr: netAddr,
				}

				p = &peer{
					connInfo: info,
				}
			)

			assert.Equal(t, netAddr.String(), p.SocketAddr().String())
		})

		t.Run("set logger", func(t *testing.T) {
			t.Parallel()

			var (
				l = slog.New(slog.NewTextHandler(io.Discard, nil))

				p = &peer{
					mConn: &mock.MConn{},
				}
			)

			p.SetLogger(l)

			assert.Equal(t, l, p.Logger)
		})

		t.Run("peer start", func(t *testing.T) {
			t.Parallel()

			var (
				expectedErr = errors.New("some error")

				mConn = &mock.MConn{
					StartFn: func() error {
						return expectedErr
					},
				}

				p = &peer{
					mConn: mConn,
				}
			)

			assert.ErrorIs(t, p.OnStart(), expectedErr)
		})

		t.Run("peer stop", func(t *testing.T) {
			t.Parallel()

			var (
				stopCalled  = false
				expectedErr = errors.New("some error")

				mConn = &mock.MConn{
					StopFn: func() error {
						stopCalled = true

						return expectedErr
					},
				}

				p = &peer{
					mConn: mConn,
				}
			)

			p.BaseService = *service.NewBaseService(nil, "Peer", p)

			p.OnStop()

			assert.True(t, stopCalled)
		})

		t.Run("flush stop", func(t *testing.T) {
			t.Parallel()

			var (
				stopCalled = false

				mConn = &mock.MConn{
					FlushFn: func() {
						stopCalled = true
					},
				}

				p = &peer{
					mConn: mConn,
				}
			)

			p.BaseService = *service.NewBaseService(nil, "Peer", p)

			p.FlushStop()

			assert.True(t, stopCalled)
		})

		t.Run("node info fetch", func(t *testing.T) {
			t.Parallel()

			var (
				info = types.NodeInfo{
					Network: "gnoland",
				}

				p = &peer{
					nodeInfo: info,
				}
			)

			assert.Equal(t, info, p.NodeInfo())
		})

		t.Run("node status fetch", func(t *testing.T) {
			t.Parallel()

			var (
				status = conn.ConnectionStatus{
					Duration: 5 * time.Second,
				}

				mConn = &mock.MConn{
					StatusFn: func() conn.ConnectionStatus {
						return status
					},
				}

				p = &peer{
					mConn: mConn,
				}
			)

			assert.Equal(t, status, p.Status())
		})

		t.Run("string representation", func(t *testing.T) {
			t.Parallel()

			testTable := []struct {
				name     string
				outbound bool
			}{
				{
					"outbound",
					true,
				},
				{
					"inbound",
					false,
				},
			}

			for _, testCase := range testTable {
				t.Run(testCase.name, func(t *testing.T) {
					t.Parallel()

					var (
						id       = types.GenerateNodeKey().ID()
						mConnStr = "description"

						p = &peer{
							mConn: &mock.MConn{
								StringFn: func() string {
									return mConnStr
								},
							},
							nodeInfo: types.NodeInfo{
								NetAddress: &types.NetAddress{
									ID: id,
								},
							},
							connInfo: &ConnInfo{
								Outbound: testCase.outbound,
							},
						}

						direction = "in"
					)

					if testCase.outbound {
						direction = "out"
					}

					assert.Contains(
						t,
						p.String(),
						fmt.Sprintf(
							"Peer{%s %s %s}",
							mConnStr,
							id,
							direction,
						),
					)
				})
			}
		})

		t.Run("outbound information", func(t *testing.T) {
			t.Parallel()

			p := &peer{
				connInfo: &ConnInfo{
					Outbound: true,
				},
			}

			assert.True(
				t,
				p.IsOutbound(),
			)
		})

		t.Run("persistent information", func(t *testing.T) {
			t.Parallel()

			p := &peer{
				connInfo: &ConnInfo{
					Persistent: true,
				},
			}

			assert.True(t, p.IsPersistent())
		})

		t.Run("initial conn close", func(t *testing.T) {
			t.Parallel()

			var (
				closeErr = errors.New("close error")

				mockConn = &mock.Conn{
					CloseFn: func() error {
						return closeErr
					},
				}

				p = &peer{
					connInfo: &ConnInfo{
						Conn: mockConn,
					},
				}
			)

			assert.ErrorIs(t, p.CloseConn(), closeErr)
		})
	})
}

func TestPeer_GetSet(t *testing.T) {
	t.Parallel()

	var (
		key  = "key"
		data = []byte("random")

		p = &peer{
			data: cmap.NewCMap(),
		}
	)

	assert.Nil(t, p.Get(key))

	// Set the key
	p.Set(key, data)

	assert.Equal(t, data, p.Get(key))
}

func TestPeer_Send(t *testing.T) {
	t.Parallel()

	t.Run("peer not running", func(t *testing.T) {
		t.Parallel()

		var (
			chID = byte(10)
			data = []byte("random")

			capturedSendID   byte
			capturedSendData []byte

			mockConn = &mock.MConn{
				SendFn: func(c byte, d []byte) bool {
					capturedSendID = c
					capturedSendData = d

					return true
				},
			}

			p = &peer{
				nodeInfo: types.NodeInfo{
					Channels: []byte{
						chID,
					},
				},
				mConn: mockConn,
			}
		)

		p.BaseService = *service.NewBaseService(nil, "Peer", p)

		// Make sure the send fails
		require.False(t, p.Send(chID, data))

		assert.Empty(t, capturedSendID)
		assert.Nil(t, capturedSendData)
	})

	t.Run("peer doesn't have channel", func(t *testing.T) {
		t.Parallel()

		var (
			chID = byte(10)
			data = []byte("random")

			capturedSendID   byte
			capturedSendData []byte

			mockConn = &mock.MConn{
				SendFn: func(c byte, d []byte) bool {
					capturedSendID = c
					capturedSendData = d

					return true
				},
			}

			p = &peer{
				nodeInfo: types.NodeInfo{
					Channels: []byte{},
				},
				mConn: mockConn,
			}
		)

		p.BaseService = *service.NewBaseService(nil, "Peer", p)

		// Start the peer "multiplexing"
		require.NoError(t, p.Start())
		t.Cleanup(func() {
			require.NoError(t, p.Stop())
		})

		// Make sure the send fails
		require.False(t, p.Send(chID, data))

		assert.Empty(t, capturedSendID)
		assert.Nil(t, capturedSendData)
	})

	t.Run("valid peer data send", func(t *testing.T) {
		t.Parallel()

		var (
			chID = byte(10)
			data = []byte("random")

			capturedSendID   byte
			capturedSendData []byte

			mockConn = &mock.MConn{
				SendFn: func(c byte, d []byte) bool {
					capturedSendID = c
					capturedSendData = d

					return true
				},
			}

			p = &peer{
				nodeInfo: types.NodeInfo{
					Channels: []byte{
						chID,
					},
				},
				mConn: mockConn,
			}
		)

		p.BaseService = *service.NewBaseService(nil, "Peer", p)

		// Start the peer "multiplexing"
		require.NoError(t, p.Start())
		t.Cleanup(func() {
			require.NoError(t, p.Stop())
		})

		// Make sure the send is valid
		require.True(t, p.Send(chID, data))

		assert.Equal(t, chID, capturedSendID)
		assert.Equal(t, data, capturedSendData)
	})
}

func TestPeer_TrySend(t *testing.T) {
	t.Parallel()

	t.Run("peer not running", func(t *testing.T) {
		t.Parallel()

		var (
			chID = byte(10)
			data = []byte("random")

			capturedSendID   byte
			capturedSendData []byte

			mockConn = &mock.MConn{
				TrySendFn: func(c byte, d []byte) bool {
					capturedSendID = c
					capturedSendData = d

					return true
				},
			}

			p = &peer{
				nodeInfo: types.NodeInfo{
					Channels: []byte{
						chID,
					},
				},
				mConn: mockConn,
			}
		)

		p.BaseService = *service.NewBaseService(nil, "Peer", p)

		// Make sure the send fails
		require.False(t, p.TrySend(chID, data))

		assert.Empty(t, capturedSendID)
		assert.Nil(t, capturedSendData)
	})

	t.Run("peer doesn't have channel", func(t *testing.T) {
		t.Parallel()

		var (
			chID = byte(10)
			data = []byte("random")

			capturedSendID   byte
			capturedSendData []byte

			mockConn = &mock.MConn{
				TrySendFn: func(c byte, d []byte) bool {
					capturedSendID = c
					capturedSendData = d

					return true
				},
			}

			p = &peer{
				nodeInfo: types.NodeInfo{
					Channels: []byte{},
				},
				mConn: mockConn,
			}
		)

		p.BaseService = *service.NewBaseService(nil, "Peer", p)

		// Start the peer "multiplexing"
		require.NoError(t, p.Start())
		t.Cleanup(func() {
			require.NoError(t, p.Stop())
		})

		// Make sure the send fails
		require.False(t, p.TrySend(chID, data))

		assert.Empty(t, capturedSendID)
		assert.Nil(t, capturedSendData)
	})

	t.Run("valid peer data send", func(t *testing.T) {
		t.Parallel()

		var (
			chID = byte(10)
			data = []byte("random")

			capturedSendID   byte
			capturedSendData []byte

			mockConn = &mock.MConn{
				TrySendFn: func(c byte, d []byte) bool {
					capturedSendID = c
					capturedSendData = d

					return true
				},
			}

			p = &peer{
				nodeInfo: types.NodeInfo{
					Channels: []byte{
						chID,
					},
				},
				mConn: mockConn,
			}
		)

		p.BaseService = *service.NewBaseService(nil, "Peer", p)

		// Start the peer "multiplexing"
		require.NoError(t, p.Start())
		t.Cleanup(func() {
			require.NoError(t, p.Stop())
		})

		// Make sure the send is valid
		require.True(t, p.TrySend(chID, data))

		assert.Equal(t, chID, capturedSendID)
		assert.Equal(t, data, capturedSendData)
	})
}

func TestPeer_NewPeer(t *testing.T) {
	t.Parallel()

	tcpAddr, err := net.ResolveTCPAddr("tcp", "localhost:8080")
	require.NoError(t, err)

	netAddr, err := types.NewNetAddress(types.GenerateNodeKey().ID(), tcpAddr)
	require.NoError(t, err)

	var (
		connInfo = &ConnInfo{
			Outbound:   false,
			Persistent: true,
			Conn:       &mock.Conn{},
			RemoteIP:   tcpAddr.IP,
			SocketAddr: netAddr,
		}

		mConfig = &ConnConfig{
			MConfig:      conn.MConfigFromP2P(config.DefaultP2PConfig()),
			ReactorsByCh: make(map[byte]Reactor),
			ChDescs:      make([]*conn.ChannelDescriptor, 0),
			OnPeerError:  nil,
		}
	)

	assert.NotPanics(t, func() {
		_ = newPeer(connInfo, types.NodeInfo{}, mConfig)
	})
}
