// generics_examples.go — ví dụ thực hành đi kèm cheatsheet Generics (docs/generics.html).
// Chạy: go run generics_examples.go   (cần Go 1.23+ cho iter; mục 11 cần Go 1.26)
package main

import (
	"cmp"
	"fmt"
	"iter"
	"maps"
	"slices"
	"strconv"
	"strings"
	"sync"
)

func title(s string) { fmt.Printf("\n=== %s ===\n", s) }

// ---------------------------------------------------------------------------
// 1. Hàm generic: [type parameters] + constraint
// ---------------------------------------------------------------------------

func Map[T, U any](s []T, f func(T) U) []U {
	out := make([]U, 0, len(s))
	for _, v := range s {
		out = append(out, f(v))
	}
	return out
}

func Filter[T any](s []T, keep func(T) bool) []T {
	var out []T
	for _, v := range s {
		if keep(v) {
			out = append(out, v)
		}
	}
	return out
}

func Reduce[T, A any](s []T, init A, f func(A, T) A) A {
	acc := init
	for _, v := range s {
		acc = f(acc, v)
	}
	return acc
}

// ---------------------------------------------------------------------------
// 2. Constraint = interface mô tả TYPE SET
// ---------------------------------------------------------------------------

type Number interface {
	~int | ~int32 | ~int64 | ~float32 | ~float64 // ~ : chấp nhận cả type có underlying là int...
}

func Sum[T Number](xs ...T) T {
	var total T // zero value của T
	for _, x := range xs {
		total += x // được phép vì MỌI type trong type set đều hỗ trợ +
	}
	return total
}

type Celsius float64 // underlying type float64 -> thỏa ~float64

func MaxOf[T cmp.Ordered](a, b T) T { // cmp.Ordered: hỗ trợ < <= > >=
	if a > b {
		return a
	}
	return b
}

// ---------------------------------------------------------------------------
// 3. Kiểu generic: Stack, Set, Cache
// ---------------------------------------------------------------------------

type Stack[T any] struct{ items []T }

func (s *Stack[T]) Push(v T) { s.items = append(s.items, v) }

func (s *Stack[T]) Pop() (T, bool) {
	if len(s.items) == 0 {
		var zero T // không thể "return nil" cho T bất kỳ
		return zero, false
	}
	v := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return v, true
}

type Set[T comparable] map[T]struct{} // key map bắt buộc comparable

func NewSet[T comparable](xs ...T) Set[T] {
	s := make(Set[T], len(xs))
	for _, x := range xs {
		s[x] = struct{}{}
	}
	return s
}
func (s Set[T]) Has(v T) bool { _, ok := s[v]; return ok }

// All trả về iterator (Go 1.23 range-over-func)
func (s Set[T]) All() iter.Seq[T] { return maps.Keys(s) }

type Cache[K comparable, V any] struct {
	mu sync.Mutex
	m  map[K]V
}

func NewCache[K comparable, V any]() *Cache[K, V] { return &Cache[K, V]{m: map[K]V{}} }

func (c *Cache[K, V]) GetOrLoad(k K, load func(K) V) V {
	c.mu.Lock()
	defer c.mu.Unlock()
	if v, ok := c.m[k]; ok {
		return v
	}
	v := load(k)
	c.m[k] = v
	return v
}

// ---------------------------------------------------------------------------
// 4. S ~[]E: giữ nguyên named slice type
// ---------------------------------------------------------------------------

type Names []string

func (n Names) Join() string { return strings.Join(n, ", ") }

func CloneLoose[E any](s []E) []E    { return append([]E(nil), s...) }
func CloneKeep[S ~[]E, E any](s S) S { return append(S(nil), s...) }
func Last[S ~[]E, E any](s S) (E, bool) {
	var z E
	if len(s) == 0 {
		return z, false
	}
	return s[len(s)-1], true
}

// ---------------------------------------------------------------------------
// 5. Không type switch trực tiếp trên T -> chuyển qua any
// ---------------------------------------------------------------------------

func Describe[T any](v T) string {
	switch x := any(v).(type) {
	case int:
		return "int " + strconv.Itoa(x)
	case string:
		return "string " + strconv.Quote(x)
	default:
		return fmt.Sprintf("khác: %T", x)
	}
}

// ---------------------------------------------------------------------------
// 6. Constraint chứa pointer method: mẫu [T any, PT interface{ *T; M() }]
// ---------------------------------------------------------------------------

type Setter[T any] interface {
	*T                  // PT phải là *T
	Set(s string) error // và có method Set (pointer receiver)
}

func ParseAll[T any, PT Setter[T]](in []string) ([]T, error) {
	out := make([]T, len(in))
	for i, s := range in {
		if err := PT(&out[i]).Set(s); err != nil {
			return nil, fmt.Errorf("phần tử %d: %w", i, err)
		}
	}
	return out, nil
}

type Port int

func (p *Port) Set(s string) error {
	n, err := strconv.Atoi(s)
	if err != nil || n < 1 || n > 65535 {
		return fmt.Errorf("port không hợp lệ %q", s)
	}
	*p = Port(n)
	return nil
}

// ---------------------------------------------------------------------------
// 7. Suy luận kiểu thất bại khi T chỉ nằm ở kết quả
// ---------------------------------------------------------------------------

func Zero[T any]() T { var z T; return z }

// ---------------------------------------------------------------------------
// 8. comparable có thể panic lúc chạy khi T là interface
// ---------------------------------------------------------------------------

func Equal[T comparable](a, b T) bool { return a == b }

func safe(f func()) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()
	f()
	return nil
}

// ---------------------------------------------------------------------------
// 9. Iterator generic (Go 1.23)
// ---------------------------------------------------------------------------

func FilterSeq[T any](seq iter.Seq[T], keep func(T) bool) iter.Seq[T] {
	return func(yield func(T) bool) {
		for v := range seq {
			if keep(v) && !yield(v) {
				return
			}
		}
	}
}

// ---------------------------------------------------------------------------
// 10. Generic type alias (Go 1.24)
// ---------------------------------------------------------------------------

type Dict[V any] = map[string]V

// ---------------------------------------------------------------------------
// 11. Constraint tự tham chiếu (Go 1.26)
// ---------------------------------------------------------------------------

type Adder[A Adder[A]] interface {
	Add(A) A
}

type Vec struct{ X, Y int }

func (v Vec) Add(o Vec) Vec { return Vec{v.X + o.X, v.Y + o.Y} }

func SumAll[A Adder[A]](xs ...A) A {
	var total A
	for _, x := range xs {
		total = total.Add(x)
	}
	return total
}

type User struct {
	Name string
	Age  int
}

func main() {
	title("1. Map / Filter / Reduce")
	nums := []int{1, 2, 3, 4, 5}
	strs := Map(nums, func(n int) string { return "#" + strconv.Itoa(n) }) // suy luận T=int, U=string
	even := Filter(nums, func(n int) bool { return n%2 == 0 })
	total := Reduce(nums, 0, func(acc, n int) int { return acc + n })
	fmt.Println(strs, even, total)
	fmt.Println(Map[int, float64](nums, func(n int) float64 { return float64(n) / 2 })) // chỉ định tường minh

	title("2. Constraint & ~")
	fmt.Println(Sum(1, 2, 3), Sum(1.5, 2.5), Sum[Celsius](36.5, 0.5))
	fmt.Println(MaxOf(3, 7), MaxOf("go", "rust"), MaxOf(2.5, -1))

	title("3. Kiểu generic")
	var st Stack[string]
	st.Push("a")
	st.Push("b")
	v, ok := st.Pop()
	fmt.Println(v, ok)
	st.Pop()
	v, ok = st.Pop()
	fmt.Printf("%q %v (zero value khi rỗng)\n", v, ok)

	set := NewSet("go", "rust", "go")
	fmt.Println(len(set), set.Has("go"), set.Has("java"), slices.Sorted(set.All()))

	cache := NewCache[int, string]()
	load := func(k int) string { fmt.Println("  load", k); return strings.Repeat("*", k) }
	fmt.Println(cache.GetOrLoad(3, load), cache.GetOrLoad(3, load)) // lần 2 không load

	title("4. S ~[]E giữ named type")
	names := Names{"An", "Bình"}
	fmt.Printf("CloneLoose: %T\n", CloneLoose(names)) // []string -> mất method Join
	fmt.Printf("CloneKeep:  %T -> %s\n", CloneKeep(names), CloneKeep(names).Join())
	last, _ := Last(names)
	fmt.Println("Last:", last)

	title("5. Type switch qua any(v)")
	fmt.Println(Describe(42), "|", Describe("hi"), "|", Describe(3.14))

	title("6. Pointer-method constraint")
	ports, err := ParseAll[Port]([]string{"80", "443"}) // PT được suy ra = *Port
	fmt.Println(ports, err)
	_, err = ParseAll[Port]([]string{"8080", "99999"})
	fmt.Println(err)

	title("7. Chỉ định type argument khi không suy luận được")
	fmt.Printf("%q %v %+v\n", Zero[string](), Zero[int](), Zero[User]())

	title("8. comparable + interface: panic lúc chạy")
	fmt.Println(Equal(1, 1), Equal("a", "b"))
	err = safe(func() { Equal[any]([]int{1}, []int{1}) })
	fmt.Println(err)

	title("9. Iterator + slices/maps/cmp chuẩn")
	users := []User{{"An", 30}, {"Bình", 25}, {"Chi", 35}}
	slices.SortFunc(users, func(a, b User) int { return cmp.Compare(a.Age, b.Age) })
	fmt.Println(users)
	oldest := slices.MaxFunc(users, func(a, b User) int { return cmp.Compare(a.Age, b.Age) })
	fmt.Println("oldest:", oldest.Name, "| index Chi:", slices.IndexFunc(users, func(u User) bool { return u.Name == "Chi" }))
	for n := range FilterSeq(slices.Values(nums), func(n int) bool { return n > 2 }) {
		fmt.Print(n, " ")
	}
	fmt.Println()

	title("10. Generic type alias")
	ages := Dict[int]{"An": 30}
	fmt.Printf("%T %v\n", ages, ages)

	title("11. Constraint tự tham chiếu (Go 1.26)")
	fmt.Println(SumAll(Vec{1, 2}, Vec{3, 4}, Vec{5, 6}))
}
