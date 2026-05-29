import { useEffect, useState } from 'react'
import { AlertCircle } from 'lucide-react'
import api from '../api/axios'

export default function ActiveAlerts() {
  const [alerts, setAlerts] = useState([])

  useEffect(() => {
    const fetch = async () => {
      try {
        const res = await api.get('/api/v1/alerts')
        setAlerts(res.data || [])
      } catch (err) {
        console.error('Failed to fetch alerts:', err)
      }
    }
    fetch()
    const interval = setInterval(fetch, 10000)
    return () => clearInterval(interval)
  }, [])

  return (
    <div className="bg-white rounded-xl border border-gray-200 p-6 mt-6">
      <h2 className="font-semibold text-gray-900 mb-4">System Alerts</h2>
      {alerts.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-8 text-gray-400">
          <AlertCircle className="w-8 h-8 mb-2" strokeWidth={1.5} />
          <p className="text-sm">No active alerts</p>
        </div>
      ) : (
        <div className="space-y-3">
          {alerts.map(alert => (
            <div
              key={alert.id}
              className="flex items-center justify-between p-3 bg-red-50 border border-red-100 rounded-lg"
            >
              <div className="flex items-center gap-3">
                <AlertCircle className="w-4 h-4 text-red-500 shrink-0" />
                <div>
                  <p className="text-sm font-medium text-gray-900">
                    {alert.metric} exceeded threshold
                  </p>
                  <p className="text-xs text-gray-500">
                    Value: {alert.value?.toFixed(1)} • Threshold: {alert.threshold?.toFixed(1)} • Time: {alert.triggered_at.split('T')[1].slice(0,8)}
                  </p>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}