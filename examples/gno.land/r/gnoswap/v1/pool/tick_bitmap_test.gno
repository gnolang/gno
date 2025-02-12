package pool

import (
	"testing"

	"gno.land/p/demo/avl"
	u256 "gno.land/p/gnoswap/uint256"
)

func TestTickBitmapPosition(t *testing.T) {
	tests := []struct {
		name        string
		tick        int32
		wantWordPos int16
		wantBitPos  uint8
	}{
		{
			name:        "positive tick",
			tick:        300,
			wantWordPos: 1,  // 300 >> 8 = 1
			wantBitPos:  44, // 300 % 256 = 44
		},
		{
			name:        "negative tick",
			tick:        -300,
			wantWordPos: -2,  // -300 >> 8 = -2
			wantBitPos:  212, // -300 % 256 = 212
		},
		{
			name:        "zero tick",
			tick:        0,
			wantWordPos: 0,
			wantBitPos:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wordPos, bitPos := tickBitmapPosition(tt.tick)
			if wordPos != tt.wantWordPos {
				t.Errorf("wordPos = %v, want %v", wordPos, tt.wantWordPos)
			}
			if bitPos != tt.wantBitPos {
				t.Errorf("bitPos = %v, want %v", bitPos, tt.wantBitPos)
			}
		})
	}
}

func TestTickBitmapFlipTick(t *testing.T) {
	tests := []struct {
		name        string
		tick        int32
		tickSpacing int32
		shouldPanic bool
	}{
		{
			name:        "valid positive tick and spacing",
			tick:        100,
			tickSpacing: 20,
			shouldPanic: false,
		},
		{
			name:        "valid negative tick and spacing",
			tick:        -300,
			tickSpacing: 20,
			shouldPanic: false,
		},
		{
			name:        "invalid tick and spacing",
			tick:        101,
			tickSpacing: 20,
			shouldPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool := &Pool{
				tickBitmaps: avl.NewTree(),
			}

			if tt.shouldPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("expected panic but got none")
					}
				}()
			}

			pool.tickBitmapFlipTick(tt.tick, tt.tickSpacing)

			if !tt.shouldPanic {
				wordPos, bitPos := tickBitmapPosition(tt.tick / tt.tickSpacing)
				expected := new(u256.Uint).Lsh(u256.NewUint(1), uint(bitPos))
				if pool.getTickBitmap(wordPos).Cmp(expected) != 0 {
					t.Errorf("bitmap not set correctly")
				}
			}
		})
	}
}

func TestTickBitmapNextInitializedTickWithInOneWord(t *testing.T) {
	tests := []struct {
		name        string
		tick        int32
		tickSpacing int32
		lte         bool
		setupBitmap func(*Pool)
		wantTick    int32
		wantInit    bool
	}{
		{
			name:        "search lte with initialized tick",
			tick:        200,
			tickSpacing: 20,
			lte:         true,
			setupBitmap: func(p *Pool) {
				p.tickBitmapFlipTick(180, 20) // Initialize a tick at 180
			},
			wantTick: 180,
			wantInit: true,
		},
		{
			name:        "search gt with initialized tick",
			tick:        100,
			tickSpacing: 20,
			lte:         false,
			setupBitmap: func(p *Pool) {
				p.tickBitmapFlipTick(140, 20) // Initialize a tick at 140
			},
			wantTick: 140,
			wantInit: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool := &Pool{
				tickBitmaps: avl.NewTree(),
			}
			if tt.setupBitmap != nil {
				tt.setupBitmap(pool)
			}

			gotTick, gotInit := pool.tickBitmapNextInitializedTickWithInOneWord(
				tt.tick,
				tt.tickSpacing,
				tt.lte,
			)

			if gotTick != tt.wantTick {
				t.Errorf("tick = %v, want %v", gotTick, tt.wantTick)
			}
			if gotInit != tt.wantInit {
				t.Errorf("initialized = %v, want %v", gotInit, tt.wantInit)
			}
		})
	}
}
