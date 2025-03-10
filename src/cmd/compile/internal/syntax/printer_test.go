// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package syntax

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func TestPrint(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	ast, _ := ParseFile(*src_, func(err error) { t.Error(err) }, nil, AllowGenerics)

	if ast != nil {
		Fprint(testOut(), ast, LineForm)
		fmt.Println()
	}
}

type shortBuffer struct {
	buf []byte
}

func (w *shortBuffer) Write(data []byte) (n int, err error) {
	w.buf = append(w.buf, data...)
	n = len(data)
	if len(w.buf) > 10 {
		err = io.ErrShortBuffer
	}
	return
}

func TestPrintError(t *testing.T) {
	const src = "package p; var x int"
	ast, err := Parse(nil, strings.NewReader(src), nil, nil, 0)
	if err != nil {
		t.Fatal(err)
	}

	var buf shortBuffer
	_, err = Fprint(&buf, ast, 0)
	if err == nil || err != io.ErrShortBuffer {
		t.Errorf("got err = %s, want %s", err, io.ErrShortBuffer)
	}
}

var stringTests = []string{
	"package p",
	"package p; type _ int; type T1 = struct{}; type ( _ *struct{}; T2 = float32 )",

	// generic type declarations
	"package p; type _[T any] struct{}",
	"package p; type _[A, B, C interface{m()}] struct{}",
	"package p; type _[T any, A, B, C interface{m()}, X, Y, Z interface{~int}] struct{}",

	// generic function declarations
	"package p; func _[T any]()",
	"package p; func _[A, B, C interface{m()}]()",
	"package p; func _[T any, A, B, C interface{m()}, X, Y, Z interface{~int}]()",

	// methods with generic receiver types
	"package p; func (R[T]) _()",
	"package p; func (*R[A, B, C]) _()",
	"package p; func (_ *R[A, B, C]) _()",

	// type constraint literals with elided interfaces
	"package p; func _[P ~int, Q int | string]() {}",
	"package p; func _[P struct{f int}, Q *P]() {}",

	// channels
	"package p; type _ chan chan int",
	"package p; type _ chan (<-chan int)",
	"package p; type _ chan chan<- int",

	"package p; type _ <-chan chan int",
	"package p; type _ <-chan <-chan int",
	"package p; type _ <-chan chan<- int",

	"package p; type _ chan<- chan int",
	"package p; type _ chan<- <-chan int",
	"package p; type _ chan<- chan<- int",

	// TODO(gri) expand
}

func TestPrintString(t *testing.T) {
	for _, want := range stringTests {
		ast, err := Parse(nil, strings.NewReader(want), nil, nil, AllowGenerics)
		if err != nil {
			t.Error(err)
			continue
		}
		if got := String(ast); got != want {
			t.Errorf("%q: got %q", want, got)
		}
	}
}

func testOut() io.Writer {
	if testing.Verbose() {
		return os.Stdout
	}
	return ioutil.Discard
}

func dup(s string) [2]string { return [2]string{s, s} }

var exprTests = [][2]string{
	// basic type literals
	dup("x"),
	dup("true"),
	dup("42"),
	dup("3.1415"),
	dup("2.71828i"),
	dup(`'a'`),
	dup(`"foo"`),
	dup("`bar`"),

	// func and composite literals
	dup("func() {}"),
	dup("[]int{}"),
	{"func(x int) complex128 { return 0 }", "func(x int) complex128 {…}"},
	{"[]int{1, 2, 3}", "[]int{…}"},

	// type expressions
	dup("[1 << 10]byte"),
	dup("[]int"),
	dup("*int"),
	dup("struct{x int}"),
	dup("func()"),
	dup("func(int, float32) string"),
	dup("interface{m()}"),
	dup("interface{m() string; n(x int)}"),
	dup("interface{~int}"),
	dup("interface{~int | ~float64 | ~string}"),
	dup("interface{~int; m()}"),
	dup("interface{~int | ~float64 | ~string; m() string; n(x int)}"),
	dup("map[string]int"),
	dup("chan E"),
	dup("<-chan E"),
	dup("chan<- E"),

	// new interfaces
	dup("interface{int}"),
	dup("interface{~int}"),
	dup("interface{~int}"),
	dup("interface{int | string}"),
	dup("interface{~int | ~string; float64; m()}"),
	dup("interface{~a | ~b | ~c; ~int | ~string; float64; m()}"),
	dup("interface{~T[int, string] | string}"),

	// non-type expressions
	dup("(x)"),
	dup("x.f"),
	dup("a[i]"),

	dup("s[:]"),
	dup("s[i:]"),
	dup("s[:j]"),
	dup("s[i:j]"),
	dup("s[:j:k]"),
	dup("s[i:j:k]"),

	dup("x.(T)"),

	dup("x.([10]int)"),
	dup("x.([...]int)"),

	dup("x.(struct{})"),
	dup("x.(struct{x int; y, z float32; E})"),

	dup("x.(func())"),
	dup("x.(func(x int))"),
	dup("x.(func() int)"),
	dup("x.(func(x, y int, z float32) (r int))"),
	dup("x.(func(a, b, c int))"),
	dup("x.(func(x ...T))"),

	dup("x.(interface{})"),
	dup("x.(interface{m(); n(x int); E})"),
	dup("x.(interface{m(); n(x int) T; E; F})"),

	dup("x.(map[K]V)"),

	dup("x.(chan E)"),
	dup("x.(<-chan E)"),
	dup("x.(chan<- chan int)"),
	dup("x.(chan<- <-chan int)"),
	dup("x.(<-chan chan int)"),
	dup("x.(chan (<-chan int))"),

	dup("f()"),
	dup("f(x)"),
	dup("int(x)"),
	dup("f(x, x + y)"),
	dup("f(s...)"),
	dup("f(a, s...)"),

	dup("*x"),
	dup("&x"),
	dup("x + y"),
	dup("x + y << (2 * s)"),
}

func TestShortString(t *testing.T) {
	for _, test := range exprTests {
		src := "package p; var _ = " + test[0]
		ast, err := Parse(nil, strings.NewReader(src), nil, nil, AllowGenerics)
		if err != nil {
			t.Errorf("%s: %s", test[0], err)
			continue
		}
		x := ast.DeclList[0].(*VarDecl).Values
		if got := String(x); got != test[1] {
			t.Errorf("%s: got %s, want %s", test[0], got, test[1])
		}
	}
}
