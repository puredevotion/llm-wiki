import { createFileRoute, useNavigate } from '@tanstack/react-router'
import { useQuery } from '@tanstack/react-query'
import { apiFetch } from '@/lib/api/client'
import type { GraphData } from '@/lib/api/types'
import ForceGraph2D from 'react-force-graph-2d'

export const Route = createFileRoute('/graph')({
  component: GraphView,
})

const LABEL_COLORS: Record<string, string> = {
  Person: '#3b82f6', // blue-500
  Team: '#6366f1',   // indigo-500
  Organization: '#8b5cf6', // purple-500
  Topic: '#a855f7',  // purple-500
  Project: '#10b981', // emerald-500
  Source: '#f59e0b', // amber-500
  Zettel: '#f97316', // orange-500
  Event: '#ef4444',  // red-500
}

function GraphView() {
  const navigate = useNavigate()
  const { data, isLoading } = useQuery({
    queryKey: ['graph'],
    queryFn: () => apiFetch<GraphData>('/graph'),
  })

  const handleNodeClick = (node: any) => {
    if (node.label === 'Zettel') {
      navigate({ to: '/zettels/$zettelId', params: { zettelId: node.id } })
    } else if (node.label === 'Event') {
      navigate({ to: '/events/$eventId', params: { eventId: node.id } })
    }
  }

  return (
    <div className="h-[calc(100vh-120px)] border rounded-xl bg-background overflow-hidden relative">
      <div className="absolute top-4 left-4 z-10 bg-background/80 backdrop-blur-sm border rounded-lg p-4 shadow-sm pointer-events-none">
        <h2 className="text-sm font-bold mb-3">Knowledge Graph</h2>
        <div className="grid grid-cols-2 gap-x-4 gap-y-2">
          {Object.entries(LABEL_COLORS).map(([label, color]) => (
            <div key={label} className="flex items-center gap-2 text-[10px] font-medium uppercase tracking-wider">
              <div className="w-2.5 h-2.5 rounded-full" style={{ backgroundColor: color }} />
              {label}
            </div>
          ))}
        </div>
      </div>

      {isLoading && (
        <div className="absolute inset-0 flex items-center justify-center bg-background/50 backdrop-blur-sm z-20">
          <p className="text-muted-foreground animate-pulse">Mapping connections...</p>
        </div>
      )}

      {data && (
        <ForceGraph2D
          graphData={data}
          nodeLabel={(node: any) => `${node.label}: ${node.name}`}
          nodeColor={(node: any) => LABEL_COLORS[node.label] || '#94a3b8'}
          nodeRelSize={6}
          onNodeClick={handleNodeClick}
          linkDirectionalArrowLength={4}
          linkDirectionalArrowRelPos={1}
          linkCurvature={0.25}
          linkColor={() => 'rgba(148, 163, 184, 0.2)'}
          backgroundColor="#00000000"
        />
      )}

      {!isLoading && (!data || data.nodes.length === 0) && (
        <div className="absolute inset-0 flex flex-col items-center justify-center text-center p-12">
          <p className="text-muted-foreground mb-1 font-semibold text-lg">The void is vast</p>
          <p className="text-sm text-muted-foreground">Start adding notes or events to populate the graph.</p>
        </div>
      )}
    </div>
  )
}
