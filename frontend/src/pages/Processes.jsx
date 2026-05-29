import { useState, useEffect } from 'react'
import { Play, Square, RotateCw, X, Plus, Pin } from 'lucide-react'
import api from '../api/axios'

function RegisterModal({ onClose, onRegistered }) {
  const [allProcesses, setAllProcesses] = useState([])
  const [search, setSearch] = useState('')
  const [loading, setLoading] = useState(true)
  const [registering, setRegistering] = useState(null)

  useEffect(() => {
    const fetch = async () => {
      try {
        const res = await api.get('/api/v1/processes')
        setAllProcesses(res.data || [])
      } catch (err) {
        console.error('Failed to fetch processes:', err)
      } finally {
        setLoading(false)
      }
    }
    fetch()
  }, [])

  const handleRegister = async (pid) => {
    setRegistering(pid)
    try {
      await api.post(`/api/v1/processes/register/${pid}`)
      onRegistered()
      onClose()
    } catch (err) {
      console.error('Register failed:', err)
    } finally {
      setRegistering(null)
    }
  }

  const filtered = allProcesses.filter(p =>
    p.name.toLowerCase().includes(search.toLowerCase())
  )

  return (
    <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50">
      <div className="bg-white rounded-xl w-full max-w-lg mx-4 shadow-xl">
        <div className="flex items-center justify-between p-4 border-b border-gray-200">
          <h2 className="font-semibold text-gray-900">Register Process</h2>
          <button onClick={onClose} className="text-gray-400 hover:text-gray-600">
            <X className="w-5 h-5" />
          </button>
        </div>
        <div className="p-4 border-b border-gray-100">
          <input
            type="text"
            placeholder="Search by name..."
            value={search}
            onChange={e => setSearch(e.target.value)}
            className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-blue-500"
          />
        </div>
        <div className="overflow-y-auto max-h-96">
          {loading ? (
            <p className="text-gray-400 text-sm p-4">Loading processes...</p>
          ) : filtered.length === 0 ? (
            <p className="text-gray-400 text-sm p-4">No processes found.</p>
          ) : (
            filtered.map(p => (
              <div key={p.pid} className="flex items-center justify-between px-4 py-3 hover:bg-gray-50 border-b border-gray-100">
                <div>
                  <p className="text-sm font-medium text-gray-900">{p.name}</p>
                  <p className="text-xs text-gray-400">PID: {p.pid} • CPU: {p.cpu_percentage?.toFixed(1)}% • Mem: {p.memory_percentage?.toFixed(1)}%</p>
                </div>
                <button
                  onClick={() => handleRegister(p.pid)}
                  disabled={registering === p.pid}
                  className="text-xs bg-blue-600 hover:bg-blue-700 disabled:opacity-50 text-white px-3 py-1.5 rounded-lg"
                >
                  {registering === p.pid ? '...' : 'Register'}
                </button>
              </div>
            ))
          )}
        </div>
      </div>
    </div>
  )
}

export default function Processes() {
  const [processes, setProcesses] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [showModal, setShowModal] = useState(false)

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
  const handlePin = async (id) => {
  try {
    await api.patch(`/api/v1/processes/${id}`)
    fetchProcesses()
  } catch (err) {
    console.error('Pin failed:', err)
  }
}

  if (loading) return <div className="min-h-screen bg-gray-50 p-8 text-gray-400">Loading...</div>
  if (error) return <div className="min-h-screen bg-gray-50 p-8 text-red-400">{error}</div>

  return (
    <div className=" bg-gray-50 p-8 mt-5 rounded-xl shadow">
      <div className="max-w-4xl">
        <div className="flex items-center justify-between gap-4 mb-8">
          <h1 className="text-xl font-bold text-gray-900">Process Control</h1>
          <button
            onClick={() => setShowModal(true)}
            className="flex items-center gap-2 bg-blue-600 hover:bg-blue-700 text-white text-sm px-4 py-2 rounded-lg"
          >
            <Plus className="w-4 h-4" />
            Register Process
          </button>
        </div>

        <div className="bg-white rounded-xl border border-gray-200 p-6">
          <div className="space-y-3">
            {processes.length === 0 ? (
              <p className="text-gray-400 text-sm">No managed processes yet.</p>
            ) : (
                  processes.map(process => (
                <div
                  key={process.id}
                  className={`flex items-center justify-between p-3 rounded-lg transition-colors ${
                    process.status === 'stopped'
                      ? 'bg-gray-100 opacity-60'
                      : 'bg-gray-50 hover:bg-gray-100'
                  }`}
                >
                  <div className="flex items-center gap-4 flex-1">
                    <div
                      className={`w-2 h-2 rounded-full ${
                        process.status === 'running' ? 'bg-green-500' : 'bg-gray-400'
                      }`}
                    />
                    <div className="flex-1">
                      <div className="flex items-center gap-2">
                        <p className="text-sm font-medium text-gray-900">{process.name}</p>
                        {process.pinned && <Pin className="w-3 h-3 text-blue-500" />}
                      </div>
                      <p className="text-xs text-gray-500">
                        PID: {process.pid} • CPU: {process.cpu_percentage?.toFixed(1)}% • Memory: {process.memory_percentage?.toFixed(1)}%
                      </p>
                    </div>
                  </div>

                  <div className="flex flex-col lg:flex-row gap-2">
                    <button
                      onClick={() => handlePin(process.id)}
                      className={`p-2 rounded-md transition-colors ${
                        process.pinned
                          ? 'bg-blue-100 text-blue-600'
                          : 'bg-gray-200 text-gray-500 hover:bg-gray-300'
                      }`}
                      title={process.pinned ? 'Unpin' : 'Pin'}
                    >
                      <Pin className="w-4 h-4" />
                    </button>
                    <button
                      onClick={() => handleStop(process.id)}
                      disabled={process.status === 'stopped'}
                      className="p-2 bg-red-500 hover:bg-red-600 disabled:opacity-40 text-white rounded-md transition-colors"
                      title="Stop"
                    >
                      <Square className="w-4 h-4" />
                    </button>
                    <button
                      onClick={() => handleRestart(process.id)}
                      className="p-2 bg-blue-500 hover:bg-blue-600 text-white rounded-md transition-colors"
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

      {showModal && (
        <RegisterModal
          onClose={() => setShowModal(false)}
          onRegistered={fetchProcesses}
        />
      )}
    </div>
  )
}