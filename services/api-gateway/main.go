package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	pathIDPattern = regexp.MustCompile(`^\d+$|^[0-9a-fA-F-]{8,}$`)

	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests.",
		},
		[]string{"service", "method", "path", "status"},
	)
	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request latency in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"service", "method", "path", "status"},
	)
)

func init() {
	prometheus.MustRegister(httpRequestsTotal, httpRequestDuration)
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func normalizePath(path string) string {
	if path == "" {
		return "/"
	}

	segments := strings.Split(path, "/")
	for i, segment := range segments {
		if pathIDPattern.MatchString(segment) {
			segments[i] = ":id"
		}
	}

	return strings.Join(segments, "/")
}

func withMetrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/metrics" {
			next.ServeHTTP(w, r)
			return
		}

		start := time.Now()
		recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(recorder, r)

		path := normalizePath(r.URL.Path)
		status := strconv.Itoa(recorder.status)
		httpRequestsTotal.WithLabelValues("api-gateway", r.Method, path, status).Inc()
		httpRequestDuration.WithLabelValues("api-gateway", r.Method, path, status).Observe(time.Since(start).Seconds())
	})
}

func newProxy(target string) *httputil.ReverseProxy {
	u, err := url.Parse(target)
	if err != nil {
		log.Fatalf("failed to parse target URL %s: %v", target, err)
	}
	return httputil.NewSingleHostReverseProxy(u)
}

type gateway struct {
	authProxy      *httputil.ReverseProxy
	inventoryProxy *httputil.ReverseProxy
	purchaseProxy  *httputil.ReverseProxy
	approvalProxy  *httputil.ReverseProxy
}

func (g *gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	path := r.URL.Path
	log.Printf("[gateway] %s %s", r.Method, path)

	if !strings.HasPrefix(path, "/api/") {
		http.NotFound(w, r)
		return
	}

	stripped := strings.TrimPrefix(path, "/api")
	r.URL.Path = stripped
	r.URL.RawPath = strings.TrimPrefix(r.URL.RawPath, "/api")

	switch {
	case strings.HasPrefix(stripped, "/auth"), strings.HasPrefix(stripped, "/users"):
		g.authProxy.ServeHTTP(w, r)
	case strings.HasPrefix(stripped, "/inventory"), strings.HasPrefix(stripped, "/dep"):
		g.inventoryProxy.ServeHTTP(w, r)
	case strings.HasPrefix(stripped, "/purchase"):
		// Remove /purchase prefix before proxying
		r.URL.Path = strings.TrimPrefix(r.URL.Path, "/purchase")
		r.URL.RawPath = strings.TrimPrefix(r.URL.RawPath, "/purchase")
		g.purchaseProxy.ServeHTTP(w, r)
	case strings.HasPrefix(stripped, "/approval"):
		// Remove /approval prefix before proxying
		r.URL.Path = strings.TrimPrefix(r.URL.Path, "/approval")
		r.URL.RawPath = strings.TrimPrefix(r.URL.RawPath, "/approval")
		g.approvalProxy.ServeHTTP(w, r)
	default:
		http.NotFound(w, r)
	}
}

func main() {
	authURL := os.Getenv("AUTH_SERVICE_URL")
	if authURL == "" {
		authURL = "http://localhost:6767"
	}
	inventoryURL := os.Getenv("INVENTORY_SERVICE_URL")
	if inventoryURL == "" {
		inventoryURL = "http://localhost:6768"
	}
	purchaseURL := os.Getenv("PURCHASE_SERVICE_URL")
	if purchaseURL == "" {
		purchaseURL = "http://localhost:6769"
	}
	approvalURL := os.Getenv("APPROVAL_SERVICE_URL")
	if approvalURL == "" {
		approvalURL = "http://localhost:6770"
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	g := &gateway{
		authProxy:      newProxy(authURL),
		inventoryProxy: newProxy(inventoryURL),
		purchaseProxy:  newProxy(purchaseURL),
		approvalProxy:  newProxy(approvalURL),
	}

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.Handle("/", g)

	log.Printf("api-gateway starting on port %s", port)
	log.Printf("  auth-service      -> %s", authURL)
	log.Printf("  inventory-service -> %s", inventoryURL)
	log.Printf("  purchase-service  -> %s", purchaseURL)
	log.Printf("  approval-service  -> %s", approvalURL)
	if err := http.ListenAndServe(":"+port, withMetrics(mux)); err != nil {
		log.Fatalf("failed to start gateway: %v", err)
	}
}
