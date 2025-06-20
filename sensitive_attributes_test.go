package main

import (
	"testing"

	"github.com/camptocamp/terraboard/db"
	"github.com/camptocamp/terraboard/internal/terraform/states"
	"github.com/camptocamp/terraboard/types"
	"github.com/stretchr/testify/assert"
	"github.com/zclconf/go-cty/cty"
)

func TestSensitiveAttributesEndToEnd(t *testing.T) {
	// Test the complete flow of sensitive attributes from state parsing to API response
	
	// Create a mock resource instance with sensitive paths
	src := &states.ResourceInstanceObjectSrc{
		AttrsJSON: []byte(`{"password":"secret123","username":"user","api_key":"key456"}`),
		AttrSensitivePaths: []cty.PathValueMarks{
			{
				Path:  cty.Path{cty.GetAttrStep{Name: "password"}},
				Marks: cty.NewValueMarks("sensitive"),
			},
			{
				Path:  cty.Path{cty.GetAttrStep{Name: "api_key"}},
				Marks: cty.NewValueMarks("sensitive"),
			},
		},
		Status: states.ObjectReady,
	}

	// Use the marshalAttributeValues function to process the attributes
	attrs := db.MarshalAttributeValues(src)

	// Verify that we have 3 attributes
	assert.Equal(t, 3, len(attrs))

	// Create a map for easier checking
	attrMap := make(map[string]types.Attribute)
	for _, attr := range attrs {
		attrMap[attr.Key] = attr
	}

	// Verify that password and api_key are marked as sensitive
	assert.True(t, attrMap["password"].Sensitive, "password should be marked as sensitive")
	assert.True(t, attrMap["api_key"].Sensitive, "api_key should be marked as sensitive")
	assert.False(t, attrMap["username"].Sensitive, "username should not be marked as sensitive")

	// Verify that the values are stored correctly
	assert.Equal(t, "\"secret123\"", attrMap["password"].Value)
	assert.Equal(t, "\"key456\"", attrMap["api_key"].Value)
	assert.Equal(t, "\"user\"", attrMap["username"].Value)
}