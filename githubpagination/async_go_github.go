package githubpagination

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"sync"

	"github.com/gofri/go-github-pagination/githubpagination/drivers"
	"github.com/gofri/go-github-pagination/githubpagination/searchresult"
)

// NewAsync creates a new Async instance for non-search results.
// Note: you can use this with search results, but the incomplete_results and total_count fields will not be available.
func NewAsync[DataType any](onNext OnNextResponseSlice[DataType]) *Async[DataType] {
	adapter := func(resp *http.Response, result *searchresult.Typed[DataType]) error {
		return onNext(resp, result.Items)
	}
	return NewAsyncSearch(adapter)
}

// NewAsyncSearch creates a new Async instance for search results.
// It is designed to be used with search results,
// so that the incomplete_results and total_count fields are available.
func NewAsyncSearch[DataType any](onNext OnNextResponse[DataType]) *Async[DataType] {
	return &Async[DataType]{
		OnNext:  onNext,
		errChan: make(chan error, 1),
	}
}

type OnNextResponse[DataType any] func(*http.Response, *searchresult.Typed[DataType]) error
type OnNextResponseSlice[DataType any] func(*http.Response, []*DataType) error

type Async[DataType any] struct {
	OnNext        OnNextResponse[DataType]
	errChan       chan error
	closeChanLock sync.Mutex
	closed        bool
}

// Paginate paginates through the results of a request function.
func (a *Async[DataType]) Paginate(requestFn any, args ...any) error {
	respType, err := a.getValidatedResponseDataType(requestFn)
	if err != nil {
		return err
	}
	isSearch := respType == searchresult.ResponseDataTypeSearch
	effectiveArgs := a.withAsyncCtx(requestFn, isSearch, args...)

	rValues := reflect.ValueOf(requestFn).Call(effectiveArgs)
	if err := <-a.errChan; err != nil {
		return err
	}

	if err := a.getReturnedError(rValues); err != nil {
		return err
	}

	return nil
}

func (a *Async[DataType]) HandlePage(data *searchresult.Typed[DataType], resp *http.Response) error {
	if a.OnNext != nil {
		return a.OnNext(resp, data)
	}
	return nil
}

func (a *Async[DataType]) HandleError(resp *http.Response, err error) {
	a.closeErrChan(err)
}

func (a *Async[DataType]) HandleFinish(resp *http.Response, pageCount int) {
	a.closeErrChan(nil)
}

func (a *Async[DataType]) closeErrChan(err error) {
	a.closeChanLock.Lock()
	defer a.closeChanLock.Unlock()
	if a.closed {
		// prevent deadlock in case of more than a single error
		return
	}
	a.errChan <- err
	close(a.errChan)
	a.closed = true
}

const (
	returnIndexData     = 0
	returnIndexResponse = 1
	returnIndexError    = 2
	returnValuesCount   = 3
	argsIndexContext    = 0
)

func (a *Async[DataType]) getReturnedError(rValues []reflect.Value) error {
	goGithubErr := rValues[returnIndexError].Interface()
	if goGithubErr == nil {
		return nil
	}
	return goGithubErr.(error)
}

func (a *Async[DataType]) getValidatedResponseDataType(requestFn any) (searchresult.ResponseDataType, error) {
	respDataType := searchresult.ResponseDataTypeUnknown

	// Check if the requestFn is a function.
	fnType := reflect.TypeOf(requestFn)
	if fnType.Kind() != reflect.Func {
		return respDataType, errors.New("not a function")
	}

	// Check if the requestFn returns the correct values.
	if fnType.NumOut() != returnValuesCount {
		return respDataType, errors.New("request function must return 3 values")
	}
	respDataType = searchresult.GetResponseDataType[DataType](fnType.Out(returnIndexData))
	if respDataType == searchresult.ResponseDataTypeUnknown {
		return respDataType, errors.New("first return value must be either a slice of pointers or a search result type")
	}
	if second := fnType.Out(returnIndexResponse).String(); second != "*github.Response" {
		return respDataType, fmt.Errorf("second return value must be *http.Response, got %v", second)
	}
	if fnType.Out(returnIndexError).String() != "error" {
		return respDataType, errors.New("third return value must be error")
	}

	// Check if the requestFn accepts context.Context as the first argument.
	if fnType.In(argsIndexContext).String() != "context.Context" {
		return respDataType, errors.New("first argument must be context.Context")
	}

	return respDataType, nil
}

func (a *Async[DataType]) withAsyncCtx(requestFn any, isSearch bool, args ...any) []reflect.Value {
	reflected := make([]reflect.Value, 0, len(args))

	ctx := WithOverrideConfig(args[argsIndexContext].(context.Context),
		WithDriver(drivers.NewGithubAsyncPaginationDriver(a, isSearch)),
		WithPaginationEnabled(), // make sure that pagination is enabled
	)

	for i, arg := range args {
		if i == 0 {
			reflected = append(reflected, reflect.ValueOf(ctx))
			continue
		}
		rValue := reflect.ValueOf(arg)
		if !rValue.IsValid() { // fix nil values
			rValue = reflect.New(reflect.TypeOf(requestFn).In(i)).Elem()
		}
		reflected = append(reflected, rValue)
	}
	return reflected
}
