package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	jaeger "github.com/uber/jaeger-client-go"
	config "github.com/uber/jaeger-client-go/config"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"
)

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
	tracer, closer := Init("formatter")
	defer closer.Close()

	http.HandleFunc("/format", func(w http.ResponseWriter, r *http.Request) {
		spanCtx, _ := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
		// it's a RPC-server side, so we need to start a span which is tagged as such
		span := tracer.StartSpan("format", ext.RPCServerOption(spanCtx))

		defer span.Finish()

		toFormat := r.FormValue("data")
		dt := r.FormValue("type")

		span.SetTag("data_type", dt)

		formatted, err := formatData(toFormat, dt)
		if err != nil {
			w.WriteHeader(400)
		}

		span.LogFields(
			otlog.String("event", "format"),
			otlog.String("value", toFormat),
		)

		w.Write([]byte(formatted))
	})

	log.Fatal(http.ListenAndServe(":8081", nil))
}

func formatData(data string, dataType string) (string, error) {
	if dataType == "json" {
		return formatJSON(data)
	}
	if dataType == "xml" {
		return formatXML(data)
	}

	return "", errors.New("unknown type")
}

func formatJSON(data string) (string, error) {
	var out bytes.Buffer
	err := json.Indent(&out, []byte(data), "", "\t")
	if err != nil {
		return "", err
	}

	return out.String(), nil
}

func formatXML(data string) (string, error) {
	var out bytes.Buffer
	decoder := xml.NewDecoder(bytes.NewReader([]byte(data)))
	encoder := xml.NewEncoder(&out)
	encoder.Indent("", "  ")
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			encoder.Flush()
			return out.String(), nil
		}
		if err != nil {
			return "", err
		}
		err = encoder.EncodeToken(token)
		if err != nil {
			return "", err
		}
	}
}
