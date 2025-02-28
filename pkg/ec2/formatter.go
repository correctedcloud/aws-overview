package ec2

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

var timeNow = time.Now

// GetInstancesSummary returns a summary of EC2 instances
func GetInstancesSummary(instances []InstanceSummary) string {
	running := 0
	stopped := 0
	other := 0

	for _, instance := range instances {
		switch instance.State {
		case "running":
			running++
		case "stopped":
			stopped++
		default:
			other++
		}
	}

	return fmt.Sprintf("%d total (%d running, %d stopped, %d other)", 
		len(instances), running, stopped, other)
}

// FormatInstances returns a formatted string of EC2 instances
func FormatInstances(instances []InstanceSummary) string {
	if len(instances) == 0 {
		return "No EC2 instances found."
	}

	// Sort instances by name, then by ID
	sort.Slice(instances, func(i, j int) bool {
		if instances[i].Name != instances[j].Name {
			return instances[i].Name < instances[j].Name
		}
		return instances[i].InstanceID < instances[j].InstanceID
	})

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("EC2 Instances (%d):\n\n", len(instances)))

	for _, instance := range instances {
		// Format instance name and ID
		nameDisplay := instance.Name
		if nameDisplay == "" {
			nameDisplay = "<unnamed>"
		}
		sb.WriteString(fmt.Sprintf("🖥️  %s (%s)\n", nameDisplay, instance.InstanceID))
		
		// Format instance type and state with color indicators
		stateIndicator := "🔴"
		if instance.State == "running" {
			stateIndicator = "🟢"
		} else if instance.State == "stopped" {
			stateIndicator = "🟠"
		}
		sb.WriteString(fmt.Sprintf("   Type: %s | State: %s %s\n", 
			instance.InstanceType, stateIndicator, instance.State))
		
		// Format IPs
		sb.WriteString(fmt.Sprintf("   Private IP: %s", instance.PrivateIP))
		if instance.PublicIP != "" {
			sb.WriteString(fmt.Sprintf(" | Public IP: %s", instance.PublicIP))
		}
		sb.WriteString("\n")
		
		// Format platform and launch time
		uptime := formatUptime(instance.LaunchTime)
		sb.WriteString(fmt.Sprintf("   Platform: %s | Launched: %s (%s)\n", 
			instance.Platform, 
			instance.LaunchTime.Format("2006-01-02 15:04:05"),
			uptime))
		
		// Format VPC and subnet
		sb.WriteString(fmt.Sprintf("   VPC: %s | Subnet: %s | AZ: %s\n", 
			instance.VpcID, instance.SubnetID, instance.AvailabilityZone))
		
		// Format security groups
		if len(instance.SecurityGroups) > 0 {
			sb.WriteString(fmt.Sprintf("   Security Groups: %s\n", 
				strings.Join(instance.SecurityGroups, ", ")))
		}
		
		// Format important tags
		importantTags := []string{"Environment", "Project", "Owner", "Role", "Application"}
		var tagStrings []string
		for _, tag := range importantTags {
			if value, ok := instance.Tags[tag]; ok {
				tagStrings = append(tagStrings, fmt.Sprintf("%s: %s", tag, value))
			}
		}
		
		if len(tagStrings) > 0 {
			sb.WriteString(fmt.Sprintf("   Tags: %s\n", strings.Join(tagStrings, " | ")))
		}
		
		sb.WriteString("\n")
	}

	return sb.String()
}

// formatUptime formats the uptime of an instance
func formatUptime(launchTime time.Time) string {
	duration := timeNow().Sub(launchTime)
	
	days := int(duration.Hours() / 24)
	hours := int(duration.Hours()) % 24
	minutes := int(duration.Minutes()) % 60
	
	if days > 0 {
		return fmt.Sprintf("%dd %dh", days, hours)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}
