package structgraphql

import (
	"reflect"

	"github.com/fatih/structtag"
	"github.com/graphql-go/graphql"
	goutils "github.com/onichandame/go-utils"
)

type Named interface{ Name() string }

func getName(t reflect.Type) string {
	t = goutils.UnwrapType(t)
	name := t.Name()
	if named, ok := reflect.New(t).Interface().(Named); ok {
		name = named.Name()
	}
	return name
}

type Described interface{ Description() string }

func getDescription(t reflect.Type) string {
	t = goutils.UnwrapType(t)
	description := ``
	if described, ok := reflect.New(t).Interface().(Described); ok {
		description = described.Description()
	}
	return description
}

func getType(ent interface{}) reflect.Type {
	if t, ok := ent.(reflect.Type); ok {
		return goutils.UnwrapType(t)
	} else {
		return goutils.UnwrapType(reflect.TypeOf(ent))
	}
}

func decorateFieldType(field *reflect.StructField, t graphql.Type) graphql.Type {
	tags, _ := structtag.Parse(string(field.Tag))
	if tags != nil {
		tag, _ := tags.Get(TAG_PREFIX)
		if tag != nil {
			if tag.HasOption(TAG_ID) {
				t = graphql.ID
			}
			if tag.HasOption(TAG_NULLABLE) {
				t = graphql.NewNonNull(t)
			}
		}
	}
	return t
}

type Defaulted interface {
	Default() interface{}
}

func getDefault(t reflect.Type) interface{} {
	var res interface{}
	if def, ok := reflect.New(t).Interface().(Defaulted); ok {
		res = def.Default()
	}
	return res
}

func getFieldName(field *reflect.StructField) string {
	name := field.Name
	tags, _ := structtag.Parse(string(field.Tag))
	if tags != nil {
		tag, _ := tags.Get(TAG_PREFIX)
		if tag != nil {
			if tag.Name != "" {
				name = tag.Name
			}
		}
	}
	return name
}

func unwrapSlice(t reflect.Type, opts ...interface{}) (reflect.Type, int) {
	var dim int
	if len(opts) > 0 {
		if d, ok := opts[0].(int); ok {
			dim = d
		}
	}
	if t.Kind() == reflect.Slice {
		dim++
		return unwrapSlice(t.Elem(), dim)
	} else {
		return t, dim
	}
}
