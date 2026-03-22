package tools

import (
	"testing"
)

// BenchmarkBase64Encode 基准测试 Base64 编码性能
func BenchmarkBase64Encode(b *testing.B) {
	input := "This is a test message for base64 encoding"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Base64encode(input)
	}
}

// BenchmarkBase64Decode 基准测试 Base64 解码性能
func BenchmarkBase64Decode(b *testing.B) {
	input := "VGhpcyBpcyBhIHRlc3QgbWVzc2FnZSBmb3IgYmFzZTY0IGVuY29kaW5n"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Base64decode(input)
	}
}

// BenchmarkRandString 基准测试随机字符串生成性能
func BenchmarkRandString(b *testing.B) {
	length := 20
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		RandString(length)
	}
}

// BenchmarkMd5 基准测试 MD5 哈希性能
func BenchmarkMd5(b *testing.B) {
	input := "test string for md5 hashing"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Md5(input)
	}
}
