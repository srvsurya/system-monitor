import {useState, useEffect} from 'react'
import { useWS } from '../context/WSContext'
import MetricCard from '../components/MetricCard'
import ProcessesManaged from '../components/Processes'
import { Cpu, HardDrive, Activity, MemoryStick, Settings, LogOut, Zap, Shield} from 'lucide-react';
import HistoryChart from '../components/HistoricalCharts';
import ActiveAlerts from '../components/AlertSection';
import { useNavigate } from 'react-router-dom'
import { useAuth } from '../context/AuthContext'
import api from '../api/axios'
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
  useEffect(() => {
  const fetchHealState = async () => {
      try {
          const res = await api.get('/api/v1/cleaner/settings')
          setHealEnabled(!!res.data.smart_heal_enabled)
      } catch (err) {
          console.error('Failed to fetch heal state')
      }
  }
  fetchHealState()
  }, [])

  const [optimizing, setOptimizing] = useState(false)
  const [toast, setToast] = useState(null)
  const [healEnabled, setHealEnabled] = useState(false)
  const [healLoading, setHealLoading] = useState(false)


const showToast = (message, type = 'success') => {
  setToast({ message, type })
  setTimeout(() => setToast(null), 4000)
}

const handleOptimize = async () => {
  setOptimizing(true)
  try {
    const res = await api.post('/api/v1/cleaner/optimize')
    const { killed, skipped, errors } = res.data
    if (killed.length === 0 && errors?.length > 0) {
      showToast('Optimization failed', 'error')
    } else if (killed.length === 0) {
      showToast(`System is clean — nothing to optimize (${skipped} processes checked)`)
    } else {
      showToast(`Killed ${killed.length} process${killed.length > 1 ? 'es' : ''}, ${skipped} skipped`)
    }
  } catch (err) {
    showToast('Optimization failed', 'error')
  } finally {
    setOptimizing(false)
  }
}

const handleHealToggle = async () => {
  setHealLoading(true)
  try {
    const next = !healEnabled
    await api.post('/api/v1/healer/toggle', { enabled: next })
    setHealEnabled(next)
    showToast(`Smart Heal ${next ? 'enabled' : 'disabled'}`)
  } catch (err) {
    showToast('Failed to update Smart Heal', 'error')
  } finally {
    setHealLoading(false)
  }
}


  return (
    <div className="min-h-screen bg-white dark:bg-gray-900 p-8">
      <div className="max-w-6xl mx-auto">

        <div className="flex items-center justify-between mb-8">
          <div className="flex-col">
            <h1 className="text-black text-2xl font-bold dark:text-gray-200">System Monitor</h1>
            <p className="text-gray-600 dark:text-gray-400">Real-time system metrics and process management</p>
          </div>
          <div className="flex flex-col gap-5">
            <div className="flex gap-2">
              <button onClick={() => navigate("/settings")} 
              className="bg-white shadow rounded-xl border border-black text-gray-600 text-xs p-2 hover:scale-110 transition-transform duration-200 cursor-pointer dark:bg-gray-600">
                <div className="flex items-center gap-1">
                  <Settings className="w-4 h-4 dark:text-white" />
                  <span className="dark:text-white">Settings</span>
                </div>
              </button>
              <button onClick={handleLogout}
              className="bg-white shadow rounded-xl border border-black text-gray-600 text-xs p-2 hover:scale-110 transition-transform duration-200 cursor-pointer dark:bg-gray-600">
                <div className="flex items-center gap-1">
                  <LogOut className="w-4 h-4 dark:text-white" />
                  <span className="dark:text-white">Logout</span>
                </div>
              </button>
            </div>
            <span className={`text-xs px-3 py-1 rounded-full font-medium ${connected ? 'bg-green-900 text-green-400' : 'bg-red-900 text-red-400'}`}>
              {connected ? 'Live' : 'Disconnected'}
            </span>
          </div>
        </div>
        
        {toast && (
          <div className={`fixed top-6 right-6 z-50 px-4 py-3 rounded-xl shadow-lg text-sm font-medium transition-all ${
            toast.type === 'error' ? 'bg-red-50 border border-red-200 text-red-700' : 'bg-green-50 border border-green-200 text-green-700'
          }`}>
            {toast.message}
          </div>
        )}

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


        
      <div className="flex flex-col lg:flex-row gap-6 mt-4">
        
        <div className="flex-1 min-w-0">
          <div>
            <HistoryChart />
            <ActiveAlerts/>
          </div>
        </div>
        <div className="flex flex-col lg:w-100 shrink-0 overflow-hidden">
          <div className="flex justify-between gap-4 mt-6">
            <button
              onClick={handleOptimize}
              disabled={optimizing}
              className="flex items-center ml-2 gap-2 bg-blue-600 text-white text-xs px-4 py-2 rounded-xl hover:scale-105 transition-transform duration-200 cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <Zap className="w-4 h-4" />
              {optimizing ? 'Optimizing...' : 'Optimize'}
            </button>
            <div className="flex items-center gap-3 bg-white border border-gray-200 rounded-xl px-4 py-2 shadow-sm dark:bg-gray-500">
              <Shield className="w-4 h-4 text-gray-500 dark:text-white" />
              <span className="text-xs text-gray-600 dark:text-white">Smart Heal</span>
              <button
                onClick={handleHealToggle}
                disabled={healLoading}
                className={`relative inline-flex h-5 w-9 items-center rounded-full transition-colors duration-200 focus:outline-none ${
                  healEnabled ? 'bg-black' : 'bg-gray-300'
                } ${healLoading ? 'opacity-50 cursor-not-allowed' : 'cursor-pointer'}`}
              >
                <span className={`inline-block h-3 w-3 transform rounded-full bg-white transition-transform duration-200 ${
                  healEnabled ? 'translate-x-5' : 'translate-x-1'
                }`} />
              </button>
            </div>
          </div>
          <div className="flex-1 min-w-0">
            <ProcessesManaged />
          </div>
          { /* System actions history button */ }
          <button
            onClick={() => navigate("/actions")}
            className="mt-4 w-fit ml-2 bg-blue-600 text-white text-xs px-4 py-2 rounded-xl hover:scale-105 transition-transform duration-200 cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
          >System Action Log
          </button>
          {/*stop*/}
        </div>
      </div>
      </div>
    </div>
  )
}