package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDoRequest(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		expectedResponse := map[string]string{"foo": "bar"}
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify path
			if r.URL.Path != "/api/v1/test" {
				t.Errorf("Expected path /api/v1/test, got %s", r.URL.Path)
			}
			// Verify headers
			if r.Header.Get("Authorization") != "Bearer test-token" {
				t.Errorf("Expected Authorization header, got %s", r.Header.Get("Authorization"))
			}
			if r.Header.Get("Content-Type") != "application/json" {
				t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
			}

			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(expectedResponse)
		}))
		defer server.Close()

		c := NewClient(server.URL, "test-token")
		var respData map[string]string
		err := c.DoRequest(context.Background(), "GET", "/api/v1/test", nil, &respData)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if respData["foo"] != "bar" {
			t.Errorf("Expected foo=bar, got %s", respData["foo"])
		}
	})

	t.Run("api error handling", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/api/v1/test" {
				t.Errorf("Expected path /api/v1/test, got %s", r.URL.Path)
			}
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("invalid request"))
		}))
		defer server.Close()

		c := NewClient(server.URL, "test-token")
		err := c.DoRequest(context.Background(), "POST", "/api/v1/test", map[string]string{"in": "put"}, nil)

		if err == nil {
			t.Fatal("Expected error for 400 status, got nil")
		}
		if !strings.Contains(err.Error(), "status=400") || !strings.Contains(err.Error(), "invalid request") {
			t.Errorf("Unexpected error message: %v", err)
		}
	})
}

func TestDoMultipartRequest(t *testing.T) {
	t.Run("successful multipart request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/api/v1/upload" {
				t.Errorf("Expected path /api/v1/upload, got %s", r.URL.Path)
			}
			// Verify content type
			if !strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
				t.Errorf("Expected multipart content type, got %s", r.Header.Get("Content-Type"))
			}

			err := r.ParseMultipartForm(10 << 20)
			if err != nil {
				t.Fatalf("Error parsing multipart form: %v", err)
			}

			if r.FormValue("name") != "test-name" {
				t.Errorf("Expected name=test-name, got %s", r.FormValue("name"))
			}

			file, _, err := r.FormFile("file")
			if err != nil {
				t.Fatalf("Error getting form file: %v", err)
			}
			defer file.Close()

			content, _ := io.ReadAll(file)
			if string(content) != "hello world" {
				t.Errorf("Expected file content 'hello world', got %s", string(content))
			}

			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id": "123"}`))
		}))
		defer server.Close()

		c := NewClient(server.URL, "test-token")
		fields := map[string]string{"name": "test-name"}
		fileContent := strings.NewReader("hello world")

		var respData struct {
			ID string `json:"id"`
		}
		err := c.DoMultipartRequest(context.Background(), "POST", "/api/v1/upload", fields, "file", "test.txt", fileContent, &respData)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if respData.ID != "123" {
			t.Errorf("Expected ID 123, got %s", respData.ID)
		}
	})
}
