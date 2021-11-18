package structgraphql_test

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/graphql-go/graphql"
	goutils "github.com/onichandame/go-utils"
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
			t.Run("int", func(t *testing.T) {
				parser := structgraphql.NewParser()
				rawType := parser.ParseInput(int16(0))
				assert.NotNil(t, rawType)
				intType := parser.ParseInput(Int(0))
				assert.NotNil(t, intType)
				assert.Equal(t, Int(0).Description(), intType.Description())
			})
			t.Run("bool", func(t *testing.T) {
				parser := structgraphql.NewParser()
				rawType := parser.ParseInput(false)
				assert.NotNil(t, rawType)
				type Bool bool
				boolType := parser.ParseInput(Bool(true))
				assert.NotNil(t, boolType)
			})
			t.Run("date", func(t *testing.T) {
				parser := structgraphql.NewParser()
				rawType := parser.ParseInput(new(time.Time))
				assert.NotNil(t, rawType)
				type Date *time.Time
				dateType := parser.ParseInput(new(Date))
				assert.NotNil(t, dateType)
			})
		})
		t.Run("can parse objects", func(t *testing.T) {
			t.Run("throws when circular dependency", func(t *testing.T) {
				parser := structgraphql.NewParser()
				type Input struct {
					Input *Input
				}
				assert.Panics(t, func() { parser.ParseInput(new(Input)) })
			})
			t.Run("plain object", func(t *testing.T) {
				parser := structgraphql.NewParser()
				type Input struct {
					ID   ID         `graphql:"id"`
					Str  Str        `graphql:"str"`
					Bool bool       `graphql:"bool"`
					Date *time.Time `graphql:"date"`
				}
				inputType := parser.ParseInput(new(Input))
				assert.NotNil(t, inputType)
				assert.IsType(t, new(graphql.InputObject), inputType)
				input := inputType.(*graphql.InputObject)
				assert.NotNil(t, input.Fields()["id"])
			})
			t.Run("nested object", func(t *testing.T) {
				parser := structgraphql.NewParser()
				type Child struct {
					ID uint `graphql:"id"`
				}
				type Input struct {
					Children []*Child `graphql:"children,nullable"`
				}
				inputType := parser.ParseInput(new(Input))
				assert.NotNil(t, inputType)
				assert.IsType(t, new(graphql.InputObject), inputType)
				input := inputType.(*graphql.InputObject)
				assert.NotNil(t, input.Fields()["children"])
				assert.IsType(t, new(graphql.List), input.Fields()["children"].Type)
				children := input.Fields()["children"].Type.(*graphql.List).OfType
				assert.IsType(t, new(graphql.InputObject), children)
				assert.IsType(t, new(graphql.NonNull), children.(*graphql.InputObject).Fields()["id"].Type)
			})
		})
	})
	t.Run("args", func(t *testing.T) {
		t.Run("throws when passed non-struct", func(t *testing.T) {
			parser := structgraphql.NewParser()
			assert.Panics(t, func() { parser.ParseArgs("") })
			assert.Panics(t, func() { parser.ParseArgs(0) })
			assert.Panics(t, func() { parser.ParseArgs(true) })
			assert.Panics(t, func() { parser.ParseArgs(map[interface{}]interface{}{}) })
			assert.Panics(t, func() { parser.ParseArgs([]interface{}{}) })
		})
		t.Run("primitives", func(t *testing.T) {
			parser := structgraphql.NewParser()
			type Args struct {
				ID     uint       `graphql:"id,nullable"`
				Name   string     `graphql:"name,nullable"`
				Active bool       `graphql:"active"`
				Date   *time.Time `graphql:"date,nullable"`
			}
			argsType := parser.ParseArgs(new(Args))
			assert.NotNil(t, argsType)
			assert.NotNil(t, argsType["id"])
			assert.NotNil(t, argsType["name"])
			assert.NotNil(t, argsType["active"])
			assert.IsType(t, new(graphql.NonNull), argsType["active"].Type)
			assert.NotNil(t, argsType["date"])
		})
		t.Run("nested objects", func(t *testing.T) {
			parser := structgraphql.NewParser()
			type Input struct {
				ID uint `graphql:"id"`
			}
			type Args struct {
				Input *Input `graphql:"input,nullable"`
			}
			argsType := parser.ParseArgs(new(Args))
			assert.NotNil(t, argsType)
			assert.NotNil(t, argsType["input"])
			assert.IsType(t, new(graphql.InputObject), argsType["input"].Type)
			input := argsType["input"].Type.(*graphql.InputObject)
			assert.NotNil(t, input.Fields()["id"])
			assert.IsType(t, new(graphql.NonNull), input.Fields()["id"].Type)
		})
	})
	t.Run("end-to-end", func(t *testing.T) {
		parser := structgraphql.NewParser()
		type Input struct {
			Name Str `graphql:"name"`
		}
		type Args struct {
			Input   *Input `graphql:"input"`
			Message string `graphql:"message"`
		}
		type Output struct {
			Name     string `graphql:"name" json:"name"`
			Message  string `graphql:"message" json:"message"`
			Greeting string `graphql:"greeting" json:"greeting"`
		}
		schema, err := graphql.NewSchema(graphql.SchemaConfig{
			Query: graphql.NewObject(graphql.ObjectConfig{
				Name: "query",
				Fields: graphql.Fields{
					"handshake": &graphql.Field{
						Args: parser.ParseArgs(new(Args)),
						Type: parser.ParseOutput(new(Output)),
						Resolve: func(p graphql.ResolveParams) (res interface{}, err error) {
							defer goutils.RecoverToErr(&err)
							var out Output
							out.Name = p.Args["input"].(map[string]interface{})["name"].(string)
							out.Message = p.Args["message"].(string)
							out.Greeting = fmt.Sprintf("hello %v", out.Name)
							res = &out
							return res, err
						},
					},
				},
			}),
		})
		assert.Nil(t, err)
		res := graphql.Do(graphql.Params{
			Schema:        schema,
			RequestString: `{handshake(input:{name:"jimmy"},message:"hi"){name message greeting}}`,
		})
		assert.Nil(t, res.Errors)
		by, err := json.Marshal(res.Data)
		assert.Nil(t, err)
		type Response struct {
			Handshake Output `json:"handshake"`
		}
		var out Response
		assert.Nil(t, json.Unmarshal(by, &out))
		assert.Equal(t, "jimmy", out.Handshake.Name)
		assert.Equal(t, "hello jimmy", out.Handshake.Greeting)
	})
}
