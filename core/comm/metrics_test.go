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

type testContext struct{}

func (tc *testContext) Deadline() (deadline time.Time, ok bool) {
	return time.Now(), true
}

func (tc *testContext) Done() <-chan struct{} {
	return nil
}

func (tc *testContext) Err() error {
	return errors.New("Dummy context error")
}

func (tc *testContext) Value(key interface{}) interface{} {
	return nil
}

func interceptorTestHandler(ctx context.Context, req interface{}) (interface{}, error) {
	hackyReturn := testStruct{ctx: ctx, req: req}
	return hackyReturn, nil
}

type testReq struct {
	value string
}

type testStruct struct {
	ctx context.Context
	req interface{}
}

func TestUnaryInterceptorReturns(t *testing.T) {

	var ctx = &testContext{}
	var req = testReq{value: "test"}
	var info = grpc.UnaryServerInfo{Server: nil, FullMethod: "/TestMethod"}

	testResult, _ := UnaryMetricsInterceptor(ctx, req, &info, interceptorTestHandler)

	if testResultConverted, ok := testResult.(testStruct); ok {
		assert.Equal(t, ctx, testResultConverted.ctx)
		assert.Equal(t, req, testResultConverted.req)
		t.Log("UnaryInterceptor passes along its inputs, exactly as it should")
	} else {
		t.Fatal("Whatever came back from UnaryMetricsInterceptor was not what we passed in")
	}
}
