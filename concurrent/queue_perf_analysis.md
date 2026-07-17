# segQueue MPMC 性能瓶颈分析

## 基准数据 (2026-05-06, i5-8500 6核, Go 1.25)

| 指标 | seg | ring | 差距 |
|------|-----|------|------|
| 1P1C ns/op | 20.7 | 19.0 | 1.09x |
| PingPong ns/op | 458~837 | 271~302 | 1.5~2.8x |
| MPMC ns/op | 117~132 | 54~58 | **2.0~2.3x** |
| Scale 1G ns/op | 121~134 | 53~60 | 2.3~2.5x |
| Scale 8G ns/op | 130~140 | 53~55 | 2.4~2.6x |

1P1C仅慢9%,MPMC慢2x+,差距随并发竞争放大.

---

## pprof CPU Profile 分析 (MPMC, 6核)

### 时间分布

| 函数 | flat | 占比 | 说明 |
|------|------|------|------|
| runtime.stdcall2 | 11.39s | 34.8% | Windows WaitForSingleObject (调度器互斥) |
| runtime.procyield | 4.73s | 14.4% | CAS自旋等待 |
| segQueue.Dequeue | 4.76s | 14.5% | 核心出队逻辑 |
| segQueue.Enqueue | 1.46s | 4.5% | 核心入队逻辑 |
| runtime.lock2 (cum) | 18.51s | 56.5% | Go调度器内部互斥竞争 |

### segQueue.Dequeue 热指令 (MPMC disassembly)

| 指令 | 耗时 | 对应代码 |
|------|------|---------|
| LOCK CMPXCHGQ (CAS headPos) | 1.69s | `atomic.CompareAndSwapUint64(&q.headPos, h, h+1)` |
| MOVQ 0x80(AX),SI (load headSeg) | 940ms | `atomic.LoadPointer(&q.headSeg)` |
| XCHGQ (store seq) | 620ms | `atomic.StoreUint64(&s.seq, h+segSize)` |

### ringQueue.Dequeue 热指令 (对比)

| 指令 | 耗时 | 对应代码 |
|------|------|---------|
| LOCK CMPXCHGQ (CAS head) | 320ms | `atomic.CompareAndSwapUint64(&q.head, h, h+1)` |
| MOVQ 0x8(AX),SI (load head) | 70ms | `atomic.LoadUint64(&q.head)` |

---

## 三大瓶颈

### 瓶颈一: 调度器竞争 (56.5% cumulative)

MPMC下Go调度器内部互斥开销占56%,这是所有MPMC队列共享的基础开销.
但seg的更长临界区(CAS前有更多指令)放大了调度器竞争:

seg Dequeue CAS前: load headPos → load headSeg → 比较seg.id → 计算slot地址 → load seq → CAS
ring Dequeue CAS前: load head → 计算slot地址 → load seq → CAS

seg多了2步(headSeg load + seg.id比较),CAS窗口更大,冲突率更高.

### 瓶颈二: headSeg间接寻址 (940ms vs 70ms, 13.4x)

seg需要`atomic.LoadPointer(&q.headSeg)`加载segment指针再解引用.
ring直接通过`slots[h&mask]`索引,无需间接寻址.
这是动态扩容vs预分配连续内存的根本差异.

### 瓶颈三: CAS竞争放大 (1.69s vs 0.32s, 5.3x)

由于瓶颈一和二的叠加,seg的CAS冲突率远高于ring.
每次CAS失败意味着一个完整的CAS循环重试,增加调度器压力.

---

## 已尝试的优化及结果

| 优化 | 方案 | 结果 | 原因 |
|------|------|------|------|
| P0: 合并cache line | headPos与headSeg放同一cache line | **退化** | advanceHead写headSeg导致headPos false sharing |
| P2: 延迟advanceHead | 落后2+segment才推进headSeg | 无改善 | advanceHead本身开销小,主要瓶颈在CAS和调度器 |

### 为什么合并cache line退化?

原始布局(headPos和headSeg分离):
- Dequeue读headPos + 读headSeg: 2次cache line加载,但无写冲突
- advanceHead写headSeg: 不影响headPos所在cache line

合并布局(headPos和headSeg同一cache line):
- advanceHead的CAS headSeg使整个cache line失效
- 所有正在读headPos的Dequeue goroutine都需重新加载cache line
- false sharing > 额外的cache miss

---

## 不可消除的固有限制

1. **间接寻址**: seg通过指针链访问segment,ring通过数组索引. 这是"动态扩容"vs"预分配"的根本权衡.
2. **更大的临界区**: seg在CAS前需要segment定位,增加了被抢占的窗口.
3. **segment管理元数据**: seg.id比较,链表遍历,segment分配/回收,ring完全不需要.

---

## 结论

seg在1P1C下接近ring(仅慢9%),说明单线程路径效率已经很好.
MPMC的2x差距主要来自:

1. Go调度器竞争(56%) — 无法优化,Go运行时层面的开销
2. headSeg间接寻址(940ms) — 动态扩容的固有代价
3. CAS窗口更大(1.69s vs 0.32s) — 算法层面无法缩短

**优化空间已基本耗尽.** 进一步改进需要根本性的算法变更(如per-CPU队列+work stealing),而非微优化.
