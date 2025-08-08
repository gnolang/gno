package internal

import (
	"errors"
	"sync"
)

type AtomicSetDeleter interface {
	Mutex() *sync.Mutex
	SetNoLock(key, value []byte)
	SetNoLockSync(key, value []byte)
	DeleteNoLock(key []byte)
	DeleteNoLockSync(key []byte)
}

type MemBatch struct {
	DB   AtomicSetDeleter
	Ops  []Operation
	Size int
}

type OpType int

const (
	OpTypeSet    OpType = 1
	OpTypeDelete OpType = 2
)

type Operation struct {
	OpType
	Key   []byte
	Value []byte
}

func (mBatch *MemBatch) Set(key, value []byte) error {
	mBatch.Size += len(key) + len(value)
	mBatch.Ops = append(mBatch.Ops, Operation{OpTypeSet, key, value})
	return nil
}

func (mBatch *MemBatch) Delete(key []byte) error {
	mBatch.Size += len(key)
	mBatch.Ops = append(mBatch.Ops, Operation{OpTypeDelete, key, nil})
	return nil
}

func (mBatch *MemBatch) Write() error {
	mBatch.write(false)
	return nil
}

func (mBatch *MemBatch) WriteSync() error {
	mBatch.write(true)
	return nil
}

func (mBatch *MemBatch) Close() error {
	mBatch.Ops = nil
	mBatch.Size = 0
	return nil
}

func (mBatch *MemBatch) GetByteSize() (int, error) {
	if mBatch.Ops == nil {
		return 0, errors.New("membatch: batch has been written or closed")
	}
	return mBatch.Size, nil
}

func (mBatch *MemBatch) write(doSync bool) {
	if mtx := mBatch.DB.Mutex(); mtx != nil {
		mtx.Lock()
		defer mtx.Unlock()
	}

	for i, op := range mBatch.Ops {
		if doSync && i == (len(mBatch.Ops)-1) {
			switch op.OpType {
			case OpTypeSet:
				mBatch.DB.SetNoLockSync(op.Key, op.Value)
			case OpTypeDelete:
				mBatch.DB.DeleteNoLockSync(op.Key)
			}
			break // we're done.
		}
		switch op.OpType {
		case OpTypeSet:
			mBatch.DB.SetNoLock(op.Key, op.Value)
		case OpTypeDelete:
			mBatch.DB.DeleteNoLock(op.Key)
		}
	}
}
