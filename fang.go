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
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Count is an integer type which is used to count the number of occurrences of
// command-line arguments. Commonly used for Verbose (incremental logging level via -vvv)
type Count int

// BytesHex is a byte array type, which is parsed by hex on the command-line arguments
type BytesHex []byte

var (
	_IPType       = reflect.TypeOf(net.IP{})
	_CountType    = reflect.TypeOf(Count(0))
	_IPNetType    = reflect.TypeOf(net.IPNet{})
	_IPMaskType   = reflect.TypeOf(net.IPMask{})
	_BytesHexType = reflect.TypeOf(BytesHex{})
	_DurationType = reflect.TypeOf(time.Duration(0))
)

// BindError represents an error that occurred during binding
type BindError struct {
	Cause   error
	Message string
	Type    reflect.Type
}

// Error returns a string indicating the error that occurred, which
// will have the type (if provided) and the original error(if provided)
func (e *BindError) Error() string {
	err := "fang: bind error"
	if e.Type != nil {
		err += "(type = " + e.Type.String() + ")"
	}

	err += ": " + e.Message
	if e.Cause != nil {
		err += "; cause by [" + e.Cause.Error() + "]"
	}
	return err
}

// Bind is an alias method, see more details from Binder.Bind
func Bind(cmd *cobra.Command, v interface{}) error {
	b, err := New(cmd)
	if err != nil {
		return err
	}

	return b.Bind(v)
}

// New creates an instance object to bind struct-pointer into cmd.
// cmd cannot be nil and Binder.Bind can be called multiple times, which helps
// to implement the binding of parameters to several struct-value
func New(cmd *cobra.Command) (*Binder, error) {
	if cmd == nil {
		return nil, &BindError{Message: "unable bind value to nil command"}
	}

	return &Binder{cmd: cmd}, nil
}

// Binder holds the cmd and provides a convenient binding method for it
type Binder struct {
	cmd *cobra.Command
}

// Bind traveling all the fields in the struct-pointer and binds
// them to the parameters of the cmd, v and cmd cannot be nil
func (b *Binder) Bind(v interface{}) error {
	if v == nil {
		return &BindError{Message: "unable bind nil value to command"}
	}

	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr {
		return &BindError{Message: "unable bind to non-pointer value", Type: rv.Type()}
	}

	if rv = rv.Elem(); rv.Kind() != reflect.Struct {
		return &BindError{Message: "unsupported type, use struct instead", Type: rv.Type()}
	}

	return b.bindToStruct(rv)
}

// bindToStruct traveling all the fields in the struct and calling the
// appropriate binding method depending on the type
func (b *Binder) bindToStruct(v reflect.Value) error {
	return visitStructField(v, func(field *structField) error {
		switch field.Type {
		case _IPType, _DurationType, _IPNetType, _IPMaskType:
			return b.bindToPrimitive(field.Value)(newInvoker(b, field))
		case _CountType:
			return b.bindToCount(field.Value)(newInvoker(b, field))
		case _BytesHexType:
			return b.bindToBytesHex(field.Value)(newInvoker(b, field))
		}

		switch field.Type.Kind() {
		case reflect.Struct:
			return b.bindToStruct(field.Value)
		case reflect.Array, reflect.Slice:
			return b.bindToSlice(field.Value)(newInvoker(b, field))
		case reflect.Map:
			return b.bindToMap(field.Value)(newInvoker(b, field))
		default:
			return b.bindToPrimitive(field.Value)(newInvoker(b, field))
		}
	})
}

// bindToSlice invoking the binding method depending on the type of the slice-element
func (b *Binder) bindToSlice(v reflect.Value) func(*invoker) error {
	return func(ivk *invoker) error {
		et := v.Type().Elem()
		if et == _IPType {
			return ivk.Invoke(ivk.IPSliceVarP)
		} else if et == _DurationType {
			return ivk.Invoke(ivk.DurationSliceVarP)
		}

		switch et.Kind() {
		case reflect.Bool:
			return ivk.Invoke(ivk.BoolSliceVarP)
		case reflect.Int:
			return ivk.Invoke(ivk.IntSliceVarP)
		case reflect.Uint:
			return ivk.Invoke(ivk.UintSliceVarP)
		case reflect.Int32:
			return ivk.Invoke(ivk.Int32SliceVarP)
		case reflect.Int64:
			return ivk.Invoke(ivk.Int64SliceVarP)
		case reflect.Float32:
			return ivk.Invoke(ivk.Float32SliceVarP)
		case reflect.Float64:
			return ivk.Invoke(ivk.Float64SliceVarP)
		case reflect.String:
			return ivk.Invoke(ivk.StringSliceVarP)
		default:
			return &BindError{Message: "unsupported slice type", Type: v.Type()}
		}
	}
}

// bindToMap invoking the binding method with customized mapValue type
func (b *Binder) bindToMap(v reflect.Value) func(*invoker) error {
	return func(ivk *invoker) error {
		return ivk.WithInvoke(func(f *structField) error {
			m, err := newMapValue(v)
			if err != nil {
				return err
			}

			ivk.VarPF(m, f.Name(), f.Shorthand(), f.Usage())
			return nil
		})
	}
}

// bindToPrimitive invoking the binding method depending on the primitive type
func (b *Binder) bindToPrimitive(v reflect.Value) func(*invoker) error {
	return func(ivk *invoker) error {
		switch v.Type() {
		case _IPType:
			return ivk.Invoke(ivk.IPVarP)
		case _IPNetType:
			return ivk.Invoke(ivk.IPNetVarP)
		case _IPMaskType:
			return ivk.Invoke(ivk.IPMaskVarP)
		case _DurationType:
			return ivk.Invoke(ivk.DurationVarP)
		}

		switch v.Kind() {
		case reflect.Bool:
			return ivk.Invoke(ivk.BoolVarP)
		case reflect.Int:
			return ivk.Invoke(ivk.IntVarP)
		case reflect.Int8:
			return ivk.Invoke(ivk.Int8VarP)
		case reflect.Int16:
			return ivk.Invoke(ivk.Int16VarP)
		case reflect.Int32:
			return ivk.Invoke(ivk.Int32VarP)
		case reflect.Int64:
			return ivk.Invoke(ivk.Int64VarP)
		case reflect.Uint:
			return ivk.Invoke(ivk.UintVarP)
		case reflect.Uint8:
			return ivk.Invoke(ivk.Uint8VarP)
		case reflect.Uint16:
			return ivk.Invoke(ivk.Uint16VarP)
		case reflect.Uint32:
			return ivk.Invoke(ivk.Uint32VarP)
		case reflect.Uint64:
			return ivk.Invoke(ivk.Uint64VarP)
		case reflect.Float32:
			return ivk.Invoke(ivk.Float32VarP)
		case reflect.Float64:
			return ivk.Invoke(ivk.Float64VarP)
		case reflect.String:
			return ivk.Invoke(ivk.StringVarP)
		default:
			return &BindError{Message: "unsupported type of field", Type: v.Type()}
		}
	}
}

// bindToCount invoking the binding method on Count type
func (b *Binder) bindToCount(v reflect.Value) func(*invoker) error {
	return func(ivk *invoker) error {
		return ivk.WithInvoke(func(f *structField) error {
			ivk.CountVarP((*int)(v.Addr().Interface().(*Count)), f.Name(), f.Shorthand(), f.Usage())
			return nil
		})
	}
}

// bindToBytesHex invoking the binding method on BytesHex type
func (b *Binder) bindToBytesHex(v reflect.Value) func(*invoker) error {
	return func(ivk *invoker) error {
		return ivk.WithInvoke(func(f *structField) error {
			ivk.BytesHexVarP((*[]byte)(v.Addr().Interface().(*BytesHex)), f.Name(), f.Shorthand(),
				f.Value.Interface().(BytesHex), f.Usage())
			return nil
		})
	}
}

// invoker holds pflag.FlagSet and structField and performed actual binding
type invoker struct {
	*pflag.FlagSet

	cmd   *cobra.Command
	field *structField
}

// Invoke invokes fVarP by reflection and add some simple verification
func (ivk *invoker) Invoke(fVarP interface{}) error {
	rVarP := reflect.ValueOf(fVarP)
	if rVarP.Kind() != reflect.Func {
		return &BindError{Message: "internal error for binding invoke", Type: rVarP.Type()}
	}

	return ivk.WithInvoke(func(f *structField) error {
		rVarP.Call([]reflect.Value{
			f.Value.Addr(),                 // pointer
			reflect.ValueOf(f.Name()),      // name
			reflect.ValueOf(f.Shorthand()), // shorthand
			f.Value,                        // default value
			reflect.ValueOf(f.Usage()),     // usage
		})
		return nil
	})
}

// WithInvoke invokes handler customized binding and add some simple verification
func (ivk *invoker) WithInvoke(handler func(field *structField) error) (err error) {
	defer func() {
		if v := recover(); v != nil {
			if e, ok := v.(error); ok {
				err = &BindError{Message: "internal error", Cause: e}
			}
			if s, ok := v.(string); ok {
				err = &BindError{Message: "internal error", Cause: errors.New(s)}
			}
			panic(v)
		}
	}()

	if err = handler(ivk.field); err != nil {
		if be, ok := err.(*BindError); ok {
			return be
		}
		return &BindError{Message: "internal error", Cause: err}
	}

	if ivk.field.Required() {
		if ivk.field.Persistent() {
			return ivk.cmd.MarkPersistentFlagRequired(ivk.field.Name())
		} else {
			return ivk.cmd.MarkFlagRequired(ivk.field.Name())
		}
	}
	return
}

// newInvoker creates invoker instance and extract the pflag.FlagSet
// according to whether the attr-persistent
func newInvoker(b *Binder, field *structField) *invoker {
	i := &invoker{cmd: b.cmd, field: field, FlagSet: b.cmd.Flags()}
	if field.Persistent() {
		i.FlagSet = b.cmd.PersistentFlags()
	}

	return i
}

// visitStructField calling the visit method for each exported field of the structure
// If the return value of the visit method is not nil, will return this error directly and exit
// The parameter v must the reflection interface of a struct value
func visitStructField(v reflect.Value, visit func(field *structField) error) error {
	for i, t := 0, v.Type(); i < v.NumField(); i++ {
		if fv := v.Field(i); fv.CanSet() && fv.CanAddr() {
			if err := visit(newStructField(t.Field(i), fv)); err != nil {
				return err
			}
		}
	}
	return nil
}

// structField represents a field in struct
type structField struct {
	Type  reflect.Type
	Value reflect.Value
	Field reflect.StructField
}

// Name returns snake-case string indicates name of the field
// The name of the field will be used by default, and can be customized using the `name` tag
func (f *structField) Name() string {
	if name, ok := f.Field.Tag.Lookup("name"); ok && len(name) != 0 {
		return name
	}
	return toSnakeCase(f.Field.Name)
}

// Shorthand returns one-letter abbreviated string indicates shorthand of argument in command
// The default value is empty(meaning no shorthand), and can be customized using the `shorthand` tag
func (f *structField) Shorthand() string {
	return f.Field.Tag.Get("shorthand")
}

// Usage returns one line string indicates help message of argument in command
// The default value is empty(no help message), and can be customized using the `usage` tag
func (f *structField) Usage() string {
	return f.Field.Tag.Get("usage")
}

// Persistent returns a boolean value indicating whether the command line argument
// should be `persist` or not, see more details from cobra.Command, and can be customized
// using the `fang` tag with some of `persistent`, `persist` or `p` values
func (f *structField) Persistent() bool {
	for _, attr := range f.attrs() {
		switch attr {
		case "persistent", "persist", "p":
			return true
		}
	}
	return false
}

// Required returns a boolean value indicating whether this command line argument is required
func (f *structField) Required() bool {
	for _, attr := range f.attrs() {
		switch attr {
		case "required", "require", "r":
			return true
		}
	}
	return false
}

// attrs returns a list of the string indicates the extra attribute for command line argument
func (f *structField) attrs() []string {
	return strings.FieldsFunc(f.Field.Tag.Get("fang"), func(r rune) bool {
		return r == ',' || r == ' '
	})
}

// newStructField creates a structField instance to keep the field type and value
// fields of pointer type are automatically created as default value depending on
// whether they are nil or not and are converted to uniform non-pointer types
func newStructField(f reflect.StructField, v reflect.Value) *structField {
	field := &structField{Type: f.Type, Value: v, Field: f}
	if field.Type.Kind() == reflect.Ptr {
		field.Type = field.Type.Elem()
		if field.Value.IsNil() && field.Value.CanSet() {
			field.Value.Set(reflect.New(field.Type))
		}
		field.Value = field.Value.Elem()
	} else if field.Type.Kind() == reflect.Map {
		field.Value.Set(reflect.MakeMap(field.Type))
	}

	return field
}

// toSnakeCase returns a string from camel-case to snake-case
// This function will only convert the uppercase letters (except the first letter) to
// the corresponding lower case form and add the midline in front(A -> -a).
// No changes will be made to other symbols such as underscores(_) or numbers
func toSnakeCase(s string) string {
	var buf bytes.Buffer

	buf.WriteRune(unicode.ToLower(rune(s[0])))
	for _, r := range s[1:] {
		if unicode.IsUpper(r) {
			buf.WriteRune('-')
		}
		buf.WriteRune(unicode.ToLower(r))
	}

	return buf.String()
}

// mapValue represents a map value on command line
type mapValue struct {
	Key  reflect.Type
	Elem reflect.Type

	Value reflect.Value
}

// String returns a string indicates default value
// for this command line argument
func (m *mapValue) String() string {
	data, err := json.Marshal(m.Value.Interface())
	if err != nil {
		panic(err)
	}
	return string(data)
}

// Set sets a command line argument into map
func (m *mapValue) Set(arg string) (err error) {
	kv := strings.SplitN(arg, "=", 2)
	if len(kv) != 2 {
		return &BindError{Message: "invalid key-value pair format, key=value"}
	}

	var key, value interface{}
	if key, err = newPrimitiveValue(m.Key, kv[0]); err != nil {
		return &BindError{Message: fmt.Sprintf("unexpected map key %q", kv[0]), Type: m.Key, Cause: err}
	}
	if value, err = newPrimitiveValue(m.Elem, kv[1]); err != nil {
		return &BindError{Message: fmt.Sprintf("unexpected map value %q", kv[0]), Type: m.Key, Cause: err}
	}

	m.Value.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(value))
	return
}

// Type returns a string indicates type of command line argument
func (m *mapValue) Type() string {
	return m.Value.Type().String()
}

// newMapValue creates a customized pflag.Value to binding map
func newMapValue(v reflect.Value) (pflag.Value, error) {
	m := &mapValue{Key: v.Type().Key(), Elem: v.Type().Elem(), Value: v}

	switch m.Key.Kind() {
	case reflect.Chan, reflect.Array, reflect.Struct, reflect.Ptr, reflect.UnsafePointer, reflect.Uintptr:
		return nil, &BindError{Message: "unsupported type of map key", Type: m.Key}
	}

	switch m.Elem.Kind() {
	case reflect.Chan, reflect.Array, reflect.Struct, reflect.Ptr, reflect.UnsafePointer, reflect.Uintptr:
		return nil, &BindError{Message: "unsupported type of map value", Type: m.Key}
	}

	return m, nil
}

// newPrimitiveValue creates primitive value by reflection
func newPrimitiveValue(t reflect.Type, s string) (interface{}, error) {
	switch t.Kind() {
	case reflect.Bool:
		b, err := strconv.ParseBool(s)
		return b, err
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return nil, err
		}

		if v := reflect.New(t).Elem(); v.OverflowInt(n) {
			return nil, errors.New("number overflow")
		}

		switch t.Kind() {
		case reflect.Int8:
			return int8(n), nil
		case reflect.Int16:
			return int16(n), nil
		case reflect.Int32:
			return int32(n), nil
		case reflect.Int:
			return int(n), nil
		default:
			return n, nil
		}
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
		n, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return nil, err
		}

		if v := reflect.New(t).Elem(); v.OverflowUint(n) {
			return nil, errors.New("unsigned number overflow")
		}

		switch t.Kind() {
		case reflect.Uint8:
			return uint8(n), nil
		case reflect.Uint16:
			return uint16(n), nil
		case reflect.Uint32:
			return uint32(n), nil
		case reflect.Uint:
			return uint(n), nil
		default:
			return n, nil
		}
	case reflect.Float32, reflect.Float64:
		n, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return nil, err
		}

		if v := reflect.ValueOf(t).Elem(); v.OverflowFloat(n) {
			return nil, errors.New("float number overflow")
		}

		if t.Kind() == reflect.Float32 {
			return float32(n), nil
		}
		return n, nil
	}
	return s, nil
}
