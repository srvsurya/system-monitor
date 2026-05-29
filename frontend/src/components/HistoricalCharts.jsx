import { useEffect, useState } from 'react'
import { LineChart, Line, XAxis, YAxis, Tooltip, ResponsiveContainer } from 'recharts'
import api from '../api/axios'

const pad = (n) => String(n).padStart(2, '0')

const formatLocal = (date) => {
  return `${date.getFullYear()}-${pad(date.getMonth()+1)}-${pad(date.getDate())}T${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(date.getSeconds())}`
}

export default function HistoryChart({ className = '' }) {
  const [data, setData] = useState([])
  const [active, setActive] = useState('cpu')
  const [range, setRange] = useState(30)

  useEffect(() => {
    const fetch = async () => {
      try {
        const now = new Date()
        const from = new Date(now.getTime() - range * 60 * 1000)
        const res = await api.get('/api/v1/stats/history', {
          params: {
            from: formatLocal(from),
            to: formatLocal(now),
            limit: 1000,
          }
        })
        const reversed = [...(res.data || [])].reverse()
        if (reversed.length === 0) return
        const formatted = reversed.map(m => ({
          time: new Date(m.timestamp.replace('Z', '')).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }),
          cpu: parseFloat(m.cpu_usage.toFixed(1)),
          memory: parseFloat(((m.memory_used / m.memory_total) * 100).toFixed(1)),
          disk: parseFloat(((m.disk_used / m.disk_total) * 100).toFixed(1)),
          network: parseFloat((m.net_download / 1024).toFixed(1)),
        }))
        setData(formatted)
      } catch (err) {
        console.error('Failed to fetch history:', err)
      }
    }
    fetch()
    const interval = setInterval(fetch, 10000)
    return () => clearInterval(interval)
  }, [range])

  const tabs = [
    { key: 'cpu',     label: 'CPU',     color: '#3b82f6' },
    { key: 'memory',  label: 'Memory',  color: '#8b5cf6' },
    { key: 'disk',    label: 'Disk',    color: '#f59e0b' },
    { key: 'network', label: 'Network', color: '#10b981' },
  ]

  const ranges = [
    { label: '30m', minutes: 30 },
    { label: '1h',  minutes: 60 },
    { label: '6h',  minutes: 360 },
    { label: '24h', minutes: 1440 },
  ]

  const current = tabs.find(t => t.key === active)

  return (
    <div className={`bg-white rounded-xl border border-gray-200 p-6 mt-6 ${className}`}>
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3 mb-4">
        <h2 className="font-semibold text-gray-900">History</h2>
        <div className="flex flex-wrap gap-2">
          {ranges.map(r => (
            <button
              key={r.minutes}
              onClick={() => setRange(r.minutes)}
              className={`text-xs px-3 py-1 rounded-full font-medium transition-colors ${
                range === r.minutes
                  ? 'bg-gray-900 text-white'
                  : 'bg-gray-100 text-gray-500 hover:bg-gray-200'
              }`}
            >
              {r.label}
            </button>
          ))}
        </div>
        <div className="flex flex-wrap gap-2">
          {tabs.map(t => (
            <button
              key={t.key}
              onClick={() => setActive(t.key)}
              className={`text-xs px-3 py-1 rounded-full font-medium transition-colors ${
                active === t.key
                  ? 'bg-blue-600 text-white'
                  : 'bg-gray-100 text-gray-500 hover:bg-gray-200'
              }`}
            >
              {t.label}
            </button>
          ))}
        </div>
      </div>

      <ResponsiveContainer width="100%" height={250}>
        <LineChart data={data} margin={{ top: 5, right: 10, left: 0, bottom: 5 }}>
          <XAxis
            dataKey="time"
            tick={{ fontSize: 11, fill: '#9ca3af' }}
            tickLine={false}
            axisLine={false}
            interval="preserveStartEnd"
          />
          <YAxis
            tick={{ fontSize: 11, fill: '#9ca3af' }}
            tickLine={false}
            axisLine={false}
            domain={active === 'network' ? ['auto', 'auto'] : [0, 100]}
            unit={active === 'network' ? ' KB/s' : '%'}
          />
          <Tooltip
            contentStyle={{ backgroundColor: '#fff', border: '1px solid #e5e7eb', borderRadius: '8px', fontSize: '12px' }}
            formatter={(val) => [
              `${val}${active === 'network' ? ' KB/s' : '%'}`,
              current.label
            ]}
          />
          <Line
            type="monotone"
            dataKey={active}
            stroke={current.color}
            strokeWidth={2}
            dot={false}
            activeDot={{ r: 4 }}
          />
        </LineChart>
      </ResponsiveContainer>
    </div>
  )
}