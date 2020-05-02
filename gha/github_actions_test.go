package gha

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testStruct struct {
	Required   string `gha:"required-field,required"`
	Optional   string `gha:"optional-field"`
	WithoutTag string
	Boolean    bool `gha:"boolean"`
}

func TestPopulateFromInputs(t *testing.T) {
	os.Clearenv()
	os.Setenv("INPUT_REQUIRED-FIELD", "foo")
	os.Setenv("INPUT_WITHOUTTAG", "bar")
	os.Setenv("INPUT_BOOLEAN", "true")

	var ts testStruct

	err := PopulateFromInputs(&ts)

	assert.NoError(t, err)
	assert.Equal(t, "foo", ts.Required)
	assert.Equal(t, "", ts.Optional)
	assert.Equal(t, "bar", ts.WithoutTag)
	assert.Equal(t, true, ts.Boolean)
}

func TestPopulateFromInputs_invalidInputType(t *testing.T) {
	os.Clearenv()

	aString := "foo"
	values := []interface{}{nil, 5, aString, &aString, testStruct{}}

	for _, v := range values {
		err := PopulateFromInputs(v)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid type")
	}
}

func TestPopulateFromInputs_invalidBooleanInput(t *testing.T) {
	os.Clearenv()
	os.Setenv("INPUT_REQUIRED-FIELD", "foo")
	os.Setenv("INPUT_BOOLEAN-FIELD", "bar")

	var ts testStruct

	err := PopulateFromInputs(&ts)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "could not parse input for field Boolean as bool")
}

func TestPopulateFromInputs_missingRequiredField(t *testing.T) {
	os.Clearenv()

	var ts testStruct

	err := PopulateFromInputs(&ts)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "is required but was not supplied")
}

func TestPopulateFromInputs_unsupportedFieldType(t *testing.T) {
	os.Clearenv()

	type unsupportedStruct struct {
		Number int `gha:"number"`
	}

	var testStruct unsupportedStruct

	err := PopulateFromInputs(&testStruct)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "fields of type int are not supported")
}
