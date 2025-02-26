# AWS Overview

A Go application designed to give a quick overview of an AWS environment with a modern terminal UI.

It provides small graph/sparkline overviews for various AWS services and relevant information in an interactive interface.

## Services

### Autoscaling/Load Balancing

- Shows the health status for each target group, grouped by load balancer

### RDS

- Shows the CPU and memory usage over the past 1 hour for each RDS instance
- Shows any recent errors in the DB error log

## Features

- Interactive terminal UI with tabs
- Parallel data fetching for quick information retrieval
- Visual sparkline graphs for numeric metrics
- Color-coded status indicators

## Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/aws-overview.git
cd aws-overview

# Build the application
go build -o aws-overview ./cmd/aws-overview

# Move to a location in your PATH (optional)
sudo mv aws-overview /usr/local/bin/
```

## Usage

```bash
# Use default AWS configuration and show all information
aws-overview

# Specify a region
aws-overview -region us-west-2

# Show only ALB information
aws-overview -rds=false

# Show only RDS information
aws-overview -alb=false

# Get help
aws-overview -h
```

### Terminal UI Navigation

- Use `Tab`, `Right Arrow`, or `l` to move to the next tab
- Use `Shift+Tab`, `Left Arrow`, or `h` to move to the previous tab
- Press `q` or `Ctrl+C` to quit the application

## AWS Credentials

This application uses the AWS SDK for Go v2, which will look for credentials in the following order:

1. Environment variables (`AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`)
2. Shared credential file (`~/.aws/credentials`)
3. EC2 Instance Profile (if running on an EC2 instance)

Make sure you have valid AWS credentials configured before using the application.

## Development

### Requirements

- Go 1.23 or higher
- AWS account with appropriate permissions

### Dependencies

- [bubbletea](https://github.com/charmbracelet/bubbletea) - Terminal UI framework
- [lipgloss](https://github.com/charmbracelet/lipgloss) - Styles for terminal applications
- [AWS SDK for Go v2](https://github.com/aws/aws-sdk-go-v2) - AWS API client

### Running Tests

```bash
go test -v ./...
```

## License

MIT