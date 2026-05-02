package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/NatalNW7/pombohook/internal/cli"
	"github.com/NatalNW7/pombohook/internal/storage"
)

func main() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: could not determine home directory: %v\n", err)
		os.Exit(1)
	}

	store := storage.NewStorage(filepath.Join(homeDir, ".pombo"))

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "ping":
		runPing(store, os.Args[2:])
	case "route":
		runRoute(store, os.Args[2:])
	case "go":
		runGo(store, os.Args[2:])
	case "sleep":
		runSleep(store)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`🕊️  pombo - PomboHook CLI

Usage:
  pombo <command> [options]

Commands:
  ping     Test connection and authenticate with server
  route    Manage webhook delivery routes
  go       Start listening and delivering webhooks
  sleep    Stop a background pombo session`)
}

func runPing(store *storage.Storage, args []string) {
	fs := flag.NewFlagSet("ping", flag.ExitOnError)
	server := fs.String("server", "", "Server URL (e.g., wss://pomboserver.fly.com)")
	token := fs.String("token", "", "Authentication token")
	fs.Parse(args)

	if err := cli.RunPing(store, os.Stdout, *server, *token); err != nil {
		os.Exit(1)
	}
}

func runRoute(store *storage.Storage, args []string) {
	fs := flag.NewFlagSet("route", flag.ExitOnError)
	path := fs.String("path", "", "Webhook path (e.g., /webhook/mp)")
	port := fs.Int("port", 0, "Local port (e.g., 8081)")
	list := fs.Bool("list", false, "List all routes")
	remove := fs.String("remove", "", "Remove a route by path")
	clear := fs.Bool("clear", false, "Clear all routes")
	fs.Parse(args)

	switch {
	case *list:
		if err := cli.RunRouteList(store, os.Stdout); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case *remove != "":
		if err := cli.RunRouteRemove(store, os.Stdout, *remove); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case *clear:
		if err := cli.RunRouteClear(store, os.Stdout); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	default:
		if *path == "" || *port == 0 {
			fmt.Fprintln(os.Stderr, "Usage: pombo route --path=/webhook/mp --port=8081")
			os.Exit(1)
		}
		if err := cli.RunRouteAdd(store, os.Stdout, *path, *port); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}
}

func runGo(store *storage.Storage, args []string) {
	fs := flag.NewFlagSet("go", flag.ExitOnError)
	background := fs.Bool("background", false, "Run in background (daemon mode)")
	daemon := fs.Bool("daemon", false, "Internal: run as daemon child process")
	fs.Parse(args)

	if *daemon {
		// We are the daemon child process — run foreground (output goes to log file)
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
		if err := cli.RunGo(store, os.Stdout, logger); err != nil {
			fmt.Fprintf(os.Stderr, "Daemon error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if *background {
		executable, _ := os.Executable()
		if err := cli.RunGoBackground(store, os.Stdout, executable); err != nil {
			os.Exit(1)
		}
		return
	}

	// Foreground mode (original behavior)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	if err := cli.RunGo(store, os.Stdout, logger); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runSleep(store *storage.Storage) {
	if err := cli.RunSleep(store, os.Stdout); err != nil {
		os.Exit(1)
	}
}
