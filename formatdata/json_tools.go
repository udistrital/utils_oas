package formatdata

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/globalsign/mgo/bson"
	"gopkg.in/go-playground/validator.v9"
)

var validate *validator.Validate

func FillStructV(m interface{}, s interface{}) (err error) {
	validate = validator.New()
	j, _ := json.Marshal(m)
	err = json.Unmarshal(j, s)
	valErr := validate.Struct(s)

	if valErr != nil {
		err = valErr
	}
	return
}

func FillStruct(m interface{}, s interface{}) error {
	j, _ := json.Marshal(m)
	err := json.Unmarshal(j, s)
	return err
}

func FillStructP(m interface{}, s interface{}) {
	j, _ := json.Marshal(m)
	err := json.Unmarshal(j, s)
	if err != nil {
		panic(err.Error())
	}
}

func SetField(obj interface{}, name string, value interface{}) error {

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

		if m, ok := value.(map[string]interface{}); ok {

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

func FillDataStruct(m map[string]interface{}, s interface{}) error {
	for k, v := range m {
		err := SetField(s, k, v)
		if err != nil {
			return err
		}
	}
	return nil
}
func FillStructDeep(m map[string]interface{}, fields string, s interface{}) (err error) {
	f := strings.Split(fields, ".")
	if len(f) == 0 {
		err = errors.New("invalid fields.")
		return
	}

	var aux map[string]interface{}
	var load interface{}
	for i, value := range f {

		if i == 0 {
			//fmt.Println(m[value])
			FillStruct(m[value], &load)
		} else {
			FillStruct(load, &aux)
			FillStruct(aux[value], &load)
			//fmt.Println(aux[value])
		}
	}
	j, _ := json.Marshal(load)
	err = json.Unmarshal(j, s)
	return
}

func JsonPrint(x interface{}) (err error) {
	b, err := json.MarshalIndent(x, "", "  ")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Print(string(b))
	return
}

// StructValidation ... Validate struct by tags
func StructValidation(data interface{}) (errMess []interface{}) {
	validate = validator.New()
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
func ToMap(in interface{}, tag string) (map[string]interface{}, error) {
	var (
		err error
	)
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf(fmt.Sprintf("%s", r))
		}
	}()
	out := make(map[string]interface{})

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
func FillStructBson(in interface{}, out interface{}) {
	j, _ := bson.Marshal(in)
	err := bson.Unmarshal(j, out)
	if err != nil {
		panic(err.Error())
	}
	return
}
