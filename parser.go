package structgraphql

import (
	"fmt"
	"reflect"
	"time"

	"github.com/graphql-go/graphql"
)

type Parser struct {
	types  map[reflect.Type]graphql.Type
	inputs map[reflect.Type]*graphql.ArgumentConfig
}

func NewParser() *Parser {
	var parser Parser
	parser.inputs = make(map[reflect.Type]*graphql.ArgumentConfig)
	parser.types = make(map[reflect.Type]graphql.Type)
	parser.types[reflect.TypeOf(time.Time{})] = graphql.DateTime
	parser.types[reflect.TypeOf(false)] = graphql.Boolean
	ints := []interface{}{int(0), int8(0), int16(0), int32(0), int64(0), uint(0), uint8(0), uint16(0), uint32(0), uint64(0)}
	floats := []interface{}{float32(0), float64(0)}
	strings := []interface{}{string(``), []byte(``)}
	for _, i := range ints {
		parser.types[reflect.TypeOf(i)] = graphql.Int
	}
	for _, f := range floats {
		parser.types[reflect.TypeOf(f)] = graphql.Float
	}
	for _, s := range strings {
		parser.types[reflect.TypeOf(s)] = graphql.String
	}
	return &parser
}

func (parser *Parser) isTypeLoaded(t reflect.Type) bool {
	_, ok := parser.types[t]
	return ok
}

func (parser *Parser) isInputLoaded(t reflect.Type) bool {
	_, ok := parser.inputs[t]
	return ok
}

func (parser *Parser) AddEnum(ent interface{}, enum *graphql.Enum) {
	t := getType(ent)
	if parser.isTypeLoaded(t) {
		return
	}
	parser.types[t] = enum
}

func (parser *Parser) AddEnumByValues(ent interface{}, values map[string]interface{}) {
	t := getType(ent)
	if parser.isTypeLoaded(t) {
		return
	}
	name := getName(t)
	description := getDescription(t)
	valuesMap := make(graphql.EnumValueConfigMap)
	for name, value := range values {
		valuesMap[name] = &graphql.EnumValueConfig{Value: value}
	}
	enum := graphql.NewEnum(graphql.EnumConfig{
		Name:        name,
		Description: description,
		Values:      valuesMap,
	})
	parser.types[t] = enum
}

func (parser *Parser) AddScalar(ent interface{}, value *graphql.Scalar) {
	t := getType(ent)
	if parser.isTypeLoaded(t) {
		return
	}
	parser.types[t] = value
}

func (parser *Parser) ParseObject(ent interface{}) *graphql.Object {
	t := getType(ent)
	if parser.isTypeLoaded(t) {
		return parser.types[t].(*graphql.Object)
	}
	var loadObjectType func(reflect.Type, map[reflect.Type]interface{})
	loadObjectType = func(t reflect.Type, loading map[reflect.Type]interface{}) {
		t = getType(t)
		if parser.isTypeLoaded(t) {
			return
		}
		fields := make(graphql.Fields)
		if t.Kind() != reflect.Struct {
			panic(fmt.Errorf("object type must be a struct"))
		}
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			fieldType := getType(field.Type)
			name := getFieldName(&field)
			description := getDescription(fieldType)
			var sliceDims int
			var unwrapSlice func(t reflect.Type) reflect.Type
			unwrapSlice = func(t reflect.Type) reflect.Type {
				if t.Kind() == reflect.Slice {
					sliceDims++
					return unwrapSlice(t.Elem())
				} else {
					return t
				}
			}
			fieldType = getType(unwrapSlice(fieldType))
			if _, ok := parser.types[fieldType]; !ok {
				if fieldType.Kind() != reflect.Struct {
					panic(fmt.Errorf("type %v not struct but %v", fieldType.Name(), fieldType.Kind()))
				}
				if _, ok := loading[fieldType]; ok {
					panic(fmt.Errorf("loading type %v hits a loop", fieldType.Name()))
				}
				newLoading := make(map[reflect.Type]interface{})
				for t := range loading {
					newLoading[t] = nil
				}
				newLoading[fieldType] = nil
				loadObjectType(fieldType, newLoading)
			}
			t := parser.types[fieldType]
			for dim := 0; dim < sliceDims; dim++ {
				t = graphql.NewList(t)
			}
			t = decorateFieldType(&field, t)
			fields[name] = &graphql.Field{Type: t, Description: description}
		}
		name := getName(t)
		description := getDescription(t)
		parser.types[t] = graphql.NewObject(graphql.ObjectConfig{
			Name:        name,
			Description: description,
			Fields:      fields,
		})
	}
	loadObjectType(t, make(map[reflect.Type]interface{}))
	return parser.types[t].(*graphql.Object)
}

func (parser *Parser) GetType(t reflect.Type) graphql.Type {
	return parser.types[t]
}

func (parser *Parser) ParseInput(ent interface{}) *graphql.ArgumentConfig {
	t := getType(ent)
	if !parser.isInputLoaded(t) {
	}
	return parser.inputs[t]
}
