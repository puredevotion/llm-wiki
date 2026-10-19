import { createFileRoute } from '@tanstack/react-router'
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/card'
import { BookOpen, Calendar, Share2, TrendingUp, Inbox } from 'lucide-react'

export const Route = createFileRoute('/')({
  component: Dashboard,
})

function Dashboard() {
  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-4xl font-bold tracking-tight">Dashboard</h1>
        <p className="text-muted-foreground mt-2 text-lg">Welcome to your Knowledge Lens.</p>
      </div>
      
      <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-4">
        <StatCard title="Total Zettels" value="42" icon={BookOpen} trend="+12% this week" />
        <StatCard title="Events" value="128" icon={Calendar} trend="+5 today" />
        <StatCard title="Graph Nodes" value="1,024" icon={Share2} trend="Optimal density" />
        <StatCard title="Ingestion Inbox" value="7" icon={Inbox} trend="Pending review" color="text-amber-500" />
      </div>

      <div className="grid gap-6 md:grid-cols-2">
        <Card className="col-span-1">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <TrendingUp className="h-5 w-5 text-primary" />
              Recent Activity
            </CardTitle>
            <CardDescription>Latest updates from your knowledge garden.</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {[1, 2, 3].map((i) => (
                <div key={i} className="flex items-start gap-3 border-b pb-3 last:border-0 last:pb-0">
                  <div className="w-2 h-2 rounded-full bg-blue-500 mt-2 shrink-0" />
                  <div>
                    <p className="text-sm font-medium">Updated "Vector Similarity Search"</p>
                    <p className="text-xs text-muted-foreground">2 hours ago</p>
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>

        <Card className="col-span-1">
          <CardHeader>
            <CardTitle>System Status</CardTitle>
            <CardDescription>Backend and database connectivity.</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              <StatusItem label="Turso (SQL)" status="Healthy" />
              <StatusItem label="LadybugDB (Graph)" status="Healthy" />
              <StatusItem label="OpenAI (Embeddings)" status="Connected" />
              <StatusItem label="Sync Engine" status="Idle" />
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}

function StatCard({ title, value, icon: Icon, trend, color = "text-primary" }: any) {
  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-sm font-medium">{title}</CardTitle>
        <Icon className={`h-4 w-4 ${color}`} />
      </CardHeader>
      <CardContent>
        <div className="text-2xl font-bold">{value}</div>
        <p className="text-xs text-muted-foreground mt-1">{trend}</p>
      </CardContent>
    </Card>
  )
}

function StatusItem({ label, status }: { label: string; status: string }) {
  return (
    <div className="flex items-center justify-between text-sm">
      <span className="text-muted-foreground">{label}</span>
      <span className="font-medium text-emerald-500 flex items-center gap-1">
        <div className="w-1.5 h-1.5 rounded-full bg-emerald-500" />
        {status}
      </span>
    </div>
  )
}
