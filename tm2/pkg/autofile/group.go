package autofile

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gnolang/gno/tm2/pkg/service"
)

const (
	defaultGroupCheckDuration = 5000 * time.Millisecond
	defaultHeadSizeLimit      = 10 * 1024 * 1024       // 10MB
	defaultTotalSizeLimit     = 1 * 1024 * 1024 * 1024 // 1GB
	maxFilesToRemove          = 4                      // needs to be greater than 1
)

/*
You can open a Group to keep restrictions on an AutoFile, like
the maximum size of each chunk, and/or the total amount of bytes
stored in the group.

The first file to be written in the Group.Dir is the head file.

	Dir/
	- <HeadPath>

Once the Head file reaches the size limit, it will be rotated.

	Dir/
	- <HeadPath>.000   // First rolled file
	- <HeadPath>       // New head path, starts empty.
										 // The implicit index is 001.

As more files are written, the index numbers grow...

	Dir/
	- <HeadPath>.000   // First rolled file
	- <HeadPath>.001   // Second rolled file
	- ...
	- <HeadPath>       // New head path

The Group can also be used to binary-search for some line,
assuming that marker lines are written occasionally.
*/
type Group struct {
	service.BaseService

	ID      string
	Head    *AutoFile // The head AutoFile to write to
	headBuf *bufio.Writer
	Dir     string // Directory that contains .Head

	mtx            sync.Mutex
	headSizeLimit  int64
	totalSizeLimit int64
	info           GroupInfo

	// TODO: When we start deleting files, we need to start tracking GroupReaders
	// and their dependencies.
}

// OpenGroup creates a new Group with head at headPath. It returns an error if
// it fails to open head file.
func OpenGroup(headPath string, groupOptions ...func(*Group)) (g *Group, err error) {
	dir := path.Dir(headPath)
	head, err := OpenAutoFile(headPath)
	if err != nil {
		return nil, err
	}

	g = &Group{
		ID:             "group:" + head.ID,
		Head:           head,
		headBuf:        bufio.NewWriterSize(head, 4096*10),
		Dir:            dir,
		headSizeLimit:  defaultHeadSizeLimit,
		totalSizeLimit: defaultTotalSizeLimit,
		info: GroupInfo{
			MinIndex:  0,
			MaxIndex:  0,
			TotalSize: 0,
			HeadSize:  0,
		},
	}

	for _, option := range groupOptions {
		option(g)
	}

	g.BaseService = *service.NewBaseService(nil, "Group", g)
	g.info = g.readGroupInfo()
	return
}

// GroupHeadSizeLimit allows you to overwrite default head size limit - 10MB.
func GroupHeadSizeLimit(limit int64) func(*Group) {
	return func(g *Group) {
		g.headSizeLimit = limit
	}
}

// GroupTotalSizeLimit allows you to overwrite default total size limit of the group - 1GB.
func GroupTotalSizeLimit(limit int64) func(*Group) {
	return func(g *Group) {
		g.totalSizeLimit = limit
	}
}

// OnStart implements service.Service by starting the goroutine that checks file
// and group limits.
func (g *Group) OnStart() error {
	return nil
}

// OnStop implements service.Service by stopping the goroutine described above.
// NOTE: g.Head must be closed separately using Close.
func (g *Group) OnStop() {
	if err := g.FlushAndSync(); err != nil {
		g.Logger.Error(
			fmt.Sprintf("unable to gracefully flush data, %s", err.Error()),
		)
	}
}

// Wait blocks until all internal goroutines are finished. Supposed to be
// called after Stop.
func (g *Group) Wait() {
	// Nothing to wait for.
}

// Close closes the head file. The group must be stopped by this moment.
func (g *Group) Close() {
	if err := g.FlushAndSync(); err != nil {
		g.Logger.Error(
			fmt.Sprintf("unable to gracefully flush data, %s", err.Error()),
		)
	}

	g.mtx.Lock()
	defer g.mtx.Unlock()

	if err := g.Head.Close(); err != nil {
		g.Logger.Error(
			fmt.Sprintf("unable to gracefully close group head, %s", err.Error()),
		)
	}
}

// HeadSizeLimit returns the current head size limit.
func (g *Group) HeadSizeLimit() int64 {
	g.mtx.Lock()
	defer g.mtx.Unlock()
	return g.headSizeLimit
}

// TotalSizeLimit returns total size limit of the group.
func (g *Group) TotalSizeLimit() int64 {
	g.mtx.Lock()
	defer g.mtx.Unlock()
	return g.totalSizeLimit
}

// MaxIndex returns index of the last file in the group.
func (g *Group) MaxIndex() int {
	g.mtx.Lock()
	defer g.mtx.Unlock()
	return g.info.MaxIndex
}

// MinIndex returns index of the first file in the group.
func (g *Group) MinIndex() int {
	g.mtx.Lock()
	defer g.mtx.Unlock()
	return g.info.MinIndex
}

func (g *Group) TotalSize() int64 {
	g.mtx.Lock()
	defer g.mtx.Unlock()
	return g.info.TotalSize
}

func (g *Group) HeadSize() int64 {
	g.mtx.Lock()
	defer g.mtx.Unlock()
	return g.info.HeadSize
}

// Write writes the contents of p into the current head of the group. It
// returns the number of bytes written. If nn < len(p), it also returns an
// error explaining why the write is short.
// NOTE: Writes are buffered so they don't write synchronously
// TODO: Make it halt if space is unavailable
func (g *Group) Write(p []byte) (nn int, err error) {
	g.mtx.Lock()
	defer g.mtx.Unlock()
	nn, err = g.headBuf.Write(p)

	// Update limits
	g.info.TotalSize += int64(nn)
	g.info.HeadSize += int64(nn)

	// Maybe rotate
	if err == nil && 0 < g.headSizeLimit && g.headSizeLimit <= g.info.HeadSize {
		g.rotateFile()
	}
	return
}

// WriteLine writes line into the current head of the group. It also appends "\n".
// NOTE: Writes are buffered so they don't write synchronously
// TODO: Make it halt if space is unavailable
func (g *Group) WriteLine(line string) error {
	g.mtx.Lock()
	defer g.mtx.Unlock()
	nn, err := g.headBuf.Write([]byte(line + "\n"))

	// Update limits
	g.info.TotalSize += int64(nn)
	g.info.HeadSize += int64(nn)

	// Maybe rotate
	if err == nil && 0 < g.headSizeLimit && g.headSizeLimit <= g.info.HeadSize {
		g.rotateFile()
	}
	return err
}

// Buffered returns the size of the currently buffered data.
func (g *Group) Buffered() int {
	g.mtx.Lock()
	defer g.mtx.Unlock()
	return g.headBuf.Buffered()
}

// FlushAndSync writes any buffered data to the underlying file and commits the
// current content of the file to stable storage (fsync).
func (g *Group) FlushAndSync() error {
	g.mtx.Lock()
	defer g.mtx.Unlock()
	err := g.headBuf.Flush()
	if err == nil {
		err = g.Head.Sync()
	}
	return err
}

func (g *Group) ensureTotalSizeLimit() {
	limit := g.totalSizeLimit
	if limit == 0 {
		return
	}

	for i := range maxFilesToRemove {
		index := g.info.MinIndex + i
		if g.info.TotalSize < limit {
			return
		}
		if index == g.info.MaxIndex {
			// Special degenerate case, just do nothing.
			// group's head may grow without bound.
			// TODO: an occasional warning?
			return
		}
		pathToRemove := filePathForIndex(g.Head.Path, index, g.info.MaxIndex)
		fInfo, err := os.Stat(pathToRemove)
		if err != nil {
			g.Logger.Error("Failed to fetch info for file", "file", pathToRemove)
			g.info.MinIndex = index + 1 // bump MinIndex.
			continue
		}
		err = os.Remove(pathToRemove)
		if err != nil {
			g.Logger.Error("Failed to remove path", "path", pathToRemove)
			return
		}
		g.info.MinIndex = index + 1 // bump MinIndex.
		g.info.TotalSize -= fInfo.Size()
	}
}

// RotateFile causes group to close the current head and assign it some index.
// After rotation, the earliest chunk may be removed if total size > totalSizeLimit.
// Note it does not create a new head.
func (g *Group) RotateFile() {
	g.mtx.Lock()
	defer g.mtx.Unlock()
	g.rotateFile()
}

func (g *Group) rotateFile() {
	headPath := g.Head.Path

	if err := g.headBuf.Flush(); err != nil {
		panic(err)
	}

	if err := g.Head.Sync(); err != nil {
		panic(err)
	}

	if err := g.Head.closeFile(); err != nil {
		panic(err)
	}

	indexPath := filePathForIndex(headPath, g.info.MaxIndex, g.info.MaxIndex+1)
	if err := os.Rename(headPath, indexPath); err != nil {
		panic(err)
	}

	g.info.HeadSize = 0
	g.info.MaxIndex++

	g.ensureTotalSizeLimit()
}

// NewReader returns a new group reader.
// If endIndex != 0, reads until endIndex exclusive.
// CONTRACT: Caller must close the returned GroupReader.
func (g *Group) NewReader(startIndex int, endIndex int) (*GroupReader, error) {
	r := newGroupReader(g, startIndex, endIndex)
	return r, nil
}

// GroupInfo holds information about the group.
type GroupInfo struct {
	MinIndex  int   // index of the first file in the group, including head
	MaxIndex  int   // index of the last file in the group, including head
	TotalSize int64 // total size of the group
	HeadSize  int64 // size of the head
}

// Returns info after scanning all files in g.Head's dir.
func (g *Group) ReadGroupInfo() GroupInfo {
	g.mtx.Lock()
	defer g.mtx.Unlock()
	return g.readGroupInfo()
}

var indexedFilePattern = regexp.MustCompile(`^.+\.([0-9]{3,})$`)

// Index includes the head.
// CONTRACT: caller should have called g.mtx.Lock
func (g *Group) readGroupInfo() GroupInfo {
	groupDir := filepath.Dir(g.Head.Path)
	headBase := filepath.Base(g.Head.Path)
	minIndex, maxIndex := -1, -1
	var totalSize, headSize int64 = 0, 0

	dir, err := os.Open(groupDir)
	if err != nil {
		panic(err)
	}
	defer dir.Close()
	fiz, err := dir.Readdir(0)
	if err != nil {
		panic(err)
	}

	// For each file in the directory, filter by pattern
	for _, fileInfo := range fiz {
		if fileInfo.Name() == headBase {
			fileSize := fileInfo.Size()
			totalSize += fileSize
			headSize = fileSize
			continue
		} else if strings.HasPrefix(fileInfo.Name(), headBase) {
			fileSize := fileInfo.Size()
			totalSize += fileSize
			submatch := indexedFilePattern.FindSubmatch([]byte(fileInfo.Name()))
			if len(submatch) != 0 {
				// Matches
				fileIndex, err := strconv.Atoi(string(submatch[1]))
				if err != nil {
					panic(err)
				}
				if maxIndex < fileIndex {
					maxIndex = fileIndex
				}
				if minIndex == -1 || fileIndex < minIndex {
					minIndex = fileIndex
				}
			}
		}
	}

	// TODO ensure that all files are present between min and max.

	// Now account for the head.
	if minIndex == -1 {
		// If there were no numbered files,
		// then the head is index 0.
		minIndex, maxIndex = 0, 0
	} else {
		// Otherwise, the head file is 1 greater
		maxIndex++
	}
	return GroupInfo{minIndex, maxIndex, totalSize, headSize}
}

func filePathForIndex(headPath string, index int, maxIndex int) string {
	if index == maxIndex {
		return headPath
	}
	return fmt.Sprintf("%v.%03d", headPath, index)
}

// --------------------------------------------------------------------------------

// GroupReader provides an interface for reading from a Group.
type GroupReader struct {
	*Group
	mtx        sync.Mutex
	startIndex int
	endIndex   int
	curIndex   int
	curFile    *os.File
	curReader  *bufio.Reader
	curLine    []byte
}

func newGroupReader(g *Group, startIndex int, endIndex int) *GroupReader {
	gr := &GroupReader{
		Group:      g,
		startIndex: startIndex,
		endIndex:   endIndex,
		curIndex:   0,
		curFile:    nil,
		curReader:  nil,
		curLine:    nil,
	}
	gr.openFile(startIndex)
	return gr
}

// Close closes the GroupReader by closing the cursor file.
func (gr *GroupReader) Close() error {
	gr.mtx.Lock()
	defer gr.mtx.Unlock()

	if gr.curReader != nil {
		err := gr.curFile.Close()
		gr.curIndex = 0
		gr.curReader = nil
		gr.curFile = nil
		gr.curLine = nil
		return err
	}
	return nil
}

// Read implements io.Reader, reading bytes from the current Reader
// incrementing index until enough bytes are read.
func (gr *GroupReader) Read(p []byte) (n int, err error) {
	lenP := len(p)
	if lenP == 0 {
		return 0, errors.New("given empty slice")
	}

	gr.mtx.Lock()
	defer gr.mtx.Unlock()

	// Open file if not open yet
	if gr.curReader == nil {
		if err = gr.openFile(gr.curIndex); err != nil {
			return 0, err
		}
	}

	// Iterate over files until enough bytes are read
	var nn int
	for {
		nn, err = gr.curReader.Read(p[n:])
		n += nn
		switch {
		case errors.Is(err, io.EOF):
			if n >= lenP {
				return n, nil
			}
			// Open the next file
			if err1 := gr.openFile(gr.curIndex + 1); err1 != nil {
				return n, err1
			}
		case err != nil:
			return n, err
		case nn == 0: // empty file
			return n, err
		}
	}
}

// IF index > gr.Group.maxIndex, returns io.EOF
// CONTRACT: caller should hold gr.mtx
func (gr *GroupReader) openFile(index int) error {
	// Lock on Group to ensure that head doesn't move in the meanwhile.
	gr.Group.mtx.Lock()
	defer gr.Group.mtx.Unlock()

	if gr.Group.info.MaxIndex < index {
		return io.EOF
	}
	if gr.endIndex != 0 && gr.endIndex <= index {
		return io.EOF
	}

	curFilePath := filePathForIndex(gr.Head.Path, index, gr.Group.info.MaxIndex)
	curFile, err := os.OpenFile(curFilePath, os.O_RDONLY|os.O_CREATE, autoFilePerms)
	if err != nil {
		return err
	}
	curReader := bufio.NewReader(curFile)

	// Update gr.cur*
	if gr.curFile != nil {
		gr.curFile.Close() // TODO return error?
	}
	gr.curIndex = index
	gr.curFile = curFile
	gr.curReader = curReader
	gr.curLine = nil
	return nil
}

// CurIndex returns cursor's file index.
func (gr *GroupReader) CurIndex() int {
	gr.mtx.Lock()
	defer gr.mtx.Unlock()
	return gr.curIndex
}
