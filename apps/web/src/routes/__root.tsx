import { createRootRoute, Link, Outlet } from '@tanstack/react-router'
import { TanStackRouterDevtools } from '@tanstack/router-devtools'
import { LayoutDashboard, BookText, Share2, History, Settings, Search } from 'lucide-react'

export const Route = createRootRoute({
  component: () => (
    <div className="flex h-screen bg-background font-sans antialiased overflow-hidden">
      {/* Sidebar */}
      <nav className="w-64 border-r bg-muted/30 flex flex-col">
        <div className="p-6 border-b flex items-center gap-2">
          <div className="w-8 h-8 rounded-lg bg-primary flex items-center justify-center">
            <Share2 className="text-primary-foreground h-5 w-5" />
          </div>
          <span className="text-xl font-bold tracking-tight">LLM Wiki</span>
        </div>
        
        <div className="flex-1 py-6 px-3 space-y-1 overflow-y-auto">
          <SidebarLink to="/" icon={LayoutDashboard} label="Dashboard" />
          <SidebarLink to="/zettels" icon={BookText} label="Zettel Garden" />
          <SidebarLink to="/graph" icon={Share2} label="Knowledge Graph" />
          <SidebarLink to="/timeline" icon={History} label="Timeline" />
          
          <div className="pt-6 pb-2 px-3 text-xs font-semibold text-muted-foreground uppercase tracking-wider">
            Identity
          </div>
          <SidebarLink to="/zettels" icon={Search} label="Search" />
        </div>

        <div className="p-4 border-t mt-auto">
          <SidebarLink to="/" icon={Settings} label="Settings" />
        </div>
      </nav>

      {/* Main Content */}
      <main className="flex-1 overflow-y-auto bg-background p-8">
        <div className="max-w-7xl mx-auto h-full">
          <Outlet />
        </div>
      </main>
      
      <TanStackRouterDevtools position="bottom-right" />
    </div>
  ),
})

function SidebarLink({ to, icon: Icon, label }: { to: string; icon: any; label: string }) {
  return (
    <Link
      to={to}
      className="flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors hover:bg-accent hover:text-accent-foreground [&.active]:bg-accent [&.active]:text-accent-foreground"
    >
      <Icon className="h-4 w-4" />
      {label}
    </Link>
  )
}
