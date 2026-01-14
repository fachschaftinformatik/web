package sid

import (
	"fmt"
	"time"

	"github.com/ruhrcloud/snowflake"
)

var node *snowflake.Node

func init() {
	var err error
	// Using Node ID 1. In a distributed system, this would be unique per instance.
	node, err = snowflake.New(1)
	if err != nil {
		panic(err)
	}

	// User requested epoch: 23.01.2026
	epoch := time.Date(2026, 1, 23, 0, 0, 0, 0, time.UTC).UnixMilli()
	node.SetEpoch(epoch)
}

// New generates a new snowflake ID as a string.
func New() string {
	id, err := node.Generate()
	if err != nil {
		// Fallback for safety
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return fmt.Sprintf("%d", id)
}
