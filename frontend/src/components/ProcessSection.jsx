import { useState, useEffect } from 'react'
import { Play, Square, RotateCw } from 'lucide-react'
import api from '../api/axios'

export default function Processes() {
  const [processes, setProcesses] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)

  const fetchProcesses = async () => {
    try {
      const res = await api.get('/api/v1/processes/managed')
      setProcesses(res.data || [])
    } catch (err) {
      setError('Failed to fetch processes')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchProcesses()
    const interval = setInterval(fetchProcesses, 5000)
    return () => clearInterval(interval)
  }, [])

  const handleStop = async (id) => {
    try {
      await api.post(`/api/v1/processes/stop/${id}`)
      fetchProcesses()
    } catch (err) {
      console.error('Stop failed:', err)
    }
  }

  const handleRestart = async (id) => {
    try {
      await api.post(`/api/v1/processes/restart/${id}`)
      fetchProcesses()
    } catch (err) {
      console.error('Restart failed:', err)
    }
  }

  if (loading) return <div className="min-h-screen bg-white p-8 text-gray-400">Loading...</div>
  if (error) return <div className="min-h-screen mt-6 rounded-xl shadow bg-white p-8 text-red-400">{error}</div>

  return (
    <div className="min-h-screen bg-white p-8">
      <div className="max-w-4xl mx-auto">
        <h1 className="text-white text-2xl font-bold mb-8">Process Control</h1>
        <div className="bg-gray-900 rounded-xl border border-gray-800 p-6">
          <div className="space-y-3">
            {processes.length === 0 ? (
              <p className="text-gray-500 text-sm">No managed processes yet.</p>
            ) : (
              processes.map(process => (
                <div
                  key={process.id}
                  className="flex items-center justify-between p-3 bg-gray-800 rounded-lg hover:bg-gray-750 transition-colors"
                >
                  <div className="flex items-center gap-4 flex-1">
                    <div className={`w-2 h-2 rounded-full ${process.status === 'running' ? 'bg-green-500' : 'bg-gray-400'}`} />
                    <div className="flex-1">
                      <p className="text-white text-sm">{process.name}</p>
                      <p className="text-xs text-gray-500">
                        CPU: {process.cpu_percentage?.toFixed(1)}% • Memory: {process.memory_percentage?.toFixed(1)}%
                      </p>
                    </div>
                  </div>
                  <div className="flex gap-2">
                    <button
                      onClick={() => handleStop(process.id)}
                      disabled={process.status === 'stopped'}
                      className="p-2 bg-red-600 hover:bg-red-700 disabled:opacity-40 text-white rounded-md transition-colors"
                      title="Stop"
                    >
                      <Square className="w-4 h-4" />
                    </button>
                    <button
                      onClick={() => handleRestart(process.id)}
                      className="p-2 bg-blue-600 hover:bg-blue-700 text-white rounded-md transition-colors"
                      title="Restart"
                    >
                      <RotateCw className="w-4 h-4" />
                    </button>
                  </div>
                </div>
              ))
            )}
          </div>
        </div>
      </div>
    </div>
  )
}