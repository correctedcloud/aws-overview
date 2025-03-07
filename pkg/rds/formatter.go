package rds

import (
	"fmt"
	"strings"

	"github.com/correctedcloud/aws-overview/pkg/common"
)

// FormatDBInstances formats DB instance summaries for terminal display
func FormatDBInstances(summaries []DBInstanceSummary) string {
	if len(summaries) == 0 {
		return "No DB instances found"
	}

	var output strings.Builder
	output.WriteString("RDS INSTANCES\n")
	output.WriteString("=============\n\n")

	for _, instance := range summaries {
		statusSymbol := getStatusSymbol(instance.Status)
		output.WriteString(fmt.Sprintf("%s %s (%s)\n", statusSymbol, instance.Identifier, instance.Engine))

		if instance.Endpoint != "" {
			output.WriteString(fmt.Sprintf("  Endpoint: %s\n", instance.Endpoint))
		}

		output.WriteString("\n  CPU Utilization (1 hour):\n")
		if len(instance.CPUData) > 0 {
			cpuGraph := common.GenerateSparkline(instance.CPUData, "CPU (%)", 3)
			output.WriteString(fmt.Sprintf("%s\n", cpuGraph))
		} else {
			output.WriteString("  No CPU data available\n")
		}

		output.WriteString("\n  Memory Utilization (1 hour):\n")
		if len(instance.MemoryData) > 0 {
			memoryGraph := common.GenerateSparkline(instance.MemoryData, "Memory (%)", 3)
			output.WriteString(fmt.Sprintf("%s\n", memoryGraph))
		} else {
			output.WriteString("  No memory data available\n")
		}

		output.WriteString("\n  Recent Errors:\n")
		if len(instance.RecentErrors) > 0 {
			for _, err := range instance.RecentErrors {
				output.WriteString(fmt.Sprintf("  - %s\n", err))
			}
		} else {
			output.WriteString("  No recent errors\n")
		}

		output.WriteString("\n")
	}

	return output.String()
}

// GetDBInstancesSummary returns a brief summary of DB instances
func GetDBInstancesSummary(summaries []DBInstanceSummary) string {
	if len(summaries) == 0 {
		return "No DB instances found"
	}

	// Count instances by status
	available := 0
	other := 0

	// Calculate average CPU and memory
	totalCPU := 0.0
	cpuDataPoints := 0
	totalMemory := 0.0
	memoryDataPoints := 0

	for _, instance := range summaries {
		if instance.Status == "available" {
			available++
		} else {
			other++
		}

		// Add the last CPU data point if available
		if len(instance.CPUData) > 0 {
			totalCPU += instance.CPUData[len(instance.CPUData)-1]
			cpuDataPoints++
		}

		// Add the last memory data point if available
		if len(instance.MemoryData) > 0 {
			totalMemory += instance.MemoryData[len(instance.MemoryData)-1]
			memoryDataPoints++
		}
	}

	// Calculate averages
	var cpuAvg, memoryAvg float64
	if cpuDataPoints > 0 {
		cpuAvg = totalCPU / float64(cpuDataPoints)
	}
	if memoryDataPoints > 0 {
		memoryAvg = totalMemory / float64(memoryDataPoints)
	}

	return fmt.Sprintf("%d instances (%d available), Avg CPU: %s, Avg Memory: %s",
		len(summaries),
		available,
		common.FormatPercentage(cpuAvg),
		common.FormatPercentage(memoryAvg))
}

// getStatusSymbol returns an appropriate symbol for an instance status
func getStatusSymbol(status string) string {
	switch status {
	case "available":
		return "✅"
	case "creating":
		return "🔄"
	case "deleting":
		return "🗑️"
	case "failed":
		return "❌"
	case "inaccessible-encryption-credentials":
		return "🔒"
	case "incompatible-network":
		return "🌐"
	case "incompatible-option-group":
		return "⚙️"
	case "incompatible-parameters":
		return "⚙️"
	case "incompatible-restore":
		return "🔄"
	case "maintenance":
		return "🔧"
	case "modifying":
		return "🔄"
	case "stopped":
		return "⏹️"
	case "stopping":
		return "⏹️"
	case "storage-full":
		return "💾"
	default:
		return "❓"
	}
}
