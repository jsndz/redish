package util_test

import (
	"testing"

	"github.com/jsndz/redish/util"
)

func BenchmarkRESP_ToRESP(b *testing.B) {
	cmd := "SET user:1000:session 9876543210abcdef EX 3600"
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = util.ToRESP(cmd)
	}
}

func BenchmarkRESP_Formatter_SimpleString(b *testing.B) {
	input := "+OK\r\n"
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = util.RESPFormatter(input)
	}
}

func BenchmarkRESP_Formatter_ArrayCommand(b *testing.B) {
	input := "*3\r\n$3\r\nSET\r\n$9\r\nbenchmark\r\n$5\r\nvalue\r\n"
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = util.RESPFormatter(input)
	}
}
