package main

import (
	"html/template"
	"net/http"

	"github.com/mayuresh-vadhyar/application-load-balancer/Response"
	"github.com/mayuresh-vadhyar/application-load-balancer/server"
)

const dashboardHTML = `<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<title>{{ .Title }}</title>
	<style>
		table { border-collapse: collapse; width: 100%; }
		th, td { border: 1px solid #ddd; padding: 8px; }
		th { background-color: #f2f2f2; text-align: left; }
		tr:nth-child(even) { background-color: #fbfbfb; }
		.status-healthy { color: #155724; font-weight: bold; }
		.status-unhealthy { color: #721c24; font-weight: bold; }
	</style>
</head>
<body>
	<h1>{{ .Title }}</h1>
	<table>
		<thead>
			<tr>
				<th>ID</th>
				<th>URL</th>
				<th>Healthy</th>
				<th>Total Requests</th>
				<th>Active Requests</th>
			</tr>
		</thead>
		<tbody>
			{{- range .Rows }}
			<tr>
				<td>{{ .ID }}</td>
				<td>{{ .URL }}</td>
				<td class="status-{{ if .IsHealthy }}healthy{{ else }}unhealthy{{ end }}">{{ if .IsHealthy }}Healthy{{ else }}Unhealthy{{ end }}</td>
				<td>{{ .RequestCount }}</td>
				<td>{{ .ActiveReqCount }}</td>
			</tr>
			{{- end }}
		</tbody>
	</table>
</body>
</html>`

var dashboardTemplate = template.Must(template.New("dashboard").Parse(dashboardHTML))

type dashboardPageData struct {
	Title string
	Rows  []serverStatsEntry
}

type serverStatsEntry struct {
	ID             int    `json:"id"`
	URL            string `json:"url"`
	IsHealthy      bool   `json:"isHealthy"`
	RequestCount   int64  `json:"requestCount"`
	ActiveReqCount int64  `json:"activeReqCount"`
}

func collectServerStats() []serverStatsEntry {
	stats := make([]serverStatsEntry, 0, len(server.GetServers()))
	for _, item := range server.GetServers() {
		item.Mutex.Lock()
		stats = append(stats, serverStatsEntry{
			ID:             item.Id,
			URL:            item.URL.String(),
			IsHealthy:      item.IsHealthy,
			RequestCount:   item.RequestCount,
			ActiveReqCount: item.ActiveReqCount,
		})
		item.Mutex.Unlock()
	}
	return stats
}

func serverDashboardHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		Response.WriteErrorResponse(w, http.StatusMethodNotAllowed, "Method Not Allowed")
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := dashboardTemplate.Execute(w, dashboardPageData{
		Title: "Server Stats Dashboard",
		Rows:  collectServerStats(),
	}); err != nil {
		Response.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
	}
}
