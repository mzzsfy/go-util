# 队列

```go
q:=BlockQueueWrapper(NewQueue[int]())

q.Enqueue(1)
q.Enqueue(2)
q.Enqueue(3)
q.Dequeue()
q.Dequeue()
q.Dequeue()
```