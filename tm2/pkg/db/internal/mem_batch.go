package internal

import "sync"

type AtomicSetDeleter interface {
	Mutex() *sync.Mutex
	SetNoLock(key, value []byte)
	SetNoLockSync(key, value []byte)
	DeleteNoLock(key []byte)
	DeleteNoLockSync(key []byte)
}

type MemBatch struct {
	DB  AtomicSetDeleter
	Ops []Operation
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

func (mBatch *MemBatch) Set(key, value []byte) {
	mBatch.Ops = append(mBatch.Ops, Operation{OpTypeSet, key, value})
}

func (mBatch *MemBatch) Delete(key []byte) {
	mBatch.Ops = append(mBatch.Ops, Operation{OpTypeDelete, key, nil})
}

func (mBatch *MemBatch) Write() {
	mBatch.write(false)
}

func (mBatch *MemBatch) WriteSync() {
	mBatch.write(true)
}

func (mBatch *MemBatch) Close() {
	mBatch.Ops = nil
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
