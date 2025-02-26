package ui

import (
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/correctedcloud/aws-overview/pkg/alb"
	"github.com/correctedcloud/aws-overview/pkg/rds"
)

var (
	titleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1)

	tabStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#888888")).
		Padding(0, 1)

	activeTabStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1).
		Bold(true)

	contentStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(1, 2) // MaxWidth is applied in View() method to ensure borders render properly
		
	tabGap = lipgloss.NewStyle().Padding(0, 1)
)

// Model is the main UI model
type Model struct {
	spinner         spinner.Model
	viewport        viewport.Model
	loadingALB      bool
	loadingRDS      bool
	loadBalancers   []alb.LoadBalancerSummary
	dbInstances     []rds.DBInstanceSummary
	albErr          error
	rdsErr          error
	width           int
	height          int
	showALB         bool
	showRDS         bool
	region          string
	activeTab       int
	tabs            []string
}

// NewModel creates a new UI model
func NewModel(showALB, showRDS bool, region string) Model {
	// Create tabs list
	tabs := []string{"Overview"}
	if showALB {
		tabs = append(tabs, "Load Balancers")
	}
	if showRDS {
		tabs = append(tabs, "RDS Instances")
	}

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4"))

	// Initialize viewport with default size (will be adjusted when window size is known)
	vp := viewport.New(80, 20)

	return Model{
		spinner:    s,
		viewport:   vp,
		loadingALB: showALB,
		loadingRDS: showRDS,
		showALB:    showALB,
		showRDS:    showRDS,
		region:     region,
		activeTab:  0,
		tabs:       tabs,
	}
}

// Init initializes the model and triggers data loading
func (m Model) Init() tea.Cmd {
	cmds := []tea.Cmd{
		m.spinner.Tick,
	}

	if m.showALB {
		cmds = append(cmds, m.loadALBData())
	}

	if m.showRDS {
		cmds = append(cmds, m.loadRDSData())
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
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		
		// Update viewport height and width
		headerHeight := 2 // Title + tabs
		footerHeight := 1 // Help text
		m.viewport.Width = m.width - 4  // Account for padding
		m.viewport.Height = m.height - headerHeight - footerHeight - 2 // Account for margins
		
		// Update content for the viewport with the new dimensions
		m.updateViewportContent()

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)

	case albDataLoadedMsg:
		m.loadingALB = false
		m.loadBalancers = msg.loadBalancers
		m.albErr = msg.err
		m.updateViewportContent()

	case rdsDataLoadedMsg:
		m.loadingRDS = false
		m.dbInstances = msg.dbInstances
		m.rdsErr = msg.err
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
	case (m.activeTab == 1 && !m.showALB && m.showRDS) || (m.activeTab == 2 && m.showRDS): // RDS tab
		content = m.renderRDS()
	}

	// Set the content for scrolling
	m.viewport.SetContent(content)
}

// View renders the UI
func (m Model) View() string {
	// Generate tabs
	var renderedTabs []string
	for i, t := range m.tabs {
		if i == m.activeTab {
			renderedTabs = append(renderedTabs, activeTabStyle.Render(t))
		} else {
			renderedTabs = append(renderedTabs, tabStyle.Render(t))
		}
	}
	tabBar := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)

	// Calculate appropriate width based on terminal size
	contentWidth := m.width - 4
	if contentWidth > 200 {
		contentWidth = 200 // Limit maximum width for readability
	}

	// Use viewport for scrollable content
	viewportContent := m.viewport.View()

	// Apply content styling with proper border rendering
	// Use copy of style to avoid modifying the original
	contentStyleCopy := contentStyle.Copy().Width(contentWidth)
	styledContent := contentStyleCopy.Render(viewportContent)

	// Show help text at the bottom
	helpText := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888")).
		Render("← → Navigate Tabs • ↑↓/j k Scroll • q Quit")

	// Ensure title and tabs are visible at the top with clear separation
	title := titleStyle.Render("AWS Overview")
	header := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		tabBar,
	)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		styledContent,
		helpText,
	)
}

// renderOverview shows a summary view
func (m Model) renderOverview() string {
	if (m.loadingALB && m.showALB) || (m.loadingRDS && m.showRDS) {
		return m.spinner.View() + " Loading AWS resources..."
	}

	var content string
	content += "Region: " + m.region + "\n\n"

	if m.showALB {
		if m.albErr != nil {
			content += "❌ Load Balancer Error: " + m.albErr.Error() + "\n\n"
		} else {
			content += "✅ Load Balancers: " + alb.GetLoadBalancersSummary(m.loadBalancers) + "\n\n"
		}
	}

	if m.showRDS {
		if m.rdsErr != nil {
			content += "❌ RDS Error: " + m.rdsErr.Error() + "\n\n"
		} else {
			content += "✅ RDS Instances: " + rds.GetDBInstancesSummary(m.dbInstances) + "\n\n"
		}
	}

	if !m.showALB && !m.showRDS {
		content += "No services selected. Use -alb=true and/or -rds=true flags."
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