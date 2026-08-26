package openapi

import (
	"strings"
	"testing"

	"github.com/swaggo/swag"
)

func TestSwaggerSpecIsRegisteredAndReadable(t *testing.T) {
	registered := swag.GetSwagger(SwaggerInfo.InstanceName())
	if registered == nil {
		t.Fatal("Swagger spec is not registered")
	}

	doc, err := swag.ReadDoc(SwaggerInfo.InstanceName())
	if err != nil {
		t.Fatalf("ReadDoc() error = %v", err)
	}
	if !strings.Contains(doc, `"/api/v1/streams"`) {
		t.Fatalf("rendered Swagger doc does not contain streams route")
	}
	if !strings.Contains(doc, `"title": "StreamPulse API"`) {
		t.Fatalf("rendered Swagger doc does not contain StreamPulse API title")
	}
}
