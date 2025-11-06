package data

import (
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/choral-io/gommerce-server-core/config"
)

const (
	DefaultIDEpoch   = int64(1704067200000)        // Defaults to: 2024-01-01T00:00:00Z
	TimestampBits    = 41                          // Number of bits for timestamp
	MaxTimestamp     = int64(1)<<TimestampBits - 1 // Maximum timestamp value
	MaxPayloadBits   = 64 - 1 - TimestampBits      // Maximum bits available for cluster + worker + sequence
	ClockWaitTimeout = 2                           // Seconds to wait for clock recovery
)

var (
	defaultIDWorker atomic.Value

	ErrSeqRequired             = errors.New("seq required when workerSeqKey is set")
	ErrIDEpochOutOfRange       = errors.New("idEpoch must be positive")
	ErrClusterIDBitsOutOfRange = errors.New("clusterIDBits must be positive")
	ErrWorkerIDBitsOutOfRange  = errors.New("workerIDBits must be positive")
	ErrSequenceBitsOutOfRange  = errors.New("sequenceBits must be positive")
	ErrClusterIDOutOfRange     = errors.New("clusterID out of range")
	ErrWorkerIDOutOfRange      = errors.New("workerID out of range")
	ErrIDBitsExceedLimit       = errors.New("total ID bits exceed limit")
)

func init() {
	var idw IDWorker = &idWorker{
		idEpoch:       DefaultIDEpoch,
		clusterID:     0,
		workerID:      0,
		clusterIDBits: 5,
		workerIDBits:  5,
		sequenceBits:  12,
		sequenceMask:  4095, // int64(1)<<12 - 1
		sequenceValue: 0,
		lastTimestamp: 0,
		maxTimestamp:  DefaultIDEpoch + (int64(1)<<TimestampBits - 1),
	}
	defaultIDWorker.Store(idw)
}

// SetDefaultIDWorker sets the default IDWorker instance.
func SetDefaultIDWorker(w IDWorker) {
	defaultIDWorker.Store(w)
}

// DefaultIDWorker returns the default IDWorker instance.
func DefaultIDWorker() IDWorker {
	return defaultIDWorker.Load().(IDWorker)
}

// IDWorker is used to generate unique id.
type IDWorker interface {
	NextInt64() int64
	NextBytes() [8]byte
	NextHex() string
}

// idWorker is used to generate unique id using the Snowflake algorithm.
// The generated ID structure: [1 bit unused][41 bits timestamp][cluster bits][worker bits][sequence bits]
type idWorker struct {
	sync.Mutex
	idEpoch       int64
	clusterID     int64
	workerID      int64
	clusterIDBits int32
	workerIDBits  int32
	sequenceBits  int32
	sequenceMask  int64
	sequenceValue int64
	lastTimestamp int64
	maxTimestamp  int64
}

// getTimestamp returns the current timestamp in milliseconds.
// It handles clock backwards detection and ensures time moves forward.
func (w *idWorker) getTimestamp() int64 {
	timestamp := time.Now().UnixMilli()

	// Handle clock moved backwards first (recoverable)
	if timestamp < w.lastTimestamp {
		slog.Warn("clock moved backwards",
			slog.Int64("last.millis", w.lastTimestamp),
			slog.Int64("current.millis", timestamp))

		timer := time.NewTimer(ClockWaitTimeout * time.Second)
		defer timer.Stop()

		for timestamp <= w.lastTimestamp {
			select {
			case <-timer.C:
				slog.Error("clock did not recover within timeout, aborting id generation")
				panic("snowflake: clock moved backwards and timeout waiting for recovery")
			default:
				time.Sleep(time.Millisecond)
				timestamp = time.Now().UnixMilli()
			}
		}
	}

	// Validate timestamp overflow after clock recovery (non-recoverable)
	if timestamp > w.maxTimestamp {
		slog.Error("timestamp exceeds maximum allowed value", slog.Int64("timestamp", timestamp), slog.Int64("max", w.maxTimestamp))
		panic(fmt.Sprintf("snowflake: timestamp (%d - %d) exceeds maximum allowed value %d", timestamp, w.idEpoch, MaxTimestamp))
	}

	return timestamp
}

// NextInt64 returns the next unique id in int64.
// This method is thread-safe and handles clock backwards movement.
func (w *idWorker) NextInt64() int64 {
	w.Lock()
	defer w.Unlock()

	timestamp := w.getTimestamp()

	if w.lastTimestamp == timestamp {
		w.sequenceValue = (w.sequenceValue + 1) & w.sequenceMask
		if w.sequenceValue == 0 {
			// Sequence overflow, wait for next millisecond
			for timestamp == w.lastTimestamp {
				timestamp = w.getTimestamp()
			}
		}
	} else {
		w.sequenceValue = 0
	}
	w.lastTimestamp = timestamp
	slog.Debug("generating new snowflake id", slog.Int64("time.millis", timestamp), slog.Int64("seq.value", w.sequenceValue))
	return ((timestamp - w.idEpoch) << int64(w.clusterIDBits+w.workerIDBits+w.sequenceBits)) |
		(w.clusterID << int64(w.workerIDBits+w.sequenceBits)) |
		(w.workerID << int64(w.sequenceBits)) |
		w.sequenceValue
}

// NextBytes returns the next unique id in bytes.
func (w *idWorker) NextBytes() [8]byte {
	v := w.NextInt64()
	return [8]byte{byte(v >> 56), byte(v >> 48), byte(v >> 40), byte(v >> 32), byte(v >> 24), byte(v >> 16), byte(v >> 8), byte(v)}
}

// NextHex returns the next unique id in hex string.
func (w *idWorker) NextHex() string {
	return fmt.Sprintf("%x", w.NextBytes())
}

// NewIDWorker creates a new IDWorker instance with the given config.
// If the workerSeqKey is not empty, the workerID will be generated from the Seq.
func NewIDWorker(cfg config.SnowflakeConfig, seq Seq) (IDWorker, error) {
	idEpoch := cfg.GetIDEpoch()
	clusterID := cfg.GetClusterID()
	workerID := cfg.GetWorkerID()
	clusterIDBits := cfg.GetClusterIDBits()
	workerIDBits := cfg.GetWorkerIDBits()
	sequenceBits := cfg.GetSequenceBits()
	workerSeqKey := cfg.GetWorkerSeqKey()

	if idEpoch <= 0 {
		return nil, ErrIDEpochOutOfRange
	}
	if clusterIDBits <= 0 {
		return nil, ErrClusterIDBitsOutOfRange
	}
	if workerIDBits <= 0 {
		return nil, ErrWorkerIDBitsOutOfRange
	}
	if sequenceBits <= 0 {
		return nil, ErrSequenceBitsOutOfRange
	}
	// Snowflake format: 1 bit (unused) + 41 bits (timestamp) + remaining bits for cluster/worker/sequence
	// Total remaining bits: 64 - 1 - 41 = 22 bits
	if m := clusterIDBits + workerIDBits + sequenceBits; m > MaxPayloadBits {
		return nil, fmt.Errorf(
			"total bits (%d) exceeds maximum %d: %w",
			m, MaxPayloadBits, ErrIDBitsExceedLimit,
		)
	}
	if m := int64(1)<<clusterIDBits - 1; clusterID < 0 || clusterID > m {
		return nil, fmt.Errorf("clusterID must be 0-%d: %w", m, ErrClusterIDOutOfRange)
	}

	if workerSeqKey != "" {
		if seq == nil {
			return nil, ErrSeqRequired
		}
		if v, err := seq.Next(workerSeqKey, int64(0), int64(1)<<cfg.GetWorkerIDBits()-1); err != nil {
			return nil, err
		} else {
			workerID = v
		}
	}
	if m := int64(1)<<workerIDBits - 1; workerID < 0 || workerID > m {
		return nil, fmt.Errorf("workerID must be 0-%d: %w", m, ErrWorkerIDOutOfRange)
	}

	return &idWorker{
		idEpoch:       idEpoch,
		clusterID:     clusterID,
		workerID:      workerID,
		clusterIDBits: clusterIDBits,
		workerIDBits:  workerIDBits,
		sequenceBits:  sequenceBits,
		sequenceMask:  int64(1)<<sequenceBits - 1,
		sequenceValue: 0,
		lastTimestamp: 0,
		maxTimestamp:  idEpoch + (int64(1)<<TimestampBits - 1),
	}, nil
}
