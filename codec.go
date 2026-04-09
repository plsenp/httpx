package httpx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	"github.com/go-playground/form/v4"
	"github.com/plsenp/httpx/encoding"
	"github.com/plsenp/httpx/errors"

	_ "github.com/plsenp/httpx/encoding/form"
	_ "github.com/plsenp/httpx/encoding/json"
	_ "github.com/plsenp/httpx/encoding/proto"
	_ "github.com/plsenp/httpx/encoding/xml"
	_ "github.com/plsenp/httpx/encoding/yaml"
)

var (
	queryDecoder  = form.NewDecoder()
	pathDecoder   = form.NewDecoder()
	headerDecoder = form.NewDecoder()
)

func init() {
	queryDecoder.SetTagName("query")
	pathDecoder.SetTagName("path")
	headerDecoder.SetTagName("header")
}

func defaultErrorHandler(w http.ResponseWriter, r *http.Request, err error) {
	w.Header().Set("Content-Type", "application/json")
	switch v := err.(type) {
	case *errors.HTTPError:
		w.WriteHeader(v.Code)
		_ = json.NewEncoder(w).Encode(v)
	default:
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(&errors.HTTPError{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
	}
}

func defaultQueryBind(r *http.Request, v any) error {
	if len(r.URL.Query()) == 0 {
		return nil
	}
	return queryDecoder.Decode(v, r.URL.Query())
}

func defaultPathBind(r *http.Request, v any) error {
	vs := make(url.Values, 0)
	for _, v := range strings.Split(r.Pattern, "/") {
		if strings.HasPrefix(v, "{") && strings.HasSuffix(v, "}") {
			v = v[1 : len(v)-1]
			v = strings.TrimSuffix(v, "...")

			vs.Add(v, r.PathValue(v))
		}
	}

	return pathDecoder.Decode(v, vs)
}

func defaultHeaderBind(r *http.Request, v any) error {
	return headerDecoder.Decode(v, url.Values(r.Header))
}

func defaultRenderFunc(w http.ResponseWriter, r *http.Request, status int, v any) error {
	if v == nil {
		w.WriteHeader(status)
		return nil
	}
	codec, _ := codecForRequest(r, "Accept")
	data, err := codec.Marshal(v)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/"+codec.Name())
	w.WriteHeader(status)
	_, err = w.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func codecForRequest(r *http.Request, name string) (encoding.Codec, bool) {
	for _, accept := range r.Header[name] {
		codec := encoding.GetCodec(ContentSubtype(accept))
		if codec != nil {
			return codec, true
		}
	}
	return encoding.GetCodec("json"), false
}

func ContentSubtype(contentType string) string {
	left := strings.Index(contentType, "/")
	if left == -1 {
		return ""
	}
	right := strings.Index(contentType, ";")
	if right == -1 {
		right = len(contentType)
	}
	if right < left {
		return ""
	}
	return contentType[left+1 : right]
}

func defaultBodyBind(r *http.Request, v any) error {
	codec, ok := codecForRequest(r, "Content-Type")
	if !ok {
		return nil
	}
	data, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	if len(data) == 0 {
		return nil
	}
	_ = r.Body.Close()
	r.Body = io.NopCloser(bytes.NewBuffer(data))
	return codec.Unmarshal(data, v)
}

// DefaultMaxMemory is the default max memory for multipart form.
const DefaultMaxMemory = 32 << 20 // 32MB

// bindFiles binds uploaded files to struct fields with "file" tag.
func bindFiles(v any, files map[string][]*multipart.FileHeader) error {
	if len(files) == 0 {
		return nil
	}

	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("target must be a non-nil pointer")
	}

	rv = rv.Elem()
	rt := rv.Type()

	if rt.Kind() != reflect.Struct {
		return fmt.Errorf("target must be a pointer to struct")
	}

	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		fieldValue := rv.Field(i)

		// Check for file tag
		tag := field.Tag.Get("file")
		if tag == "" || tag == "-" {
			continue
		}

		// Get uploaded files for this field
		uploadedFiles, ok := files[tag]
		if !ok || len(uploadedFiles) == 0 {
			continue
		}

		// Handle single file (*multipart.FileHeader)
		if field.Type == reflect.TypeOf(&multipart.FileHeader{}) {
			fieldValue.Set(reflect.ValueOf(uploadedFiles[0]))
		}

		// Handle multiple files ([]*multipart.FileHeader)
		if field.Type == reflect.TypeOf([]*multipart.FileHeader{}) {
			fieldValue.Set(reflect.ValueOf(uploadedFiles))
		}
	}

	return nil
}
