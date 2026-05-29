import { useWS } from '../context/WSContext'
import MetricCard from '../components/MetricCard'
import Processes from './Processes'
import { Cpu, HardDrive, Activity, MemoryStick, Settings, LogOut} from 'lucide-react';
import HistoryChart from '../components/HistoricalCharts';
import ActiveAlerts from '../components/AlertSection';
import { useNavigate } from 'react-router-dom'
import { useAuth } from '../context/AuthContext'
export default function Dashboard() {
  const navigate = useNavigate()
  const { metrics, connected } = useWS()
  const { logout } = useAuth()
  const handleLogout = async () => {
    try {
      await api.post('/api/v1/logout')
    } catch (err) {
      console.error('Logout error:', err)
    } finally {
      logout()
      navigate('/login')
    }
  }

  return (
    <div className="min-h-screen bg-white p-8">
      <div className="max-w-6xl mx-auto">

        <div className="flex items-center justify-between mb-8">
          <div className="flex-col">
            <h1 className="text-black text-2xl font-bold">System Monitor</h1>
            <p className="text-gray-600">Real-time system metrics and process management</p>
          </div>
          <div className="flex flex-col gap-5">
            <div className="flex gap-2">
              <button onClick={() => navigate("/settings")} 
              className="bg-white shadow rounded-xl border border-black text-gray-600 text-xs p-2 hover:scale-110 transition-transform duration-200 cursor-pointer">
                <div className="flex items-center gap-1">
                  <Settings className="w-4 h-4" />
                  <span>Settings</span>
                </div>
              </button>
              <button onClick={handleLogout}
              className="bg-white shadow rounded-xl border border-black text-gray-600 text-xs p-2 hover:scale-110 transition-transform duration-200 cursor-pointer">
                <div className="flex items-center gap-1">
                  <LogOut className="w-4 h-4" />
                  <span>Logout</span>
                </div>
              </button>
            </div>
            <span className={`text-xs px-3 py-1 rounded-full font-medium ${connected ? 'bg-green-900 text-green-400' : 'bg-red-900 text-red-400'}`}>
              {connected ? 'Live' : 'Disconnected'}
            </span>
          </div>
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
        
      <div className="flex flex-col lg:flex-row gap-6 mt-6">
        <div className="flex-2 min-w-0">
          <HistoryChart />
        </div>
        <div className="flex-1 min-w-0">
          <Processes />
        </div>
      </div>
      <ActiveAlerts/>
      </div>
    </div>
  )
}