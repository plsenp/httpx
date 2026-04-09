package openapi

import (
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3gen"
)

// Spec wraps OpenAPI specification
type Spec struct {
	mu      sync.RWMutex
	spec    *openapi3.T
	gen     *openapi3gen.Generator
	servers []*openapi3.Server
}

// Config for Spec
type Config struct {
	Title          string
	Description    string
	Version        string
	OpenAPIVersion string
}

// DefaultConfig returns default configuration
func DefaultConfig() Config {
	return Config{
		Title:          "API",
		Description:    "Auto-generated API documentation",
		Version:        "1.0.0",
		OpenAPIVersion: "3.0.3",
	}
}

// NewSpec creates a new OpenAPI Spec
func NewSpec(config Config) *Spec {
	if config.OpenAPIVersion == "" {
		config.OpenAPIVersion = "3.0.3"
	}

	return &Spec{
		spec: &openapi3.T{
			OpenAPI: config.OpenAPIVersion,
			Info: &openapi3.Info{
				Title:       config.Title,
				Description: config.Description,
				Version:     config.Version,
			},
			Paths: openapi3.NewPaths(),
			Components: &openapi3.Components{
				Schemas: make(openapi3.Schemas),
			},
		},
		gen: openapi3gen.NewGenerator(
			openapi3gen.UseAllExportedFields(),
			openapi3gen.SchemaCustomizer(schemaCustomizer),
		),
	}
}

// schemaCustomizer is a callback function that customizes schema during generation
// It extracts description and example from struct tags
func schemaCustomizer(name string, t reflect.Type, tag reflect.StructTag, schema *openapi3.Schema) error {
	// Skip root level
	if name == "_root" {
		return nil
	}

	temp := make([]string, 0, 2)

	// Extract description
	if desc := tag.Get("description"); desc != "" {
		temp = append(temp, desc)
	}
	if vd := tag.Get("validate"); vd != "" {
		temp = append(temp, "[constraint: "+vd+"]")
	}
	if len(temp) > 0 {
		schema.Description = strings.Join(temp, "\n")
	}

	// Extract example
	if example := tag.Get("example"); example != "" {
		schema.Example = parseExample(example, t)
	}

	return nil
}

// SetInfo sets API information
func (s *Spec) SetInfo(title, description, version string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.spec.Info.Title = title
	s.spec.Info.Description = description
	s.spec.Info.Version = version
}

// AddServer adds a server
func (s *Spec) AddServer(url, description string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.servers = append(s.servers, &openapi3.Server{
		URL:         url,
		Description: description,
	})
	s.spec.Servers = s.servers
}

// RegisterOperation registers an operation to OpenAPI spec
func (s *Spec) RegisterOperation(method, path, summary, description string, tags []string, reqType, respType any) {
	s.mu.Lock()
	defer s.mu.Unlock()

	op := openapi3.NewOperation()
	op.Summary = summary
	op.Description = description
	op.Tags = tags

	// Extract path parameters
	pathParams := extractPathParams(path)
	for _, paramName := range pathParams {
		param := &openapi3.ParameterRef{
			Value: &openapi3.Parameter{
				Name:        paramName,
				In:          "path",
				Required:    true,
				Description: "Path parameter " + paramName,
				Schema: &openapi3.SchemaRef{
					Value: openapi3.NewStringSchema(),
				},
			},
		}
		op.Parameters = append(op.Parameters, param)
	}

	// Handle request body for non-GET/DELETE/HEAD methods
	if reqType != nil && method != "GET" && method != "DELETE" && method != "HEAD" {
		reqRef, err := s.gen.NewSchemaRefForValue(reqType, s.spec.Components.Schemas)
		if err == nil && reqRef != nil {
			requiredFields := extractRequiredFields(reqType)
			if len(requiredFields) > 0 && reqRef.Value != nil {
				reqRef.Value.Required = requiredFields
			}

			// Extract content types from request struct tags
			reqContentTypes := extractContentTypes(reqType)
			reqContent := buildContent(reqRef, reqContentTypes)

			op.RequestBody = &openapi3.RequestBodyRef{
				Value: &openapi3.RequestBody{
					Required: true,
					Content:  reqContent,
				},
			}
		}
	}

	// Handle query parameters for GET requests
	if method == "GET" && reqType != nil {
		queryParams := extractQueryParams(reqType)
		for _, param := range queryParams {
			op.Parameters = append(op.Parameters, param)
		}
	}

	// Handle response
	respDesc := description
	if respDesc == "" {
		respDesc = "Success"
	}

	if respType != nil {
		respRef, err := s.gen.NewSchemaRefForValue(respType, s.spec.Components.Schemas)
		if err == nil && respRef != nil {
			// Extract content types from response struct tags
			respContentTypes := extractContentTypes(respType)
			respContent := buildContent(respRef, respContentTypes)

			op.Responses = openapi3.NewResponses(
				openapi3.WithStatus(200, &openapi3.ResponseRef{
					Value: &openapi3.Response{
						Description: &respDesc,
						Content:     respContent,
					},
				}),
			)
		}
	} else {
		op.Responses = openapi3.NewResponses(
			openapi3.WithStatus(200, &openapi3.ResponseRef{
				Value: &openapi3.Response{
					Description: &respDesc,
				},
			}),
		)
	}

	s.spec.AddOperation(path, method, op)
}

// MarshalJSON returns JSON format of OpenAPI spec
func (s *Spec) MarshalJSON() ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.spec.MarshalJSON()
}

// Spec returns the underlying openapi3.T (read-only, do not modify)
func (s *Spec) Spec() *openapi3.T {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.spec
}

// extractPathParams extracts path parameters from path, e.g., /users/{id}
func extractPathParams(path string) []string {
	var params []string
	parts := strings.Split(path, "/")
	for _, part := range parts {
		if len(part) > 2 && part[0] == '{' && part[len(part)-1] == '}' {
			params = append(params, part[1:len(part)-1])
		}
	}
	return params
}

// extractQueryParams extracts query parameters from struct tags
func extractQueryParams(reqType any) []*openapi3.ParameterRef {
	var params []*openapi3.ParameterRef

	t := reflect.TypeOf(reqType)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return params
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Skip fields with path tag (they are path parameters, already handled)
		if tag := field.Tag.Get("path"); tag != "" {
			continue
		}

		// Check query tag
		if tag := field.Tag.Get("query"); tag != "" && tag != "-" {
			param := &openapi3.ParameterRef{
				Value: &openapi3.Parameter{
					Name:        tag,
					In:          "query",
					Description: field.Tag.Get("description"),
					Schema:      goTypeToSchema(field.Type),
				},
			}
			params = append(params, param)
		}
	}

	return params
}

// extractRequiredFields extracts required fields based on validate tag
func extractRequiredFields(reqType any) []string {
	var required []string

	t := reflect.TypeOf(reqType)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return required
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Check json tag for field name
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}

		fieldName := strings.Split(jsonTag, ",")[0]

		// Check if required
		validateTag := field.Tag.Get("validate")
		if strings.Contains(validateTag, "required") {
			required = append(required, fieldName)
		}
	}

	return required
}

// goTypeToSchema converts Go type to OpenAPI schema
func goTypeToSchema(t reflect.Type) *openapi3.SchemaRef {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	switch t.Kind() {
	case reflect.String:
		return &openapi3.SchemaRef{Value: openapi3.NewStringSchema()}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return &openapi3.SchemaRef{Value: openapi3.NewInt64Schema()}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return &openapi3.SchemaRef{Value: openapi3.NewIntegerSchema()}
	case reflect.Float32, reflect.Float64:
		return &openapi3.SchemaRef{Value: openapi3.NewFloat64Schema()}
	case reflect.Bool:
		return &openapi3.SchemaRef{Value: openapi3.NewBoolSchema()}
	case reflect.Slice, reflect.Array:
		return &openapi3.SchemaRef{Value: openapi3.NewArraySchema()}
	default:
		return &openapi3.SchemaRef{Value: openapi3.NewStringSchema()}
	}
}

// contentTypeMapping maps struct tag names to MIME types
var contentTypeMapping = map[string]string{
	"json":      "application/json",
	"xml":       "application/xml",
	"yaml":      "application/yaml",
	"yml":       "application/yaml",
	"form":      "application/x-www-form-urlencoded",
	"multipart": "multipart/form-data",
	"protobuf":  "application/protobuf",
	"proto":     "application/protobuf",
	"msgpack":   "application/msgpack",
	"bson":      "application/bson",
	"toml":      "application/toml",
}

// extractContentTypes extracts supported content types from struct tags
// It checks for various serialization format tags and returns corresponding MIME types
func extractContentTypes(v any) []string {
	if v == nil {
		return []string{"application/json"}
	}

	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return []string{"application/json"}
	}

	contentTypes := make(map[string]bool)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		for tagName, mimeType := range contentTypeMapping {
			if tag := field.Tag.Get(tagName); tag != "" && tag != "-" {
				contentTypes[mimeType] = true
			}
		}
	}

	// Default to JSON if no tags found
	if len(contentTypes) == 0 {
		return []string{"application/json"}
	}

	// Convert map to slice (ensure consistent order by putting JSON first)
	result := make([]string, 0, len(contentTypes))
	if contentTypes["application/json"] {
		result = append(result, "application/json")
	}
	for ct := range contentTypes {
		if ct != "application/json" {
			result = append(result, ct)
		}
	}
	return result
}

// buildContent builds Content map for request/response body
func buildContent(schemaRef *openapi3.SchemaRef, contentTypes []string) openapi3.Content {
	content := make(openapi3.Content)
	for _, ct := range contentTypes {
		content[ct] = &openapi3.MediaType{
			Schema: schemaRef,
		}
	}
	return content
}

// parseExample parses example string based on field type
func parseExample(example string, fieldType reflect.Type) any {
	// Dereference pointer if needed
	if fieldType.Kind() == reflect.Ptr {
		fieldType = fieldType.Elem()
	}

	switch fieldType.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if v, err := strconv.ParseInt(example, 10, 64); err == nil {
			return v
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if v, err := strconv.ParseUint(example, 10, 64); err == nil {
			return v
		}
	case reflect.Float32, reflect.Float64:
		if v, err := strconv.ParseFloat(example, 64); err == nil {
			return v
		}
	case reflect.Bool:
		if v, err := strconv.ParseBool(example); err == nil {
			return v
		}
	default:
		// For strings and other types, return as-is
		return example
	}
	// If parsing fails, return as string
	return example
}
