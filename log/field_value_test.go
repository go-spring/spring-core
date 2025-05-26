/*
 * Copyright 2025 The Go-Spring Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package log

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"testing"

	"github.com/lvan100/go-assert"
)

var _ Encoder = (*MockEncoder)(nil)

type MockEncoder struct {
	AppendEncoderBeginFunc func() error
	AppendEncoderEndFunc   func() error
	AppendObjectBeginFunc  func() error
	AppendObjectEndFunc    func() error
	AppendArrayBeginFunc   func() error
	AppendArrayEndFunc     func() error
	AppendKeyFunc          func(key string) error
	AppendBoolFunc         func(v bool) error
	AppendInt64Func        func(v int64) error
	AppendUint64Func       func(v uint64) error
	AppendFloat64Func      func(v float64) error
	AppendStringFunc       func(v string) error
	AppendReflectFunc      func(v interface{}) error
}

func (m MockEncoder) AppendEncoderBegin() error {
	if m.AppendEncoderBeginFunc != nil {
		return m.AppendEncoderBeginFunc()
	}
	panic("AppendEncoderBeginFunc is nil")
}

func (m MockEncoder) AppendEncoderEnd() error {
	if m.AppendEncoderEndFunc != nil {
		return m.AppendEncoderEndFunc()
	}
	panic("AppendEncoderEndFunc is nil")
}

func (m MockEncoder) AppendObjectBegin() error {
	if m.AppendObjectBeginFunc != nil {
		return m.AppendObjectBeginFunc()
	}
	panic("AppendObjectBeginFunc is nil")
}

func (m MockEncoder) AppendObjectEnd() error {
	if m.AppendObjectEndFunc != nil {
		return m.AppendObjectEndFunc()
	}
	panic("AppendObjectEndFunc is nil")
}

func (m MockEncoder) AppendArrayBegin() error {
	if m.AppendArrayBeginFunc != nil {
		return m.AppendArrayBeginFunc()
	}
	panic("AppendArrayBeginFunc is nil")
}

func (m MockEncoder) AppendArrayEnd() error {
	if m.AppendArrayEndFunc != nil {
		return m.AppendArrayEndFunc()
	}
	panic("AppendArrayEndFunc is nil")
}

func (m MockEncoder) AppendKey(key string) error {
	if m.AppendKeyFunc != nil {
		return m.AppendKeyFunc(key)
	}
	panic("AppendKeyFunc is nil")
}

func (m MockEncoder) AppendBool(v bool) error {
	if m.AppendBoolFunc != nil {
		return m.AppendBoolFunc(v)
	}
	panic("AppendBoolFunc is nil")
}

func (m MockEncoder) AppendInt64(v int64) error {
	if m.AppendInt64Func != nil {
		return m.AppendInt64Func(v)
	}
	panic("AppendInt64Func is nil")
}

func (m MockEncoder) AppendUint64(v uint64) error {
	if m.AppendUint64Func != nil {
		return m.AppendUint64Func(v)
	}
	panic("AppendUint64Func is nil")
}

func (m MockEncoder) AppendFloat64(v float64) error {
	if m.AppendFloat64Func != nil {
		return m.AppendFloat64Func(v)
	}
	panic("AppendFloat64Func is nil")
}

func (m MockEncoder) AppendString(v string) error {
	if m.AppendStringFunc != nil {
		return m.AppendStringFunc(v)
	}
	panic("AppendStringFunc is nil")
}

func (m MockEncoder) AppendReflect(v interface{}) error {
	if m.AppendReflectFunc != nil {
		return m.AppendReflectFunc(v)
	}
	panic("AppendReflectFunc is nil")
}

func TestBoolValue(t *testing.T) {

	t.Run("error", func(t *testing.T) {
		enc := MockEncoder{AppendBoolFunc: func(v bool) error {
			return errors.New("encoder error")
		}}

		err := BoolValue(true).Encode(enc)
		assert.ThatError(t, err).Matches("encoder error")

		err = BoolValue(false).Encode(enc)
		assert.ThatError(t, err).Matches("encoder error")
	})

	t.Run("success", func(t *testing.T) {
		var buf bytes.Buffer
		enc := MockEncoder{AppendBoolFunc: func(v bool) error {
			_, err := buf.WriteString(strconv.FormatBool(v))
			return err
		}}

		err := BoolValue(true).Encode(enc)
		assert.Nil(t, err)
		assert.ThatString(t, buf.String()).Equal("true")

		buf.Reset()
		err = BoolValue(false).Encode(enc)
		assert.Nil(t, err)
		assert.ThatString(t, buf.String()).Equal("false")
	})
}

func TestInt64Value(t *testing.T) {

	t.Run("error", func(t *testing.T) {
		enc := MockEncoder{AppendInt64Func: func(v int64) error {
			return errors.New("encoder error")
		}}

		err := Int64Value(9).Encode(enc)
		assert.ThatError(t, err).Matches("encoder error")
	})

	t.Run("success", func(t *testing.T) {
		var buf bytes.Buffer
		enc := MockEncoder{AppendInt64Func: func(v int64) error {
			_, err := buf.WriteString(strconv.FormatInt(v, 10))
			return err
		}}

		err := Int64Value(9).Encode(enc)
		assert.Nil(t, err)
		assert.ThatString(t, buf.String()).Equal("9")
	})
}

func TestUint64Value(t *testing.T) {

	t.Run("error", func(t *testing.T) {
		enc := MockEncoder{AppendUint64Func: func(v uint64) error {
			return errors.New("encoder error")
		}}

		err := Uint64Value(9).Encode(enc)
		assert.ThatError(t, err).Matches("encoder error")
	})

	t.Run("success", func(t *testing.T) {
		var buf bytes.Buffer
		enc := MockEncoder{AppendUint64Func: func(v uint64) error {
			_, err := buf.WriteString(strconv.FormatUint(v, 10))
			return err
		}}

		err := Uint64Value(9).Encode(enc)
		assert.Nil(t, err)
		assert.ThatString(t, buf.String()).Equal("9")
	})
}

func TestFloat64Value(t *testing.T) {

	t.Run("error", func(t *testing.T) {
		enc := MockEncoder{AppendFloat64Func: func(v float64) error {
			return errors.New("encoder error")
		}}

		err := Float64Value(9.9).Encode(enc)
		assert.ThatError(t, err).Matches("encoder error")
	})

	t.Run("success", func(t *testing.T) {
		var buf bytes.Buffer
		enc := MockEncoder{AppendFloat64Func: func(v float64) error {
			_, err := buf.WriteString(strconv.FormatFloat(v, 'f', -1, 64))
			return err
		}}

		err := Float64Value(9.9).Encode(enc)
		assert.Nil(t, err)
		assert.ThatString(t, buf.String()).Equal("9.9")
	})
}

func TestStringValue(t *testing.T) {

	t.Run("error", func(t *testing.T) {
		enc := MockEncoder{AppendStringFunc: func(v string) error {
			return errors.New("encoder error")
		}}

		err := StringValue("9.9").Encode(enc)
		assert.ThatError(t, err).Matches("encoder error")
	})

	t.Run("success", func(t *testing.T) {
		var buf bytes.Buffer
		enc := MockEncoder{AppendStringFunc: func(v string) error {
			_, err := buf.WriteString(v)
			return err
		}}

		err := StringValue("9.9").Encode(enc)
		assert.Nil(t, err)
		assert.ThatString(t, buf.String()).Equal("9.9")
	})
}

func TestReflectValue(t *testing.T) {
	type A struct {
		V string
	}

	t.Run("error", func(t *testing.T) {
		enc := MockEncoder{AppendReflectFunc: func(v any) error {
			return errors.New("encoder error")
		}}

		v := &ReflectValue{Val: &A{V: "a"}}
		err := v.Encode(enc)
		assert.ThatError(t, err).Matches("encoder error")
	})

	t.Run("success", func(t *testing.T) {
		var buf bytes.Buffer
		enc := MockEncoder{AppendReflectFunc: func(v any) error {
			_, err := buf.WriteString(fmt.Sprint(v))
			return err
		}}

		v := &ReflectValue{Val: &A{V: "a"}}
		err := v.Encode(enc)
		assert.Nil(t, err)
		assert.ThatString(t, buf.String()).Equal("&{a}")
	})
}

func TestBoolsValue(t *testing.T) {

	t.Run("error - array begin", func(t *testing.T) {
		enc := MockEncoder{
			AppendArrayBeginFunc: func() error { return errors.New("encoder error") },
		}

		err := BoolsValue([]bool{true, false, true}).Encode(enc)
		assert.ThatError(t, err).Matches("encoder error")
	})

	t.Run("error - append bool", func(t *testing.T) {
		enc := MockEncoder{
			AppendArrayBeginFunc: func() error { return nil },
			AppendBoolFunc:       func(v bool) error { return errors.New("encoder error") },
		}

		err := BoolsValue([]bool{true, false, true}).Encode(enc)
		assert.ThatError(t, err).Matches("encoder error")
	})

	t.Run("error - array end", func(t *testing.T) {
		enc := MockEncoder{
			AppendArrayBeginFunc: func() error { return nil },
			AppendBoolFunc:       func(v bool) error { return nil },
			AppendArrayEndFunc:   func() error { return errors.New("encoder error") },
		}

		err := BoolsValue([]bool{true, false, true}).Encode(enc)
		assert.ThatError(t, err).Matches("encoder error")
	})

	t.Run("success", func(t *testing.T) {
		var buf bytes.Buffer
		enc := MockEncoder{
			AppendArrayBeginFunc: func() error {
				_, err := buf.WriteString("[")
				return err
			},
			AppendArrayEndFunc: func() error {
				_, err := buf.WriteString("]")
				return err
			},
			AppendBoolFunc: func(v bool) error {
				if buf.Len() > 1 {
					if _, err := buf.WriteString(","); err != nil {
						return err
					}
				}
				_, err := buf.WriteString(strconv.FormatBool(v))
				return err
			},
		}

		err := BoolsValue([]bool{true, false, true}).Encode(enc)
		assert.Nil(t, err)
		assert.ThatString(t, buf.String()).Equal("[true,false,true]")
	})
}

func TestIntsValue(t *testing.T) {

	t.Run("error - array begin", func(t *testing.T) {
		enc := MockEncoder{
			AppendArrayBeginFunc: func() error { return errors.New("encoder error") },
		}
		err := IntsValue([]int{1, 2, 3}).Encode(enc)
		assert.ThatError(t, err).Matches("encoder error")
	})

	t.Run("error - append int", func(t *testing.T) {
		enc := MockEncoder{
			AppendArrayBeginFunc: func() error { return nil },
			AppendInt64Func:      func(v int64) error { return errors.New("encoder error") },
		}
		err := IntsValue([]int{1, 2, 3}).Encode(enc)
		assert.ThatError(t, err).Matches("encoder error")
	})

	t.Run("error - array end", func(t *testing.T) {
		enc := MockEncoder{
			AppendArrayBeginFunc: func() error { return nil },
			AppendInt64Func:      func(v int64) error { return nil },
			AppendArrayEndFunc:   func() error { return errors.New("encoder error") },
		}
		err := IntsValue([]int{1, 2, 3}).Encode(enc)
		assert.ThatError(t, err).Matches("encoder error")
	})

	t.Run("success", func(t *testing.T) {
		var buf bytes.Buffer
		enc := MockEncoder{
			AppendArrayBeginFunc: func() error {
				_, err := buf.WriteString("[")
				return err
			},
			AppendArrayEndFunc: func() error {
				_, err := buf.WriteString("]")
				return err
			},
			AppendInt64Func: func(v int64) error {
				if buf.Len() > 1 {
					if _, err := buf.WriteString(","); err != nil {
						return err
					}
				}
				_, err := buf.WriteString(strconv.FormatInt(v, 10))
				return err
			},
		}

		err := IntsValue([]int{1, 2, 3}).Encode(enc)
		assert.Nil(t, err)
		assert.ThatString(t, buf.String()).Equal("[1,2,3]")
	})
}

func TestInt8sValue(t *testing.T) {

	t.Run("error - array begin", func(t *testing.T) {
		enc := MockEncoder{
			AppendArrayBeginFunc: func() error {
				return errors.New("array begin error")
			},
		}
		err := Int8sValue([]int8{1, 2, 3}).Encode(enc)
		assert.ThatError(t, err).Matches("array begin error")
	})

	t.Run("error - append int8", func(t *testing.T) {
		enc := MockEncoder{
			AppendArrayBeginFunc: func() error { return nil },
			AppendInt64Func:      func(v int64) error { return errors.New("append error") },
		}
		err := Int8sValue([]int8{1, 2, 3}).Encode(enc)
		assert.ThatError(t, err).Matches("append error")
	})

	t.Run("error - array end", func(t *testing.T) {
		enc := MockEncoder{
			AppendArrayBeginFunc: func() error { return nil },
			AppendInt64Func:      func(v int64) error { return nil },
			AppendArrayEndFunc:   func() error { return errors.New("array end error") },
		}
		err := Int8sValue([]int8{1, 2, 3}).Encode(enc)
		assert.ThatError(t, err).Matches("array end error")
	})

	t.Run("success", func(t *testing.T) {
		var buf bytes.Buffer
		enc := MockEncoder{
			AppendArrayBeginFunc: func() error {
				_, err := buf.WriteString("[")
				return err
			},
			AppendArrayEndFunc: func() error {
				_, err := buf.WriteString("]")
				return err
			},
			AppendInt64Func: func(v int64) error {
				if buf.Len() > 1 {
					if _, err := buf.WriteString(","); err != nil {
						return err
					}
				}
				_, err := buf.WriteString(strconv.FormatInt(v, 10))
				return err
			},
		}

		err := Int8sValue([]int8{1, 2, 3}).Encode(enc)
		assert.Nil(t, err)
		assert.ThatString(t, buf.String()).Equal("[1,2,3]")
	})
}

func TestInt16sValue(t *testing.T) {

	t.Run("error - array begin", func(t *testing.T) {
		enc := MockEncoder{
			AppendArrayBeginFunc: func() error { return errors.New("array begin error") },
		}
		err := Int16sValue([]int16{1, 2, 3}).Encode(enc)
		assert.ThatError(t, err).Matches("array begin error")
	})

	t.Run("error - append int16", func(t *testing.T) {
		enc := MockEncoder{
			AppendArrayBeginFunc: func() error { return nil },
			AppendInt64Func:      func(v int64) error { return errors.New("append error") },
		}
		err := Int16sValue([]int16{1, 2, 3}).Encode(enc)
		assert.ThatError(t, err).Matches("append error")
	})

	t.Run("error - array end", func(t *testing.T) {
		enc := MockEncoder{
			AppendArrayBeginFunc: func() error { return nil },
			AppendInt64Func:      func(v int64) error { return nil },
			AppendArrayEndFunc:   func() error { return errors.New("array end error") },
		}
		err := Int16sValue([]int16{1, 2, 3}).Encode(enc)
		assert.ThatError(t, err).Matches("array end error")
	})

	t.Run("success", func(t *testing.T) {
		var buf bytes.Buffer
		enc := MockEncoder{
			AppendArrayBeginFunc: func() error {
				_, err := buf.WriteString("[")
				return err
			},
			AppendArrayEndFunc: func() error {
				_, err := buf.WriteString("]")
				return err
			},
			AppendInt64Func: func(v int64) error {
				if buf.Len() > 1 {
					if _, err := buf.WriteString(","); err != nil {
						return err
					}
				}
				_, err := buf.WriteString(strconv.FormatInt(v, 10))
				return err
			},
		}

		err := Int16sValue([]int16{1, 2, 3}).Encode(enc)
		assert.Nil(t, err)
		assert.ThatString(t, buf.String()).Equal("[1,2,3]")
	})
}

func TestInt32sValue(t *testing.T) {

	t.Run("error - array begin", func(t *testing.T) {
		enc := MockEncoder{
			AppendArrayBeginFunc: func() error { return errors.New("array begin error") },
		}
		err := Int32sValue([]int32{1, 2, 3}).Encode(enc)
		assert.ThatError(t, err).Matches("array begin error")
	})

	t.Run("error - append int32", func(t *testing.T) {
		enc := MockEncoder{
			AppendArrayBeginFunc: func() error { return nil },
			AppendInt64Func:      func(v int64) error { return errors.New("append error") },
		}
		err := Int32sValue([]int32{1, 2, 3}).Encode(enc)
		assert.ThatError(t, err).Matches("append error")
	})

	t.Run("error - array end", func(t *testing.T) {
		enc := MockEncoder{
			AppendArrayBeginFunc: func() error { return nil },
			AppendInt64Func:      func(v int64) error { return nil },
			AppendArrayEndFunc:   func() error { return errors.New("array end error") },
		}
		err := Int32sValue([]int32{1, 2, 3}).Encode(enc)
		assert.ThatError(t, err).Matches("array end error")
	})

	t.Run("success", func(t *testing.T) {
		var buf bytes.Buffer
		enc := MockEncoder{
			AppendArrayBeginFunc: func() error {
				_, err := buf.WriteString("[")
				return err
			},
			AppendArrayEndFunc: func() error {
				_, err := buf.WriteString("]")
				return err
			},
			AppendInt64Func: func(v int64) error {
				if buf.Len() > 1 {
					if _, err := buf.WriteString(","); err != nil {
						return err
					}
				}
				_, err := buf.WriteString(strconv.FormatInt(v, 10))
				return err
			},
		}

		err := Int32sValue([]int32{1, 2, 3}).Encode(enc)
		assert.Nil(t, err)
		assert.ThatString(t, buf.String()).Equal("[1,2,3]")
	})
}

func TestInt64sValue(t *testing.T) {

	t.Run("error - array begin", func(t *testing.T) {
		enc := MockEncoder{
			AppendArrayBeginFunc: func() error { return errors.New("array begin error") },
		}
		err := Int64sValue([]int64{1, 2, 3}).Encode(enc)
		assert.ThatError(t, err).Matches("array begin error")
	})

	t.Run("error - append int64", func(t *testing.T) {
		enc := MockEncoder{
			AppendArrayBeginFunc: func() error { return nil },
			AppendInt64Func:      func(v int64) error { return errors.New("append error") },
		}
		err := Int64sValue([]int64{1, 2, 3}).Encode(enc)
		assert.ThatError(t, err).Matches("append error")
	})

	t.Run("error - array end", func(t *testing.T) {
		enc := MockEncoder{
			AppendArrayBeginFunc: func() error { return nil },
			AppendInt64Func:      func(v int64) error { return nil },
			AppendArrayEndFunc:   func() error { return errors.New("array end error") },
		}
		err := Int64sValue([]int64{1, 2, 3}).Encode(enc)
		assert.ThatError(t, err).Matches("array end error")
	})

	t.Run("success", func(t *testing.T) {
		var buf bytes.Buffer
		enc := MockEncoder{
			AppendArrayBeginFunc: func() error {
				_, err := buf.WriteString("[")
				return err
			},
			AppendArrayEndFunc: func() error {
				_, err := buf.WriteString("]")
				return err
			},
			AppendInt64Func: func(v int64) error {
				if buf.Len() > 1 {
					if _, err := buf.WriteString(","); err != nil {
						return err
					}
				}
				_, err := buf.WriteString(strconv.FormatInt(v, 10))
				return err
			},
		}

		err := Int64sValue([]int64{1, 2, 3}).Encode(enc)
		assert.Nil(t, err)
		assert.ThatString(t, buf.String()).Equal("[1,2,3]")
	})
}

func TestUintsValue(t *testing.T) {

	t.Run("error - array begin", func(t *testing.T) {
		enc := MockEncoder{
			AppendArrayBeginFunc: func() error { return errors.New("array begin error") },
		}
		err := UintsValue([]uint{1, 2, 3}).Encode(enc)
		assert.ThatError(t, err).Matches("array begin error")
	})

	t.Run("error - append uint", func(t *testing.T) {
		enc := MockEncoder{
			AppendArrayBeginFunc: func() error { return nil },
			AppendUint64Func:     func(v uint64) error { return errors.New("append error") },
		}
		err := UintsValue([]uint{1, 2, 3}).Encode(enc)
		assert.ThatError(t, err).Matches("append error")
	})

	t.Run("error - array end", func(t *testing.T) {
		enc := MockEncoder{
			AppendArrayBeginFunc: func() error { return nil },
			AppendUint64Func:     func(v uint64) error { return nil },
			AppendArrayEndFunc:   func() error { return errors.New("array end error") },
		}
		err := UintsValue([]uint{1, 2, 3}).Encode(enc)
		assert.ThatError(t, err).Matches("array end error")
	})

	t.Run("success", func(t *testing.T) {
		var buf bytes.Buffer
		enc := MockEncoder{
			AppendArrayBeginFunc: func() error {
				_, err := buf.WriteString("[")
				return err
			},
			AppendArrayEndFunc: func() error {
				_, err := buf.WriteString("]")
				return err
			},
			AppendUint64Func: func(v uint64) error {
				if buf.Len() > 1 {
					if _, err := buf.WriteString(","); err != nil {
						return err
					}
				}
				_, err := buf.WriteString(strconv.FormatUint(v, 10))
				return err
			},
		}

		err := UintsValue([]uint{1, 2, 3}).Encode(enc)
		assert.Nil(t, err)
		assert.ThatString(t, buf.String()).Equal("[1,2,3]")
	})
}

func TestUint8sValue(t *testing.T) {

	t.Run("error - array begin", func(t *testing.T) {
		enc := MockEncoder{
			AppendArrayBeginFunc: func() error { return errors.New("array begin error") },
		}
		err := Uint8sValue([]uint8{1, 2, 3}).Encode(enc)
		assert.ThatError(t, err).Matches("array begin error")
	})

	t.Run("error - append uint8", func(t *testing.T) {
		enc := MockEncoder{
			AppendArrayBeginFunc: func() error { return nil },
			AppendUint64Func:     func(v uint64) error { return errors.New("append error") },
		}
		err := Uint8sValue([]uint8{1, 2, 3}).Encode(enc)
		assert.ThatError(t, err).Matches("append error")
	})

	t.Run("error - array end", func(t *testing.T) {
		enc := MockEncoder{
			AppendArrayBeginFunc: func() error { return nil },
			AppendUint64Func:     func(v uint64) error { return nil },
			AppendArrayEndFunc:   func() error { return errors.New("array end error") },
		}
		err := Uint8sValue([]uint8{1, 2, 3}).Encode(enc)
		assert.ThatError(t, err).Matches("array end error")
	})

	t.Run("success", func(t *testing.T) {
		var buf bytes.Buffer
		enc := MockEncoder{
			AppendArrayBeginFunc: func() error {
				_, err := buf.WriteString("[")
				return err
			},
			AppendArrayEndFunc: func() error {
				_, err := buf.WriteString("]")
				return err
			},
			AppendUint64Func: func(v uint64) error {
				if buf.Len() > 1 {
					if _, err := buf.WriteString(","); err != nil {
						return err
					}
				}
				_, err := buf.WriteString(strconv.FormatUint(v, 10))
				return err
			},
		}

		err := Uint8sValue([]uint8{1, 2, 3}).Encode(enc)
		assert.Nil(t, err)
		assert.ThatString(t, buf.String()).Equal("[1,2,3]")
	})
}

func TestUint16sValue(t *testing.T) {

	t.Run("error - array begin", func(t *testing.T) {
		enc := MockEncoder{
			AppendArrayBeginFunc: func() error { return errors.New("array begin error") },
		}
		err := Uint16sValue([]uint16{1, 2, 3}).Encode(enc)
		assert.ThatError(t, err).Matches("array begin error")
	})

	t.Run("error - append uint16", func(t *testing.T) {
		enc := MockEncoder{
			AppendArrayBeginFunc: func() error { return nil },
			AppendUint64Func:     func(v uint64) error { return errors.New("append error") },
		}
		err := Uint16sValue([]uint16{1, 2, 3}).Encode(enc)
		assert.ThatError(t, err).Matches("append error")
	})

	t.Run("error - array end", func(t *testing.T) {
		enc := MockEncoder{
			AppendArrayBeginFunc: func() error { return nil },
			AppendUint64Func:     func(v uint64) error { return nil },
			AppendArrayEndFunc:   func() error { return errors.New("array end error") },
		}
		err := Uint16sValue([]uint16{1, 2, 3}).Encode(enc)
		assert.ThatError(t, err).Matches("array end error")
	})

	t.Run("success", func(t *testing.T) {
		var buf bytes.Buffer
		enc := MockEncoder{
			AppendArrayBeginFunc: func() error {
				_, err := buf.WriteString("[")
				return err
			},
			AppendArrayEndFunc: func() error {
				_, err := buf.WriteString("]")
				return err
			},
			AppendUint64Func: func(v uint64) error {
				if buf.Len() > 1 {
					if _, err := buf.WriteString(","); err != nil {
						return err
					}
				}
				_, err := buf.WriteString(strconv.FormatUint(v, 10))
				return err
			},
		}

		err := Uint16sValue([]uint16{1, 2, 3}).Encode(enc)
		assert.Nil(t, err)
		assert.ThatString(t, buf.String()).Equal("[1,2,3]")
	})
}

func TestUint32sValue(t *testing.T) {

	t.Run("error - array begin", func(t *testing.T) {
		enc := MockEncoder{
			AppendArrayBeginFunc: func() error { return errors.New("array begin error") },
		}
		err := Uint32sValue([]uint32{1, 2, 3}).Encode(enc)
		assert.ThatError(t, err).Matches("array begin error")
	})

	t.Run("error - append uint32", func(t *testing.T) {
		enc := MockEncoder{
			AppendArrayBeginFunc: func() error { return nil },
			AppendUint64Func:     func(v uint64) error { return errors.New("append error") },
		}
		err := Uint32sValue([]uint32{1, 2, 3}).Encode(enc)
		assert.ThatError(t, err).Matches("append error")
	})

	t.Run("error - array end", func(t *testing.T) {
		enc := MockEncoder{
			AppendArrayBeginFunc: func() error { return nil },
			AppendUint64Func:     func(v uint64) error { return nil },
			AppendArrayEndFunc:   func() error { return errors.New("array end error") },
		}
		err := Uint32sValue([]uint32{1, 2, 3}).Encode(enc)
		assert.ThatError(t, err).Matches("array end error")
	})

	t.Run("success", func(t *testing.T) {
		var buf bytes.Buffer
		enc := MockEncoder{
			AppendArrayBeginFunc: func() error {
				_, err := buf.WriteString("[")
				return err
			},
			AppendArrayEndFunc: func() error {
				_, err := buf.WriteString("]")
				return err
			},
			AppendUint64Func: func(v uint64) error {
				if buf.Len() > 1 {
					if _, err := buf.WriteString(","); err != nil {
						return err
					}
				}
				_, err := buf.WriteString(strconv.FormatUint(v, 10))
				return err
			},
		}

		err := Uint32sValue([]uint32{1, 2, 3}).Encode(enc)
		assert.Nil(t, err)
		assert.ThatString(t, buf.String()).Equal("[1,2,3]")
	})
}

func TestUint64sValue(t *testing.T) {

	t.Run("error - array begin", func(t *testing.T) {
		enc := MockEncoder{
			AppendArrayBeginFunc: func() error { return errors.New("array begin error") },
		}
		err := Uint64sValue([]uint64{1, 2, 3}).Encode(enc)
		assert.ThatError(t, err).Matches("array begin error")
	})

	t.Run("error - append uint64", func(t *testing.T) {
		enc := MockEncoder{
			AppendArrayBeginFunc: func() error { return nil },
			AppendUint64Func:     func(v uint64) error { return errors.New("append error") },
		}
		err := Uint64sValue([]uint64{1, 2, 3}).Encode(enc)
		assert.ThatError(t, err).Matches("append error")
	})

	t.Run("error - array end", func(t *testing.T) {
		enc := MockEncoder{
			AppendArrayBeginFunc: func() error { return nil },
			AppendUint64Func:     func(v uint64) error { return nil },
			AppendArrayEndFunc:   func() error { return errors.New("array end error") },
		}
		err := Uint64sValue([]uint64{1, 2, 3}).Encode(enc)
		assert.ThatError(t, err).Matches("array end error")
	})

	t.Run("success", func(t *testing.T) {
		var buf bytes.Buffer
		enc := MockEncoder{
			AppendArrayBeginFunc: func() error {
				_, err := buf.WriteString("[")
				return err
			},
			AppendArrayEndFunc: func() error {
				_, err := buf.WriteString("]")
				return err
			},
			AppendUint64Func: func(v uint64) error {
				if buf.Len() > 1 {
					if _, err := buf.WriteString(","); err != nil {
						return err
					}
				}
				_, err := buf.WriteString(strconv.FormatUint(v, 10))
				return err
			},
		}

		err := Uint64sValue([]uint64{1, 2, 3}).Encode(enc)
		assert.Nil(t, err)
		assert.ThatString(t, buf.String()).Equal("[1,2,3]")
	})
}

func TestFloat32sValue(t *testing.T) {

	t.Run("error - array begin", func(t *testing.T) {
		enc := MockEncoder{
			AppendArrayBeginFunc: func() error { return errors.New("array begin error") },
		}
		err := Float32sValue([]float32{1.1, 2.2, 3.3}).Encode(enc)
		assert.ThatError(t, err).Matches("array begin error")
	})

	t.Run("error - append float32", func(t *testing.T) {
		enc := MockEncoder{
			AppendArrayBeginFunc: func() error { return nil },
			AppendFloat64Func:    func(v float64) error { return errors.New("append error") },
		}
		err := Float32sValue([]float32{1.1, 2.2, 3.3}).Encode(enc)
		assert.ThatError(t, err).Matches("append error")
	})

	t.Run("error - array end", func(t *testing.T) {
		enc := MockEncoder{
			AppendArrayBeginFunc: func() error { return nil },
			AppendFloat64Func:    func(v float64) error { return nil },
			AppendArrayEndFunc:   func() error { return errors.New("array end error") },
		}
		err := Float32sValue([]float32{1.1, 2.2, 3.3}).Encode(enc)
		assert.ThatError(t, err).Matches("array end error")
	})

	t.Run("success", func(t *testing.T) {
		var buf bytes.Buffer
		enc := MockEncoder{
			AppendArrayBeginFunc: func() error {
				_, err := buf.WriteString("[")
				return err
			},
			AppendArrayEndFunc: func() error {
				_, err := buf.WriteString("]")
				return err
			},
			AppendFloat64Func: func(v float64) error {
				if buf.Len() > 1 {
					if _, err := buf.WriteString(","); err != nil {
						return err
					}
				}
				_, err := buf.WriteString(strconv.FormatFloat(v, 'f', 1, 64))
				return err
			},
		}

		err := Float32sValue([]float32{1.1, 2.2, 3.3}).Encode(enc)
		assert.Nil(t, err)
		assert.ThatString(t, buf.String()).Equal("[1.1,2.2,3.3]")
	})
}

func TestFloat64sValue(t *testing.T) {

	t.Run("error - array begin", func(t *testing.T) {
		enc := MockEncoder{
			AppendArrayBeginFunc: func() error { return errors.New("array begin error") },
		}
		err := Float64sValue([]float64{1.1, 2.2, 3.3}).Encode(enc)
		assert.ThatError(t, err).Matches("array begin error")
	})

	t.Run("error - append float64", func(t *testing.T) {
		enc := MockEncoder{
			AppendArrayBeginFunc: func() error { return nil },
			AppendFloat64Func:    func(v float64) error { return errors.New("append error") },
		}
		err := Float64sValue([]float64{1.1, 2.2, 3.3}).Encode(enc)
		assert.ThatError(t, err).Matches("append error")
	})

	t.Run("error - array end", func(t *testing.T) {
		enc := MockEncoder{
			AppendArrayBeginFunc: func() error { return nil },
			AppendFloat64Func:    func(v float64) error { return nil },
			AppendArrayEndFunc:   func() error { return errors.New("array end error") },
		}
		err := Float64sValue([]float64{1.1, 2.2, 3.3}).Encode(enc)
		assert.ThatError(t, err).Matches("array end error")
	})

	t.Run("success", func(t *testing.T) {
		var buf bytes.Buffer
		enc := MockEncoder{
			AppendArrayBeginFunc: func() error {
				_, err := buf.WriteString("[")
				return err
			},
			AppendArrayEndFunc: func() error {
				_, err := buf.WriteString("]")
				return err
			},
			AppendFloat64Func: func(v float64) error {
				if buf.Len() > 1 {
					if _, err := buf.WriteString(","); err != nil {
						return err
					}
				}
				_, err := buf.WriteString(strconv.FormatFloat(v, 'f', -1, 64))
				return err
			},
		}

		err := Float64sValue([]float64{1.1, 2.2, 3.3}).Encode(enc)
		assert.Nil(t, err)
		assert.ThatString(t, buf.String()).Equal("[1.1,2.2,3.3]")
	})
}

func TestStringsValue(t *testing.T) {

	t.Run("error - array begin", func(t *testing.T) {
		enc := MockEncoder{
			AppendArrayBeginFunc: func() error { return errors.New("array begin error") },
		}
		err := StringsValue([]string{"a", "b", "c"}).Encode(enc)
		assert.ThatError(t, err).Matches("array begin error")
	})

	t.Run("error - append string", func(t *testing.T) {
		enc := MockEncoder{
			AppendArrayBeginFunc: func() error { return nil },
			AppendStringFunc:     func(v string) error { return errors.New("append error") },
		}
		err := StringsValue([]string{"a", "b", "c"}).Encode(enc)
		assert.ThatError(t, err).Matches("append error")
	})

	t.Run("error - array end", func(t *testing.T) {
		enc := MockEncoder{
			AppendArrayBeginFunc: func() error { return nil },
			AppendStringFunc:     func(v string) error { return nil },
			AppendArrayEndFunc:   func() error { return errors.New("array end error") },
		}
		err := StringsValue([]string{"a", "b", "c"}).Encode(enc)
		assert.ThatError(t, err).Matches("array end error")
	})

	t.Run("success", func(t *testing.T) {
		var buf bytes.Buffer
		enc := MockEncoder{
			AppendArrayBeginFunc: func() error {
				_, err := buf.WriteString("[")
				return err
			},
			AppendArrayEndFunc: func() error {
				_, err := buf.WriteString("]")
				return err
			},
			AppendStringFunc: func(v string) error {
				if buf.Len() > 1 {
					if _, err := buf.WriteString(","); err != nil {
						return err
					}
				}
				_, err := buf.WriteString(v)
				return err
			},
		}

		err := StringsValue([]string{"a", "b", "c"}).Encode(enc)
		assert.Nil(t, err)
		assert.ThatString(t, buf.String()).Equal("[a,b,c]")
	})
}

func TestObjectValue(t *testing.T) {

	t.Run("error - object begin", func(t *testing.T) {
		enc := MockEncoder{
			AppendObjectBeginFunc: func() error { return errors.New("object begin error") },
		}

		fields := []Field{{Key: "key", Val: StringValue("value")}}
		err := ObjectValue(fields).Encode(enc)
		assert.ThatError(t, err).Matches("object begin error")
	})

	t.Run("error - append key", func(t *testing.T) {
		enc := MockEncoder{
			AppendObjectBeginFunc: func() error { return nil },
			AppendKeyFunc:         func(key string) error { return errors.New("key error") },
		}

		fields := []Field{{Key: "key", Val: StringValue("value")}}
		err := ObjectValue(fields).Encode(enc)
		assert.ThatError(t, err).Matches("key error")
	})

	t.Run("error - value encode", func(t *testing.T) {
		enc := MockEncoder{
			AppendObjectBeginFunc: func() error { return nil },
			AppendKeyFunc:         func(key string) error { return nil },
			AppendStringFunc:      func(v string) error { return errors.New("value encode error") },
		}

		fields := []Field{{Key: "key", Val: StringValue("value")}}
		err := ObjectValue(fields).Encode(enc)
		assert.ThatError(t, err).Matches("value encode error")
	})

	t.Run("error - object end", func(t *testing.T) {
		enc := MockEncoder{
			AppendObjectBeginFunc: func() error { return nil },
			AppendKeyFunc:         func(key string) error { return nil },
			AppendStringFunc:      func(v string) error { return nil },
			AppendObjectEndFunc:   func() error { return errors.New("object end error") },
		}

		fields := []Field{{Key: "key", Val: StringValue("value")}}
		err := ObjectValue(fields).Encode(enc)
		assert.ThatError(t, err).Matches("object end error")
	})

	t.Run("success", func(t *testing.T) {
		var buf bytes.Buffer
		enc := MockEncoder{
			AppendObjectBeginFunc: func() error {
				_, err := buf.WriteString("{")
				return err
			},
			AppendObjectEndFunc: func() error {
				_, err := buf.WriteString("}")
				return err
			},
			AppendKeyFunc: func(key string) error {
				if buf.Len() > 1 {
					if _, err := buf.WriteString(","); err != nil {
						return err
					}
				}
				_, err := buf.WriteString(strconv.Quote(key))
				return err
			},
			AppendStringFunc: func(v string) error {
				_, err := buf.WriteString(":" + strconv.Quote(v))
				return err
			},
		}

		fields := []Field{{Key: "key", Val: StringValue("value")}}
		err := ObjectValue(fields).Encode(enc)
		assert.Nil(t, err)
		assert.ThatString(t, buf.String()).Equal(`{"key":"value"}`)
	})
}

func TestArrayValue(t *testing.T) {

	t.Run("error - array begin", func(t *testing.T) {
		enc := MockEncoder{
			AppendArrayBeginFunc: func() error { return errors.New("array begin error") },
		}

		values := []Value{StringValue("a"), StringValue("b"), StringValue("c")}
		err := ArrayValue(values).Encode(enc)
		assert.ThatError(t, err).Matches("array begin error")
	})

	t.Run("error - value encode", func(t *testing.T) {
		enc := MockEncoder{
			AppendArrayBeginFunc: func() error { return nil },
			AppendStringFunc:     func(v string) error { return errors.New("value encode error") },
		}

		values := []Value{StringValue("a"), StringValue("b"), StringValue("c")}
		err := ArrayValue(values).Encode(enc)
		assert.ThatError(t, err).Matches("value encode error")
	})

	t.Run("error - array end", func(t *testing.T) {
		enc := MockEncoder{
			AppendArrayBeginFunc: func() error { return nil },
			AppendStringFunc:     func(v string) error { return nil },
			AppendArrayEndFunc:   func() error { return errors.New("array end error") },
		}

		values := []Value{StringValue("a"), StringValue("b"), StringValue("c")}
		err := ArrayValue(values).Encode(enc)
		assert.ThatError(t, err).Matches("array end error")
	})

	t.Run("success", func(t *testing.T) {
		var buf bytes.Buffer
		enc := MockEncoder{
			AppendArrayBeginFunc: func() error {
				_, err := buf.WriteString("[")
				return err
			},
			AppendArrayEndFunc: func() error {
				_, err := buf.WriteString("]")
				return err
			},
			AppendStringFunc: func(v string) error {
				if buf.Len() > 1 {
					if _, err := buf.WriteString(","); err != nil {
						return err
					}
				}
				_, err := buf.WriteString(v)
				return err
			},
		}

		values := []Value{StringValue("a"), StringValue("b"), StringValue("c")}
		err := ArrayValue(values).Encode(enc)
		assert.Nil(t, err)
		assert.ThatString(t, buf.String()).Equal(`[a,b,c]`)
	})
}
