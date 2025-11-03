package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/reybrally/TCP-Load-Balancer/internal/adapter/balancer"
	"github.com/reybrally/TCP-Load-Balancer/internal/adapter/config"
	"github.com/reybrally/TCP-Load-Balancer/internal/adapter/health"
	"github.com/reybrally/TCP-Load-Balancer/internal/adapter/listener"
	prommetrics "github.com/reybrally/TCP-Load-Balancer/internal/adapter/metrics"
	"github.com/reybrally/TCP-Load-Balancer/internal/adapter/repository"
	"github.com/reybrally/TCP-Load-Balancer/internal/application/usecase"
	appcfg "github.com/reybrally/TCP-Load-Balancer/internal/config"
	"github.com/reybrally/TCP-Load-Balancer/internal/domain/model"
	"github.com/reybrally/TCP-Load-Balancer/internal/pkg/logger"
)

const (
	Version             = "v1.0.0"
	HealthCheckInterval = 10 * time.Second
	HealthCheckTimeout  = 2 * time.Second
	ShutdownTimeout     = 30 * time.Second
)

func main() {
	log := logger.New("development")
	defer log.Sync()

	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.PrintBanner(Version, fmt.Sprintf("%d", cfg.Server.Port))

	metrics := prommetrics.NewPrometheusMetrics()
	log.Infof("Prometheus metrics collector initialized")

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		})
		log.Infof("Metrics endpoint started on :9090/metrics")
		if err := http.ListenAndServe(":9090", mux); err != nil {
			log.Errorf("Metrics server error: %v", err)
		}
	}()

	repo := repository.New()

	if err := initBackends(cfg, repo, log); err != nil {
		log.Fatalf("Failed to initialize backends: %v", err)
	}

	lb := balancer.New()
	log.Infof("Load balancing algorithm: Round-Robin")

	healthChecker := health.New(HealthCheckTimeout, metrics, log)
	log.Infof("Health checker initialized (interval: %v, timeout: %v)", HealthCheckInterval, HealthCheckTimeout)

	handleConnUseCase := usecase.New(lb, repo, metrics, log)

	tcpListener, err := listener.New(cfg.Server.Host, cfg.Server.Port, log)
	if err != nil {
		log.Fatalf("Failed to create listener: %v", err)
	}
	defer tcpListener.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var totalConnections atomic.Int64
	var activeConnections atomic.Int64

	wg.Add(1)
	go func() {
		defer wg.Done()
		runHealthChecks(ctx, repo, healthChecker, log)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		printStats(ctx, repo, &totalConnections, &activeConnections, log)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := tcpListener.Listen(ctx, handleConnUseCase); err != nil && err != context.Canceled {
			log.Warnf("Listener error: %v", err)
		}
	}()

	log.Infof("TCP Load Balancer is ready to accept connections")
	log.Infof("Listening on %s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Infof("Press Ctrl+C to gracefully shutdown")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Warnf("⚠Shutdown signal received, initiating graceful shutdown...")

	cancel()
	tcpListener.Close()

	log.Infof("Waiting for active connections to complete (max %v)...", ShutdownTimeout)

	done := make(chan struct{})
	go func() {
		wg.Wait()
		done <- struct{}{}
	}()

	select {
	case <-done:
		log.Infof("All connections closed gracefully")
	case <-time.After(ShutdownTimeout):
		log.Warnf("⚠Timeout waiting for connections to close, forcing shutdown")
	}

	printFinalStats(repo, &totalConnections, log)

	log.Infof("TCP Load Balancer stopped successfully")
}

func initBackends(cfg *appcfg.Config, repo interface {
	Add(context.Context, *model.Backend) error
}, log *logger.Logger) error {
	log.Infof("Initializing %d backend servers...", len(cfg.Backends))

	for i, backendCfg := range cfg.Backends {
		backend := model.NewBackend(
			fmt.Sprintf("backend-%d", i),
			backendCfg.Address,
			backendCfg.Port,
			backendCfg.Weight,
		)
		if err := repo.Add(context.Background(), backend); err != nil {
			return fmt.Errorf("failed to add backend %s: %w", backend.GetAddress(), err)
		}
		log.Infof("  ✓ Backend %d: %s (weight: %d)", i+1, backend.GetAddress(), backend.Weight)
	}

	return nil
}

func runHealthChecks(ctx context.Context, repo interface {
	GetAll(context.Context) []*model.Backend
}, healthChecker interface {
	Check(context.Context, *model.Backend) bool
}, log *logger.Logger) {
	ticker := time.NewTicker(HealthCheckInterval)
	defer ticker.Stop()

	log.Debugf("Health check daemon started")

	for {
		select {
		case <-ctx.Done():
			log.Debugf("Health check daemon stopped")
			return
		case <-ticker.C:
			backends := repo.GetAll(ctx)
			var wg sync.WaitGroup
			var healthyCount atomic.Int32
			var unhealthyCount atomic.Int32

			for _, backend := range backends {
				wg.Add(1)
				go func(b *model.Backend) {
					defer wg.Done()

					wasHealthy := b.GetHealthy()
					isHealthy := healthChecker.Check(ctx, b)
					b.SetHealthy(isHealthy)

					if isHealthy {
						healthyCount.Add(1)
					} else {
						unhealthyCount.Add(1)
					}

					if wasHealthy != isHealthy {
						if isHealthy {
							log.Infof("Backend %s recovered", b.GetAddress())
						} else {
							log.Warnf("Backend %s went down", b.GetAddress())
						}
					}

					status := "healthy"
					if !isHealthy {
						status = "unhealthy"
					}
					log.Debugf("Health check for %s: %s", b.GetAddress(), status)
				}(backend)
			}

			wg.Wait()
			log.Debugf("Health check completed: %d healthy, %d unhealthy", healthyCount.Load(), unhealthyCount.Load())
		}
	}
}

func printStats(ctx context.Context, repo interface {
	GetAll(context.Context) []*model.Backend
	GetHealthy(context.Context) []*model.Backend
}, totalConns, activeConns *atomic.Int64, log *logger.Logger) {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			allBackends := repo.GetAll(ctx)
			healthyBackends := repo.GetHealthy(ctx)

			log.Infof("Statistics: Total Backends: %d, Healthy: %d, Unhealthy: %d",
				len(allBackends),
				len(healthyBackends),
				len(allBackends)-len(healthyBackends))

			for _, b := range allBackends {
				status := "OK"
				if !b.GetHealthy() {
					status = "NOT OK"
				}
				log.Infof("  %s %s - Active connections: %d",
					status,
					b.GetAddress(),
					b.GetActiveConnections())
			}
		}
	}
}

func printFinalStats(repo interface {
	GetAll(context.Context) []*model.Backend
}, totalConns *atomic.Int64, log *logger.Logger) {
	log.Infof("Final Statistics:")
	log.Infof("  Total connections processed: %d", totalConns.Load())

	backends := repo.GetAll(context.Background())
	for _, b := range backends {
		log.Infof("  Backend %s: %d active connections",
			b.GetAddress(),
			b.GetActiveConnections())
	}
}
