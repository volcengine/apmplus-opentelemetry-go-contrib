// Copyright 2026 Beijing Volcano Engine Technology Co., Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package filters

import (
	"context"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/bytedance/mockey"
	. "github.com/smartystreets/goconvey/convey"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv7 "go.opentelemetry.io/otel/semconv/v1.7.0"
	apitrace "go.opentelemetry.io/otel/trace"
	"trpc.group/trpc-go/trpc-go/codec"
	"trpc.group/trpc-go/trpc-go/errs"
	trpcpb "trpc.group/trpc/trpc-protocol/pb/go/trpc"
)

type mockExporter struct {
	metrics []metricdata.ScopeMetrics
	spans   []trace.ReadOnlySpan
	mu      sync.RWMutex
}

func newMockExporter() *mockExporter {
	return &mockExporter{
		metrics: make([]metricdata.ScopeMetrics, 0),
		spans:   make([]trace.ReadOnlySpan, 0),
	}
}

func (m *mockExporter) Temporality(kind metric.InstrumentKind) metricdata.Temporality {
	return metricdata.DeltaTemporality
}

func (m *mockExporter) Aggregation(kind metric.InstrumentKind) metric.Aggregation {
	return nil
}

func (m *mockExporter) Export(ctx context.Context, metrics *metricdata.ResourceMetrics) error {
	if len(metrics.ScopeMetrics) == 0 {
		return nil
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.metrics = append(m.metrics, metrics.ScopeMetrics...)
	return nil
}

func (m *mockExporter) ForceFlush(ctx context.Context) error {
	return nil
}

func (m *mockExporter) ExportSpans(ctx context.Context, spans []trace.ReadOnlySpan) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.spans = append(m.spans, spans...)
	return nil
}

func (m *mockExporter) Shutdown(ctx context.Context) error {
	return nil
}

func (m *mockExporter) GetMetrics() []metricdata.ScopeMetrics {
	for {
		m.mu.Lock()
		if len(m.metrics) > 0 {
			metrics := m.metrics
			m.metrics = make([]metricdata.ScopeMetrics, 0)
			m.mu.Unlock()
			return metrics
		}
		m.mu.Unlock()
		time.Sleep(time.Second)
	}
}

func (m *mockExporter) GetSpans() []trace.ReadOnlySpan {
	m.mu.Lock()
	defer m.mu.Unlock()
	spans := m.spans
	m.spans = make([]trace.ReadOnlySpan, 0)
	return spans
}

func initOTELSDK(exporter *mockExporter) {
	otel.SetTracerProvider(trace.NewTracerProvider(trace.WithSpanProcessor(trace.NewSimpleSpanProcessor(exporter))))
	otel.SetMeterProvider(metric.NewMeterProvider(metric.WithReader(metric.NewPeriodicReader(exporter, metric.WithInterval(time.Second)))))
}

var (
	exporter = newMockExporter()
)

func init() {
	initOTELSDK(exporter)
}

func TestServerFilter(t *testing.T) {
	mockey.PatchConvey("ServerFilter", t, func() {
		Convey("Return Error", func() {
			ctx := context.Background()
			ctx, msg := codec.WithCloneMessage(ctx)
			msg.WithServerRPCName("/trpc.helloworld.Greeter/Hello")
			msg.WithCallerServiceName("test")
			msg.WithCallerService("Greeter")
			msg.WithCallerMethod("Hello")

			_, err := ServerFilter(ctx, nil, func(ctx context.Context, req interface{}) (interface{}, error) {
				return nil, errs.New(trpcpb.TrpcRetCode_TRPC_SERVER_AUTH_ERR, "auth failed")
			})
			So(err, ShouldNotBeNil)
			spans := exporter.GetSpans()
			So(spans, ShouldHaveLength, 1)
			So(spans[0].Name(), ShouldEqual, "/trpc.helloworld.Greeter/Hello")
			So(spans[0].Status().Code, ShouldEqual, codes.Error)
			So(spans[0].SpanKind(), ShouldEqual, apitrace.SpanKindServer)
			So(spans[0].SpanContext().HasTraceID(), ShouldBeTrue)
			So(len(spans[0].Attributes()) > 0, ShouldBeTrue)
		})

		Convey("Return Success", func() {
			ctx := context.Background()
			ctx, msg := codec.WithCloneMessage(ctx)
			msg.WithServerRPCName("/trpc.helloworld.Greeter/Hello")
			msg.WithCallerServiceName("test")
			msg.WithCallerService("Greeter")
			msg.WithCallerMethod("Hello")
			addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:8080")
			So(err, ShouldBeNil)
			msg.WithRemoteAddr(addr)
			_, err = ServerFilter(ctx, nil, func(ctx context.Context, req interface{}) (interface{}, error) {
				return nil, nil
			})
			So(err, ShouldBeNil)
			spans := exporter.GetSpans()
			So(spans, ShouldHaveLength, 1)
			So(spans[0].Name(), ShouldEqual, "/trpc.helloworld.Greeter/Hello")
			So(spans[0].Status().Code, ShouldEqual, codes.Ok)
			So(spans[0].SpanKind(), ShouldEqual, apitrace.SpanKindServer)
			So(spans[0].SpanContext().HasTraceID(), ShouldBeTrue)
			So(len(spans[0].Attributes()) > 0, ShouldBeTrue)
		})

		metrics := exporter.GetMetrics()
		So(len(metrics) > 0, ShouldBeTrue)
		So(len(metrics[0].Metrics) > 0, ShouldBeTrue)
		So(metrics[0].Metrics[0].Name, ShouldEqual, "rpc.server.duration")
		So(metrics[0].Metrics[0].Unit, ShouldEqual, "ms")
	})
}

func TestClientFilter(t *testing.T) {
	mockey.PatchConvey("ClientFilter", t, func() {
		Convey("Return Error", func() {
			ctx := context.Background()
			ctx, msg := codec.WithCloneMessage(ctx)
			msg.WithClientRPCName("/trpc.helloworld.Greeter/Hello")
			msg.WithCalleeServiceName("test")
			msg.WithCalleeService("Greeter")
			msg.WithCalleeMethod("Hello")
			addr, err := net.ResolveTCPAddr("tcp", "localhost:8080")
			So(err, ShouldBeNil)
			msg.WithRemoteAddr(addr)
			msg.WithSerializationType(codec.SerializationTypePB)
			err = ClientFilter(ctx, nil, nil, func(ctx context.Context, req, rsp interface{}) error {
				return errs.New(trpcpb.TrpcRetCode_TRPC_SERVER_AUTH_ERR, "auth failed")
			})
			So(err, ShouldNotBeNil)
			spans := exporter.GetSpans()
			So(spans, ShouldHaveLength, 1)
			So(spans[0].Name(), ShouldEqual, "/trpc.helloworld.Greeter/Hello")
			So(spans[0].Status().Code, ShouldEqual, codes.Error)
			So(spans[0].SpanKind(), ShouldEqual, apitrace.SpanKindClient)
			So(spans[0].SpanContext().HasTraceID(), ShouldBeTrue)
			So(len(spans[0].Attributes()) > 0, ShouldBeTrue)
		})

		Convey("Return Success", func() {
			ctx := context.Background()
			ctx, msg := codec.WithCloneMessage(ctx)
			msg.WithClientRPCName("/trpc.helloworld.Greeter/Hello")
			msg.WithCalleeServiceName("test")
			msg.WithCalleeService("Greeter")
			msg.WithCalleeMethod("Hello")
			msg.WithCommonMeta(codec.CommonMeta{
				"":                   "test",
				semconv7.DBSystemKey: "test",
			})
			err := ClientFilter(ctx, nil, nil, func(ctx context.Context, req, rsp interface{}) error {
				return nil
			})
			So(err, ShouldBeNil)
			spans := exporter.GetSpans()
			So(spans, ShouldHaveLength, 1)
			So(spans[0].Name(), ShouldEqual, "/trpc.helloworld.Greeter/Hello")
			So(spans[0].Status().Code, ShouldEqual, codes.Ok)
			So(spans[0].SpanKind(), ShouldEqual, apitrace.SpanKindClient)
			So(spans[0].SpanContext().HasTraceID(), ShouldBeTrue)
			So(len(spans[0].Attributes()) > 0, ShouldBeTrue)
		})

		metrics := exporter.GetMetrics()
		So(len(metrics) > 0, ShouldBeTrue)
		So(len(metrics[0].Metrics) > 0, ShouldBeTrue)
		So(metrics[0].Metrics[0].Name, ShouldEqual, "rpc.client.duration")
		So(metrics[0].Metrics[0].Unit, ShouldEqual, "ms")
	})
}
