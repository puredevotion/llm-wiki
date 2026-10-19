import { createFileRoute, Link } from '@tanstack/react-router'
import { useQuery } from '@tanstack/react-query'
import { apiFetch } from '@/lib/api/client'
import type { TimelineEvent } from '@/lib/api/types'
import { format, parseISO } from 'date-fns'
import { ArrowLeft, Calendar, User, Tag, Clock, MapPin } from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Separator } from '@/components/ui/separator'

export const Route = createFileRoute('/events/$eventId')({
  component: EventDetail,
})

function EventDetail() {
  const { eventId } = Route.useParams()

  const { data: event, isLoading, error } = useQuery({
    queryKey: ['events', eventId],
    queryFn: () => apiFetch<TimelineEvent>(`/events/${eventId}`),
  })

  if (isLoading) return <div className="p-20 text-center italic text-muted-foreground animate-pulse">Revisiting the past...</div>
  if (error || !event) return <div className="p-20 text-center text-destructive">Error retrieving event. It may have been a temporal anomaly.</div>

  const date = parseISO(event.occurred_at || event.starts_at || event.recorded_at)

  return (
    <div className="max-w-4xl mx-auto space-y-8 animate-in fade-in slide-in-from-bottom-4 duration-500">
      <div className="flex items-center gap-4">
        <Link to="/timeline">
          <Button variant="ghost" size="sm" className="gap-2">
            <ArrowLeft className="h-4 w-4" />
            Timeline
          </Button>
        </Link>
        <Separator orientation="vertical" className="h-4" />
        <div className="flex items-center gap-2 text-sm text-muted-foreground">
          <Clock className="h-3.5 w-3.5" />
          <span className="capitalize">{event.kind}</span>
        </div>
      </div>

      <div className="space-y-4">
        <div className="flex items-start justify-between gap-4">
          <h1 className="text-4xl font-extrabold tracking-tight lg:text-5xl">{event.title}</h1>
          <Badge className="mt-2 capitalize">
            {event.kind}
          </Badge>
        </div>

        <div className="flex flex-wrap items-center gap-6 text-sm text-muted-foreground">
          <div className="flex items-center gap-2">
            <Calendar className="h-4 w-4 text-primary" />
            <span className="font-semibold text-foreground">{format(date, 'PPPP')}</span>
          </div>
          <div className="flex items-center gap-2 text-primary font-medium bg-primary/5 px-2 py-1 rounded-md">
            <Clock className="h-4 w-4" />
            {format(date, 'p')}
          </div>
          <div className="flex items-center gap-2">
            <User className="h-4 w-4" />
            {event.created_by}
          </div>
        </div>
      </div>

      <Separator />

      <div className="bg-muted/20 p-8 rounded-2xl border border-muted-foreground/10">
        <p className="text-lg text-foreground whitespace-pre-wrap leading-relaxed">
          {event.body}
        </p>
      </div>

      {event.metadata && Object.keys(event.metadata).length > 0 && (
        <div className="pt-8">
          <h3 className="text-lg font-semibold mb-4 flex items-center gap-2 text-muted-foreground">
            <Tag className="h-4 w-4" />
            Event Context
          </h3>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {Object.entries(event.metadata).map(([k, v]) => (
              <div key={k} className="flex items-center justify-between p-4 rounded-xl border bg-card/50 shadow-sm hover:border-primary/20 transition-colors">
                <span className="text-sm font-medium text-muted-foreground capitalize">{k.replace(/_/g, ' ')}</span>
                <span className="text-sm font-bold">{String(v)}</span>
              </div>
            ))}
          </div>
        </div>
      )}
      
      <div className="pt-12 text-[10px] text-muted-foreground text-center uppercase tracking-[0.2em] font-bold">
        Reference ID: {event.id}
      </div>
    </div>
  )
}
