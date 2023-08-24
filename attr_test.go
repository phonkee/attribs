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

}
