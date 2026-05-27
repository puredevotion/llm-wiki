import { createFileRoute, Link } from '@tanstack/react-router'
import { useQuery } from '@tanstack/react-query'
import { apiFetch } from '@/lib/api/client'
import type { Zettel } from '@/lib/api/types'
import ReactMarkdown from 'react-markdown'
import { format, parseISO } from 'date-fns'
import { ArrowLeft, Calendar, User, Tag, Archive } from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Separator } from '@/components/ui/separator'

export const Route = createFileRoute('/zettels/$zettelId')({
  component: ZettelDetail,
})

function ZettelDetail() {
  const { zettelId } = Route.useParams()

  const { data: zettel, isLoading, error } = useQuery({
    queryKey: ['zettels', zettelId],
    queryFn: () => apiFetch<Zettel>(`/zettels/${zettelId}`),
  })

  if (isLoading) return <div className="p-20 text-center italic text-muted-foreground animate-pulse">Unfolding zettel...</div>
  if (error || !zettel) return <div className="p-20 text-center text-destructive">Error retrieving zettel. It may have been archived or lost in the graph.</div>

  return (
    <div className="max-w-4xl mx-auto space-y-8 animate-in fade-in slide-in-from-bottom-4 duration-500">
      <div className="flex items-center gap-4">
        <Link to="/zettels">
          <Button variant="ghost" size="sm" className="gap-2">
            <ArrowLeft className="h-4 w-4" />
            Garden
          </Button>
        </Link>
        <Separator orientation="vertical" className="h-4" />
        <div className="flex items-center gap-2 text-sm text-muted-foreground">
          <Tag className="h-3.5 w-3.5" />
          <span className="capitalize">{zettel.lifecycle}</span>
        </div>
      </div>

      <div className="space-y-4">
        <div className="flex items-start justify-between gap-4">
          <h1 className="text-4xl font-extrabold tracking-tight lg:text-5xl">{zettel.title}</h1>
          <Badge variant={zettel.status === 'active' ? 'default' : 'secondary'} className="mt-2 capitalize">
            {zettel.status}
          </Badge>
        </div>

        <div className="flex flex-wrap items-center gap-6 text-sm text-muted-foreground">
          <div className="flex items-center gap-2">
            <User className="h-4 w-4" />
            {zettel.created_by}
          </div>
          <div className="flex items-center gap-2">
            <Calendar className="h-4 w-4" />
            {format(parseISO(zettel.updated_at), 'PPP p')}
          </div>
          {zettel.status === 'archived' && (
            <div className="flex items-center gap-2 text-amber-500">
              <Archive className="h-4 w-4" />
              Archived
            </div>
          )}
        </div>
      </div>

      <Separator />

      <article className="prose prose-stone dark:prose-invert max-w-none">
        <ReactMarkdown>{zettel.body}</ReactMarkdown>
      </article>

      <div className="pt-12">
        <h3 className="text-lg font-semibold mb-4 flex items-center gap-2">
          <Tag className="h-4 w-4" />
          Metadata
        </h3>
        <div className="grid grid-cols-2 md:grid-cols-3 gap-4 text-sm bg-muted/30 p-6 rounded-xl border">
          <div>
            <p className="text-muted-foreground font-medium mb-1 uppercase tracking-wider text-[10px]">ID</p>
            <code className="bg-muted px-1.5 py-0.5 rounded text-xs">{zettel.id}</code>
          </div>
          <div>
            <p className="text-muted-foreground font-medium mb-1 uppercase tracking-wider text-[10px]">Created</p>
            <p>{format(parseISO(zettel.created_at), 'PP')}</p>
          </div>
          <div>
            <p className="text-muted-foreground font-medium mb-1 uppercase tracking-wider text-[10px]">Updated</p>
            <p>{format(parseISO(zettel.updated_at), 'PP')}</p>
          </div>
        </div>
      </div>
    </div>
  )
}
