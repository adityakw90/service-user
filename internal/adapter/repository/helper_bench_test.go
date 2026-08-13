package repository

import "testing"

// Method 1: Current implementation using ceiling division formula
func method1(total int64, limit int) int {
	if total <= 0 || limit <= 0 {
		return 1
	}
	return int((total + int64(limit) - 1) / int64(limit))
}

// Method 2: Alternative using division + modulo check
func method2(total int64, limit int) int {
	if total <= 0 || limit <= 0 {
		return 1
	}
	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}
	return totalPages
}

// Generic benchmarks for each method
func BenchmarkPageCalculationMethod1(b *testing.B) {
	for b.Loop() {
		method1(100, 10)
	}
}

func BenchmarkPageCalculationMethod2(b *testing.B) {
	for b.Loop() {
		method2(100, 10)
	}
}

// Exact division scenario: total % limit == 0
func BenchmarkMethod1_ExactDivision(b *testing.B) {
	b.Run("100/10", func(b *testing.B) {
		for b.Loop() {
			method1(100, 10)
		}
	})
	b.Run("1000/50", func(b *testing.B) {
		for b.Loop() {
			method1(1000, 50)
		}
	})
}

func BenchmarkMethod2_ExactDivision(b *testing.B) {
	b.Run("100/10", func(b *testing.B) {
		for b.Loop() {
			method2(100, 10)
		}
	})
	b.Run("1000/50", func(b *testing.B) {
		for b.Loop() {
			method2(1000, 50)
		}
	})
}

// Remainder scenario: total % limit > 0
func BenchmarkMethod1_WithRemainder(b *testing.B) {
	b.Run("105/10", func(b *testing.B) {
		for b.Loop() {
			method1(105, 10)
		}
	})
	b.Run("1003/50", func(b *testing.B) {
		for b.Loop() {
			method1(1003, 50)
		}
	})
}

func BenchmarkMethod2_WithRemainder(b *testing.B) {
	b.Run("105/10", func(b *testing.B) {
		for b.Loop() {
			method2(105, 10)
		}
	})
	b.Run("1003/50", func(b *testing.B) {
		for b.Loop() {
			method2(1003, 50)
		}
	})
}

// Edge case: total < limit (should return 1 page)
func BenchmarkMethod1_TotalLessThanLimit(b *testing.B) {
	b.Run("5/10", func(b *testing.B) {
		for b.Loop() {
			method1(5, 10)
		}
	})
	b.Run("1/100", func(b *testing.B) {
		for b.Loop() {
			method1(1, 100)
		}
	})
}

func BenchmarkMethod2_TotalLessThanLimit(b *testing.B) {
	b.Run("5/10", func(b *testing.B) {
		for b.Loop() {
			method2(5, 10)
		}
	})
	b.Run("1/100", func(b *testing.B) {
		for b.Loop() {
			method2(1, 100)
		}
	})
}

// Edge case: limit = 1 (worst case for modulo operation)
func BenchmarkMethod1_LimitOne(b *testing.B) {
	for b.Loop() {
		method1(1000, 1)
	}
}

func BenchmarkMethod2_LimitOne(b *testing.B) {
	for b.Loop() {
		method2(1000, 1)
	}
}

// Large dataset scenario
func BenchmarkMethod1_LargeDataset(b *testing.B) {
	b.Run("1M/50", func(b *testing.B) {
		for b.Loop() {
			method1(1_000_000, 50)
		}
	})
	b.Run("10M/100", func(b *testing.B) {
		for b.Loop() {
			method1(10_000_000, 100)
		}
	})
}

func BenchmarkMethod2_LargeDataset(b *testing.B) {
	b.Run("1M/50", func(b *testing.B) {
		for b.Loop() {
			method2(1_000_000, 50)
		}
	})
	b.Run("10M/100", func(b *testing.B) {
		for b.Loop() {
			method2(10_000_000, 100)
		}
	})
}

// Zero total edge case
func BenchmarkMethod1_ZeroTotal(b *testing.B) {
	for b.Loop() {
		method1(0, 10)
	}
}

func BenchmarkMethod2_ZeroTotal(b *testing.B) {
	for b.Loop() {
		method2(0, 10)
	}
}

// Existing benchmark for buildMeta
func Benchmark_BuildMeta(b *testing.B) {
	for b.Loop() {
		buildMeta(100, 0, 10)
	}
}
