package main

import (
	"net/http"
	"fmt"
	"log"
	"time"

	jaegercfg "github.com/uber/jaeger-client-go/config"
	"github.com/opentracing/opentracing-go"
	jaegerlog "github.com/uber/jaeger-client-go/log"
	_ "github.com/uber/jaeger-lib/metrics"
	openlog "github.com/opentracing/opentracing-go/log"
	"github.com/uber/jaeger-lib/metrics/prometheus"
)

var cfg jaegercfg.Configuration
var jLogger = jaegerlog.StdLogger
var jMetricsFactory = prometheus.New()
//var jMetricsFactory = metrics.NullFactory

func init() {
	cfg = jaegercfg.Configuration {
		Sampler : &jaegercfg.SamplerConfig {
			Type : "const",
			Param : 1,
		},
		Reporter : &jaegercfg.ReporterConfig {
			LogSpans : true,
		},
	}
}

func spanLog(sp opentracing.Span, funcName string) {
	logString := "In The %s function."
	fmt.Sprintf(logString, funcName)
	sp.LogFields(
		openlog.String("Function_Name", logString),
		openlog.Int("Wating millis", 1000))
}

var testFunc = func(w http.ResponseWriter, r *http.Request) {
	sp := opentracing.StartSpan("test_function")
	defer sp.Finish()
	time.Sleep(1 * time.Second)
	fmt.Printf("In the test Handler.\n")
	spanLog(sp, "test_function")
	functionOne(sp)
}

var getFunc = func(w http.ResponseWriter, r *http.Request) {
	sp := opentracing.StartSpan("get_function")
	defer sp.Finish()
	time.Sleep(1 * time.Second)
	fmt.Printf("In the getFunc Handler.\n")
	spanLog(sp, "test_function")
	functionOne(sp)
	client := &http.Client{}
	httpReq, _ := http.NewRequest("GET", "http://localhost:10000/second_services", nil)
	opentracing.GlobalTracer().Inject(
		sp.Context(),
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(httpReq.Header))
	res, err := client.Do(httpReq)
	fmt.Printf("%v\n", res)
	if err != nil {
		fmt.Printf("Http Error is : %v.\n", err)
		return
	}
	functionOne(sp)
}

func functionOne(parentSpan opentracing.Span) {
	sp := opentracing.StartSpan(
		"one_function",
		opentracing.ChildOf(parentSpan.Context()))
	defer sp.Finish()
	time.Sleep(1 * time.Second)
	spanLog(sp, "one_function")
	fmt.Print("In The one function.\n")
	functionTwo(sp)
}

func functionTwo(parentSpan opentracing.Span) {
	sp := opentracing.StartSpan(
		"two_function",
		opentracing.ChildOf(parentSpan.Context()))
	defer sp.Finish()
	time.Sleep(1 * time.Second)
	spanLog(sp, "two_function")
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

	http.HandleFunc("/test", testFunc)
	http.HandleFunc("/get", getFunc)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
