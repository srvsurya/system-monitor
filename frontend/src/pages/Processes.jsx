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
      <div className="bg-white rounded-xl w-full max-w-lg mx-4 shadow-xl dark:bg-gray-700">
        <div className="flex items-center justify-between p-4 border-b border-gray-200">
          <h2 className="font-semibold text-gray-900 dark:text-gray-200">Register Process</h2>
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
            className="w-full border border-gray-300 dark:text-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-blue-500"
          />
        </div>
        <div className="overflow-y-auto max-h-96">
          {loading ? (
            <p className="text-gray-400 dark:text-gray-300 text-sm p-4">Loading processes...</p>
          ) : filtered.length === 0 ? (
            <p className="text-gray-400 dark:text-gray-300 text-sm p-4">No processes found.</p>
          ) : (
            filtered.map(p => (
              <div key={p.pid} className="flex items-center justify-between px-4 py-3 hover:bg-gray-50 dark:hover:bg-gray-700 border-b border-gray-100">
                <div>
                  <p className="text-sm font-medium text-gray-900 dark:text-gray-300">{p.name}</p>
                  <p className="text-xs text-gray-400">PID: {p.pid} • CPU: {p.cpu_percentage?.toFixed(2)}% • Mem: {p.memory_percentage?.toFixed(2)}%</p>
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
  const [toast, setToast] = useState(null)

  const showToast = (msg, type = 'info') => {
    setToast({ msg, type })
    setTimeout(() => setToast(null), 3000)
  }

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
const handleRemove = (process) => {
  if (process.pinned) {
    showToast("You must remove Pin in order to remove this process from the list", "error")
    return
  }
  api.delete(`/api/v1/processes/${process.id}/remove`)
    .then(() => fetchProcesses())
    .catch(() => showToast("Failed to remove process", "error"))
}


  if (loading) return <div className="min-h-screen bg-gray-50 p-8 text-gray-400 dark:bg-gray-700">Loading...</div>
  if (error) return <div className="min-h-screen bg-gray-50 p-8 text-red-400">{error}</div>

  return (
    <div className=" bg-gray-50 p-8 mt-5 rounded-xl shadow dark:bg-gray-700 border border-gray-200">
      <div className="max-w-4xl">
        <div className="flex items-center justify-between gap-4 mb-8">
          <h1 className="text-xl font-bold text-gray-900 dark:text-gray-300">Process Control</h1>
          <button
            onClick={() => setShowModal(true)}
            className="flex items-center gap-2 bg-blue-600 hover:bg-blue-700 text-white text-sm px-4 py-2 rounded-lg hover:cursor-pointer"
          >
            <Plus className="w-4 h-4" />
            Register Process
          </button>
        </div>

        <div className="bg-white rounded-xl border border-gray-200 p-6 dark:bg-gray-700">
          <div className="space-y-3">
            {processes.length === 0 ? (
              <p className="text-gray-400 text-sm dark:text-gray-300">No managed processes yet.</p>
            ) : (
                  processes.map(process => (
                <div
                  key={process.id}
                  className={`relative flex items-center justify-between p-3 rounded-lg transition-colors dark:bg-gray-600 ${
                    process.status === 'stopped'
                      ? 'bg-gray-100 opacity-60'
                      : 'bg-gray-50'
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
                        <p className="text-sm font-medium text-gray-900 dark:text-gray-300">{process.name}</p>
                        {process.pinned && <Pin className="w-3 h-3 text-blue-500" />}
                      </div>
                      <p className="text-xs text-gray-500 dark:text-gray-300">
                        PID: {process.pid} • CPU: {process.cpu_percentage?.toFixed(2)}% • Memory: {process.memory_percentage?.toFixed(2)}%
                      </p>
                    </div>
                  </div>
                   <button
                        onClick={() => handleRemove(process)}
                        className="absolute -top-2 -left-2 w-4 h-4 bg-gray-200 hover:bg-red-100 hover:text-red-500 text-gray-400 rounded-full flex items-center justify-center text-xs transition-colors cursor-pointer"
                        title="Remove from managed"
                      >
                        ×
                    </button>

                  <div className="flex flex-col gap-2">
                    <button
                      onClick={() => handlePin(process.id)}
                      className={`p-2 rounded-md transition-colors hover:cursor-pointer ${
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
                      className="p-2 bg-red-500 hover:bg-red-600 disabled:opacity-40 text-white rounded-md transition-colors hover:cursor-pointer"
                      title="Stop"
                    >
                      <Square className="w-4 h-4" />
                    </button>
                    <button
                      onClick={() => handleRestart(process.id)}
                      className="p-2 bg-blue-500 hover:bg-blue-600 text-white rounded-md transition-colors hover:cursor-pointer"
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
      {toast && (
        <div className={`fixed bottom-6 right-6 px-4 py-3 rounded-lg text-white text-sm shadow-lg z-50 ${
          toast.type === 'error' ? 'bg-gray-500' : 'bg-green-500'
        }`}>
          {toast.msg}
        </div>
      )}
    </div>
    
  )
}