import { Badge, LayerCard } from "@cloudflare/kumo"
import { useEffect, useState } from "react"
import { getJson } from "./api"
import Incidents from "./Incidents"
import ProjectDetail from "./ProjectDetail"

type Project = {
  id: number
  name: string
  url: string
  is_up: boolean
  last_checked_at: string | null
  response_time_ms: number | null
  uptime_24h_percent: number
}

function Overview() {
  const [projects, setProjects] = useState<Project[]>([])
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const load = () => getJson<Project[]>("/api/projects").then(setProjects).catch((e) => setError(String(e)))
    load()
    const id = setInterval(load, 30000)
    return () => clearInterval(id)
  }, [])

  if (error) return <div style={{ padding: 24 }}>{error}</div>

  return (
    <div style={{ padding: 24 }}>
      <div style={{ display: "flex", gap: 16 }}>
        <a href="#/">Overview</a>
        <a href="#/incidents">Incidents</a>
      </div>
      <h1>Orpheus Overview</h1>
      <div
        style={{
          display: "grid",
          gap: 16,
          gridTemplateColumns: "repeat(auto-fill, minmax(280px, 1fr))",
          marginTop: 16,
        }}
      >
        {projects.map((p) => (
          <a key={p.id} href={`#/projects/${p.id}`}>
            <LayerCard>
              <LayerCard.Secondary>{p.name}</LayerCard.Secondary>
              <LayerCard.Primary>
                <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
                  <Badge variant={p.is_up ? "green" : "red"}>{p.is_up ? "Up" : "Down"}</Badge>
                  <span>{p.url}</span>
                </div>
                <div>Response: {p.response_time_ms ?? "-"} ms</div>
                <div>Uptime 24h: {p.uptime_24h_percent}%</div>
                <div>
                  Last check: {p.last_checked_at ? new Date(p.last_checked_at).toLocaleString() : "-"}
                </div>
              </LayerCard.Primary>
            </LayerCard>
          </a>
        ))}
      </div>
      {projects.length === 0 && <div>No projects</div>}
    </div>
  )
}

export default function App() {
  const [hash, setHash] = useState(() => window.location.hash)

  useEffect(() => {
    const onHash = () => setHash(window.location.hash)
    window.addEventListener("hashchange", onHash)
    return () => window.removeEventListener("hashchange", onHash)
  }, [])

  if (hash === "#/incidents") return <Incidents />
  const match = /^#\/projects\/(\d+)$/.exec(hash)
  if (match) return <ProjectDetail projectId={Number(match[1])} />
  return <Overview />
}
