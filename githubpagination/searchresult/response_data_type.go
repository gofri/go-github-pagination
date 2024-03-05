package searchresult

import (
	"reflect"
)

type ResponseDataType string

const (
	ResponseDataTypeUnknown ResponseDataType = ""
	ResponseDataTypeSliced  ResponseDataType = "sliced"
	ResponseDataTypeSearch  ResponseDataType = "search"
)

func GetResponseDataType[DataType any](rType reflect.Type) ResponseDataType {
	if rType.Kind() == reflect.Slice &&
		rType.Elem().Kind() == reflect.Ptr &&
		rType.Elem().Elem() == reflect.TypeOf(*new(DataType)) {
		return ResponseDataTypeSliced
	}

	if err := VerifySearchType[DataType](rType); err == nil {
		return ResponseDataTypeSearch
	}

	return ResponseDataTypeUnknown
}
