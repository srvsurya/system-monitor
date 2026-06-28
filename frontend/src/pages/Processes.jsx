import { useState, useEffect, useRef, useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import { ChevronUp, ChevronDown, ChevronsUpDown } from 'lucide-react'
import api from '../api/axios'

const PAGE_SIZE = 30

export default function Processes() {
  const [allProcesses, setAllProcesses] = useState([])
  const [search, setSearch] = useState('')
  const [loading, setLoading] = useState(true)
  const [registering, setRegistering] = useState(null)
  const [sortKey, setSortKey] = useState('cpu_percentage')
  const [sortDir, setSortDir] = useState('desc')
  const navigate = useNavigate()

  const [page, setPage] = useState(1)
  

  useEffect(() => {
    const fetchProcesses = async () => {
      try {
        const res = await api.get('/api/v1/processes')
        setAllProcesses(res.data || [])
      } catch (err) {
        console.error('Failed to fetch processes:', err)
      } finally {
        setLoading(false)
      }
    }
    fetchProcesses()
  }, [])

  // Reset visible count when search or sort changes
  useEffect(() => {
  setPage(1)
  }, [search, sortKey, sortDir])

  const handleSort = (key) => {
    if (sortKey === key) {
      setSortDir(d => d === 'asc' ? 'desc' : 'asc')
    } else {
      setSortKey(key)
      setSortDir(key === 'name' ? 'asc' : 'desc') // names default A-Z, numbers default high-low
    }
  }

  const handleRegister = async (pid) => {
    setRegistering(pid)
    try {
      await api.post(`/api/v1/processes/register/${pid}`)
      navigate('/')
    } catch (err) {
      console.error('Register failed:', err)
    } finally {
      setRegistering(null)
    }
  }

  const sorted = [...allProcesses]
    .filter(p => p.name.toLowerCase().includes(search.toLowerCase()))
    .sort((a, b) => {
      const aVal = a[sortKey] ?? 0
      const bVal = b[sortKey] ?? 0
      if (typeof aVal === 'string') {
        return sortDir === 'asc'
          ? aVal.localeCompare(bVal)
          : bVal.localeCompare(aVal)
      }
      return sortDir === 'asc' ? aVal - bVal : bVal - aVal
    })

  const totalPages = Math.ceil(sorted.length / PAGE_SIZE)
  const visible = sorted.slice((page - 1) * PAGE_SIZE, page * PAGE_SIZE)

  const SortIcon = ({ col }) => {
    if (sortKey !== col) return <ChevronsUpDown className="w-3 h-3 text-gray-400" />
    return sortDir === 'asc'
      ? <ChevronUp className="w-3 h-3 text-blue-500" />
      : <ChevronDown className="w-3 h-3 text-blue-500" />
  }

  const HeaderCell = ({ col, label }) => (
    <th
      onClick={() => handleSort(col)}
      className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider cursor-pointer select-none hover:text-gray-900 dark:hover:text-gray-200"
    >
      <span className="flex items-center gap-1">
        {label}
        <SortIcon col={col} />
      </span>
    </th>
  )

  return (
  <div className="bg-white p-4 dark:bg-gray-800 shadow border border-gray-200 dark:border-gray-700 min-h-[calc(100vh-0px)]">
    <div className="flex justify-between">
      <h1 className="text-xl font-semibold text-gray-900 dark:text-gray-100 mb-6">
        Register Process
      </h1>
      <button
        onClick={() => navigate('/')}
        className="text-sm text-gray-500 hover:text-gray-700 hover:cursor-pointer mb-6 dark:hover:text-white flex items-center gap-1"
      >
        ← Back to Dashboard
      </button>
    </div>

    <div className="bg-white dark:bg-gray-800 rounded-xl shadow border border-gray-200 dark:border-gray-700 min-h-[calc(100vh-200px)]">
      <div className="p-4 border-b border-gray-100 dark:border-gray-700">
        <input
          type="text"
          placeholder="Search by name..."
          value={search}
          onChange={e => setSearch(e.target.value)}
          className="w-full border border-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-blue-500"
        />
      </div>

      <div className="overflow-y-auto">
        {loading ? (
          <p className="text-gray-400 dark:text-gray-300 text-sm p-4">Loading processes...</p>
        ) : sorted.length === 0 ? (
          <p className="text-gray-400 dark:text-gray-300 text-sm p-4">No processes found.</p>
        ) : (
          <table className="w-full">
            <thead className="bg-gray-50 dark:bg-gray-900 sticky top-0">
              <tr>
                <HeaderCell col="name" label="Name" />
                <HeaderCell col="cpu_percentage" label="CPU %" />
                <HeaderCell col="memory_percentage" label="Mem %" />
                <th className="px-4 py-3" />
              </tr>
            </thead>
            <tbody>
              {visible.map(p => (
                <tr
                  key={p.pid}
                  className="border-t border-gray-100 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-700"
                >
                  <td className="px-4 py-3">
                    <p className="text-sm font-medium text-gray-900 dark:text-gray-300">{p.name}</p>
                    <p className="text-xs text-gray-400">PID: {p.pid}</p>
                  </td>
                  <td className="px-4 py-3 text-sm text-gray-600 dark:text-gray-400">
                    {p.cpu_percentage?.toFixed(2)}%
                  </td>
                  <td className="px-4 py-3 text-sm text-gray-600 dark:text-gray-400">
                    {p.memory_percentage?.toFixed(2)}%
                  </td>
                  <td className="px-4 py-3 text-right">
                    <button
                      onClick={() => handleRegister(p.pid)}
                      disabled={registering === p.pid}
                      className="text-xs bg-blue-600 hover:bg-blue-700 disabled:opacity-50 text-white px-3 py-1.5 rounded-lg"
                    >
                      {registering === p.pid ? '...' : 'Register'}
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>

      {!loading && sorted.length > 0 && (
        <div className="flex items-center justify-between px-4 py-3 border-t border-gray-100 dark:border-gray-700">
          <span className="text-xs text-gray-400">
            Page {page} of {totalPages}
          </span>
          <div className="flex gap-2">
            <button
              onClick={() => setPage(p => p - 1)}
              disabled={page === 1}
              className="text-xs px-3 py-1.5 rounded-lg bg-gray-100 dark:bg-gray-700 disabled:opacity-40 hover:bg-gray-200 dark:hover:bg-gray-600 hover:cursor-pointer dark:text-gray-300"
            >
              Prev
            </button>
            <button
              onClick={() => setPage(p => p + 1)}
              disabled={page === totalPages}
              className="text-xs px-3 py-1.5 rounded-lg bg-gray-100 dark:bg-gray-700 disabled:opacity-40 hover:bg-gray-200 dark:hover:bg-gray-600 hover:cursor-pointer dark:text-gray-300 "
            >
              Next
            </button>
          </div>
        </div>
      )}
    </div>
  </div>
)
}