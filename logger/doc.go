// Package logger 提供高性能日志库, 双版本 API + 可插拔格式化
//
// 核心特性:
//   - 双版本 API: 性能版链式 (零分配) + 便捷版命令式 (D/I/L/DF/IF/LF)
//   - 类型安全: Str/Int/Bool 等方法避免 interface 装箱
//   - Formatter 接口: ConsoleFormatter (默认, 控制台易读) / JSONFormatter (JSON 行)
//   - 级别过滤零开销: 不达级别返回 disabled Event, 所有方法 noop
//
// 性能版 (链式, 零分配):
//
//	log := logger.New("app")
//	log.Info().Str("user", "moke").Int("id", 42).Msg("login")
//	// 控制台: 26-08-05 14:30:00.123[  app]I: user=moke id=42 login
//
// 便捷版 (命令式, 有装箱开销, 适合非热路径):
//
//	log.I("login", "user", "moke", "id", 42)
//	log.IF("detail", func() []any { return heavyArgs() }) // 级别不够时 f 不执行
//
// 切换 JSON 格式:
//
//	logger.SetFormatter(logger.JSONFormatter{})
//	log.Info().Str("user", "moke").Msg("login")
//	// {"t":"2026-08-05T14:30:00.123","lv":"Info","logger":"app","user":"moke","msg":"login"}
//
// 级别过滤零开销:
//
//	log.Debug().Str("k","v").Msg("filtered") // 级别不够, 返回 disabled Event
package logger
