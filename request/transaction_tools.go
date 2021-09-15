package request

const rethrow_panic = "_____rethrow"

type Excep struct {
	error interface{}
}

func Commit(f func()) (err Excep) {
	defer func() {
		err.error = recover()
	}()
	f()
	return
}

func (err Excep) Rollback(f func(response interface{}, excep interface{}), params interface{}) {
	if err.error != nil {
		defer func() {
			if excep := recover(); excep != nil {
				if excep == rethrow_panic {
					excep = err.error
				}
				panic(excep)
			}
		}()
		f(params, err.error)
	}
}
