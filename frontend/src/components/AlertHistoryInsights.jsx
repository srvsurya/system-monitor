import { useMemo, useState } from "react"

const formatMetric = (metric) => {
  const map = {
    cpu_usage: "CPU Usage",
    memory_usage: "Memory Usage",
    disk_usage: "Disk Usage",
    net_upload: "Network Upload",
    net_download: "Network Download",
  }
  return map[metric] || metric
}

const getSeverity = (value, threshold) => {
  if (value >= threshold * 1.5) return { label: "CRITICAL", color: "bg-red-100 text-red-700" }
  if (value >= threshold * 1.2) return { label: "HIGH",     color: "bg-orange-100 text-orange-700" }
  return                               { label: "WARNING",  color: "bg-yellow-100 text-yellow-700" }
}

export default function AlertHistorySection({ alertHistory, ruleMap, timeRange, filterByTimeRange }) {
  const [sortBy, setSortBy] = useState("triggered_at") // "triggered_at" | "value" | "metric"
  const [sortDir, setSortDir] = useState("desc")        // "asc" | "desc"
  const [metricFilter, setMetricFilter] = useState("all")

  // Filter by time range
  const timeFiltered = useMemo(
    () => filterByTimeRange(alertHistory, "triggered_at"),
    [alertHistory, timeRange]
  )

  // Unique metrics for the filter dropdown
  const availableMetrics = useMemo(() => {
    const set = new Set(timeFiltered.map(a => a.metric).filter(Boolean))
    return ["all", ...set]
  }, [timeFiltered])

  // Filter by metric
  const metricFiltered = useMemo(() => {
    if (metricFilter === "all") return timeFiltered
    return timeFiltered.filter(a => a.metric === metricFilter)
  }, [timeFiltered, metricFilter])

  // Sort
  const sorted = useMemo(() => {
    return [...metricFiltered].sort((a, b) => {
      let aVal, bVal
      if (sortBy === "triggered_at") {
        aVal = new Date(a.triggered_at).getTime()
        bVal = new Date(b.triggered_at).getTime()
      } else if (sortBy === "value") {
        aVal = a.value
        bVal = b.value
      } else if (sortBy === "metric") {
        aVal = a.metric
        bVal = b.metric
      }
      if (aVal < bVal) return sortDir === "asc" ? -1 : 1
      if (aVal > bVal) return sortDir === "asc" ? 1 : -1
      return 0
    })
  }, [metricFiltered, sortBy, sortDir])

  const handleSort = (col) => {
    if (sortBy === col) {
      setSortDir(d => d === "asc" ? "desc" : "asc")
    } else {
      setSortBy(col)
      setSortDir("desc")
    }
  }

  const SortIcon = ({ col }) => {
    if (sortBy !== col) return <span className="text-gray-300 ml-1">↕</span>
    return <span className="text-purple-500 ml-1">{sortDir === "asc" ? "↑" : "↓"}</span>
  }

  return (
    <div className="bg-white rounded-xl border border-gray-200 p-6 mb-6">
      {/* Header */}
      <div className="flex items-center justify-between mb-4">
        <div className="flex items-center gap-2">
          <span className="text-yellow-500">⚠</span>
          <h2 className="text-base font-semibold text-gray-900">Alert History</h2>
          <span className="text-xs text-gray-400 bg-gray-100 px-2 py-0.5 rounded-full">
            {sorted.length} alert{sorted.length !== 1 ? "s" : ""}
          </span>
        </div>

        {/* Metric filter */}
        <select
          value={metricFilter}
          onChange={e => setMetricFilter(e.target.value)}
          className="text-xs border border-gray-200 rounded-lg px-3 py-1.5 bg-white text-gray-700 focus:outline-none focus:ring-1 focus:ring-purple-500"
        >
          {availableMetrics.map(m => (
            <option key={m} value={m}>
              {m === "all" ? "All metrics" : formatMetric(m)}
            </option>
          ))}
        </select>
      </div>

      {/* Empty state */}
      {sorted.length === 0 && (
        <div className="flex flex-col items-center justify-center py-10 text-center">
          <span className="text-3xl mb-2">✅</span>
          <p className="text-sm text-gray-500">
            No alerts were triggered in this time range.
          </p>
        </div>
      )}

      {/* Table */}
      {sorted.length > 0 && (
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-gray-100">
                <th
                  onClick={() => handleSort("metric")}
                  className="text-left text-xs font-medium text-gray-500 pb-2 cursor-pointer hover:text-purple-500 select-none"
                >
                  Metric <SortIcon col="metric" />
                </th>
                <th
                  onClick={() => handleSort("value")}
                  className="text-left text-xs font-medium text-gray-500 pb-2 cursor-pointer hover:text-purple-500 select-none"
                >
                  Value <SortIcon col="value" />
                </th>
                <th className="text-left text-xs font-medium text-gray-500 pb-2">
                  Threshold
                </th>
                <th className="text-left text-xs font-medium text-gray-500 pb-2">
                  Severity
                </th>
                <th className="text-left text-xs font-medium text-gray-500 pb-2">
                  Status
                </th>
                <th
                  onClick={() => handleSort("triggered_at")}
                  className="text-left text-xs font-medium text-gray-500 pb-2 cursor-pointer hover:text-purple-500 select-none"
                >
                  Triggered <SortIcon col="triggered_at" />
                </th>
                <th className="text-left text-xs font-medium text-gray-500 pb-2">
                  Resolved
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-50">
              {sorted.map(alert => {
                const severity = getSeverity(alert.value, alert.threshold)
                const metricName = formatMetric(alert.metric)
                // Derive unit for display
                const isPercent = ["cpu_usage", "memory_usage", "disk_usage"].includes(alert.metric)
                const unit = isPercent ? "%" : " KB/s"

                return (
                  <tr key={alert.id} className="hover:bg-gray-50 transition-colors">
                    <td className="py-3 font-medium text-gray-800">
                      {metricName}
                    </td>
                    <td className="py-3 text-gray-700">
                      {alert.value.toFixed(1)}{unit}
                    </td>
                    <td className="py-3 text-gray-500">
                      {alert.threshold.toFixed(1)}{unit}
                    </td>
                    <td className="py-3">
                      <span className={`text-xs font-semibold px-2 py-0.5 rounded-full ${severity.color}`}>
                        {severity.label}
                      </span>
                    </td>
                    <td className="py-3">
                      <span className={`text-xs px-2 py-0.5 rounded-full ${
                        alert.status
                          ? "bg-red-100 text-red-600"
                          : "bg-green-100 text-green-600"
                      }`}>
                        {alert.status ? "Active" : "Resolved"}
                      </span>
                    </td>
                    <td className="py-3 text-gray-500 text-xs">
                      {new Date(alert.triggered_at).toLocaleString()}
                    </td>
                    <td className="py-3 text-gray-500 text-xs">
                      {alert.resolved_at
                        ? new Date(alert.resolved_at).toLocaleString()
                        : <span className="text-gray-300">—</span>
                      }
                    </td>
                  </tr>
                )
              })}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}