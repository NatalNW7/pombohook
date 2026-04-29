package cli

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/NatalNW7/pombohook/internal/storage"
)

// RunPing tests the connection to the server and saves config on success.
func RunPing(store *storage.Storage, w io.Writer, server, token string) error {
	if server == "" {
		return fmt.Errorf("--server is required")
	}
	if token == "" {
		return fmt.Errorf("--token is required")
	}

	// Build the ping URL (convert ws:// to http:// for the REST ping)
	pingURL := server + "/ping"

	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest(http.MethodGet, pingURL, nil)
	if err != nil {
		fmt.Fprintf(w, "✗ Could not reach server at %s\n  Check the URL and try again.\n", server)
		return fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintf(w, "✗ Could not reach server at %s\n  Check the URL and try again.\n", server)
		return fmt.Errorf("connecting to server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		fmt.Fprintf(w, "✗ Authentication failed. Invalid token.\n")
		return fmt.Errorf("authentication failed")
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(w, "✗ Server returned status %d\n", resp.StatusCode)
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	// Save config
	cfg := storage.PomboConfig{Server: server, Token: token}
	if err := store.SaveConfig(cfg); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	fmt.Fprintf(w, `🕊️  Connection established! Your pigeon is ready.
    Server: %s
    Auth:   ✓ Valid

    Next steps:
      pombo route --path=/webhook/test --port=8081   # register a delivery route
      pombo go                                        # start delivering
`, server)

	return nil
}
