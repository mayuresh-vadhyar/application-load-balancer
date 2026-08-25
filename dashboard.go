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
</head>
<body>
	<h1>{{ .Title }}</h1>
	<table>
		<thead>
			<tr>
				<th>ID</th>
				<th>URL</th>
			</tr>
		</thead>
		<tbody>
			{{- range .Rows }}
			<tr>
				<td>{{ .ID }}</td>
				<td>{{ .URL }}</td>
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
