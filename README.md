# Auditoria
Repositorio que contiene la líbreria de middleware, la cual permite capturar la traza en cada transacción

## Middleware in Beego


Basic example for creating middlewares in Beego framework.

### Installation

Standard `go get`:

```
$ go get github.com/udistrital/auditoria
```

#### Dependencies

```
  go get github.com/astaxie/beego
```

#### Configuring Beego Middleware

Add the following lines into the ```routers/routers.go``` file which will initialize the filter to run on all requests (BeforeStatic, BeforeRouter, BeforeStatic, AfterExec and FinishRouter)


```go
 import "github.com/udistrital/auditoria"

 func init() {
    auditoria.InitMiddleware()


 }
```
