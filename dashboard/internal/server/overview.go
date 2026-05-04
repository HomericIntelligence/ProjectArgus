package server

var overviewHTML = []byte(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Atlas — HomericIntelligence</title>
<link rel="stylesheet" href="/static/css/atlas.css">
<script src="https://unpkg.com/htmx.org@1.9.12" defer></script>
<script src="/static/js/atlas.js" defer></script>
</head>
<body>
<nav>
  <span class="brand">Atlas</span>
  <a href="/">Overview</a>
  <a href="/hosts">Hosts</a>
  <a href="/agents">Agents</a>
  <a href="/tasks">Tasks</a>
  <a href="/nats">NATS</a>
  <a href="/grafana">Grafana</a>
  <a href="/mnemosyne">Mnemosyne</a>
  <span class="conn-dot" title="SSE disconnected"></span>
</nav>
<main>
  <div class="stat-grid">
    <div class="stat-card"><div class="label">Agents</div><div class="value" id="stat-agents">0</div></div>
    <div class="stat-card"><div class="label">Tasks</div><div class="value" id="stat-tasks">0</div></div>
    <div class="stat-card"><div class="label">Streams</div><div class="value" id="stat-streams">0</div></div>
    <div class="stat-card"><div class="label">Hosts</div><div class="value" id="stat-hosts">0</div></div>
  </div>
  <p style="color:#8b949e">Live feed active after M2 (NATS JetStream subscribers).</p>
</main>
</body>
</html>`)
