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

package comm_test

import (
	"errors"
	"net"
	"testing"
	"time"

	"github.com/hyperledger/fabric/core/comm"
	"github.com/stretchr/testify/assert"

	"github.com/cactus/go-statsd-client/statsd"
	testpb "github.com/hyperledger/fabric/core/comm/testdata/grpc"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// MockContext is an argument to be used to test UnaryInterceptor
type MockContext struct{}

func (tc *MockContext) Deadline() (deadline time.Time, ok bool) { return time.Now(), true }

func (tc *MockContext) Done() <-chan struct{} { return nil }

func (tc *MockContext) Err() error { return errors.New("Dummy context error") }

func (tc *MockContext) Value(key interface{}) interface{} { return nil }

// MockStat represents a statsd metric that has been passed to our MockStatSender
type MockStat struct {
	Name       string
	Prefix     string
	ValueNum   int64
	ValueTime  time.Duration
	ValueStr   string
	SampleRate float32
	Function   string
}

// MockStatSender lets us capture the metrics that a statsd generating client may send out. The
// channel object is used to send out the metrics that MockStatSender receives, so that they can
// be validated elsewhere.
type MockStatSender struct {
	Prefix    string
	responder chan<- MockStat
}

func NewMockStatSender(prefix string, responder chan<- MockStat) MockStatSender {
	return MockStatSender{Prefix: prefix, responder: responder}
}
func (mstats MockStatSender) Inc(statPrefix string, statValue int64, sampleRate float32) error {
	go func() {
		mstats.responder <- MockStat{Prefix: statPrefix, ValueNum: statValue, SampleRate: sampleRate, Function: "Inc"}
	}()
	return nil
}
func (mstats MockStatSender) Dec(statPrefix string, statValue int64, sampleRate float32) error {
	go func() {
		mstats.responder <- MockStat{Prefix: statPrefix, ValueNum: statValue, SampleRate: sampleRate, Function: "Dec"}
	}()
	return nil
}
func (mstats MockStatSender) Gauge(statPrefix string, statValue int64, sampleRate float32) error {
	go func() {
		mstats.responder <- MockStat{Prefix: statPrefix, ValueNum: statValue, SampleRate: sampleRate, Function: "Gauge"}
	}()
	return nil
}
func (mstats MockStatSender) GaugeDelta(statPrefix string, statValue int64, sampleRate float32) error {
	go func() {
		mstats.responder <- MockStat{Prefix: statPrefix, ValueNum: statValue, SampleRate: sampleRate, Function: "GaugeDelta"}
	}()
	return nil
}
func (mstats MockStatSender) Timing(statPrefix string, statValue int64, sampleRate float32) error {
	go func() {
		mstats.responder <- MockStat{Prefix: statPrefix, ValueNum: statValue, SampleRate: sampleRate, Function: "Timing"}
	}()
	return nil
}
func (mstats MockStatSender) SetInt(statPrefix string, statValue int64, sampleRate float32) error {
	go func() {
		mstats.responder <- MockStat{Prefix: statPrefix, ValueNum: statValue, SampleRate: sampleRate, Function: "SetInt"}
	}()
	return nil
}
func (mstats MockStatSender) Set(statPrefix string, statValue string, sampleRate float32) error {
	go func() {
		mstats.responder <- MockStat{Prefix: statPrefix, ValueStr: statValue, SampleRate: sampleRate, Function: "Set"}
	}()
	return nil
}
func (mstats MockStatSender) Raw(statPrefix string, statValue string, sampleRate float32) error {
	go func() {
		mstats.responder <- MockStat{Prefix: statPrefix, ValueStr: statValue, SampleRate: sampleRate, Function: "Raw"}
	}()
	return nil
}
func (mstats MockStatSender) TimingDuration(statPrefix string, statValue time.Duration, sampleRate float32) error {
	go func() {
		mstats.responder <- MockStat{Prefix: statPrefix, ValueTime: statValue, SampleRate: sampleRate, Function: "TimingDuration"}
	}()
	return nil
}
func (mstats MockStatSender) SetSamplerFunc(statsd.SamplerFunc) {
	return
}
func (mstats MockStatSender) NewSubStatter(prefix string) statsd.SubStatter {
	return NewMockStatSender(prefix, mstats.responder)
}
func (mstats MockStatSender) SetPrefix(prefix string) {
	mstats.Prefix = prefix
}
func (mstats MockStatSender) Close() error {
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

// MockStreamHandlerResult lets MockStreamHandler pass along all it's arguments for
// validation in our unit tests
type MockStreamHandlerResult struct {
	srv interface{}
	ss  grpc.ServerStream
}

func CreateMockStreamHandler(response chan<- MockStreamHandlerResult) func(interface{}, grpc.ServerStream) error {
	// The handler should just package up everything it was sent and send it
	// through the channel to be validated in a unit test.
	return func(srv interface{}, stream grpc.ServerStream) error {
		var result = MockStreamHandlerResult{srv: srv, ss: stream}
		go func() {
			response <- result
		}()
		return nil
	}
}

func MockStreamHandler(srv interface{}, stream grpc.ServerStream) error {
	return nil
}

func TestUnaryInterceptorTransparency(t *testing.T) {
	t.Parallel()
	var ctx = &MockContext{}
	var req = MockRequest{value: "test"}
	var info = grpc.UnaryServerInfo{Server: nil, FullMethod: "/TestService/TestUnaryMethod"}

	// No need to actually send metrics for this test
	var sender, _ = statsd.NewNoopClient()
	var interceptor = comm.NewStatsdInterceptorWithStatter(sender)
	testResult, _ := interceptor.UnaryMetricsInterceptor(ctx, req, &info, MockUnaryHandler)

	if testResultConverted, ok := testResult.(MockUnaryHandlerResult); ok {
		assert.Equal(t, ctx, testResultConverted.ctx)
		assert.Equal(t, req, testResultConverted.req)
		t.Log("UnaryMetricsInterceptor passes along all of its inputs to its handler, exactly as it should")
	} else {
		t.Fatal("UnaryMetricsInterceptor did not pass its inputs to its handler properly")
	}
}

func TestUnaryInterceptorSendsMetrics(t *testing.T) {
	t.Parallel()
	var ctx = &MockContext{}
	var req = MockRequest{value: "test"}
	var info = grpc.UnaryServerInfo{Server: nil, FullMethod: "/TestService/TestUnaryMethod"}

	// Use a MockStatSender so we can capture the metrics that get sent
	response := make(chan MockStat, 1)
	var sender = NewMockStatSender("mockPrefix", response)
	var interceptor = comm.NewStatsdInterceptorWithStatter(sender)
	interceptor.UnaryMetricsInterceptor(ctx, req, &info, MockUnaryHandler)

	// Use a timeout to complete the test if the metrics are never sent
	timeout := make(chan bool, 1)
	go func() {
		time.Sleep(500 * time.Millisecond)
		timeout <- true
	}()

	select {
	case mockStat := <-response:
		assert.Equal(t, mockStat.Function, "Inc")
		assert.Equal(t, mockStat.Prefix, "messages")
		assert.Equal(t, mockStat.ValueNum, int64(1))
		t.Log("Received a metric from the UnaryMetricsInterceptor")
	case <-timeout:
		t.Fatal("Never received a metric from the UnaryMetricsInterceptor")
	}
}

func TestStreamInterceptorTransparency(t *testing.T) {
	t.Parallel()
	var srv interface{}
	var ss grpc.ServerStream
	var info = grpc.StreamServerInfo{FullMethod: "/TestService/TestStreamMethod",
		IsClientStream: false, IsServerStream: false}

	// No need to actually send metrics for this test
	var sender, _ = statsd.NewNoopClient()
	var interceptor = comm.NewStatsdInterceptorWithStatter(sender)
	response := make(chan MockStreamHandlerResult, 1)
	interceptor.StreamMetricsInterceptor(srv, ss, &info, CreateMockStreamHandler(response))

	// Use a timeout to complete the test if the metrics are never sent
	timeout := make(chan bool, 1)
	go func() {
		time.Sleep(500 * time.Millisecond)
		timeout <- true
	}()

	select {
	case mockResult := <-response:
		assert.Equal(t, mockResult.srv, srv)
		assert.Equal(t, mockResult.ss, ss)
		t.Log("StreamMetricsInterceptor passes along all of its inputs to its handler, exactly as it should")
	case <-timeout:
		t.Fatal("StreamMetricsInterceptor did not pass its inputs to its handler properly")
	}
}

func TestStreamInterceptorSendsMetrics(t *testing.T) {
	t.Parallel()
	var srv interface{}
	var ss grpc.ServerStream
	var info = grpc.StreamServerInfo{FullMethod: "/TestService/TestStreamMethod",
		IsClientStream: false, IsServerStream: false}

	// No need to actually send metrics for this test
	response := make(chan MockStat, 1)
	var sender = NewMockStatSender("mockPrefix", response)
	var interceptor = comm.NewStatsdInterceptorWithStatter(sender)
	interceptor.StreamMetricsInterceptor(srv, ss, &info, MockStreamHandler)

	// Use a timeout to complete the test if the metrics are never sent
	timeout := make(chan bool, 1)
	go func() {
		time.Sleep(500 * time.Millisecond)
		timeout <- true
	}()

	select {
	case mockStat := <-response:
		assert.Equal(t, mockStat.Function, "Inc")
		assert.Equal(t, mockStat.Prefix, "messages")
		assert.Equal(t, mockStat.ValueNum, int64(1))
		t.Log("Received a metric from the StreamMetricsInterceptor")
	case <-timeout:
		t.Fatal("Never received a metric from the StreamMetricsInterceptor")
	}
}

func TestNewGRPCServerWithInterceptor(t *testing.T) {
	t.Parallel()
	testAddress := "localhost:9053"
	testStatsdAddress := "localhost:9125"
	srv, err := comm.NewGRPCServer(testAddress,
		comm.SecureServerConfig{UseTLS: false, SendMetrics: true,
			StatsdAddress: testStatsdAddress})
	//check for error
	if err != nil {
		t.Fatalf("Failed to return new GRPC server with metrics: %v", err)
	}

	//register the GRPC test server
	testpb.RegisterTestServiceServer(srv.Server(), &testServiceServer{})

	//start the server
	go srv.Start()

	defer srv.Stop()
	//should not be needed
	time.Sleep(10 * time.Millisecond)

	// Setup a goroutine to listen for the statsd request from the NewGRPCServer
	//response := make(chan string, 1)
	listener, err := net.ListenPacket("udp", testStatsdAddress)
	if err != nil {
		t.Fatalf("Failed to create a UDP listener to listen for statsd metrics: %s", err.Error())
	}
	listener.SetDeadline(time.Now().Add(100 * time.Millisecond))
	listener.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	listener.SetWriteDeadline(time.Now().Add(100 * time.Millisecond))
	l := listener.(*net.UDPConn)

	defer l.Close()

	go func() {
		// Create a buffer and read any metrics data sent by the interceptor
		data := make([]byte, 128)
		n, _, err := l.ReadFrom(data)

		if err != nil {
			t.Fatalf("Failed to read metrics data: %s", err.Error())
		} else {
			metric := string(data[:n])

			// The interceptor should just count the number of messages for now
			assert.Equal(t, "interceptor.unary.messages:1|c", metric)
		}
	}()

	//GRPC client options
	var dialOptions []grpc.DialOption
	dialOptions = append(dialOptions, grpc.WithInsecure())

	//invoke the EmptyCall service
	_, err = invokeEmptyCall(testAddress, dialOptions)

	if err != nil {
		t.Fatalf("GRPC client failed to invoke the EmptyCall service on %s: %v",
			testAddress, err)
	} else {
		t.Log("GRPC client successfully invoked the EmptyCall service: " + testAddress)
	}
}

func newUDPListener(addr string) (*net.UDPConn, error) {
	l, err := net.ListenPacket("udp", addr)
	if err != nil {
		return nil, err
	}
	l.SetDeadline(time.Now().Add(100 * time.Millisecond))
	l.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	l.SetWriteDeadline(time.Now().Add(100 * time.Millisecond))
	return l.(*net.UDPConn), nil
}
