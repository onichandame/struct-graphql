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
