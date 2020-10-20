# Deep Variable Equality for Humans

This is a fork of [go-deep](https://github.com/go-test/deep) to add some functionality we required for working with our operators.

## Original function

This package provides a functions: `deep.Equal`. It's like [reflect.DeepEqual](http://golang.org/pkg/reflect/#DeepEqual) but much friendlier to humans (or any sentient being) for two reason:

* `deep.Equal` returns a map of differences in the form *path-of-difference* => *actusal difference* (the original implementation returned a list of strings)
* `deep.Equal` does not compare unexported fields (by default)

`reflect.DeepEqual` is good (like all things Golang!), but it's a game of [Hunt the Wumpus](https://en.wikipedia.org/wiki/Hunt_the_Wumpus). For large maps, slices, and structs, finding the difference is difficult.

`deep.Equal` doesn't play games with you, it lists the differences:

```go
package main_test

import (
	"testing"
	"github.com/instana/deep"
)

type T struct {
	Name    string
	Numbers []float64
}

func TestDeepEqual(t *testing.T) {
	// Can you spot the difference?
	t1 := T{
		Name:    "Isabella",
		Numbers: []float64{1.13459, 2.29343, 3.010100010},
	}
	t2 := T{
		Name:    "Isabella",
		Numbers: []float64{1.13459, 2.29843, 3.010100010},
	}

	if diff := deep.Equal(t1, t2); diff != nil {
		t.Error(diff)
	}
}
```


```
$ go test
--- FAIL: TestDeepEqual (0.00s)
        main_test.go:25: [Numbers.slice[1]: 2.29343 != 2.29843]
```

The difference is in `Numbers.slice[1]`: the two values aren't equal using Go `==`.

## Extended function

This package now provides a function: `deep.EqualSubset`.
This function does a match to verify if the left element is a subset of the right element.
It additionally allows to specify that arrays/slices are equal as long as they contain the same elements, even if not in 
the same order. 

```go
package main_test

import (
	"testing"
	"github.com/instana/deep"
)

func Example(t *testing.T) {
	type student struct {
		Name string
		Age  int
		Arr  []string
	}

	left := student{"mark", 10, []string{"same1", "different", "same2"}}
	right := student{"mark", 10, []string{"same1", "same2", "different"}}

	diff, isSame := deep.EqualSubset(left, right, nil, true, false)
}
```