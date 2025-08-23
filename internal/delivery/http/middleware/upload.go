package middleware

import (
	"io"

	"github.com/gin-gonic/gin"
)

// UploadProgressMiddleware tracks upload progress for large files
func UploadProgressMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if this is a wordlist upload
		if c.Request.URL.Path == "/api/v1/wordlists/upload" && c.Request.Method == "POST" {
			// Set longer timeout for large file uploads
			c.Request.Body = &progressReader{
				reader: c.Request.Body,
				onProgress: func(bytesRead int64) {
					// Log progress every 10MB
					if bytesRead%(10*1024*1024) == 0 {
						gin.DefaultWriter.Write([]byte(
							"Upload progress: " + formatBytes(bytesRead) + "\n"))
					}
				},
			}
		}
		c.Next()
	}
}

// progressReader wraps io.Reader to track upload progress
type progressReader struct {
	reader     io.Reader
	bytesRead  int64
	onProgress func(int64)
}

func (pr *progressReader) Read(p []byte) (n int, err error) {
	n, err = pr.reader.Read(p)
	pr.bytesRead += int64(n)
	if pr.onProgress != nil {
		pr.onProgress(pr.bytesRead)
	}
	return
}

// Close implements io.Closer interface
func (pr *progressReader) Close() error {
	if closer, ok := pr.reader.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

// formatBytes converts bytes to human readable format
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return string(rune(bytes)) + " B"
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return string(rune(bytes/div)) + " " + string("KMGTPE"[exp]) + "B"
}
