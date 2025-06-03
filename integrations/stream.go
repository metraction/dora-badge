package integrations

func NewMono(element any) chan any {
	outChan := make(chan any)
	go func() {
		outChan <- element
		close(outChan)
	}()
	return outChan
}

type ResponseWithError[T any] struct {
	Response T
	Err      error
}

func Right[T any](e ResponseWithError[T]) T {
	return e.Response
}

func ErrorCollector[T any](errors chan any) func(e ResponseWithError[T]) bool {
	return func(e ResponseWithError[T]) bool {
		if e.Err != nil {
			errors <- e.Err
			return false
		}
		return true
	}
}
