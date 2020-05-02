// Package gha provides functions to interact with the GitHub Actions runtime.
package gha

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/sethvargo/go-githubactions"
)

// InGitHubActions indicates whether this application is being run within the
// GitHub Actions environment.
func InGitHubActions() bool {
	return os.Getenv("GITHUB_ACTIONS") == "true"
}

// PopulateFromInputs will populate the given struct with inputs supplied by
// the GitHub Actions environment. Fields that should be populated must be
// tagged with `gha:"<name of input>"`. If the empty string is given (`gha:""`)
// the field name will be used as input key.
//
// Additional options can be supplied through the tags, separated by comma's:
//  - required: returns an error if the input is not present or empty string
//
// Example struct:
//
//     type Example struct {
//         RunID     string `gha:"run-id,required"`
//         Directory string `gha:""`
//         DryRun    bool   `gha:"dry-run"`
//     }
//
func PopulateFromInputs(v interface{}) (err error) {
	rv := reflect.ValueOf(v)

	if rv.Kind() != reflect.Ptr || rv.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("invalid type %v, must be a pointer to a struct", reflect.TypeOf(v))
	}

	structType := reflect.TypeOf(v).Elem()

	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)

		tag := field.Tag.Get("gha")

		inputName, isRequired := parseTagOptions(tag)
		if inputName == "" {
			inputName = field.Name
		}

		value := githubactions.GetInput(inputName)

		if isRequired && value == "" {
			return fmt.Errorf("field %v is required but was not supplied", field.Name)
		}

		valueField := rv.Elem().Field(i)

		if !valueField.CanSet() {
			return fmt.Errorf("field %v can not be set, is it exported?", field.Name)
		}

		switch valueField.Kind() {
		case reflect.String:
			valueField.SetString(value)
		case reflect.Bool:
			boolValue, err := strconv.ParseBool(value)
			if err != nil {
				return fmt.Errorf("could not parse input for field %v as bool, value: %v: %w", field.Name, value, err)
			}
			valueField.SetBool(boolValue)
		default:
			return fmt.Errorf("fields of type %v are not supported, only strings and booleans are", valueField.Kind())
		}
	}

	return nil
}

func parseTagOptions(tag string) (inputName string, isRequired bool) {
	if tag == "" {
		return "", false
	}
	splitTag := strings.Split(tag, ",")
	inputName, options := splitTag[0], splitTag[1:]

	for _, option := range options {
		if option == "required" {
			isRequired = true
			break
		}
	}

	return
}

// WriteOutput writes an output parameter.
func WriteOutput(name, value string) {
	githubactions.SetOutput(name, value)
}
