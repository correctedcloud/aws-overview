package ecs

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

var timeNow = time.Now

// GetServicesSummary returns a summary of ECS services
func GetServicesSummary(services []ServiceSummary) string {
	active := 0
	draining := 0
	other := 0
	healthyServices := 0

	clusters := make(map[string]bool)

	for _, service := range services {
		clusters[service.ClusterName] = true

		switch service.Status {
		case "ACTIVE":
			active++
		case "DRAINING":
			draining++
		default:
			other++
		}

		// Count healthy services (running tasks match desired)
		if service.RunningCount == service.DesiredCount && service.DesiredCount > 0 {
			healthyServices++
		}
	}

	return fmt.Sprintf("%d services in %d clusters (%d active, %d draining, %d other, %d/%d healthy)",
		len(services), len(clusters), active, draining, other, healthyServices, len(services))
}

// FormatServices returns a formatted string of ECS services
func FormatServices(services []ServiceSummary) string {
	if len(services) == 0 {
		return "No ECS services found."
	}

	// First, group services by cluster
	servicesByCluster := make(map[string][]ServiceSummary)
	for _, service := range services {
		servicesByCluster[service.ClusterName] = append(servicesByCluster[service.ClusterName], service)
	}

	// Get sorted cluster names
	clusterNames := make([]string, 0, len(servicesByCluster))
	for cluster := range servicesByCluster {
		clusterNames = append(clusterNames, cluster)
	}
	sort.Strings(clusterNames)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("ECS Services (%d):\n\n", len(services)))

	// Format by cluster
	for _, clusterName := range clusterNames {
		clusterServices := servicesByCluster[clusterName]

		// Sort services by name within cluster
		sort.Slice(clusterServices, func(i, j int) bool {
			return clusterServices[i].ServiceName < clusterServices[j].ServiceName
		})

		// Cluster header
		sb.WriteString(fmt.Sprintf("🚀 Cluster: %s (%d services)\n", clusterName, len(clusterServices)))
		sb.WriteString(strings.Repeat("-", 40) + "\n")

		// Format each service
		for _, service := range clusterServices {
			// Health status indicator
			healthIndicator := "🔴"
			if service.RunningCount == service.DesiredCount && service.DesiredCount > 0 {
				healthIndicator = "🟢"
			} else if service.RunningCount > 0 {
				healthIndicator = "🟠"
			} else if service.DesiredCount == 0 && service.RunningCount == 0 {
				healthIndicator = "⚪"
			}

			sb.WriteString(fmt.Sprintf("%s %s\n", healthIndicator, service.ServiceName))

			// Status and deployment status
			deploymentInfo := ""
			if service.DeploymentStatus != "stable" {
				deploymentInfo = fmt.Sprintf(" (deployment: %s)", service.DeploymentStatus)
			}
			sb.WriteString(fmt.Sprintf("   Status: %s%s\n", service.Status, deploymentInfo))

			// Task counts
			sb.WriteString(fmt.Sprintf("   Tasks: %d/%d running (%d pending)\n",
				service.RunningCount, service.DesiredCount, service.PendingCount))

			// Task definition and launch type
			sb.WriteString(fmt.Sprintf("   Task Definition: %s | %s | %s\n",
				service.TaskDefinition, service.LaunchType, service.NetworkMode))

			// Last deployment time
			lastDeploymentTime := formatUptime(service.LastDeploymentTime)
			sb.WriteString(fmt.Sprintf("   Last Deployment: %s (%s ago)\n",
				service.LastDeploymentTime.Format("2006-01-02 15:04:05"), lastDeploymentTime))

			// Load balancers
			if len(service.LoadBalancers) > 0 {
				sb.WriteString(fmt.Sprintf("   Load Balancers: %s\n",
					strings.Join(service.LoadBalancers, ", ")))
			}

			// Format important tags
			importantTags := []string{"Environment", "Project", "Owner", "Application"}
			var tagStrings []string
			for _, tag := range importantTags {
				if value, ok := service.Tags[tag]; ok {
					tagStrings = append(tagStrings, fmt.Sprintf("%s: %s", tag, value))
				}
			}

			if len(tagStrings) > 0 {
				sb.WriteString(fmt.Sprintf("   Tags: %s\n", strings.Join(tagStrings, " | ")))
			}

			sb.WriteString("\n")
		}

		// Add a separator between clusters
		sb.WriteString("\n")
	}

	return sb.String()
}

// formatUptime formats the uptime of a service
func formatUptime(createdTime time.Time) string {
	duration := timeNow().Sub(createdTime)

	days := int(duration.Hours() / 24)
	months := days / 30
	years := months / 12

	if years > 0 {
		return fmt.Sprintf("%dy %dm", years, months%12)
	}
	if months > 0 {
		return fmt.Sprintf("%dm %dd", months, days%30)
	}
	if days > 0 {
		hours := int(duration.Hours()) % 24
		return fmt.Sprintf("%dd %dh", days, hours)
	}

	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}

	return fmt.Sprintf("%dm", minutes)
}
