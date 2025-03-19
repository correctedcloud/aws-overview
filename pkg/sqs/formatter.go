package sqs

import (
	"fmt"
	"strings"

	"github.com/correctedcloud/aws-overview/pkg/common"
)

// FormatQueues formats queue summaries for terminal display
func FormatQueues(summaries []QueueSummary) string {
	if len(summaries) == 0 {
		return "No SQS queues found"
	}

	var output strings.Builder
	output.WriteString("SQS QUEUES\n")
	output.WriteString("==========\n\n")

	for _, queue := range summaries {
		queueTypeSymbol := getQueueTypeSymbol(queue.Type)
		output.WriteString(fmt.Sprintf("%s %s (%s)\n", queueTypeSymbol, queue.Name, queue.Type))

		output.WriteString("\n  Messages Sent (1 hour):\n")
		if len(queue.SentMessages) > 0 {
			sentGraph := common.GenerateSparkline(queue.SentMessages, "Messages Sent", 3)
			output.WriteString(fmt.Sprintf("%s\n", sentGraph))
		} else {
			output.WriteString("  No message sent data available\n")
		}

		output.WriteString("\n  Visible Messages (1 hour):\n")
		if len(queue.VisibleMessages) > 0 {
			visibleGraph := common.GenerateSparkline(queue.VisibleMessages, "Visible Messages", 3)
			output.WriteString(fmt.Sprintf("%s\n", visibleGraph))
		} else {
			output.WriteString("  No visible message data available\n")
		}

		output.WriteString("\n")
	}

	return output.String()
}

// GetQueuesSummary returns a brief summary of SQS queues
func GetQueuesSummary(summaries []QueueSummary) string {
	if len(summaries) == 0 {
		return "No SQS queues found"
	}

	// Count queues by type
	standard := 0
	fifo := 0

	// Calculate average sent and visible messages
	totalSent := 0.0
	sentDataPoints := 0
	totalVisible := 0.0
	visibleDataPoints := 0

	for _, queue := range summaries {
		if queue.Type == "FIFO" {
			fifo++
		} else {
			standard++
		}

		// Add the last sent messages data point if available
		if len(queue.SentMessages) > 0 {
			totalSent += queue.SentMessages[len(queue.SentMessages)-1]
			sentDataPoints++
		}

		// Add the last visible messages data point if available
		if len(queue.VisibleMessages) > 0 {
			totalVisible += queue.VisibleMessages[len(queue.VisibleMessages)-1]
			visibleDataPoints++
		}
	}

	// Calculate averages
	var sentAvg, visibleAvg float64
	if sentDataPoints > 0 {
		sentAvg = totalSent / float64(sentDataPoints)
	}
	if visibleDataPoints > 0 {
		visibleAvg = totalVisible / float64(visibleDataPoints)
	}

	return fmt.Sprintf("%d queues (%d standard, %d FIFO), Recent Avg Sent: %.1f, Recent Avg Visible: %.1f",
		len(summaries),
		standard,
		fifo,
		sentAvg,
		visibleAvg)
}

// getQueueTypeSymbol returns an appropriate symbol for a queue type
func getQueueTypeSymbol(queueType string) string {
	switch queueType {
	case "FIFO":
		return "🔄" // Shows ordered processing
	case "Standard":
		return "📬" // Regular mailbox
	default:
		return "❓"
	}
}
