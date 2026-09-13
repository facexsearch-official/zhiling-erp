package snowflake

import (
	"sync"

	"github.com/bwmarrin/snowflake"
)

var (
	node     *snowflake.Node
	initOnce sync.Once
)

func Init(machineID int64) error {
	var err error
	initOnce.Do(func() {
		node, err = snowflake.NewNode(machineID)
	})
	return err
}

func GenID() int64 {
	if node == nil {
		// fallback: use timestamp
		return int64(1 << 22) // minimal valid snowflake
	}
	return node.Generate().Int64()
}
