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

var logger = logging.MustGetLogger("orderer/metrics")

// UnaryMetricsInterceptor intercepts Unary traffic and sends statsd metrics.
func UnaryMetricsInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	// Print out the grpc traffic info.  Anything will do
	logger.Debug("UnaryMetricsInterceptor called")

	// Send a metric to statsd
	logger.Debug("Sending statsd metric")
	client, _ := statsd.NewClient("127.0.0.1:8125", "unary_interceptor")
	client.Inc("interceptions", 1, 1.0)

	// Call handler to complete the RPC reqest, the same as next() in express
	return handler(ctx, req)
}
