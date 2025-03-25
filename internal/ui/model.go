package ui

import (
	"os"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/correctedcloud/aws-overview/pkg/alb"
	"github.com/correctedcloud/aws-overview/pkg/ec2"
	"github.com/correctedcloud/aws-overview/pkg/ecs"
	"github.com/correctedcloud/aws-overview/pkg/rds"
	"github.com/correctedcloud/aws-overview/pkg/sqs"
)

// Color scheme for the UI
var (
	// Define color palette
	primaryColor    = lipgloss.Color("#7D56F4") // Vibrant purple
	secondaryColor  = lipgloss.Color("#5AD4E6") // Bright cyan
	accentColor     = lipgloss.Color("#FFB938") // Warm amber
	errorColor      = lipgloss.Color("#FF5F87") // Soft red
	successColor    = lipgloss.Color("#39DA8A") // Vibrant green
	warningColor    = lipgloss.Color("#FFBD54") // Amber
	backgroundColor = lipgloss.Color("#1A1B26") // Dark background
	textColor       = lipgloss.Color("#FAFAFA") // Light text
	dimTextColor    = lipgloss.Color("#9699B7") // Dimmed text
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(textColor).
			Background(primaryColor).
			Padding(0, 2).
			Margin(0, 0, 0, 0).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor)

	tabStyle = lipgloss.NewStyle().
			Foreground(textColor).
			Background(dimTextColor).
			Padding(0, 2).
			Margin(0, 1, 0, 0).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderBottom(true).
			BorderForeground(dimTextColor)

	activeTabStyle = lipgloss.NewStyle().
			Foreground(textColor).
			Background(accentColor).
			Padding(0, 2).
			Margin(0, 1, 0, 0).
			Bold(true).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderBottom(true).
			BorderForeground(accentColor)

	contentStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(secondaryColor).
			Margin(1, 0, 0, 0).
			Padding(1, 2)

	tabGap = lipgloss.NewStyle().Padding(0, 1)

	headerStyle = lipgloss.NewStyle().
			Foreground(textColor).
			Background(backgroundColor).
			Width(100).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Margin(0, 0, 1, 0).
			Padding(0, 2).
			Bold(true)
)

// Model is the main UI model
type Model struct {
	spinner       spinner.Model
	viewport      viewport.Model
	loadingALB    bool
	loadingRDS    bool
	loadingEC2    bool
	loadingECS    bool
	loadingSQS    bool
	loadBalancers []alb.LoadBalancerSummary
	dbInstances   []rds.DBInstanceSummary
	ec2Instances  []ec2.InstanceSummary
	ecsServices   []ecs.ServiceSummary
	sqsQueues     []sqs.QueueSummary
	albErr        error
	rdsErr        error
	ec2Err        error
	ecsErr        error
	sqsErr        error
	width         int
	height        int
	showALB       bool
	showRDS       bool
	showEC2       bool
	showECS       bool
	showSQS       bool
	region        string
	activeTab     int
	tabs          []string
	lastRefresh   time.Time
}

// NewModel creates a new UI model
func NewModel(showALB, showRDS, showEC2, showECS, showSQS bool, region string) Model {
	// Create tabs list
	tabs := []string{"Overview"}
	if showALB {
		tabs = append(tabs, "Load Balancers")
	}
	if showRDS {
		tabs = append(tabs, "RDS Instances")
	}
	if showEC2 {
		tabs = append(tabs, "EC2 Instances")
	}
	if showECS {
		tabs = append(tabs, "ECS Services")
	}
	if showSQS {
		tabs = append(tabs, "SQS Queues")
	}

	// Create a fancier spinner with custom styling
	s := spinner.New()
	s.Spinner = spinner.MiniDot
	s.Style = lipgloss.NewStyle().Foreground(accentColor).Bold(true)

	// Initialize viewport with default size (will be adjusted when window size is known)
	vp := viewport.New(80, 20)

	return Model{
		spinner:     s,
		viewport:    vp,
		loadingALB:  showALB,
		loadingRDS:  showRDS,
		loadingEC2:  showEC2,
		loadingECS:  showECS,
		loadingSQS:  showSQS,
		showALB:     showALB,
		showRDS:     showRDS,
		showEC2:     showEC2,
		showECS:     showECS,
		showSQS:     showSQS,
		region:      region,
		activeTab:   0,
		tabs:        tabs,
		lastRefresh: time.Now(),
	}
}

// Init initializes the model and triggers data loading
func (m Model) Init() tea.Cmd {
	cmds := []tea.Cmd{
		m.spinner.Tick,
		refreshTimer(),
	}

	if m.showALB {
		cmds = append(cmds, m.loadALBData())
	}

	if m.showRDS {
		cmds = append(cmds, m.loadRDSData())
	}

	if m.showEC2 {
		cmds = append(cmds, m.loadEC2Data())
	}

	if m.showECS {
		cmds = append(cmds, m.loadECSData())
	}

	if m.showSQS {
		cmds = append(cmds, m.loadSQSData())
	}

	return tea.Batch(cmds...)
}

// Update handles various events and messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Let viewport handle keys first if not a tab-switching key
		if msg.String() != "tab" && msg.String() != "right" && msg.String() != "l" &&
			msg.String() != "shift+tab" && msg.String() != "left" && msg.String() != "h" &&
			msg.String() != "q" && msg.String() != "ctrl+c" {
			var cmd tea.Cmd
			m.viewport, cmd = m.viewport.Update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
				break
			}
		}

		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "tab", "right", "l":
			// Cycle to next tab
			m.activeTab = (m.activeTab + 1) % len(m.tabs)
			// Update content for the new tab
			m.updateViewportContent()
		case "shift+tab", "left", "h":
			// Cycle to previous tab
			m.activeTab = (m.activeTab - 1 + len(m.tabs)) % len(m.tabs)
			// Update content for the new tab
			m.updateViewportContent()
		case "r": // Manual refresh
			cmds = append(cmds, m.refreshData())
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Update viewport height and width
		headerHeight := 12                                             // Increased space for header elements
		footerHeight := 1                                              // Help text
		m.viewport.Width = m.width - 4                                 // Account for padding
		m.viewport.Height = m.height - headerHeight - footerHeight - 2 // Account for margins

		// Update content for the viewport with the new dimensions
		m.updateViewportContent()

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)

	case refreshTimerMsg:
		// Update last refresh time
		m.lastRefresh = time.Now()

		// Start data refresh
		if !m.loadingALB && !m.loadingRDS && !m.loadingEC2 && !m.loadingECS && !m.loadingSQS {
			cmds = append(cmds, m.refreshData())
		}

		// Schedule next refresh
		cmds = append(cmds, refreshTimer())

	case albDataLoadedMsg:
		m.loadingALB = false
		m.loadBalancers = msg.loadBalancers
		m.albErr = msg.err
		// Update region if it was empty and we got it from AWS config
		if m.region == "" && msg.region != "" {
			m.region = msg.region
		}
		m.updateViewportContent()

	case rdsDataLoadedMsg:
		m.loadingRDS = false
		m.dbInstances = msg.dbInstances
		m.rdsErr = msg.err
		// Update region if it was empty and we got it from AWS config
		if m.region == "" && msg.region != "" {
			m.region = msg.region
		}
		m.updateViewportContent()

	case ec2DataLoadedMsg:
		m.loadingEC2 = false
		m.ec2Instances = msg.instances
		m.ec2Err = msg.err
		// Update region if it was empty and we got it from AWS config
		if m.region == "" && msg.region != "" {
			m.region = msg.region
		}
		m.updateViewportContent()

	case ecsDataLoadedMsg:
		m.loadingECS = false
		m.ecsServices = msg.services
		m.ecsErr = msg.err
		// Update region if it was empty and we got it from AWS config
		if m.region == "" && msg.region != "" {
			m.region = msg.region
		}
		m.updateViewportContent()

	case sqsDataLoadedMsg:
		m.loadingSQS = false
		m.sqsQueues = msg.queues
		m.sqsErr = msg.err
		// Update region if it was empty and we got it from AWS config
		if m.region == "" && msg.region != "" {
			m.region = msg.region
		}
		m.updateViewportContent()
	}

	return m, tea.Batch(cmds...)
}

// updateViewportContent updates the viewport content based on the active tab
func (m *Model) updateViewportContent() {
	var content string

	switch {
	case m.activeTab == 0: // Overview tab
		content = m.renderOverview()
	case m.activeTab == 1 && m.showALB: // Load Balancers tab
		content = m.renderALB()
	case (m.activeTab == 1 && !m.showALB && m.showRDS) || (m.activeTab == 2 && m.showALB && m.showRDS): // RDS tab
		content = m.renderRDS()
	case (m.activeTab == 1 && !m.showALB && !m.showRDS && m.showEC2) ||
		(m.activeTab == 2 && !m.showALB && m.showEC2) ||
		(m.activeTab == 2 && !m.showRDS && m.showEC2) ||
		(m.activeTab == 3 && m.showALB && m.showRDS && m.showEC2): // EC2 tab
		content = m.renderEC2()
	case (m.activeTab == 1 && !m.showALB && !m.showRDS && !m.showEC2 && m.showECS) ||
		(m.activeTab == 2 && !m.showALB && !m.showRDS && m.showECS) ||
		(m.activeTab == 2 && !m.showALB && !m.showEC2 && m.showECS) ||
		(m.activeTab == 2 && !m.showRDS && !m.showEC2 && m.showECS) ||
		(m.activeTab == 3 && !m.showALB && m.showECS) ||
		(m.activeTab == 3 && !m.showRDS && m.showECS) ||
		(m.activeTab == 3 && !m.showEC2 && m.showECS) ||
		(m.activeTab == 4 && m.showALB && m.showRDS && m.showEC2 && m.showECS): // ECS tab
		content = m.renderECS()
	case m.activeTab >= 1 && m.activeTab <= 5 && m.showSQS &&
		((m.activeTab == len(m.tabs)-1) || m.tabs[m.activeTab] == "SQS Queues"): // SQS tab
		content = m.renderSQS()
	}

	// Set the content for scrolling
	m.viewport.SetContent(content)
}

// View renders the UI
func (m Model) View() string {
	// Generate tabs with prominent styling
	var renderedTabs []string
	for i, t := range m.tabs {
		if i == m.activeTab {
			renderedTabs = append(renderedTabs, activeTabStyle.Render(t))
		} else {
			renderedTabs = append(renderedTabs, tabStyle.Render(t))
		}
	}
	tabBar := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)

	// Make tab bar more prominent
	tabBar = lipgloss.NewStyle().Margin(0, 0, 1, 0).Render(tabBar)

	// Use viewport for scrollable content
	viewportContent := m.viewport.View()

	// Apply content styling with proper border rendering using full width
	contentStyleCopy := contentStyle.Copy().Width(m.width - 4) // Subtract padding
	styledContent := contentStyleCopy.Render(viewportContent)

	// Show help text at the bottom
	helpText := lipgloss.NewStyle().
		Foreground(dimTextColor).
		Background(backgroundColor).
		Bold(true).
		Padding(0, 2).
		Margin(1, 0, 0, 0).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(primaryColor).
		Render("← → Navigate Tabs • ↑↓/j k Scroll • r Refresh • q Quit")

	// Force tabs to top of screen with no margins above
	header := lipgloss.JoinVertical(
		lipgloss.Left,
		tabBar,
	)

	// Ensure content has adequate spacing from header
	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		styledContent,
		helpText,
	)
}

// getRegionFlag returns the flag emoji for a given AWS region
func getRegionFlag(region string) string {
	// Map AWS regions to flag emoji with location suffix
	regionToFlag := map[string]string{
		// North America
		"us-east-1":     "🇺🇸",
		"us-east-2":     "🇺🇸",
		"us-west-1":     "🇺🇸",
		"us-west-2":     "🇺🇸",
		"us-gov-east-1": "🇺🇸",
		"us-gov-west-1": "🇺🇸",
		"mx-central-1":  "🇲🇽",

		// South America
		"sa-east-1": "🇧🇷",

		// Europe
		"eu-west-1":    "🇮🇪",
		"eu-west-2":    "🇬🇧",
		"eu-west-3":    "🇫🇷",
		"eu-central-1": "🇩🇪",
		"eu-central-2": "🇨🇭",
		"eu-south-1":   "🇮🇹",
		"eu-south-2":   "🇪🇸",
		"eu-north-1":   "🇸🇪",

		// Middle East
		"me-central-1": "🇦🇪",
		"me-south-1":   "🇧🇭",
		"il-central-1": "🇮🇱",

		// Asia Pacific
		"ap-southeast-1": "🇸🇬",
		"ap-southeast-2": "🇦🇺",
		"ap-southeast-3": "🇸🇬",
		"ap-southeast-4": "🇦🇺",
		"ap-southeast-5": "🇳🇿",
		"ap-southeast-7": "🇹🇭",
		"ap-east-1":      "🇭🇰",
		"ap-south-1":     "🇮🇳",
		"ap-south-2":     "🇮🇳",
		"ap-northeast-1": "🇯🇵",
		"ap-northeast-2": "🇰🇷",
		"ap-northeast-3": "🇯🇵",

		// Canada
		"ca-central-1": "🇨🇦",
		"ca-west-1":    "🇨🇦",

		// Africa
		"af-south-1": "🇿🇦",

		// China
		"cn-north-1":     "🇨🇳",
		"cn-northwest-1": "🇨🇳",
	}

	flag, ok := regionToFlag[region]
	if !ok {
		return "🌐" // Default global symbol if region not found
	}

	return flag
}

// getAWSProfile returns the current AWS profile from environment variables
func getAWSProfile() string {
	profile := os.Getenv("AWS_PROFILE")
	if profile == "" {
		profile = os.Getenv("AWS_DEFAULT_PROFILE")
	}
	return profile
}

// renderOverview shows a summary view
func (m Model) renderOverview() string {
	if (m.loadingALB && m.showALB) || (m.loadingRDS && m.showRDS) || (m.loadingEC2 && m.showEC2) {
		return m.spinner.View() + " Loading AWS resources..."
	}

	var content string
	flag := getRegionFlag(m.region)
	content += lipgloss.NewStyle().Foreground(accentColor).Bold(true).Render("Region: "+flag+" "+m.region) + "\n"

	// Display AWS profile if set
	profile := getAWSProfile()
	if profile != "" {
		content += lipgloss.NewStyle().Foreground(secondaryColor).Bold(true).Render("Profile: "+profile) + "\n"
	}

	// Display last refresh time
	content += lipgloss.NewStyle().Foreground(dimTextColor).Render("Last refresh: "+m.lastRefresh.Format("15:04:05")+" (auto-refreshes every minute)") + "\n\n"

	if m.showALB {
		if m.albErr != nil {
			content += lipgloss.NewStyle().Foreground(errorColor).Bold(true).Render("❌ Load Balancer Error: ") +
				lipgloss.NewStyle().Foreground(errorColor).Render(m.albErr.Error()) + "\n\n"
		} else {
			content += lipgloss.NewStyle().Foreground(successColor).Bold(true).Render("✅ Load Balancers: ") +
				lipgloss.NewStyle().Foreground(textColor).Render(alb.GetLoadBalancersSummary(m.loadBalancers)) + "\n\n"
		}
	}

	if m.showRDS {
		if m.rdsErr != nil {
			content += lipgloss.NewStyle().Foreground(errorColor).Bold(true).Render("❌ RDS Error: ") +
				lipgloss.NewStyle().Foreground(errorColor).Render(m.rdsErr.Error()) + "\n\n"
		} else {
			content += lipgloss.NewStyle().Foreground(successColor).Bold(true).Render("✅ RDS Instances: ") +
				lipgloss.NewStyle().Foreground(textColor).Render(rds.GetDBInstancesSummary(m.dbInstances)) + "\n\n"
		}
	}

	if m.showEC2 {
		if m.ec2Err != nil {
			content += lipgloss.NewStyle().Foreground(errorColor).Bold(true).Render("❌ EC2 Error: ") +
				lipgloss.NewStyle().Foreground(errorColor).Render(m.ec2Err.Error()) + "\n\n"
		} else {
			content += lipgloss.NewStyle().Foreground(successColor).Bold(true).Render("✅ EC2 Instances: ") +
				lipgloss.NewStyle().Foreground(textColor).Render(ec2.GetInstancesSummary(m.ec2Instances)) + "\n\n"
		}
	}

	if m.showECS {
		if m.ecsErr != nil {
			content += lipgloss.NewStyle().Foreground(errorColor).Bold(true).Render("❌ ECS Error: ") +
				lipgloss.NewStyle().Foreground(errorColor).Render(m.ecsErr.Error()) + "\n\n"
		} else {
			content += lipgloss.NewStyle().Foreground(successColor).Bold(true).Render("✅ ECS Services: ") +
				lipgloss.NewStyle().Foreground(textColor).Render(ecs.GetServicesSummary(m.ecsServices)) + "\n\n"
		}
	}

	if m.showSQS {
		if m.sqsErr != nil {
			content += lipgloss.NewStyle().Foreground(errorColor).Bold(true).Render("❌ SQS Error: ") +
				lipgloss.NewStyle().Foreground(errorColor).Render(m.sqsErr.Error()) + "\n\n"
		} else {
			content += lipgloss.NewStyle().Foreground(successColor).Bold(true).Render("✅ SQS Queues: ") +
				lipgloss.NewStyle().Foreground(textColor).Render(sqs.GetQueuesSummary(m.sqsQueues)) + "\n\n"
		}
	}

	if !m.showALB && !m.showRDS && !m.showEC2 && !m.showECS && !m.showSQS {
		content += "No services selected. Use -alb=true, -rds=true, -ec2=true, and/or -ecs=true flags."
	}

	return content
}

// renderALB shows detailed ALB information
func (m Model) renderALB() string {
	if m.loadingALB {
		return m.spinner.View() + " Loading ALB data..."
	}

	if m.albErr != nil {
		return "Error loading ALB data: " + m.albErr.Error()
	}

	return alb.FormatLoadBalancers(m.loadBalancers)
}

// renderRDS shows detailed RDS information
func (m Model) renderRDS() string {
	if m.loadingRDS {
		return m.spinner.View() + " Loading RDS data..."
	}

	if m.rdsErr != nil {
		return "Error loading RDS data: " + m.rdsErr.Error()
	}

	return rds.FormatDBInstances(m.dbInstances)
}

// renderEC2 shows detailed EC2 information
func (m Model) renderEC2() string {
	if m.loadingEC2 {
		return m.spinner.View() + " Loading EC2 data..."
	}

	if m.ec2Err != nil {
		return "Error loading EC2 data: " + m.ec2Err.Error()
	}

	return ec2.FormatInstances(m.ec2Instances)
}

// renderECS shows detailed ECS information
func (m Model) renderECS() string {
	if m.loadingECS {
		return m.spinner.View() + " Loading ECS data..."
	}

	if m.ecsErr != nil {
		return "Error loading ECS data: " + m.ecsErr.Error()
	}

	return ecs.FormatServices(m.ecsServices)
}

// renderSQS shows detailed SQS information
func (m Model) renderSQS() string {
	if m.loadingSQS {
		return m.spinner.View() + " Loading SQS data..."
	}

	if m.sqsErr != nil {
		return "Error loading SQS data: " + m.sqsErr.Error()
	}

	return sqs.FormatQueues(m.sqsQueues)
}
