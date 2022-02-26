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

/*
Package fang provides a simple and elegant way to bind command line
arguments to struct fields.

For example

	type Person struct {
		Name 	string `shorthand:"n" usage:"person name" fang:"required p"`
		Age 	int
		Gender 	*string
		Body struct {
			Height int `shorthand:"h"`
			Weight float64 `shorthand:"w"`
		}
	}

	var p Person
	fang.Bind(&cobra.Command{}, &p)

Assigned fields in the struct will be used as default values for command line arguments,
fields of pointer type will be automatically initialized to get a zero value as default value.

fang also supports map type binding, but it should be noted that the map keys and values
must be primitive types (such as int float64 string, etc.)

For example

	type Query struct {
		Labels map[string]string `shorthand:"l"`
	}

	var q Query
	fang.Bind(&cobra.Command{}, &q)

	// ./cmdline -l a=b -l c=d

The parameter v passed to the function Bind or Binder.Bind must be a pointer to a struct,
any other type of value will get an error.

Available tags

	* name: customize the full name of this command line argument, the default will use
	  the field name (converted to snake-case format) as the name
	* shorthand: one-letter abbreviated string indicates shorthand of argument in command.
	  fang does not verify the uniqueness of this tag, and spf13/cobra gives an error when
	  there are multi identical abbreviations.
	* usage: one line string indicates help message of argument in command.
	* fang: the extra attribute for fang to binding command line argument. The following
	  attributes can be configured:
		1) persistent, persist, p: meaning arguments should be persisted to subcommands
		2) required, require, r: meaning arguments is required
*/

package fang
