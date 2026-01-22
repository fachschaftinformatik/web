package id

import (
	"encoding/json"
	"strconv"
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

// ID is a custom type for Snowflake IDs that serializes to/from strings in JSON.
type ID int64

func New() ID {
	id, err := node.Generate()
	if err != nil {
		return ID(time.Now().UnixNano())
	}

	return ID(id)
}

func (id ID) String() string {
	return strconv.FormatInt(int64(id), 10)
}

func (id ID) Int64() int64 {
	return int64(id)
}

func Parse(s string) (ID, error) {
	val, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, err
	}
	return ID(val), nil
}

func (id ID) MarshalJSON() ([]byte, error) {
	return json.Marshal(strconv.FormatInt(int64(id), 10))
}

func (id *ID) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		// Try unmarshaling as int if string fails
		var i int64
		if err := json.Unmarshal(b, &i); err != nil {
			return err
		}
		*id = ID(i)
		return nil
	}

	val, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return err
	}
	*id = ID(val)
	return nil
}
