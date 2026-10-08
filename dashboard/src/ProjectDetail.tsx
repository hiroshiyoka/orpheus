import { ChartPalette, LayerCard, TimeseriesChart } from "@cloudflare/kumo"
import { useEffect, useState } from "react"
import { getJson } from "./api"
import echarts from "./echarts"

type Project = { id: number; name: string; url: string }
type Check = { checked_at: string; response_time_ms: number | null }
type Metric = { period_start: string; requests_count: number; error_count: number }

export default function ProjectDetail({ projectId }: { projectId: number }) {
  const [project, setProject] = useState<Project | null>(null)
  const [checks, setChecks] = useState<Check[]>([])
  const [metrics, setMetrics] = useState<Metric[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let alive = true
    Promise.all([
      getJson<Project[]>("/api/projects"),
      getJson<Check[]>(`/api/projects/${projectId}/checks?range=24h`),
      getJson<Metric[]>(`/api/projects/${projectId}/metrics?range=7d`),
    ])
      .then(([projects, c, m]) => {
        if (!alive) return
        setProject(projects.find((p) => p.id === projectId) ?? null)
        setChecks(c)
        setMetrics(m)
        setLoading(false)
      })
      .catch((e) => alive && setError(String(e)))
    return () => {
      alive = false
    }
  }, [projectId])

  if (error) return <div style={{ padding: 24 }}>{error}</div>

  const responseTime: [number, number][] = checks
    .filter((c) => c.response_time_ms != null)
    .map((c) => [Date.parse(c.checked_at), c.response_time_ms as number])

  const errorRate: [number, number][] = metrics
    .filter((m) => m.requests_count > 0)
    .map((m) => [Date.parse(m.period_start), Math.round((m.error_count / m.requests_count) * 10000) / 100])

  return (
    <div style={{ padding: 24 }}>
      <a href="#/">← Overview</a>
      <h1>{project ? project.name : `Project #${projectId}`}</h1>

      <LayerCard style={{ marginTop: 16 }}>
        <LayerCard.Secondary>Response time — last 24h</LayerCard.Secondary>
        <LayerCard.Primary>
          <TimeseriesChart
            echarts={echarts}
            loading={loading}
            data={[{ name: "Response time (ms)", data: responseTime, color: ChartPalette.categorical(0) }]}
            yAxisName="ms"
            ariaDescription="Response time history for the last 24 hours"
          />
        </LayerCard.Primary>
      </LayerCard>

      <LayerCard style={{ marginTop: 16 }}>
        <LayerCard.Secondary>Error rate — last 7 days</LayerCard.Secondary>
        <LayerCard.Primary>
          <TimeseriesChart
            echarts={echarts}
            loading={loading}
            data={[{ name: "Error rate (%)", data: errorRate, color: ChartPalette.categorical(1) }]}
            yAxisName="%"
            ariaDescription="Error rate history for the last 7 days"
          />
        </LayerCard.Primary>
      </LayerCard>
    </div>
  )
}
