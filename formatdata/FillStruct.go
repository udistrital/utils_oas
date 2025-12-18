package formatdata

import (
	"encoding/json"
)

// FillStruct attempts to fill an `interface{}` with another `interface{}`
//
// Example Usage
//
//	type User struct {
//	  Id   int
//	  Name string
//	}
//	userStruct := User{2, "Pepe"}
//
//	var userMap map[string]interface{}
//	if err := formatdata.FillStruct(userStruct, &userMap); err != nil {
//	  panic(err)
//	}
//	logs.Info("struct --> map OK", userMap)
//
//	var userStruct2 User
//	if err := formatdata.FillStruct(userMap, &userStruct2); err != nil {
//	  panic(err)
//	}
//	logs.Info("map --> struct OK", userStruct2)
func FillStruct(in interface{}, out interface{}) (err error) {
	var str []byte
	if str, err = json.Marshal(in); err != nil {
		return
	}
	err = json.Unmarshal(str, &out)
	return
}
