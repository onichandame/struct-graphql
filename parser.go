package structgraphql

import (
	"fmt"
	"reflect"
	"time"

	"github.com/graphql-go/graphql"
)

type Parser struct {
	types  map[reflect.Type]graphql.Type
	inputs map[reflect.Type]graphql.Input
}

func NewParser() *Parser {
	var parser Parser
	parser.inputs = make(map[reflect.Type]graphql.Input)
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

func (parser *Parser) AddEnum(ent interface{}, enum *graphql.Enum) {
	t := getType(ent)
	if parser.isTypeLoaded(t) {
		return
	}
	parser.types[t] = enum
	parser.inputs[t] = enum
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
	parser.inputs[t] = enum
}

func (parser *Parser) AddScalar(ent interface{}, value *graphql.Scalar) {
	t := getType(ent)
	if parser.isTypeLoaded(t) {
		return
	}
	parser.types[t] = value
	parser.inputs[t] = value
}

func (parser *Parser) ParseOutput(ent interface{}, opts ...interface{}) graphql.Type {
	t := getType(ent)
	t, sliceDims := unwrapSlice(t)
	t = getType(t)
	if !parser.isTypeLoaded(t) {
		visited := make(map[reflect.Type]interface{})
		if len(opts) > 0 {
			if v, ok := opts[0].(map[reflect.Type]interface{}); ok {
				visited = v
			}
		}
		if _, ok := visited[t]; ok {
			panic(fmt.Errorf("when loading output type there must not be a cyclic reference at %v", t.Name()))
		}
		visited[t] = nil
		if t != reflect.TypeOf(time.Time{}) && t.Kind() == reflect.Struct {
			fields := make(graphql.Fields)
			var loadStruct func(t reflect.Type)
			loadStruct = func(t reflect.Type) {
				for i := 0; i < t.NumField(); i++ {
					field := t.Field(i)
					if field.Anonymous {
						loadStruct(field.Type)
					} else {
						fieldType := getType(field.Type)
						name := getFieldName(&field)
						var sliceDims int
						elemType, sliceDims := unwrapSlice(fieldType)
						fieldType = getType(elemType)
						var fieldtype graphql.Type
						if ft, ok := parser.types[fieldType]; !ok {
							fieldtype = parser.ParseOutput(fieldType, visited)
						} else {
							fieldtype = ft
						}
						for dim := 0; dim < sliceDims; dim++ {
							fieldtype = graphql.NewList(fieldtype)
						}
						fieldtype = decorateFieldType(&field, fieldtype)
						fields[name] = &graphql.Field{Type: fieldtype, Description: getDescription(fieldType), Name: getName(fieldType)}
					}
				}
			}
			loadStruct(t)
			parser.types[t] = graphql.NewObject(graphql.ObjectConfig{
				Fields:      fields,
				Name:        getName(t),
				Description: getDescription(t),
			})
		} else {
			var baseType *graphql.Scalar
			if isID(t) {
				baseType = graphql.ID
			} else if t == reflect.TypeOf(time.Time{}) {
				baseType = graphql.DateTime
			} else {
				switch t.Kind() {
				case reflect.Float32, reflect.Float64:
					baseType = graphql.Float
				case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int8, reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint8:
					baseType = graphql.Int
				case reflect.String:
					baseType = graphql.String
				case reflect.Bool:
					baseType = graphql.Boolean
				default:
					panic(fmt.Errorf("type %v not supported", t.Kind()))
				}
			}
			parser.types[t] = graphql.NewScalar(graphql.ScalarConfig{Serialize: baseType.Serialize, ParseValue: baseType.ParseValue, ParseLiteral: baseType.ParseLiteral, Name: getName(t), Description: getDescription(t)})
		}
	}
	res := parser.types[t]
	for i := 0; i < sliceDims; i++ {
		res = graphql.NewList(res)
	}
	return res
}

func (parser *Parser) ParseInput(ent interface{}, opts ...interface{}) graphql.Input {
	t := getType(ent)
	t, sliceDims := unwrapSlice(t)
	t = getType(t)
	if _, ok := parser.inputs[t]; !ok {
		visited := make(map[reflect.Type]interface{})
		if len(opts) > 0 {
			if v, ok := opts[0].(map[reflect.Type]interface{}); ok {
				visited = v
			}
		}
		if _, ok := visited[t]; ok {
			panic(fmt.Errorf("when loading input type there must not be a cyclic reference at %v", t.Name()))
		}
		visited[t] = nil
		if t.Kind() == reflect.Struct && t != reflect.TypeOf(time.Time{}) {
			fields := make(graphql.InputObjectConfigFieldMap)
			var loadStruct func(t reflect.Type)
			loadStruct = func(t reflect.Type) {
				for i := 0; i < t.NumField(); i++ {
					field := t.Field(i)
					if field.Anonymous {
						loadStruct(field.Type)
					} else {
						fieldType := getType(field.Type)
						name := getFieldName(&field)
						var sliceDims int
						elemType, sliceDims := unwrapSlice(fieldType)
						fieldType = getType(elemType)
						var fieldtype graphql.Type
						if ft, ok := parser.inputs[fieldType]; !ok {
							fieldtype = parser.ParseInput(fieldType, visited)
						} else {
							fieldtype = ft
						}
						for dim := 0; dim < sliceDims; dim++ {
							fieldtype = graphql.NewList(fieldtype)
						}
						fieldtype = decorateFieldType(&field, fieldtype)
						fields[name] = &graphql.InputObjectFieldConfig{Type: fieldtype, Description: getDescription(fieldType), DefaultValue: getDefault(fieldType)}
					}
				}
			}
			loadStruct(t)
			parser.inputs[t] = graphql.NewInputObject(graphql.InputObjectConfig{
				Name:        getName(t),
				Description: getDescription(t),
				Fields:      fields,
			})
		} else {
			var basetype *graphql.Scalar
			if t == reflect.TypeOf(time.Time{}) {
				basetype = graphql.DateTime
			} else {
				switch t.Kind() {
				case reflect.Float32, reflect.Float64:
					basetype = graphql.Float
				case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int8, reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint8:
					basetype = graphql.Int
				case reflect.String:
					basetype = graphql.String
				case reflect.Bool:
					basetype = graphql.Boolean
				default:
					panic(fmt.Errorf("type %v not supported", t.Kind()))
				}
			}
			parser.inputs[t] = graphql.NewScalar(graphql.ScalarConfig{Name: getName(t), Description: getDescription(t), Serialize: basetype.Serialize, ParseValue: basetype.ParseValue, ParseLiteral: basetype.ParseLiteral})
		}
	}
	res := parser.inputs[t]
	for i := 0; i < sliceDims; i++ {
		res = graphql.NewList(res)
	}
	return res
}

// parse all args as a struct
func (parser *Parser) ParseArgs(ent interface{}) graphql.FieldConfigArgument {
	t := getType(ent)
	if t.Kind() != reflect.Struct {
		panic(fmt.Errorf("args must be passed as a struct"))
	}
	args := make(graphql.FieldConfigArgument)
	var loadStruct func(t reflect.Type)
	loadStruct = func(t reflect.Type) {
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			if field.Anonymous {
				loadStruct(field.Type)
			} else {
				fieldType := getType(field.Type)
				fieldType, sliceDims := unwrapSlice(fieldType)
				argType := parser.ParseInput(fieldType)
				for i := 0; i < sliceDims; i++ {
					argType = graphql.NewList(argType)
				}
				argType = decorateFieldType(&field, argType)
				args[getFieldName(&field)] = &graphql.ArgumentConfig{
					Type:         argType,
					Description:  getDescription(fieldType),
					DefaultValue: getDefault(fieldType),
				}
			}
		}
	}
	loadStruct(t)
	return args
}
