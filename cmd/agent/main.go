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
	ID         uuid.UUID
	Name       string
	ServerURL  string
	Client     *http.Client
	CurrentJob *domain.Job
	UploadDir  string
	LocalFiles map[string]LocalFile // filename -> LocalFile
}

type LocalFile struct {
	Name    string    `json:"name"`
	Path    string    `json:"path"`
	Size    int64     `json:"size"`
	Type    string    `json:"type"` // wordlist, hash_file
	Hash    string    `json:"hash"` // MD5 hash for integrity
	ModTime time.Time `json:"mod_time"`
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
	rootCmd.Flags().String("capabilities", "GPU", "Agent capabilities")
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
	uploadDir := viper.GetString("upload-dir")

	if name == "" {
		hostname, _ := os.Hostname()
		name = fmt.Sprintf("agent-%s", hostname)
	}

	if ip == "" {
		// Try to get local IP
		ip = getLocalIP()
	}

	agent := &Agent{
		Name:       name,
		ServerURL:  serverURL,
		Client:     &http.Client{Timeout: 30 * time.Second},
		UploadDir:  uploadDir,
		LocalFiles: make(map[string]LocalFile),
	}

	// Create upload directory structure
	if err := agent.initializeDirectories(); err != nil {
		log.Fatalf("Failed to initialize directories: %v", err)
	}

	// Scan local files
	if err := agent.scanLocalFiles(); err != nil {
		log.Printf("Warning: Failed to scan local files: %v", err)
	}

	// Register agent with server
	if err := agent.registerWithServer(name, ip, port, capabilities); err != nil {
		log.Fatalf("Failed to register with server: %v", err)
	}

	log.Printf("Agent %s (%s) registered successfully", agent.Name, agent.ID.String())
	log.Printf("Local upload directory: %s", agent.UploadDir)
	log.Printf("Found %d local files", len(agent.LocalFiles))

	// Register local files with server
	if err := agent.registerLocalFiles(); err != nil {
		log.Printf("Warning: Failed to register local files: %v", err)
	}

	// Start background processes
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go agent.startHeartbeat(ctx)
	go agent.pollForJobs(ctx)
	go agent.watchLocalFiles(ctx) // Watch for file changes

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down agent...")

	// Update status to offline
	agent.updateStatus("offline")

	log.Println("Agent exited")
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
	log.Println("🔍 Scanning local files...")

	// Scan wordlists
	wordlistDir := filepath.Join(a.UploadDir, "wordlists")
	if err := a.scanDirectory(wordlistDir, "wordlist"); err != nil {
		log.Printf("Warning: Failed to scan wordlists directory: %v", err)
	}

	// Scan hash files
	hashFileDir := filepath.Join(a.UploadDir, "hash-files")
	if err := a.scanDirectory(hashFileDir, "hash_file"); err != nil {
		log.Printf("Warning: Failed to scan hash-files directory: %v", err)
	}

	// Also scan root upload directory for legacy files
	if err := a.scanDirectory(a.UploadDir, "auto"); err != nil {
		log.Printf("Warning: Failed to scan root upload directory: %v", err)
	}

	log.Printf("✅ Scanned %d local files", len(a.LocalFiles))
	for filename, file := range a.LocalFiles {
		log.Printf("  📄 %s (%s, %s)", filename, file.Type, formatFileSize(file.Size))
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

func (a *Agent) registerWithServer(name, ip string, port int, capabilities string) error {
	req := domain.CreateAgentRequest{
		Name:         name,
		IPAddress:    ip,
		Port:         port,
		Capabilities: capabilities,
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
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

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
	url := fmt.Sprintf("%s/api/v1/agents/%s/heartbeat", a.ServerURL, a.ID.String())
	resp, err := a.Client.Post(url, "application/json", nil)
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

func (a *Agent) updateStatus(status string) error {
	req := struct {
		Status string `json:"status"`
	}{Status: status}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/api/v1/agents/%s/status", a.ServerURL, a.ID.String())
	httpReq, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := a.Client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
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
	// Get jobs assigned to this agent
	url := fmt.Sprintf("%s/api/v1/jobs?status=pending", a.ServerURL)
	resp, err := a.Client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var response struct {
		Data []domain.Job `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return err
	}

	// Find a job assigned to this agent
	for _, job := range response.Data {
		if job.AgentID != nil && *job.AgentID == a.ID {
			log.Printf("Found assigned job: %s", job.Name)
			a.CurrentJob = &job
			go a.executeJob(&job)
			break
		}
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
	// Resolve hash file (local first, download if needed)
	var localHashFile string
	var err error

	if job.HashFileID != nil {
		// Try to find hash file locally first
		hashFileName := job.HashFile
		if hashFileName != "" {
			if localPath, found := a.findLocalFile(filepath.Base(hashFileName)); found {
				localHashFile = localPath
				log.Printf("📁 Using local hash file: %s", localHashFile)
			}
		}

		// Download if not found locally
		if localHashFile == "" {
			localHashFile, err = a.downloadHashFile(*job.HashFileID)
			if err != nil {
				return fmt.Errorf("failed to download hash file: %w", err)
			}
			defer os.Remove(localHashFile) // Clean up downloaded file
			log.Printf("📥 Downloaded hash file: %s", localHashFile)
		}
	} else {
		// Fallback to original path
		localHashFile = job.HashFile
	}

	// Resolve wordlist (local first, download if needed)
	var localWordlist string
	if job.Wordlist != "" {
		if localPath, found := a.findLocalFile(job.Wordlist); found {
			localWordlist = localPath
			log.Printf("📁 Using local wordlist: %s", localWordlist)
		} else {
			// If not found locally, try to download or use direct path
			localWordlist = job.Wordlist
			log.Printf("⚠️  Wordlist not found locally, using path: %s", localWordlist)
		}
	}

	// Build hashcat command
	args := []string{
		"-m", strconv.Itoa(job.HashType),
		"-a", strconv.Itoa(job.AttackMode),
		localHashFile,
		localWordlist,
		"-w", "4",
		"--status",
		"--status-timer=5",
		"--potfile-disable",
	}

	if job.Rules != "" {
		args = append(args, "-r", job.Rules)
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

	// Wait for command to complete
	if err := cmd.Wait(); err != nil {
		// Check if hashcat found the password (exit code 0) or exhausted (exit code 1)
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode := exitError.ExitCode()
			if exitCode == 1 {
				// Exhausted - not an error
				a.completeJob(job.ID, "Password not found - exhausted")
				return nil
			}
		}
		return err
	}

	// Success - password found
	a.completeJob(job.ID, "Password found")
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

	log.Printf("Downloaded hash file to: %s", localPath)
	return localPath, nil
}

func (a *Agent) monitorHashcatOutput(job *domain.Job, stdout, stderr io.Reader) {
	// Parse hashcat output for progress updates
	progressRegex := regexp.MustCompile(`Progress\.+:\s*(\d+)/(\d+)\s*\((\d+\.\d+)%\)`)
	speedRegex := regexp.MustCompile(`Speed\.+:\s*(\d+)\s*H/s`)

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

				a.updateJobProgress(job.ID, progress, speed)
			}
		}
	}

	go scanner(stdout)
	go scanner(stderr)
}

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

	a.Client.Do(httpReq)
}

func (a *Agent) completeJob(jobID uuid.UUID, result string) {
	req := struct {
		Result string `json:"result"`
	}{Result: result}

	jsonData, _ := json.Marshal(req)
	url := fmt.Sprintf("%s/api/v1/jobs/%s/complete", a.ServerURL, jobID.String())

	a.Client.Post(url, "application/json", bytes.NewBuffer(jsonData))
}

func (a *Agent) failJob(jobID uuid.UUID, reason string) {
	req := struct {
		Reason string `json:"reason"`
	}{Reason: reason}

	jsonData, _ := json.Marshal(req)
	url := fmt.Sprintf("%s/api/v1/jobs/%s/fail", a.ServerURL, jobID.String())

	a.Client.Post(url, "application/json", bytes.NewBuffer(jsonData))
}

func getLocalIP() string {
	// Simple implementation - could be enhanced
	return "127.0.0.1"
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
