package main

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"go-distributed-hashcat/internal/domain"
	"go-distributed-hashcat/internal/infrastructure"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Agent struct {
	ID           uuid.UUID
	Name         string
	ServerURL    string
	Client       *http.Client
	CurrentJob   *domain.Job
	UploadDir    string
	LocalFiles   map[string]LocalFile // filename -> LocalFile
	AgentKey     string               // Add agent key field
	OriginalPort int                  // Store original port from database
	ServerIP     string               // Store server IP for validation
	Status       string               // Current agent status (online, offline, busy)
}

type LocalFile struct {
	Name    string    `json:"name"`
	Path    string    `json:"path"`
	Size    int64     `json:"size"`
	Type    string    `json:"type"` // wordlist, hash_file
	Hash    string    `json:"hash"` // MD5 hash for integrity
	ModTime time.Time `json:"mod_time"`
}

type AgentInfo struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	IPAddress    string    `json:"ip_address"`
	Port         int       `json:"port"`
	Capabilities string    `json:"capabilities"`
	AgentKey     string    `json:"agent_key"`
}

func main() {
	var rootCmd = &cobra.Command{
		Use:   "agent",
		Short: "Hashcat distributed cracking agent",
		Run:   runAgent,
	}

	rootCmd.Flags().String("server", "http://localhost:1337", "Server URL")
	rootCmd.Flags().String("name", "", "Agent name")
	rootCmd.Flags().String("ip", "", "Agent IP address")
	rootCmd.Flags().Int("port", 8081, "Agent port")
	rootCmd.Flags().String("capabilities", "auto", "Agent capabilities (auto, CPU, GPU, or custom)")
	rootCmd.Flags().String("agent-key", "", "Agent key")
	rootCmd.Flags().String("upload-dir", "/root/uploads", "Local uploads directory")

	viper.BindPFlags(rootCmd.Flags())

	if err := rootCmd.Execute(); err != nil {
		infrastructure.AgentLogger.Fatal("%v", err)
	}
}

func runAgent(cmd *cobra.Command, args []string) {
	serverURL := viper.GetString("server")
	name := viper.GetString("name")
	ip := viper.GetString("ip")
	port := viper.GetInt("port")
	capabilities := viper.GetString("capabilities")
	agentKey := viper.GetString("agent-key")
	uploadDir := viper.GetString("upload-dir")

	if agentKey == "" {
		infrastructure.AgentLogger.Fatal("Agent key is required. Please provide --agent-key parameter.")
	}

	// Create temporary agent client to check agent key
	tempAgent := &Agent{
		ServerURL: serverURL,
		Client:    &http.Client{Timeout: 30 * time.Second},
	}

	// Check if agent key exists in database
	info, lookupErr := getAgentByKeyOnly(tempAgent, agentKey)
	if lookupErr != nil {
		infrastructure.AgentLogger.Fatal("Agent key '%s' not registered in the database. Agent failed to run.", agentKey)
	}

	// Validate IP address with local IP
	if ip != "" {
		if err := validateLocalIP(ip); err != nil {
			infrastructure.AgentLogger.Fatal("%v", err)
		}
	} else {
		// If IP is empty, get automatically
		ip = getLocalIP()
		infrastructure.AgentLogger.Info("Auto-detected local IP: %s", ip)
	}

	// Auto-detect capabilities using hashcat -I if not specified or empty
	if capabilities == "" || capabilities == "auto" {
		infrastructure.AgentLogger.Info("Auto-detection mode: Running hashcat -I to detect capabilities...")
		capabilities = detectCapabilitiesWithHashcat()
		infrastructure.AgentLogger.Success("Auto-detected capabilities using hashcat -I: %s", capabilities)
	} else {
		infrastructure.AgentLogger.Info("Using manually specified capabilities: %s", capabilities)
	}

	// Update capabilities in database if different from detected
	if info.Capabilities == "" || info.Capabilities != capabilities {
		infrastructure.AgentLogger.Info("Updating capabilities from '%s' to '%s'", info.Capabilities, capabilities)
		if err := updateAgentCapabilities(tempAgent, agentKey, capabilities); err != nil {
			infrastructure.AgentLogger.Warning("Failed to update capabilities: %v", err)
		} else {
			infrastructure.AgentLogger.Success("Capabilities updated successfully")
		}
	} else {
		infrastructure.AgentLogger.Info("Capabilities already up-to-date: %s", capabilities)
	}

	// If name is empty, use hostname
	if name == "" {
		hostname, _ := os.Hostname()
		name = fmt.Sprintf("agent-%s", hostname)
	}

	// Save original port from database for restoration
	originalPort := info.Port
	if originalPort == 0 {
		originalPort = 8080 // Default port
	}

	// Create the actual agent object
	agent := &Agent{
		ID:           info.ID,
		Name:         name,
		ServerURL:    serverURL,
		Client:       &http.Client{Timeout: 30 * time.Second},
		UploadDir:    uploadDir,
		LocalFiles:   make(map[string]LocalFile),
		AgentKey:     agentKey,
		OriginalPort: originalPort, // Store original port from database
		ServerIP:     ip,           // Store server IP for validation
	}

	// Inisialisasi direktori
	if err := agent.initializeDirectories(); err != nil {
		infrastructure.AgentLogger.Fatal("Failed to initialize directories: %v", err)
	}

	if err := agent.scanLocalFiles(); err != nil {
		infrastructure.AgentLogger.Warning("Failed to scan local files: %v", err)
	}

	// Registrasi ke server
	err := agent.registerWithServer(name, ip, port, capabilities, agentKey)
	if err != nil && strings.Contains(err.Error(), "already registered") {
		if info.Name != name {
			infrastructure.AgentLogger.Fatal("Agent key '%s' already used by another agent: %s", agentKey, info.Name)
		}

		// If IP, Port, or Capabilities are empty → update data
		if info.IPAddress == "" || info.Port == 0 || info.Capabilities == "" {
			infrastructure.AgentLogger.Info("Agent data '%s' is incomplete, being updated...", name)
			if err := agent.updateAgentInfo(info.ID, ip, port, capabilities, "online"); err != nil {
				infrastructure.AgentLogger.Fatal("Failed to update agent info: %v", err)
			}
			// Still use "registered successfully" log
			infrastructure.AgentLogger.Success("Agent %s (%s) registered successfully", agent.Name, agent.ID.String())
			agent.updateStatus("online")
		} else {
			// Complete data → log already exists with its data
			infrastructure.AgentLogger.Info("Agent key already exists with complete data:")
			infrastructure.AgentLogger.Info("Name: %s", agent.Name)
			infrastructure.AgentLogger.Info("    ID: %s", info.ID.String())
			infrastructure.AgentLogger.Info("Server URL: %s", agent.ServerURL)
			infrastructure.AgentLogger.Info("Port: %d", info.Port)
			infrastructure.AgentLogger.Info("Capabilities: %s", info.Capabilities)
			infrastructure.AgentLogger.Success("Agent %s (%s) is running", agent.Name, agent.ID.String())
			agent.updateStatus("online")
		}
	} else if err != nil {
		infrastructure.AgentLogger.Fatal("Failed to register and lookup agent: %v", err)
	} else {
		// New registration successful
		infrastructure.AgentLogger.Success("Agent %s (%s) registered successfully", agent.Name, agent.ID.String())
	}

	// Update status to online and port to 8081 when agent starts running
	infrastructure.AgentLogger.Info("Updating agent status to online and port to 8081...")
	if err := agent.updateAgentInfo(agent.ID, ip, 8081, capabilities, "online"); err != nil {
		infrastructure.AgentLogger.Warning("Failed to update agent status to online: %v", err)
	} else {
		infrastructure.AgentLogger.Success("Agent status updated to online with port 8081")
		// Set agent status for real-time monitoring
		agent.Status = "online"
	}

	// Run hashcat benchmark to detect and update agent speed
	// This should run after status is set to online to ensure proper speed update
	infrastructure.AgentLogger.Info("Running hashcat benchmark to detect agent speed...")
	if err := agent.runHashcatBenchmark(); err != nil {
		infrastructure.AgentLogger.Warning("Failed to run hashcat benchmark: %v", err)
		infrastructure.AgentLogger.Info("Agent will continue without speed information")
	} else {
		infrastructure.AgentLogger.Success("Hashcat benchmark completed and speed updated")
	}

	infrastructure.AgentLogger.Info("Upload Directory: %s", agent.UploadDir)
	infrastructure.AgentLogger.Info("Found %d local files", len(agent.LocalFiles))

	if err := agent.registerLocalFiles(); err != nil {
		infrastructure.AgentLogger.Warning("Failed to register local files: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start real-time speed monitoring in background
	agent.startRealTimeSpeedMonitoring(ctx)

	go agent.startHeartbeat(ctx)
	go agent.pollForJobs(ctx)
	go agent.watchLocalFiles(ctx)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	infrastructure.AgentLogger.Info("Shutting down agent...")

	// Update status to offline and restore original port 8080 before shutdown
	infrastructure.AgentLogger.Info("Updating agent status to offline and restoring port to 8080...")
	infrastructure.AgentLogger.Info("Preserving capabilities: %s", capabilities)

	// Set agent status for real-time monitoring
	agent.Status = "offline"

	// Note: Speed is no longer reset to 0 during shutdown to preserve speed data
	// Speed will be updated when agent comes back online

	// Update agent data (IP, port, capabilities) first
	if err := agent.updateAgentInfo(agent.ID, ip, 8080, capabilities, ""); err != nil {
		infrastructure.AgentLogger.Warning("Failed to update agent data: %v", err)
	} else {
		infrastructure.AgentLogger.Success("Agent data updated successfully")
	}

	// Update status to offline without resetting speed using the new endpoint
	if err := agent.updateAgentStatusOffline(); err != nil {
		infrastructure.AgentLogger.Warning("Failed to update agent status to offline: %v", err)
	} else {
		infrastructure.AgentLogger.Success("Agent status updated to offline (speed preserved)")
	}

	infrastructure.AgentLogger.Info("Agent exited")
}

func getAgentByKeyOnly(a *Agent, key string) (AgentInfo, error) {
	var info AgentInfo
	url := fmt.Sprintf("%s/api/v1/agents?agent_key=%s", a.ServerURL, key)
	resp, err := a.Client.Get(url)
	if err != nil {
		return info, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return info, fmt.Errorf("failed to get agent: %s", string(body))
	}

	var res struct {
		Data []AgentInfo `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return info, err
	}

	if len(res.Data) == 0 {
		return info, fmt.Errorf("agent key not found")
	}

	return res.Data[0], nil
}

func (a *Agent) updateAgentInfo(agentID uuid.UUID, ip string, port int, capabilities string, status string) error {
	// Use the correct endpoint for updating agent data
	req := struct {
		AgentKey     string `json:"agent_key"`
		IPAddress    string `json:"ip_address"`
		Port         int    `json:"port"`
		Capabilities string `json:"capabilities"`
	}{
		AgentKey:     a.AgentKey,
		IPAddress:    ip,
		Port:         port,
		Capabilities: capabilities,
	}

	jsonData, _ := json.Marshal(req)
	url := fmt.Sprintf("%s/api/v1/agents/update-data", a.ServerURL)

	httpReq, _ := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonData))
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := a.Client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to update agent data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to update agent data: %s", string(body))
	}

	// If status needs to be updated, use the status endpoint
	if status != "" {
		statusReq := struct {
			Status string `json:"status"`
		}{
			Status: status,
		}

		statusData, _ := json.Marshal(statusReq)
		statusURL := fmt.Sprintf("%s/api/v1/agents/%s/status", a.ServerURL, agentID.String())

		statusHttpReq, _ := http.NewRequest(http.MethodPut, statusURL, bytes.NewBuffer(statusData))
		statusHttpReq.Header.Set("Content-Type", "application/json")

		statusResp, err := a.Client.Do(statusHttpReq)
		if err != nil {
			return fmt.Errorf("failed to update agent status: %w", err)
		}
		defer statusResp.Body.Close()

		if statusResp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(statusResp.Body)
			return fmt.Errorf("failed to update agent status: %s", string(body))
		}
	}

	return nil
}

func (a *Agent) initializeDirectories() error {
	dirs := []string{
		a.UploadDir,
		filepath.Join(a.UploadDir, "wordlists"),
		filepath.Join(a.UploadDir, "hash-files"),
		filepath.Join(a.UploadDir, "temp"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	infrastructure.AgentLogger.Info("Initialized directory structure in %s", a.UploadDir)
	return nil
}

func (a *Agent) scanLocalFiles() error {
	infrastructure.AgentLogger.Info("Scanning local files...")

	// Scan wordlists
	wordlistDir := filepath.Join(a.UploadDir, "wordlists")
	if err := a.scanDirectory(wordlistDir, "wordlist"); err != nil {
		infrastructure.AgentLogger.Warning("Failed to scan wordlists directory: %v", err)
	}

	// Scan hash files
	hashFileDir := filepath.Join(a.UploadDir, "hash-files")
	if err := a.scanDirectory(hashFileDir, "hash_file"); err != nil {
		infrastructure.AgentLogger.Warning("Failed to scan hash-files directory: %v", err)
	}

	// Also scan root upload directory for legacy files
	if err := a.scanDirectory(a.UploadDir, "auto"); err != nil {
		infrastructure.AgentLogger.Warning("Failed to scan root upload directory: %v", err)
	}

	infrastructure.AgentLogger.Info("Scanned %d local files", len(a.LocalFiles))
	for filename, file := range a.LocalFiles {
		infrastructure.AgentLogger.Info("  %s (%s, %s)", filename, file.Type, formatFileSize(file.Size))
	}

	return nil
}

func (a *Agent) scanDirectory(dir, fileType string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil // Directory doesn't exist, skip
	}

	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Skip hidden files and temp files
		if strings.HasPrefix(info.Name(), ".") || strings.HasPrefix(info.Name(), "~") {
			return nil
		}

		// Determine file type if auto
		detectedType := fileType
		if fileType == "auto" {
			detectedType = a.detectFileType(info.Name())
		}

		// Calculate file hash for integrity
		hash, err := a.calculateFileHash(path)
		if err != nil {
			infrastructure.AgentLogger.Warning("Failed to calculate hash for %s: %v", path, err)
			hash = ""
		}

		localFile := LocalFile{
			Name:    info.Name(),
			Path:    path,
			Size:    info.Size(),
			Type:    detectedType,
			Hash:    hash,
			ModTime: info.ModTime(),
		}

		a.LocalFiles[info.Name()] = localFile
		return nil
	})
}

func (a *Agent) detectFileType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	name := strings.ToLower(filename)

	// Hash file extensions
	if ext == ".hccapx" || ext == ".hccap" || ext == ".cap" || ext == ".pcap" {
		return "hash_file"
	}

	// Wordlist extensions and common names
	if ext == ".txt" || ext == ".lst" || ext == ".dic" || ext == ".wordlist" {
		return "wordlist"
	}

	// Common wordlist names
	if strings.Contains(name, "rockyou") || strings.Contains(name, "password") ||
		strings.Contains(name, "wordlist") || strings.Contains(name, "dict") {
		return "wordlist"
	}

	return "unknown"
}

func (a *Agent) calculateFileHash(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func (a *Agent) watchLocalFiles(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute) // Check every 5 minutes
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			oldCount := len(a.LocalFiles)
			if err := a.scanLocalFiles(); err != nil {
				infrastructure.AgentLogger.Error("Error rescanning local files: %v", err)
				continue
			}

			newCount := len(a.LocalFiles)
			if newCount != oldCount {
				infrastructure.AgentLogger.Info("Local files changed: %d -> %d", oldCount, newCount)
				if err := a.registerLocalFiles(); err != nil {
					infrastructure.AgentLogger.Error("Error re-registering local files: %v", err)
				}
			}
		}
	}
}

func (a *Agent) registerLocalFiles() error {
	if len(a.LocalFiles) == 0 {
		return nil
	}

	infrastructure.AgentLogger.Info("Registering local files with server...")

	req := struct {
		AgentID uuid.UUID            `json:"agent_id"`
		Files   map[string]LocalFile `json:"files"`
	}{
		AgentID: a.ID,
		Files:   a.LocalFiles,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/api/v1/agents/%s/files", a.ServerURL, a.ID.String())
	resp, err := a.Client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to register local files: %s", string(body))
	}

	infrastructure.AgentLogger.Success("Registered %d local files with server", len(a.LocalFiles))
	return nil
}

func (a *Agent) findLocalFile(filename string) (string, bool) {
	// Direct filename lookup
	if file, exists := a.LocalFiles[filename]; exists {
		return file.Path, true
	}

	// Search by pattern (case-insensitive)
	lowerFilename := strings.ToLower(filename)
	for name, file := range a.LocalFiles {
		if strings.ToLower(name) == lowerFilename {
			return file.Path, true
		}
	}

	// Search in subdirectories
	for _, subdir := range []string{"wordlists", "hash-files"} {
		fullPath := filepath.Join(a.UploadDir, subdir, filename)
		if _, err := os.Stat(fullPath); err == nil {
			return fullPath, true
		}
	}

	return "", false
}

func (a *Agent) registerWithServer(name, ip string, port int, capabilities, agentKey string) error {
	req := domain.CreateAgentRequest{
		Name:         name,
		IPAddress:    ip,
		Port:         port,
		Capabilities: capabilities,
		AgentKey:     agentKey, // ← kirim agentKey ke server
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return err
	}

	resp, err := a.Client.Post(
		fmt.Sprintf("%s/api/v1/agents", a.ServerURL),
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to register agent: %s", string(body))
	}

	var response struct {
		Data domain.Agent `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return err
	}

	a.ID = response.Data.ID
	return nil
}

func (a *Agent) startHeartbeat(ctx context.Context) {
	// Ultra-fast real-time heartbeat: every 1 second for instant detection
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	// Send initial heartbeat immediately
	if err := a.sendHeartbeat(); err != nil {
		infrastructure.AgentLogger.Warning("Initial heartbeat failed: %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := a.sendHeartbeat(); err != nil {
				infrastructure.AgentLogger.Error("Failed to send heartbeat: %v", err)
			}
		}
	}
}

func (a *Agent) sendHeartbeat() error {
	// Use new endpoint with agent key instead of agent ID
	url := fmt.Sprintf("%s/api/v1/agents/heartbeat", a.ServerURL)

	// Create request body with agent key
	reqBody := struct {
		AgentKey string `json:"agent_key"`
	}{
		AgentKey: a.AgentKey,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal heartbeat request: %v", err)
	}

	resp, err := a.Client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("heartbeat failed: %s", string(body))
	}

	return nil
}

func (a *Agent) updateStatus(status string) {
	if a.ID == uuid.Nil {
		infrastructure.AgentLogger.Warning("Agent ID not yet available, cannot update status")
		return
	}

	req := struct {
		Status string `json:"status"`
	}{
		Status: status,
	}

	jsonData, _ := json.Marshal(req)
	url := fmt.Sprintf("%s/api/v1/agents/%s", a.ServerURL, a.ID.String())

	httpReq, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(jsonData))
	if err != nil {
		infrastructure.AgentLogger.Error("Failed to create status update request: %v", err)
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := a.Client.Do(httpReq)
	if err != nil {
		infrastructure.AgentLogger.Error("Failed to update agent status: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		infrastructure.AgentLogger.Error("Failed to update agent status: %s", string(body))
		return
	}

	infrastructure.AgentLogger.Success("Agent '%s' status successfully updated to '%s'", a.Name, status)
}

func (a *Agent) pollForJobs(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if a.CurrentJob == nil {
				if err := a.checkForNewJob(); err != nil {
					infrastructure.AgentLogger.Error("Error checking for new job: %v", err)
				}
			}
		}
	}
}

func (a *Agent) checkForNewJob() error {
	// Use the specific endpoint for getting available job for this agent
	url := fmt.Sprintf("%s/api/v1/jobs/agent/%s", a.ServerURL, a.ID.String())
	resp, err := a.Client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var response struct {
		Data *domain.Job `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return err
	}

	// Check if we got a job
	if response.Data != nil {
		infrastructure.AgentLogger.Info("Found assigned job: %s", response.Data.Name)
		a.CurrentJob = response.Data
		go a.executeJob(response.Data)
	}

	return nil
}

func (a *Agent) executeJob(job *domain.Job) {
	defer func() {
		a.CurrentJob = nil
		a.updateStatus("online")
	}()

	infrastructure.AgentLogger.Info("Starting job: %s", job.Name)
	a.updateStatus("busy")

	// Start the job
	if err := a.startJob(job.ID); err != nil {
		infrastructure.AgentLogger.Error("Failed to start job: %v", err)
		a.failJob(job.ID, fmt.Sprintf("Failed to start job: %v", err))
		return
	}

	// Execute hashcat command
	if err := a.runHashcat(job); err != nil {
		infrastructure.AgentLogger.Error("Hashcat execution failed: %v", err)
		a.failJob(job.ID, fmt.Sprintf("Hashcat execution failed: %v", err))
		return
	}

	infrastructure.AgentLogger.Success("Job completed: %s", job.Name)
}

func (a *Agent) startJob(jobID uuid.UUID) error {
	url := fmt.Sprintf("%s/api/v1/jobs/%s/start", a.ServerURL, jobID.String())
	resp, err := a.Client.Post(url, "application/json", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func (a *Agent) runHashcat(job *domain.Job) error {
	// Send initial job data to server immediately
	a.sendInitialJobData(job)

	// Resolve hash file (local first, download if needed)
	var localHashFile string
	var err error

	if job.HashFileID != nil {
		// Try to find hash file locally first
		hashFileName := job.HashFile
		if hashFileName != "" {
			if localPath, found := a.findLocalFile(filepath.Base(hashFileName)); found {
				localHashFile = localPath
				infrastructure.AgentLogger.Info("Using local hash file: %s", localHashFile)
			}
		}

		// Download if not found locally
		if localHashFile == "" {
			localHashFile, err = a.downloadHashFile(*job.HashFileID)
			if err != nil {
				return fmt.Errorf("failed to download hash file: %w", err)
			}
			// defer os.Remove(localHashFile) // Keep file for debugging/--show command
			infrastructure.AgentLogger.Success("Downloaded hash file: %s", localHashFile)
		}
	} else {
		// Fallback to original path
		localHashFile = job.HashFile
	}

	// Resolve wordlist (local first, download if needed, or create from content)
	var localWordlist string

	// Prioritize WordlistID if available
	if job.WordlistID != nil {
		downloadedPath, err := a.downloadWordlist(*job.WordlistID)
		if err != nil {
			return fmt.Errorf("failed to download wordlist %s: %w", job.WordlistID.String(), err)
		}
		localWordlist = downloadedPath
		infrastructure.AgentLogger.Success("Downloaded wordlist from ID: %s", localWordlist)
	} else if job.Wordlist != "" {
		// Check if wordlist contains newlines (indicating it's content, not a path)
		if strings.Contains(job.Wordlist, "\n") {
			// This is wordlist content, create a temporary file
			tempDir := filepath.Join(a.UploadDir, "temp")
			if err := os.MkdirAll(tempDir, 0755); err != nil {
				return fmt.Errorf("failed to create temp directory: %w", err)
			}

			wordlistFile := filepath.Join(tempDir, fmt.Sprintf("wordlist-%s.txt", job.ID.String()))
			if err := os.WriteFile(wordlistFile, []byte(job.Wordlist), 0644); err != nil {
				return fmt.Errorf("failed to create wordlist file: %w", err)
			}

			localWordlist = wordlistFile
			infrastructure.AgentLogger.Info("Created wordlist file from content: %s", localWordlist)
			infrastructure.AgentLogger.Info("Wordlist content preview: %s", strings.Split(job.Wordlist, "\n")[0])
		} else {
			// Fallback to wordlist filename resolution
			if localPath, found := a.findLocalFile(job.Wordlist); found {
				localWordlist = localPath
				infrastructure.AgentLogger.Info("Using local wordlist: %s", localWordlist)
			} else {
				// Try to parse as UUID and download
				if wordlistUUID, err := uuid.Parse(job.Wordlist); err == nil {
					downloadedPath, err := a.downloadWordlist(wordlistUUID)
					if err != nil {
						return fmt.Errorf("failed to download wordlist %s: %w", job.Wordlist, err)
					}
					localWordlist = downloadedPath
					infrastructure.AgentLogger.Success("Downloaded wordlist: %s", localWordlist)
				} else {
					// If not UUID, use as direct path
					localWordlist = job.Wordlist
					infrastructure.AgentLogger.Info("Using wordlist path directly: %s", localWordlist)
				}
			}
		}
	}

	// Build hashcat command with UUID-based outfile
	tempDir := filepath.Join(a.UploadDir, "temp")
	outfile := filepath.Join(tempDir, fmt.Sprintf("cracked-%s.txt", job.ID.String()))
	infrastructure.AgentLogger.Info("Outfile will be: %s", outfile)
	args := []string{
		"-m", strconv.Itoa(job.HashType),
		"-a", strconv.Itoa(job.AttackMode),
		localHashFile,
		localWordlist,
		"-w", "4",
		"--status",
		"--status-timer=2",
		"--potfile-disable",
		"--outfile", outfile,
		"--outfile-format", "2", // Format: hash:plain
	}

	// Add skip and limit parameters for distributed cracking
	if job.Skip != nil && *job.Skip >= 0 {
		args = append(args, "--skip", strconv.FormatInt(*job.Skip, 10))
		infrastructure.AgentLogger.Info("Using --skip parameter: %d", *job.Skip)
	}

	if job.WordLimit != nil && *job.WordLimit > 0 {
		args = append(args, "--limit", strconv.FormatInt(*job.WordLimit, 10))
		infrastructure.AgentLogger.Info("Using --limit parameter: %d", *job.WordLimit)
	}

	if job.Rules != "" {
		args = append(args, "-r", job.Rules)
	}

	infrastructure.AgentLogger.Info("Running hashcat with args: %v", args)

	cmd := exec.Command("hashcat", args...)

	// Set up pipes for stdout and stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return err
	}

	// Monitor output for progress updates
	go a.monitorHashcatOutput(job, stdout, stderr)

	// Monitor job status for cancellation/pause
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go a.monitorJobStatus(ctx, job.ID, cmd)

	// Wait for command to complete
	if err := cmd.Wait(); err != nil {
		// Check if hashcat found the password (exit code 0) or exhausted (exit code 1)
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode := exitError.ExitCode()
			switch exitCode {
			case 1:
				// Exhausted - not an error
				a.completeJob(job.ID, "Password not found - exhausted")
				a.cleanupJobFiles(job.ID)
				return nil
			case 255:
				// Exit code 255 usually means invalid arguments or file not found
				// Check if this is due to password not being found vs other errors
				// For now, treat exit 255 as password not found scenario
				a.failJob(job.ID, "Password not found")
				a.cleanupJobFiles(job.ID)
				return nil
			}
		}
		// Cleanup on other errors too
		a.cleanupJobFiles(job.ID)
		return err
	}

	// Success - password found, now capture the actual password
	password, err := a.extractPassword(job.ID)
	if err != nil {
		infrastructure.AgentLogger.Warning("Failed to extract password: %v", err)
		a.completeJob(job.ID, "Password found (extraction failed)")
	} else {
		a.completeJob(job.ID, fmt.Sprintf("Password found: %s", password))
	}

	// Cleanup outfile after job completion
	a.cleanupJobFiles(job.ID)
	return nil
}

func (a *Agent) downloadHashFile(hashFileID uuid.UUID) (string, error) {
	// Create temp directory for downloaded files
	tempDir := filepath.Join(a.UploadDir, "temp")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Download file from server
	url := fmt.Sprintf("%s/api/v1/hashfiles/%s/download", a.ServerURL, hashFileID.String())
	resp, err := a.Client.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download file: status %d", resp.StatusCode)
	}

	// Get filename from Content-Disposition header or use UUID
	filename := hashFileID.String()
	if cd := resp.Header.Get("Content-Disposition"); cd != "" {
		// Extract filename from Content-Disposition header
		if parts := regexp.MustCompile(`filename=([^;]+)`).FindStringSubmatch(cd); len(parts) > 1 {
			filename = parts[1]
		}
	}

	// Create local file
	localPath := filepath.Join(tempDir, filename)
	file, err := os.Create(localPath)
	if err != nil {
		return "", fmt.Errorf("failed to create local file: %w", err)
	}
	defer file.Close()

	// Copy downloaded content to local file
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		os.Remove(localPath) // Clean up on error
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	infrastructure.AgentLogger.Success("Downloaded hash file to: %s", localPath)
	return localPath, nil
}

func (a *Agent) downloadWordlist(wordlistID uuid.UUID) (string, error) {
	// Create temp directory for downloaded files
	tempDir := filepath.Join(a.UploadDir, "temp")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Download file from server
	url := fmt.Sprintf("%s/api/v1/wordlists/%s/download", a.ServerURL, wordlistID.String())
	resp, err := a.Client.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download file: status %d", resp.StatusCode)
	}

	// Get filename from Content-Disposition header or use UUID
	filename := wordlistID.String()
	if cd := resp.Header.Get("Content-Disposition"); cd != "" {
		// Extract filename from Content-Disposition header
		if parts := regexp.MustCompile(`filename=([^;]+)`).FindStringSubmatch(cd); len(parts) > 1 {
			filename = parts[1]
		}
	}

	// Create local file
	localPath := filepath.Join(tempDir, filename)
	file, err := os.Create(localPath)
	if err != nil {
		return "", fmt.Errorf("failed to create local file: %w", err)
	}
	defer file.Close()

	// Copy downloaded content to local file
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		os.Remove(localPath) // Clean up on error
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	infrastructure.AgentLogger.Success("Downloaded wordlist to: %s", localPath)
	return localPath, nil
}

func (a *Agent) extractPassword(jobID uuid.UUID) (string, error) {
	// Read password from outfile created during cracking
	tempDir := filepath.Join(a.UploadDir, "temp")
	outfile := filepath.Join(tempDir, fmt.Sprintf("cracked-%s.txt", jobID.String()))

	content, err := os.ReadFile(outfile)
	if err != nil {
		return "", fmt.Errorf("failed to read outfile %s: %w", outfile, err)
	}

	// Parse output format: plain password (one per line)
	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			// Return first non-empty password found
			return line, nil
		}
	}

	return "", fmt.Errorf("no password found in outfile")
}

func (a *Agent) cleanupJobFiles(jobID uuid.UUID) {
	// Clean up job-specific files after completion
	tempDir := filepath.Join(a.UploadDir, "temp")
	outfile := filepath.Join(tempDir, fmt.Sprintf("cracked-%s.txt", jobID.String()))

	if err := os.Remove(outfile); err != nil && !os.IsNotExist(err) {
		infrastructure.AgentLogger.Warning("Failed to cleanup outfile %s: %v", outfile, err)
	} else {
		infrastructure.AgentLogger.Info("Cleaned up outfile: %s", outfile)
	}
}

func (a *Agent) monitorHashcatOutput(job *domain.Job, stdout, stderr io.Reader) {
	// Parse hashcat output for progress updates
	progressRegex := regexp.MustCompile(`Progress\.+:\s*(\d+)/(\d+)\s*\((\d+\.\d+)%\)`)
	speedRegex := regexp.MustCompile(`Speed\.+:\s*(\d+)\s*H/s`)
	etaRegex := regexp.MustCompile(`ETA\.+:\s*(\d+):(\d+):(\d+)`)

	scanner := func(reader io.Reader) {
		buf := make([]byte, 1024)
		for {
			n, err := reader.Read(buf)
			if err != nil {
				return
			}

			output := string(buf[:n])

			// Parse progress
			if matches := progressRegex.FindStringSubmatch(output); len(matches) > 3 {
				progress, _ := strconv.ParseFloat(matches[3], 64)

				// Parse speed
				var speed int64
				if speedMatches := speedRegex.FindStringSubmatch(output); len(speedMatches) > 1 {
					speed, _ = strconv.ParseInt(speedMatches[1], 10, 64)
				}

				// Parse ETA
				var eta *string
				if etaMatches := etaRegex.FindStringSubmatch(output); len(etaMatches) > 3 {
					hours, _ := strconv.Atoi(etaMatches[1])
					minutes, _ := strconv.Atoi(etaMatches[2])
					seconds, _ := strconv.Atoi(etaMatches[3])

					etaTime := time.Now().Add(time.Duration(hours)*time.Hour +
						time.Duration(minutes)*time.Minute +
						time.Duration(seconds)*time.Second)
					etaStr := etaTime.Format(time.RFC3339)
					eta = &etaStr
				}

				// Send complete data to new endpoint
				a.updateJobDataFromAgent(job.ID, progress, speed, eta)
			}
		}
	}

	go scanner(stdout)
	go scanner(stderr)
}

func (a *Agent) sendInitialJobData(job *domain.Job) {
	req := struct {
		AgentID    string  `json:"agent_id"`
		AttackMode int     `json:"attack_mode"`
		Rules      string  `json:"rules"`
		Speed      int64   `json:"speed"`
		ETA        *string `json:"eta,omitempty"`
		Progress   float64 `json:"progress"`
	}{
		AgentID:    a.ID.String(),
		AttackMode: job.AttackMode,
		Rules:      job.Rules,
		Speed:      0,   // Initial speed is 0
		ETA:        nil, // No ETA initially
		Progress:   0,   // Initial progress is 0
	}

	jsonData, _ := json.Marshal(req)
	url := fmt.Sprintf("%s/api/v1/jobs/%s/data", a.ServerURL, job.ID.String())

	httpReq, _ := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(jsonData))
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := a.Client.Do(httpReq)
	if err != nil {
		infrastructure.AgentLogger.Error("Failed to send initial job data to server: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		infrastructure.AgentLogger.Error("Initial job data failed with status %d: %s", resp.StatusCode, string(body))
	} else {
		infrastructure.AgentLogger.Success("Initial job data sent successfully to server")
	}
}

func (a *Agent) updateJobDataFromAgent(jobID uuid.UUID, progress float64, speed int64, eta *string) {
	// Get current job data to include attack_mode and rules
	var attackMode int
	var rules string

	if a.CurrentJob != nil && a.CurrentJob.ID == jobID {
		attackMode = a.CurrentJob.AttackMode
		rules = a.CurrentJob.Rules
	}

	req := struct {
		AgentID    string  `json:"agent_id"`
		AttackMode int     `json:"attack_mode"`
		Rules      string  `json:"rules"`
		Speed      int64   `json:"speed"`
		ETA        *string `json:"eta,omitempty"`
		Progress   float64 `json:"progress"`
	}{
		AgentID:    a.ID.String(),
		AttackMode: attackMode,
		Rules:      rules,
		Speed:      speed,
		ETA:        eta,
		Progress:   progress,
	}

	jsonData, _ := json.Marshal(req)
	url := fmt.Sprintf("%s/api/v1/jobs/%s/data", a.ServerURL, jobID.String())

	httpReq, _ := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(jsonData))
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := a.Client.Do(httpReq)
	if err != nil {
		infrastructure.AgentLogger.Error("Failed to send job data update to server: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		infrastructure.AgentLogger.Error("Job data update failed with status %d: %s", resp.StatusCode, string(body))
	} else {
		infrastructure.AgentLogger.Info("Job data update sent successfully (Progress: %.2f%%, Speed: %d H/s)", progress, speed)
	}
}

func (a *Agent) monitorJobStatus(ctx context.Context, jobID uuid.UUID, cmd *exec.Cmd) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Check job status from server
			status, err := a.checkJobStatus(jobID)
			if err != nil {
				infrastructure.AgentLogger.Error("Failed to check job status: %v", err)
				continue
			}

			// Handle status changes
			switch status {
			case "paused":
				infrastructure.AgentLogger.Warning("Job %s status changed to %s, pausing hashcat", jobID, status)
				if cmd.Process != nil {
					cmd.Process.Signal(syscall.SIGSTOP)
				}
				return
			case "failed":
				infrastructure.AgentLogger.Warning("Job %s status changed to %s, terminating hashcat", jobID, status)
				if cmd.Process != nil {
					cmd.Process.Kill()
				}

				// Check if this is a coordination stop (password found by another agent)
				if a.isCoordinationStop(jobID) {
					infrastructure.AgentLogger.Info("Job stopped due to password found by another agent - updating progress to 100%%")
					// Update progress to 100% and send final status
					a.updateJobProgress(jobID, 100.0, 0)
					a.failJob(jobID, "Password found by another agent - stopping")
				}
				return
			case "cancelled":
				infrastructure.AgentLogger.Warning("Job %s status changed to %s, terminating hashcat", jobID, status)
				if cmd.Process != nil {
					cmd.Process.Kill()
				}

				// Check if this is a coordination stop (password found by another agent)
				if a.isCoordinationStop(jobID) {
					infrastructure.AgentLogger.Info("Job cancelled due to password found by another agent - updating progress to 100%%")
					// Update progress to 100% and send final status
					a.updateJobProgress(jobID, 100.0, 0)
					a.failJob(jobID, "Password found by another agent - job cancelled")
				}
				return
			}
		}
	}
}

// isCoordinationStop checks if the job was stopped due to password being found by another agent
func (a *Agent) isCoordinationStop(jobID uuid.UUID) bool {
	// Get job details to check the failure reason
	url := fmt.Sprintf("%s/api/v1/jobs/%s", a.ServerURL, jobID.String())
	resp, err := a.Client.Get(url)
	if err != nil {
		infrastructure.AgentLogger.Error("Failed to get job details for coordination check: %v", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false
	}

	var jobResp struct {
		Data struct {
			Result string `json:"result"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&jobResp); err != nil {
		infrastructure.AgentLogger.Error("Failed to decode job response: %v", err)
		return false
	}

	// Check if the failure reason indicates coordination stop
	return strings.Contains(jobResp.Data.Result, "Password found by another agent") ||
		strings.Contains(jobResp.Data.Result, "Password found by another agent - stopping") ||
		strings.Contains(jobResp.Data.Result, "Password found by another agent - job cancelled")
}

func (a *Agent) checkJobStatus(jobID uuid.UUID) (string, error) {
	url := fmt.Sprintf("%s/api/v1/jobs/%s", a.ServerURL, jobID.String())
	resp, err := a.Client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get job status: %d", resp.StatusCode)
	}

	var jobResp struct {
		Data struct {
			Status string `json:"status"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&jobResp); err != nil {
		return "", err
	}

	return jobResp.Data.Status, nil
}

func (a *Agent) completeJob(jobID uuid.UUID, result string) {
	req := struct {
		Result string `json:"result"`
	}{Result: result}

	jsonData, _ := json.Marshal(req)
	url := fmt.Sprintf("%s/api/v1/jobs/%s/complete", a.ServerURL, jobID.String())

	resp, err := a.Client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		infrastructure.AgentLogger.Error("Failed to send job completion to server: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		infrastructure.AgentLogger.Error("Job completion failed with status %d: %s", resp.StatusCode, string(body))
	} else {
		infrastructure.AgentLogger.Success("Job completion sent successfully to server")
	}
}

func (a *Agent) failJob(jobID uuid.UUID, reason string) {
	req := struct {
		Reason string `json:"reason"`
	}{Reason: reason}

	jsonData, _ := json.Marshal(req)
	url := fmt.Sprintf("%s/api/v1/jobs/%s/fail", a.ServerURL, jobID.String())

	resp, err := a.Client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		infrastructure.AgentLogger.Error("Failed to send job failure to server: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		infrastructure.AgentLogger.Error("Job failure notification failed with status %d: %s", resp.StatusCode, string(body))
	} else {
		infrastructure.AgentLogger.Success("Job failure notification sent successfully to server")
	}
}

func getLocalIP() string {
	// Use hostname -I to get local IP addresses
	cmd := exec.Command("hostname", "-I")
	output, err := cmd.Output()
	if err != nil {
		infrastructure.AgentLogger.Warning("Failed to get local IP using hostname -I: %v", err)
		return "127.0.0.1" // Fallback to localhost
	}

	// Parse output and get first non-localhost IP
	ips := strings.Fields(string(output))
	for _, ip := range ips {
		ip = strings.TrimSpace(ip)
		// Skip localhost and loopback addresses
		if ip != "127.0.0.1" && ip != "::1" && ip != "" {
			infrastructure.AgentLogger.Info("Found local IP: %s", ip)
			return ip
		}
	}

	infrastructure.AgentLogger.Warning("No valid local IP found, using fallback")
	return "127.0.0.1"
}

// validateLocalIP validates if the provided IP is a valid local IP address
func validateLocalIP(providedIP string) error {
	// Get actual local IPs using hostname -I
	cmd := exec.Command("hostname", "-I")
	output, err := cmd.Output()
	if err != nil {
		infrastructure.AgentLogger.Warning("Failed to get local IP using hostname -I: %v", err)
		// If we can't validate, allow the IP to pass
		return nil
	}

	// Parse output and check if provided IP exists in local IPs
	localIPs := strings.Fields(string(output))
	for _, localIP := range localIPs {
		localIP = strings.TrimSpace(localIP)
		if localIP == providedIP {
			infrastructure.AgentLogger.Success("IP address validation passed: %s is a valid local IP", providedIP)
			return nil
		}
	}

	// IP not found in local IPs
	return fmt.Errorf("IP address validation failed: provided IP '%s' is not a valid local IP address. Local IPs: %v", providedIP, localIPs)
}

// detectCapabilitiesWithHashcat detects server capabilities using hashcat -I command
func detectCapabilitiesWithHashcat() string {
	infrastructure.AgentLogger.Info("Starting hashcat -I capabilities detection...")

	// Check if hashcat is available
	if _, err := exec.LookPath("hashcat"); err != nil {
		infrastructure.AgentLogger.Warning("hashcat not found, falling back to basic detection")
		infrastructure.AgentLogger.Debug("Error details: %v", err)
		return detectCapabilitiesBasic()
	}

	infrastructure.AgentLogger.Info("hashcat command found, executing hashcat -I...")

	// Run hashcat -I to get device information
	cmd := exec.Command("hashcat", "-I")
	output, err := cmd.Output()
	if err != nil {
		infrastructure.AgentLogger.Warning("Failed to run hashcat -I: %v", err)
		infrastructure.AgentLogger.Info("Falling back to basic detection")
		return detectCapabilitiesBasic()
	}

	infrastructure.AgentLogger.Success("hashcat -I executed successfully")

	// Parse output to find device types
	outputStr := string(output)
	lines := strings.Split(outputStr, "\n")

	infrastructure.AgentLogger.Debug("Hashcat -I output lines count: %d", len(lines))
	infrastructure.AgentLogger.Debug("Raw output preview (first 10 lines):")
	for i, line := range lines[:min(10, len(lines))] {
		infrastructure.AgentLogger.Debug("   Line %d: %s", i+1, line)
	}

	var deviceTypes []string

	for i, line := range lines {
		line = strings.TrimSpace(line)

		// Look for device section headers
		if strings.Contains(line, "Backend Device ID #") {
			infrastructure.AgentLogger.Debug("Found device section header at line %d: %s", i+1, line)
			continue
		}

		// Look for Type line
		if strings.HasPrefix(line, "Type...........:") {
			infrastructure.AgentLogger.Debug("Found Type line at line %d: %s", i+1, line)
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				deviceType := strings.TrimSpace(parts[1])
				if deviceType != "" {
					deviceTypes = append(deviceTypes, deviceType)
					infrastructure.AgentLogger.Info("Detected device type: %s", deviceType)
				}
			}
		}
	}

	infrastructure.AgentLogger.Info("Total device types found: %d", len(deviceTypes))
	infrastructure.AgentLogger.Info("Device types: %v", deviceTypes)

	// Determine capabilities based on detected devices
	if len(deviceTypes) == 0 {
		infrastructure.AgentLogger.Warning("No device types found in hashcat -I output, falling back to basic detection")
		return detectCapabilitiesBasic()
	}

	// Check if any GPU devices are found
	for _, deviceType := range deviceTypes {
		infrastructure.AgentLogger.Debug("Checking device type for GPU: %s", deviceType)
		if strings.Contains(strings.ToUpper(deviceType), "GPU") {
			infrastructure.AgentLogger.Success("GPU device detected: %s", deviceType)
			return "GPU"
		}
	}

	// If no GPU, check for CPU
	for _, deviceType := range deviceTypes {
		infrastructure.AgentLogger.Debug("Checking device type for CPU: %s", deviceType)
		if strings.Contains(strings.ToUpper(deviceType), "CPU") {
			infrastructure.AgentLogger.Info("CPU device detected: %s", deviceType)
			return "CPU"
		}
	}

	// If we can't determine, log all found types and fallback
	infrastructure.AgentLogger.Warning("Could not determine capabilities from device types: %v", deviceTypes)
	infrastructure.AgentLogger.Info("Falling back to basic detection")
	return detectCapabilitiesBasic()
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// detectCapabilitiesBasic is the fallback detection method
func detectCapabilitiesBasic() string {
	// Try to detect GPU first
	if hasGPU() {
		return "GPU"
	}

	// Fallback to CPU
	return "CPU"
}

// hasGPU checks if GPU is available on the system
func hasGPU() bool {
	infrastructure.AgentLogger.Info("Starting GPU detection...")

	// Check for NVIDIA GPU
	if _, err := exec.LookPath("nvidia-smi"); err == nil {
		infrastructure.AgentLogger.Info("nvidia-smi command found, checking if GPU is working...")
		// Try to run nvidia-smi to verify GPU is working
		cmd := exec.Command("nvidia-smi", "--query-gpu=name", "--format=csv,noheader,nounits")
		if output, err := cmd.Output(); err == nil && len(strings.TrimSpace(string(output))) > 0 {
			gpuName := strings.TrimSpace(string(output))
			infrastructure.AgentLogger.Success("Detected NVIDIA GPU: %s", gpuName)
			return true
		} else {
			infrastructure.AgentLogger.Warning("nvidia-smi found but failed to run or no output: %v", err)
		}
	} else {
		infrastructure.AgentLogger.Debug("nvidia-smi command not found")
	}

	// Check for AMD GPU
	if _, err := exec.LookPath("rocm-smi"); err == nil {
		infrastructure.AgentLogger.Info("rocm-smi command found, checking if GPU is working...")
		cmd := exec.Command("rocm-smi", "--list-gpus")
		if output, err := cmd.Output(); err == nil && len(strings.TrimSpace(string(output))) > 0 {
			infrastructure.AgentLogger.Success("Detected AMD GPU (ROCm): %s", strings.TrimSpace(string(output)))
			return true
		} else {
			infrastructure.AgentLogger.Warning("rocm-smi found but failed to run or no output: %v", err)
		}
	} else {
		infrastructure.AgentLogger.Debug("rocm-smi command not found")
	}

	// Check for Intel GPU
	if _, err := exec.LookPath("intel_gpu_top"); err == nil {
		infrastructure.AgentLogger.Info("intel_gpu_top command found, checking if GPU is working...")
		cmd := exec.Command("intel_gpu_top", "-J", "-s", "1")
		if output, err := cmd.Output(); err == nil && len(strings.TrimSpace(string(output))) > 0 {
			infrastructure.AgentLogger.Success("Detected Intel GPU: %s", strings.TrimSpace(string(output)))
			return true
		} else {
			infrastructure.AgentLogger.Warning("intel_gpu_top found but failed to run or no output: %v", err)
		}
	} else {
		infrastructure.AgentLogger.Debug("intel_gpu_top command not found")
	}

	// Additional check: look for GPU devices in /proc
	if _, err := os.Stat("/proc/driver/nvidia"); err == nil {
		infrastructure.AgentLogger.Info("Found NVIDIA driver in /proc/driver/nvidia")
		return true
	}

	if _, err := os.Stat("/sys/class/drm"); err == nil {
		// Check if there are GPU devices
		if files, err := os.ReadDir("/sys/class/drm"); err == nil {
			for _, file := range files {
				if strings.HasPrefix(file.Name(), "card") && file.Name() != "card0" {
					infrastructure.AgentLogger.Info("Found GPU device: %s", file.Name())
					return true
				}
			}
		}
	}

	infrastructure.AgentLogger.Info("No GPU detected, using CPU")
	return false
}

// updateAgentCapabilities updates agent capabilities in the database
func updateAgentCapabilities(a *Agent, agentKey, capabilities string) error {
	req := struct {
		AgentKey     string `json:"agent_key"`
		Capabilities string `json:"capabilities"`
	}{
		AgentKey:     agentKey,
		Capabilities: capabilities,
	}

	jsonData, _ := json.Marshal(req)
	url := fmt.Sprintf("%s/api/v1/agents/update-data", a.ServerURL)

	httpReq, _ := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonData))
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := a.Client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to update capabilities: %s", string(body))
	}

	return nil
}

func formatFileSize(bytes int64) string {
	if bytes == 0 {
		return "0 B"
	}
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// updateJobProgress updates job progress and speed
func (a *Agent) updateJobProgress(jobID uuid.UUID, progress float64, speed int64) {
	req := struct {
		Progress float64 `json:"progress"`
		Speed    int64   `json:"speed"`
	}{
		Progress: progress,
		Speed:    speed,
	}

	jsonData, _ := json.Marshal(req)
	url := fmt.Sprintf("%s/api/v1/jobs/%s/progress", a.ServerURL, jobID.String())

	httpReq, _ := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(jsonData))
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := a.Client.Do(httpReq)
	if err != nil {
		infrastructure.AgentLogger.Error("Failed to send job progress update to server: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		infrastructure.AgentLogger.Error("Job progress update failed with status %d: %s", resp.StatusCode, string(body))
	} else {
		infrastructure.AgentLogger.Info("Job progress update sent successfully (Progress: %.2f%%, Speed: %d H/s)", progress, speed)
	}
}

// runHashcatBenchmark runs hashcat benchmark and updates agent speed
func (a *Agent) runHashcatBenchmark() error {
	infrastructure.AgentLogger.Info("Starting hashcat benchmark to detect agent speed...")

	// Check if hashcat is available
	if _, err := exec.LookPath("hashcat"); err != nil {
		infrastructure.AgentLogger.Warning("hashcat not found in PATH: %v", err)
		infrastructure.AgentLogger.Info("Skipping automatic speed detection. You can manually set speed via API:")
		infrastructure.AgentLogger.Info("PUT /api/v1/agents/%s/speed", a.ID.String())
		return nil
	}

	// Run hashcat -b -m 2500 for WPA benchmark
	cmd := exec.Command("hashcat", "-b", "-m", "2500")

	// Capture stdout and stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start hashcat benchmark: %w", err)
	}

	// Parse output for speed
	speedRegex := regexp.MustCompile(`Speed\.#\d+\.+:\s*(\d+)\s*H/s`)
	var detectedSpeed int64

	scanner := func(reader io.Reader) {
		buf := make([]byte, 1024)
		for {
			n, err := reader.Read(buf)
			if err != nil {
				return
			}

			output := string(buf[:n])

			// Parse speed from output
			if matches := speedRegex.FindStringSubmatch(output); len(matches) > 1 {
				if speed, err := strconv.ParseInt(matches[1], 10, 64); err == nil {
					detectedSpeed = speed
					infrastructure.AgentLogger.Info("Detected hashcat speed: %d H/s", detectedSpeed)
				}
			}
		}
	}

	go scanner(stdout)
	go scanner(stderr)

	// Wait for command to complete
	if err := cmd.Wait(); err != nil {
		infrastructure.AgentLogger.Warning("Hashcat benchmark completed with error: %v", err)
		// Continue anyway to use detected speed if available
	}

	// Update agent speed in database if speed was detected
	if detectedSpeed > 0 {
		if err := a.updateAgentSpeed(detectedSpeed); err != nil {
			infrastructure.AgentLogger.Error("Failed to update agent speed: %v", err)
			return err
		}
		infrastructure.AgentLogger.Success("Agent speed updated successfully: %d H/s", detectedSpeed)
	} else {
		infrastructure.AgentLogger.Warning("No speed detected from hashcat benchmark")
		infrastructure.AgentLogger.Info("You can manually set speed via API:")
		infrastructure.AgentLogger.Info("PUT /api/v1/agents/%s/speed", a.ID.String())
	}

	return nil
}

// updateAgentSpeed updates the agent speed in the database
// This method is called during benchmark detection and real-time monitoring
func (a *Agent) updateAgentSpeed(speed int64) error {
	req := struct {
		Speed int64 `json:"speed"`
	}{
		Speed: speed,
	}

	jsonData, _ := json.Marshal(req)
	url := fmt.Sprintf("%s/api/v1/agents/%s/speed", a.ServerURL, a.ID.String())

	httpReq, _ := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(jsonData))
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := a.Client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send speed update request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("speed update failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// updateAgentStatusOnly updates only the agent status without changing speed
// This method is used for status consistency during monitoring
func (a *Agent) updateAgentStatusOnly(status string) error {
	url := fmt.Sprintf("%s/api/v1/agents/%s/status", a.ServerURL, a.ID.String())

	req := struct {
		Status string `json:"status"`
	}{
		Status: status,
	}

	jsonData, _ := json.Marshal(req)
	httpReq, _ := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(jsonData))
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := a.Client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send status update request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("status update failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// updateAgentStatusOffline updates agent status to offline without resetting speed
// This method is used for normal shutdown scenarios to preserve speed data
func (a *Agent) updateAgentStatusOffline() error {
	url := fmt.Sprintf("%s/api/v1/agents/%s/status-offline", a.ServerURL, a.ID.String())

	infrastructure.AgentLogger.Info("Sending status offline request to: %s", url)

	httpReq, _ := http.NewRequest(http.MethodPut, url, nil)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := a.Client.Do(httpReq)
	if err != nil {
		infrastructure.AgentLogger.Error("❌ Network error during status offline update: %v", err)
		return fmt.Errorf("failed to send status offline request: %w", err)
	}
	defer resp.Body.Close()

	infrastructure.AgentLogger.Info("📡 Status offline response status: %d", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		infrastructure.AgentLogger.Error("❌ Status offline update failed with status %d: %s", resp.StatusCode, string(body))
		return fmt.Errorf("status offline update failed with status %d: %s", resp.StatusCode, string(body))
	}

	infrastructure.AgentLogger.Success("✅ Status offline request successful")
	return nil
}

// startRealTimeSpeedMonitoring starts a background goroutine for continuous speed monitoring
// This method runs independently and doesn't interfere with main agent operations
func (a *Agent) startRealTimeSpeedMonitoring(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(30 * time.Second) // Monitor every 30 seconds
		defer ticker.Stop()

		infrastructure.AgentLogger.Info("🚀 Starting real-time speed monitoring...")

		for {
			select {
			case <-ctx.Done():
				infrastructure.AgentLogger.Info("🛑 Real-time speed monitoring stopped")
				return
			case <-ticker.C:
				// Only monitor if agent is online
				if a.Status == "online" {
					// Only update status to ensure consistency, don't update speed continuously
					// Speed is only updated once during startup/benchmark
					go func() {
						if err := a.updateAgentStatusOnly("online"); err != nil {
							infrastructure.AgentLogger.Warning("Failed to update agent status during monitoring: %v", err)
						}
					}()
				}
			}
		}
	}()
}
