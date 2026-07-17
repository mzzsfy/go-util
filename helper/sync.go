package helper

import (
	"sync"
)

// NewWaitGroup 创建一个带有初始计数器的WaitGroup
// init参数为初始计数器值,用于在创建时预设等待数量
func NewWaitGroup(init int) *sync.WaitGroup {
	waitGroup := sync.WaitGroup{}
	waitGroup.Add(init)
	return &waitGroup
}
