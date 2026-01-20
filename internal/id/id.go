package id

import (
	"fmt"
	"time"

	"github.com/ruhrcloud/snowflake"
)

var node *snowflake.Node

func init() {
	var instance int64 = 1
	var err error
	node, err = snowflake.New(instance)
	if err != nil {
		panic(err)
	}

	// epoch set to 01.01.2025
	ts := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	epoch := ts.UnixMilli()
	node.SetEpoch(epoch)
}

func New() string {
	id, err := node.Generate()
	if err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}

	return fmt.Sprintf("%d", id)
}
