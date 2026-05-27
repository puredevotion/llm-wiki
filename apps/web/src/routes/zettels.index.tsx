import { ChangeEvent } from 'react'
import { createFileRoute, useNavigate, useSearch, Link } from '@tanstack/react-router'
import { useQuery } from '@tanstack/react-query'
import { apiFetch } from '@/lib/api/client'
import type { SearchResult } from '@/lib/api/types'
import { Input } from '@/components/ui/input'
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Search, BookOpen, Filter } from 'lucide-react'
import { 
  Select, 
  SelectContent, 
  SelectItem, 
  SelectTrigger, 
  SelectValue 
} from '@/components/ui/select'

type ZettelSearch = {
  q?: string
  lifecycle?: string
}

export const Route = createFileRoute('/zettels/')({
  validateSearch: (search: Record<string, unknown>): ZettelSearch => {
    return {
      q: (search.q as string) || '',
      lifecycle: (search.lifecycle as string) || 'all',
    }
  },
  component: ZettelGarden,
})

function ZettelGarden() {
  const { q, lifecycle } = useSearch({ from: '/zettels/' })
  const navigate = useNavigate({ from: '/zettels/' })

  const { data: results, isLoading } = useQuery({
    queryKey: ['search', q, lifecycle],
    queryFn: () => {
      const params = new URLSearchParams()
      params.append('q', q || '')
      if (lifecycle && lifecycle !== 'all') {
        params.append('lifecycle', lifecycle)
      }
      params.append('limit', '20')
      return apiFetch<SearchResult[]>(`/search?${params.toString()}`)
    },
    enabled: (q?.length ?? 0) > 2,
  })

  const handleSearchChange = (e: ChangeEvent<HTMLInputElement>) => {
    navigate({
      search: (prev) => ({ ...prev, q: e.target.value }),
      replace: true,
    })
  }

  const handleLifecycleChange = (val: string) => {
    navigate({
      search: (prev) => ({ ...prev, lifecycle: val }),
    })
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
        <h1 className="text-4xl font-bold tracking-tight">Zettel Garden</h1>
        <div className="flex items-center gap-2">
          <Filter className="h-4 w-4 text-muted-foreground" />
          <Select value={lifecycle} onValueChange={handleLifecycleChange}>
            <SelectTrigger className="w-[180px]">
              <SelectValue placeholder="Lifecycle" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All Stages</SelectItem>
              <SelectItem value="source">Source</SelectItem>
              <SelectItem value="zettel">Zettel</SelectItem>
              <SelectItem value="topic">Topic</SelectItem>
              <SelectItem value="project">Project</SelectItem>
              <SelectItem value="evergreen">Evergreen</SelectItem>
            </SelectContent>
          </Select>
        </div>
      </div>

      <div className="relative">
        <Search className="absolute left-3 top-3.5 h-5 w-5 text-muted-foreground" />
        <Input
          placeholder="Semantic search across your knowledge..."
          className="pl-10 h-12 text-lg shadow-sm"
          value={q}
          onChange={handleSearchChange}
        />
      </div>

      {isLoading && (
        <div className="flex flex-col items-center justify-center p-20 text-muted-foreground italic gap-4">
          <div className="w-8 h-8 border-4 border-primary border-t-transparent rounded-full animate-spin" />
          Thinking...
        </div>
      )}

      <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
        {results?.map((res) => (
          <Link key={res.id} to="/zettels/$zettelId" params={{ zettelId: res.id }}>
            <Card className="hover:shadow-lg transition-all cursor-pointer group border-muted-foreground/20 h-full">
              <CardHeader className="flex flex-row items-start justify-between space-y-0 pb-4">
                <CardTitle className="text-xl font-semibold leading-tight group-hover:text-primary transition-colors line-clamp-2">
                  {res.title}
                </CardTitle>
                <BookOpen className="h-5 w-5 text-muted-foreground shrink-0 ml-4 mt-1 opacity-50 group-hover:opacity-100 transition-opacity" />
              </CardHeader>
              <CardContent>
                <p className="text-sm text-muted-foreground line-clamp-4 mb-6 leading-relaxed italic">
                  {res.snippet}
                </p>
                <div className="flex items-center gap-2">
                  <Badge variant="outline" className="capitalize font-medium border-primary/20 text-primary bg-primary/5">
                    {res.lifecycle}
                  </Badge>
                  {res.score > 0 && (
                    <div className="ml-auto text-xs font-semibold text-muted-foreground bg-muted px-2 py-1 rounded">
                      {Math.round(res.score * 100)}% match
                    </div>
                  )}
                </div>
              </CardContent>
            </Card>
          </Link>
        ))}

        {!isLoading && (q?.length ?? 0) > 2 && results?.length === 0 && (
          <div className="col-span-full flex flex-col items-center justify-center p-20 text-center border-2 border-dashed rounded-xl bg-muted/20">
            <p className="text-muted-foreground mb-2 font-bold text-xl">No matches in the garden</p>
            <p className="text-sm text-muted-foreground max-w-xs">
              Try broadening your search or check if the stage filter is too restrictive.
            </p>
          </div>
        )}

        {(q?.length ?? 0) <= 2 && (
          <div className="col-span-full flex flex-col items-center justify-center p-20 text-center opacity-50">
            <Search className="h-12 w-12 text-muted-foreground mb-4" />
            <p className="text-muted-foreground italic text-lg font-medium">
              Cultivate knowledge by searching for ideas, people, or projects.
            </p>
          </div>
        )}
      </div>
    </div>
  )
}
