// interface_examples.go — ví dụ thực hành đi kèm cheatsheet Interface (docs/interface.html).
// Chạy: go run interface_examples.go
package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

func title(s string) { fmt.Printf("\n=== %s ===\n", s) }

// ---------------------------------------------------------------------------
// 1. Khai báo & thỏa mãn NGẦM ĐỊNH (không có từ khóa implements)
// ---------------------------------------------------------------------------

type Speaker interface {
	Speak() string
}

type Dog struct{ Name string }

func (d Dog) Speak() string { return d.Name + ": Gâu!" } // value receiver

type Robot struct{ ID int }

func (r *Robot) Speak() string { return fmt.Sprintf("R-%d: Bíp!", r.ID) } // pointer receiver

// Kiểm tra lúc compile: nếu Dog / *Robot thiếu method, build sẽ fail ngay tại đây.
var (
	_ Speaker = Dog{}
	_ Speaker = (*Robot)(nil)
	// _ Speaker = Robot{} // ✗ compile error: method Speak has pointer receiver
)

// ---------------------------------------------------------------------------
// 3. Bẫy nil: interface chứa con trỏ nil KHÁC interface nil
// ---------------------------------------------------------------------------

type MyErr struct{ Msg string }

func (e *MyErr) Error() string { return e.Msg }

// ✗ SAI: trả về biến có kiểu cụ thể *MyErr -> interface error = (*MyErr, nil) != nil
func validateBad(ok bool) error {
	var p *MyErr
	if !ok {
		p = &MyErr{"dữ liệu sai"}
	}
	return p
}

// ✓ ĐÚNG: trả về nil tường minh
func validateGood(ok bool) error {
	if !ok {
		return &MyErr{"dữ liệu sai"}
	}
	return nil
}

// ---------------------------------------------------------------------------
// 4. Type switch
// ---------------------------------------------------------------------------

func describe(v any) string {
	switch x := v.(type) {
	case nil:
		return "nil"
	case int, int64: // nhiều type trong 1 case -> x vẫn có kiểu any
		return fmt.Sprintf("số nguyên %v", x)
	case string:
		return fmt.Sprintf("chuỗi dài %d", len(x)) // x là string
	case error: // đặt error TRƯỚC Stringer/Speaker nếu type có thể thỏa cả hai
		return "error: " + x.Error()
	case Speaker:
		return "biết nói -> " + x.Speak()
	default:
		return fmt.Sprintf("không rõ (%T)", x)
	}
}

// ---------------------------------------------------------------------------
// 5. io.Writer, embedding interface, optional interface
// ---------------------------------------------------------------------------

type upperWriter struct{ w io.Writer }

func (u upperWriter) Write(p []byte) (int, error) {
	return u.w.Write([]byte(strings.ToUpper(string(p))))
}

func capabilities(w io.Writer) {
	_, isSW := w.(io.StringWriter) // "nâng cấp" nếu type hỗ trợ thêm
	_, isCloser := w.(io.Closer)
	fmt.Printf("%-20T StringWriter=%-5v Closer=%v\n", w, isSW, isCloser)
}

// Nhúng interface vào struct: chỉ override 1 method, phần còn lại ủy quyền.
type loudSpeaker struct {
	Speaker // field nhúng; nếu để nil mà gọi Speak() -> panic
}

func (l loudSpeaker) Speak() string { return strings.ToUpper(l.Speaker.Speak()) }

// ---------------------------------------------------------------------------
// 6. So sánh interface có thể PANIC
// ---------------------------------------------------------------------------

func safeEqual(a, b any) (eq bool, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()
	return a == b, nil
}

// ---------------------------------------------------------------------------
// 7. fmt.Stringer
// ---------------------------------------------------------------------------

type Celsius float64

func (c Celsius) String() string { return fmt.Sprintf("%.1f°C", float64(c)) }

// ✗ Đừng: func (c Celsius) String() string { return fmt.Sprint(c) } -> đệ quy vô hạn

// ---------------------------------------------------------------------------
// 8. Interface phía CONSUMER + fake để test (Dependency Injection)
// ---------------------------------------------------------------------------

type User struct {
	ID   int
	Name string
}

var ErrNotFound = errors.New("not found")

// --- "package store": trả về struct cụ thể, KHÔNG định nghĩa interface ---
type MemStore struct{ data map[int]User }

func NewMemStore() *MemStore { return &MemStore{data: map[int]User{1: {1, "An"}}} }

func (m *MemStore) GetUser(id int) (User, error) {
	u, ok := m.data[id]
	if !ok {
		return User{}, fmt.Errorf("user %d: %w", id, ErrNotFound)
	}
	return u, nil
}
func (m *MemStore) SaveUser(u User) error { m.data[u.ID] = u; return nil }
func (m *MemStore) DeleteUser(id int)     { delete(m.data, id) }

// --- "package greet": tự khai báo ĐÚNG những gì mình cần (1 method) ---
type UserGetter interface {
	GetUser(id int) (User, error)
}

type Greeter struct{ users UserGetter }

func (g Greeter) Greet(id int) string {
	u, err := g.users.GetUser(id)
	if errors.Is(err, ErrNotFound) {
		return "Xin chào người lạ"
	}
	if err != nil {
		return "lỗi: " + err.Error()
	}
	return "Xin chào " + u.Name
}

// --- fake trong test: chỉ cần 1 method, không cần mock framework ---
type fakeUsers map[int]User

func (f fakeUsers) GetUser(id int) (User, error) {
	if u, ok := f[id]; ok {
		return u, nil
	}
	return User{}, ErrNotFound
}

// ---------------------------------------------------------------------------
// 9. Func adapter (mẫu http.HandlerFunc) + middleware
// ---------------------------------------------------------------------------

type Handler interface{ Handle(msg string) string }

type HandlerFunc func(string) string

func (f HandlerFunc) Handle(m string) string { return f(m) }

func logging(next Handler) Handler {
	return HandlerFunc(func(m string) string {
		fmt.Println("  [log] nhận:", m)
		return next.Handle(m)
	})
}

// ---------------------------------------------------------------------------
// 11. Interface làm constraint cho generics (type set)
// ---------------------------------------------------------------------------

type Number interface {
	~int | ~int64 | ~float64 // chỉ dùng được làm constraint, KHÔNG làm kiểu biến
}

func Sum[T Number](xs []T) T {
	var total T
	for _, x := range xs {
		total += x
	}
	return total
}

func main() {
	title("1. Thỏa mãn ngầm định")
	for _, s := range []Speaker{Dog{"Rex"}, &Robot{7}} {
		fmt.Println(s.Speak())
	}

	title("2. Dynamic type + dynamic value")
	var s Speaker
	fmt.Printf("zero value: %v, %T, s==nil: %v\n", s, s, s == nil)
	d := Dog{"Rex"}
	s = d // COPY giá trị d vào interface
	d.Name = "Max"
	fmt.Printf("%T -> %s (không bị đổi thành Max)\n", s, s.Speak())
	s = &Robot{1}
	fmt.Printf("%T -> %s\n", s, s.Speak())

	title("3. Bẫy nil interface")
	errBad := validateBad(true)
	fmt.Printf("validateBad:  err==nil? %-5v type=%T\n", errBad == nil, errBad)
	errGood := validateGood(true)
	fmt.Printf("validateGood: err==nil? %-5v type=%T\n", errGood == nil, errGood)
	var me *MyErr
	fmt.Println("errors.As tìm *MyErr:", errors.As(validateGood(false), &me), me.Msg)

	title("4. Type assertion & type switch")
	var x any = 42
	n, ok := x.(int)
	fmt.Println("x.(int)    ->", n, ok)
	str, ok := x.(string) // có ok -> không panic, str = zero value
	fmt.Printf("x.(string) -> %q %v\n", str, ok)
	for _, v := range []any{nil, 7, "hello", Dog{"Rex"}, errors.New("hỏng"), 3.14} {
		fmt.Println(" ", describe(v))
	}

	title("5. io.Writer, optional interface, embedding")
	io.Copy(upperWriter{os.Stdout}, strings.NewReader("hello interface\n"))
	capabilities(os.Stdout)
	capabilities(&strings.Builder{})
	capabilities(upperWriter{})
	fmt.Println(loudSpeaker{Dog{"Rex"}}.Speak())

	title("6. So sánh interface")
	var a, b any = 1, 1
	fmt.Println("1 == 1:", a == b)
	fmt.Println("int(1) == int64(1):", a == any(int64(1)), "(khác dynamic type)")
	fmt.Println("Dog == Dog:", any(Dog{"Rex"}) == any(Dog{"Rex"}))
	_, err := safeEqual([]int{1}, []int{1})
	fmt.Println("[]int == []int ->", err)

	title("7. Stringer")
	fmt.Println("Nhiệt độ:", Celsius(36.6))
	fmt.Printf("%v | %s | %.2f\n", Celsius(20), Celsius(20), Celsius(20))

	title("8. Interface phía consumer + fake")
	fmt.Println("prod:", Greeter{NewMemStore()}.Greet(1)) // *MemStore thỏa UserGetter
	fmt.Println("prod:", Greeter{NewMemStore()}.Greet(99))
	fmt.Println("test:", Greeter{fakeUsers{5: {5, "Bình"}}}.Greet(5))

	title("9. HandlerFunc adapter + middleware")
	echo := HandlerFunc(func(m string) string { return "echo " + m })
	fmt.Println(logging(echo).Handle("ping"))

	title("10. []Dog không tự chuyển thành []Speaker")
	dogs := []Dog{{"A"}, {"B"}}
	// var ss []Speaker = dogs // ✗ compile error: cannot use dogs ([]Dog) as []Speaker
	ss := make([]Speaker, len(dogs))
	for i, dg := range dogs {
		ss[i] = dg
	}
	fmt.Println(len(ss), ss[1].Speak())

	title("11. Interface làm constraint (generics)")
	fmt.Println(Sum([]int{1, 2, 3}), Sum([]float64{1.5, 2.5}), Sum([]Celsius{10, 20.5}))
}
