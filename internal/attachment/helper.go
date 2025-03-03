package attachment

import (
	"github.com/google/uuid"
	"reflect"
)

func MapFormToStruct(form map[string][]string, dst interface{}) error {
	dstVal := reflect.ValueOf(dst).Elem()
	dstType := dstVal.Type()

	for i := 0; i < dstType.NumField(); i++ {
		field := dstType.Field(i)
		formTag := field.Tag.Get("form")
		if values, ok := form[formTag]; ok && len(values) > 0 {
			fieldVal := dstVal.FieldByName(field.Name)
			if fieldVal.CanSet() {
				fieldVal.SetString(values[0])
			}
		}
	}
	return nil
}

func GetUUID() string {
	return uuid.New().String()
}
