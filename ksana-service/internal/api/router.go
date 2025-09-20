package api

import (
	"log/slog"
	"net/http"
	"strings"
	"time"
)

func NewRouter(handler *JobHandler, logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/jobs", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			handler.CreateJob(w, r)
		case http.MethodGet:
			handler.ListJobs(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/jobs/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/jobs/")
		parts := strings.Split(path, "/")

		if len(parts) == 1 && parts[0] != "" {
			switch r.Method {
			case http.MethodGet:
				handler.GetJob(w, r)
			case http.MethodPatch:
				handler.UpdateJob(w, r)
			case http.MethodDelete:
				handler.DeleteJob(w, r)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
			return
		}

		if len(parts) == 2 && parts[0] != "" && r.Method == http.MethodPost {
			switch parts[1] {
			case "run-now":
				handler.RunNow(w, r)
			case "pause":
				handler.PauseJob(w, r)
			case "resume":
				handler.ResumeJob(w, r)
			default:
				http.Error(w, "Not found", http.StatusNotFound)
			}
			return
		}

		http.Error(w, "Not found", http.StatusNotFound)
	})

	mux.HandleFunc("/health", handler.Health)

	return corsMiddleware(loggingMiddleware(logger)(mux))
}

func loggingMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
			next.ServeHTTP(wrapped, r)

			logger.Info("HTTP request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", wrapped.statusCode,
				"latency_ms", time.Since(start).Milliseconds(),
				"user_agent", r.UserAgent(),
			)
		})
	}
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 设置 CORS 头
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "86400") // 24小时

		// 处理预检请求
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		// 继续处理其他请求
		next.ServeHTTP(w, r)
	})
}