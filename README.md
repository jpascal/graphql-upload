# graphql-upload

[![GitHub Actions status](https://github.com/jpascal/graphql-upload/workflows/Test/badge.svg)](https://github.com/jpascal/graphql-upload/actions)

Middleware and an [`Upload` scalar](#class-graphqlupload) to add support for [GraphQL multipart requests](https://github.com/jaydenseric/graphql-multipart-request-spec) (file uploads via queries and mutations) to various golang GraphQL servers.

## Installation
```bash
go get github.com/jpascal/graphql-upload
```

## Usage

```go
server := &http.Server{
	Addr: "0.0.0.0:5000", 
	Handler: handler.New(func(request *handler.Request) interface{} {
		return graphql.Do(graphql.Params{
			RequestString:  request.Query,
			OperationName:  request.OperationName,
			VariableValues: request.Variables,
			Schema:         schema.New(),
			Context:        request.Context,
		})
	}, &handler.Config {MaxBodySize: 1024}),
}
    server.ListenAndServe()
```

## Contributing

1. Fork it ( https://github.com/jpascal/graphql-upload/fork )
2. Create your feature branch (`git checkout -b my-new-feature`)
3. Commit your changes (`git commit -am 'Add some feature'`)
4. Push to the branch (`git push origin my-new-feature`)
5. Create a new Pull Request
