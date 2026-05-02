package tests

import (
	"bytes"
	"net/http"
	"testing"
	"time"

	"github.com/NatalNW7/pombohook/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestE2E_Delivery(t *testing.T) {
	env := setupE2EEnvironment(t, 20)
	defer env.tsServer.Close()
	defer env.targetServer.Close()

	// 1. Connect and register client
	err := env.client.Connect()
	require.NoError(t, err)
	err = env.client.SendRegister([]config.RouteMapping{{Path: "/webhook/test", Port: env.targetPort}})
	require.NoError(t, err)
	
	// Start listening in background
	go env.client.Listen()
	time.Sleep(100 * time.Millisecond) // wait for connection to register

	t.Run("e2e_should_deliver_POST_webhook", func(t *testing.T) {
		env.setReceivedBody("")
		env.setReceivedMethod("")

		payload := `{"event":"charge"}`
		req, _ := http.NewRequest(http.MethodPost, env.tsServer.URL+"/webhook/test", bytes.NewReader([]byte(payload)))
		
		// 2. Send webhook to server
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// 3. Verify server immediate response
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// 4. Wait for delivery and verify target received it
		time.Sleep(100 * time.Millisecond)
		assert.Equal(t, "POST", env.getReceivedMethod())
		assert.Equal(t, payload, env.getReceivedBody())
	})

	t.Run("e2e_should_deliver_GET_webhook", func(t *testing.T) {
		env.setReceivedBody("")
		env.setReceivedMethod("")

		req, _ := http.NewRequest(http.MethodGet, env.tsServer.URL+"/webhook/test", nil)
		
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		time.Sleep(100 * time.Millisecond)
		assert.Equal(t, "GET", env.getReceivedMethod())
	})

	t.Run("e2e_should_return_200_immediately", func(t *testing.T) {
		// Just to explicitly verify the response structure
		req, _ := http.NewRequest(http.MethodPut, env.tsServer.URL+"/webhook/test", bytes.NewReader([]byte(`{}`)))
		
		start := time.Now()
		resp, err := http.DefaultClient.Do(req)
		duration := time.Since(start)
		
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Less(t, duration, 50*time.Millisecond, "Server must respond immediately without waiting for tunnel")
	})
}

func TestE2E_Queueing(t *testing.T) {
	env := setupE2EEnvironment(t, 2)
	defer env.tsServer.Close()
	defer env.targetServer.Close()

	// Pre-register the route via connect then disconnect.
	// After disconnect, route is cleaned up, so register directly in registry for offline scenario.
	env.registry.Register("/webhook/test", env.targetPort, "offline-tunnel")

	t.Run("e2e_should_queue_when_cli_offline", func(t *testing.T) {
		// CLI is NOT connected
		req, _ := http.NewRequest(http.MethodPost, env.tsServer.URL+"/webhook/test", bytes.NewReader([]byte(`{"queued":true}`)))
		
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusAccepted, resp.StatusCode) // 202 Queued
		assert.Empty(t, env.getReceivedBody()) // Target did NOT receive it
	})

	t.Run("e2e_should_flush_queue_on_reconnect", func(t *testing.T) {
		// Queue a message
		req, _ := http.NewRequest(http.MethodPost, env.tsServer.URL+"/webhook/test", bytes.NewReader([]byte(`{"flush":true}`)))
		http.DefaultClient.Do(req)

		// Connect CLI again
		err := env.client.Connect()
		require.NoError(t, err)
		err = env.client.SendRegister([]config.RouteMapping{{Path: "/webhook/test", Port: env.targetPort}})
		require.NoError(t, err)
		go env.client.Listen()

		// Server should flush queue to connected client automatically
		time.Sleep(200 * time.Millisecond) // Wait for flush and delivery

		assert.Equal(t, "POST", env.getReceivedMethod())
		assert.Contains(t, env.getReceivedBody(), `"flush":true`) // Ensure we received the queued message
	})
}

func TestE2E_QueueFull(t *testing.T) {
	env := setupE2EEnvironment(t, 1) // Max capacity = 1
	defer env.tsServer.Close()
	defer env.targetServer.Close()

	// Register route directly in registry for offline scenario
	env.registry.Register("/webhook/test", env.targetPort, "offline-tunnel")

	t.Run("e2e_should_return_503_when_queue_full", func(t *testing.T) {
		// CLI offline, fill the queue
		req1, _ := http.NewRequest(http.MethodPost, env.tsServer.URL+"/webhook/test", bytes.NewReader([]byte(`1`)))
		resp1, err := http.DefaultClient.Do(req1)
		require.NoError(t, err)
		defer resp1.Body.Close()
		assert.Equal(t, http.StatusAccepted, resp1.StatusCode) // Queue has 1 item (now full)

		// Try to queue another one
		req2, _ := http.NewRequest(http.MethodPost, env.tsServer.URL+"/webhook/test", bytes.NewReader([]byte(`2`)))
		resp2, err := http.DefaultClient.Do(req2)
		require.NoError(t, err)
		defer resp2.Body.Close()

		assert.Equal(t, http.StatusServiceUnavailable, resp2.StatusCode) // 503 Full
	})
}

func TestE2E_PingAndAuth(t *testing.T) {
	env := setupE2EEnvironment(t, 20)
	defer env.tsServer.Close()
	defer env.targetServer.Close()

	t.Run("e2e_ping_should_return_pong_with_auth", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, env.tsServer.URL+"/ping", nil)
		req.Header.Set("Authorization", "Bearer e2e-token")
		
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("e2e_ping_should_reject_without_auth", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, env.tsServer.URL+"/ping", nil)
		// Missing auth header
		
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}
