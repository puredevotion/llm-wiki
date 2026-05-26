import { createFileRoute } from '@tanstack/react-router'
import ForceGraph2D from 'react-force-graph-2d'
import { useMemo } from 'react'

export const Route = createFileRoute('/graph')({
  component: GraphView,
})

function GraphView() {
  const data = useMemo(() => {
    return {
      nodes: [
        { id: '1', name: 'Alice', val: 10, group: 'Person' },
        { id: '2', name: 'Go Backend', val: 15, group: 'Project' },
        { id: '3', name: 'Vector Search', val: 8, group: 'Topic' },
        { id: '4', name: 'Zettel 1', val: 5, group: 'Zettel' },
        { id: '5', name: 'Zettel 2', val: 5, group: 'Zettel' },
      ],
      links: [
        { source: '1', target: '2' },
        { source: '2', target: '3' },
        { source: '4', target: '3' },
        { source: '5', target: '3' },
      ]
    }
  }, [])

  return (
    <div className="h-[calc(100vh-120px)] border rounded-xl bg-background overflow-hidden relative">
      <div className="absolute top-4 left-4 z-10 bg-background/80 backdrop-blur-sm border rounded-lg p-3 shadow-sm">
        <h2 className="text-sm font-semibold mb-2">Knowledge Graph</h2>
        <div className="flex flex-col gap-1">
          <div className="flex items-center gap-2 text-xs">
            <div className="w-3 h-3 rounded-full bg-blue-500" /> Person
          </div>
          <div className="flex items-center gap-2 text-xs">
            <div className="w-3 h-3 rounded-full bg-green-500" /> Project
          </div>
          <div className="flex items-center gap-2 text-xs">
            <div className="w-3 h-3 rounded-full bg-purple-500" /> Topic
          </div>
          <div className="flex items-center gap-2 text-xs">
            <div className="w-3 h-3 rounded-full bg-orange-500" /> Zettel
          </div>
        </div>
      </div>
      <ForceGraph2D
        graphData={data}
        nodeLabel="name"
        nodeAutoColorBy="group"
        linkDirectionalArrowLength={3.5}
        linkDirectionalArrowRelPos={1}
        backgroundColor="#00000000"
      />
    </div>
  )
}
