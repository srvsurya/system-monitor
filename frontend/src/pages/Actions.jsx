import api from '../api/axios'

import { useState, useEffect, useRef, useCallback } from 'react'
import { ChevronUp, ChevronDown, ChevronsUpDown } from 'lucide-react'

import { useNavigate } from 'react-router-dom' 

const PAGE_SIZE = 30
export default function Action(){
    const navigate = useNavigate()
    const [sortKey, setSortKey] = useState('created_at')
    const [sortDir, setSortDir] = useState('desc')
    const [visibleCount, setVisibleCount] = useState(PAGE_SIZE)
    const [search, setSearch] = useState('')
    const bottomRef = useRef(null)
    const [actions,setActions] = useState([])
    const [loading, setLoading] = useState(true)

    useEffect(() => {
    setVisibleCount(PAGE_SIZE)
    }, [search, sortKey, sortDir])


    const handleSort = (key) => {
    if (sortKey === key) {
        setSortDir(d => d === 'asc' ? 'desc' : 'asc')
    } else {
        setSortKey(key)
        setSortDir(key === 'action_type' ? 'asc' : 'desc')
    }
    }

    const sorted = [...actions]
    .filter(a => a.action_type.toLowerCase().includes(search.toLowerCase()))
    .sort((a, b) => {
        const aVal = a[sortKey] ?? ''
        const bVal = b[sortKey] ?? ''
        if (sortKey === 'created_at') {
        return sortDir === 'asc'
            ? new Date(aVal) - new Date(bVal)
            : new Date(bVal) - new Date(aVal)
        }
        return sortDir === 'asc'
        ? String(aVal).localeCompare(String(bVal))
        : String(bVal).localeCompare(String(aVal))
    })

    const visible = sorted.slice(0, visibleCount)
    const hasMore = visibleCount < sorted.length

    const observerRef = useCallback((node) => {
    if (!node) return
    const observer = new IntersectionObserver(
        (entries) => {
        if (entries[0].isIntersecting && hasMore) {
            setVisibleCount(c => c + PAGE_SIZE)
        }
        },
        { threshold: 0.1 }
    )
    observer.observe(node)
    bottomRef.current = observer
    return () => observer.disconnect()
    }, [hasMore])

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
    useEffect(() => {
    const fetchActions = async () => {
      try {
        const res = await api.get('/api/v1/processes/actions')
        setActions(res.data || [])
      } catch (err) {
        console.error('Failed to fetch actions:', err)
      } finally {
        setLoading(false)
      }
    }
    fetchActions()
  }, [])


return (
  <div className="bg-white p-4 dark:bg-gray-800 shadow border border-gray-200 dark:border-gray-700 min-h-[calc(100vh-0px)]">
    <div className="flex justify-between">
      <h1 className="text-xl font-semibold text-gray-900 dark:text-gray-200 mb-6">
        System Actions
      </h1>
      <button onClick={() => navigate("/")} className="text-sm text-gray-500 hover:text-gray-700 hover:cursor-pointer mb-6 dark:hover:text-white flex items-center gap-1">
        ← Back to Dashboard
      </button>
    </div>

    <div className="bg-white dark:bg-gray-800 rounded-xl shadow border border-gray-200 dark:border-gray-700 min-h-[calc(100vh-200px)]">
      <div className="p-4 border-b border-gray-100 dark:border-gray-700">
        <input
          type="text"
          placeholder="Search by action type..."
          value={search}
          onChange={e => setSearch(e.target.value)}
          className="w-full border border-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-blue-500"
        />
      </div>

      <div className="overflow-y-auto">
        {loading ? (
          <p className="text-gray-400 dark:text-gray-300 text-sm p-4">Loading actions...</p>
        ) : sorted.length === 0 ? (
          <p className="text-gray-400 dark:text-gray-300 text-sm p-4">No actions found.</p>
        ) : (
          <table className="w-full">
            <thead className="bg-gray-50 dark:bg-gray-900 sticky top-0">
              <tr>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                  Process ID
                </th>
                <HeaderCell col="action_type" label="Action Type" />
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                  Reason
                </th>
                <HeaderCell col="created_at" label="Created At" />
              </tr>
            </thead>
            <tbody>
              {visible.map(a => (
                <tr
                  key={a.id}
                  className="border-t border-gray-100 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-700"
                >
                  <td className="px-4 py-3 text-sm text-gray-600 dark:text-gray-400">
                    {a.process_id ?? '—'}
                  </td>
                  <td className="px-4 py-3">
                    <span className="text-xs font-medium bg-blue-100 dark:bg-blue-900 text-blue-700 dark:text-blue-300 px-2 py-1 rounded-full">
                      {a.action_type}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-sm text-gray-600 dark:text-gray-400">
                    {a.reason}
                  </td>
                  <td className="px-4 py-3 text-sm text-gray-400">
                    {new Date(a.created_at).toLocaleString()}
                  </td>
                </tr>
              ))}
              {hasMore && (
                <tr>
                  <td colSpan={5}>
                    <div ref={observerRef} className="h-8" />
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        )}
      </div>
    </div>
  </div>
)
}