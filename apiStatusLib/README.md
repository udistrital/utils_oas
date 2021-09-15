# API Status Check

## Setup

```go
import (
  ...
  apistatus "github.com/udistrital/utils_oas/apiStatusLib"
)
```

## Usage in "basic APIs": `Init()`

For basic MID APIs that, as example "only perfoms requests to another APIs"
(Note that this can even be checked with an specific healthcheck)

Usage Example:

```go
func main() {
  ...
  apistatus.Init()
  beego.Run()
}
```

## Accepting a handler function: `InitWithHandler(errorCheckHandler)`

For another kind of APIs that would require a handler that performs the healthcheck.

When specified (`errorCheckHandler != nil`) only will return `Status: Ok` when the `errorCheckHandler`
returns `nil`. Otherwise, let it be `!= nil` or whenever the `errorCheckHandler` crashes (throwing a `panic()`)
that will be indicated in the status

Usage Example:

```go
// statusCheckHandler better to be in another file. Even inside another
// package (folder) would also be fine. Keep in mind that this handler
// function can be called many times as it would be being called on every GET "/"
func statusCheckHandler() (checkError interface{}) {
  ... // Perform healthchecks

  // Once checked everything, return one of the following ones:
  return nil // when everything is OK
  return // (Same as above if checkError was left "pristine")
  return ... // any interface{}: string, a map[string]interface{}, ...

  // Alternatively/complementarily, if this healthcheck crashes and/or throws
  // a panic, the panic() will be catched and shown
  panic(...) // any interface{}: string, error, a map[string]interface{}, ...
}

func main() {
  ...
  apistatus.InitWithHandler(statusCheckHandler)
  beego.Run()
}
```
