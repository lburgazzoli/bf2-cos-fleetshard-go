package decoder

import (
	"github.com/mitchellh/mapstructure"
	"net/url"
	"reflect"
)

// StringToURLHookFunc returns a DecodeHookFunc that converts
// strings to url.URL.
func StringToURLHookFunc() mapstructure.DecodeHookFunc {
	return func(
		f reflect.Type,
		t reflect.Type,
		data interface{}) (interface{}, error) {
		if f.Kind() != reflect.String {
			return data, nil
		}
		if t != reflect.TypeOf(url.URL{}) {
			return data, nil
		}

		// Convert it by parsing
		return url.Parse(data.(string))
	}
}

// BytesToURLHookFunc returns a DecodeHookFunc that converts
// []byte] to url.URL.
func BytesToURLHookFunc() mapstructure.DecodeHookFunc {
	return func(
		f reflect.Type,
		t reflect.Type,
		data interface{}) (interface{}, error) {
		if f != reflect.TypeOf([]byte{}) {
			return data, nil
		}
		if t != reflect.TypeOf(url.URL{}) {
			return data, nil
		}

		// Convert it by parsing
		return url.Parse(data.(string))
	}
}
