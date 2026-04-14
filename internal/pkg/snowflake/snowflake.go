// Package snowflake 封装 bwmarrin/snowflake 节点初始化。
package snowflake

import (
	"fmt"
	"sync"

	sf "github.com/bwmarrin/snowflake"
)

var node *sf.Node
var once sync.Once

// Init 使用节点号（0-1023）初始化。
func Init(nodeID int64) error {
	var err error
	once.Do(func() {
		node, err = sf.NewNode(nodeID)
	})
	return err
}

// NextID 生成下一个 ID；未初始化时返回错误。
func NextID() (int64, error) {
	if node == nil {
		return 0, fmt.Errorf("snowflake: not initialized")
	}
	return node.Generate().Int64(), nil
}
