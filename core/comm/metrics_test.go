/*
Copyright IBM Corp. 2016 All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

                 http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package comm

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// MockContext is an argument to be used to test UnaryInterceptor
type MockContext struct{}

func (tc *MockContext) Deadline() (deadline time.Time, ok bool) {
	return time.Now(), true
}

func (tc *MockContext) Done() <-chan struct{} {
	return nil
}

func (tc *MockContext) Err() error {
	return errors.New("Dummy context error")
}

func (tc *MockContext) Value(key interface{}) interface{} {
	return nil
}

// MockUnaryHandlerResult lets MockUnaryHandler return all its arguments will still
// implementing the UnaryHandler interface
type MockUnaryHandlerResult struct {
	ctx context.Context
	req interface{}
}

// MockUnaryHandler gives us back everything that the UnaryInterceptor passed to it
func MockUnaryHandler(ctx context.Context, req interface{}) (interface{}, error) {
	hackyReturn := MockUnaryHandlerResult{ctx: ctx, req: req}
	return hackyReturn, nil
}

// MockRequest is something we can pass into interceptors to test them
type MockRequest struct {
	value string
}

func MockStreamHandler(srv interface{}, stream grpc.ServerStream) error {

}

func TestUnaryInterceptorReturns(t *testing.T) {

	t.Parallel()
	var ctx = &MockContext{}
	var req = MockRequest{value: "test"}
	var info = grpc.UnaryServerInfo{Server: nil, FullMethod: "/TestService/TestUnaryMethod"}

	testResult, _ := UnaryMetricsInterceptor(ctx, req, &info, MockUnaryHandler)

	if testResultConverted, ok := testResult.(MockUnaryHandlerResult); ok {
		assert.Equal(t, ctx, testResultConverted.ctx)
		assert.Equal(t, req, testResultConverted.req)
		t.Log("UnaryInterceptor passes along its inputs, exactly as it should")
	} else {
		t.Fatal("Whatever came back from UnaryMetricsInterceptor was not what we passed in")
	}
}

func TestStreamInterceptorReturns(t *testing.T) {
	t.Parallel()

	var srv interface{}
	var ss grpc.ServerStream
	var info = grpc.StreamServerInfo{FullMethod: "/TestService/TestStreamMethod",
		IsClientStream: false, IsServerStream: false}

	err := StreamMetricsInterceptor(srv, ss, &info, MockStreamHandler)

}
