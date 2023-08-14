package testvalidate

import (
	"context"
	testvalidatev1 "github.com/grpc-ecosystem/go-grpc-middleware/v2/testing/testvalidate/v1"
)

type TestValidateService struct {
	testvalidatev1.UnimplementedTestValidateServiceServer
}

func (v *TestValidateService) Send(
	_ context.Context,
	_ *testvalidatev1.SendRequest,
) (*testvalidatev1.SendResponse, error) {
	return &testvalidatev1.SendResponse{}, nil
}

func (v *TestValidateService) SendStream(
	_ *testvalidatev1.SendStreamRequest,
	stream testvalidatev1.TestValidateService_SendStreamServer,
) error {
	for i := 0; i < 10; i++ {
		if err := stream.Send(&testvalidatev1.SendStreamResponse{}); err != nil {
			return err
		}
	}

	return nil
}

// Unary requests for unit testing
var (
	BadUnaryRequest = &testvalidatev1.SendRequest{
		Message: "%any",
	}

	GoodUnaryRequest = &testvalidatev1.SendRequest{
		Message: "good@example.com",
	}
)

// Stream requests for unit testing
var (
	BadStreamRequest = &testvalidatev1.SendStreamRequest{
		Message: "%any",
	}

	GoodStreamRequest = &testvalidatev1.SendStreamRequest{
		Message: "good@example.com",
	}
)
