package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/correctedcloud/aws-overview/internal/ui"
)

func main() {
	// Parse command line flags
	var showALB bool
	var showRDS bool
	var showEC2 bool
	var showECS bool
	var region string

	flag.BoolVar(&showALB, "alb", false, "Show ALB resources")
	flag.BoolVar(&showRDS, "rds", false, "Show RDS resources")
	flag.BoolVar(&showEC2, "ec2", false, "Show EC2 resources")
	flag.BoolVar(&showECS, "ecs", false, "Show ECS services")
	flag.StringVar(&region, "region", "", "AWS region (defaults to AWS_REGION env var)")
	flag.Parse()

	// Check if at least one resource type is selected
	if !showALB && !showRDS && !showEC2 && !showECS {
		// Default to showing all resource types if none specified
		showALB = true
		showRDS = true
		showEC2 = true
		showECS = true
	}

	// Create the UI model
	m := ui.NewModel(showALB, showRDS, showEC2, showECS, region)

	// Initialize the terminal UI
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running UI: %v\n", err)
		os.Exit(1)
	}
}