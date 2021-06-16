package rkgintrace

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/rookie-ninja/rk-entry/entry"
	"github.com/rookie-ninja/rk-gin/interceptor/basic"
	"github.com/rookie-ninja/rk-gin/interceptor/extension"
	"github.com/rookie-ninja/rk-gin/interceptor/log/zap"
	"github.com/rookie-ninja/rk-logger"
	"go.opentelemetry.io/contrib"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/stdout"
	"go.opentelemetry.io/otel/exporters/trace/jaeger"
	"go.opentelemetry.io/otel/propagation"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/semconv"
	oteltrace "go.opentelemetry.io/otel/trace"
	"log"
	"os"
	"path"
)

func CreateFileExporter(outputPath string, opts ...stdout.Option) sdktrace.SpanExporter {
	if opts == nil {
		opts = make([]stdout.Option, 0)
	}

	if outputPath == "" {
		outputPath = "stdout"
	}

	if outputPath == "stdout" {
		opts = append(opts,
			stdout.WithPrettyPrint(),
			stdout.WithoutMetricExport())
	} else {
		// init lumberjack logger
		writer := rklogger.NewLumberjackConfigDefault()

		if !path.IsAbs(outputPath) {
			wd, _ := os.Getwd()
			outputPath = path.Join(wd, outputPath)
		}

		writer.Filename = outputPath

		opts = append(opts, stdout.WithWriter(writer))
	}

	exporter, _ := stdout.NewExporter(opts...)

	return exporter
}

// TODO: Wait for opentelemetry update version of jeager exporter. Current exporter is not compatible with jaeger agent.
func CreateJaegerExporter(host, port string) sdktrace.SpanExporter {
	if len(host) < 1 {
		host = "localhost"
	}

	if len(port) < 1 {
		port = "6832"
	}

	exporter, _ := jaeger.NewRawExporter(
		jaeger.WithAgentEndpoint(
			jaeger.WithAgentHost(host),
			jaeger.WithAgentPort(port),
			jaeger.WithLogger(log.New(os.Stdout, "", 0))))

	return exporter
}

func TelemetryInterceptor(opts ...Option) gin.HandlerFunc {
	set := &optionSet{
		EntryName: rkginbasic.RkEntryNameValue,
		EntryType: rkginbasic.RkEntryTypeValue,
	}

	for i := range opts {
		opts[i](set)
	}

	if set.Exporter == nil {
		set.Exporter, _ = stdout.NewExporter(stdout.WithPrettyPrint())
	}

	if set.Processor == nil {
		set.Processor = sdktrace.NewBatchSpanProcessor(set.Exporter)
	}

	if set.Provider == nil {
		set.Provider = sdktrace.NewTracerProvider(
			sdktrace.WithSampler(sdktrace.AlwaysSample()),
			sdktrace.WithSpanProcessor(set.Processor),
			sdktrace.WithResource(
				sdkresource.NewWithAttributes(
					attribute.String("service.name", rkentry.GlobalAppCtx.GetAppInfoEntry().AppName),
					attribute.String("service.version", rkentry.GlobalAppCtx.GetAppInfoEntry().Version),
					attribute.String("service.entryName", set.EntryName),
					attribute.String("service.entryType", set.EntryType),
				)),
		)
	}

	set.Tracer = set.Provider.Tracer(set.EntryName, oteltrace.WithInstrumentationVersion(contrib.SemVersion()))

	if set.Propagator == nil {
		set.Propagator = propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{})
	}

	if _, ok := optionsMap[set.EntryName]; !ok {
		optionsMap[set.EntryName] = set
	}

	return func(ctx *gin.Context) {
		opts := []oteltrace.SpanOption{
			oteltrace.WithAttributes(semconv.NetAttributesFromHTTPRequest("tcp", ctx.Request)...),
			oteltrace.WithAttributes(semconv.EndUserAttributesFromHTTPRequest(ctx.Request)...),
			oteltrace.WithAttributes(semconv.HTTPServerAttributesFromHTTPRequest(rkentry.GlobalAppCtx.GetAppInfoEntry().AppName, ctx.FullPath(), ctx.Request)...),
			oteltrace.WithAttributes(localeToAttributes()...),
			oteltrace.WithSpanKind(oteltrace.SpanKindServer),
		}

		// 1: extract tracing info from request header
		spanCtx := oteltrace.SpanContextFromContext(
			set.Propagator.Extract(ctx.Request.Context(), propagation.HeaderCarrier(ctx.Request.Header)))

		spanName := ctx.FullPath()
		if len(spanName) < 1 {
			spanName = "rk-span-default"
		}

		// 2: start new span
		newRequestCtx, span := set.Tracer.Start(
			oteltrace.ContextWithRemoteSpanContext(ctx.Request.Context(), spanCtx),
			spanName, opts...)
		// 2.1: pass the span through the request context
		ctx.Request = ctx.Request.WithContext(newRequestCtx)

		defer span.End()

		// 3: read trace id, tracer, traceProvider, propagator and logger into event data and gin context
		rkginlog.GetEvent(ctx).SetTraceId(span.SpanContext().TraceID().String())
		ctx.Set(rkginbasic.RkTraceIdKey, span.SpanContext().TraceID().String())
		ctx.Set(rkginbasic.RkTracerKey, span.Tracer())
		ctx.Set(rkginbasic.RkTracerProviderKey, set.Provider)
		ctx.Set(rkginbasic.RkPropagatorKey, set.Propagator)

		// 4: set trace id into response header with header key defined in extension
		if extensionSet := rkginextension.GetOptionSet(ctx); extensionSet != nil {
			ctx.Writer.Header().Set(extensionSet.TraceIdKey, span.SpanContext().TraceID().String())
		} else {
			ctx.Writer.Header().Set(rkginextension.TraceIdHeaderKeyDefault, span.SpanContext().TraceID().String())
		}

		// 5: Call rest of function
		ctx.Next()

		// 6: Set attribute of response code into span
		attrs := semconv.HTTPAttributesFromHTTPStatusCode(ctx.Writer.Status())
		spanStatus, spanMessage := semconv.SpanStatusFromHTTPStatusCode(ctx.Writer.Status())
		span.SetAttributes(attrs...)
		span.SetStatus(spanStatus, spanMessage)
		if len(ctx.Errors) > 0 {
			span.SetAttributes(attribute.String("errors", ctx.Errors.String()))
		}
	}
}

func localeToAttributes() []attribute.KeyValue {
	res := []attribute.KeyValue{
		attribute.String(rkginbasic.Realm.Key, rkginbasic.Realm.String),
		attribute.String(rkginbasic.Region.Key, rkginbasic.Region.String),
		attribute.String(rkginbasic.AZ.Key, rkginbasic.AZ.String),
		attribute.String(rkginbasic.Domain.Key, rkginbasic.Domain.String),
	}

	return res
}

var optionsMap = make(map[string]*optionSet)

// options which is used while initializing tracing interceptor
type optionSet struct {
	EntryName  string
	EntryType  string
	Exporter   sdktrace.SpanExporter
	Processor  sdktrace.SpanProcessor
	Provider   *sdktrace.TracerProvider
	Propagator propagation.TextMapPropagator
	Tracer     oteltrace.Tracer
}

type Option func(*optionSet)

func WithExporter(exporter sdktrace.SpanExporter) Option {
	return func(opt *optionSet) {
		if exporter != nil {
			opt.Exporter = exporter
		}
	}
}

func WithSpanProcessor(processor sdktrace.SpanProcessor) Option {
	return func(opt *optionSet) {
		if processor != nil {
			opt.Processor = processor
		}
	}
}

func WithTracerProvider(provider *sdktrace.TracerProvider) Option {
	return func(opt *optionSet) {
		if provider != nil {
			opt.Provider = provider
		}
	}
}

func WithPropagator(propagator propagation.TextMapPropagator) Option {
	return func(opt *optionSet) {
		if propagator != nil {
			opt.Propagator = propagator
		}
	}
}

func WithEntryNameAndType(entryName, entryType string) Option {
	return func(opt *optionSet) {
		opt.EntryName = entryName
		opt.EntryType = entryType
	}
}

func ShutdownExporters() {
	for _, v := range optionsMap {
		v.Exporter.Shutdown(context.Background())
	}
}
