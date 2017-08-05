package millumin

import (
	"fmt"
	"github.com/hypebeast/go-osc/osc"
	"reflect"
	"regexp"
)

func parseAddressLayer(msg *osc.Message, layerPtr *string) error {
	return parseAddress(msg, regexp.MustCompile("/layer:(.*?)/"), layerPtr)
}

func parseAddress(msg *osc.Message, pattern *regexp.Regexp, values ...*string) error {
	if matches := pattern.FindStringSubmatch(msg.Address); matches == nil {
		return fmt.Errorf("Address did not match pattern %v: %v", pattern, msg.Address)
	} else if len(matches) != len(values)+1 {
		return fmt.Errorf("Address pattern values count %d mismatch: %d", len(values), len(matches)-1)
	} else {
		for i, value := range matches {
			if i == 0 {
				continue
			}

			*values[i-1] = value
		}

		return nil
	}
}

func unmarshalArgument(msg *osc.Message, argIndex int, value interface{}) error {
	var valueType = reflect.TypeOf(value)

	if valueType.Kind() != reflect.Ptr {
		panic("value is not a pointer")
	}

	if msg.CountArguments() <= argIndex {
		return fmt.Errorf("Missing argument %d", argIndex)
	}

	var arg = msg.Arguments[argIndex]
	var argValue = reflect.ValueOf(arg)
	var valueValue = reflect.ValueOf(value)

	if argValue.Type().Kind() != valueType.Elem().Kind() {
		return fmt.Errorf("Invalid arugment %d: expected %v, got %T: %#v", argIndex, valueType.Elem(), arg, arg)
	}

	valueValue.Elem().Set(argValue)

	return nil
}
