package main

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"log"
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
		log.Fatal(err)
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
		log.Fatalf("❌ Agent key is required. Please provide --agent-key parameter.")
	}

	// Buat temporary agent client untuk cek agent key
	tempAgent := &Agent{
		ServerURL: serverURL,
		Client:    &http.Client{Timeout: 10 * time.Minute}, // Increased timeout for large file downloads
	}

	// ✅ Cek apakah agent key ada di database
	info, lookupErr := getAgentByKeyOnly(tempAgent, agentKey)
	if lookupErr != nil {
		log.Fatalf("❌ Agent key '%s' not registered in the database. Agent failed to run.", agentKey)
	}

	// ✅ Validasi IP address dengan IP lokal
	if ip != "" {
		if err := validateLocalIP(ip); err != nil {
			log.Fatalf("%v", err)
		}
	} else {
		// Jika IP kosong, ambil otomatis
		ip = getLocalIP()
		log.Printf("🔍 Auto-detected local IP: %s", ip)
	}

	// ✅ Auto-detect capabilities menggunakan hashcat -I jika tidak dispecify atau kosong
	if capabilities == "" || capabilities == "auto" {
		log.Printf("🔍 Auto-detection mode: Running hashcat -I to detect capabilities...")
		capabilities = detectCapabilitiesWithHashcat()
		log.Printf("🔍 Auto-detected capabilities using hashcat -I: %s", capabilities)
	} else {
		log.Printf("ℹ️ Using manually specified capabilities: %s", capabilities)
	}

	// ✅ Update capabilities di database jika berbeda dengan yang terdeteksi
	if info.Capabilities == "" || info.Capabilities != capabilities {
		log.Printf("🔄 Updating capabilities from '%s' to '%s'", info.Capabilities, capabilities)
		if err := updateAgentCapabilities(tempAgent, agentKey, capabilities); err != nil {
			log.Printf("⚠️ Warning: Failed to update capabilities: %v", err)
		} else {
			log.Printf("✅ Capabilities updated successfully")
		}
	} else {
		log.Printf("ℹ️ Capabilities already up-to-date: %s", capabilities)
	}

	// Jika name kosong, pakai hostname
	if name == "" {
		hostname, _ := os.Hostname()
		name = fmt.Sprintf("agent-%s", hostname)
	}

	// Simpan port asli dari database untuk restoration
	originalPort := info.Port
	if originalPort == 0 {
		originalPort = 8080 // Default port
	}

	// Buat object agent sesungguhnya
	agent := &Agent{
		ID:           info.ID,
		Name:         name,
		ServerURL:    serverURL,
		Client:       &http.Client{Timeout: 10 * time.Minute}, // Increased timeout for large file downloads
		UploadDir:    uploadDir,
		LocalFiles:   make(map[string]LocalFile),
		AgentKey:     agentKey,
		OriginalPort: originalPort, // Store original port from database
		ServerIP:     ip,           // Store server IP for validation
	}

	// Inisialisasi direktori
	if err := agent.initializeDirectories(); err != nil {
		log.Fatalf("❌ Failed to initialize directories: %v", err)
	}

	if err := agent.scanLocalFiles(); err != nil {
		log.Printf("⚠️ Warning: Failed to scan local files: %v", err)
	}

	// Registrasi ke server
	err := agent.registerWithServer(name, ip, port, capabilities, agentKey)
	if err != nil && strings.Contains(err.Error(), "already registered") {
		if info.Name != name {
			log.Fatalf("❌ Agent key '%s' already used by another agent: %s", agentKey, info.Name)
		}

		// Jika IP, Port, atau Capabilities kosong → update data
		if info.IPAddress == "" || info.Port == 0 || info.Capabilities == "" {
			log.Printf("⚠️ Agent data '%s' is incomplete, being updated...", name)
			if err := agent.updateAgentInfo(info.ID, ip, port, capabilities, "online"); err != nil {
				log.Fatalf("❌ Failed to update agent info: %v", err)
			}
			// Tetap gunakan log "registered successfully"
			log.Printf("✅ Agent %s (%s) registered successfully", agent.Name, agent.ID.String())
			agent.updateStatus("online")
		} else {
			// Data lengkap → log already exists beserta datanya
			log.Printf("ℹ️ Agent key already exists with complete data:")
			log.Printf("    Name: %s", info.Name)
			log.Printf("    ID: %s", info.ID.String())
			log.Printf("    IP: %s", info.IPAddress)
			log.Printf("    Port: %d", info.Port)
			log.Printf("    Capabilities: %s", info.Capabilities)
			log.Printf("✅ Agent %s (%s) is running", agent.Name, agent.ID.String())
			agent.updateStatus("online")
		}
	} else if err != nil {
		log.Fatalf("❌ Failed to register and lookup agent: %v", err)
	} else {
		// Registrasi baru sukses
		log.Printf("✅ Agent %s (%s) registered successfully", agent.Name, agent.ID.String())
	}

	// ✅ Update status to online and port to 8081 when agent starts running
	log.Printf("🔄 Updating agent status to online and port to 8081...")
	if err := agent.updateAgentInfo(agent.ID, ip, 8081, capabilities, "online"); err != nil {
		log.Printf("⚠️ Warning: Failed to update agent status to online: %v", err)
	} else {
		log.Printf("✅ Agent status updated to online with port 8081")
	}

	log.Printf("Local upload directory: %s", agent.UploadDir)
	log.Printf("Found %d local files", len(agent.LocalFiles))

	if err := agent.registerLocalFiles(); err != nil {
		log.Printf("⚠️ Warning: Failed to register local files: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go agent.startHeartbeat(ctx)
	go agent.pollForJobs(ctx)
	go agent.watchLocalFiles(ctx)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down agent...")

	// ✅ Update status to offline and restore original port 8080 before shutdown
	log.Printf("🔄 Updating agent status to offline and restoring port to 8080...")
	log.Printf("🔄 Preserving capabilities: %s", capabilities)
	if err := agent.updateAgentInfo(agent.ID, ip, 8080, capabilities, "offline"); err != nil {
		log.Printf("⚠️ Warning: Failed to update agent status to offline: %v", err)
	} else {
		log.Printf("✅ Agent status updated to offline with port 8080 and capabilities preserved")
	}

	// ✅ Note: restoreOriginalPort() is no longer needed since we already updated everything above
	// The single updateAgentInfo call above handles both status and port updates
	log.Printf("ℹ️ Skipping restoreOriginalPort() to avoid capabilities override")

	log.Println("Agent exited")
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
		return info, fmt.Errorf("gagal ambil agent: %s", string(body))
	}

	var res struct {
		Data []AgentInfo `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return info, err
	}

	if len(res.Data) == 0 {
		return info, fmt.Errorf("agent key tidak ditemukan")
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

	log.Printf("📁 Initialized directory structure in %s", a.UploadDir)
	return nil
}

func (a *Agent) scanLocalFiles() error {
	log.Println("🔍 DEBUG: Starting local files scan...")
	log.Printf("🔍 DEBUG: Upload directory: %s", a.UploadDir)

	// Scan wordlists
	wordlistDir := filepath.Join(a.UploadDir, "wordlists")
	log.Printf("🔍 DEBUG: Scanning wordlists directory: %s", wordlistDir)
	if err := a.scanDirectory(wordlistDir, "wordlist"); err != nil {
		log.Printf("⚠️  WARNING: Failed to scan wordlists directory: %v", err)
	}

	// Scan hash files
	hashFileDir := filepath.Join(a.UploadDir, "hash-files")
	log.Printf("🔍 DEBUG: Scanning hash-files directory: %s", hashFileDir)
	if err := a.scanDirectory(hashFileDir, "hash_file"); err != nil {
		log.Printf("⚠️  WARNING: Failed to scan hash-files directory: %v", err)
	}

	// Also scan root upload directory for legacy files
	log.Printf("🔍 DEBUG: Scanning root upload directory: %s", a.UploadDir)
	if err := a.scanDirectory(a.UploadDir, "auto"); err != nil {
		log.Printf("⚠️  WARNING: Failed to scan root upload directory: %v", err)
	}

	log.Printf("✅ SUCCESS: Scanned %d local files", len(a.LocalFiles))
	log.Printf("🔍 DEBUG: Detailed file list:")
	for filename, file := range a.LocalFiles {
		log.Printf("  📄 %s -> %s (%s, %s, Hash: %s)", 
			filename, file.Path, file.Type, formatFileSize(file.Size), file.Hash)
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
			log.Printf("Warning: Failed to calculate hash for %s: %v", path, err)
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
				log.Printf("Error rescanning local files: %v", err)
				continue
			}

			newCount := len(a.LocalFiles)
			if newCount != oldCount {
				log.Printf("📁 Local files changed: %d -> %d", oldCount, newCount)
				if err := a.registerLocalFiles(); err != nil {
					log.Printf("Error re-registering local files: %v", err)
				}
			}
		}
	}
}

func (a *Agent) registerLocalFiles() error {
	if len(a.LocalFiles) == 0 {
		return nil
	}

	log.Println("📤 Registering local files with server...")

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

	log.Printf("✅ Registered %d local files with server", len(a.LocalFiles))
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
	// ✅ Ultra-fast real-time heartbeat: every 1 second for instant detection
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	// Send initial heartbeat immediately
	if err := a.sendHeartbeat(); err != nil {
		log.Printf("Initial heartbeat failed: %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := a.sendHeartbeat(); err != nil {
				log.Printf("Failed to send heartbeat: %v", err)
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
		log.Printf("⚠️ Agent ID belum tersedia, tidak bisa update status")
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
		log.Printf("⚠️ Gagal membuat request update status: %v", err)
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := a.Client.Do(httpReq)
	if err != nil {
		log.Printf("⚠️ Gagal update status agent: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("⚠️ Gagal update status agent: %s", string(body))
		return
	}

	log.Printf("✅ Status agent '%s' berhasil diupdate menjadi '%s'", a.Name, status)
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
					log.Printf("Error checking for new job: %v", err)
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
		log.Printf("Found assigned job: %s", response.Data.Name)
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

	log.Printf("Starting job: %s", job.Name)
	a.updateStatus("busy")

	// Start the job
	if err := a.startJob(job.ID); err != nil {
		log.Printf("Failed to start job: %v", err)
		a.failJob(job.ID, fmt.Sprintf("Failed to start job: %v", err))
		return
	}

	// Execute hashcat command
	if err := a.runHashcat(job); err != nil {
		log.Printf("Hashcat execution failed: %v", err)
		a.failJob(job.ID, fmt.Sprintf("Hashcat execution failed: %v", err))
		return
	}

	log.Printf("Job completed: %s", job.Name)
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
	log.Printf("🚀 DEBUG: Starting hashcat job execution")
	log.Printf("🔍 DEBUG: Job details:")
	log.Printf("  📋 Job ID: %s", job.ID.String())
	log.Printf("  📋 Job Name: %s", job.Name)
	log.Printf("  📋 Hash Type: %d", job.HashType)
	log.Printf("  📋 Attack Mode: %d", job.AttackMode)
	if job.HashFileID != nil {
		log.Printf("  📋 Hash File ID: %s", job.HashFileID.String())
	}
	if job.WordlistID != nil {
		log.Printf("  📋 Wordlist ID: %s", job.WordlistID.String())
	}
	log.Printf("  📋 Hash File: %s", job.HashFile)
	log.Printf("  📋 Wordlist: %s", job.Wordlist)
	log.Printf("  📋 Rules: %s", job.Rules)
	
	// Send initial job data to server immediately
	a.sendInitialJobData(job)

	// Resolve hash file (local first, download if needed)
	var localHashFile string
	var err error

	if job.HashFileID != nil {
		// First, try to find the hash file locally by checking if we have a file with matching UUID
		hashFileIDStr := job.HashFileID.String()
		log.Printf("🔍 DEBUG: Looking for hash file with ID: %s", hashFileIDStr)
		log.Printf("🔍 DEBUG: Current LocalFiles count: %d", len(a.LocalFiles))
		
		// Debug: Show all local hash file files
		log.Printf("🔍 DEBUG: Available local hash files:")
		for filename, localFile := range a.LocalFiles {
			if localFile.Type == "hash_file" {
				log.Printf("  📄 %s -> %s (Size: %s)", filename, localFile.Path, formatFileSize(localFile.Size))
			}
		}
		
		localPath := a.findLocalHashFileByUUID(hashFileIDStr)
		
		if localPath != "" {
			localHashFile = localPath
			log.Printf("✅ SUCCESS: Using local hash file for ID %s: %s", hashFileIDStr, localPath)
		} else {
			// Not found locally, download from server
			log.Printf("⚠️  WARNING: Hash file %s not found locally, downloading from server...", hashFileIDStr)
			log.Printf("🔍 DEBUG: Will attempt download from: %s/api/v1/hashfiles/%s/download", a.ServerURL, hashFileIDStr)
			
			localHashFile, err = a.downloadHashFile(*job.HashFileID)
			if err != nil {
				log.Printf("❌ ERROR: Download failed for hash file %s: %v", hashFileIDStr, err)
				return fmt.Errorf("failed to download hash file: %w", err)
			}
			// defer os.Remove(localHashFile) // Keep file for debugging/--show command
			log.Printf("✅ SUCCESS: Downloaded hash file from ID: %s to %s", hashFileIDStr, localHashFile)
		}
	} else {
		// Fallback to original path
		localHashFile = job.HashFile
	}

	// Resolve wordlist (local first, download if needed, or create from content)
	var localWordlist string

	// Prioritize WordlistID if available
	if job.WordlistID != nil {
		// First, try to find the wordlist locally by checking if we have a file with matching UUID
		wordlistIDStr := job.WordlistID.String()
		log.Printf("🔍 DEBUG: Looking for wordlist with ID: %s", wordlistIDStr)
		log.Printf("🔍 DEBUG: Current LocalFiles count: %d", len(a.LocalFiles))
		
		// Debug: Show all local wordlist files
		log.Printf("🔍 DEBUG: Available local wordlist files:")
		for filename, localFile := range a.LocalFiles {
			if localFile.Type == "wordlist" {
				log.Printf("  📄 %s -> %s (Size: %s)", filename, localFile.Path, formatFileSize(localFile.Size))
			}
		}
		
		localPath := a.findLocalWordlistByUUID(wordlistIDStr)
		
		if localPath != "" {
			localWordlist = localPath
			log.Printf("✅ SUCCESS: Using local wordlist for ID %s: %s", wordlistIDStr, localPath)
		} else {
			// Not found locally, download from server
			log.Printf("⚠️  WARNING: Wordlist %s not found locally, downloading from server...", wordlistIDStr)
			log.Printf("🔍 DEBUG: Will attempt download from: %s/api/v1/wordlists/%s/download", a.ServerURL, wordlistIDStr)
			
			downloadedPath, err := a.downloadWordlist(*job.WordlistID)
			if err != nil {
				log.Printf("❌ ERROR: Download failed for wordlist %s: %v", wordlistIDStr, err)
				return fmt.Errorf("failed to download wordlist %s: %w", job.WordlistID.String(), err)
			}
			localWordlist = downloadedPath
			log.Printf("✅ SUCCESS: Downloaded wordlist from ID: %s to %s", wordlistIDStr, localWordlist)
		}
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
			log.Printf("📝 Created wordlist file from content: %s", localWordlist)
			log.Printf("📋 Wordlist content preview: %s", strings.Split(job.Wordlist, "\n")[0])
		} else {
			// Fallback to wordlist filename resolution
			if localPath, found := a.findLocalFile(job.Wordlist); found {
				localWordlist = localPath
				log.Printf("📁 Using local wordlist: %s", localWordlist)
			} else {
				// Try to parse as UUID and download
				if wordlistUUID, err := uuid.Parse(job.Wordlist); err == nil {
					downloadedPath, err := a.downloadWordlist(wordlistUUID)
					if err != nil {
						return fmt.Errorf("failed to download wordlist %s: %w", job.Wordlist, err)
					}
					localWordlist = downloadedPath
					log.Printf("📥 Downloaded wordlist: %s", localWordlist)
				} else {
					// If not UUID, use as direct path
					localWordlist = job.Wordlist
					log.Printf("⚠️  Using wordlist path directly: %s", localWordlist)
				}
			}
		}
	}

	// Build hashcat command with UUID-based outfile
	tempDir := filepath.Join(a.UploadDir, "temp")
	outfile := filepath.Join(tempDir, fmt.Sprintf("cracked-%s.txt", job.ID.String()))
	log.Printf("📁 Outfile will be: %s", outfile)
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

	if job.Rules != "" {
		args = append(args, "-r", job.Rules)
	}

	log.Printf("🔨 DEBUG: Final hashcat command:")
	log.Printf("  📋 Hash File: %s", localHashFile)
	log.Printf("  📋 Wordlist: %s", localWordlist)
	log.Printf("  📋 Outfile: %s", outfile)
	log.Printf("  📋 Hash Type: %d", job.HashType)
	log.Printf("  📋 Attack Mode: %d", job.AttackMode)
	if job.Rules != "" {
		log.Printf("  📋 Rules: %s", job.Rules)
	}
	log.Printf("🔨 Running hashcat with args: %v", args)

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
			if exitCode == 1 {
				// Exhausted - not an error
				a.completeJob(job.ID, "Password not found - exhausted")
				a.cleanupJobFiles(job.ID)
				return nil
			} else if exitCode == 255 {
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
		log.Printf("⚠️  Failed to extract password: %v", err)
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

	// Get file size for progress tracking
	contentLength := resp.ContentLength
	if contentLength > 0 {
		log.Printf("📥 Downloading hash file: %s (Size: %s)", filename, formatFileSize(contentLength))
	} else {
		log.Printf("📥 Downloading hash file: %s (Size: unknown)", filename)
	}

	// Create local file
	localPath := filepath.Join(tempDir, filename)
	file, err := os.Create(localPath)
	if err != nil {
		return "", fmt.Errorf("failed to create local file: %w", err)
	}
	defer file.Close()

	// Use chunked download with progress tracking for large files
	if contentLength > 10*1024*1024 { // 10MB threshold
		_, err = a.chunkedDownloadWithProgress(resp.Body, file, contentLength, filename)
	} else {
		// Simple copy for small files
		_, err = io.Copy(file, resp.Body)
	}
	
	if err != nil {
		os.Remove(localPath) // Clean up on error
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	log.Printf("✅ Downloaded hash file to: %s", localPath)
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

	// Get file size for progress tracking
	contentLength := resp.ContentLength
	if contentLength > 0 {
		log.Printf("📥 Downloading wordlist: %s (Size: %s)", filename, formatFileSize(contentLength))
	} else {
		log.Printf("📥 Downloading wordlist: %s (Size: unknown)", filename)
	}

	// Create local file
	localPath := filepath.Join(tempDir, filename)
	file, err := os.Create(localPath)
	if err != nil {
		return "", fmt.Errorf("failed to create local file: %w", err)
	}
	defer file.Close()

	// Use chunked download with progress tracking for large files
	if contentLength > 10*1024*1024 { // 10MB threshold
		_, err = a.chunkedDownloadWithProgress(resp.Body, file, contentLength, filename)
	} else {
		// Simple copy for small files
		_, err = io.Copy(file, resp.Body)
	}
	
	if err != nil {
		os.Remove(localPath) // Clean up on error
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	log.Printf("✅ Downloaded wordlist to: %s", localPath)
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
		log.Printf("⚠️  Failed to cleanup outfile %s: %v", outfile, err)
	} else {
		log.Printf("🗑️  Cleaned up outfile: %s", outfile)
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
		log.Printf("❌ Failed to send initial job data to server: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("❌ Initial job data failed with status %d: %s", resp.StatusCode, string(body))
	} else {
		log.Printf("✅ Initial job data sent successfully to server")
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
		log.Printf("❌ Failed to send job data update to server: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("❌ Job data update failed with status %d: %s", resp.StatusCode, string(body))
	} else {
		log.Printf("✅ Job data update sent successfully (Progress: %.2f%%, Speed: %d H/s)", progress, speed)
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
				log.Printf("⚠️  Failed to check job status: %v", err)
				continue
			}

			// Handle status changes
			switch status {
			case "paused", "failed", "cancelled":
				log.Printf("🛑 Job %s status changed to %s, terminating hashcat", jobID, status)
				if cmd.Process != nil {
					cmd.Process.Kill()
				}
				return
			}
		}
	}
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
		log.Printf("❌ Failed to send job completion to server: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("❌ Job completion failed with status %d: %s", resp.StatusCode, string(body))
	} else {
		log.Printf("✅ Job completion sent successfully to server")
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
		log.Printf("❌ Failed to send job failure to server: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("❌ Job failure notification failed with status %d: %s", resp.StatusCode, string(body))
	} else {
		log.Printf("✅ Job failure notification sent successfully to server")
	}
}

func getLocalIP() string {
	// Use hostname -I to get local IP addresses
	cmd := exec.Command("hostname", "-I")
	output, err := cmd.Output()
	if err != nil {
		log.Printf("⚠️ Warning: Failed to get local IP using hostname -I: %v", err)
		return "127.0.0.1" // Fallback to localhost
	}

	// Parse output and get first non-localhost IP
	ips := strings.Fields(string(output))
	for _, ip := range ips {
		ip = strings.TrimSpace(ip)
		// Skip localhost and loopback addresses
		if ip != "127.0.0.1" && ip != "::1" && ip != "" {
			log.Printf("🔍 Found local IP: %s", ip)
			return ip
		}
	}

	log.Printf("⚠️ Warning: No valid local IP found, using fallback")
	return "127.0.0.1"
}

// validateLocalIP validates if the provided IP is a valid local IP address
func validateLocalIP(providedIP string) error {
	// Get actual local IPs using hostname -I
	cmd := exec.Command("hostname", "-I")
	output, err := cmd.Output()
	if err != nil {
		log.Printf("⚠️ Warning: Failed to get local IP using hostname -I: %v", err)
		// If we can't validate, allow the IP to pass
		return nil
	}

	// Parse output and check if provided IP exists in local IPs
	localIPs := strings.Fields(string(output))
	for _, localIP := range localIPs {
		localIP = strings.TrimSpace(localIP)
		if localIP == providedIP {
			log.Printf("✅ IP address validation passed: %s is a valid local IP", providedIP)
			return nil
		}
	}

	// IP not found in local IPs
	return fmt.Errorf("❌ IP address validation failed: provided IP '%s' is not a valid local IP address. Local IPs: %v", providedIP, localIPs)
}

// detectCapabilitiesWithHashcat detects server capabilities using hashcat -I command
func detectCapabilitiesWithHashcat() string {
	log.Printf("🔍 Starting hashcat -I capabilities detection...")

	// Check if hashcat is available
	if _, err := exec.LookPath("hashcat"); err != nil {
		log.Printf("⚠️ Warning: hashcat not found, falling back to basic detection")
		log.Printf("🔍 Error details: %v", err)
		return detectCapabilitiesBasic()
	}

	log.Printf("✅ hashcat command found, executing hashcat -I...")

	// Run hashcat -I to get device information
	cmd := exec.Command("hashcat", "-I")
	output, err := cmd.Output()
	if err != nil {
		log.Printf("⚠️ Warning: Failed to run hashcat -I: %v", err)
		log.Printf("⚠️ Falling back to basic detection")
		return detectCapabilitiesBasic()
	}

	log.Printf("✅ hashcat -I executed successfully")

	// Parse output to find device types
	outputStr := string(output)
	lines := strings.Split(outputStr, "\n")

	log.Printf("🔍 Hashcat -I output lines count: %d", len(lines))
	log.Printf("🔍 Raw output preview (first 10 lines):")
	for i, line := range lines[:min(10, len(lines))] {
		log.Printf("   Line %d: %s", i+1, line)
	}

	var deviceTypes []string

	for i, line := range lines {
		line = strings.TrimSpace(line)

		// Look for device section headers
		if strings.Contains(line, "Backend Device ID #") {
			log.Printf("🔍 Found device section header at line %d: %s", i+1, line)
			continue
		}

		// Look for Type line
		if strings.HasPrefix(line, "Type...........:") {
			log.Printf("🔍 Found Type line at line %d: %s", i+1, line)
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				deviceType := strings.TrimSpace(parts[1])
				if deviceType != "" {
					deviceTypes = append(deviceTypes, deviceType)
					log.Printf("🔍 Detected device type: %s", deviceType)
				}
			}
		}
	}

	log.Printf("🔍 Total device types found: %d", len(deviceTypes))
	log.Printf("🔍 Device types: %v", deviceTypes)

	// Determine capabilities based on detected devices
	if len(deviceTypes) == 0 {
		log.Printf("⚠️ No device types found in hashcat -I output, falling back to basic detection")
		return detectCapabilitiesBasic()
	}

	// Check if any GPU devices are found
	for _, deviceType := range deviceTypes {
		log.Printf("🔍 Checking device type for GPU: %s", deviceType)
		if strings.Contains(strings.ToUpper(deviceType), "GPU") {
			log.Printf("✅ GPU device detected: %s", deviceType)
			return "GPU"
		}
	}

	// If no GPU, check for CPU
	for _, deviceType := range deviceTypes {
		log.Printf("🔍 Checking device type for CPU: %s", deviceType)
		if strings.Contains(strings.ToUpper(deviceType), "CPU") {
			log.Printf("✅ CPU device detected: %s", deviceType)
			return "CPU"
		}
	}

	// If we can't determine, log all found types and fallback
	log.Printf("⚠️ Could not determine capabilities from device types: %v", deviceTypes)
	log.Printf("⚠️ Falling back to basic detection")
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
	log.Printf("🔍 Starting GPU detection...")

	// Check for NVIDIA GPU
	if _, err := exec.LookPath("nvidia-smi"); err == nil {
		log.Printf("🔍 nvidia-smi command found, checking if GPU is working...")
		// Try to run nvidia-smi to verify GPU is working
		cmd := exec.Command("nvidia-smi", "--query-gpu=name", "--format=csv,noheader,nounits")
		if output, err := cmd.Output(); err == nil && len(strings.TrimSpace(string(output))) > 0 {
			gpuName := strings.TrimSpace(string(output))
			log.Printf("✅ Detected NVIDIA GPU: %s", gpuName)
			return true
		} else {
			log.Printf("⚠️ nvidia-smi found but failed to run or no output: %v", err)
		}
	} else {
		log.Printf("🔍 nvidia-smi command not found")
	}

	// Check for AMD GPU
	if _, err := exec.LookPath("rocm-smi"); err == nil {
		log.Printf("🔍 rocm-smi command found, checking if GPU is working...")
		cmd := exec.Command("rocm-smi", "--list-gpus")
		if output, err := cmd.Output(); err == nil && len(strings.TrimSpace(string(output))) > 0 {
			log.Printf("✅ Detected AMD GPU (ROCm): %s", strings.TrimSpace(string(output)))
			return true
		} else {
			log.Printf("⚠️ rocm-smi found but failed to run or no output: %v", err)
		}
	} else {
		log.Printf("🔍 rocm-smi command not found")
	}

	// Check for Intel GPU
	if _, err := exec.LookPath("intel_gpu_top"); err == nil {
		log.Printf("🔍 intel_gpu_top command found, checking if GPU is working...")
		cmd := exec.Command("intel_gpu_top", "-J", "-s", "1")
		if output, err := cmd.Output(); err == nil && len(strings.TrimSpace(string(output))) > 0 {
			log.Printf("✅ Detected Intel GPU: %s", strings.TrimSpace(string(output)))
			return true
		} else {
			log.Printf("⚠️ intel_gpu_top found but failed to run or no output: %v", err)
		}
	} else {
		log.Printf("🔍 intel_gpu_top command not found")
	}

	// Additional check: look for GPU devices in /proc
	if _, err := os.Stat("/proc/driver/nvidia"); err == nil {
		log.Printf("✅ Found NVIDIA driver in /proc/driver/nvidia")
		return true
	}

	if _, err := os.Stat("/sys/class/drm"); err == nil {
		// Check if there are GPU devices
		if files, err := os.ReadDir("/sys/class/drm"); err == nil {
			for _, file := range files {
				if strings.HasPrefix(file.Name(), "card") && file.Name() != "card0" {
					log.Printf("✅ Found GPU device: %s", file.Name())
					return true
				}
			}
		}
	}

	log.Printf("🔍 No GPU detected, using CPU")
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

// chunkedDownloadWithProgress downloads large files in chunks with progress tracking
func (a *Agent) chunkedDownloadWithProgress(src io.Reader, dst *os.File, totalSize int64, filename string) (int64, error) {
	const chunkSize = 1024 * 1024 // 1MB chunks
	buffer := make([]byte, chunkSize)
	var totalWritten int64
	lastProgress := time.Now()

	for {
		n, err := src.Read(buffer)
		if n > 0 {
			written, writeErr := dst.Write(buffer[:n])
			if writeErr != nil {
				return totalWritten, writeErr
			}
			totalWritten += int64(written)
		}

		// Show progress every 5 seconds or every 50MB
		if time.Since(lastProgress) >= 5*time.Second || totalWritten%50*1024*1024 < int64(n) {
			progress := float64(totalWritten) / float64(totalSize) * 100
			log.Printf("📊 Download progress: %s - %.1f%% (%s / %s)", 
				filename, progress, formatFileSize(totalWritten), formatFileSize(totalSize))
			lastProgress = time.Now()
		}

		if err != nil {
			if err == io.EOF {
				break
			}
			return totalWritten, err
		}
	}

	return totalWritten, nil
}

// findLocalWordlistByUUID searches for a local wordlist file that matches the given UUID
// This function checks if we have a local file that corresponds to the server's wordlist ID
func (a *Agent) findLocalWordlistByUUID(wordlistID string) string {
	log.Printf("🔍 DEBUG: Starting search for wordlist UUID: %s", wordlistID)
	log.Printf("🔍 DEBUG: Searching in LocalFiles map (%d files)...", len(a.LocalFiles))
	
	// First, check if we have any wordlist files locally
	for filename, localFile := range a.LocalFiles {
		if localFile.Type == "wordlist" {
			log.Printf("🔍 DEBUG: Checking wordlist file: %s -> %s", filename, localFile.Path)
			
			// Check if the filename contains the UUID (common pattern: wordlist-UUID.txt)
			if strings.Contains(filename, wordlistID) {
				log.Printf("✅ SUCCESS: Found local wordlist by filename match: %s (UUID: %s)", filename, wordlistID)
				return localFile.Path
			}
			
			// Also check if the file path contains the UUID
			if strings.Contains(localFile.Path, wordlistID) {
				log.Printf("✅ SUCCESS: Found local wordlist by path match: %s (UUID: %s)", localFile.Path, wordlistID)
				return localFile.Path
			}
		}
	}
	
	log.Printf("🔍 DEBUG: No match found in LocalFiles, checking temp directory...")
	
	// If no exact match found, try to find by scanning the temp directory
	// This is useful when files are downloaded to temp but not yet moved to wordlists directory
	tempDir := filepath.Join(a.UploadDir, "temp")
	log.Printf("🔍 DEBUG: Scanning temp directory: %s", tempDir)
	
	if files, err := os.ReadDir(tempDir); err == nil {
		log.Printf("🔍 DEBUG: Found %d files in temp directory", len(files))
		for _, file := range files {
			if !file.IsDir() {
				log.Printf("🔍 DEBUG: Checking temp file: %s", file.Name())
				if strings.Contains(file.Name(), wordlistID) {
					fullPath := filepath.Join(tempDir, file.Name())
					log.Printf("✅ SUCCESS: Found wordlist in temp directory: %s (UUID: %s)", fullPath, wordlistID)
					return fullPath
				}
			}
		}
	} else {
		log.Printf("⚠️  WARNING: Failed to read temp directory %s: %v", tempDir, err)
	}
	
	log.Printf("❌ FAILED: No local wordlist found for UUID: %s", wordlistID)
	log.Printf("🔍 DEBUG: Search completed for wordlist UUID: %s", wordlistID)
	return ""
}

// findLocalHashFileByUUID searches for a local hash file that matches the given UUID
// This function checks if we have a local file that corresponds to the server's hash file ID
func (a *Agent) findLocalHashFileByUUID(hashFileID string) string {
	log.Printf("🔍 DEBUG: Starting search for hash file UUID: %s", hashFileID)
	log.Printf("🔍 DEBUG: Searching in LocalFiles map (%d files)...", len(a.LocalFiles))
	
	// First, check if we have any hash file files locally
	for filename, localFile := range a.LocalFiles {
		if localFile.Type == "hash_file" {
			log.Printf("🔍 DEBUG: Checking hash file: %s -> %s", filename, localFile.Path)
			
			// Check if the filename contains the UUID (common pattern: hashfile-UUID.hccapx)
			if strings.Contains(filename, hashFileID) {
				log.Printf("✅ SUCCESS: Found local hash file by filename match: %s (UUID: %s)", filename, hashFileID)
				return localFile.Path
			}
			
			// Also check if the file path contains the UUID
			if strings.Contains(localFile.Path, hashFileID) {
				log.Printf("✅ SUCCESS: Found local hash file by path match: %s (UUID: %s)", localFile.Path, hashFileID)
				return localFile.Path
			}
		}
	}
	
	log.Printf("🔍 DEBUG: No match found in LocalFiles, checking temp directory...")
	
	// If no exact match found, try to find by scanning the temp directory
	// This is useful when files are downloaded to temp but not yet moved to hash-files directory
	tempDir := filepath.Join(a.UploadDir, "temp")
	log.Printf("🔍 DEBUG: Scanning temp directory: %s", tempDir)
	
	if files, err := os.ReadDir(tempDir); err == nil {
		log.Printf("🔍 DEBUG: Found %d files in temp directory", len(files))
		for _, file := range files {
			if !file.IsDir() {
				log.Printf("🔍 DEBUG: Checking temp file: %s", file.Name())
				if strings.Contains(file.Name(), hashFileID) {
					fullPath := filepath.Join(tempDir, file.Name())
					log.Printf("✅ SUCCESS: Found hash file in temp directory: %s (UUID: %s)", fullPath, hashFileID)
					return fullPath
				}
			}
		}
	} else {
		log.Printf("⚠️  WARNING: Failed to read temp directory %s: %v", tempDir, err)
	}
	
	log.Printf("❌ FAILED: No local hash file found for UUID: %s", hashFileID)
	log.Printf("🔍 DEBUG: Search completed for hash file UUID: %s", hashFileID)
	return ""
}
