import { useWS } from '../context/WSContext'
import MetricCard from '../components/MetricCard'

export default function Dashboard() {
  const { metrics, connected } = useWS()

  return (
    <div className="min-h-screen bg-gray-950 p-8">
      <div className="max-w-6xl mx-auto">

        <div className="flex items-center justify-between mb-8">
          <h1 className="text-white text-2xl font-bold">System Monitor</h1>
          <span className={`text-xs px-3 py-1 rounded-full font-medium ${connected ? 'bg-green-900 text-green-400' : 'bg-red-900 text-red-400'}`}>
            {connected ? 'Live' : 'Disconnected'}
          </span>
        </div>

        {!metrics ? (
          <p className="text-gray-500">Waiting for metrics...</p>
        ) : (
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
            <MetricCard title="CPU Usage" value={metrics.cpu_usage.toFixed(1)} unit="%" color="blue" />
            <MetricCard title="Memory Usage" value={((metrics.memory_used/metrics.memory_total)*100).toFixed(1)} unit="%" color="green" />
            <MetricCard title="Disk Usage" value={((metrics.disk_used/metrics.disk_total)*100).toFixed(1)} unit="%" color="yellow" />
            <MetricCard title="Network In" value={(metrics.net_download / 1024).toFixed(1)} unit="KB/s" color="red" />
          </div>
        )}

      </div>
    </div>
  )
}