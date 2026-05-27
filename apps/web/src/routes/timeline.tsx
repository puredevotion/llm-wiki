import { createFileRoute, useNavigate, useSearch } from '@tanstack/react-router'
import { useQuery } from '@tanstack/react-query'
import { apiFetch } from '@/lib/api/client'
import type { TimelineEvent } from '@/lib/api/types'
import { format, parseISO, startOfDay, endOfDay } from 'date-fns'
import { Calendar as CalendarIcon, CheckCircle2, Flag, MessageSquare, StickyNote, Filter } from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { 
  Select, 
  SelectContent, 
  SelectItem, 
  SelectTrigger, 
  SelectValue 
} from '@/components/ui/select'

type TimelineSearch = {
  kind?: string
  start?: string
  end?: string
}

export const Route = createFileRoute('/timeline')({
  validateSearch: (search: Record<string, unknown>): TimelineSearch => {
    return {
      kind: (search.kind as string) || 'all',
      start: (search.start as string) || '',
      end: (search.end as string) || '',
    }
  },
  component: TimelineView,
})

const KIND_ICONS: Record<string, any> = {
  meeting: MessageSquare,
  milestone: Flag,
  task: CheckCircle2,
  decision: StickyNote,
  log: CalendarIcon,
}

function TimelineView() {
  const { kind, start, end } = useSearch({ from: '/timeline' })
  const navigate = useNavigate({ from: '/timeline' })

  const { data: events, isLoading } = useQuery({
    queryKey: ['timeline', kind, start, end],
    queryFn: () => {
      const params = new URLSearchParams()
      if (kind && kind !== 'all') params.append('kind', kind)
      if (start) params.append('starts_at', start)
      if (end) params.append('ends_at', end)
      params.append('limit', '50')
      return apiFetch<TimelineEvent[]>(`/timeline?${params.toString()}`)
    },
  })

  const handleKindChange = (val: string) => {
    navigate({
      search: (prev) => ({ ...prev, kind: val }),
    })
  }

  return (
    <div className="space-y-8 max-w-4xl mx-auto">
      <div className="flex flex-col gap-4 md:flex-row md:items-center md:justify-between border-b pb-6">
        <div>
          <h1 className="text-4xl font-bold tracking-tight">Timeline</h1>
          <p className="text-muted-foreground mt-1">The chronological history of your knowledge.</p>
        </div>
        <div className="flex items-center gap-2">
          <Filter className="h-4 w-4 text-muted-foreground" />
          <Select value={kind} onValueChange={handleKindChange}>
            <SelectTrigger className="w-[150px]">
              <SelectValue placeholder="Event Type" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All Events</SelectItem>
              <SelectItem value="meeting">Meetings</SelectItem>
              <SelectItem value="milestone">Milestones</SelectItem>
              <SelectItem value="decision">Decisions</SelectItem>
              <SelectItem value="task">Tasks</SelectItem>
              <SelectItem value="log">Logs</SelectItem>
            </SelectContent>
          </Select>
        </div>
      </div>

      {isLoading && (
        <div className="flex flex-col items-center justify-center py-20 text-muted-foreground italic gap-4 animate-pulse">
          <CalendarIcon className="h-8 w-8 text-muted-foreground/30" />
          Recalling history...
        </div>
      )}

      <div className="relative space-y-0 before:absolute before:inset-0 before:ml-5 before:-translate-x-px before:h-full before:w-0.5 before:bg-gradient-to-b before:from-transparent before:via-muted-foreground/20 before:to-transparent">
        {events?.map((event) => {
          const Icon = KIND_ICONS[event.kind] || CalendarIcon
          const date = parseISO(event.occurred_at || event.starts_at || event.recorded_at)
          
          return (
            <Link key={event.id} to="/events/$eventId" params={{ eventId: event.id }} className="block group">
              <div className="relative flex items-start pb-10">
                <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-full border bg-background shadow-sm group-hover:scale-110 group-hover:border-primary transition-all z-10">
                  <Icon className="h-5 w-5 text-primary" />
                </div>
                <div className="ml-6 pt-0.5 w-full">
                  <div className="flex items-center justify-between mb-2">
                    <time className="text-sm font-semibold text-muted-foreground group-hover:text-primary transition-colors">
                      {format(date, 'PPP p')}
                    </time>
                    <Badge variant="secondary" className="capitalize text-[10px] font-bold px-2 py-0 h-5">
                      {event.kind}
                    </Badge>
                  </div>
                  <div className="rounded-xl border bg-card p-5 text-card-foreground shadow-sm group-hover:shadow-md group-hover:border-primary/30 transition-all">
                    <h3 className="font-bold text-xl leading-snug mb-3">{event.title}</h3>
                    <p className="text-sm text-muted-foreground whitespace-pre-wrap leading-relaxed">
                      {event.body}
                    </p>
                    
                    {event.metadata && Object.keys(event.metadata).length > 0 && (
                      <div className="mt-4 pt-4 border-t flex flex-wrap gap-2">
                        {Object.entries(event.metadata).map(([k, v]) => (
                          <div key={k} className="text-[10px] bg-muted px-2 py-0.5 rounded-md font-mono">
                            {k}: {String(v)}
                          </div>
                        ))}
                      </div>
                    )}
                  </div>
                </div>
              </div>
            </Link>
          )
        })}

        {!isLoading && events?.length === 0 && (
          <div className="flex flex-col items-center justify-center p-20 text-center border-2 border-dashed rounded-2xl bg-muted/10">
            <CalendarIcon className="h-16 w-12 text-muted-foreground/20 mb-4" />
            <p className="font-bold text-muted-foreground text-xl">The past is quiet</p>
            <p className="text-sm text-muted-foreground mt-1">No events match your current filters.</p>
          </div>
        )}
      </div>
    </div>
  )
}
