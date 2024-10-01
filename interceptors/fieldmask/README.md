# fieldmask-middlware
Go grpc middleware for field mask
This package provides an interceptor that will `filter` server response by client intention.

# How to use
`Pre-condition`

You have to add `field_mask` into proto's request like
``` Protobuf
message Request{
  // main fields
  google.protobuf.FieldMask field_mask = 100;
}

```
1. install via go get 

 `go get github.com/grpc-ecosystem/go-grpc-middleware/v2`
`

2. Import and inject into grpc interceptor
The code in your application should be like that:
``` Go
import(
        // ...
        "google.golang.org/grpc"
    	fieldmaskpkg "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/fieldmask"


)
// ...

func main(){
    var unaryOpts []grpc.UnaryServerInterceptor{
		fieldmaskpkg.UnaryServerInterceptor(fieldmaskpkg.DefaultFilterFunc),
    }
    // Should append others interceptors
}
```
3. Client usage
   	
	 `Protobuf definition`
	 ```Protobuf
	   message GetProductRequest{
	     string id = 1;
	     google.protobuf.FieldMask field_mask = 2;
	   }
	   message Response{
	     message Result{
	        repeated Product products = 1;
	     }
	     Result result = 1;
	   }
	   message Product{
	      string id = 1;
	      string name = 2;
	      string img = 3;
	      decimal price = 4;
	   }
	   
 	```

	Client interaction

	 ``` Go
         func GetProduct(id string) *Product{
		  request := pb.GetProductRequest{
			Id: "axaxaxx",
			FieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{
					"result.products.id", "result.products.name"
				},
			}
		   }
           // ...
         }
	 ```
