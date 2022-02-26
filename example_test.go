// Copyright (c) 2022 Jayson Wang
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package fang

import (
	"fmt"

	"github.com/spf13/cobra"
)

func ExampleBind() {
	kubectl := &cobra.Command{}
	var kubernetes struct {
		File      *string `shorthand:"f" usage:"that contains the configuration to apply"`
		Namespace string  `shorthand:"n" fang:"If present, the namespace scope for this CLI request"`
	}

	if err := Bind(kubectl, &kubernetes); err != nil {
		panic(err)
	}

	if err := kubectl.ParseFlags([]string{"-n", "app", "-f", "pod.yaml"}); err != nil {
		panic(err)
	}

	fmt.Println(kubernetes.Namespace, *kubernetes.File)

	// Output:
	// app pod.yaml
}

func ExampleBinder_Bind() {
	kubectl := &cobra.Command{}
	var global struct {
		Namespace string `shorthand:"n" fang:"If present, the namespace scope for this CLI request"`
	}
	var apply struct {
		File *string `shorthand:"f" usage:"that contains the configuration to apply"`
	}

	b, err := New(kubectl)
	if err != nil {
		panic(err)
	}

	if err = b.Bind(&global); err != nil {
		panic(err)
	}
	if err = b.Bind(&apply); err != nil {
		panic(err)
	}

	if err = kubectl.ParseFlags([]string{"-n", "app", "-f", "pod.yaml"}); err != nil {
		panic(err)
	}

	fmt.Println(global.Namespace, *apply.File)

	// Output:
	// app pod.yaml
}
