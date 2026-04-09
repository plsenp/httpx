package openapi

import (
	"encoding/json"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestRegisterOperation(t *testing.T) {
	spec := NewSpec(Config{
		Title:   "Test API",
		Version: "1.0.0",
	})

	type GetUserReq struct {
		ID int `path:"id"`
	}

	type GetUserResp struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	spec.RegisterOperation("GET", "/users/{id}", "Get User", "Get user by ID", []string{"Users"}, &GetUserReq{}, &GetUserResp{})

	s := spec.Spec()
	if s.Paths == nil {
		t.Fatal("Paths should not be nil")
	}

	pathItem := s.Paths.Find("/users/{id}")
	if pathItem == nil {
		t.Fatal("Path /users/{id} not found")
	}

	if pathItem.Get == nil {
		t.Fatal("GET operation not found")
	}

	if pathItem.Get.Summary != "Get User" {
		t.Errorf("Expected summary 'Get User', got '%s'", pathItem.Get.Summary)
	}

	if len(pathItem.Get.Tags) != 1 || pathItem.Get.Tags[0] != "Users" {
		t.Errorf("Expected tags ['Users'], got %v", pathItem.Get.Tags)
	}
}

func TestRegisterPOSTOperation(t *testing.T) {
	spec := NewSpec(Config{
		Title:   "Test API",
		Version: "1.0.0",
	})

	type CreateUserReq struct {
		Name  string `json:"name" validate:"required"`
		Email string `json:"email" validate:"required,email"`
	}

	type CreateUserResp struct {
		ID int `json:"id"`
	}

	spec.RegisterOperation("POST", "/users", "Create User", "Create a new user", []string{"Users"}, &CreateUserReq{}, &CreateUserResp{})

	s := spec.Spec()
	pathItem := s.Paths.Find("/users")
	if pathItem == nil {
		t.Fatal("Path /users not found")
	}

	if pathItem.Post == nil {
		t.Fatal("POST operation not found")
	}

	// Check request body
	if pathItem.Post.RequestBody == nil || pathItem.Post.RequestBody.Value == nil {
		t.Fatal("RequestBody not found")
	}

	// Check required fields
	reqSchema := pathItem.Post.RequestBody.Value.Content["application/json"].Schema.Value
	if len(reqSchema.Required) != 2 {
		t.Errorf("Expected 2 required fields, got %d", len(reqSchema.Required))
	}
}

func TestSetInfo(t *testing.T) {
	spec := NewSpec(DefaultConfig())
	spec.SetInfo("New Title", "New Description", "2.0.0")

	s := spec.Spec()
	if s.Info.Title != "New Title" {
		t.Errorf("Expected title 'New Title', got '%s'", s.Info.Title)
	}
	if s.Info.Description != "New Description" {
		t.Errorf("Expected description 'New Description', got '%s'", s.Info.Description)
	}
	if s.Info.Version != "2.0.0" {
		t.Errorf("Expected version '2.0.0', got '%s'", s.Info.Version)
	}
}

func TestAddServer(t *testing.T) {
	spec := NewSpec(DefaultConfig())
	spec.AddServer("https://api.example.com", "Production server")

	s := spec.Spec()
	if len(s.Servers) != 1 {
		t.Fatalf("Expected 1 server, got %d", len(s.Servers))
	}

	if s.Servers[0].URL != "https://api.example.com" {
		t.Errorf("Expected URL 'https://api.example.com', got '%s'", s.Servers[0].URL)
	}
}

func TestMarshalJSON(t *testing.T) {
	spec := NewSpec(Config{
		Title:   "Test API",
		Version: "1.0.0",
	})

	type GetUserReq struct {
		ID int `path:"id"`
	}

	type GetUserResp struct {
		Name string `json:"name"`
	}

	spec.RegisterOperation("GET", "/users/{id}", "Get User", "Get user by ID", []string{"Users"}, &GetUserReq{}, &GetUserResp{})

	data, err := spec.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	if len(data) == 0 {
		t.Error("JSON data should not be empty")
	}

	// Verify it's valid JSON
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Errorf("Invalid JSON: %v", err)
	}
}

func TestExtractPathParams(t *testing.T) {
	tests := []struct {
		path     string
		expected []string
	}{
		{"/users/{id}", []string{"id"}},
		{"/users/{userId}/posts/{postId}", []string{"userId", "postId"}},
		{"/users", []string{}},
		{"/users/{id}/profile", []string{"id"}},
	}

	for _, test := range tests {
		result := extractPathParams(test.path)
		if len(result) != len(test.expected) {
			t.Errorf("Path %s: expected %v, got %v", test.path, test.expected, result)
			continue
		}
		for i, v := range test.expected {
			if result[i] != v {
				t.Errorf("Path %s: expected %v, got %v", test.path, test.expected, result)
				break
			}
		}
	}
}

func TestExtractQueryParams(t *testing.T) {
	type ListUsersReq struct {
		Page  int    `query:"page" description:"Page number"`
		Limit int    `query:"limit"`
		ID    int    `path:"id"`
		Sort  string `query:"sort"`
	}

	params := extractQueryParams(&ListUsersReq{})

	if len(params) != 3 {
		t.Errorf("Expected 3 query params, got %d", len(params))
	}

	// Check that path param is not included
	for _, p := range params {
		if p.Value.Name == "id" {
			t.Error("Path parameter 'id' should not be in query params")
		}
	}

	// Check description is extracted
	for _, p := range params {
		if p.Value.Name == "page" && p.Value.Description != "Page number" {
			t.Errorf("Expected description 'Page number', got '%s'", p.Value.Description)
		}
	}
}

func TestExtractRequiredFields(t *testing.T) {
	type CreateUserReq struct {
		Name     string `json:"name" validate:"required"`
		Email    string `json:"email" validate:"required,email"`
		Optional string `json:"optional"`
		Age      int    `json:"age" validate:"required,min=0"`
	}

	required := extractRequiredFields(&CreateUserReq{})

	if len(required) != 3 {
		t.Errorf("Expected 3 required fields, got %d", len(required))
	}

	expected := map[string]bool{
		"name":  true,
		"email": true,
		"age":   true,
	}

	for _, field := range required {
		if !expected[field] {
			t.Errorf("Unexpected required field: %s", field)
		}
	}
}

// Test recursive types
type Menu struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Children []*Menu `json:"children,omitempty"`
}

func TestRecursiveType(t *testing.T) {
	spec := NewSpec(Config{
		Title:   "Test API",
		Version: "1.0.0",
	})

	type GetMenuReq struct {
		ID int `path:"id"`
	}

	spec.RegisterOperation("GET", "/menus/{id}", "Get Menu", "Get menu by ID", []string{"Menus"}, &GetMenuReq{}, &Menu{})

	s := spec.Spec()
	pathItem := s.Paths.Find("/menus/{id}")
	if pathItem == nil {
		t.Fatal("Path /menus/{id} not found")
	}

	if pathItem.Get == nil {
		t.Fatal("GET operation not found")
	}

	// Check if response schema is generated
	resp := pathItem.Get.Responses.Status(200)
	if resp == nil || resp.Value == nil || resp.Value.Content == nil {
		t.Fatal("Response content not found")
	}

	mediaType := resp.Value.Content["application/json"]
	if mediaType == nil || mediaType.Schema == nil || mediaType.Schema.Value == nil {
		t.Fatal("Schema not found in response")
	}

	schema := mediaType.Schema.Value

	// Check if schema has the expected properties
	if schema.Properties == nil {
		t.Fatal("Schema properties not found")
	}

	if _, ok := schema.Properties["id"]; !ok {
		t.Error("Schema should have 'id' property")
	}
	if _, ok := schema.Properties["name"]; !ok {
		t.Error("Schema should have 'name' property")
	}

	// Check for recursive 'children' property
	childrenProp, ok := schema.Properties["children"]
	if !ok {
		t.Error("Schema should have 'children' property")
	} else {
		if childrenProp.Value == nil || !childrenProp.Value.Type.Is("array") {
			t.Error("Children property should be an array type")
		}
	}
}

// Test deeply nested types
type Item struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Category struct {
	ID            int        `json:"id"`
	Name          string     `json:"name"`
	Description   string     `json:"description"`
	Items         []Item     `json:"items"`
	SubCategories []Category `json:"sub_categories"`
}

func TestDeepNestedType(t *testing.T) {
	spec := NewSpec(Config{
		Title:   "Test API",
		Version: "1.0.0",
	})

	type GetCategoryReq struct {
		ID int `path:"id"`
	}

	spec.RegisterOperation("GET", "/categories/{id}", "Get Category", "Get category by ID", []string{"Categories"}, &GetCategoryReq{}, &Category{})

	s := spec.Spec()
	pathItem := s.Paths.Find("/categories/{id}")
	if pathItem == nil {
		t.Fatal("Path /categories/{id} not found")
	}

	if pathItem.Get == nil {
		t.Fatal("GET operation not found")
	}

	// Check if response schema is generated
	resp := pathItem.Get.Responses.Status(200)
	if resp == nil || resp.Value == nil || resp.Value.Content == nil {
		t.Fatal("Response content not found")
	}

	mediaType := resp.Value.Content["application/json"]
	if mediaType == nil || mediaType.Schema == nil || mediaType.Schema.Value == nil {
		t.Fatal("Schema not found in response")
	}

	schema := mediaType.Schema.Value

	// Check if schema has the expected properties
	if schema.Properties == nil {
		t.Fatal("Schema properties not found")
	}

	// Check for basic properties
	if _, ok := schema.Properties["id"]; !ok {
		t.Error("Schema should have 'id' property")
	}
	if _, ok := schema.Properties["name"]; !ok {
		t.Error("Schema should have 'name' property")
	}
	if _, ok := schema.Properties["description"]; !ok {
		t.Error("Schema should have 'description' property")
	}

	// Check for recursive 'sub_categories' property
	subCatProp, ok := schema.Properties["sub_categories"]
	if !ok {
		t.Error("Schema should have 'sub_categories' property")
	} else {
		if subCatProp.Value == nil || !subCatProp.Value.Type.Is("array") {
			t.Error("SubCategories property should be an array type")
		}
	}

	// Check for 'items' property (nested struct slice)
	itemsProp, ok := schema.Properties["items"]
	if !ok {
		t.Error("Schema should have 'items' property")
	} else {
		if itemsProp.Value == nil || !itemsProp.Value.Type.Is("array") {
			t.Error("Items property should be an array type")
		}
	}
}

func TestRecursiveTypeMarshalJSON(t *testing.T) {
	spec := NewSpec(Config{
		Title:   "Test API",
		Version: "1.0.0",
	})

	type GetMenuReq struct {
		ID int `path:"id"`
	}

	spec.RegisterOperation("GET", "/menus/{id}", "Get Menu", "Get menu by ID", []string{"Menus"}, &GetMenuReq{}, &Menu{})

	data, err := spec.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	if len(data) == 0 {
		t.Error("JSON data should not be empty")
	}

	// Verify it's valid JSON and contains expected fields
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Errorf("Invalid JSON: %v", err)
	}

	// Check basic structure
	if result["openapi"] == nil {
		t.Error("OpenAPI version not found")
	}
	if result["info"] == nil {
		t.Error("Info not found")
	}
	if result["paths"] == nil {
		t.Error("Paths not found")
	}

	// Check if the paths contain our endpoint
	paths, ok := result["paths"].(map[string]interface{})
	if !ok {
		t.Fatal("Paths is not a map")
	}

	if paths["/menus/{id}"] == nil {
		t.Error("Path /menus/{id} not found in JSON")
	}
}

// Test types for content type extraction
type ArticleJSONOnly struct {
	ID      int    `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

type ArticleXML struct {
	ID      int    `json:"id" xml:"id"`
	Title   string `json:"title" xml:"title"`
	Content string `json:"content" xml:"content"`
}

type ArticleMultiFormat struct {
	ID      int    `json:"id" xml:"id" yaml:"id"`
	Title   string `json:"title" xml:"title" yaml:"title"`
	Content string `json:"content" xml:"content" yaml:"content"`
}

// Test types for various content types
type FormData struct {
	Name  string `form:"name"`
	Email string `form:"email"`
}

type MultipartForm struct {
	File     string `multipart:"file"`
	Filename string `multipart:"filename"`
}

type ProtobufData struct {
	ID   int    `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	Name string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
}

type MsgpackData struct {
	ID   int    `msgpack:"id"`
	Name string `msgpack:"name"`
}

type BSONData struct {
	ID   int    `bson:"id"`
	Name string `bson:"name"`
}

type TOMLData struct {
	ID   int    `toml:"id"`
	Name string `toml:"name"`
}

type YAMLData struct {
	ID   int    `yaml:"id"`
	Name string `yaml:"name"`
}

func TestExtractContentTypes(t *testing.T) {
	tests := []struct {
		name     string
		v        any
		expected []string
	}{
		{
			name:     "nil value defaults to JSON",
			v:        nil,
			expected: []string{"application/json"},
		},
		{
			name:     "JSON only struct",
			v:        &ArticleJSONOnly{},
			expected: []string{"application/json"},
		},
		{
			name:     "JSON and XML struct",
			v:        &ArticleXML{},
			expected: []string{"application/json", "application/xml"},
		},
		{
			name:     "Multi format struct",
			v:        &ArticleMultiFormat{},
			expected: []string{"application/json", "application/xml", "application/yaml"},
		},
		{
			name:     "Form data struct",
			v:        &FormData{},
			expected: []string{"application/x-www-form-urlencoded"},
		},
		{
			name:     "Multipart form struct",
			v:        &MultipartForm{},
			expected: []string{"multipart/form-data"},
		},
		{
			name:     "Protobuf data struct",
			v:        &ProtobufData{},
			expected: []string{"application/json", "application/protobuf"},
		},
		{
			name:     "Msgpack data struct",
			v:        &MsgpackData{},
			expected: []string{"application/msgpack"},
		},
		{
			name:     "BSON data struct",
			v:        &BSONData{},
			expected: []string{"application/bson"},
		},
		{
			name:     "TOML data struct",
			v:        &TOMLData{},
			expected: []string{"application/toml"},
		},
		{
			name:     "YAML data struct",
			v:        &YAMLData{},
			expected: []string{"application/yaml"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := extractContentTypes(test.v)
			if len(result) != len(test.expected) {
				t.Errorf("Expected %v, got %v", test.expected, result)
				return
			}
			// Use map for order-independent comparison
			resultMap := make(map[string]bool)
			for _, v := range result {
				resultMap[v] = true
			}
			for _, v := range test.expected {
				if !resultMap[v] {
					t.Errorf("Expected %v, got %v", test.expected, result)
					break
				}
			}
		})
	}
}

func TestBuildContent(t *testing.T) {
	schemaRef := &openapi3.SchemaRef{
		Value: &openapi3.Schema{
			Type: &openapi3.Types{"object"},
		},
	}

	contentTypes := []string{"application/json", "application/xml"}
	content := buildContent(schemaRef, contentTypes)

	if len(content) != 2 {
		t.Errorf("Expected 2 content types, got %d", len(content))
	}

	if _, ok := content["application/json"]; !ok {
		t.Error("Expected 'application/json' in content")
	}

	if _, ok := content["application/xml"]; !ok {
		t.Error("Expected 'application/xml' in content")
	}
}

func TestRegisterOperationWithMultipleContentTypes(t *testing.T) {
	spec := NewSpec(Config{
		Title:   "Test API",
		Version: "1.0.0",
	})

	// Register operation with multi-format request and response
	spec.RegisterOperation("POST", "/articles", "Create Article", "Create a new article", []string{"Articles"}, &ArticleXML{}, &ArticleXML{})

	s := spec.Spec()
	pathItem := s.Paths.Find("/articles")
	if pathItem == nil {
		t.Fatal("Path /articles not found")
	}

	if pathItem.Post == nil {
		t.Fatal("POST operation not found")
	}

	// Check request body content types
	if pathItem.Post.RequestBody == nil || pathItem.Post.RequestBody.Value == nil {
		t.Fatal("RequestBody not found")
	}

	reqContent := pathItem.Post.RequestBody.Value.Content
	if len(reqContent) != 2 {
		t.Errorf("Expected 2 request content types, got %d", len(reqContent))
	}

	if _, ok := reqContent["application/json"]; !ok {
		t.Error("Request should support 'application/json'")
	}

	if _, ok := reqContent["application/xml"]; !ok {
		t.Error("Request should support 'application/xml'")
	}

	// Check response content types
	resp := pathItem.Post.Responses.Status(200)
	if resp == nil || resp.Value == nil || resp.Value.Content == nil {
		t.Fatal("Response content not found")
	}

	respContent := resp.Value.Content
	if len(respContent) != 2 {
		t.Errorf("Expected 2 response content types, got %d", len(respContent))
	}

	if _, ok := respContent["application/json"]; !ok {
		t.Error("Response should support 'application/json'")
	}

	if _, ok := respContent["application/xml"]; !ok {
		t.Error("Response should support 'application/xml'")
	}
}

// Test schema enhancement with description and example tags
func TestEnhanceSchemaWithDescription(t *testing.T) {
	spec := NewSpec(Config{
		Title:   "Test API",
		Version: "1.0.0",
	})

	type CreateProductReq struct {
		Name        string  `json:"name" description:"Product name" example:"iPhone 15"`
		Description string  `json:"description" description:"Product description" example:"Latest iPhone model"`
		Price       float64 `json:"price" description:"Product price in USD" example:"999.99"`
		Category    string  `json:"category" description:"Product category" example:"electronics"`
		Email       string  `json:"email" description:"Contact email" example:"john@example.com"`
		InStock     bool    `json:"in_stock" description:"Whether product is in stock" example:"true"`
		Count       int     `json:"count" description:"Product count" example:"100"`
	}

	type CreateProductResp struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	spec.RegisterOperation("POST", "/products", "Create Product", "Create a new product", []string{"Products"}, &CreateProductReq{}, &CreateProductResp{})

	s := spec.Spec()
	pathItem := s.Paths.Find("/products")
	if pathItem == nil || pathItem.Post == nil {
		t.Fatal("POST operation not found")
	}

	// Check request body schema
	reqBody := pathItem.Post.RequestBody.Value.Content["application/json"]
	if reqBody == nil || reqBody.Schema == nil || reqBody.Schema.Value == nil {
		t.Fatal("Request schema not found")
	}

	schema := reqBody.Schema.Value

	// Check Name field description and example
	nameProp := schema.Properties["name"]
	if nameProp == nil || nameProp.Value == nil {
		t.Fatal("Name property not found")
	}
	if nameProp.Value.Description != "Product name" {
		t.Errorf("Expected description 'Product name', got '%s'", nameProp.Value.Description)
	}
	if nameProp.Value.Example != "iPhone 15" {
		t.Errorf("Expected example 'iPhone 15', got '%v'", nameProp.Value.Example)
	}

	// Check Description field
	descProp := schema.Properties["description"]
	if descProp == nil || descProp.Value == nil {
		t.Fatal("Description property not found")
	}
	if descProp.Value.Description != "Product description" {
		t.Errorf("Expected description 'Product description', got '%s'", descProp.Value.Description)
	}
	if descProp.Value.Example != "Latest iPhone model" {
		t.Errorf("Expected example 'Latest iPhone model', got '%v'", descProp.Value.Example)
	}

	// Check Price field (float example)
	priceProp := schema.Properties["price"]
	if priceProp == nil || priceProp.Value == nil {
		t.Fatal("Price property not found")
	}
	if priceProp.Value.Description != "Product price in USD" {
		t.Errorf("Expected description 'Product price in USD', got '%s'", priceProp.Value.Description)
	}
	if priceProp.Value.Example != 999.99 {
		t.Errorf("Expected example 999.99, got %v", priceProp.Value.Example)
	}

	// Check Category field
	catProp := schema.Properties["category"]
	if catProp == nil || catProp.Value == nil {
		t.Fatal("Category property not found")
	}
	if catProp.Value.Description != "Product category" {
		t.Errorf("Expected description 'Product category', got '%s'", catProp.Value.Description)
	}
	if catProp.Value.Example != "electronics" {
		t.Errorf("Expected example 'electronics', got '%v'", catProp.Value.Example)
	}

	// Check Email field
	emailProp := schema.Properties["email"]
	if emailProp == nil || emailProp.Value == nil {
		t.Fatal("Email property not found")
	}
	if emailProp.Value.Description != "Contact email" {
		t.Errorf("Expected description 'Contact email', got '%s'", emailProp.Value.Description)
	}
	if emailProp.Value.Example != "john@example.com" {
		t.Errorf("Expected example 'john@example.com', got '%v'", emailProp.Value.Example)
	}

	// Check InStock field (bool example)
	stockProp := schema.Properties["in_stock"]
	if stockProp == nil || stockProp.Value == nil {
		t.Fatal("InStock property not found")
	}
	if stockProp.Value.Description != "Whether product is in stock" {
		t.Errorf("Expected description 'Whether product is in stock', got '%s'", stockProp.Value.Description)
	}
	if stockProp.Value.Example != true {
		t.Errorf("Expected example true, got %v", stockProp.Value.Example)
	}

	// Check Count field (int example)
	countProp := schema.Properties["count"]
	if countProp == nil || countProp.Value == nil {
		t.Fatal("Count property not found")
	}
	if countProp.Value.Description != "Product count" {
		t.Errorf("Expected description 'Product count', got '%s'", countProp.Value.Description)
	}
	if countProp.Value.Example != int64(100) {
		t.Errorf("Expected example 100, got %v", countProp.Value.Example)
	}
}
