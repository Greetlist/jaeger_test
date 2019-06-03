package main

import (
	"fmt"
	"net/http"
	"time"
	"log"

	jaegercfg "github.com/uber/jaeger-client-go/config"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	jaegerlog "github.com/uber/jaeger-client-go/log"
	_ "github.com/uber/jaeger-lib/metrics"
	openlog "github.com/opentracing/opentracing-go/log"
	"github.com/uber/jaeger-lib/metrics/prometheus"
)

var cfg jaegercfg.Configuration
var jLogger = jaegerlog.StdLogger
var jMetricsFactory = prometheus.New()

func init() {
	cfg = jaegercfg.Configuration{
		Sampler : &jaegercfg.SamplerConfig{
			Type : "const",
			Param : 1,
		},
		Reporter : &jaegercfg.ReporterConfig {
			LogSpans : true,
		},
	}
}

var secondFunc = func(w http.ResponseWriter, r *http.Request) {
	var sp opentracing.Span
	wireContext, err := opentracing.GlobalTracer().Extract(
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(r.Header))
	if err != nil {
		fmt.Printf("Extract Error is : %v.\n", err)
		return
	}
	sp = opentracing.StartSpan(
		"second_service_function",
		ext.RPCServerOption(wireContext))
	defer sp.Finish()
	time.Sleep(2 * time.Second)
	secondSeviceFunc(sp)
	w.Write([]byte("Finish"))
}

func secondSeviceFunc(parentSpan opentracing.Span) {
	sp := opentracing.StartSpan(
		"two_function",
		opentracing.ChildOf(parentSpan.Context()))
	defer sp.Finish()
	time.Sleep(1 * time.Second)
	sp.LogFields(
		openlog.String("Function_name", "secondSeviceFunc"),
		openlog.Int("Wating millis", 1000))
	fmt.Print("In The two function.\n")
}

func main() {
	tracer, closer, err := cfg.New(
		"elasticSearch-Jaeger",
		jaegercfg.Logger(jLogger),
		jaegercfg.Metrics(jMetricsFactory),
	)
	if err != nil {
		fmt.Printf("Init Tracer Error is : %v.\n", err)
		return
	}
	defer closer.Close()
	opentracing.SetGlobalTracer(tracer)

	http.HandleFunc("/second_services", secondFunc)
	log.Fatal(http.ListenAndServe(":10000", nil))
}
