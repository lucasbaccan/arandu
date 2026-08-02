package ids

import (
	"sync"
	"time"
)

const (
	epoch        int64 = 1704067200000 // 2024-01-01T00:00:00Z
	nodeBits     uint  = 10
	sequenceBits uint  = 12
	maxSequence  int64 = 1<<sequenceBits - 1
)

// Generator produz Snowflake IDs (41 bits timestamp + 10 bits node + 12 bits sequência).
type Generator struct {
	mu         sync.Mutex
	nodeID     int64
	sequence   int64
	lastMillis int64
}

func NewGenerator(nodeID int64) *Generator {
	return &Generator{nodeID: nodeID}
}

func (g *Generator) NextID() int64 {
	g.mu.Lock()
	defer g.mu.Unlock()

	now := time.Now().UnixMilli()
	if now < g.lastMillis {
		now = g.lastMillis
	}
	if now == g.lastMillis {
		g.sequence = (g.sequence + 1) & maxSequence
		if g.sequence == 0 {
			for now <= g.lastMillis {
				now = time.Now().UnixMilli()
			}
		}
	} else {
		g.sequence = 0
	}
	g.lastMillis = now

	return (now-epoch)<<(nodeBits+sequenceBits) | (g.nodeID << sequenceBits) | g.sequence
}
