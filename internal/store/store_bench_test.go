package store_test

import (
	"fmt"
	"math/rand"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jsndz/redish/internal/store"
)

func BenchmarkStore_Set(b *testing.B) {
	s := store.New()
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		s.Set(fmt.Sprintf("key_%d", i), "benchmark_value", 0)
	}
}

func BenchmarkStore_Get(b *testing.B) {
	s := store.New()
	const numKeys = 10000
	for i := 0; i < numKeys; i++ {
		s.Set(fmt.Sprintf("key_%d", i), "benchmark_value", 0)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		s.Get(fmt.Sprintf("key_%d", i%numKeys))
	}
}

func BenchmarkStore_Incr(b *testing.B) {
	s := store.New()
	s.Set("counter", "0", 0)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = s.Incr("counter")
	}
}

func BenchmarkStore_Rpush(b *testing.B) {
	s := store.New()
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = s.Rpush("list_key", "elem")
	}
}

func BenchmarkStore_Lrange(b *testing.B) {
	s := store.New()
	for i := 0; i < 100; i++ {
		_, _ = s.Rpush("list_key", fmt.Sprintf("elem_%d", i))
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = s.Lrange("list_key", 0, 50)
	}
}

func BenchmarkStore_ParallelGet(b *testing.B) {
	s := store.New()
	const numKeys = 10000
	for i := 0; i < numKeys; i++ {
		s.Set(fmt.Sprintf("key_%d", i), "benchmark_value", 0)
	}

	var counter uint64
	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			idx := atomic.AddUint64(&counter, 1) % numKeys
			s.Get(fmt.Sprintf("key_%d", idx))
		}
	})
}

func BenchmarkStore_ParallelSet(b *testing.B) {
	s := store.New()
	var counter uint64
	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			idx := atomic.AddUint64(&counter, 1)
			s.Set(fmt.Sprintf("key_%d", idx%10000), "val", 0)
		}
	})
}

func BenchmarkStore_ParallelMixed_80Read_20Write(b *testing.B) {
	s := store.New()
	const numKeys = 10000
	for i := 0; i < numKeys; i++ {
		s.Set(fmt.Sprintf("key_%d", i), "initial_value", 0)
	}

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		for pb.Next() {
			key := fmt.Sprintf("key_%d", r.Intn(numKeys))
			if r.Intn(100) < 80 {
				s.Get(key)
			} else {
				s.Set(key, "updated_value", 0)
			}
		}
	})
}
