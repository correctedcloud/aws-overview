package alb

import (
	"fmt"
	"strings"
)

// FormatLoadBalancers formats load balancer summaries for terminal display
func FormatLoadBalancers(summaries []LoadBalancerSummary) string {
	if len(summaries) == 0 {
		return "No load balancers found"
	}

	var output strings.Builder
	output.WriteString("LOAD BALANCERS\n")
	output.WriteString("==============\n\n")

	for _, lb := range summaries {
		output.WriteString(fmt.Sprintf("🔄 %s (%s)\n", lb.Name, lb.DNSName))

		if len(lb.TargetGroups) == 0 {
			output.WriteString("  No target groups\n\n")
			continue
		}

		for _, tg := range lb.TargetGroups {
			output.WriteString(fmt.Sprintf("  📋 %s\n", tg.Name))

			if len(tg.Targets) == 0 {
				output.WriteString("    No targets\n")
				continue
			}

			for _, target := range tg.Targets {
				statusSymbol := getStatusSymbol(target.Status)
				output.WriteString(fmt.Sprintf("    %s %s:%d - %s",
					statusSymbol,
					target.ID,
					target.Port,
					target.Status))

				if target.Reason != "" {
					output.WriteString(fmt.Sprintf(" (%s)", target.Reason))
				}

				output.WriteString("\n")
			}
		}

		output.WriteString("\n")
	}

	return output.String()
}

// GetLoadBalancersSummary returns a brief summary of load balancers
func GetLoadBalancersSummary(summaries []LoadBalancerSummary) string {
	if len(summaries) == 0 {
		return "No load balancers found"
	}

	// Count LBs and healthy/unhealthy targets
	totalTargets := 0
	healthyTargets := 0

	for _, lb := range summaries {
		for _, tg := range lb.TargetGroups {
			totalTargets += len(tg.Targets)

			for _, target := range tg.Targets {
				if target.Status == "healthy" {
					healthyTargets++
				}
			}
		}
	}

	return fmt.Sprintf("%d LBs, %d/%d healthy targets",
		len(summaries),
		healthyTargets,
		totalTargets)
}

// getStatusSymbol returns an appropriate symbol for a health status
func getStatusSymbol(status string) string {
	switch status {
	case "healthy":
		return "✅"
	case "unhealthy":
		return "❌"
	case "draining":
		return "🔄"
	case "unavailable":
		return "⚠️"
	case "initial":
		return "🔍"
	default:
		return "❓"
	}
}
