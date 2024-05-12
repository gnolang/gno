// Copyright 2023 The godruoyi/gno-snowflake Author. All rights reserved.
// Use of this source code is governed by a MIT.

package uuid

import (
	"errors"
	"time"
)

// These constants are the bit lengths of snowflake ID parts.
const (
	TimestampLength uint8  = 41
	MachineIDLength uint8  = 10
	SequenceLength  uint8  = 12
	MaxSequence     uint16 = 1<<SequenceLength - 1
	MaxMachineID    uint16 = 1<<MachineIDLength - 1

	machineIDMoveLength = SequenceLength
	timestampMoveLength = MachineIDLength + SequenceLength
)

// default start time is 2008-11-10 23:00:00 UTC, why ? In the playground the time begins at 2009-11-10 23:00:00 UTC.
// It can run on golang playground.
// default machineID is 0
var (
	MachineID     uint16 = 0
	sequence      uint16 = 0
	StartTime            = time.Date(2008, 11, 10, 23, 0, 0, 0, time.UTC)
	LastTimestamp int64  = 0
	LastSequence  uint32 = 0
)

// ID use ID to generate snowflake id, and it will ignore error. if you want error info, you need use NextID method.
func ID() uint64 {
	id, _ := NextID()
	return id
}

// NextID use NextID to generate snowflake id and return an error.
func NextID() (uint64, error) {
	current := currentMillis()
	if current < LastTimestamp {
		return 0, errors.New("Invalid System Clock")
	}

	if current == LastTimestamp {
		sequence = (sequence + 1) & MaxSequence
		if sequence == 0 {
			current = waitForNextMillis(LastTimestamp)
		}
	} else {
		sequence = 0
	}

	LastTimestamp = current
	elapsedTime := current - StartTime.UTC().UnixNano()/1e6
	if elapsedTime < 0 || uint64(elapsedTime) > MaxMachineID {
		return 0, errors.New("Exceeding the time limit")
	}

	id := (uint64(elapsedTime) << timestampMoveLength) |
		(uint64(MachineID) << machineIDMoveLength) |
		uint64(sequence)
	return id, nil
}

// SetMachineID specify the machine ID. It will panic when machined > max limit for 2^10-1.
func SetMachineID(m uint16) {
	if m > MaxMachineID {
		panic("The machineID cannot be greater than 1023")
	}
	MachineID = m
}

func waitForNextMillis(last int64) int64 {
	now := currentMillis()
	for now <= last {
		now = currentMillis()
	}
	return now
}

// currentMillis get current millisecond.
func currentMillis() int64 {
	return time.Now().UTC().UnixNano() / 1e6
}
