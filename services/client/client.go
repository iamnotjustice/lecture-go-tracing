package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
	jaeger "github.com/uber/jaeger-client-go"
	config "github.com/uber/jaeger-client-go/config"

	"github.com/iamnotjustice/lecture-go-tracing/xhttp"
)

// Init returns an instance of Jaeger Tracer that samples all of the traces and logs all spans to stdout.
func Init(service string) (opentracing.Tracer, io.Closer) {
	cfg := &config.Configuration{
		ServiceName: service,
		Sampler: &config.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &config.ReporterConfig{
			LogSpans: true,
		},
	}
	tracer, closer, err := cfg.NewTracer(config.Logger(jaeger.StdLogger))
	if err != nil {
		panic(fmt.Sprintf("ERROR: cannot init Jaeger: %v\n", err))
	}
	return tracer, closer
}

func main() {
	if len(os.Args) != 3 {
		panic("ERROR: Expecting one argument")
	}

	tracer, closer := Init("pretty-print")
	defer closer.Close()
	opentracing.SetGlobalTracer(tracer)

	data := os.Args[1]
	dataType := os.Args[2]

	span := tracer.StartSpan("prettify-input")
	span.SetTag("data_type", dataType)
	defer span.Finish()

	ctx := opentracing.ContextWithSpan(context.Background(), span)

	formatted := formatString(ctx, data, dataType)

	fmt.Print(formatted)
}

func formatString(ctx context.Context, data, dataType string) string {
	span, _ := opentracing.StartSpanFromContext(ctx, "formatString")
	defer span.Finish()

	v := url.Values{}
	v.Set("data", data)
	v.Set("type", dataType)
	url := "http://localhost:8081/format?" + v.Encode()
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err.Error())
	}

	// we mark this span as RPC-client (by adding a default tag)
	ext.SpanKindRPCClient.Set(span)
	ext.HTTPUrl.Set(span, url)
	ext.HTTPMethod.Set(span, "GET")
	span.Tracer().Inject(
		span.Context(),
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(req.Header),
	)

	resp, err := xhttp.Do(req)
	if err != nil {
		ext.LogError(span, err)
		panic(err.Error())
	}

	formatted := string(resp)

	span.LogFields(
		log.String("event", "string-format"),
		log.String("value", formatted),
	)

	return formatted
}
