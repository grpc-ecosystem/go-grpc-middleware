package metrics

type grpcType string

const (
	Unary        grpcType = "unary"
	ClientStream grpcType = "client_stream"
	ServerStream grpcType = "server_stream"
	BidiStream   grpcType = "bidi_stream"
)

type Kind string

const (
	KindClient Kind = "client"
	KindServer Kind = "server"
)

type RPCMethod string

const (
	Send    RPCMethod = "send"
	Receive RPCMethod = "recv"
)
