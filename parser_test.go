package structgraphql_test

import (
	"testing"
	"time"

	"github.com/graphql-go/graphql"
	structgraphql "github.com/onichandame/struct-graphql"
	"github.com/stretchr/testify/assert"
)

type Str string

func (Str) Name() string         { return `Str` }
func (Str) Default() interface{} { return `Str` }

type Int int8

func (Int) Description() string { return `Int8bit` }

type ID uint

func (ID) ID() bool { return true }
func TestParser(t *testing.T) {
	t.Run("output", func(t *testing.T) {
		t.Run("can parse primitives", func(t *testing.T) {
			t.Run("string", func(t *testing.T) {
				parser := structgraphql.NewParser()
				rawType := parser.ParseInput(string(""))
				assert.NotNil(t, rawType)
				strType := parser.ParseOutput(Str(""))
				assert.NotNil(t, strType)
				assert.IsType(t, &graphql.Scalar{}, strType)
				assert.Equal(t, Str("").Name(), strType.Name())
			})
			t.Run("int", func(t *testing.T) {
				parser := structgraphql.NewParser()
				rawType := parser.ParseOutput(uint(0))
				assert.NotNil(t, rawType)
				intType := parser.ParseOutput(Int(0))
				assert.NotNil(t, intType)
				assert.IsType(t, new(graphql.Scalar), intType)
				assert.Equal(t, Int(0).Description(), intType.Description())
			})
			t.Run("bool", func(t *testing.T) {
				parser := structgraphql.NewParser()
				rawType := parser.ParseOutput(true)
				assert.NotNil(t, rawType)
				type Bool bool
				boolType := parser.ParseOutput(Bool(false))
				assert.NotNil(t, boolType)
				assert.IsType(t, new(graphql.Scalar), boolType)
			})
			t.Run("date", func(t *testing.T) {
				parser := structgraphql.NewParser()
				rawType := parser.ParseOutput(new(time.Time))
				assert.NotNil(t, rawType)
				type Date *time.Time
				dateType := parser.ParseOutput(Date(new(time.Time)))
				assert.NotNil(t, dateType)
				assert.IsType(t, new(graphql.Scalar), dateType)
			})
		})
		t.Run("can parse objects", func(t *testing.T) {
			t.Run("throws at cyclic reference", func(t *testing.T) {
				parser := structgraphql.NewParser()
				type Obj struct {
					Obj *Obj
				}
				assert.Panics(t, func() { parser.ParseOutput(new(Obj)) })
			})
			t.Run("plain object", func(t *testing.T) {
				parser := structgraphql.NewParser()
				type Obj struct {
					Str  string `graphql:"str"`
					I    int
					Boo  bool
					Date *time.Time
				}
				objType := parser.ParseOutput(new(Obj))
				assert.NotNil(t, objType)
				assert.IsType(t, new(graphql.Object), objType)
				obj := objType.(*graphql.Object)
				assert.NotNil(t, obj.Fields()["str"])
				assert.NotNil(t, obj.Fields()["I"])
				assert.NotNil(t, obj.Fields()["Boo"])
				assert.NotNil(t, obj.Fields()["Date"])
			})
			t.Run("nested object", func(t *testing.T) {
				parser := structgraphql.NewParser()
				type Child struct {
					ID ID `graphql:"id"`
				}
				type Parent struct {
					Children []*Child `graphql:"children,nullable"`
				}
				objType := parser.ParseOutput(new(Parent))
				assert.NotNil(t, objType)
				assert.IsType(t, new(graphql.Object), objType)
				obj := objType.(*graphql.Object)
				assert.NotNil(t, obj.Fields()["children"])
				assert.IsType(t, new(graphql.List), obj.Fields()["children"].Type)
				children := obj.Fields()["children"].Type.(*graphql.List)
				assert.IsType(t, new(graphql.Object), children.OfType)
				child := children.OfType.(*graphql.Object)
				assert.NotNil(t, child.Fields()["id"])
				assert.IsType(t, new(graphql.NonNull), child.Fields()["id"].Type)
			})
		})
	})
	t.Run("input", func(t *testing.T) {
		t.Run("can parse primitives", func(t *testing.T) {
			t.Run("string", func(t *testing.T) {
				parser := structgraphql.NewParser()
				rawType := parser.ParseInput(string(""))
				assert.NotNil(t, rawType)
				strType := parser.ParseInput(Str(""))
				assert.NotNil(t, strType)
				assert.Equal(t, Str("").Name(), strType.Name())
			})
		})
	})
	//t.Run("can parse nested object", func(t *testing.T) {
	//	parser := NewParser()
	//	type GrandChild struct {
	//		ID uint `graphql:",id"`
	//	}
	//	type Child struct {
	//		ID        uint
	//		Offspring *GrandChild
	//	}
	//	type Parent struct {
	//		ID        uint
	//		Offspring []*Child
	//	}
	//	objType := parser.ParseOutput(new(Parent))
	//	assert.NotNil(t, objType)
	//	assert.IsType(t, &graphql.List{}, objType.Fields()["Offspring"].Type)
	//	assert.IsType(t, &graphql.Object{}, objType.Fields()["Offspring"].Type.(*graphql.List).OfType)
	//	assert.IsType(t, &graphql.Object{}, objType.Fields()["Offspring"].Type.(*graphql.List).OfType.(*graphql.Object).Fields()["Offspring"].Type)
	//	assert.IsType(t, graphql.ID, objType.Fields()["Offspring"].Type.(*graphql.List).OfType.(*graphql.Object).Fields()["Offspring"].Type.(*graphql.Object).Fields()["ID"].Type)
	//})
	//t.Run("can double load object", func(t *testing.T) {
	//	parser := NewParser()
	//	type Object struct{}
	//	assert.NotPanics(t, func() {
	//		parser.ParseOutput(new(Object))
	//		parser.ParseOutput(new(Object))
	//	})
	//})
	//t.Run("can load scalar", func(t *testing.T) {
	//	parser := NewParser()
	//	type String string
	//	StringType := graphql.NewScalar(graphql.ScalarConfig{
	//		Name:        "String",
	//		Description: "String",
	//		Serialize: func(value interface{}) interface{} {
	//			switch value := value.(type) {
	//			case String, string:
	//				return value
	//			default:
	//				panic(fmt.Errorf("type must be String"))
	//			}
	//		},
	//		ParseValue: func(value interface{}) interface{} {
	//			switch value := value.(type) {
	//			case String:
	//				return value
	//			case string:
	//				return String(value)
	//			default:
	//				panic(fmt.Errorf("type must be string"))
	//			}
	//		},
	//		ParseLiteral: func(valueAST ast.Value) interface{} {
	//			switch value := valueAST.(type) {
	//			case *ast.StringValue:
	//				return String(value.Value)
	//			default:
	//				panic(fmt.Errorf("type must be string"))
	//			}
	//		},
	//	})
	//	parser.AddScalar(String(""), StringType)
	//	type Object struct {
	//		Name String
	//	}
	//	objType := parser.ParseOutput(new(Object))
	//	assert.NotNil(t, objType)
	//	assert.Equal(t, StringType, objType.Fields()["Name"].Type)
	//})
	//t.Run("can load enum", func(t *testing.T) {
	//	t.Run("raw enum", func(t *testing.T) {
	//		parser := NewParser()
	//		type String string
	//		StringEnum := graphql.NewEnum(graphql.EnumConfig{
	//			Name: "String",
	//			Values: graphql.EnumValueConfigMap{
	//				"A": &graphql.EnumValueConfig{Value: String("a")},
	//				"B": &graphql.EnumValueConfig{Value: String("b")},
	//			},
	//		})
	//		parser.AddEnum(String(""), StringEnum)
	//		type Object struct {
	//			Name String
	//		}
	//		objType := parser.ParseOutput(new(Object))
	//		assert.NotNil(t, objType)
	//		assert.Equal(t, StringEnum, objType.Fields()["Name"].Type)
	//	})
	//	t.Run("wrapped enum", func(t *testing.T) {
	//		parser := NewParser()
	//		type String string
	//		parser.AddEnumByValues(String(""), map[string]interface{}{"A": String("a")})
	//		type Object struct {
	//			Name String
	//		}
	//		objType := parser.ParseOutput(new(Object))
	//		assert.NotNil(t, objType)
	//		assert.IsType(t, &graphql.Enum{}, objType.Fields()["Name"].Type)
	//		assert.Len(t, objType.Fields()["Name"].Type.(*graphql.Enum).Values(), 1)
	//		assert.Equal(t, "A", objType.Fields()["Name"].Type.(*graphql.Enum).Values()[0].Name)
	//		assert.Equal(t, String("a"), objType.Fields()["Name"].Type.(*graphql.Enum).Values()[0].Value)
	//	})
	//})
	//t.Run("can load inputs", func(t *testing.T) {
	//	type Status string
	//	type Args struct {
	//		ID         uint
	//		Name       string `graphql:"name"`
	//		Active     Active
	//		Date       *time.Time
	//		Nested     *Input
	//		Status     Status
	//		NestedList []*Input
	//	}
	//	parser := NewParser()
	//	parser.AddEnum(Status(""), graphql.NewEnum(graphql.EnumConfig{
	//		Name:   "Status",
	//		Values: graphql.EnumValueConfigMap{"Active": &graphql.EnumValueConfig{Value: "active"}, "Inactive": &graphql.EnumValueConfig{Value: "inactive"}},
	//	}))
	//	argsType := parser.ParseArgs(new(Args))
	//	assert.NotNil(t, argsType)
	//	assert.Equal(t, graphql.Int, argsType["ID"].Type)
	//	assert.Equal(t, graphql.String, argsType["name"].Type)
	//	assert.Equal(t, graphql.Boolean, argsType["Active"].Type)
	//	assert.Equal(t, true, argsType["Active"].DefaultValue)
	//	assert.Equal(t, graphql.DateTime, argsType["Date"].Type)
	//	assert.Equal(t, parser.inputs[reflect.TypeOf(Status(""))], argsType["Status"].Type)
	//	assert.IsType(t, &graphql.InputObject{}, argsType["Nested"].Type)
	//	nested := argsType["Nested"].Type.(*graphql.InputObject)
	//	assert.Equal(t, `input`, nested.Name())
	//	assert.Equal(t, `input`, nested.Description())
	//	assert.NotNil(t, nested.Fields())
	//	nestedField := nested.Fields()
	//	assert.Equal(t, graphql.ID, nestedField["id"].Type)
	//	assert.Equal(t, graphql.Boolean, nestedField["Active"].Type)
	//	assert.Equal(t, true, nestedField["Active"].DefaultValue)
	//	assert.IsType(t, graphql.NewList(&graphql.InputObject{}), argsType["NestedList"].Type)
	//	t.Run("throws when circular dependency", func(t *testing.T) {
	//		type Args struct {
	//			Args *Args
	//		}
	//		parser := NewParser()
	//		assert.Panics(t, func() { parser.ParseArgs(new(Args)) })
	//	})
	//})
}
