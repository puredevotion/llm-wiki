export interface Zettel {
  id: string;
  title: string;
  body: string;
  lifecycle: 'source' | 'zettel' | 'topic' | 'project' | 'evergreen';
  status: 'active' | 'archived';
  created_by: string;
  created_at: string;
  updated_at: string;
}

export interface SearchResult {
  id: string;
  title: string;
  snippet: string;
  lifecycle: string;
  score: number;
}

export interface TimelineEvent {
  id: string;
  kind: 'meeting' | 'milestone' | 'task' | 'decision' | 'log';
  title: string;
  body: string;
  occurred_at?: string;
  starts_at?: string;
  ends_at?: string;
  recorded_at: string;
  created_by: string;
  metadata: Record<string, any>;
}
