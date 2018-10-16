package handler

import (
	"bytes"
	"encoding/json"
	"github.com/graphql-go/graphql"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

var UploadType = graphql.NewScalar(graphql.ScalarConfig{
	Name:        "Upload",
	Description: "Scalar upload object",
})

var FileObject = graphql.NewObject(graphql.ObjectConfig{
	Name:        "File",
	Description: "File object",
	Fields: graphql.Fields{
		"fileName": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				file := p.Source.(File)
				return file.Filename, nil
			},
		},
		"size": &graphql.Field{
			Type: graphql.Int,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				file := p.Source.(File)
				return file.Size, nil
			},
		},
	},
})

func Schema() graphql.Schema {
	if schema, err := graphql.NewSchema(
		graphql.SchemaConfig{
			Query: graphql.NewObject(
				graphql.ObjectConfig{
					Name: "Query",
					Fields: graphql.Fields{
						"version": &graphql.Field{
							Type: graphql.String,
							Resolve: func(params graphql.ResolveParams) (interface{}, error) {
								return "v0.0.0", nil
							},
						},
					},
				}),
			Mutation: graphql.NewObject(
				graphql.ObjectConfig{
					Name: "Mutation",
					Fields: graphql.Fields{
						"version": &graphql.Field{
							Type: graphql.String,
							Resolve: func(params graphql.ResolveParams) (interface{}, error) {
								return "sdd", nil
							},
						},
						"singleUpload": &graphql.Field{
							Args: graphql.FieldConfigArgument{
								"file": &graphql.ArgumentConfig{
									Type: UploadType,
								},
							},
							Type: FileObject,
							Resolve: func(params graphql.ResolveParams) (interface{}, error) {
								return params.Args["file"], nil
							},
						},
						"multipleUpload": &graphql.Field{
							Args: graphql.FieldConfigArgument{
								"files": &graphql.ArgumentConfig{
									Type: graphql.NewList(UploadType),
								},
							},
							Type: graphql.NewList(FileObject),
							Resolve: func(params graphql.ResolveParams) (interface{}, error) {
								return params.Args["files"], nil
							},
						},
					},
				}),
		}); err != nil {
		panic(err)
	} else {
		return schema
	}
}

func TestHandler(t *testing.T) {
	handler := New(func(request *Request) interface{} {
		return graphql.Do(graphql.Params{
			RequestString:  request.Query,
			OperationName:  request.OperationName,
			VariableValues: request.Variables,
			Schema:         Schema(),
			Context:        request.Context,
		})
	});
	t.Run("GET Regular", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/", nil)
		if err != nil {
			t.Fatal(err)
		}

		req.URL.RawQuery = url.Values{
			"operationName": {"version"},
			"query":         {"query version { version }"},
			"variables":     {"{}"},
		}.Encode()

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", rr.Code, http.StatusOK)
		}

		expected := `{"data":{"version":"v0.0.0"}}` + "\n"
		got := string(rr.Body.Bytes())
		if got != expected {
			t.Errorf("handler returned unexpected body: got %v want %v",
				got, expected)
		}
	})
	t.Run("POST Regular", func(t *testing.T) {
		values := map[string]interface{}{
			"operationName": "version",
			"query":         "query version { version }",
			"variables":     map[string]interface{}{},
		}
		body, _ := json.Marshal(values)

		req, err := http.NewRequest("POST", "/", bytes.NewBuffer(body))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", rr.Code, http.StatusOK)
		}

		expected := `{"data":{"version":"v0.0.0"}}` + "\n"
		got := string(rr.Body.Bytes())
		if got != expected {
			t.Errorf("handler returned unexpected body: got %v want %v",
				got, expected)
		}

	})
	t.Run("POST Single File", func(t *testing.T) {
		bodyBuf := &bytes.Buffer{}
		bodyWriter := multipart.NewWriter(bodyBuf)
		bodyWriter.WriteField("operations",`{ "query": "mutation ($file: Upload!) { singleUpload(file: $file) { fileName, size } }", "variables": { "file": null } }`)
		bodyWriter.WriteField("map",`{ "0": ["variables.file"] }`)
		w, _ := bodyWriter.CreateFormFile("0", "a.txt")
		w.Write([]byte("test"))
		bodyWriter.Close()

		req, err := http.NewRequest("POST", "/", bodyBuf)
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Content-Type", bodyWriter.FormDataContentType())
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", rr.Code, http.StatusOK)
		}

		expected := `{"data":{"singleUpload":{"fileName":"a.txt","size":4}}}` + "\n"
		got := string(rr.Body.Bytes())
		if got != expected {
			t.Errorf("handler returned unexpected body: got %v want %v",
				got, expected)
		}
	})
	t.Run("POST File List", func(t *testing.T) {
		bodyBuf := &bytes.Buffer{}
		bodyWriter := multipart.NewWriter(bodyBuf)
		bodyWriter.WriteField("operations",`{ "query": "mutation($files: [Upload!]!) { multipleUpload(files: $files) { fileName } }", "variables": { "files": [null, null] } }`)
		bodyWriter.WriteField("map",`{ "0": ["variables.files.0"], "1": ["variables.files.1"] }`)
		w0, _ := bodyWriter.CreateFormFile("0", "a.txt")
		w0.Write([]byte("test"))
		w1, _ := bodyWriter.CreateFormFile("1", "b.txt")
		w1.Write([]byte("test"))
		bodyWriter.Close()

		req, err := http.NewRequest("POST", "/", bodyBuf)
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Content-Type", bodyWriter.FormDataContentType())
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", rr.Code, http.StatusOK)
		}

		expected := `{"data":{"multipleUpload":[{"fileName":"a.txt"},{"fileName":"b.txt"}]}}` + "\n"
		got := string(rr.Body.Bytes())
		if got != expected {
			t.Errorf("handler returned unexpected body: got %v want %v",
				got, expected)
		}
	})
	t.Run("POST Batch with files", func(t *testing.T) {
		bodyBuf := &bytes.Buffer{}
		bodyWriter := multipart.NewWriter(bodyBuf)
		bodyWriter.WriteField("operations",`[{ "query": "mutation ($file: Upload!) { singleUpload(file: $file) { fileName } }", "variables": { "file": null } }, { "query": "mutation($files: [Upload!]!) { multipleUpload(files: $files) { fileName } }", "variables": { "files": [null, null] } }]`)
		bodyWriter.WriteField("map",`{ "0": ["0.variables.file"], "1": ["1.variables.files.0"], "2": ["1.variables.files.1"] }`)
		w0, _ := bodyWriter.CreateFormFile("0", "a.txt")
		w0.Write([]byte("test"))
		w1, _ := bodyWriter.CreateFormFile("1", "b.txt")
		w1.Write([]byte("test"))
		w2, _ := bodyWriter.CreateFormFile("2", "c.txt")
		w2.Write([]byte("test"))
		bodyWriter.Close()

		req, err := http.NewRequest("POST", "/", bodyBuf)
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Content-Type", bodyWriter.FormDataContentType())
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", rr.Code, http.StatusOK)
		}

		expected := `[{"data":{"singleUpload":{"fileName":"a.txt"}}},{"data":{"multipleUpload":[{"fileName":"b.txt"},{"fileName":"c.txt"}]}}]` + "\n"
		got := string(rr.Body.Bytes())
		if got != expected {
			t.Errorf("handler returned unexpected body: got %v want %v",
				got, expected)
		}
	})
}
