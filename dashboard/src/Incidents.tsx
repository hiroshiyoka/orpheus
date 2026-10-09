import { Table } from "@cloudflare/kumo"
import { useEffect, useState } from "react"
import { getJson } from "./api"

type Incident = {
  id: number
  project_id: number
  type: string
  started_at: string
  resolved_at: string | null
  description: string | null
}

function formatDuration(start: string, end: string | null) {
  const s = Date.parse(start)
  const e = end ? Date.parse(end) : Date.now()
  const ms = e - s
  const m = Math.floor(ms / 60000)
  if (m < 60) return `${m}m`
  const h = Math.floor(m / 60)
  const rm = m % 60
  return `${h}h ${rm}m`
}

export default function Incidents() {
  const [incidents, setIncidents] = useState<Incident[]>([])
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    getJson<Incident[]>("/api/incidents").then(setIncidents).catch((e) => setError(String(e)))
  }, [])

  if (error) return <div style={{ padding: 24 }}>{error}</div>

  return (
    <div style={{ padding: 24 }}>
      <a href="#/">← Overview</a>
      <h1>Incidents</h1>
      <Table>
        <Table.Header>
          <Table.Row>
            <Table.Head>ID</Table.Head>
            <Table.Head>Project</Table.Head>
            <Table.Head>Type</Table.Head>
            <Table.Head>Started</Table.Head>
            <Table.Head>Duration</Table.Head>
            <Table.Head>Status</Table.Head>
          </Table.Row>
        </Table.Header>
        <Table.Body>
          {incidents.map((inc) => (
            <Table.Row key={inc.id}>
              <Table.Cell>{inc.id}</Table.Cell>
              <Table.Cell>{inc.project_id}</Table.Cell>
              <Table.Cell>{inc.type}</Table.Cell>
              <Table.Cell>{new Date(inc.started_at).toLocaleString()}</Table.Cell>
              <Table.Cell>{formatDuration(inc.started_at, inc.resolved_at)}</Table.Cell>
              <Table.Cell>{inc.resolved_at ? "Resolved" : "Open"}</Table.Cell>
            </Table.Row>
          ))}
        </Table.Body>
      </Table>
      {incidents.length === 0 && <div>No incidents</div>}
    </div>
  )
}
