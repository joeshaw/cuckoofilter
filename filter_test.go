package cuckoofilter

import (
	"encoding/binary"
	"fmt"
	"testing"
)

func TestXor(t *testing.T) {
	f := New(1000000)
	h := hash([]byte("foo"))
	fp := f.fingerprint(uint32(h))
	i1 := f.bucketIndex(uint32(h >> 32))
	i2 := f.alternateIndex(i1, fp)

	fmt.Println("f = fingerprint(x):", fp)
	fmt.Println("i1 = hash(x):", i1)
	fmt.Println("i2 = i1 XOR hash(f)", i2)
	fmt.Println("i1 = i2 XOR hash(f)", f.alternateIndex(i2, fp))

	if actual := f.alternateIndex(i2, fp); actual != i1 {
		t.Fatalf("expected %d, got %d", i1, actual)
	}
}

func key(i int) []byte {
	d := make([]byte, 8)
	binary.LittleEndian.PutUint64(d, uint64(i))
	return d
}

func TestFilter(t *testing.T) {
	//rand.Seed(time.Now().UnixNano())
	n := 100000
	f := New(uint32(n))
	for i := 0; i < n; i++ {
		err := f.Add(key(i))
		if err != nil {
			t.Fatalf("%d: expected nil, got %s", i, err)
		}
	}

	for i := 0; i < n; i++ {
		if !f.Contains(key(i)) {
			t.Fatalf("%d: expected true, got false", i)
		}
	}

	falseCount := 0
	for i := n; i < n*2; i++ {
		if f.Contains(key(i)) {
			falseCount++
		}
	}
	fmt.Printf("False positive rate (before deletes): %d / %d\n", falseCount, n)

	// Remove half the keys, make sure the still-existing ones
	// always return true for Contains

	for i := 0; i < n/2; i++ {
		f.Delete(key(i))
	}

	for i := n / 2; i < n; i++ {
		if !f.Contains(key(i)) {
			t.Fatalf("%d: expected true, got false", i)
		}
	}

	falseCount = 0
	for i := n; i < n*2; i++ {
		if f.Contains(key(i)) {
			falseCount++
		}
	}
	fmt.Printf("False positive rate (after deletes): %d / %d\n", falseCount, n)
}

func benchmarkNew(b *testing.B, maxKeys uint32) {
	for i := 0; i < b.N; i++ {
		New(maxKeys)
	}
}

func BenchmarkNew1(b *testing.B)     { benchmarkNew(b, 1) }
func BenchmarkNew10(b *testing.B)    { benchmarkNew(b, 10) }
func BenchmarkNew100(b *testing.B)   { benchmarkNew(b, 100) }
func BenchmarkNew1000(b *testing.B)  { benchmarkNew(b, 1000) }
func BenchmarkNew10000(b *testing.B) { benchmarkNew(b, 10000) }

func BenchmarkAdd(b *testing.B) {
	f := New(uint32(b.N))
	for i := 0; i < b.N; i++ {
		f.Add(key(i))
	}
}

func BenchmarkContains(b *testing.B) {
	f := New(uint32(b.N))
	for i := 0; i < b.N; i++ {
		f.Add(key(i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f.Contains(key(i))
	}
}

func BenchmarkDelete(b *testing.B) {
	f := New(uint32(b.N))
	for i := 0; i < b.N; i++ {
		f.Add(key(i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f.Delete(key(i))
	}
}
