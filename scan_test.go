// Copyright 2012 Gary Burd
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package ssgo

import (
	"reflect"
	"testing"
)

type s0 struct {
	X  int
	Y  int `ssgo:"y"`
	Bt bool
}

type s1 struct {
	X  int    `ssgo:"-"`
	I  int    `ssgo:"i"`
	U  uint   `ssgo:"u"`
	S  string `ssgo:"s"`
	P  []byte `ssgo:"p"`
	B  bool   `ssgo:"b"`
	Bt bool
	Bf bool
	s0
}

var scanStructTests = []struct {
	title string
	reply []string
	value interface{}
}{
	{"basic",
		[]string{"i", "-1234", "u", "5678", "s", "hello", "p", "world", "b", "t", "Bt", "1", "Bf", "0", "X", "123", "y", "456"},
		&s1{I: -1234, U: 5678, S: "hello", P: []byte("world"), B: true, Bt: true, Bf: false, s0: s0{X: 123, Y: 456}},
	},
}

func TestScanStruct(t *testing.T) {
	for _, tt := range scanStructTests {

		value := reflect.New(reflect.ValueOf(tt.value).Type().Elem())

		if err := ScanStruct(tt.reply, value.Interface()); err != nil {
			t.Fatalf("ScanStruct(%s) returned error %v", tt.title, err)
		}

		if !reflect.DeepEqual(value.Interface(), tt.value) {
			t.Fatalf("ScanStruct(%s) returned %v, want %v", tt.title, value.Interface(), tt.value)
		}
	}
}

func TestBadScanStructArgs(t *testing.T) {
	x := []string{"A", "b"}
	test := func(v interface{}) {
		if err := ScanStruct(x, v); err == nil {
			t.Errorf("Expect error for ScanStruct(%T, %T)", x, v)
		}
	}

	test(nil)

	var v0 *struct{}
	test(v0)

	var v1 int
	test(&v1)

	x = x[:1]
	v2 := struct{ A string }{}
	test(&v2)
}
