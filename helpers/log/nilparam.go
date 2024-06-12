package log

import "fmt"

// Input is nil error

type ParameterNilError struct {
	Name string
}

func (pne ParameterNilError) Error() string {
	if pne.Name == "" {
		return fmt.Sprintf("argument error: Unknown parameter is nil")

	}
	return fmt.Sprintf("argument error: The %s is nil", pne.Name)
}
func PanicIfNil(arr ...interface{}) {
	err := errorIfNil(arr...)
	if err != nil {
		Panic0(err)
	}
}
func errorIfNil(arr ...interface{}) error {
	var str = "parameter 0"
	for i, v := range arr {
		if v != nil {
			t, ok0 := (v).([]interface{})
			u, ok1 := (v).(*interface{})
			w, ok2 := (v).(string)
			if ok2 {
				str = w
			} else {
				if ok0 && t == nil {
					return ParameterNilError{str}
				}
				if ok1 && u == nil {
					return ParameterNilError{str}
				}
				str = fmt.Sprintf("parameter %d", i)
			}
		} else {
			return ParameterNilError{str}
		}
	}
	return nil
}
