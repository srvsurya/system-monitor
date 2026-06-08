import { useState, useEffect } from "react"
import { useNavigate } from "react-router-dom"
import axios from "../api/axios"

export default function Insights() {
  const navigate = useNavigate()

  // --- Raw data from API ---
  const [metrics, setMetrics] = useState([])        // /stats/history
  const [alertHistory, setAlertHistory] = useState([]) // /alerts/history
  const [rules, setRules] = useState([])            // /alerts/rules
  const [processes, setProcesses] = useState([])    // /processes

  // --- Derived/computed ---
  const [ruleMap, setRuleMap] = useState({})         // { rule_id: { metric, threshold } }
  const [thresholds, setThresholds] = useState({})   // { cpu_usage: 80, memory_usage: 75 }

  // --- UI state ---
  const [loading, setLoading] = useState(true)
  const [timeRange, setTimeRange] = useState("24h")  // "24h" | "7d" | "30d" | "all"
  const [consumerTab, setConsumerTab] = useState("cpu") // "cpu" | "memory"
  const [insights, setInsights] = useState(null)     // null until generated
  const [generating, setGenerating] = useState(false)

  // --- Fetch all data on mount ---
  useEffect(() => {
    fetchAll()
  }, [])

  const fetchAll = async () => {
    setLoading(true)
    try {
      const [metricsRes, alertsRes, rulesRes, processesRes] = await Promise.all([
        axios.get("/api/v1/stats/history?limit=500"),
        axios.get("/api/v1/alerts/history"),
        axios.get("/api/v1/alerts/rules"),
        axios.get("/api/v1/processes"),
      ])

      setMetrics(metricsRes.data || [])
      setAlertHistory(alertsRes.data || [])
      setProcesses(processesRes.data || [])

      // Build ruleMap: { rule_id → { metric, threshold } }
      const map = {}
      const thresh = {}
      ;(rulesRes.data || []).forEach(rule => {
        map[rule.id] = { metric: rule.metric, threshold: rule.threshold }
        thresh[rule.metric] = rule.threshold
      })
      setRules(rulesRes.data || [])
      setRuleMap(map)
      setThresholds(thresh)

    } catch (err) {
      console.error("Failed to fetch insights data:", err)
    } finally {
      setLoading(false)
    }
  }

  // --- Time range filter helper ---
  // Call this wherever you need to filter alertHistory or metrics by timeRange
  const filterByTimeRange = (items, dateField) => {
    if (timeRange === "all") return items
    const now = new Date()
    const cutoff = {
      "24h": new Date(now - 24 * 60 * 60 * 1000),
      "7d":  new Date(now - 7  * 24 * 60 * 60 * 1000),
      "30d": new Date(now - 30 * 24 * 60 * 60 * 1000),
    }[timeRange]
    return items.filter(item => new Date(item[dateField]) >= cutoff)
  }

  // --- Health summary algorithm --- 
  const generateInsights = () => {
  setGenerating(true)
  setTimeout(() => {
    const filtered = filterByTimeRange(metrics, "timestamp")
    const filteredAlerts = filterByTimeRange(alertHistory, "triggered_at")

    const rangeLabel = {
      "24h": "the last 24 hours",
      "7d":  "the last 7 days",
      "30d": "the last 30 days",
      "all": "all recorded time",
    }[timeRange]

    // --- CPU computations ---
    const cpuValues = filtered.map(m => m.cpu_usage).filter(v => v != null)
    const avgCPU = cpuValues.length
      ? cpuValues.reduce((a, b) => a + b, 0) / cpuValues.length
      : null
    const peakCPU = cpuValues.length ? Math.max(...cpuValues) : null
    const cpuThreshold = thresholds["cpu_usage"]

    // --- Memory computations ---
    const memValues = filtered
      .filter(m => m.memory_total > 0)
      .map(m => (m.memory_used / m.memory_total) * 100)
    const avgMem = memValues.length
      ? memValues.reduce((a, b) => a + b, 0) / memValues.length
      : null
    const peakMem = memValues.length ? Math.max(...memValues) : null
    const memThreshold = thresholds["memory_usage"]

    // --- Alert computations ---
    const cpuAlerts = filteredAlerts.filter(
      a => ruleMap[a.rule_id]?.metric === "cpu_usage"
    )
    const memAlerts = filteredAlerts.filter(
      a => ruleMap[a.rule_id]?.metric === "memory_usage"
    )
    const totalAlerts = filteredAlerts.length

    // --- Alert clustering: check if >50% of alerts fall within any 2-hour window ---
    let clusterNote = null
    if (filteredAlerts.length >= 4) {
      const timestamps = filteredAlerts
        .map(a => new Date(a.triggered_at).getTime())
        .sort((a, b) => a - b)
      const windowMs = 2 * 60 * 60 * 1000
      let maxInWindow = 1
      let windowPeak = null
      for (let i = 0; i < timestamps.length; i++) {
        const inWindow = timestamps.filter(
          t => t >= timestamps[i] && t <= timestamps[i] + windowMs
        )
        if (inWindow.length > maxInWindow) {
          maxInWindow = inWindow.length
          windowPeak = new Date(timestamps[i])
        }
      }
      if (maxInWindow >= Math.ceil(filteredAlerts.length * 0.5)) {
        const hour = windowPeak.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" })
        clusterNote = `Notably, ${maxInWindow} of those alerts clustered around ${hour}, which could point to a scheduled job, a deployment, or a recurring load pattern worth investigating.`
      }
    }

    // --- Sentence 1: CPU ---
    let cpuSentence = ""
    if (avgCPU === null) {
      cpuSentence = "There isn't enough metric data to assess CPU behavior over this period."
    } else if (!cpuThreshold) {
      cpuSentence = `CPU averaged ${avgCPU.toFixed(1)}% over ${rangeLabel}, peaking at ${peakCPU.toFixed(1)}%, but no alert rule has been configured for it — so it's hard to say whether this is concerning or not.`
    } else if (avgCPU >= cpuThreshold) {
      if (cpuAlerts.length > 0) {
        cpuSentence = `CPU has been under real stress over ${rangeLabel} — it averaged ${avgCPU.toFixed(1)}% against a threshold of ${cpuThreshold}%, peaked at ${peakCPU.toFixed(1)}%, and triggered ${cpuAlerts.length} alert${cpuAlerts.length > 1 ? "s" : ""} in that window.`
      } else {
        cpuSentence = `CPU averaged ${avgCPU.toFixed(1)}% over ${rangeLabel} — technically above its configured threshold of ${cpuThreshold}% — though no alerts fired, which may mean breaches were brief or the duration window wasn't met.`
      }
    } else if (avgCPU >= cpuThreshold * 0.75) {
      cpuSentence = `CPU has been running moderately over ${rangeLabel}, averaging ${avgCPU.toFixed(1)}% with a peak of ${peakCPU.toFixed(1)}% — comfortably under the ${cpuThreshold}% threshold${cpuAlerts.length === 0 ? ", and no alerts were triggered" : ""}.`
    } else {
      cpuSentence = `CPU looks healthy over ${rangeLabel}, holding a steady average of ${avgCPU.toFixed(1)}% well below the configured ${cpuThreshold}% threshold${peakCPU < cpuThreshold ? " — even at its peak" : ""}.`
    }

    // --- Sentence 2: Memory ---
    let memSentence = ""
    if (avgMem === null) {
      memSentence = "Memory data is unavailable for this period."
    } else if (!memThreshold) {
      memSentence = `Memory usage averaged ${avgMem.toFixed(1)}% over ${rangeLabel}, peaking at ${peakMem.toFixed(1)}%, but no memory alert rule is set up.`
    } else if (avgMem >= memThreshold) {
      if (memAlerts.length > 0) {
        memSentence = `Memory has been consistently high — averaging ${avgMem.toFixed(1)}% over ${rangeLabel} with a peak of ${peakMem.toFixed(1)}%, and firing ${memAlerts.length} alert${memAlerts.length > 1 ? "s" : ""}. This level of sustained pressure could impact application stability.`
      } else {
        memSentence = `Memory averaged ${avgMem.toFixed(1)}% over ${rangeLabel}, edging above its ${memThreshold}% threshold on average, though the alert duration window wasn't sustained long enough to trigger a formal alert.`
      }
    } else if (avgMem >= memThreshold * 0.75) {
      memSentence = `Memory is within limits but has been climbing — averaging ${avgMem.toFixed(1)}% over ${rangeLabel} and peaking at ${peakMem.toFixed(1)}%. Still under the ${memThreshold}% threshold, but worth watching if the trend continues.`
    } else {
      memSentence = `Memory has been in good shape over ${rangeLabel}, averaging ${avgMem.toFixed(1)}% with no signs of pressure.`
    }

    // --- Sentence 3: Alert pattern ---
    let alertSentence = ""
    if (totalAlerts === 0) {
      alertSentence = `No alerts were triggered over ${rangeLabel} — a good sign overall.`
    } else if (cpuAlerts.length > 0 && memAlerts.length > 0) {
      alertSentence = `Across ${rangeLabel}, the system generated ${totalAlerts} alert${totalAlerts > 1 ? "s" : ""} in total — ${cpuAlerts.length} CPU and ${memAlerts.length} memory. When both metrics are alerting together, it often indicates the system is genuinely resource-constrained rather than a single runaway process.`
    } else if (cpuAlerts.length > memAlerts.length) {
      alertSentence = `Of the ${totalAlerts} alert${totalAlerts > 1 ? "s" : ""} over ${rangeLabel}, the majority were CPU-related (${cpuAlerts.length}), suggesting compute demand has been the primary pressure point.`
    } else if (memAlerts.length > 0) {
      alertSentence = `Of the ${totalAlerts} alert${totalAlerts > 1 ? "s" : ""} over ${rangeLabel}, memory was the dominant concern (${memAlerts.length} alerts), which can sometimes indicate a slow leak or a process that isn't releasing memory properly.`
    } else {
      alertSentence = `${totalAlerts} alert${totalAlerts > 1 ? "s were" : " was"} triggered over ${rangeLabel}.`
    }

    // --- Sentence 4: Cross-signal or cluster note ---
    let crossSentence = clusterNote || (() => {
      if (avgCPU !== null && avgMem !== null) {
        const bothHigh = avgCPU >= (cpuThreshold || 70) * 0.75 && avgMem >= (memThreshold || 70) * 0.75
        const bothLow  = avgCPU < (cpuThreshold || 70) * 0.5  && avgMem < (memThreshold || 70) * 0.5
        if (bothHigh) return `Both CPU and memory have been elevated over this period — if this is a pattern rather than a one-off, it may be worth reviewing what's running regularly on this machine.`
        if (bothLow)  return `Both CPU and memory have been comfortably low across the board — the system appears to have had plenty of headroom throughout this period.`
        return `CPU and memory pressures appear to be independent of each other over this period, which usually means there's no single systemic issue — just isolated spikes.`
      }
      return null
    })()

    // --- Sentence 5: Overall verdict ---
    let verdict = ""
    const stressScore =
      (avgCPU !== null && cpuThreshold && avgCPU >= cpuThreshold ? 2 : 0) +
      (avgMem !== null && memThreshold && avgMem >= memThreshold ? 2 : 0) +
      (totalAlerts > 5 ? 2 : totalAlerts > 0 ? 1 : 0)

    if (stressScore === 0) {
      verdict = `Overall, the system looks healthy over ${rangeLabel}. No action needed.`
    } else if (stressScore <= 2) {
      verdict = `Overall, the system has been mostly stable over ${rangeLabel} with minor stress in places. Nothing urgent, but keeping an eye on the trend wouldn't hurt.`
    } else if (stressScore <= 4) {
      verdict = `Overall, there are some signs of strain over ${rangeLabel}. It's worth reviewing which processes are consuming the most resources and whether any thresholds need adjusting.`
    } else {
      verdict = `Overall, the system has been under significant pressure over ${rangeLabel}. This warrants a closer look — check the resource consumers below and consider whether current workloads are sustainable.`
    }

    const sentences = [cpuSentence, memSentence, alertSentence, crossSentence, verdict].filter(Boolean)
    setInsights(sentences)
    setGenerating(false)
  }, 900)
}

  if (loading) return (
    <div className="min-h-screen flex items-center justify-center">
      <p className="text-gray-500 dark:text-gray-400">Loading insights...</p>
    </div>
  )

  return (
    <div className="min-h-screen bg-gray-50 p-6">
      {/* Header */}
      <div className="max-w-5xl mx-auto">
        <button
          onClick={() => navigate("/")}
          className="text-sm text-gray-500 hover:text-gray-700 mb-4 flex items-center gap-1 hover:cursor-pointer"
        >
          ← Back to Dashboard
        </button>
        <div className="flex flex-col items-center justify-between mb-6">
          <div className="flex flex-col items-center gap-3 mb-2">
            <h1 className="text-2xl font-bold text-gray-900">Historical Insights</h1>
            <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">
              Analyze system behavior and resource consumption patterns
            </p>
          </div>
          {/* Health Summary Card */}
        <div className="bg-white rounded-xl border border-gray-200 p-6 mb-6">
        <div className="flex items-center gap-6 justify-between mb-4">
            <div className="flex items-center gap-2">
            <span className="text-green-500 text-lg">✦</span>
            <h2 className="text-base font-semibold text-gray-900 dark:text-white">System Health Summary</h2>
            </div>
            <button
            onClick={generateInsights}
            disabled={generating}
            className="flex items-center gap-2 bg-purple-600 hover:bg-purple-700 hover:cursor-pointer text-white text-sm px-4 py-2 rounded-lg disabled:opacity-60 transition-colors"
            >
            {generating ? (
                <>
                <svg className="animate-spin h-4 w-4" viewBox="0 0 24 24" fill="none">
                    <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"/>
                    <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v8z"/>
                </svg>
                Analyzing...
                </>
            ) : (
                <>⚙ {insights ? "Regenerate Insights" : "Generate System Insights"}</>
            )}
            </button>
        </div>

        {/* Not yet generated */}
        {!insights && !generating && (
            <div className="flex flex-col items-center justify-center py-10 text-center">
            <span className="text-4xl mb-3">🔍</span>
            <p className="text-sm text-gray-500 dark:text-gray-400 max-w-sm">
                Click <span className="font-medium text-purple-500">Generate System Insights</span> to run
                an analysis of your system's health based on metrics and alert history.
            </p>
            </div>
        )}

        {/* Generating state */}
        {generating && (
            <div className="flex flex-col items-center justify-center py-10 text-center">
            <span className="text-4xl mb-3 animate-pulse">⚙</span>
            <p className="text-sm text-gray-400">Analyzing metrics and alert patterns...</p>
            </div>
        )}

        {/* Insights rendered */}
        {insights && !generating && (
            <div className="space-y-3">
            {/* Stress score badge */}
            {(() => {
                const cpuValues = filterByTimeRange(metrics, "timestamp").map(m => m.cpu_usage)
                const avgCPU = cpuValues.length
                ? cpuValues.reduce((a, b) => a + b, 0) / cpuValues.length
                : null
                const memValues = filterByTimeRange(metrics, "timestamp")
                .filter(m => m.memory_total > 0)
                .map(m => (m.memory_used / m.memory_total) * 100)
                const avgMem = memValues.length
                ? memValues.reduce((a, b) => a + b, 0) / memValues.length
                : null
                const totalAlerts = filterByTimeRange(alertHistory, "triggered_at").length

                const stressScore =
                (avgCPU !== null && thresholds["cpu_usage"] && avgCPU >= thresholds["cpu_usage"] ? 2 : 0) +
                (avgMem !== null && thresholds["memory_usage"] && avgMem >= thresholds["memory_usage"] ? 2 : 0) +
                (totalAlerts > 5 ? 2 : totalAlerts > 0 ? 1 : 0)

                const badge = stressScore === 0
                ? { label: "Healthy",  color: "bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400" }
                : stressScore <= 2
                ? { label: "Stable",   color: "bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400" }
                : stressScore <= 4
                ? { label: "Moderate", color: "bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400" }
                : { label: "Stressed", color: "bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400" }

                return (
                <div className="flex items-center gap-2 mb-2">
                    <span className={`text-xs font-semibold px-3 py-1 rounded-full ${badge.color}`}>
                    {badge.label}
                    </span>
                    <span className="text-xs text-gray-400 dark:text-gray-500">
                    Based on {timeRange === "all" ? "all recorded data" : `the last ${timeRange}`}
                    </span>
                </div>
                )
            })()}

            {/* Insight sentences */}
            <div className="border-l-2 border-purple-300 dark:border-purple-700 pl-4 space-y-2">
                {insights.map((sentence, i) => (
                <p key={i} className="text-sm text-gray-700 leading-relaxed">
                    {sentence}
                </p>
                ))}
            </div>

            {/* Timestamp */}
            <p className="text-xs text-gray-400 dark:text-gray-500 pt-2">
                Generated at {new Date().toLocaleTimeString()}
            </p>
            </div>
        )}
        </div>

        {/* Time range selector — shared across Alert History and System Trend */}
        <div className="flex items-center gap-2 mb-4">
            <span className="text-xs text-gray-500">Time range:</span>
        {["24h", "7d", "30d", "all"].map(range => (
            <button
            key={range}
            onClick={() => {
                setTimeRange(range)
                setInsights(null) // reset insights when time range changes
            }}
            className={`text-xs px-3 py-1 rounded-full border transition-colors ${
                timeRange === range
                ? "bg-purple-600 text-white border-purple-600"
                : "text-gray-500 dark:text-gray-400 border-gray-300 hover:border-purple-400 hover:cursor-pointer"
            }`}
            >
            {range}
            </button>
        ))}
        </div>

        {/* Placeholder for next sections */}
        <p className="text-gray-400">Alert History + Resource Consumers coming next...</p>
            </div>

                {/* Sections will go here */}
                <p className="text-gray-400">Sections coming next...</p>
            </div>
        </div>
  )
}