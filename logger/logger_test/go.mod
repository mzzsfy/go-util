module logger_test

go 1.18

replace github.com/mzzsfy/go-util => ../../

require (
	github.com/mzzsfy/go-util v0.0.0-00010101000000-000000000000
	go.uber.org/zap v1.27.1
)

require go.uber.org/multierr v1.11.0 // indirect
