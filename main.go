package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ibhiyassine/GoKnot/internal/admin"
	"github.com/ibhiyassine/GoKnot/internal/config"
	"github.com/ibhiyassine/GoKnot/internal/health"
	"github.com/ibhiyassine/GoKnot/internal/loadbalancer"
	"github.com/ibhiyassine/GoKnot/internal/proxy"
	"github.com/ibhiyassine/GoKnot/internal/tui"
)

var pool = &loadbalancer.ServerPool{}
var Strats = map[string]loadbalancer.LoadBalancer{
	"round_robin":      loadbalancer.NewRoundRobin(pool),
	"least_connection": loadbalancer.NewLeastConnections(pool),
}

func main() {
	// This is the entry point for the reverse proxy

	//NOTE: This logging is completely written by AI, i like it :)
	// =========================================================================
	// 1. Setup Logging (File + Console)
	// =========================================================================
	if err := os.MkdirAll("logs", 0755); err != nil {
		log.Fatalf("Failed to create logs directory: %v", err)
	}

	logFile, err := os.OpenFile("logs/goknot.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer logFile.Close()

	// Write to both Terminal and File
	log.SetOutput(logFile)
	log.Println("Initializing GoKnot Load Balancer...")

	// Loading configuration
	cfg, err := config.LoadConfig("config.json")
	if err != nil {
		log.Fatal("Error loading configuration of reverse proxy")
	}

	// Initialize load Balancer
	lb := Strats[cfg.Strategy]
	if lb == nil {
		log.Fatal("Error loading the correct strategy")
	}

	// Start healthchecker and admin api
	checker := health.NewHealthChecker(lb, cfg.HealthCheckFreq)
	admin := admin.NewAdminServer(lb)

	checker.Start()

	// We just run the admin and don't care of it halting, no need for wait group
	go func() {
		//NOTE: admin listens in port 3333
		admin.Start(":" + strconv.Itoa(cfg.AdminPort))
	}()

	proxyHandler := proxy.NewProxyHandler(lb)

	serverAddr := fmt.Sprintf(":%d", cfg.Port)
	go func() {

		err = http.ListenAndServe(serverAddr, proxyHandler)
		if err != nil {
			log.Fatal("Proxy server failed...")
		}
	}()
	log.Printf("Proxy server listening on %s (Admin listening on :%d)", serverAddr, cfg.AdminPort)

	// Start the TUI
	p := tea.NewProgram(tui.InitialModel(lb, cfg.AdminPort))
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
