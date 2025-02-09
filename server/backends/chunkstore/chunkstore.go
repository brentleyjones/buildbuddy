package chunkstore

import (
	"context"
	"fmt"
	"io"
	"math"
	"syscall"
	"time"

	"github.com/buildbuddy-io/buildbuddy/server/interfaces"
	"github.com/buildbuddy-io/buildbuddy/server/util/status"
)

const (
	mb   = 1 << 20
	year = time.Hour * 24 * 365
)

// This implements a chunking reader/writer interface on top of an arbitrary
// blobstore, averting the need to access blobs all at once.
type Chunkstore struct {
	internalBlobstore    interfaces.Blobstore
	writeBlockSize       int
	writeTimeoutDuration time.Duration
}

type ChunkstoreOptions struct {
	WriteBlockSize       int
	WriteTimeoutDuration time.Duration
}

func New(blobstore interfaces.Blobstore, co *ChunkstoreOptions) *Chunkstore {
	writeBlockSize := co.WriteBlockSize
	if writeBlockSize == 0 {
		writeBlockSize = 1 * mb
	}
	writeTimeoutDuration := co.WriteTimeoutDuration
	if writeTimeoutDuration == 0 {
		writeTimeoutDuration = 5 * time.Second
	}
	return &Chunkstore{
		internalBlobstore:    blobstore,
		writeBlockSize:       writeBlockSize,
		writeTimeoutDuration: writeTimeoutDuration,
	}
}

func (c *Chunkstore) BlobExists(ctx context.Context, blobName string) (bool, error) {
	return c.chunkExists(ctx, blobName, 0)
}

func (c *Chunkstore) ReadBlob(ctx context.Context, blobName string) ([]byte, error) {
	data, err := io.ReadAll(c.Reader(ctx, blobName))
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (c *Chunkstore) WriteBlob(ctx context.Context, blobName string, data []byte) (int, error) {
	c.DeleteBlob(ctx, blobName)
	w := c.Writer(ctx, blobName)
	bytesWritten, err := w.Write(data)
	if err != nil {
		return bytesWritten, err
	}
	bytesFlushed, err := w.Flush()
	if err != nil {
		return bytesWritten + bytesFlushed, err
	}

	return bytesWritten + bytesFlushed, w.Close()
}

func (c *Chunkstore) DeleteBlob(ctx context.Context, blobName string) error {
	index := uint16(0)
	for {
		if exists, err := c.chunkExists(ctx, blobName, index); err != nil {
			return err
		} else if !exists {
			return nil
		}
		if err := c.deleteChunk(ctx, blobName, index); err != nil {
			return err
		}
		index++
	}
}

func chunkName(blobName string, index uint16) string {
	return blobName + "_" + fmt.Sprintf("%04x", index)
}

func (c *Chunkstore) chunkExists(ctx context.Context, blobName string, index uint16) (bool, error) {
	return c.internalBlobstore.BlobExists(ctx, chunkName(blobName, index))
}

func (c *Chunkstore) readChunk(ctx context.Context, blobName string, index uint16) ([]byte, error) {
	return c.internalBlobstore.ReadBlob(ctx, chunkName(blobName, index))
}

func (c *Chunkstore) writeChunk(ctx context.Context, blobName string, index uint16, data []byte) (int, error) {
	return c.internalBlobstore.WriteBlob(ctx, chunkName(blobName, index), data)
}

func (c *Chunkstore) deleteChunk(ctx context.Context, blobName string, index uint16) error {
	return c.internalBlobstore.DeleteBlob(ctx, chunkName(blobName, index))
}

type chunkstoreReader struct {
	ctx        context.Context
	chunkstore *Chunkstore
	blobName   string
	chunk      []byte
	off        int64
	chunkOff   int
	chunkIndex uint16
	startIndex uint16
	reverse    bool
}

func (r *chunkstoreReader) advanceOffset(adv int64) {
	r.chunkOff += int(adv)
	r.off += adv
}

func (r *chunkstoreReader) nextChunkIndex() uint16 {
	if r.reverse {
		return r.chunkIndex - 1
	}
	return r.chunkIndex + 1
}

func (r *chunkstoreReader) noChunksRead() bool {
	return r.chunkIndex == r.startIndex
}

func (r *chunkstoreReader) copyAndAdvanceOffset(dst, src []byte) int {
	bytesRead := copy(dst, src)
	r.advanceOffset(int64(bytesRead))
	return bytesRead
}

func (r *chunkstoreReader) copyToReadBuffer(p []byte) int {
	if r.chunkOff >= len(r.chunk) {
		return 0
	}
	if r.reverse {
		remainingBytesInChunk := len(r.chunk) - r.chunkOff

		// If we intend to saturate our read buffer with this copy,
		// there will be no more bytes to read after the copy and
		// there will be len(p) fewer bytes remaining in the chunk
		remainingBytesInChunkAfterCopy := remainingBytesInChunk - len(p)
		bytesToReadAfterCopy := 0
		if len(p) > remainingBytesInChunk {
			// If we intend to copy the entire rest of the chunk into the
			// read buffer, there will be no bytes remaining in the chunk
			// after the copy, and the bytes left to read will be reduced
			// by the number of bytes currently remaining in the chunk
			bytesToReadAfterCopy = len(p) - remainingBytesInChunk
			remainingBytesInChunkAfterCopy = 0
		}
		return r.copyAndAdvanceOffset(p[bytesToReadAfterCopy:], r.chunk[remainingBytesInChunkAfterCopy:remainingBytesInChunk])
	}
	return r.copyAndAdvanceOffset(p, r.chunk[r.chunkOff:])
}

func (r *chunkstoreReader) nextChunkExists() (bool, error) {
	return r.chunkstore.chunkExists(r.ctx, r.blobName, r.nextChunkIndex())
}

func (r *chunkstoreReader) getNextChunk() error {
	if exists, err := r.nextChunkExists(); err != nil {
		return err
	} else if !exists {
		return status.NotFoundErrorf("Opening %v: Couldn't find blob.", chunkName(r.blobName, r.nextChunkIndex()))
	}
	r.chunkIndex = r.nextChunkIndex()
	// Decrementing the chunk offset by the length of the chunk instead
	// of zeroing it. This is important for the Seek operation.
	r.chunkOff -= len(r.chunk)
	var err error
	r.chunk, err = r.chunkstore.readChunk(r.ctx, r.blobName, r.chunkIndex)
	return err
}

func (r *chunkstoreReader) eof() (bool, error) {
	if r.chunkOff < len(r.chunk) {
		return false, nil
	}
	exists, err := r.nextChunkExists()
	return !exists, err
}

func (r *chunkstoreReader) shiftToFront(p []byte, bytesRead *int) {
	if *bytesRead < len(p) {
		copy(p, p[(len(p)-*bytesRead):])
		for i := *bytesRead; i < len(p); i++ {
			p[i] = 0
		}
	}
}

func (r *chunkstoreReader) Read(p []byte) (int, error) {
	bytesRead := r.copyToReadBuffer(p)
	if r.reverse {
		defer r.shiftToFront(p, &bytesRead)
	}
	for bytesRead < len(p) {
		err := r.getNextChunk()
		if r.reverse {
			//Exclude bytes already read in the read buffer
			bytesRead += r.copyToReadBuffer(p[:(len(p) - bytesRead)])
		} else {
			//Exclude bytes already read in the read buffer
			bytesRead += r.copyToReadBuffer(p[bytesRead:])
		}
		if err != nil {
			if status.IsNotFoundError(err) && !r.noChunksRead() {
				return bytesRead, io.EOF
			}
			return bytesRead, err
		}
	}
	if eof, err := r.eof(); err != nil {
		return bytesRead, err
	} else if eof {
		return bytesRead, io.EOF
	}
	return bytesRead, nil
}

func (r *chunkstoreReader) Seek(offset int64, whence int) (int64, error) {
	if offset == r.off {
		return r.off, nil
	}
	switch whence {
	case io.SeekStart:
	case io.SeekCurrent:
		offset = r.off + offset
	case io.SeekEnd:
		for {
			if exists, err := r.nextChunkExists(); err != nil {
				return r.off, err
			} else if !exists {
				break
			}
			r.getNextChunk()
		}
		offset += r.off + int64(len(r.chunk)-r.chunkOff)
	default:
		return r.off, syscall.EINVAL
	}

	if offset < 0 {
		return r.off, syscall.EINVAL
	}

	// If the requested offset is within the current chunk, just change the offset and return
	if offset >= r.off-int64(r.chunkOff) && offset < r.off+int64(len(r.chunk)-r.chunkOff) {
		r.advanceOffset(offset - r.off)
		return r.off, nil
	}

	// If the offset is before the current offset (and not within the current chunk),
	// reset the offset too the beginning of the file
	if offset < r.off {
		r.chunkIndex = r.startIndex
		r.chunk = []byte{}
		r.off = 0
		r.chunkOff = 0
	}
	r.advanceOffset(offset - r.off)
	return r.off, nil
}

func (r *chunkstoreReader) Close() error {
	r.blobName = ""
	r.chunk = nil
	r.ctx = nil
	r.chunkstore = nil
	return nil
}

func (c *Chunkstore) Reader(ctx context.Context, blobName string) *chunkstoreReader {
	return &chunkstoreReader{
		chunkstore: c,
		chunkIndex: math.MaxUint16,
		startIndex: math.MaxUint16,
		chunk:      []byte{},
		blobName:   blobName,
		ctx:        ctx,
	}
}

func (c *Chunkstore) ReverseReader(ctx context.Context, blobName string) (*chunkstoreReader, error) {
	chunkIndex := uint16(0)
	for {
		if exists, err := c.chunkExists(ctx, blobName, chunkIndex); err != nil {
			return nil, err
		} else if !exists {
			break
		}
		chunkIndex++
	}
	return &chunkstoreReader{
		chunkstore: c,
		chunkIndex: chunkIndex,
		startIndex: chunkIndex,
		chunk:      []byte{},
		blobName:   blobName,
		ctx:        ctx,
		reverse:    true,
	}, nil
}

type writeResult struct {
	err  error
	size int
}

type chunkstoreWriter struct {
	chunkstore         *Chunkstore
	blobName           string
	writeChannel       chan []byte
	writeResultChannel chan writeResult
}

func (w *chunkstoreWriter) readFromWriteResultChannel() (int, error) {
	select {
	case result, open := <-w.writeResultChannel:
		if !open {
			return 0, status.UnavailableErrorf("Error accessing %v: Already closed.", w.blobName)
		}
		return result.size, result.err
	case <-time.After(w.chunkstore.writeTimeoutDuration):
		return 0, status.DeadlineExceededErrorf("Error accessing %v: Deadline exceeded.", w.blobName)
	}
}

func (w *chunkstoreWriter) Write(p []byte) (int, error) {
	w.writeChannel <- p
	return w.readFromWriteResultChannel()
}

func (w *chunkstoreWriter) Flush() (int, error) {
	w.writeChannel <- nil
	return w.readFromWriteResultChannel()
}

func (w *chunkstoreWriter) Close() error {
	close(w.writeChannel)
	_, err := w.readFromWriteResultChannel()
	return err
}

type writeLoop struct {
	chunkstore         *Chunkstore
	blobName           string
	ctx                context.Context
	chunkIndex         uint16
	chunk              []byte
	flushTime          time.Time
	writeError         error
	bytesFlushed       int
	open               bool
	timeout            bool
	writeChannel       chan []byte
	writeResultChannel chan writeResult
}

func (l *writeLoop) write() error {
	size := l.chunkstore.writeBlockSize
	if len(l.chunk) < l.chunkstore.writeBlockSize {
		size = len(l.chunk)
	}
	bytesWritten, err := l.chunkstore.writeChunk(l.ctx, l.blobName, l.chunkIndex, l.chunk[:size])
	l.bytesFlushed += bytesWritten
	l.chunk = l.chunk[bytesWritten:]
	if bytesWritten > 0 {
		l.chunkIndex++
	}
	if err != nil {
		l.writeError = err
	}
	return err
}

func (l *writeLoop) readFromWriteChannel() {
	l.timeout = false
	var p []byte
	select {
	case p, l.open = <-l.writeChannel:
		if l.open && p == nil {
			l.flushTime = time.Unix(0, 0)
		} else {
			l.chunk = append(l.chunk, p...)
		}
	case <-time.After(time.Until(l.flushTime)):
		l.timeout = true
	case <-l.ctx.Done():
		l.open = false
	}
}

func (l *writeLoop) run() {
	l.flushTime = time.Now().Add(year)
	l.open = true
	for l.open {
		l.readFromWriteChannel()
		var err error
		for err == nil && len(l.chunk) >= l.chunkstore.writeBlockSize {
			err = l.write()
		}
		if err == nil && ((!l.open && l.chunkIndex == 0) || (len(l.chunk) > 0 && (time.Now().After(l.flushTime) || !l.open))) {
			l.write()
		}
		if len(l.chunk) == 0 {
			l.flushTime = time.Now().Add(year)
		} else if time.Until(l.flushTime) > (l.chunkstore.writeTimeoutDuration) {
			l.flushTime = time.Now().Add(l.chunkstore.writeTimeoutDuration)
		}
		if l.timeout {
			continue
		}
		l.writeResultChannel <- writeResult{size: l.bytesFlushed, err: l.writeError}
		l.writeError = nil
		l.bytesFlushed = 0
	}
	close(l.writeResultChannel)
}

func (c *Chunkstore) Writer(ctx context.Context, blobName string) *chunkstoreWriter {
	writer := &chunkstoreWriter{
		chunkstore:         c,
		blobName:           blobName,
		writeChannel:       make(chan []byte),
		writeResultChannel: make(chan writeResult),
	}

	loop := &writeLoop{
		chunkstore:         c,
		ctx:                ctx,
		blobName:           blobName,
		writeChannel:       writer.writeChannel,
		writeResultChannel: writer.writeResultChannel,
	}

	go loop.run()

	return writer
}
