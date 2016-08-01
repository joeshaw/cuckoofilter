## cuckoofilter

This is an implementation of a data structure known as a **cuckoo
filter**.  The data structure is described in a paper called *[Cuckoo
Filter: Practically Better Than
Bloom](https://www.cs.cmu.edu/~dga/papers/cuckoo-conext2014.pdf)* by
Bin Fan, David G. Andersen, Michael Kaminsky and Michael
D. Mitzenmacher.

Cuckoo filters, like Bloom filters, are probabilistic data structures
useful for determining whether a piece of data is present in a set.
Like Bloom filters, cuckoo filters do not store the key being looked
up, or the value of the data, so they are appropriate only for
checking whether the primary data source should be queried.

Cuckoo filters (and Bloom filters) can return false positives when
checking for presence, but will never return a false negative.

Unlike (standard, non-counting) Bloom filters, data can be deleted
from cuckoo filters.

To use,

```go
maxKeys := uint32(1000000)
f := cuckoofilter.New(maxKeys)

f.Add([]byte("hello"))
f.Add([]byte("world"))

f.Contains("hello") // => true
f.Contains("earth") // => true (if false positive) or false

f.Delete("world")
```

That's pretty much it.  API docs are available [on
GoDoc](https://godoc.org/github.com/joeshaw/cuckoofilter).
