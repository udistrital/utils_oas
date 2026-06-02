package formatdata

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/astaxie/beego/logs"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/v2/bson"
)

var validate = validator.New()

func FillStructV(m, s any) error {
	if err := FillStruct(m, &s); err != nil {
		return err
	}

	return validate.Struct(s)
}

func FillStructP(m, s any) {
	if err := FillStruct(m, &s); err != nil {
		panic(err.Error())
	}
}

func SetField(obj any, name string, value any) error {

	structValue := reflect.ValueOf(obj).Elem()
	fieldVal := structValue.FieldByName(name)

	if !fieldVal.IsValid() {
		return fmt.Errorf("No such field: %s in obj", name)
	}

	if !fieldVal.CanSet() {
		return fmt.Errorf("Cannot set %s field value", name)
	}

	val := reflect.ValueOf(value)

	if fieldVal.Type() != val.Type() {

		if m, ok := value.(map[string]any); ok {

			// if field value is struct
			if fieldVal.Kind() == reflect.Struct {
				return FillStruct(m, fieldVal.Addr().Interface())
			}

			// if field value is a pointer to struct
			if fieldVal.Kind() == reflect.Ptr && fieldVal.Type().Elem().Kind() == reflect.Struct {
				if fieldVal.IsNil() {
					fieldVal.Set(reflect.New(fieldVal.Type().Elem()))
				}
				// fmt.Printf("recursive: %v %v\n", m,fieldVal.Interface())
				return FillStruct(m, fieldVal.Interface())
			}

		}

		return fmt.Errorf("Provided value type didn't match obj field type")
	}

	fieldVal.Set(val)
	return nil

}

func FillDataStruct(m map[string]any, s any) error {
	for k, v := range m {
		if err := SetField(s, k, v); err != nil {
			return err
		}
	}

	return nil
}

func FillStructDeep(m map[string]any, fields string, s any) error {
	f := strings.Split(fields, ".")
	if len(f) == 0 {
		return fmt.Errorf("invalid fields.")
	}

	var aux map[string]any
	var load any
	for i, value := range f {
		if i == 0 {
			if err := FillStruct(m[value], &load); err != nil {
				return err
			}
		} else {
			if err := FillStruct(load, &aux); err != nil {
				return err
			}
			if err := FillStruct(aux[value], &load); err != nil {
				return err
			}
		}
	}

	return FillStruct(load, &s)
}

func JsonPrint(x any) error {
	b, err := json.MarshalIndent(x, "", "  ")
	if err != nil {
		logs.Error("error:", err)
		return err
	}
	logs.Info(string(b))
	return nil
}

// StructValidation ... Validate struct by tags
func StructValidation(data any) (errMess []any) {
	valErr := validate.Struct(data)
	if valErr != nil {

		for _, err := range valErr.(validator.ValidationErrors) {
			errMess = append(errMess, fmt.Sprintf("%s", err))
		}

	}
	return
}

// ToMap usa los tags en los campos del struct para decidir cuales campos se agregan
// al map retornado.
func ToMap(in any, tag string) (map[string]any, error) {
	var err error

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	out := make(map[string]any)

	v := reflect.ValueOf(in)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// we only accept structs
	if v.Kind() != reflect.Struct {
		if v.Kind() == reflect.Map {
			FillStructP(in, &out)
			return out, err
		}
		err = fmt.Errorf("ToMap only accepts bson.M, map[string]intefrace{} or structs; got %T", v)
	}

	typ := v.Type()
	for i := 0; i < v.NumField(); i++ {
		// gets us a StructField
		fi := typ.Field(i)
		if tagv, _ := fi.Tag.Lookup(tag); tagv != "" && tagv != "-" {
			tagSplit := strings.Split(tagv, ",")
			out[tagSplit[0]] = v.Field(i).Interface()
		}
	}
	return out, err
}

// FillStructBson ... Unmarshal Bson types to struct
func FillStructBson(in, out any) {
	j, _ := bson.Marshal(in)
	if err := bson.Unmarshal(j, out); err != nil {
		panic(err.Error())
	}
}
