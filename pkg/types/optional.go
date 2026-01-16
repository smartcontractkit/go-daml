package types

func WrapOptional(value interface{}) map[string]interface{} {
	if value == nil {
		return map[string]interface{}{"_type": "optional"}
	}
	return map[string]interface{}{
		"_type": "optional",
		"value": value,
	}
}

func WrapOptionalPointer[T any](ptr *T) map[string]interface{} {
	if ptr == nil {
		return map[string]interface{}{"_type": "optional"}
	}
	return map[string]interface{}{
		"_type": "optional",
		"value": *ptr,
	}
}
