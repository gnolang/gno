package gnoweb

import (
	"context"
	"errors"
	"testing"
	"time"
)

// TestCheckResponseSizeRejectsOversized — defense in depth against a
// misbehaving or compromised RPC node returning a multi-MB amino blob:
// gnoweb caps every per-query response so the decode pipeline cannot be
// pressured into a memory amplification attack.
func TestCheckResponseSizeRejectsOversized(t *testing.T) {
	for _, tc := range []struct {
		name    string
		size    int
		wantErr error
	}{
		{"empty", 0, nil},
		{"under cap", 1024, nil},
		{"at cap", maxRPCResponseSize, nil},
		{"one byte over cap", maxRPCResponseSize + 1, ErrClientResponseTooLarge},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := checkResponseSize(make([]byte, tc.size))
			if tc.wantErr == nil {
				if err != nil {
					t.Fatalf("err = %v, want nil", err)
				}
				return
			}
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("err = %v, want errors.Is(%v)", err, tc.wantErr)
			}
		})
	}
}

// TestAcquireRPCSlotBoundsConcurrency pins the semaphore contract:
// (a) cap parallelism, (b) honour ctx cancellation while waiting,
// (c) release frees exactly one slot.
func TestAcquireRPCSlotBoundsConcurrency(t *testing.T) {
	slots := make(chan struct{}, 2)

	rel1, err := acquireRPCSlot(context.Background(), slots)
	if err != nil {
		t.Fatalf("first acquire failed: %v", err)
	}
	rel2, err := acquireRPCSlot(context.Background(), slots)
	if err != nil {
		t.Fatalf("second acquire failed: %v", err)
	}

	// Bucket full — third acquire must block; ctx deadline triggers an
	// orderly cancellation rather than a stuck goroutine.
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	if _, err := acquireRPCSlot(ctx, slots); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected DeadlineExceeded waiting for slot, got %v", err)
	}

	// Release frees a slot — the next acquire succeeds immediately.
	rel1()
	ctx2, cancel2 := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel2()
	rel3, err := acquireRPCSlot(ctx2, slots)
	if err != nil {
		t.Fatalf("expected acquire after release, got %v", err)
	}
	rel3()
	rel2()

	// Bucket fully drained — len must be 0 (no slot leak from release fn).
	if len(slots) != 0 {
		t.Fatalf("slot leak: len(slots)=%d, want 0", len(slots))
	}
}
