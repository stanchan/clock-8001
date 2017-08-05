package millumin

import (
	"fmt"
	"github.com/hypebeast/go-osc/osc"
	"reflect"
)

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
