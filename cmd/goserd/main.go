package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/kardianos/service"

	"github.com/BAIGUANGMEI/goser/internal/config"
	"github.com/BAIGUANGMEI/goser/internal/daemon"
	"github.com/BAIGUANGMEI/goser/internal/logger"
)

// program implements the service.Interface for kardianos/service.
type program struct {
	srv *daemon.Server
}

func (p *program) Start(s service.Service) error {
	go p.run()
	return nil
}

func (p *program) run() {
	if err := p.srv.Run(); err != nil {
		logger.Get().Fatalf("daemon error: %v", err)
	}
}

func (p *program) Stop(s service.Service) error {
	logger.Get().Info("service stop requested")
	// The daemon.Server handles graceful shutdown via OS signals
	proc, _ := os.FindProcess(os.Getpid())
	_ = proc.Signal(os.Interrupt)
	return nil
}

func main() {
	installFlag := flag.Bool("install", false, "Install as Windows service")
	uninstallFlag := flag.Bool("uninstall", false, "Uninstall Windows service")
	flag.Parse()

	// Ensure directories exist
	if err := config.EnsureDirs(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to create directories: %v\n", err)
		os.Exit(1)
	}

	// Load configuration
	loader := config.NewLoader()
	if err := loader.LoadGlobal(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to load global config: %v\n", err)
		os.Exit(1)
	}
	if err := loader.LoadServices(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to load services: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	cfg := loader.GetGlobal()
	if err := logger.Init(cfg.Daemon.LogDir); err != nil {
		fmt.Fprintf(os.Stderr, "failed to init logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	// Create daemon server
	srv := daemon.New(loader)

	// Configure as system service
	svcConfig := &service.Config{
		Name:        "GoSerDaemon",
		DisplayName: "GoSer Service Manager Daemon",
		Description: "GoSer non-blocking service manager daemon for managing background processes.",
	}

	prg := &program{srv: srv}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create service: %v\n", err)
		os.Exit(1)
	}

	// Handle install/uninstall flags
	if *installFlag {
		if err := s.Install(); err != nil {
			fmt.Fprintf(os.Stderr, "failed to install service: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Service installed successfully. Start with: sc start GoSerDaemon")
		return
	}

	if *uninstallFlag {
		if err := s.Uninstall(); err != nil {
			fmt.Fprintf(os.Stderr, "failed to uninstall service: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Service uninstalled successfully.")
		return
	}

	log := logger.Get()
	log.Info("GoSer daemon starting...")

	// Run (as service or standalone)
	if err := s.Run(); err != nil {
		log.Fatalf("daemon error: %v", err)
	}
}
