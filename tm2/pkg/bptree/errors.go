package bptree

import "errors"

var (
	ErrVersionDoesNotExist = errors.New("version does not exist")
	ErrKeyDoesNotExist     = errors.New("key does not exist")
	ErrExportDone          = errors.New("export done")
	ErrNotInitializedTree  = errors.New("tree not initialized")
	ErrNoImport            = errors.New("no import in progress")
	ErrNodeMissingNodeKey  = errors.New("node missing node key")
	ErrEmptyTree           = errors.New("tree is empty")
	ErrActiveReaders       = errors.New("version has active readers")
	ErrEmptyKey            = errors.New("key must not be empty")
)
