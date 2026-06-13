package report

import (
	"encoding/json"
	"fmt"
	"os"
	"text/template"
	"time"

	"github.com/Shashank0701-byte/Loadster/pkg/metrics"
)

// ReportData represents the schema of the generated load testing report.
type ReportData struct {
	Target            string    `json:"target"`
	TotalRequests     int64     `json:"total_requests"`
	TotalErrors       int64     `json:"total_errors"`
	SuccessRate       float64   `json:"success_rate"`
	AverageLatencyMs  float64   `json:"average_latency_ms"`
	P50LatencyMs      float64   `json:"p50_latency_ms"`
	P90LatencyMs      float64   `json:"p90_latency_ms"`
	P95LatencyMs      float64   `json:"p95_latency_ms"`
	P99LatencyMs      float64   `json:"p99_latency_ms"`
	TestDurationSec   float64   `json:"test_duration_seconds"`
	ThroughputRps     float64   `json:"throughput_rps"`
	RawLatenciesMs    []float64 `json:"raw_latencies_ms"`
}

// BuildReportData maps metrics collector snapshots and raw latencies into ReportData.
func BuildReportData(target string, duration time.Duration, collector *metrics.Collector) *ReportData {
	stats := collector.Snapshot()
	latencies := collector.GetLatencies()

	rawMs := make([]float64, len(latencies))
	for i, l := range latencies {
		rawMs[i] = float64(l) / float64(time.Millisecond)
	}

	throughput := 0.0
	if duration.Seconds() > 0 {
		throughput = float64(stats.TotalRequests) / duration.Seconds()
	}

	return &ReportData{
		Target:           target,
		TotalRequests:    stats.TotalRequests,
		TotalErrors:      stats.TotalErrors,
		SuccessRate:      stats.SuccessRate,
		AverageLatencyMs: float64(stats.Average) / float64(time.Millisecond),
		P50LatencyMs:     float64(stats.P50) / float64(time.Millisecond),
		P90LatencyMs:     float64(stats.P90) / float64(time.Millisecond),
		P95LatencyMs:     float64(stats.P95) / float64(time.Millisecond),
		P99LatencyMs:     float64(stats.P99) / float64(time.Millisecond),
		TestDurationSec:  duration.Seconds(),
		ThroughputRps:    throughput,
		RawLatenciesMs:   rawMs,
	}
}

// WriteJSON exports report data to a JSON file.
func WriteJSON(data *ReportData, filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create JSON file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("failed to encode report JSON: %w", err)
	}
	return nil
}

// WriteHTML renders report data into a standalone, styled HTML visualization file.
func WriteHTML(data *ReportData, filePath string) error {
	tmpl, err := template.New("report").Parse(htmlTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse HTML template: %w", err)
	}

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create HTML file: %w", err)
	}
	defer file.Close()

	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("failed to execute HTML template: %w", err)
	}
	return nil
}

const htmlTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Loadster Test Report</title>
    <link href="https://fonts.googleapis.com/css2?family=Outfit:wght@300;400;600;800&display=swap" rel="stylesheet">
    <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
    <style>
        :root {
            --bg-primary: #0F172A;
            --bg-secondary: #1E293B;
            --accent-blue: #6366F1;
            --accent-green: #10B981;
            --accent-red: #EF4444;
            --text-primary: #F8FAFC;
            --text-secondary: #94A3B8;
        }

        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }

        body {
            font-family: 'Outfit', sans-serif;
            background-color: var(--bg-primary);
            color: var(--text-primary);
            line-height: 1.6;
            padding: 3rem 1rem;
        }

        .container {
            max-width: 1100px;
            margin: 0 auto;
        }

        header {
            text-align: center;
            margin-bottom: 4rem;
            position: relative;
        }

        h1 {
            font-size: 3rem;
            font-weight: 800;
            background: linear-gradient(135deg, #818CF8, #34D399);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            margin-bottom: 0.5rem;
        }

        .subtitle {
            color: var(--text-secondary);
            font-size: 1.2rem;
            font-weight: 300;
            display: flex;
            align-items: center;
            justify-content: center;
        }

        .summary-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
            gap: 1.5rem;
            margin-bottom: 3rem;
        }

        .card {
            background: rgba(30, 41, 59, 0.7);
            backdrop-filter: blur(12px);
            -webkit-backdrop-filter: blur(12px);
            border: 1px solid rgba(255, 255, 255, 0.1);
            border-radius: 16px;
            padding: 1.5rem;
            box-shadow: 0 10px 30px rgba(0, 0, 0, 0.2);
            transition: transform 0.3s ease, border-color 0.3s ease;
        }

        .card:hover {
            transform: translateY(-5px);
            border-color: var(--accent-blue);
        }

        .card-title {
            font-size: 0.9rem;
            text-transform: uppercase;
            letter-spacing: 0.05em;
            color: var(--text-secondary);
            margin-bottom: 0.5rem;
        }

        .card-value {
            font-size: 2.2rem;
            font-weight: 600;
            word-break: break-all;
        }

        .card-value.success {
            color: var(--accent-green);
        }

        .card-value.danger {
            color: var(--accent-red);
        }

        .chart-section {
            display: grid;
            grid-template-columns: 1fr 1fr;
            gap: 2rem;
            margin-bottom: 3rem;
        }

        @media (max-width: 768px) {
            .chart-section {
                grid-template-columns: 1fr;
            }
        }

        .chart-card {
            background: var(--bg-secondary);
            border-radius: 16px;
            padding: 2rem;
            border: 1px solid rgba(255, 255, 255, 0.05);
            min-height: 400px;
        }

        .chart-title {
            font-size: 1.2rem;
            font-weight: 600;
            margin-bottom: 1.5rem;
            color: var(--text-primary);
            border-left: 4px solid var(--accent-blue);
            padding-left: 0.5rem;
        }

        .table-section {
            background: var(--bg-secondary);
            border-radius: 16px;
            padding: 2rem;
            border: 1px solid rgba(255, 255, 255, 0.05);
        }

        .table-title {
            font-size: 1.2rem;
            font-weight: 600;
            margin-bottom: 1.5rem;
            border-left: 4px solid var(--accent-green);
            padding-left: 0.5rem;
        }

        table {
            width: 100%;
            border-collapse: collapse;
            text-align: left;
        }

        th, td {
            padding: 1rem;
            border-bottom: 1px solid rgba(255, 255, 255, 0.1);
        }

        th {
            color: var(--text-secondary);
            font-weight: 600;
            text-transform: uppercase;
            font-size: 0.85rem;
            letter-spacing: 0.05em;
        }

        td {
            font-size: 1rem;
        }

        .pulsing-indicator {
            display: inline-block;
            width: 10px;
            height: 10px;
            background-color: var(--accent-green);
            border-radius: 50%;
            margin-right: 0.6rem;
            box-shadow: 0 0 0 0 rgba(16, 185, 129, 0.7);
            animation: pulse 1.5s infinite;
        }

        @keyframes pulse {
            0% {
                transform: scale(0.95);
                box-shadow: 0 0 0 0 rgba(16, 185, 129, 0.7);
            }
            70% {
                transform: scale(1);
                box-shadow: 0 0 0 6px rgba(16, 185, 129, 0);
            }
            100% {
                transform: scale(0.95);
                box-shadow: 0 0 0 0 rgba(16, 185, 129, 0);
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <header>
            <h1>LOADSTER REPORT</h1>
            <div class="subtitle"><span class="pulsing-indicator"></span>Target: {{.Target}}</div>
        </header>

        <section class="summary-grid">
            <div class="card">
                <div class="card-title">Total Requests</div>
                <div class="card-value">{{.TotalRequests}}</div>
            </div>
            <div class="card">
                <div class="card-title">Success Rate</div>
                <div class="card-value success">{{printf "%.2f" .SuccessRate}}%</div>
            </div>
            <div class="card">
                <div class="card-title">Throughput</div>
                <div class="card-value">{{printf "%.2f" .ThroughputRps}} RPS</div>
            </div>
            <div class="card">
                <div class="card-title">Duration</div>
                <div class="card-value">{{printf "%.1f" .TestDurationSec}}s</div>
            </div>
        </section>

        <section class="chart-section">
            <div class="chart-card">
                <div class="chart-title">Latency Summary (ms)</div>
                <canvas id="latencySummaryChart"></canvas>
            </div>
            <div class="chart-card">
                <div class="chart-title">Latency Distribution Histogram</div>
                <canvas id="latencyDistributionChart"></canvas>
            </div>
        </section>

        <section class="table-section">
            <div class="table-title">Metric Breakdown</div>
            <table>
                <thead>
                    <tr>
                        <th>Metric</th>
                        <th>Value</th>
                    </tr>
                </thead>
                <tbody>
                    <tr>
                        <td>Average Latency</td>
                        <td>{{printf "%.2f" .AverageLatencyMs}} ms</td>
                    </tr>
                    <tr>
                        <td>P50 Latency (Median)</td>
                        <td>{{printf "%.2f" .P50LatencyMs}} ms</td>
                    </tr>
                    <tr>
                        <td>P90 Latency</td>
                        <td>{{printf "%.2f" .P90LatencyMs}} ms</td>
                    </tr>
                    <tr>
                        <td>P95 Latency</td>
                        <td>{{printf "%.2f" .P95LatencyMs}} ms</td>
                    </tr>
                    <tr>
                        <td>P99 Latency</td>
                        <td>{{printf "%.2f" .P99LatencyMs}} ms</td>
                    </tr>
                    <tr>
                        <td>Failed Requests</td>
                        <td style="color: var(--accent-red);">{{.TotalErrors}}</td>
                    </tr>
                </tbody>
            </table>
        </section>
    </div>

    <script>
        const latencyData = {
            avg: {{.AverageLatencyMs}},
            p50: {{.P50LatencyMs}},
            p90: {{.P90LatencyMs}},
            p95: {{.P95LatencyMs}},
            p99: {{.P99LatencyMs}}
        };

        const rawLatencies = [{{range $i, $el := .RawLatenciesMs}}{{if $i}},{{end}}{{$el}}{{end}}];

        // 1. Latency Summary Chart
        const ctxSummary = document.getElementById('latencySummaryChart').getContext('2d');
        new Chart(ctxSummary, {
            type: 'bar',
            data: {
                labels: ['Average', 'P50 (Median)', 'P90', 'P95', 'P99'],
                datasets: [{
                    label: 'Latency (ms)',
                    data: [latencyData.avg, latencyData.p50, latencyData.p90, latencyData.p95, latencyData.p99],
                    backgroundColor: [
                        'rgba(99, 102, 241, 0.7)',
                        'rgba(16, 185, 129, 0.7)',
                        'rgba(245, 158, 11, 0.7)',
                        'rgba(239, 68, 68, 0.7)',
                        'rgba(224, 76, 244, 0.7)'
                    ],
                    borderColor: [
                        '#6366F1',
                        '#10B981',
                        '#F59E0B',
                        '#EF4444',
                        '#E04CF4'
                    ],
                    borderWidth: 1.5,
                    borderRadius: 8
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: { display: false }
                },
                scales: {
                    y: {
                        beginAtZero: true,
                        grid: { color: 'rgba(255, 255, 255, 0.05)' },
                        ticks: { color: '#94A3B8' }
                    },
                    x: {
                        grid: { display: false },
                        ticks: { color: '#94A3B8' }
                    }
                }
            }
        });

        // 2. Latency Distribution Chart
        const ctxDistribution = document.getElementById('latencyDistributionChart').getContext('2d');
        
        let histLabels = [];
        let histData = [];

        if (rawLatencies.length > 0) {
            const min = Math.min(...rawLatencies);
            const max = Math.max(...rawLatencies);
            const binCount = 10;
            const binSize = (max - min) / binCount || 1;

            let bins = Array(binCount).fill(0);

            rawLatencies.forEach(val => {
                let index = Math.floor((val - min) / binSize);
                if (index >= binCount) index = binCount - 1;
                bins[index]++;
            });

            for (let i = 0; i < binCount; i++) {
                const startRange = (min + i * binSize).toFixed(1);
                const endRange = (min + (i + 1) * binSize).toFixed(1);
                histLabels.push(startRange + '-' + endRange + 'ms');
                histData.push(bins[i]);
            }
        }

        new Chart(ctxDistribution, {
            type: 'line',
            data: {
                labels: histLabels,
                datasets: [{
                    label: 'Frequency',
                    data: histData,
                    fill: true,
                    backgroundColor: 'rgba(16, 185, 129, 0.1)',
                    borderColor: '#10B981',
                    borderWidth: 2,
                    tension: 0.4,
                    pointRadius: 3
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: { display: false }
                },
                scales: {
                    y: {
                        beginAtZero: true,
                        grid: { color: 'rgba(255, 255, 255, 0.05)' },
                        ticks: { color: '#94A3B8' }
                    },
                    x: {
                        grid: { display: false },
                        ticks: { color: '#94A3B8' }
                    }
                }
            }
        });
    </script>
</body>
</html>`
type WriteJSON_Type = func(*ReportData, string) error // For linking symbol reference
type WriteHTML_Type = func(*ReportData, string) error // For linking symbol reference
type BuildReportData_Type = func(string, time.Duration, *metrics.Collector) *ReportData // For linking symbol reference
