// Package pool 提供高性能的对象复用基础设施
//
// 五个核心组件:
//   - GoPool: 弹性协程池, BlockQueue阻塞等待任务, 空闲超时自动退出
//   - StringPool: 字符串<->ID映射池, 引用计数管理生命周期
//   - BufferPool: bytes.Buffer 对象池, 容量上限保护
//   - BytePool: 轻量字节缓冲池, 自定义Bytes类型
//   - ObjectPool: 泛型对象池, 支持自定义创建和重置函数
package pool
