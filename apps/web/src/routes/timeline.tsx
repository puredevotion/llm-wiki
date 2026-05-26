import { createFileRoute } from '@tanstack/react-router'
import { useQuery } from '@tanstack/react-query'
import { apiFetch } from '@/lib/api/client'
import type { TimelineEvent } from '@/lib/api/types'
import { format, parseISO } from 'date-fns'
import { Calendar, CheckCircle2, Flag, MessageSquare, StickyNote } from 'lucide-react'
import { Badge } from '@/components/ui/badge'

export const Route = createFileRoute('/timeline')({
  component: TimelineView,
})

const KIND_ICONS = {
  meeting: MessageSquare,
  milestone: Flag,
  task: CheckCircle2,
  decision: StickyNote,
  log: Calendar,
}

function TimelineView() {
  const { data: events, isLoading } = useQuery({
    queryKey: ['timeline'],
    queryFn: () => apiFetch<TimelineEvent[]>('/timeline?limit=50'),
  })

  return (
    <div className="space-y-8 max-w-4xl mx-auto">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">Timeline</h1>
        <p className="text-muted-foreground">The chronological history of your knowledge.</p>
      </div>

      {isLoading && <div className="italic text-muted-foreground text-center py-20">Recalling history...</div>}

      <div className="relative space-y-0 before:absolute before:inset-0 before:ml-5 before:-translate-x-px before:h-full before:w-0.5 before:bg-gradient-to-b before:from-transparent before:via-muted-foreground/20 before:to-transparent">
        {events?.map((event) => {
          const Icon = KIND_ICONS[event.kind] || Calendar
          const date = parseISO(event.occurred_at || event.starts_at || event.recorded_at)
          
          return (
            <div key={event.id} className="relative flex items-start pb-8 group">
              <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-full border bg-background shadow-sm group-hover:scale-110 transition-transform z-10">
                <Icon className="h-5 w-5 text-primary" />
              </div>
              <div className="ml-4 pt-0.5 w-full">
                <div className="flex items-center justify-between mb-1">
                  <time className="text-sm font-medium text-muted-foreground">
                    {format(date, 'PPP p')}
                  </time>
                  <Badge variant="outline" className="capitalize text-[10px] h-5">
                    {event.kind}
                  </Badge>
                </div>
                <div className="rounded-lg border bg-card p-4 text-card-foreground shadow-sm group-hover:border-primary/50 transition-colors">
                  <h3 className="font-semibold text-lg leading-none mb-2">{event.title}</h3>
                  <p className="text-sm text-muted-foreground whitespace-pre-wrap leading-relaxed">
                    {event.body}
                  </p>
                </div>
              </div>
            </div>
          )
        })}

        {!isLoading && events?.length === 0 && (
          <div className="flex flex-col items-center justify-center p-20 text-center border-2 border-dashed rounded-xl">
            <Calendar className="h-12 w-12 text-muted-foreground/40 mb-4" />
            <p className="font-medium text-muted-foreground text-lg">Your timeline is empty</p>
            <p className="text-sm text-muted-foreground">Events and milestones will appear here as they are captured.</p>
          </div>
        )}
      </div>
    </div>
  )
}
