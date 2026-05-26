import { useState } from 'react'
import type { ChangeEvent } from 'react'
import { createFileRoute } from '@tanstack/react-router'
import { useQuery } from '@tanstack/react-query'
import { apiFetch } from '@/lib/api/client'
import type { SearchResult } from '@/lib/api/types'
import { Input } from '@/components/ui/input'
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Search, BookOpen } from 'lucide-react'

export const Route = createFileRoute('/zettels')({
  component: ZettelGarden,
})

function ZettelGarden() {
  const [query, setQuery] = useState('')

  const { data: results, isLoading } = useQuery({
    queryKey: ['search', query],
    queryFn: () => 
      query.length > 2 
        ? apiFetch<SearchResult[]>(`/search?q=${encodeURIComponent(query)}&limit=20`)
        : Promise.resolve([]),
    enabled: query.length > 2,
  })

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold tracking-tight">Zettel Garden</h1>
      </div>

      <div className="relative">
        <Search className="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
        <Input
          placeholder="Semantic search across your knowledge..."
          className="pl-9 h-12 text-lg"
          value={query}
          onChange={(e: ChangeEvent<HTMLInputElement>) => setQuery(e.target.value)}
        />
      </div>

      {isLoading && (
        <div className="flex justify-center p-12 text-muted-foreground italic">
          Thinking...
        </div>
      )}

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {results?.map((res) => (
          <Card key={res.id} className="hover:shadow-md transition-shadow cursor-pointer">
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-lg font-medium leading-tight line-clamp-2">
                {res.title}
              </CardTitle>
              <BookOpen className="h-4 w-4 text-muted-foreground shrink-0 ml-2" />
            </CardHeader>
            <CardContent>
              <p className="text-sm text-muted-foreground line-clamp-3 mb-4 italic">
                {res.snippet}
              </p>
              <div className="flex gap-2 mt-auto">
                <Badge variant="outline" className="capitalize">
                  {res.lifecycle}
                </Badge>
                {res.score > 0 && (
                  <Badge variant="secondary">
                    {Math.round(res.score * 100)}% match
                  </Badge>
                )}
              </div>
            </CardContent>
          </Card>
        ))}

        {!isLoading && query.length > 2 && results?.length === 0 && (
          <div className="col-span-full flex flex-col items-center justify-center p-12 text-center border rounded-lg border-dashed">
            <p className="text-muted-foreground mb-1 font-semibold text-lg">No matches found</p>
            <p className="text-sm text-muted-foreground">Try a different query or explore the graph.</p>
          </div>
        )}

        {query.length <= 2 && (
          <div className="col-span-full flex flex-col items-center justify-center p-12 text-center">
            <p className="text-muted-foreground italic">Start typing to search your knowledge garden...</p>
          </div>
        )}
      </div>
    </div>
  )
}
