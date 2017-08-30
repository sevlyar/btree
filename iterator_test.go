package btree

import (
	"runtime"
	"strconv"
	"testing"
)

func TestIterator_Next(t *testing.T) {
	tr := New(4)
	for i := Int(0); i < 100; i++ {
		tr.ReplaceOrInsert(i)
	}
	iter := NewIterator(tr)
	for i := Int(0); i < 100; i++ {
		if !iter.Next(Int(99)) {
			t.Fatalf("Next should return true (%d)", i)
		}
		if iter.Item() != i {
			t.Fatalf("Item returns %d instead of %d", iter.Item(), i)
		}
	}
	if iter.Next(Int(100)) {
		t.Fatal("Next should return false at the end")
	}
	for i := 0; i < 100; i++ {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			iter := NewIterator(tr)
			var last Int
			for iter.Next(Int(i)) {
				last = iter.Item().(Int)
			}
			if last != Int(i) {
				t.Fatal("should return the last value", i, iter.Item())
			}
		})
	}
}

func TestIterator_Find(t *testing.T) {
	for i := -1; i <= 100; i++ {
		t.Run(strconv.Itoa(i), func(t *testing.T) { testIteratorFind(t, Int(i)) })
	}
}

func testIteratorFind(t *testing.T, value Int) {
	tr := New(4)
	for i := Int(0); i < 100; i++ {
		tr.ReplaceOrInsert(i)
	}
	iter := NewIterator(tr)
	iter.SkipLess(value)
	var next = value
	if next < 0 {
		next = 0
	}
	hasNext := iter.Next(Int(100))
	if hasNext != (value < 100) {
		t.Error("Next should return true if end is not reach", value)
	}
	if hasNext && iter.Item() != next {
		t.Error("SkipLess should stop on searched element", value, iter.Item())
		return
	}
}

func BenchmarkIterator(b *testing.B) {
	const a = 500000
	const l = 100
	from := Item(Int(a))
	to := Item(Int(a + l))
	toWalk := Item(Int(a + l + 1))
	const count = 10 * 1000 * 1000
	series := []int{16, 32, 64, 128} // 32 is optimal
	for _, n := range series {
		runtime.GC()
		tr := New(n)
		for i := 0; i < count; i++ {
			tr.ReplaceOrInsert(Int(i))
		}
		nstr := strconv.Itoa(n)
		b.Run("walk"+nstr, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				tr.AscendRange(from, toWalk, func(Item) bool {
					return true
				})
			}
		})
		b.Run("iterator"+nstr, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				iter := NewIterator(tr)
				iter.SkipLess(from)
				for iter.Next(to) {
					iter.Item()
				}
				iter.Close()
			}
		})
	}
}
