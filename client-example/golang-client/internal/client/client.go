package client

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// tokenResponse represents the OAuth token response from OpenShift
type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

// getOAuthToken obtains an OAuth token from OpenShift using username/password
// Uses the challenge-response flow similar to oc login
func getOAuthToken(server, username, password string) (string, error) {
	// Create HTTP client that accepts insecure certificates
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Don't follow redirects automatically
			return http.ErrUseLastResponse
		},
	}

	// Step 1: Request authorization with challenge
	authURL := fmt.Sprintf("%s/oauth/authorize?client_id=openshift-challenging-client&response_type=token", strings.TrimSuffix(server, "/"))
	req, err := http.NewRequest("GET", authURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create auth request: %w", err)
	}

	// Use basic auth
	auth := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("X-CSRF-Token", "1")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to request authorization: %w", err)
	}
	defer resp.Body.Close()

	// Check for redirect with token in fragment
	if resp.StatusCode == http.StatusFound || resp.StatusCode == http.StatusSeeOther {
		location := resp.Header.Get("Location")
		if location != "" {
			// Parse the token from the redirect URL fragment
			parsedURL, err := url.Parse(location)
			if err == nil && parsedURL.Fragment != "" {
				values, _ := url.ParseQuery(parsedURL.Fragment)
				if token := values.Get("access_token"); token != "" {
					return token, nil
				}
			}
		}
	}

	// If challenge-response didn't work, try direct token endpoint
	tokenURL := fmt.Sprintf("%s/oauth/token", strings.TrimSuffix(server, "/"))
	data := url.Values{}
	data.Set("grant_type", "password")
	data.Set("username", username)
	data.Set("password", password)
	data.Set("client_id", "openshift-challenging-client")

	req2, err := http.NewRequest("POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create token request: %w", err)
	}

	req2.Header.Set("Authorization", "Basic "+auth)
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req2.Header.Set("Accept", "application/json")

	resp2, err := client.Do(req2)
	if err != nil {
		return "", fmt.Errorf("failed to request token: %w", err)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp2.Body)
		return "", fmt.Errorf("token request failed with status %d: %s", resp2.StatusCode, string(body))
	}

	// Parse the response
	var tokenResp tokenResponse
	if err := json.NewDecoder(resp2.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("failed to decode token response: %w", err)
	}

	if tokenResp.AccessToken == "" {
		return "", fmt.Errorf("no access token in response")
	}

	return tokenResp.AccessToken, nil
}

// tryKubeconfig attempts to load config from kubeconfig file
func tryKubeconfig() (*rest.Config, error) {
	// Try KUBECONFIG environment variable first
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		// Default to ~/.kube/config
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		kubeconfig = filepath.Join(home, ".kube", "config")
	}

	// Check if file exists
	if _, err := os.Stat(kubeconfig); os.IsNotExist(err) {
		return nil, fmt.Errorf("kubeconfig file not found: %s", kubeconfig)
	}

	// Load kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	// Set insecure if needed
	config.TLSClientConfig.Insecure = true

	return config, nil
}

// CreateClient creates a Kubernetes client using username/password authentication
// First tries to use kubeconfig if available, otherwise falls back to OAuth token
func CreateClient(server, username, password string) (*kubernetes.Clientset, error) {
	var config *rest.Config
	var err error

	// First, try to use kubeconfig (preferred method, works with oc login)
	config, err = tryKubeconfig()
	if err == nil {
		// Successfully loaded from kubeconfig
		clientset, err := kubernetes.NewForConfig(config)
		if err == nil {
			return clientset, nil
		}
		// If kubeconfig load failed, fall through to username/password
	}

	// Fall back to username/password OAuth flow
	if username == "" || password == "" {
		return nil, fmt.Errorf("kubeconfig not available and username/password not provided")
	}

	// Get an OAuth token using username/password
	token, err := getOAuthToken(server, username, password)
	if err != nil {
		return nil, fmt.Errorf("failed to obtain OAuth token: %w", err)
	}

	fmt.Printf("Bearer token (from OAuth): %s\n", token)

	// Create REST config with Bearer token authentication
	config = &rest.Config{
		Host:        server,
		BearerToken: token,
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: true, // Accept unsigned/self-signed certificates
		},
	}

	// Create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	return clientset, nil
}

// GetRESTConfig returns the REST config used by the client
// This is a helper to create dynamic clients for OpenShift-specific resources
func GetRESTConfig(server, username, password string) (*rest.Config, error) {
	var config *rest.Config
	var err error

	// First, try to use kubeconfig (preferred method, works with oc login)
	config, err = tryKubeconfig()
	if err == nil {
		return config, nil
	}

	// Fall back to username/password OAuth flow
	if username == "" || password == "" {
		return nil, fmt.Errorf("kubeconfig not available and username/password not provided")
	}

	// Get an OAuth token using username/password
	token, err := getOAuthToken(server, username, password)
	if err != nil {
		return nil, fmt.Errorf("failed to obtain OAuth token: %w", err)
	}

	// Create REST config with Bearer token authentication
	config = &rest.Config{
		Host:        server,
		BearerToken: token,
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: true, // Accept unsigned/self-signed certificates
		},
	}

	return config, nil
}
