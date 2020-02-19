package main

import (
	"fmt"
	"github.com/graphql-go/graphql"
	handler "github.com/jpascal/graphql-upload"
	"net/http"
	"os"
	"path"
)

var UploadType = graphql.NewScalar(graphql.ScalarConfig{
	Name:        "Upload",
	Description: "Scalar upload object",
})

type FileWrapper struct {
	File *os.File
	Name string
}

var File = graphql.NewObject(graphql.ObjectConfig{
	Name:        "File",
	Description: "File object",
	Fields: graphql.Fields{
		"name": &graphql.Field{
			Type: graphql.String,
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				file := params.Source.(*FileWrapper)
				name := path.Base(file.Name)
				return name, nil
			},
		},
		"url": &graphql.Field{
			Type: graphql.String,
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				file := params.Source.(*FileWrapper)
				name := path.Base(file.File.Name())
				return path.Join("/uploads/", name), nil
			},
		},
		"size": &graphql.Field{
			Type: graphql.Int,
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				file := params.Source.(*FileWrapper)
				if info, err := file.File.Stat(); err != nil {
					return nil, err
				} else {
					return info.Size(), nil
				}
			},
		},
	},
})

func main() {
	schema, err := graphql.NewSchema(
		graphql.SchemaConfig{
			Query: graphql.NewObject(
				graphql.ObjectConfig{
					Name: "Query",
					Fields: graphql.Fields{
						"file": &graphql.Field{
							Type: File,
							Args: graphql.FieldConfigArgument{
								"id": &graphql.ArgumentConfig{
									Type: graphql.NewNonNull(graphql.String),
								},
							},
							Resolve: func(params graphql.ResolveParams) (interface{}, error) {
								if fileId, ok := params.Args["id"].(string); ok {
									file, _ := os.Open(fileId)
									return &FileWrapper{File: file, Name: "some-file-name"}, nil
								} else {
									return nil, nil
								}
							},
						},
					},
				}),
			Mutation: graphql.NewObject(
				graphql.ObjectConfig{
					Name: "Mutation",
					Fields: graphql.Fields{
						"upload": &graphql.Field{
							Type: graphql.String,
							Args: graphql.FieldConfigArgument{
								"file": &graphql.ArgumentConfig{
									Type: UploadType,
								},
							},
							Resolve: func(params graphql.ResolveParams) (interface{}, error) {
								upload, uploadPresent := params.Args["file"].(handler.File)
								if uploadPresent {
									fmt.Print(upload.Filename)
									// Store file somewhere...
								}
								return "file-id", nil
							},
						},
					},
				}),
		})
	if err != nil {
		panic(err)
	}

	server := &http.Server{Addr: "0.0.0.0:5000", Handler: handler.New(func(request *handler.Request) interface{} {
		return graphql.Do(graphql.Params{
			RequestString:  request.Query,
			OperationName:  request.OperationName,
			VariableValues: request.Variables,
			Schema:         schema,
			Context:        request.Context,
		})
	}, &handler.Config{
		MaxBodySize: 1024,
	}),
	}
	server.ListenAndServe()
}
