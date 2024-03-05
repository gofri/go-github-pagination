package searchresult

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// see the main README.md for an explanation of this file.
type Untyped struct {
	TotalCount        int              `json:"total_count"`
	IncompleteResults bool             `json:"incomplete_results"`
	Items             *json.RawMessage `json:"items"`
}

type Typed[DataType any] struct {
	TotalCount        int
	IncompleteResults bool
	Items             []*DataType
}

func UntypedToTyped[DataType any](untyped *Untyped) (*Typed[DataType], error) {
	var items []*DataType
	err := json.Unmarshal(*untyped.Items, &items)
	if err != nil {
		return nil, err
	}

	return &Typed[DataType]{
		TotalCount:        untyped.TotalCount,
		IncompleteResults: untyped.IncompleteResults,
		Items:             items,
	}, nil
}

func FromSlice[DataType any](slice []*DataType) *Typed[DataType] {
	return &Typed[DataType]{
		TotalCount:        len(slice),
		IncompleteResults: false,
		Items:             slice,
	}
}

func VerifySearchType[DataType any](rType reflect.Type) error {
	if rType.Kind() == reflect.Ptr && rType.Elem().Kind() != reflect.Struct {
		return errors.New("search response must be a pointer to a struct")
	}

	expectedTags := map[string]reflect.Type{
		"total_count":        reflect.PointerTo(reflect.TypeOf(int(0))),
		"incomplete_results": reflect.PointerTo(reflect.TypeOf(false)),
		"items":              reflect.TypeOf([]*DataType{}),
	}
	numOfFields := rType.Elem().NumField()
	if got, want := numOfFields, len(expectedTags); got < want {
		return fmt.Errorf("search response must have %v fields, got %v", want, got)
	}
	found := 0
	for i := 0; i < numOfFields; i++ {
		rField := rType.Elem().Field(i)
		tag := rField.Tag.Get("json")
		tagName := strings.Split(tag, ",")[0]
		if expectedType, exists := expectedTags[tagName]; !exists {
			continue
		} else if expectedType == nil {
			return fmt.Errorf("search response field has duplicate tag: %v", tagName)
		} else if rField.Type != expectedType {
			return fmt.Errorf("search response field %v must be %v, got %v", tagName, expectedType, rField.Type)
		}
		expectedTags[tagName] = nil
		found++
	}

	if found < len(expectedTags) {
		return fmt.Errorf("search response must have all of the following fields: %v", expectedTags)
	}

	return nil
}
