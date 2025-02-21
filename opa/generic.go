package main

import "reflect"

func getFieldsFromInstance(instance interface{}) []string {

	result := []string{}

	fields := reflect.VisibleFields(reflect.TypeOf(instance))

	for _, field := range fields {

		result = append(result, field.Type.Name())

	}

	return result

}

func getValueFromField(instance interface{}, fieldName string) interface{} {

	reflectedValue := reflect.ValueOf(instance)

	value := reflect.Indirect(reflectedValue).FieldByName(fieldName)

	return value.Interface()

}
