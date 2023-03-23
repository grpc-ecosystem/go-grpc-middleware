# Example Go app instrumented with go-grpc-middleware

This directory has server and client application communicating using [testpb.PingService](../testing/testpb/v1/test.proto) gRPC service.
Both are instrumented with example interceptors for auth, observability correlation, timeouts and more.

Feel free to copy and play with it.

## Usage

1. Run server in one terminal:

    ```
    go run server/main.go
    ```

2. Run client in second terminal:
    
    ```
    go run client/client.go
    ```
   
3. You should see logs and tracing in the output of both terminals thanks to logging and otlpgrpc interceptors. To check metrics instrumented with prometheus interceptor you can curl OpenMetrics (so exemplars are included):

    For server metrics:
    ```
    curl http://localhost:8081/metrics -H 'Accept: application/openmetrics-text; version=0.0.1'
    ```
   
    For client metrics:
    ```
    curl http://localhost:8082/metrics -H 'Accept: application/openmetrics-text; version=0.0.1'
    ```
