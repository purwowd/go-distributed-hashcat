package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestSortAgentsByPriority(t *testing.T) {
	now := time.Now()
	gpuID := uuid.New()
	cpuID := uuid.New()
	poclID := uuid.New()

	agents := []Agent{
		{ID: poclID, Name: "pocl-agent", ResourceType: "GPU", Processor: "Portable Computing Language", Speed: 202700, CreatedAt: now},
		{ID: cpuID, Name: "cpu-1", ResourceType: "CPU", Processor: "Intel Xeon", Speed: 500000, CreatedAt: now},
		{ID: gpuID, Name: "rtx-agent", ResourceType: "GPU", Processor: "NVIDIA RTX 6000 Ada Generation", Speed: 4334600, CreatedAt: now.Add(-time.Hour)},
	}

	SortAgentsByPriority(agents)

	assert.Equal(t, "rtx-agent", agents[0].Name)
	assert.Equal(t, "cpu-1", agents[1].Name)
	assert.Equal(t, "pocl-agent", agents[2].Name)
}
