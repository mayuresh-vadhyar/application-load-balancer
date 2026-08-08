package main

import (
	"html/template"

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
