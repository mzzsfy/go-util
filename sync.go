package util

import (
    "sync"
)

func NewWaitGroup(init int) *sync.WaitGroup {
    waitGroup := sync.WaitGroup{}
    waitGroup.Add(init)
    return &waitGroup
}
