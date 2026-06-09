import { useState, useMemo } from "react"
import {
  BarChart, Bar, XAxis, YAxis, Tooltip,
  LineChart, Line, CartesianGrid, ResponsiveContainer, Legend
} from "recharts"

export default function ResourceConsumers({ processes, metrics, filterByTimeRange, timeRange }) {
  const [tab, setTab] = useState("cpu") // "cpu" | "memory"

  // --- Top 5 processes by selected metric ---
const topProcesses = useMemo(() => {
  const field = tab === "cpu" ? "cpu_percentage" : "memory_percentage"

  // Group by name, sum their values
  const grouped = {}
  processes.forEach(p => {
    const val = p[field] || 0
    if (grouped[p.name]) {
      grouped[p.name] += val
    } else {
      grouped[p.name] = val
    }
  })

  // Convert to array, sort, take top 5
  return Object.entries(grouped)
    .map(([name, value]) => ({ name, value: parseFloat(value.toFixed(2)) }))
    .sort((a, b) => b.value - a.value)
    .slice(0, 5)
}, [processes, tab])

  // --- System trend data from /stats/history ---

  return (
    <div className="bg-white rounded-xl border border-gray-200 dark:bg-gray-700 p-6 mb-6">
      {/* Current Top Consumers */}
      <div>
        <div className="flex items-center justify-between mb-4">
          <div className="flex items-center gap-4">
            <span className="text-orange-500">⚡</span>
            <h2 className="text-base font-semibold text-gray-900 dark:text-gray-200">Current Top Consumers</h2>
            <span className="text-xs text-gray-400 dark:bg-gray-400 dark:text-gray-900 bg-gray-100 px-2 py-0.5 rounded-full">
              Live Snapshot
            </span>
          </div>
          {/* CPU / Memory toggle */}
          <div className="flex items-center bg-gray-100 rounded-lg p-1 dark:bg-gray-400">
            {["cpu", "memory"].map(t => (
              <button
                key={t}
                onClick={() => setTab(t)}
                className={`text-xs px-3 py-1 rounded-md transition-colors${
                  tab === t
                    ? "bg-white text-gray-900 shadow-sm dark:bg-gray-200 font-medium hover:cursor-pointer"
                    : "text-gray-900 hover:text-gray-700 dark:hover:text-gray-900 hover:cursor-pointer"
                }`}
              >
                {t === "cpu" ? "CPU" : "Memory"}
              </button>
            ))}
          </div>
        </div>

        {topProcesses.length === 0 ? (
          <div className="flex items-center justify-center py-10">
            <p className="text-sm text-gray-400 dark:text-gray-300">No process data available.</p>
          </div>
        ) : (
          <ResponsiveContainer width="100%" height={220}>
            <BarChart
              data={topProcesses}
              layout="vertical"
              margin={{ left: 80, right: 40, top: 4, bottom: 4 }}
            >
              <XAxis
                type="number"
                domain={[0, 100]}
                tick={{ fontSize: 11, fill: "#9ca3af" }}
                tickFormatter={v => `${v}%`
                }
              />
              <YAxis
                type="category"
                dataKey="name"
                tick={{ fontSize: 11, fill: "#6b7280" }}
                width={75}
              />
              <Tooltip
                formatter={(value) => [`${value}%`, tab === "cpu" ? "CPU" : "Memory"]}
                contentStyle={{ fontSize: 12, borderRadius: 8, border: "1px solid #e5e7eb" }}
              />
              <Bar
                dataKey="value"
                fill={tab === "cpu" ? "#8b5cf6" : "#3b82f6"}
                radius={[0, 4, 4, 0]}
              />
            </BarChart>
          </ResponsiveContainer>
        )}
      </div>
    </div>
  )
}