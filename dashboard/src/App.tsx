import { useEffect, useState } from "react"
import { Badge, LayerCard } from "@cloudflare/kumo"

type Project = {
  id: number
  name: string
  url: string
  is_up: boolean
  last_checked_at: string | null
  response_time_ms: number | null
  uptime_24h_percent: number
}

export default function App() {
  const [projects, setProjects] = useState<Project[]>([])
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    fetch("/api/projects")
      .then((r) => {
        if (!r.ok) throw new Error(String(r.status))
        return r.json()
      })
      .then(setProjects)
      .catch((e) => setError(String(e)))
  }, [])

  if (error) return <div>{error}</div>

  return (
    <div style={{ padding: 24 }}>
      <h1>Orpheus Overview</h1>
      <div style={{ display: "grid", gap: 16, gridTemplateColumns: "repeat(auto-fill, minmax(280px, 1fr))", marginTop: 16 }}>
        {projects.map((p) => (
          <LayerCard key={p.id} title={p.name}>
            <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
              <Badge>{p.is_up ? "Up" : "Down"}</Badge>
              <span>{p.url}</span>
            </div>
            <div>Response: {p.response_time_ms ?? "-"} ms</div>
            <div>Uptime 24h: {p.uptime_24h_percent}%</div>
            <div>Last check: {p.last_checked_at ? new Date(p.last_checked_at).toLocaleString() : "-"}</div>
          </LayerCard>
        ))}
      </div>
      {projects.length === 0 && <div>No projects</div>}
    </div>
  )
}
