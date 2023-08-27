package attribs

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func ptrTo[T any](v T) *T {
	return &v
}

func TestInspect(t *testing.T) {

	t.Run("test integer", func(t *testing.T) {
		t.Run("test basic", func(t *testing.T) {
			data := []struct {
				input    interface{}
				nullable bool
				signed   bool
			}{
				{int(1), false, true},
				{int8(1), false, true},
				{int16(1), false, true},
				{int32(1), false, true},
				{int64(1), false, true},
				{ptrTo(int(1)), true, true},
				{ptrTo(int8(1)), true, true},
				{ptrTo(int16(1)), true, true},
				{ptrTo(int32(1)), true, true},
				{ptrTo(int64(1)), true, true},
				{uint(1), false, false},
				{uint8(1), false, false},
				{uint16(1), false, false},
				{uint32(1), false, false},
				{uint64(1), false, false},
				{ptrTo(uint(1)), true, false},
				{ptrTo(uint8(1)), true, false},
				{ptrTo(uint16(1)), true, false},
				{ptrTo(uint32(1)), true, false},
				{ptrTo(uint64(1)), true, false},
			}

			for _, item := range data {
				a, err := inspect(item.input, map[reflect.Type]*attr{})
				assert.NoError(t, err)
				assert.NotNil(t, a)
				assert.Equal(t, attrTypeInteger, a.Type)
				assert.Equal(t, item.nullable, a.Nullable)
			}
		})
	})

	type TestStruct struct {
		ID          string  `attr:"name=id"`
		Description *string `attr:"name=description"`
		Age         *int    `attr:"name=age"`
	}

	t.Run("test struct", func(t *testing.T) {
		a, err := inspect(TestStruct{}, map[reflect.Type]*attr{})
		assert.NoError(t, err)
		assert.NotNil(t, a)
	})

	t.Run("test map support", func(t *testing.T) {
		type Test struct {
			Hello string `attr:"name=hello"`
		}

		t.Run("test valid definitions", func(t *testing.T) {
			data := []struct {
				input                any
				expectedElemAttrType attrType
				nullable             bool
			}{
				{map[string]string{}, attrTypeString, false},
				{map[string]map[string]string{}, attrTypeMap, false},
				{&map[string]map[string]string{}, attrTypeMap, true},
				{&map[string]*map[string]string{}, attrTypeMap, true},
				{map[string]any{}, attrTypeAny, false},
				{&map[string]string{}, attrTypeString, true},
				{&map[string]any{}, attrTypeAny, true},
				{&map[string]string{}, attrTypeString, true},
				{&map[string]Test{}, attrTypeStruct, true},
				{map[string]*Test{}, attrTypeStruct, false},
				{&map[string]*int{}, attrTypeInteger, true},
				{map[string]*int{}, attrTypeInteger, false},
				{map[string]*[]string{}, attrTypeArray, false},
				{map[string]*[]*string{}, attrTypeArray, false},
				{map[string][]*string{}, attrTypeArray, false},
			}
			for _, item := range data {
				d, err := inspect(item.input, nil)
				assert.NoError(t, err)
				assert.NotNil(t, d)
				assert.Equal(t, item.expectedElemAttrType, d.Elem.Type, "input: %T, type: %s", item.input, d.Type.String())
				assert.Equal(t, item.nullable, d.Nullable, "input: %T, type: %s", item.input, d.Type.String())
			}
		})

		t.Run("test invalid definitions", func(t *testing.T) {
			data := []struct {
				input interface{}
				err   error
			}{
				{map[int]int{}, ErrMapKeyNotStr},
				{map[bool]int{}, ErrMapKeyNotStr},
				{map[*string]int{}, ErrMapKeyNotStr},
				{map[string]chan struct{}{}, ErrUnsupportedType},
			}
			for _, item := range data {
				d, err := inspect(item.input, nil)
				assert.ErrorIs(t, err, item.err)
				assert.Nil(t, d)
			}
		})

	})

}
