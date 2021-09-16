// REFERENCIA PARA EL TESTEO:
// https://golang.org/doc/tutorial/add-a-test

package apistatus

import (
	"testing"
	// "github.com/astaxie/beego"
	// "github.com/astaxie/beego/context"
)

// TESTS DE EXITO

func TestInitWithHandler(t *testing.T) {
	sucessHandler := func() (err interface{}) {
		return
	}
	InitWithHandler(sucessHandler)
	// TODO: Simular llamado al controlador "/"
	if true {
		t.Fatalf("TEST POR IMPLEMENTAR - InitWithHandler no retornó {Status: Ok}")
	}
}

func TestInit(t *testing.T) {
	Init()
	// TODO: Simular llamado al controlador "/"
	if true {
		t.Fatalf("TEST POR IMPLEMENTAR - Init no retornó {Status: Ok}")
	}
}

// TESTS DE FALLOS

func TestInitWithHandlerPanicFail(t *testing.T) {
	failureHandlerWithPanic := func() (err interface{}) {
		panic("ErrWithPanic")
	}
	InitWithHandler(failureHandlerWithPanic)
	// TODO: Simular llamado al controlador "/"
	if true {
		t.Fatalf("TEST POR IMPLEMENTAR - InitWithHandler no retornó {Status: ERROR: ...} ante panic()")
	}
}

func TestInitWithHandlerNotNilFail(t *testing.T) {
	failureHandlerWithNotNil := func() (err interface{}) {
		return "NotNilErr"
	}
	InitWithHandler(failureHandlerWithNotNil)
	// TODO: Simular llamado al controlador "/"
	if true {
		t.Fatalf("TEST POR IMPLEMENTAR - InitWithHandler no retornó {Status: ERROR: ...} con retorno no nulo")
	}
}
