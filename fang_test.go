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
	"net"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestBind_InvalidValue(t *testing.T) {
	var value struct{}

	assert.Error(t, Bind(nil, nil))
	assert.Error(t, Bind(nil, &value))
	assert.Error(t, Bind(&cobra.Command{}, nil))
	assert.Error(t, Bind(&cobra.Command{}, value))

	var number int
	assert.Error(t, Bind(&cobra.Command{}, &number))
}

func TestBind_PointerValue(t *testing.T) {
	var value struct {
		Boolean *bool
	}

	if b, err := New(&cobra.Command{}); assert.NoError(t, err) {
		if err = b.Bind(&value); assert.NoError(t, err) {
			if err = b.cmd.ParseFlags([]string{"--boolean", "yes"}); assert.NoError(t, err) {
				assert.Equal(t, true, *value.Boolean)
			}
		}
	}
}

func TestBind_NestedStruct(t *testing.T) {
	var value struct {
		Nested struct {
			Number int
		}
	}

	if b, err := New(&cobra.Command{}); assert.NoError(t, err) {
		if err = b.Bind(&value); assert.NoError(t, err) {
			if err = b.cmd.ParseFlags([]string{"--number", "1123456"}); assert.NoError(t, err) {
				assert.Equal(t, 1123456, value.Nested.Number)
			}
		}
	}
}

func TestBind_IPType(t *testing.T) {
	var value struct {
		IP net.IP `name:"ip"`
	}

	if b, err := New(&cobra.Command{}); assert.NoError(t, err) {
		if err = b.Bind(&value); assert.NoError(t, err) {
			if err = b.cmd.ParseFlags([]string{"--ip", "192.168.1.1"}); assert.NoError(t, err) {
				assert.Equal(t, net.ParseIP("192.168.1.1"), value.IP)
			}
		}
	}
}

func TestBind_DurationType(t *testing.T) {
	var value struct {
		Duration time.Duration
	}

	if b, err := New(&cobra.Command{}); assert.NoError(t, err) {
		if err = b.Bind(&value); assert.NoError(t, err) {
			if expected, err := time.ParseDuration("1h2m3s"); assert.NoError(t, err) {
				if err = b.cmd.ParseFlags([]string{"--duration", expected.String()}); assert.NoError(t, err) {
					assert.Equal(t, expected, value.Duration)
				}
			}
		}
	}
}

func TestBind_IPSlice(t *testing.T) {
	var value struct {
		IPs []net.IP `name:"ips"`
	}

	if b, err := New(&cobra.Command{}); assert.NoError(t, err) {
		if err = b.Bind(&value); assert.NoError(t, err) {
			args := []string{"--ips", "192.168.1.1", "--ips", "192.168.1.2"}
			if err = b.cmd.ParseFlags(args); assert.NoError(t, err) {
				assert.Equal(t, net.ParseIP("192.168.1.1"), value.IPs[0])
				assert.Equal(t, net.ParseIP("192.168.1.2"), value.IPs[1])
			}
		}
	}
}

func TestBind_DurationSlice(t *testing.T) {
	var value struct {
		Durations []time.Duration
	}

	if b, err := New(&cobra.Command{}); assert.NoError(t, err) {
		if err = b.Bind(&value); assert.NoError(t, err) {
			if expected, err := time.ParseDuration("1h2m3s"); assert.NoError(t, err) {
				args := []string{"--durations", expected.String(), "--durations", expected.String()}
				if err = b.cmd.ParseFlags(args); assert.NoError(t, err) {
					assert.Equal(t, expected, value.Durations[0])
					assert.Equal(t, expected, value.Durations[1])
				}
			}
		}
	}
}

func TestBind_MapValue(t *testing.T) {
	var value struct {
		Scores map[string]int `shorthand:"s"`
	}

	if b, err := New(&cobra.Command{}); assert.NoError(t, err) {
		if err = b.Bind(&value); assert.NoError(t, err) {
			args := []string{"-s", "a=1", "-s", "b=2"}
			if err = b.cmd.ParseFlags(args); assert.NoError(t, err) {
				if v, ok := value.Scores["a"]; assert.True(t, ok) {
					assert.Equal(t, 1, v)
				}
				if v, ok := value.Scores["b"]; assert.True(t, ok) {
					assert.Equal(t, 2, v)
				}
			}
		}
	}
}

func TestBind_toSnakeCase(t *testing.T) {
	table := []struct {
		CamelCase string
		SnakeCase string
	}{
		{CamelCase: "ILoveYou", SnakeCase: "i-love-you"},
		{CamelCase: "HELLO", SnakeCase: "h-e-l-l-o"},
		{CamelCase: "Hi0_1-2AxxBC", SnakeCase: "hi0_1-2-axx-b-c"},
	}

	for _, item := range table {
		assert.Equal(t, item.SnakeCase, toSnakeCase(item.CamelCase))
	}
}
