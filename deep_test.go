package deep_test

import (
	"errors"
	"fmt"
	"github.com/instana/deep"
	"reflect"
	"testing"
	"time"
	v1 "github.com/instana/deep/test/v1"
	v2 "github.com/instana/deep/test/v2"

)

func TestString(t *testing.T) {
	diff := deep.Equal("foo", "foo")
	if len(diff) > 0 {
		t.Error("should be equal:", diff)
	}

	diff = deep.Equal("foo", "bar")
	if diff == nil {
		t.Fatal("no diff")
	}
	if len(diff) != 1 {
		t.Error("too many diff:", diff)
	}
	if diff[deep.Root] != "foo != bar" {
		t.Error("wrong diff:", diff[deep.Root])
	}
}

func TestFloat(t *testing.T) {
	diff := deep.Equal(1.1, 1.1)
	if len(diff) > 0 {
		t.Error("should be equal:", diff)
	}

	diff = deep.Equal(1.1234561, 1.1234562)
	if diff == nil {
		t.Error("no diff")
	}

	defaultFloatPrecision := deep.FloatPrecision
	deep.FloatPrecision = 6
	defer func() { deep.FloatPrecision = defaultFloatPrecision }()

	diff = deep.Equal(1.1234561, 1.1234562)
	if len(diff) > 0 {
		t.Error("should be equal:", diff)
	}

	diff = deep.Equal(1.123456, 1.123457)
	if diff == nil {
		t.Fatal("no diff")
	}
	if len(diff) != 1 {
		t.Error("too many diff:", diff)
	}
	if diff[deep.Root] != "1.123456 != 1.123457" {
		t.Error("wrong diff:", diff[deep.Root])
	}

}

func TestInt(t *testing.T) {
	diff := deep.Equal(1, 1)
	if len(diff) > 0 {
		t.Error("should be equal:", diff)
	}

	diff = deep.Equal(1, 2)
	if diff == nil {
		t.Fatal("no diff")
	}
	if len(diff) != 1 {
		t.Error("too many diff:", diff)
	}
	if diff[deep.Root] != "1 != 2" {
		t.Error("wrong diff:", diff[deep.Root])
	}
}

func TestUint(t *testing.T) {
	diff := deep.Equal(uint(2), uint(2))
	if len(diff) > 0 {
		t.Error("should be equal:", diff)
	}

	diff = deep.Equal(uint(2), uint(3))
	if diff == nil {
		t.Fatal("no diff")
	}
	if len(diff) != 1 {
		t.Error("too many diff:", diff)
	}
	if diff[deep.Root] != "2 != 3" {
		t.Error("wrong diff:", diff[deep.Root])
	}
}

func TestBool(t *testing.T) {
	diff := deep.Equal(true, true)
	if len(diff) > 0 {
		t.Error("should be equal:", diff)
	}

	diff = deep.Equal(false, false)
	if len(diff) > 0 {
		t.Error("should be equal:", diff)
	}

	diff = deep.Equal(true, false)
	if diff == nil {
		t.Fatal("no diff")
	}
	if len(diff) != 1 {
		t.Error("too many diff:", diff)
	}
	if diff[deep.Root] != "true != false" { // unless you're fipar
		t.Error("wrong diff:", diff[deep.Root])
	}
}

func TestKindMismatch(t *testing.T) {
	deep.LogErrors = true

	var x int = 100
	var y float64 = 100
	diff := deep.Equal(x, y)
	if diff == nil {
		t.Fatal("no diff")
	}
	if len(diff) != 1 {
		t.Error("too many diff:", diff)
	}
	if diff[deep.Root] != "int != float64" {
		t.Error("wrong diff:", diff[deep.Root])
	}

	deep.LogErrors = false
}

func TestDeepRecursion(t *testing.T) {
	deep.MaxDepth = 2
	defer func() { deep.MaxDepth = 10 }()

	type s3 struct {
		S int
	}
	type s2 struct {
		S s3
	}
	type s1 struct {
		S s2
	}
	foo := map[string]s1{
		"foo": { // 1
			S: s2{ // 2
				S: s3{ // 3
					S: 42, // 4
				},
			},
		},
	}
	bar := map[string]s1{
		"foo": {
			S: s2{
				S: s3{
					S: 100,
				},
			},
		},
	}
	// No diffs because MaxDepth=2 prevents seeing the diff at 3rd level down
	diff := deep.Equal(foo, bar)
	if diff != nil {
		t.Errorf("got %d diffs, expected none: %v", len(diff), diff)
	}

	defaultMaxDepth := deep.MaxDepth
	deep.MaxDepth = 4
	defer func() { deep.MaxDepth = defaultMaxDepth }()

	diff = deep.Equal(foo, bar)
	if diff == nil {
		t.Fatal("no diff")
	}
	if len(diff) != 1 {
		t.Error("too many diff:", diff)
	}
	if diff["map[foo].S.S.S"] != "42 != 100" {
		t.Error("wrong diff:", diff["map[foo].S.S.S"])
	}
}

func TestMaxDiff(t *testing.T) {
	a := []int{1, 2, 3, 4, 5, 6, 7}
	b := []int{0, 0, 0, 0, 0, 0, 0}

	defaultMaxDiff := deep.MaxDiff
	deep.MaxDiff = 3
	defer func() { deep.MaxDiff = defaultMaxDiff }()

	diff := deep.Equal(a, b)
	if diff == nil {
		t.Fatal("no diffs")
	}
	if len(diff) != deep.MaxDiff {
		t.Errorf("got %d diffs, expected %d", len(diff), deep.MaxDiff)
	}

	defaultCompareUnexportedFields := deep.CompareUnexportedFields
	deep.CompareUnexportedFields = true
	defer func() { deep.CompareUnexportedFields = defaultCompareUnexportedFields }()
	type fiveFields struct {
		a int // unexported fields require ^
		b int
		c int
		d int
		e int
	}
	t1 := fiveFields{1, 2, 3, 4, 5}
	t2 := fiveFields{0, 0, 0, 0, 0}
	diff = deep.Equal(t1, t2)
	if diff == nil {
		t.Fatal("no diffs")
	}
	if len(diff) != deep.MaxDiff {
		t.Errorf("got %d diffs, expected %d", len(diff), deep.MaxDiff)
	}

	// Same keys, too many diffs
	m1 := map[int]int{
		1: 1,
		2: 2,
		3: 3,
		4: 4,
		5: 5,
	}
	m2 := map[int]int{
		1: 0,
		2: 0,
		3: 0,
		4: 0,
		5: 0,
	}
	diff = deep.Equal(m1, m2)
	if diff == nil {
		t.Fatal("no diffs")
	}
	if len(diff) != deep.MaxDiff {
		t.Log(diff)
		t.Errorf("got %d diffs, expected %d", len(diff), deep.MaxDiff)
	}

	// Too many missing keys
	m1 = map[int]int{
		1: 1,
		2: 2,
	}
	m2 = map[int]int{
		1: 1,
		2: 2,
		3: 0,
		4: 0,
		5: 0,
		6: 0,
		7: 0,
	}
	diff = deep.Equal(m1, m2)
	if diff == nil {
		t.Fatal("no diffs")
	}
	if len(diff) != deep.MaxDiff {
		t.Log(diff)
		t.Errorf("got %d diffs, expected %d", len(diff), deep.MaxDiff)
	}
}

func TestNotHandled(t *testing.T) {
	a := func(int) {}
	b := func(int) {}
	diff := deep.Equal(a, b)
	if len(diff) > 0 {
		t.Error("got diffs:", diff)
	}
}

func TestStruct(t *testing.T) {
	type s1 struct {
		id     int
		Name   string
		Number int
	}
	sa := s1{
		id:     1,
		Name:   "foo",
		Number: 2,
	}
	sb := sa
	diff := deep.Equal(sa, sb)
	if len(diff) > 0 {
		t.Error("should be equal:", diff)
	}

	sb.Name = "bar"
	diff = deep.Equal(sa, sb)
	if diff == nil {
		t.Fatal("no diff")
	}
	if len(diff) != 1 {
		t.Error("too many diff:", diff)
	}
	if diff["Name"] != "foo != bar" {
		t.Error("wrong diff:", diff["Name"])
	}

	sb.Number = 22
	diff = deep.Equal(sa, sb)
	if diff == nil {
		t.Fatal("no diff")
	}
	if len(diff) != 2 {
		t.Error("too many diff:", diff)
	}
	if diff["Name"] != "foo != bar" {
		t.Error("wrong diff:", diff["Name"])
	}
	if diff["Number"] != "2 != 22" {
		t.Error("wrong diff:", diff["Number"])
	}

	sb.id = 11
	diff = deep.Equal(sa, sb)
	if diff == nil {
		t.Fatal("no diff")
	}
	if len(diff) != 2 {
		t.Error("too many diff:", diff)
	}
	if diff["Name"] != "foo != bar" {
		t.Error("wrong diff:", diff["Name"])
	}
	if diff["Number"] != "2 != 22" {
		t.Error("wrong diff:", diff["Number"])
	}
}

func TestStructWithTags(t *testing.T) {
	type s1 struct {
		same                    int
		modified                int
		sameIgnored             int `deep:"-"`
		modifiedIgnored         int `deep:"-"`
		ExportedSame            int
		ExportedModified        int
		ExportedSameIgnored     int `deep:"-"`
		ExportedModifiedIgnored int `deep:"-"`
	}
	type s2 struct {
		s1
		same                    int
		modified                int
		sameIgnored             int `deep:"-"`
		modifiedIgnored         int `deep:"-"`
		ExportedSame            int
		ExportedModified        int
		ExportedSameIgnored     int `deep:"-"`
		ExportedModifiedIgnored int `deep:"-"`
		recurseInline           s1
		recursePtr              *s2
	}
	sa := s2{
		s1: s1{
			same:                    0,
			modified:                1,
			sameIgnored:             2,
			modifiedIgnored:         3,
			ExportedSame:            4,
			ExportedModified:        5,
			ExportedSameIgnored:     6,
			ExportedModifiedIgnored: 7,
		},
		same:                    0,
		modified:                1,
		sameIgnored:             2,
		modifiedIgnored:         3,
		ExportedSame:            4,
		ExportedModified:        5,
		ExportedSameIgnored:     6,
		ExportedModifiedIgnored: 7,
		recurseInline: s1{
			same:                    0,
			modified:                1,
			sameIgnored:             2,
			modifiedIgnored:         3,
			ExportedSame:            4,
			ExportedModified:        5,
			ExportedSameIgnored:     6,
			ExportedModifiedIgnored: 7,
		},
		recursePtr: &s2{
			same:                    0,
			modified:                1,
			sameIgnored:             2,
			modifiedIgnored:         3,
			ExportedSame:            4,
			ExportedModified:        5,
			ExportedSameIgnored:     6,
			ExportedModifiedIgnored: 7,
		},
	}
	sb := s2{
		s1: s1{
			same:                    0,
			modified:                10,
			sameIgnored:             2,
			modifiedIgnored:         30,
			ExportedSame:            4,
			ExportedModified:        50,
			ExportedSameIgnored:     6,
			ExportedModifiedIgnored: 70,
		},
		same:                    0,
		modified:                10,
		sameIgnored:             2,
		modifiedIgnored:         30,
		ExportedSame:            4,
		ExportedModified:        50,
		ExportedSameIgnored:     6,
		ExportedModifiedIgnored: 70,
		recurseInline: s1{
			same:                    0,
			modified:                10,
			sameIgnored:             2,
			modifiedIgnored:         30,
			ExportedSame:            4,
			ExportedModified:        50,
			ExportedSameIgnored:     6,
			ExportedModifiedIgnored: 70,
		},
		recursePtr: &s2{
			same:                    0,
			modified:                10,
			sameIgnored:             2,
			modifiedIgnored:         30,
			ExportedSame:            4,
			ExportedModified:        50,
			ExportedSameIgnored:     6,
			ExportedModifiedIgnored: 70,
		},
	}

	orig := deep.CompareUnexportedFields
	deep.CompareUnexportedFields = true
	diff := deep.Equal(sa, sb)
	deep.CompareUnexportedFields = orig

	want := map[string]string{
		"s1.modified": "1 != 10",
		"s1.ExportedModified": "5 != 50",
		"modified": "1 != 10",
		"ExportedModified": "5 != 50",
		"recurseInline.modified": "1 != 10",
		"recurseInline.ExportedModified": "5 != 50",
		"recursePtr.modified": "1 != 10",
		"recursePtr.ExportedModified": "5 != 50",
	}

	if !reflect.DeepEqual(want, diff) {
		t.Errorf("\ngot  %v, \nwant %v", diff, want)
	}
}

func TestNestedStruct(t *testing.T) {
	type s2 struct {
		Nickname string
	}
	type s1 struct {
		Name  string
		Alias s2
	}
	sa := s1{
		Name:  "Robert",
		Alias: s2{Nickname: "Bob"},
	}
	sb := sa
	diff := deep.Equal(sa, sb)
	if len(diff) > 0 {
		t.Error("should be equal:", diff)
	}

	sb.Alias.Nickname = "Bobby"
	diff = deep.Equal(sa, sb)
	if diff == nil {
		t.Fatal("no diff")
	}
	if len(diff) != 1 {
		t.Error("too many diff:", diff)
	}
	if diff["Alias.Nickname"] != "Bob != Bobby" {
		t.Error("wrong diff:", diff["Alias.Nickname"])
	}
}

func TestMap(t *testing.T) {
	ma := map[string]int{
		"foo": 1,
		"bar": 2,
	}
	mb := map[string]int{
		"foo": 1,
		"bar": 2,
	}
	diff := deep.Equal(ma, mb)
	if len(diff) > 0 {
		t.Error("should be equal:", diff)
	}

	diff = deep.Equal(ma, ma)
	if len(diff) > 0 {
		t.Error("should be equal:", diff)
	}

	mb["foo"] = 111
	diff = deep.Equal(ma, mb)
	if diff == nil {
		t.Fatal("no diff")
	}
	if len(diff) != 1 {
		t.Error("too many diff:", diff)
	}
	if diff["map[foo]"] != "1 != 111" {
		t.Error("wrong diff:", diff["map[foo]"])
	}

	delete(mb, "foo")
	diff = deep.Equal(ma, mb)
	if diff == nil {
		t.Fatal("no diff")
	}
	if len(diff) != 1 {
		t.Error("too many diff:", diff)
	}
	if diff["map[foo]"] != "1 != <does not have key>" {
		t.Error("wrong diff:", diff["map[foo]"])
	}

	diff = deep.Equal(mb, ma)
	if diff == nil {
		t.Fatal("no diff")
	}
	if len(diff) != 1 {
		t.Error("too many diff:", diff)
	}
	if diff["map[foo]"] != "<does not have key> != 1" {
		t.Error("wrong diff:", diff["map[foo]"])
	}

	var mc map[string]int
	diff = deep.Equal(ma, mc)
	if diff == nil {
		t.Fatal("no diff")
	}
	if len(diff) != 1 {
		t.Error("too many diff:", diff)
	}
	// handle hash order randomness
	if diff[deep.Root] != "map[foo:1 bar:2] != <nil map>" && diff[deep.Root] != "map[bar:2 foo:1] != <nil map>" {
		t.Error("wrong diff:", diff[deep.Root])
	}

	diff = deep.Equal(mc, ma)
	if diff == nil {
		t.Fatal("no diff")
	}
	if len(diff) != 1 {
		t.Error("too many diff:", diff)
	}
	if diff[deep.Root] != "<nil map> != map[foo:1 bar:2]" && diff[deep.Root] != "<nil map> != map[bar:2 foo:1]" {
		t.Error("wrong diff:", diff[deep.Root])
	}
}

func TestArray(t *testing.T) {
	a := [3]int{1, 2, 3}
	b := [3]int{1, 2, 3}

	diff := deep.Equal(a, b)
	if len(diff) > 0 {
		t.Error("should be equal:", diff)
	}

	diff = deep.Equal(a, a)
	if len(diff) > 0 {
		t.Error("should be equal:", diff)
	}

	b[2] = 333
	diff = deep.Equal(a, b)
	if diff == nil {
		t.Fatal("no diff")
	}
	if len(diff) != 1 {
		t.Error("too many diff:", diff)
	}
	if diff["array[2]"] != "3 != 333" {
		t.Error("wrong diff:", diff["array[2]"])
	}

	c := [3]int{1, 2, 2}
	diff = deep.Equal(a, c)
	if diff == nil {
		t.Fatal("no diff")
	}
	if len(diff) != 1 {
		t.Error("too many diff:", diff)
	}
	if diff["array[2]"] != "3 != 2" {
		t.Error("wrong diff:", diff["array[2]"])
	}

	var d [2]int
	diff = deep.Equal(a, d)
	if diff == nil {
		t.Fatal("no diff")
	}
	if len(diff) != 1 {
		t.Error("too many diff:", diff)
	}
	if diff[deep.Root] != "[3]int != [2]int" {
		t.Error("wrong diff:", diff[deep.Root])
	}

	e := [12]int{}
	f := [12]int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	diff = deep.Equal(e, f)
	if diff == nil {
		t.Fatal("no diff")
	}
	if len(diff) != deep.MaxDiff {
		t.Error("not enough diffs:", diff)
	}
	for i := 0; i < deep.MaxDiff; i++ {
		if diff[fmt.Sprintf("array[%d]", i+1)] != fmt.Sprintf("0 != %d", i+1) {
			t.Error("wrong diff:", diff[fmt.Sprintf("array[%d]", i+1)])
		}
	}
}

func TestSlice(t *testing.T) {
	a := []int{1, 2, 3}
	b := []int{1, 2, 3}

	diff := deep.Equal(a, b)
	if len(diff) > 0 {
		t.Error("should be equal:", diff)
	}

	diff = deep.Equal(a, a)
	if len(diff) > 0 {
		t.Error("should be equal:", diff)
	}

	b[2] = 333
	diff = deep.Equal(a, b)
	if diff == nil {
		t.Fatal("no diff")
	}
	if len(diff) != 1 {
		t.Error("too many diff:", diff)
	}
	if diff["slice[2]"] != "3 != 333" {
		t.Error("wrong diff:", diff["slice[2]"])
	}

	b = b[0:2]
	diff = deep.Equal(a, b)
	if diff == nil {
		t.Fatal("no diff")
	}
	if len(diff) != 1 {
		t.Error("too many diff:", diff)
	}
	if diff["slice[2]"] != "3 != <no value>" {
		t.Error("wrong diff:", diff["slice[2]"])
	}

	diff = deep.Equal(b, a)
	if diff == nil {
		t.Fatal("no diff")
	}
	if len(diff) != 1 {
		t.Error("too many diff:", diff)
	}
	if diff["slice[2]"] != "<no value> != 3" {
		t.Error("wrong diff:", diff["slice[2]"])
	}

	var c []int
	diff = deep.Equal(a, c)
	if diff == nil {
		t.Fatal("no diff")
	}
	if len(diff) != 1 {
		t.Error("too many diff:", diff)
	}
	if diff[deep.Root] != "[1 2 3] != <nil slice>" {
		t.Error("wrong diff:", diff[deep.Root])
	}

	diff = deep.Equal(c, a)
	if diff == nil {
		t.Fatal("no diff")
	}
	if len(diff) != 1 {
		t.Error("too many diff:", diff)
	}
	if diff[deep.Root] != "<nil slice> != [1 2 3]" {
		t.Error("wrong diff:", diff[deep.Root])
	}
}

func TestSiblingSlices(t *testing.T) {
	father := []int{1, 2, 3, 4}
	a := father[0:3]
	b := father[0:3]

	diff := deep.Equal(a, b)
	if len(diff) > 0 {
		t.Error("should be equal:", diff)
	}
	diff = deep.Equal(b, a)
	if len(diff) > 0 {
		t.Error("should be equal:", diff)
	}

	a = father[0:3]
	b = father[0:2]
	diff = deep.Equal(a, b)
	if diff == nil {
		t.Fatal("no diff")
	}
	if len(diff) != 1 {
		t.Error("too many diff:", diff)
	}
	if diff["slice[2]"] != "3 != <no value>" {
		t.Error("wrong diff:", diff["slice[2]"])
	}

	a = father[0:2]
	b = father[0:3]

	diff = deep.Equal(a, b)
	if diff == nil {
		t.Fatal("no diff")
	}
	if len(diff) != 1 {
		t.Error("too many diff:", diff)
	}
	if diff["slice[2]"] != "<no value> != 3" {
		t.Error("wrong diff:", diff["slice[2]"])
	}

	a = father[0:2]
	b = father[2:4]

	diff = deep.Equal(a, b)
	if diff == nil {
		t.Fatal("no diff")
	}
	if len(diff) != 2 {
		t.Error("too many diff:", diff)
	}
	if diff["slice[0]"] != "1 != 3" {
		t.Error("wrong diff:", diff["slice[0]"])
	}
	if diff["slice[1]"] != "2 != 4" {
		t.Error("wrong diff:", diff["slice[1]"])
	}

	a = father[0:0]
	b = father[1:1]

	diff = deep.Equal(a, b)
	if len(diff) > 0 {
		t.Error("should be equal:", diff)
	}
	diff = deep.Equal(b, a)
	if len(diff) > 0 {
		t.Error("should be equal:", diff)
	}
}

func TestEmptySlice(t *testing.T) {
	a := []int{1}
	b := []int{}
	var c []int

	// Non-empty is not equal to empty.
	diff := deep.Equal(a, b)
	if diff == nil {
		t.Fatal("no diff")
	}
	if len(diff) != 1 {
		t.Error("too many diff:", diff)
	}
	if diff["slice[0]"] != "1 != <no value>" {
		t.Error("wrong diff:", diff["slice[0]"])
	}

	// Empty is not equal to non-empty.
	diff = deep.Equal(b, a)
	if diff == nil {
		t.Fatal("no diff")
	}
	if len(diff) != 1 {
		t.Error("too many diff:", diff)
	}
	if diff["slice[0]"] != "<no value> != 1" {
		t.Error("wrong diff:", diff["slice[0]"])
	}

	// Empty is not equal to nil.
	diff = deep.Equal(b, c)
	if diff == nil {
		t.Fatal("no diff")
	}
	if len(diff) != 1 {
		t.Error("too many diff:", diff)
	}
	if diff[deep.Root] != "[] != <nil slice>" {
		t.Error("wrong diff:", diff[deep.Root])
	}

	// Nil is not equal to empty.
	diff = deep.Equal(c, b)
	if diff == nil {
		t.Fatal("no diff")
	}
	if len(diff) != 1 {
		t.Error("too many diff:", diff)
	}
	if diff[deep.Root] != "<nil slice> != []" {
		t.Error("wrong diff:", diff[deep.Root])
	}
}

func TestNilSlicesAreEmpty(t *testing.T) {
	defaultNilSlicesAreEmpty := deep.NilSlicesAreEmpty
	deep.NilSlicesAreEmpty = true
	defer func() { deep.NilSlicesAreEmpty = defaultNilSlicesAreEmpty }()

	a := []int{1}
	b := []int{}
	var c []int

	// Empty is equal to nil.
	diff := deep.Equal(b, c)
	if len(diff) > 0 {
		t.Error("should be equal:", diff)
	}

	// Nil is equal to empty.
	diff = deep.Equal(c, b)
	if len(diff) > 0 {
		t.Error("should be equal:", diff)
	}

	// Non-empty is not equal to nil.
	diff = deep.Equal(a, c)
	if diff == nil {
		t.Fatal("no diff")
	}
	if len(diff) != 1 {
		t.Error("too many diff:", diff)
	}
	if diff[deep.Root] != "[1] != <nil slice>" {
		t.Error("wrong diff:", diff[deep.Root])
	}

	// Nil is not equal to non-empty.
	diff = deep.Equal(c, a)
	if diff == nil {
		t.Fatal("no diff")
	}
	if len(diff) != 1 {
		t.Error("too many diff:", diff)
	}
	if diff[deep.Root] != "<nil slice> != [1]" {
		t.Error("wrong diff:", diff[deep.Root])
	}

	// Non-empty is not equal to empty.
	diff = deep.Equal(a, b)
	if diff == nil {
		t.Fatal("no diff")
	}
	if len(diff) != 1 {
		t.Error("too many diff:", diff)
	}
	if diff["slice[0]"] != "1 != <no value>" {
		t.Error("wrong diff:", diff["slice[0]"])
	}

	// Empty is not equal to non-empty.
	diff = deep.Equal(b, a)
	if diff == nil {
		t.Fatal("no diff")
	}
	if len(diff) != 1 {
		t.Error("too many diff:", diff)
	}
	if diff["slice[0]"] != "<no value> != 1" {
		t.Error("wrong diff:", diff["slice[0]"])
	}
}

func TestNilMapsAreEmpty(t *testing.T) {
	defaultNilMapsAreEmpty := deep.NilSlicesAreEmpty
	deep.NilMapsAreEmpty = true
	defer func() { deep.NilSlicesAreEmpty = defaultNilMapsAreEmpty }()

	a := map[int]int{1: 1}
	b := map[int]int{}
	var c map[int]int

	// Empty is equal to nil.
	diff := deep.Equal(b, c)
	if len(diff) > 0 {
		t.Error("should be equal:", diff)
	}

	// Nil is equal to empty.
	diff = deep.Equal(c, b)
	if len(diff) > 0 {
		t.Error("should be equal:", diff)
	}

	// Non-empty is not equal to nil.
	diff = deep.Equal(a, c)
	if diff == nil {
		t.Fatal("no diff")
	}
	if len(diff) != 1 {
		t.Error("too many diff:", diff)
	}
	if diff[deep.Root] != "map[1:1] != <nil map>" {
		t.Error("wrong diff:", diff[deep.Root])
	}

	// Nil is not equal to non-empty.
	diff = deep.Equal(c, a)
	if diff == nil {
		t.Fatal("no diff")
	}
	if len(diff) != 1 {
		t.Error("too many diff:", diff)
	}
	if diff[deep.Root] != "<nil map> != map[1:1]" {
		t.Error("wrong diff:", diff[deep.Root])
	}

	// Non-empty is not equal to empty.
	diff = deep.Equal(a, b)
	if diff == nil {
		t.Fatal("no diff")
	}
	if len(diff) != 1 {
		t.Error("too many diff:", diff)
	}
	if diff["map[1]"] != "1 != <does not have key>" {
		t.Error("wrong diff:", diff["map[1]"])
	}

	// Empty is not equal to non-empty.
	diff = deep.Equal(b, a)
	if diff == nil {
		t.Fatal("no diff")
	}
	if len(diff) != 1 {
		t.Error("too many diff:", diff)
	}
	if diff["map[1]"] != "<does not have key> != 1" {
		t.Error("wrong diff:", diff["map[1]"])
	}
}

func TestNilInterface(t *testing.T) {
	type T struct{ i int }

	a := &T{i: 1}
	diff := deep.Equal(nil, a)
	if diff == nil {
		t.Fatal("no diff")
	}
	if len(diff) != 1 {
		t.Error("too many diff:", diff)
	}
	if diff[deep.Root] != "<nil pointer> != &{1}" {
		t.Error("wrong diff:", diff[deep.Root])
	}

	diff = deep.Equal(a, nil)
	if diff == nil {
		t.Fatal("no diff")
	}
	if len(diff) != 1 {
		t.Error("too many diff:", diff)
	}
	if diff[deep.Root] != "&{1} != <nil pointer>" {
		t.Error("wrong diff:", diff[deep.Root])
	}

	diff = deep.Equal(nil, nil)
	if len(diff) > 0 {
		t.Error("should be equal:", diff)
	}
}

func TestPointer(t *testing.T) {
	type T struct{ i int }

	a, b := &T{i: 1}, &T{i: 1}
	diff := deep.Equal(a, b)
	if len(diff) > 0 {
		t.Error("should be equal:", diff)
	}

	a, b = nil, &T{}
	diff = deep.Equal(a, b)
	if diff == nil {
		t.Fatal("no diff")
	}
	if len(diff) != 1 {
		t.Error("too many diff:", diff)
	}
	if diff[deep.Root] != "<nil pointer> != deep_test.T" {
		t.Error("wrong diff:", diff[deep.Root])
	}

	a, b = &T{}, nil
	diff = deep.Equal(a, b)
	if diff == nil {
		t.Fatal("no diff")
	}
	if len(diff) != 1 {
		t.Error("too many diff:", diff)
	}
	if diff[deep.Root] != "deep_test.T != <nil pointer>" {
		t.Error("wrong diff:", deep.Root)
	}

	a, b = nil, nil
	diff = deep.Equal(a, b)
	if len(diff) > 0 {
		t.Error("should be equal:", diff)
	}
}

func TestTime(t *testing.T) {
	// In an interable kind (i.e. a struct)
	type sTime struct {
		T time.Time
	}
	now := time.Now()
	got := sTime{T: now}
	expect := sTime{T: now.Add(1 * time.Second)}
	diff := deep.Equal(got, expect)
	if len(diff) != 1 {
		t.Error("expected 1 diff:", diff)
	}

	// Directly
	a := now
	b := now
	diff = deep.Equal(a, b)
	if len(diff) > 0 {
		t.Error("should be equal:", diff)
	}

	// https://github.com/go-test/deep/issues/15
	type Time15 struct {
		time.Time
	}
	a15 := Time15{now}
	b15 := Time15{now}
	diff = deep.Equal(a15, b15)
	if len(diff) > 0 {
		t.Error("should be equal:", diff)
	}

	later := now.Add(1 * time.Second)
	b15 = Time15{later}
	diff = deep.Equal(a15, b15)
	if len(diff) != 1 {
		t.Errorf("got %d diffs, expected 1: %s", len(diff), diff)
	}

	// No diff in Equal should not affect diff of other fields (Foo)
	type Time17 struct {
		time.Time
		Foo int
	}
	a17 := Time17{Time: now, Foo: 1}
	b17 := Time17{Time: now, Foo: 2}
	diff = deep.Equal(a17, b17)
	if len(diff) != 1 {
		t.Errorf("got %d diffs, expected 1: %s", len(diff), diff)
	}
}

func TestTimeUnexported(t *testing.T) {
	// https://github.com/go-test/deep/issues/18
	// Can't call Call() on exported Value func
	defaultCompareUnexportedFields := deep.CompareUnexportedFields
	deep.CompareUnexportedFields = true
	defer func() { deep.CompareUnexportedFields = defaultCompareUnexportedFields }()

	now := time.Now()
	type hiddenTime struct {
		t time.Time
	}
	htA := &hiddenTime{t: now}
	htB := &hiddenTime{t: now}
	diff := deep.Equal(htA, htB)
	if len(diff) > 0 {
		t.Error("should be equal:", diff)
	}

	// This doesn't call time.Time.Equal(), it compares the unexported fields
	// in time.Time, causing a diff like:
	// [t.wall: 13740788835924462040 != 13740788836998203864 t.ext: 1447549 != 1001447549]
	later := now.Add(1 * time.Second)
	htC := &hiddenTime{t: later}
	diff = deep.Equal(htA, htC)

	expected := 1
	if _, ok := reflect.TypeOf(htA.t).FieldByName("ext"); ok {
		expected = 2
	}
	if len(diff) != expected {
		t.Errorf("got %d diffs, expected %d: %s", len(diff), expected, diff)
	}
}

func TestInterface(t *testing.T) {
	a := map[string]interface{}{
		"foo": map[string]string{
			"bar": "a",
		},
	}
	b := map[string]interface{}{
		"foo": map[string]string{
			"bar": "b",
		},
	}
	diff := deep.Equal(a, b)
	if len(diff) == 0 {
		t.Fatalf("expected 1 diff, got zero")
	}
	if len(diff) != 1 {
		t.Errorf("expected 1 diff, got %d: %s", len(diff), diff)
	}
}

func TestInterface2(t *testing.T) {
	defer func() {
		if val := recover(); val != nil {
			t.Fatalf("panic: %v", val)
		}
	}()

	a := map[string]interface{}{
		"bar": 1,
	}
	b := map[string]interface{}{
		"bar": 1.23,
	}
	diff := deep.Equal(a, b)
	if len(diff) == 0 {
		t.Fatalf("expected 1 diff, got zero")
	}
	if len(diff) != 1 {
		t.Errorf("expected 1 diff, got %d: %s", len(diff), diff)
	}
}

func TestInterface3(t *testing.T) {
	type Value struct{ int }
	a := map[string]interface{}{
		"foo": &Value{},
	}
	b := map[string]interface{}{
		"foo": 1.23,
	}
	diff := deep.Equal(a, b)
	if len(diff) == 0 {
		t.Fatalf("expected 1 diff, got zero")
	}

	if len(diff) != 1 {
		t.Errorf("expected 1 diff, got: %s", diff)
	}
}

func TestError(t *testing.T) {
	a := errors.New("it broke")
	b := errors.New("it broke")

	diff := deep.Equal(a, b)
	if len(diff) != 0 {
		t.Fatalf("expected zero diffs, got %d: %s", len(diff), diff)
	}

	b = errors.New("it fell apart")
	diff = deep.Equal(a, b)
	if len(diff) != 1 {
		t.Fatalf("expected 1 diff, got %d: %s", len(diff), diff)
	}
	if diff[deep.Root] != "it broke != it fell apart" {
		t.Errorf("got '%s', expected 'it broke != it fell apart'", diff[deep.Root])
	}

	// Both errors set
	type tWithError struct {
		Error error
	}
	t1 := tWithError{
		Error: a,
	}
	t2 := tWithError{
		Error: b,
	}
	diff = deep.Equal(t1, t2)
	if len(diff) != 1 {
		t.Fatalf("expected 1 diff, got %d: %s", len(diff), diff)
	}
	if diff["Error"] != "it broke != it fell apart" {
		t.Errorf("got '%s', expected 'Error: it broke != it fell apart'", diff["Error"])
	}

	// Both errors nil
	t1 = tWithError{
		Error: nil,
	}
	t2 = tWithError{
		Error: nil,
	}
	diff = deep.Equal(t1, t2)
	if len(diff) != 0 {
		t.Log(diff)
		t.Fatalf("expected 0 diff, got %d: %s", len(diff), diff)
	}

	// One error is nil
	t1 = tWithError{
		Error: errors.New("foo"),
	}
	t2 = tWithError{
		Error: nil,
	}
	diff = deep.Equal(t1, t2)
	if len(diff) != 1 {
		t.Log(diff)
		t.Fatalf("expected 1 diff, got %d: %s", len(diff), diff)
	}
	if diff["Error"] != "*errors.errorString != <nil pointer>" {
		t.Errorf("got '%s', expected 'Error: *errors.errorString != <nil pointer>'", diff["Error"])
	}
}

func TestErrorWithOtherFields(t *testing.T) {
	a := errors.New("it broke")
	b := errors.New("it broke")

	diff := deep.Equal(a, b)
	if len(diff) != 0 {
		t.Fatalf("expected zero diffs, got %d: %s", len(diff), diff)
	}

	b = errors.New("it fell apart")
	diff = deep.Equal(a, b)
	if len(diff) != 1 {
		t.Fatalf("expected 1 diff, got %d: %s", len(diff), diff)
	}
	if diff[deep.Root] != "it broke != it fell apart" {
		t.Errorf("got '%s', expected 'it broke != it fell apart'", diff[deep.Root])
	}

	// Both errors set
	type tWithError struct {
		Error error
		Other string
	}
	t1 := tWithError{
		Error: a,
		Other: "ok",
	}
	t2 := tWithError{
		Error: b,
		Other: "ok",
	}
	diff = deep.Equal(t1, t2)
	if len(diff) != 1 {
		t.Fatalf("expected 1 diff, got %d: %s", len(diff), diff)
	}
	if diff["Error"] != "it broke != it fell apart" {
		t.Errorf("got '%s', expected 'Error: it broke != it fell apart'", diff["Error"])
	}

	// Both errors nil
	t1 = tWithError{
		Error: nil,
		Other: "ok",
	}
	t2 = tWithError{
		Error: nil,
		Other: "ok",
	}
	diff = deep.Equal(t1, t2)
	if len(diff) != 0 {
		t.Log(diff)
		t.Fatalf("expected 0 diff, got %d: %s", len(diff), diff)
	}

	// Different Other value
	t1 = tWithError{
		Error: nil,
		Other: "ok",
	}
	t2 = tWithError{
		Error: nil,
		Other: "nope",
	}
	diff = deep.Equal(t1, t2)
	if len(diff) != 1 {
		t.Fatalf("expected 1 diff, got %d: %s", len(diff), diff)
	}
	if diff["Other"] != "ok != nope" {
		t.Errorf("got '%s', expected 'Other: ok != nope'", diff["Other"])
	}

	// Different Other value, same error
	t1 = tWithError{
		Error: a,
		Other: "ok",
	}
	t2 = tWithError{
		Error: a,
		Other: "nope",
	}
	diff = deep.Equal(t1, t2)
	if len(diff) != 1 {
		t.Fatalf("expected 1 diff, got %d: %s", len(diff), diff)
	}
	if diff["Other"] != "ok != nope" {
		t.Errorf("got '%s', expected 'Other: ok != nope'", diff["Other"])
	}
}

type primKindError string

func (e primKindError) Error() string {
	return string(e)
}

func TestErrorPrimitiveKind(t *testing.T) {
	// The primKindError type above is valid and used by Go, e.g.
	// url.EscapeError and url.InvalidHostError. Before fixing this bug
	// (https://github.com/go-test/deep/issues/31), we presumed a and b
	// were ptr or interface (and not nil), so a.Elem() worked. But when
	// a/b are primitive kinds, Elem() causes a panic.
	var err1 primKindError = "abc"
	var err2 primKindError = "abc"
	diff := deep.Equal(err1, err2)
	if len(diff) != 0 {
		t.Fatalf("expected zero diffs, got %d: %s", len(diff), diff)
	}
}

func TestNil(t *testing.T) {
	type student struct {
		name string
		age  int
	}

	mark := student{"mark", 10}
	var someNilThing interface{} = nil
	diff := deep.Equal(someNilThing, mark)
	if diff == nil {
		t.Error("Nil value to comparison should not be equal")
	}
	diff = deep.Equal(mark, someNilThing)
	if diff == nil {
		t.Error("Nil value to comparison should not be equal")
	}
	diff = deep.Equal(someNilThing, someNilThing)
	if diff != nil {
		t.Error("Nil value to comparison should not be equal")
	}
}


func TestEqualSubsetWithMixedKeys(t *testing.T) {
	type student struct {
		Name string
		Age  int
		Arr  []string
	}

	left := student{"mark", 10, []string{"same1", "different", "same2"}}
	right := student{"mark", 10, []string{"same1", "same2", "different"}}

	diff, isSame := deep.EqualSubset(left, right, nil, true, false)

	if !isSame {
		for i, s := range diff {
			println(fmt.Sprint(i) + " " + s)
		}
		t.Error("This should be a subset match")
	}
}


func TestEqualSubsetWithDifferentNumberOfKeys(t *testing.T) {
	type student struct {
		Name string
		Age  int
		Arr  []string
	}

	left := student{"mark", 10, []string{"same1", "same2"}}
	right := student{"mark", 10, []string{"same1", "same2", "different"}}

	diff, isSame := deep.EqualSubset(left, right, nil,true, false)

	if !isSame {
		for i, s := range diff {
			println(fmt.Sprint(i) + " " + s)
		}
		t.Error("This should be a subset match")
	}
}

func TestEqualSubsetWithEmptyString(t *testing.T) {
	type student struct {
		Name string
		Age  int
		Arr  []string
	}

	left := student{"", 10, []string{"same1", "same2", "different"}}
	right := student{"mark", 10, []string{"same1", "same2", "different"}}

	_, isSame := deep.EqualSubset(left, right, nil,true, false)

	if !isSame {
		t.Error("This should be a subset match")
	}
}

func TestEqualSubsetWithMaps(t *testing.T) {
	type student struct {
		Name string
		Age  int
		Arr  []string
		Map  map[string]string
	}

	leftMap := make(map[string]string)
	rightMap := make(map[string]string)
	rightMap["bla"] = "blue"

	left := student{"", 10, []string{"same1", "same2", "different"}, leftMap}
	right := student{"mark", 10, []string{"same1", "same2", "different"}, rightMap}

	diff, isSame := deep.EqualSubset(left, right, nil,true, false)

	if !isSame {
		for s, s2 := range diff {
			println(s+ " " +s2)
		}
		t.Error("This should be a subset match")
	}
}

func TestEqualSubsetWithOneDifferringKey(t *testing.T) {
	type student struct {
		Name string
		Age  int
		Arr  []string
	}

	left := student{"mark", 10, []string{"same1", "same2", "really-different"}}
	right := student{"mark", 10, []string{"same1", "same2", "different"}}

	_, isSame := deep.EqualSubset(left, right, nil,true, false)

	if isSame {
		t.Error("This should not be a subset match")
	}
}


func TestEqualSubsetWithNils(t *testing.T) {
	type Nested struct {
		NestedName string
	}

	type student struct {
		Name string
		Age  int
		Nested *Nested
		Arr  []string
	}

	nested := &Nested{"I AM NESTED!!!"}
	left := student{Name: "mark", Age: 10, Arr: []string{"same1", "same2"}}
	right := student{"mark", 10, nested, []string{"same1", "same2"}}

	diff, isSame := deep.EqualSubset(left, right, nil, true, false)

	if !isSame {
		for i, s := range diff {
			println(fmt.Sprint(i) + " " + s)
		}
		t.Error("This should be a subset match")
	}
}

func TestEqualSubsetWithNestedStruct(t *testing.T) {
	type Nested struct {
		NestedName string
	}

	type student struct {
		Name string
		Age  int
		Nested *Nested
		Arr  []string
	}

	nestedLeft := &Nested{"I AM NESTED!!!"}
	nestedRight := &Nested{"I AM NESTED!!!"}
	left := student{ "mark",  10, nestedLeft, []string{"same1", "same2"}}
	right := student{"mark", 10, nestedRight, []string{"same1", "same2"}}

	diff, isSame := deep.EqualSubset(left, right, nil,true, false)

	if !isSame {
		for i, s := range diff {
			println(fmt.Sprint(i) + " " + s)
		}
		t.Error("This should be a subset match")
	}
}


func TestEqualSubsetWithNestedStructInArray(t *testing.T) {
	type Nested struct {
		NestedName string
	}

	type student struct {
		Name string
		Age  int
		Nested []Nested
		Arr  []string
	}

	nestedLeft := Nested{"I AM NESTED!!!"}
	nestedRight := Nested{"I AM NESTED!!!"}
	left := student{ "mark",  10, []Nested{nestedLeft}, []string{"same1", "same2"}}
	right := student{"mark", 10, []Nested{nestedRight}, []string{"same1", "same2"}}

	diff, isSame := deep.EqualSubset(left, right, nil,true, false)

	if !isSame {
		for i, s := range diff {
			println(fmt.Sprint(i) + " " + s)
		}
		t.Error("This should be a subset match")
	}
}


func TestEqualSubsetWithIgrnoredPath(t *testing.T) {
	type DeeplyNested struct {
		NestedName string
	}

	type Nested struct {
		NestedName string
		DeeplyNested *DeeplyNested
	}

	type student struct {
		Name string
		Age  int
		Nested *Nested
		Arr  []string
	}

	deeplyNested := &DeeplyNested{"I AM DEEPLY NESTED!!!"}
	nestedRight := &Nested{"I AM NESTED!!!", nil}
	nestedLeft := &Nested{"I AM NESTED!!!", deeplyNested}
	left := student{ "mark",  10, nestedLeft, []string{"same1", "same2"}}
	right := student{"mark", 10, nestedRight, []string{"same1", "same2"}}
	ignored := []string{"Nested.DeeplyNested"}

	diff, isSame := deep.EqualSubset(left, right, ignored,true, false)

	if !isSame {
		for i, s := range diff {
			println(fmt.Sprint(i) + " " + s)
		}
		t.Error("This should be a subset match")
	}
}


func TestTypeMismatch(t *testing.T) {
	type T1 int // same type kind (int)
	type T2 int // but different type
	var t1 T1 = 1
	var t2 T2 = 1
	diff := deep.Equal(t1, t2)
	if diff == nil {
		t.Fatal("no diff")
	}
	if len(diff) != 1 {
		t.Error("too many diff:", diff)
	}
	if diff[deep.Root] != "deep_test.T1 != deep_test.T2" {
		t.Error("wrong diff:", diff[deep.Root])
	}

	// Same pkg name but differnet full paths
	// https://github.com/go-test/deep/issues/39
	err1 := v1.Error{}
	err2 := v2.Error{}
	diff = deep.Equal(err1, err2)
	if diff == nil {
		t.Fatal("no diff")
	}
	if len(diff) != 1 {
		t.Error("too many diff:", diff)
	}
	if diff[deep.Root] != "github.com/instana/deep/test/v1.Error != github.com/instana/deep/test/v2.Error" {
		t.Error("wrong diff:", diff[deep.Root])
	}
}