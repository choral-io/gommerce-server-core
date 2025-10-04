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
	DefaultIdEpoch   = int64(1704067200000)   // Defaults to: 2024-01-01T00:00:00Z
	TimestampBits    = 41                     // Number of bits for timestamp
	MaxPayloadBits   = 64 - 1 - TimestampBits // Maximum bits available for cluster + worker + sequence
	ClockWaitTimeout = 2                      // Seconds to wait for clock recovery
)

var (
	defaultIdWorker atomic.Value

	ErrSeqRequired             = errors.New("seq required when workerSeqKey is set")
	ErrIdEpochOutOfRange       = errors.New("idEpoch must be positive")
	ErrClusterIdBitsOutOfRange = errors.New("clusterIdBits must be positive")
	ErrWorkerIdBitsOutOfRange  = errors.New("workerIdBits must be positive")
	ErrSequenceBitsOutOfRange  = errors.New("sequenceBits must be positive")
	ErrClusterIdOutOfRange     = errors.New("clusterId out of range")
	ErrWorkerIdOutOfRange      = errors.New("workerId out of range")
	ErrIdBitsExceedLimit       = errors.New("total id bits exceed limit")
)

func init() {
	var idw IdWorker = &idWorker{
		idEpoch:       DefaultIdEpoch,
		clusterId:     0,
		workerId:      0,
		clusterIdBits: 5,
		workerIdBits:  5,
		sequenceBits:  12,
		sequenceMask:  4095, // int64(1)<<12 - 1
		sequenceValue: 0,
		lastMillis:    0,
	}
	defaultIdWorker.Store(idw)
}

// SetDefaultIdWorker sets the default IdWorker instance.
func SetDefaultIdWorker(w IdWorker) {
	defaultIdWorker.Store(w)
}

// DefaultIdWorker returns the default IdWorker instance.
func DefaultIdWorker() IdWorker {
	return defaultIdWorker.Load().(IdWorker)
}

// IdWorker is used to generate unique id.
type IdWorker interface {
	NextInt64() int64
	NextBytes() [8]byte
	NextHex() string
}

// idWorker is used to generate unique id using the Snowflake algorithm.
// The generated ID structure: [1 bit unused][41 bits timestamp][cluster bits][worker bits][sequence bits]
type idWorker struct {
	sync.Mutex
	idEpoch       int64
	clusterId     int64
	workerId      int64
	clusterIdBits int32
	workerIdBits  int32
	sequenceBits  int32
	sequenceMask  int64
	sequenceValue int64
	lastMillis    int64
}

// getNextMillis returns the next valid millisecond timestamp.
// It handles clock backwards detection and ensures time moves forward.
func (w *idWorker) getNextMillis() int64 {
	nextMillis := time.Now().UnixMilli()

	// Handle clock moved backwards
	if nextMillis < w.lastMillis {
		slog.Warn("clock moved backwards",
			slog.Int64("last.millis", w.lastMillis),
			slog.Int64("current.millis", nextMillis))

		// Wait for clock to catch up or timeout
		timeout := time.After(ClockWaitTimeout * time.Second)
		for nextMillis <= w.lastMillis {
			select {
			case <-timeout:
				slog.Error("clock did not recover within timeout, aborting id generation")
				panic("snowflake: clock moved backwards and timeout waiting for recovery")
			default:
				time.Sleep(time.Millisecond)
				nextMillis = time.Now().UnixMilli()
			}
		}
	}

	return nextMillis
}

// NextInt64 returns the next unique id in int64.
// This method is thread-safe and handles clock backwards movement.
func (w *idWorker) NextInt64() int64 {
	w.Lock()
	defer w.Unlock()

	nextMillis := w.getNextMillis()

	if w.lastMillis == nextMillis {
		w.sequenceValue = (w.sequenceValue + 1) & w.sequenceMask
		if w.sequenceValue == 0 {
			// Sequence overflow, wait for next millisecond
			for nextMillis == w.lastMillis {
				nextMillis = w.getNextMillis()
			}
		}
	} else {
		w.sequenceValue = 0
	}
	w.lastMillis = nextMillis
	slog.Debug("generating new snowflake id", slog.Int64("time.millis", nextMillis), slog.Int64("seq.value", w.sequenceValue))
	return ((nextMillis - w.idEpoch) << int64(w.clusterIdBits+w.workerIdBits+w.sequenceBits)) |
		(w.clusterId << int64(w.workerIdBits+w.sequenceBits)) |
		(w.workerId << int64(w.sequenceBits)) |
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

// NewIdWorker creates a new IdWorker instance with the given config.
// If the workerSeqKey is not empty, the workerId will be generated from the Seq.
func NewIdWorker(cfg config.SnowflakeConfig, seq Seq) (IdWorker, error) {
	idEpoch := cfg.GetIdEpoch()
	clusterId := cfg.GetClusterId()
	workerId := cfg.GetWorkerId()
	clusterIdBits := cfg.GetClusterIdBits()
	workerIdBits := cfg.GetWorkerIdBits()
	sequenceBits := cfg.GetSequenceBits()

	workerSeqKey := cfg.GetWorkerSeqKey()
	if workerSeqKey != "" {
		if seq == nil {
			return nil, ErrSeqRequired
		}
		if v, err := seq.Next(workerSeqKey, int64(0), int64(1)<<cfg.GetWorkerIdBits()-1); err != nil {
			return nil, err
		} else {
			workerId = v
		}
	}

	if idEpoch < 0 {
		return nil, ErrIdEpochOutOfRange
	}
	if clusterIdBits <= 0 {
		return nil, ErrClusterIdBitsOutOfRange
	}
	if workerIdBits <= 0 {
		return nil, ErrWorkerIdBitsOutOfRange
	}
	if sequenceBits <= 0 {
		return nil, ErrSequenceBitsOutOfRange
	}
	// Snowflake format: 1 bit (unused) + 41 bits (timestamp) + remaining bits for cluster/worker/sequence
	// Total remaining bits: 64 - 1 - 41 = 22 bits
	if m := clusterIdBits + workerIdBits + sequenceBits; m > MaxPayloadBits {
		return nil, fmt.Errorf(
			"total bits (%d) exceeds maximum %d: %w",
			m, MaxPayloadBits, ErrIdBitsExceedLimit,
		)
	}
	if m := int64(1)<<clusterIdBits - 1; clusterId < 0 || clusterId > m {
		return nil, fmt.Errorf("clusterId must be 0-%d: %w", m, ErrClusterIdOutOfRange)
	}
	if m := int64(1)<<workerIdBits - 1; workerId < 0 || workerId > m {
		return nil, fmt.Errorf("workerId must be 0-%d: %w", m, ErrWorkerIdOutOfRange)
	}

	return &idWorker{
		idEpoch:       idEpoch,
		clusterId:     clusterId,
		workerId:      workerId,
		clusterIdBits: clusterIdBits,
		workerIdBits:  workerIdBits,
		sequenceBits:  sequenceBits,
		sequenceMask:  int64(1)<<sequenceBits - 1,
		sequenceValue: 0,
		lastMillis:    0,
	}, nil
}
