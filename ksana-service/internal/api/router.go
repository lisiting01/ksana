package api

import (
	"ksana-service/internal/auth"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

func NewRouter(handler *JobHandler, authManager *auth.Manager, logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()

	authMiddleware := apiKeyMiddleware(authManager, logger)

	mux.HandleFunc("/jobs", authMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			handler.CreateJob(w, r)
		case http.MethodGet:
			handler.ListJobs(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	mux.HandleFunc("/jobs/", authMiddleware(func(w http.ResponseWriter, r *http.Request) {
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
	}))

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
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key")
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

func apiKeyMiddleware(authManager *auth.Manager, logger *slog.Logger) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var apiKey string

			authHeader := r.Header.Get("Authorization")
			if authHeader != "" {
				const prefix = "ApiKey "
				if strings.HasPrefix(authHeader, prefix) {
					apiKey = strings.TrimPrefix(authHeader, prefix)
				}
			}

			if apiKey == "" {
				apiKey = r.Header.Get("X-API-Key")
			}

			if apiKey == "" {
				logger.Warn("API key authentication failed: missing key",
					"client_ip", getClientIP(r),
					"path", r.URL.Path,
					"method", r.Method,
				)
				w.Header().Set("WWW-Authenticate", "ApiKey")
				http.Error(w, "Authentication required", http.StatusUnauthorized)
				return
			}

			if !authManager.Validate(apiKey) {
				logger.Warn("API key authentication failed: invalid key",
					"client_ip", getClientIP(r),
					"path", r.URL.Path,
					"method", r.Method,
				)
				w.Header().Set("WWW-Authenticate", "ApiKey")
				http.Error(w, "Invalid API key", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func getClientIP(r *http.Request) string {
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = r.Header.Get("X-Real-IP")
	}
	if ip == "" {
		ip = r.RemoteAddr
	}
	return ip
}