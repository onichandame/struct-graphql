package structgraphql

import (
	"reflect"
	"testing"
	"time"

	"github.com/graphql-go/graphql"
	"github.com/stretchr/testify/assert"
)

func TestParser(t *testing.T) {
	t.Run("throws when parsing non-struct", func(t *testing.T) {
		parser := NewParser()
		assert.Panics(t, func() { parser.ParseObject("") })
		assert.Panics(t, func() { parser.ParseObject(1) })
		assert.Panics(t, func() { parser.ParseObject(true) })
		assert.Panics(t, func() { parser.ParseObject([]interface{}{}) })
		assert.Panics(t, func() { parser.ParseObject(map[interface{}]interface{}{}) })
	})
	t.Run("can parse simple plain object", func(t *testing.T) {
		parser := NewParser()
		type Object struct {
			Name string
			Int  int        `graphql:"int"`
			Bool bool       `graphql:",nullable"`
			Date *time.Time `graphql:",id"`
		}
		objType := parser.ParseObject(new(Object))
		assert.NotNil(t, objType)
		fields := objType.Fields()
		assert.Len(t, fields, 4)
		assert.Contains(t, fields, "Name")
		assert.Contains(t, fields, "int")
		assert.Contains(t, fields, "Bool")
		assert.Contains(t, fields, "Date")
		assert.Equal(t, graphql.String, fields["Name"].Type)
		assert.Equal(t, graphql.Int, fields["int"].Type)
		assert.IsType(t, reflect.TypeOf(&graphql.NonNull{}), fields["Bool"].Type)
	})
}
