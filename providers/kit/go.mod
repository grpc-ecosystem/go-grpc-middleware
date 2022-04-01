module github.com/grpc-ecosystem/go-grpc-middleware/providers/kit/v2

go 1.14

require (
	github.com/go-kit/log v0.2.0
	github.com/grpc-ecosystem/go-grpc-middleware/v2 v2.0.0-rc.2.0.20201002093600-73cf2ae9d891
	golang.org/x/net v0.0.0-20220401154927-543a649e0bdd // indirect
	golang.org/x/sys v0.0.0-20220330033206-e17cdc41300f // indirect
	google.golang.org/genproto v0.0.0-20220401170504-314d38edb7de // indirect
	google.golang.org/grpc v1.45.0
)

replace github.com/grpc-ecosystem/go-grpc-middleware/v2 => ../..
