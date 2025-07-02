package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewAPIClient(t *testing.T) {
	baseURL := "http://example.com"
	timeout := 30 * time.Second

	client := NewAPIClient(baseURL, timeout)

	if client.baseURL != baseURL {
		t.Errorf("Expected baseURL %s, got %s", baseURL, client.baseURL)
	}

	if client.httpClient.Timeout != timeout {
		t.Errorf("Expected timeout %v, got %v", timeout, client.httpClient.Timeout)
	}
}

func TestAPIClient_Get(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse string
		serverStatus   int
		expectedBody   string
		expectedError  bool
	}{
		{
			name:           "successful GET request",
			serverResponse: `{"message": "success"}`,
			serverStatus:   http.StatusOK,
			expectedBody:   `{"message": "success"}`,
			expectedError:  false,
		},
		{
			name:           "server error",
			serverResponse: `{"error": "internal server error"}`,
			serverStatus:   http.StatusInternalServerError,
			expectedBody:   "",
			expectedError:  true,
		},
		{
			name:           "not found error",
			serverResponse: `{"error": "not found"}`,
			serverStatus:   http.StatusNotFound,
			expectedBody:   "",
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "GET" {
					t.Errorf("Expected GET request, got %s", r.Method)
				}
				w.WriteHeader(tt.serverStatus)
				w.Write([]byte(tt.serverResponse))
			}))
			defer server.Close()

			// Create client
			client := NewAPIClient(server.URL, 5*time.Second)

			// Make request
			body, err := client.Get(context.Background(), "/test")

			// Verify results
			if tt.expectedError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				if apiErr, ok := err.(*APIError); ok {
					if apiErr.StatusCode != tt.serverStatus {
						t.Errorf("Expected status code %d, got %d", tt.serverStatus, apiErr.StatusCode)
					}
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if string(body) != tt.expectedBody {
					t.Errorf("Expected body %s, got %s", tt.expectedBody, string(body))
				}
			}
		})
	}
}

func TestAPIClient_Post(t *testing.T) {
	requestBody := map[string]string{"name": "test"}
	expectedResponse := `{"id": "123", "name": "test"}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(expectedResponse))
	}))
	defer server.Close()

	client := NewAPIClient(server.URL, 5*time.Second)

	body, err := client.Post(context.Background(), "/test", requestBody)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if string(body) != expectedResponse {
		t.Errorf("Expected body %s, got %s", expectedResponse, string(body))
	}
}

func TestAPIClient_Put(t *testing.T) {
	requestBody := map[string]string{"name": "updated"}
	expectedResponse := `{"id": "123", "name": "updated"}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("Expected PUT request, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(expectedResponse))
	}))
	defer server.Close()

	client := NewAPIClient(server.URL, 5*time.Second)

	body, err := client.Put(context.Background(), "/test", requestBody)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if string(body) != expectedResponse {
		t.Errorf("Expected body %s, got %s", expectedResponse, string(body))
	}
}

func TestAPIClient_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("Expected DELETE request, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewAPIClient(server.URL, 5*time.Second)

	body, err := client.Delete(context.Background(), "/test")

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(body) != 0 {
		t.Errorf("Expected empty body, got %s", string(body))
	}
}

func TestAPIClient_Timeout(t *testing.T) {
	// Create server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("response"))
	}))
	defer server.Close()

	// Create client with very short timeout
	client := NewAPIClient(server.URL, 10*time.Millisecond)

	_, err := client.Get(context.Background(), "/test")

	if err == nil {
		t.Error("Expected timeout error, got nil")
	}
}

func TestAPIError_Error(t *testing.T) {
	err := &APIError{
		StatusCode: 404,
		Message:    "Not Found",
		Response:   `{"error": "resource not found"}`,
	}

	expectedMessage := "API error 404: Not Found"
	if err.Error() != expectedMessage {
		t.Errorf("Expected error message %s, got %s", expectedMessage, err.Error())
	}
}