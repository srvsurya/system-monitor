import { useWS } from '../context/WSContext'
import MetricCard from '../components/MetricCard'
import Processes from '../components/ProcessSection'
import { Cpu, HardDrive, Activity, MemoryStick} from 'lucide-react';

export default function Dashboard() {
  const { metrics, connected } = useWS()

  return (
    <div className="min-h-screen bg-white p-8">
      <div className="max-w-6xl mx-auto">

        <div className="flex items-center justify-between mb-8">
          <div className="flex-col">
          <h1 className="text-black text-2xl font-bold">System Monitor</h1>
          <p className="text-gray-600">Real-time system metrics and process management</p>
          </div>
          <span className={`text-xs px-3 py-1 rounded-full font-medium ${connected ? 'bg-green-900 text-green-400' : 'bg-red-900 text-red-400'}`}>
            {connected ? 'Live' : 'Disconnected'}
          </span>
        </div>

        {!metrics ? (
          <p className="text-gray-500">Waiting for metrics...</p>
        ) : (
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
            <MetricCard title="CPU Usage" value={metrics.cpu_usage.toFixed(1)} unit="%"icon={<Cpu className="w-6 h-6" />}/>
            <MetricCard title="Memory Usage" value={((metrics.memory_used/metrics.memory_total)*100).toFixed(1)} unit="%" icon={<MemoryStick className="w-6 h-6" />} />
            <MetricCard title="Disk Usage" value={((metrics.disk_used/metrics.disk_total)*100).toFixed(1)} unit="%" icon={<HardDrive className="w-6 h-6" />} />
            <MetricCard title="Network In" value={(metrics.net_download / 1024).toFixed(1)} unit="KB/s" icon={<Activity className="w-6 h-6" />}/>
          </div>
        )}
      <div className="flex justify-end"><Processes/></div>
      </div>
    </div>
  )
}