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
	"github.com/cactus/go-statsd-client/statsd"
	logging "github.com/op/go-logging"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var logger = logging.MustGetLogger("comm/metrics")

// StatsdInterceptor provides GRPC stream and unary interceptors that can be used to
// generate metrics based on GRPC traffic
type StatsdInterceptor struct {
	statsdClient              statsd.Statter
	unaryClient, streamClient statsd.SubStatter
}

// NewStatsdInterceptor creates a StatsdInterceptor that will send metrics to the given endpoint
func NewStatsdInterceptor(endpoint string) (StatsdInterceptor, error) {
	var statsdClient, err = statsd.NewClient(endpoint, "interceptor")
	if err != nil {
		logger.Errorf("Failed to create statsd client: %s", err.Error())
	}

	return NewStatsdInterceptorWithStatter(statsdClient), nil
}

// NewStatsdInterceptorWithStatter creates a StatsdInterceptor from a given statsd.Statter
func NewStatsdInterceptorWithStatter(statsdClient statsd.Statter) StatsdInterceptor {
	var unaryStatter, streamStatter statsd.SubStatter
	unaryStatter = statsdClient.NewSubStatter("unary")
	streamStatter = statsdClient.NewSubStatter("stream")

	return StatsdInterceptor{
		statsdClient: statsdClient,
		unaryClient:  unaryStatter,
		streamClient: streamStatter,
	}
}

// UnaryMetricsInterceptor intercepts Unaruy traffic and sends statsd metrics
func (interceptor *StatsdInterceptor) UnaryMetricsInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	// Print out the grpc traffic info.  Anything will do
	logger.Debugf("UnaryMetricsInterceptor intercepted GRPC traffic")
	interceptor.unaryClient.Inc("messages", 1, 1.0)

	// Call handler to complete the RPC reqest, the same as next() in express
	return handler(ctx, req)
}

// StreamMetricsInterceptor intercepts stream traffic and sends statsd metrics.
func (interceptor *StatsdInterceptor) StreamMetricsInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	// Print out information about the steam
	logger.Debugf("StreamMetricsInterceptor intercepted GRPC traffic")
	interceptor.streamClient.Inc("messages", 1, 1.0)

	// Call handler to pass the RPC request along
	return handler(srv, ss)
}
