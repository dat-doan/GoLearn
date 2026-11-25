// slice_array_examples.go — ví dụ thực hành đi kèm cheatsheet Slice & Array.
// Chạy: go run slice_array_examples.go
package main

import (
	"fmt"
	"slices"
)

func section(s string) { fmt.Printf("\n=== %s ===\n", s) }

func main() {
	// 1. ARRAY: kích thước là một phần của type, là value type
	section("1. Array là value type")
	a := [3]int{1, 2, 3}
	b := a // copy toàn bộ
	b[0] = 99
	fmt.Println("a:", a, "b:", b) // a không đổi
	fmt.Printf("%T vs %T\n", [3]int{}, [4]int{})
	c := [...]string{"x", "y"} // compiler tự đếm
	fmt.Println(len(c), c == [2]string{"x", "y"}) // array so sánh được bằng ==

	// 2. SLICE header: ptr, len, cap
	section("2. len / cap")
	s := make([]int, 3, 5)
	fmt.Println("len:", len(s), "cap:", cap(s), s)
	s = append(s, 4)
	fmt.Println("append (còn cap):", len(s), cap(s))
	s = append(s, 5, 6) // vượt cap -> cấp phát mảng mới
	fmt.Println("append (vượt cap):", len(s), cap(s))

	// 3. Chia sẻ backing array
	section("3. Shared backing array")
	arr := [5]int{0, 1, 2, 3, 4}
	s1 := arr[1:3] // [1 2], cap = 4
	fmt.Println("s1:", s1, "len:", len(s1), "cap:", cap(s1))
	s1[0] = 100
	fmt.Println("arr sau khi sửa s1:", arr)
	s1 = append(s1, 200) // ghi đè arr[3]!
	fmt.Println("arr sau append (còn cap):", arr)

	// 4. Full slice expression chặn ghi đè
	section("4. s[low:high:max]")
	arr2 := [5]int{0, 1, 2, 3, 4}
	s2 := arr2[1:3:3] // cap = 2
	s2 = append(s2, 200)
	fmt.Println("arr2 giữ nguyên:", arr2, "s2:", s2)

	// 5. Bẫy append trong hàm
	section("5. Truyền slice vào hàm")
	x := []int{1, 2, 3}
	modify(x)
	fmt.Println("sau modify:", x) // phần tử đổi
	appendWrong(x)
	fmt.Println("sau appendWrong:", x) // len không đổi
	x = appendRight(x)
	fmt.Println("sau appendRight:", x)

	// 6. nil vs empty
	section("6. nil vs empty slice")
	var n []int
	e := []int{}
	fmt.Println(n == nil, e == nil, len(n), len(e))
	n = append(n, 1) // nil slice append được
	fmt.Println(n)

	// 7. Copy đúng cách
	section("7. copy / clone")
	src := []int{1, 2, 3}
	dst := make([]int, len(src))
	copy(dst, src)
	dst[0] = 42
	fmt.Println(src, dst)
	cl := slices.Clone(src)
	cl[1] = 7
	fmt.Println(src, cl)

	// 8. Xoá phần tử
	section("8. Xoá phần tử")
	d := []int{10, 20, 30, 40, 50}
	d = slices.Delete(d, 1, 2) // xoá index 1
	fmt.Println(d)
	// xoá nhanh, không giữ thứ tự
	i := 0
	d[i] = d[len(d)-1]
	d = d[:len(d)-1]
	fmt.Println(d)

	// 9. Memory leak: slice nhỏ giữ mảng lớn
	section("9. Giữ mảng lớn")
	big := make([]int, 1_000_000)
	small := big[:2]
	fmt.Println("small cap:", cap(small)) // vẫn giữ 1_000_000
	small = slices.Clone(small)
	fmt.Println("sau Clone cap:", cap(small) >= 2 && cap(small) < 1000)

	// 10. Preallocate
	section("10. Preallocate")
	out := make([]int, 0, 5)
	for i := range 5 {
		out = append(out, i*i)
	}
	fmt.Println(out, len(out), cap(out))

	// 11. Bẫy vòng lặp: range cho bản copy của phần tử
	section("11. range copy")
	type P struct{ N int }
	ps := []P{{1}, {2}}
	for _, p := range ps {
		p.N *= 10 // không ảnh hưởng ps
	}
	fmt.Println(ps)
	for i := range ps {
		ps[i].N *= 10
	}
	fmt.Println(ps)

	// 12. 2D slice
	section("12. Slice 2 chiều")
	grid := make([][]int, 2)
	for i := range grid {
		grid[i] = make([]int, 3)
	}
	grid[1][2] = 5
	fmt.Println(grid)
}

func modify(s []int)      { s[0] = 999 }
func appendWrong(s []int) { s = append(s, 4) }
func appendRight(s []int) []int {
	return append(s, 4)
}
