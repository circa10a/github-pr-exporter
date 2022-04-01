package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	"github.com/google/go-github/v43/github"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	flag "github.com/spf13/pflag"
)

// User is a github user
type User string

// Users are a slice of github users
type Users []User

// Config represents config file read in from --config
type Config struct {
	Config struct {
		Users `yaml:"users"`
	} `yaml:"config"`
}

// read reads in configuration from the config file
func (c *Config) read() *Config {
	yamlFile, err := ioutil.ReadFile(*configFile)
	if err != nil {
		log.Fatal(err)
	}
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		log.Fatal(err)
	}
	return c
}

var (
	pullRequestsGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "pull_request",
		Help: "GitHub pull request",
	}, []string{"user", "created_at", "link", "status"})
	pullRequestsTotalGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "pull_request_total",
		Help: "Total number of pull requests opened by all users",
	}, []string{})
)

func init() {
	flag.Parse()
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
}

func main() {
	config := &Config{}
	config.read()

	numberOfUsers := len(config.Config.Users)
	if numberOfUsers == 0 {
		log.Fatal("no users in config file. nothing to do. exiting...")
	}
	log.Infof("read %d users from config file", numberOfUsers)

	// To cancel our goroutines
	ctx, cancel := context.WithCancel(context.Background())
	client := github.NewClient(nil)
	server := createHttpServer(*port)

	// Ensure we can cancel all of our goroutines
	go func(ctx context.Context) {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		<-c

		if err := server.Shutdown(ctx); err != nil {
			if err != context.Canceled {
				log.Fatalf("Error shutting down web server")
			}
			log.Info("Web server shutdown successfully")
		}

		cancel()
	}(ctx)

	// Run then exporter
	go func(ctx context.Context) {
		log.Infof("Starting github-pr-exporter on port %d", *port)
		// Get initial metrics
		collectPRMetrics(ctx, config, client)
		for {
			select {
			case <-ctx.Done():
				log.Info("Exporter shutdown successfully")
				return
			// Interval between fetching all new pull requests
			case <-time.After(time.Duration(*interval) * time.Second):
				collectPRMetrics(ctx, config, client)
			}
		}
	}(ctx)

	// Start server
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

// createHttpServer creates a new http.Server for our prometheus handler
func createHttpServer(port int) *http.Server {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	return server
}

// collectPRMetrics fetches all the pull request data and converts to prometheus metrics
func collectPRMetrics(ctx context.Context, config *Config, client *github.Client) {
	now := time.Now()
	past := now.AddDate(0, 0, -*daysAgo)
	beginningSearchDate := fmt.Sprintf("%d-%02d-%02d", past.Year(), past.Month(), past.Day())
	// Track total pull requests for the total counter
	pullRequestCount := 0
	log.Info("Searching for pull requests")
	// Loop all users passed in via config file
	for _, user := range config.Config.Users {
		// Get all pull requests opened by a user
		pullRequests := user.getPullRequests(ctx, client, beginningSearchDate, *ignoreUserNamespace)
		for _, pullRequest := range pullRequests {
			// Create new metric for each pull request
			pullRequestsGauge.With(prometheus.Labels{
				"user":       user.String(),
				"created_at": pullRequest.CreatedAt,
				"link":       pullRequest.PullRequestURL,
				"status":     pullRequest.Status,
			}).Set(1)
			// Increase the total counter
			pullRequestCount++
		}
		// Unauthenticated clients rate limit is 10 requests per minute
		// This sleep ensures no rate limits occur. The result is 1000 user searches every 90 minutes
		time.Sleep(time.Second * defaultRateLimitInteral)
	}
	// Set total counter
	pullRequestsTotalGauge.With(prometheus.Labels{}).Set(float64(pullRequestCount))
	// Wait for configured refresh interval
	log.Infof("Finished searching for pull requests. Sleeping for %d seconds", *interval)
}
