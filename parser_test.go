package structgraphql

import (
	"fmt"
	"testing"
	"time"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
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
			ID   uint `graphql:",id"`
			Int  int  `graphql:"int"`
			Bool bool `graphql:",nullable"`
			Date *time.Time
		}
		objType := parser.ParseObject(new(Object))
		assert.NotNil(t, objType)
		fields := objType.Fields()
		assert.Contains(t, fields, "Name")
		assert.Contains(t, fields, "int")
		assert.Contains(t, fields, "Bool")
		assert.Contains(t, fields, "Date")
		assert.Contains(t, fields, "ID")
		assert.Equal(t, graphql.String, fields["Name"].Type)
		assert.Equal(t, graphql.Int, fields["int"].Type)
		assert.IsType(t, &graphql.NonNull{}, fields["Bool"].Type)
		assert.Equal(t, graphql.Boolean, fields["Bool"].Type.(*graphql.NonNull).OfType)
		assert.Equal(t, graphql.DateTime, fields["Date"].Type)
		assert.Equal(t, graphql.ID, fields["ID"].Type)
	})
	t.Run("can parse nested object", func(t *testing.T) {
		parser := NewParser()
		type GrandChild struct {
			ID uint `graphql:",id"`
		}
		type Child struct {
			ID        uint
			Offspring *GrandChild
		}
		type Parent struct {
			ID        uint
			Offspring []*Child
		}
		objType := parser.ParseObject(new(Parent))
		assert.NotNil(t, objType)
		assert.IsType(t, &graphql.List{}, objType.Fields()["Offspring"].Type)
		assert.IsType(t, &graphql.Object{}, objType.Fields()["Offspring"].Type.(*graphql.List).OfType)
		assert.IsType(t, &graphql.Object{}, objType.Fields()["Offspring"].Type.(*graphql.List).OfType.(*graphql.Object).Fields()["Offspring"].Type)
		assert.IsType(t, graphql.ID, objType.Fields()["Offspring"].Type.(*graphql.List).OfType.(*graphql.Object).Fields()["Offspring"].Type.(*graphql.Object).Fields()["ID"].Type)
	})
	t.Run("can double load object", func(t *testing.T) {
		parser := NewParser()
		type Object struct{}
		assert.NotPanics(t, func() {
			parser.ParseObject(new(Object))
			parser.ParseObject(new(Object))
		})
	})
	t.Run("can load scalar", func(t *testing.T) {
		parser := NewParser()
		type String string
		StringType := graphql.NewScalar(graphql.ScalarConfig{
			Name:        "String",
			Description: "String",
			Serialize: func(value interface{}) interface{} {
				switch value := value.(type) {
				case String, string:
					return value
				default:
					panic(fmt.Errorf("type must be String"))
				}
			},
			ParseValue: func(value interface{}) interface{} {
				switch value := value.(type) {
				case String:
					return value
				case string:
					return String(value)
				default:
					panic(fmt.Errorf("type must be string"))
				}
			},
			ParseLiteral: func(valueAST ast.Value) interface{} {
				switch value := valueAST.(type) {
				case *ast.StringValue:
					return String(value.Value)
				default:
					panic(fmt.Errorf("type must be string"))
				}
			},
		})
		parser.AddScalar(String(""), StringType)
		type Object struct {
			Name String
		}
		objType := parser.ParseObject(new(Object))
		assert.NotNil(t, objType)
		assert.Equal(t, StringType, objType.Fields()["Name"].Type)
	})
}
