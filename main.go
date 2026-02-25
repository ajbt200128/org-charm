package main

import (
	"context"
	"errors"
	"flag"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"org-charm/org"
	"org-charm/ui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/activeterm"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/charmbracelet/wish/logging"
	"github.com/muesli/termenv"
)

func main() {
	// Command line flags
	host := flag.String("host", "localhost", "Host to listen on")
	port := flag.String("port", "2222", "Port to listen on")
	orgDir := flag.String("dir", "./orgfiles", "Directory containing org files")
	keyPath := flag.String("key", ".ssh/id_ed25519", "Path to host key")
	flag.Parse()

	// Setup logging with charm's log library
	log.SetLevel(log.DebugLevel)
	log.SetReportTimestamp(true)
	log.SetReportCaller(false)

	// Verify org directory exists
	if _, err := os.Stat(*orgDir); os.IsNotExist(err) {
		log.Warn("Org directory does not exist, creating it", "dir", *orgDir)
		if err := os.MkdirAll(*orgDir, 0755); err != nil {
			log.Fatal("Failed to create org directory", "error", err)
		}
	}

	// List org files
	files, err := org.ListOrgFiles(*orgDir)
	if err != nil {
		log.Fatal("Failed to list org files", "error", err)
	}
	log.Info("Found org files", "count", len(files))

	// Create the bubbletea handler
	teaHandler := makeTeaHandler(files)

	// Create SSH server with wish
	srv, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(*host, *port)),
		wish.WithHostKeyPath(*keyPath),
		wish.WithMiddleware(
			// Bubbletea middleware with forced TrueColor - serves the TUI to each SSH session
			bubbletea.MiddlewareWithColorProfile(teaHandler, termenv.TrueColor),
			// Require an active terminal
			activeterm.Middleware(),
			// Logging middleware using charm's log
			logging.Middleware(),
		),
	)
	if err != nil {
		log.Fatal("Could not create server", "error", err)
	}

	// Handle graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	log.Info("Starting SSH server",
		"host", *host,
		"port", *port,
		"org_dir", *orgDir,
	)
	log.Info("Connect with: ssh localhost -p " + *port)

	// Start server in goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			log.Fatal("Server error", "error", err)
		}
	}()

	// Wait for shutdown signal
	<-done
	log.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("Failed to shutdown server gracefully", "error", err)
	}

	log.Info("Server stopped")
}

// makeTeaHandler creates a bubbletea handler function for wish
func makeTeaHandler(files []string) bubbletea.Handler {
	return func(sess ssh.Session) (tea.Model, []tea.ProgramOption) {
		// Get the renderer for this SSH session and force TrueColor
		renderer := bubbletea.MakeRenderer(sess)
		renderer.SetColorProfile(termenv.TrueColor)

		// Get PTY info for window size
		pty, _, _ := sess.Pty()

		log.Info("New SSH session",
			"user", sess.User(),
			"term", pty.Term,
			"width", pty.Window.Width,
			"height", pty.Window.Height,
		)

		// Create the model with session-specific renderer
		model := ui.NewModel(renderer, files)

		return model, []tea.ProgramOption{
			tea.WithAltScreen(),
			tea.WithMouseCellMotion(),
		}
	}
}
