# 队列

一个简单的队列,目前有待优化,大部分功能可以用chan代替

```go
q:=BlockQueueWrapper(NewQueue[int]())

q.Enqueue(1)
q.Enqueue(2)
q.Enqueue(3)
q.Dequeue()
q.Dequeue()
q.Dequeue()
```