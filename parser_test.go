package structgraphql

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/stretchr/testify/assert"
)

type Active bool

func (*Active) Default() interface{} { return true }

type Input struct {
	ID     uint `graphql:"id,id"`
	Active Active
}

func (*Input) Name() string        { return `input` }
func (*Input) Description() string { return `input` }

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
	t.Run("can load enum", func(t *testing.T) {
		t.Run("raw enum", func(t *testing.T) {
			parser := NewParser()
			type String string
			StringEnum := graphql.NewEnum(graphql.EnumConfig{
				Name: "String",
				Values: graphql.EnumValueConfigMap{
					"A": &graphql.EnumValueConfig{Value: String("a")},
					"B": &graphql.EnumValueConfig{Value: String("b")},
				},
			})
			parser.AddEnum(String(""), StringEnum)
			type Object struct {
				Name String
			}
			objType := parser.ParseObject(new(Object))
			assert.NotNil(t, objType)
			assert.Equal(t, StringEnum, objType.Fields()["Name"].Type)
		})
		t.Run("wrapped enum", func(t *testing.T) {
			parser := NewParser()
			type String string
			parser.AddEnumByValues(String(""), map[string]interface{}{"A": String("a")})
			type Object struct {
				Name String
			}
			objType := parser.ParseObject(new(Object))
			assert.NotNil(t, objType)
			assert.IsType(t, &graphql.Enum{}, objType.Fields()["Name"].Type)
			assert.Len(t, objType.Fields()["Name"].Type.(*graphql.Enum).Values(), 1)
			assert.Equal(t, "A", objType.Fields()["Name"].Type.(*graphql.Enum).Values()[0].Name)
			assert.Equal(t, String("a"), objType.Fields()["Name"].Type.(*graphql.Enum).Values()[0].Value)
		})
	})
	t.Run("can load inputs", func(t *testing.T) {
		type Status string
		type Args struct {
			ID         uint
			Name       string `graphql:"name"`
			Active     Active
			Date       *time.Time
			Nested     *Input
			Status     Status
			NestedList []*Input
		}
		parser := NewParser()
		parser.AddEnum(Status(""), graphql.NewEnum(graphql.EnumConfig{
			Name:   "Status",
			Values: graphql.EnumValueConfigMap{"Active": &graphql.EnumValueConfig{Value: "active"}, "Inactive": &graphql.EnumValueConfig{Value: "inactive"}},
		}))
		argsType := parser.ParseArgs(new(Args))
		assert.NotNil(t, argsType)
		assert.Equal(t, graphql.Int, argsType["ID"].Type)
		assert.Equal(t, graphql.String, argsType["name"].Type)
		assert.Equal(t, graphql.Boolean, argsType["Active"].Type)
		assert.Equal(t, true, argsType["Active"].DefaultValue)
		assert.Equal(t, graphql.DateTime, argsType["Date"].Type)
		assert.Equal(t, parser.inputs[reflect.TypeOf(Status(""))], argsType["Status"].Type)
		assert.IsType(t, &graphql.InputObject{}, argsType["Nested"].Type)
		nested := argsType["Nested"].Type.(*graphql.InputObject)
		assert.Equal(t, `input`, nested.Name())
		assert.Equal(t, `input`, nested.Description())
		assert.NotNil(t, nested.Fields())
		nestedField := nested.Fields()
		assert.Equal(t, graphql.ID, nestedField["id"].Type)
		assert.Equal(t, graphql.Boolean, nestedField["Active"].Type)
		assert.Equal(t, true, nestedField["Active"].DefaultValue)
		assert.IsType(t, graphql.NewList(&graphql.InputObject{}), argsType["NestedList"].Type)
		t.Run("throws when circular dependency", func(t *testing.T) {
			type Args struct {
				Args *Args
			}
			parser := NewParser()
			assert.Panics(t, func() { parser.ParseArgs(new(Args)) })
		})
	})
}
