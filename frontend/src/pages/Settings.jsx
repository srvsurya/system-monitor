import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { Settings as SettingsIcon } from 'lucide-react'
import api from '../api/axios'

export default function Settings() {
  const navigate = useNavigate()
  const [alertEmail, setAlertEmail] = useState('')
  const [cpuThreshold, setCpuThreshold] = useState(80)
  const [memThreshold, setMemThreshold] = useState(80)
  const [duration, setDuration] = useState(1)
  const [saving, setSaving] = useState(false)
  const [saved, setSaved] = useState(false)
  const [error, setError] = useState(null)

  useEffect(() => {
    const fetchSettings = async () => {
      try {
        const [settingsRes, rulesRes] = await Promise.all([
          api.get('/api/v1/user/settings'),
          api.get('/api/v1/alerts/rules'),
        ])
        if (settingsRes.data.alert_email) {
          setAlertEmail(settingsRes.data.alert_email)
        }
        const rules = rulesRes.data || []
        const cpu = rules.find(r => r.metric === 'cpu_usage')
        const mem = rules.find(r => r.metric === 'memory_used')
        if (cpu) {
          setCpuThreshold(cpu.threshold)
          setDuration(Math.round(cpu.duration_seconds / 60))
        }
        if (mem) setMemThreshold(mem.threshold)
      } catch (err) {
        console.error('Failed to fetch settings:', err)
      }
    }
    fetchSettings()
  }, [])

  const handleSave = async () => {
    console.log(alertEmail)
    setSaving(true)
    setError(null)
    try {
      await Promise.all([
        api.patch('/api/v1/user/settings', { alert_email: alertEmail }),
        api.post('/api/v1/alerts/rules', {
          metric: 'cpu_usage',
          operator: '>',
          threshold: cpuThreshold,
          duration_seconds: duration * 60,
        }),
        api.post('/api/v1/alerts/rules', {
          metric: 'memory_used',
          operator: '>',
          threshold: memThreshold,
          duration_seconds: duration * 60,
        }),
      ])
      setSaved(true)
      setTimeout(() => setSaved(false), 3000)
    } catch (err) {
      setError('Failed to save settings')
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="min-h-screen bg-gray-50 p-8">
      <div className="max-w-2xl mx-auto">

        <button onClick={() => navigate('/')} className="text-sm text-gray-500 hover:text-gray-700 mb-6 flex items-center gap-1">
          ← Back to Dashboard
        </button>

        <div className="flex items-center gap-3 mb-8">
          <div className="bg-blue-600 p-2 rounded-lg">
            <SettingsIcon className="w-5 h-5 text-white" />
          </div>
          <div>
            <h1 className="text-xl font-bold text-gray-900">Settings</h1>
            <p className="text-sm text-gray-500">Configure alert thresholds and notifications</p>
          </div>
        </div>

        {/* Email Notifications */}
        <div className="bg-white rounded-xl border border-gray-200 p-6 mb-4">
          <h2 className="font-semibold text-gray-900 mb-4">✉ Email Notifications</h2>
          <label className="text-sm text-gray-600 block mb-1">Alert Email Address</label>
          <input
            type="email"
            value={alertEmail}
            onChange={e => setAlertEmail(e.target.value)}
            placeholder="user@example.com"
            className="w-full border border-gray-300 rounded-lg px-4 py-2 text-sm focus:outline-none focus:border-blue-500 mb-2"
          />
          <p className="text-xs text-gray-400">Critical alerts will be sent to this email address</p>

          <label className="text-sm text-gray-600 block mt-4 mb-1">
            Email Alert Trigger Duration (minutes)
          </label>
          <div className="flex items-center gap-4">
            <input
              type="range"
              min={1}
              max={60}
              value={duration}
              onChange={e => setDuration(Number(e.target.value))}
              className="flex-1"
            />
            <span className="text-sm font-medium w-14 text-right">{duration} min</span>
          </div>
          <p className="text-xs text-gray-400 mt-1">Send email only if alert persists for this duration</p>
        </div>

        {/* Alert Thresholds */}
        <div className="bg-white rounded-xl border border-gray-200 p-6 mb-4">
          <h2 className="font-semibold text-gray-900 mb-6">🔔 Alert Thresholds</h2>

          <div className="mb-6">
            <div className="flex items-center gap-2 mb-3">
              <div className="w-2 h-2 rounded-full bg-blue-500" />
              <span className="text-sm font-medium text-gray-700">CPU Threshold</span>
              <span className="ml-auto text-sm font-semibold text-yellow-600 bg-yellow-50 px-2 py-0.5 rounded">{cpuThreshold}%</span>
            </div>
            <input
              type="range"
              min={1}
              max={100}
              value={cpuThreshold}
              onChange={e => setCpuThreshold(Number(e.target.value))}
              className="w-full"
            />
            <p className="text-xs text-gray-400 mt-1">Alert when CPU usage exceeds {cpuThreshold}%</p>
          </div>

          <div>
            <div className="flex items-center gap-2 mb-3">
              <div className="w-2 h-2 rounded-full bg-purple-500" />
              <span className="text-sm font-medium text-gray-700">Memory Threshold</span>
              <span className="ml-auto text-sm font-semibold text-yellow-600 bg-yellow-50 px-2 py-0.5 rounded">{memThreshold}%</span>
            </div>
            <input
              type="range"
              min={1}
              max={100}
              value={memThreshold}
              onChange={e => setMemThreshold(Number(e.target.value))}
              className="w-full"
            />
            <p className="text-xs text-gray-400 mt-1">Alert when memory usage exceeds {memThreshold}%</p>
          </div>
        </div>

        {/* Smart Heal - UI only */}
        <div className="bg-white rounded-xl border border-gray-200 p-6 mb-8 opacity-60">
          <h2 className="font-semibold text-gray-900 mb-1">✨ Smart Heal Auto-Optimization</h2>
          <p className="text-xs text-gray-400">Coming in Week 4</p>
        </div>

        {error && <p className="text-red-500 text-sm mb-4">{error}</p>}

        <button
          onClick={handleSave}
          disabled={saving}
          className="w-full bg-blue-600 hover:bg-blue-700 disabled:opacity-50 text-white py-3 rounded-xl font-medium transition-colors"
        >
          {saving ? 'Saving...' : saved ? '✓ Saved' : 'Save Settings'}
        </button>

      </div>
    </div>
  )
}