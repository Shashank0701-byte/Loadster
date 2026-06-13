package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/Shashank0701-byte/Loadster/pkg/metrics"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TickMsg is sent periodically to trigger terminal dashboard redraws.
type TickMsg time.Time

// Dashboard implements tea.Model and represents the interactive CLI visualization.
type Dashboard struct {
	collector   *metrics.Collector
	startTime   time.Time
	totalStages int

	// Current stage stats updated dynamically
	stageNum   int
	targetVUs  int
	stageDur   time.Duration
	stageStart time.Time
}

// NewDashboard creates a new BubbleTea Dashboard instance.
func NewDashboard(collector *metrics.Collector, totalStages int) *Dashboard {
	return &Dashboard{
		collector:   collector,
		startTime:   time.Now(),
		totalStages: totalStages,
	}
}

// UpdateStage updates the dashboard state with new stage variables.
func (d *Dashboard) UpdateStage(stageNum int, targetVUs int, stageDur time.Duration) {
	d.stageNum = stageNum
	d.targetVUs = targetVUs
	d.stageDur = stageDur
	d.stageStart = time.Now()
}

// Init sets up the initial command tick loop.
func (d *Dashboard) Init() tea.Cmd {
	return tick()
}

// Update handles redrawing ticks and exit key presses.
func (d *Dashboard) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return d, tea.Quit
		}
	case TickMsg:
		return d, tick()
	}
	return d, nil
}

// View compiles and renders the dashboard UI elements.
func (d *Dashboard) View() string {
	stats := d.collector.Snapshot()

	// --- Styles ---
	var (
		headerStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(lipgloss.Color("#4F46E5")). // Indigo background
				Padding(0, 2).
				MarginBottom(1)

		panelTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#818CF8")). // Light Indigo
				MarginBottom(1)

		metricBoxStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#4F46E5")).
				Padding(1, 2).
				Width(38).
				Height(7)

		statsBoxStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#10B981")). // Emerald Green border
				Padding(1, 2).
				Width(44).
				Height(9)

		labelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#9CA3AF")).
				Width(18)

		valStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#F3F4F6")).
				Bold(true)

		successStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#10B981")).
				Bold(true)

		warnStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#F59E0B")).
				Bold(true)

		dangerStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#EF4444")).
				Bold(true)
	)

	// --- Header ---
	title := headerStyle.Render(" ⚡ LOADSTER LOAD TESTING CONSOLE ")
	elapsedTotal := time.Since(d.startTime).Round(time.Second)
	headerText := fmt.Sprintf("\n%s  (Total Elapsed: %s)\n", title, elapsedTotal)

	// --- Execution Stage Info ---
	stageElapsed := time.Since(d.stageStart)
	if stageElapsed > d.stageDur {
		stageElapsed = d.stageDur
	}
	percent := 0.0
	if d.stageDur > 0 {
		percent = float64(stageElapsed) / float64(d.stageDur)
	}

	progressStr := progressBar(percent, 25)
	stageInfo := fmt.Sprintf(
		"%s Stage %d of %d\n"+
			"%s %d VUs (Target: %d)\n"+
			"%s %s / %s\n"+
			"%s [%s] %d%%\n",
		labelStyle.Render("Test Stage:"), d.stageNum, d.totalStages,
		labelStyle.Render("Active Users:"), stats.ActiveUsers, d.targetVUs,
		labelStyle.Render("Stage Duration:"), stageElapsed.Round(time.Second), d.stageDur.Round(time.Second),
		labelStyle.Render("Progress:"), progressStr, int(percent*100),
	)

	stagePanel := panelTitleStyle.Render("⚙️  EXECUTION STATUS") + "\n" + stageInfo

	// --- Real-time Metrics ---
	successRateColor := successStyle
	if stats.SuccessRate < 90.0 {
		successRateColor = dangerStyle
	} else if stats.SuccessRate < 95.0 {
		successRateColor = warnStyle
	}

	statsInfo := fmt.Sprintf(
		"%s %s\n"+
			"%s %s\n"+
			"%s %s\n"+
			"%s %s\n"+
			"%s %s\n"+
			"%s %s\n",
		labelStyle.Render("RPS:"), valStyle.Render(fmt.Sprintf("%.2f rps", stats.CurrentRPS)),
		labelStyle.Render("Total Requests:"), valStyle.Render(fmt.Sprintf("%d", stats.TotalRequests)),
		labelStyle.Render("Errors/Failures:"), dangerStyle.Render(fmt.Sprintf("%d", stats.TotalErrors)),
		labelStyle.Render("Success Rate:"), successRateColor.Render(fmt.Sprintf("%.2f%%", stats.SuccessRate)),
		labelStyle.Render("Average Latency:"), valStyle.Render(stats.Average.Round(time.Millisecond).String()),
		labelStyle.Render("Percentiles:"), valStyle.Render(fmt.Sprintf("P50: %s | P95: %s | P99: %s",
			stats.P50.Round(time.Millisecond),
			stats.P95.Round(time.Millisecond),
			stats.P99.Round(time.Millisecond))),
	)

	metricsPanel := panelTitleStyle.Render("📈 PERFORMANCE TELEMETRY") + "\n" + statsInfo

	// --- Layout Generation ---
	leftBox := metricBoxStyle.Render(stagePanel)
	rightBox := statsBoxStyle.Render(metricsPanel)

	dashboardLayout := lipgloss.JoinHorizontal(lipgloss.Top, leftBox, rightBox)

	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#4B5563")).
		Italic(true).
		Render("\nPress [q] or [Ctrl+C] to abort load test execution.")

	return headerText + "\n" + dashboardLayout + "\n" + footer
}

func tick() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

func progressBar(percent float64, width int) string {
	if percent < 0 {
		percent = 0
	} else if percent > 1 {
		percent = 1
	}
	filledWidth := int(percent * float64(width))
	emptyWidth := width - filledWidth
	if emptyWidth < 0 {
		emptyWidth = 0
	}

	filled := strings.Repeat("█", filledWidth)
	empty := strings.Repeat("░", emptyWidth)

	barStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6366F1"))
	emptyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#374151"))

	return barStyle.Render(filled) + emptyStyle.Render(empty)
}
type Dashboard_Type = Dashboard // For linking symbol reference
type TickMsg_Type = TickMsg // For linking symbol reference
