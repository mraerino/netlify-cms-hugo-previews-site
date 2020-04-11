package hugojs

import (
	"syscall/js"

	"github.com/pkg/errors"
)

type jsResult struct {
	err  error
	vals []js.Value
}

func jsCallback() (js.Func, <-chan jsResult) {
	resultCh := make(chan jsResult)
	newErr := func(err error) interface{} {
		resultCh <- jsResult{err, nil}
		return nil
	}
	fn := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		defer close(resultCh)
		if len(args) < 1 {
			return newErr(errors.New("too few arguments"))
		}

		errArg := args[0]
		var res jsResult
		if errArg.Type() == js.TypeString {
			if errStr := errArg.String(); errStr != "" {
				res.err = errors.New(errStr)
			}
			res.vals = args[1:]
		} else if errArg.Type() == js.TypeNull {
			res.vals = args
		} else {
			res.err = errors.New("invalid error type")
		}

		resultCh <- res
		close(resultCh)
		return nil
	})
	return fn, resultCh
}
