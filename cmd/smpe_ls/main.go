package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/cybersorcerer/smpe_ls/internal/handler"
	"github.com/cybersorcerer/smpe_ls/internal/logger"
	"github.com/cybersorcerer/smpe_ls/pkg/lsp"
)

var (
	version  = "0.7.2"
	debug    = flag.Bool("debug", false, "Enable debug logging")
	showVer  = flag.Bool("version", false, "Show version")
	dataPath = flag.String("data", "", "Path to smpe.json data file (default: ~/.local/share/smpe_ls/smpe.json)")
)

func getDefaultDataPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	// Platform-specific data directory
	if runtime.GOOS == "windows" {
		// Windows: %LOCALAPPDATA%\smpe_ls\smpe.json
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData != "" {
			return filepath.Join(localAppData, "smpe_ls", "smpe.json")
		}
		// Fallback if LOCALAPPDATA not set
		return filepath.Join(homeDir, "AppData", "Local", "smpe_ls", "smpe.json")
	}

	// Linux/macOS: ~/.local/share/smpe_ls/smpe.json
	return filepath.Join(homeDir, ".local", "share", "smpe_ls", "smpe.json")
}

func main() {
	// Custom usage message to show --debug instead of -debug
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  --data string\n")
		fmt.Fprintf(os.Stderr, "    	Path to smpe.json data file (default: ~/.local/share/smpe_ls/smpe.json)\n")
		fmt.Fprintf(os.Stderr, "  --debug\n")
		fmt.Fprintf(os.Stderr, "    	Enable debug logging\n")
		fmt.Fprintf(os.Stderr, "  --version\n")
		fmt.Fprintf(os.Stderr, "    	Show version\n")
	}

	flag.Parse()

	if *showVer {
		fmt.Printf("smpe_ls version %s\n", version)
		os.Exit(0)
	}

	// Initialize logger
	if err := logger.Init(*debug); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Close()

	logger.Info("smpe_ls version %s starting", version)

	// Determine data path
	finalDataPath := *dataPath
	if finalDataPath == "" {
		finalDataPath = getDefaultDataPath()
	}
	logger.Info("Data file: %s", finalDataPath)

	// Create handler
	h, err := handler.New(version, finalDataPath)
	if err != nil {
		logger.Fatal("Failed to create handler: %v", err)
	}

	// Create LSP server
	server := lsp.NewServer(os.Stdin, os.Stdout, h)
	h.SetServer(server)

	logger.Info("LSP server starting")

	// Start server (blocks until client disconnects)
	if err := server.Start(); err != nil {
		logger.Fatal("Server error: %v", err)
	}

	logger.Info("Server stopped")
}
