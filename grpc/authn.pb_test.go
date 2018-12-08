package grpc

import (
	"context"
	"log"
	"net"
	"runtime"
	"testing"

	"google.golang.org/grpc"
)

func TestServer(t *testing.T) {

	addr := "localhost:0"
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to initializa TCP listen: %v", err)
	}
	defer lis.Close()
	srv := NewServer(lis, nil)

	ctx := context.Background()
	go srv.RunGRPC(ctx)

	runtime.Gosched()

	cc, err := grpc.Dial(lis.Addr().String(), grpc.WithInsecure())
	if err != nil {
		t.Logf("failed to dial: %s", err)
		t.FailNow()
	}

	client := NewPublicAuthNClient(cc)

	signupForm := &SignupRequest{
		Username: "cool-username",
		Password: "supersecurepassword",
	}

	res, signErr := client.Signup(ctx, signupForm)
	if signErr != nil {
		t.Logf("failed to call Signup: %s", signErr)
		t.FailNow()
	}
	t.Logf("%s", res)

}
