import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { Settings as SettingsIcon } from 'lucide-react'
import api from '../api/axios'
import { useDarkMode } from '../hooks/DarkMode'

export default function Settings() {
  const [theme, setTheme] = useDarkMode();
  const navigate = useNavigate()
  const [alertEmail, setAlertEmail] = useState('')
  const [retention,setRetention] = useState(30)
  const [cpuThreshold, setCpuThreshold] = useState(80)
  const [memThreshold, setMemThreshold] = useState(80)
  const [duration, setDuration] = useState(1)
  const [saving, setSaving] = useState(false)
  const [saved, setSaved] = useState(false)
  const [error, setError] = useState(null)

  const [cleanerCPU, setCleanerCPU] = useState(80)
  const [cleanerMem, setCleanerMem] = useState(500)
  const [cleanerDupes, setCleanerDupes] = useState(3)

  const [ignoreList, setIgnoreList] = useState([])
  const [newIgnore, setNewIgnore] = useState('')

  useEffect(() => {
    const fetchSettings = async () => {
      try {
        const [settingsRes, rulesRes, cleanerRes, ignoreRes] = await Promise.all([
          api.get('/api/v1/user/settings'),
          api.get('/api/v1/alerts/rules'),
          api.get('/api/v1/cleaner/settings'),
          api.get('/api/v1/cleaner/ignore'),
        ])
        if (settingsRes.data.alert_email) {
          setAlertEmail(settingsRes.data.alert_email)
        }
        if (settingsRes.data.retention_days){
          setRetention(settingsRes.data.retention_days)
        }
        const rules = rulesRes.data || []
        const cpu = rules.find(r => r.metric === 'cpu_usage')
        const mem = rules.find(r => r.metric === 'memory_used')
        if (cpu) {
          setCpuThreshold(cpu.threshold)
          setDuration(Math.round(cpu.duration_seconds / 60))
        }
        if (mem) setMemThreshold(mem.threshold)
        
        const c = cleanerRes.data
        if (c){
          setCleanerCPU(c.cpu_threshold)
          setCleanerMem(c.mem_threshold_mb)
          setCleanerDupes(c.duplicate_threshold)
        }

        setIgnoreList(ignoreRes.data || [])
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
        api.patch('/api/v1/user/settings', { alert_email: alertEmail,retention_days: retention}),
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
        api.put('/api/v1/cleaner/settings',{
          cpu_threshold: cleanerCPU,
          mem_threshold_mb: cleanerMem,
          duplicate_threshold: cleanerDupes,
        })
      ])
      setSaved(true)
      setTimeout(() => setSaved(false), 3000)
    } catch (err) {
      setError('Failed to save settings')
    } finally {
      setSaving(false)
    }
  }

  const handleAddIgnore = async () => {
    if (!newIgnore.trim()) return
    try {
      await api.post('/api/v1/cleaner/ignore', { process_name: newIgnore.trim() })
      const res = await api.get('/api/v1/cleaner/ignore')
      setIgnoreList(res.data || [])
      setNewIgnore('')
    } catch (err) {
      console.error('Failed to add to ignore list')
    }
  }

  const handleRemoveIgnore = async (id) => {
    try {
      await api.delete(`/api/v1/cleaner/ignore/${id}`)
      setIgnoreList(prev => prev.filter(p => p.id !== id))
    } catch (err) {
      console.error('Failed to remove from ignore list')
    }
  }

  return (
    <div className="min-h-screen bg-gray-5 dark:bg-gray-900 p-8">
      <div className="max-w-2xl mx-auto">

        <button onClick={() => navigate('/')} className="text-sm text-gray-500 hover:text-gray-700 hover:cursor-pointer mb-6 dark:hover:text-white flex items-center gap-1">
          ← Back to Dashboard
        </button>

        <div className="flex items-center gap-3 mb-8">
          <div className="bg-blue-600 p-2 rounded-lg">
            <SettingsIcon className="w-5 h-5 text-white" />
          </div>
          <div>
            <h1 className="text-xl font-bold text-gray-900 dark:text-white">Settings</h1>
            <p className="text-sm text-gray-500 dark:text-gray-200">Configure alert thresholds and notifications</p>
          </div>
        </div>

        {/* Dark Mode */}

        <div className="bg-white rounded-xl border border-gray-200 dark:bg-gray-800 p-6 mb-4">
          <h2 className="font-semibold text-gray-900 mb-4 dark:text-gray-200">User Preferences</h2>
          <div className="flex flex-col gap-1.5">
            <label className="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
              Interface Theme
            </label>
            <select
              value={theme}
              onChange={(e) => setTheme(e.target.value) 
              }
              className="w-full max-w-40 rounded-lg border border-gray-200 bg-white px-3 py-2 text-sm text-gray-800 shadow-xs outline-hidden transition-all focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 dark:border-gray-700 dark:bg-gray-800 dark:text-gray-200 dark:focus:border-blue-500 mt-2"
            >
              <option value="light" className="bg-white text-gray-800 dark:bg-gray-800 dark:text-gray-200">
                ☀️ Light Mode
              </option>
              <option value="dark" className="bg-white text-gray-800 dark:bg-gray-800 dark:text-gray-200">
                🌙 Dark Mode
              </option>
            </select>
            <label className="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider mt-2">
              Data Retention Policy
            </label>
            <input
            type="number"
            value={retention}
            onChange={e => setRetention(Number(e.target.value))}
            className="w-full border border-gray-300 dark:text-gray-200 rounded-lg px-4 py-2 text-sm focus:outline-none focus:border-blue-500 mb-2"
          />
          </div>
          
        </div>

        {/* Email Notifications */}
        <div className="bg-white rounded-xl border border-gray-200 dark:bg-gray-800 p-6 mb-4">
          <h2 className="font-semibold text-gray-900 mb-4 dark:text-gray-200">✉ Email Notifications</h2>
          <label className="text-sm text-gray-600 dark:text-gray-200 block mb-1">Alert Email Address</label>
          <input
            type="email"
            value={alertEmail}
            onChange={e => setAlertEmail(e.target.value)}
            placeholder="user@example.com"
            className="w-full border border-gray-300 dark:text-gray-200 rounded-lg px-4 py-2 text-sm focus:outline-none focus:border-blue-500 mb-2"
          />
          <p className="text-xs text-gray-400">Critical alerts will be sent to this email address</p>

          <label className="text-sm text-gray-600 block mt-4 mb-1 dark:text-gray-200">
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
            <span className="text-sm font-medium w-14 text-right dark:text-gray-200">{duration} min</span>
          </div>
          <p className="text-xs text-gray-400 mt-1 dark:text-gray-400">Send email only if alert persists for this duration</p>
        </div>

        {/* Alert Thresholds */}
        <div className="bg-white rounded-xl border border-gray-200 p-6 mb-4 dark:bg-gray-800">
          <h2 className="font-semibold text-gray-900 mb-6 dark:text-gray-200">🔔 Alert Thresholds</h2>

          <div className="mb-6">
            <div className="flex items-center gap-2 mb-3">
              <div className="w-2 h-2 rounded-full bg-blue-500" />
              <span className="text-sm font-medium text-gray-700 dark:text-gray-200">CPU Threshold</span>
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
              <span className="text-sm font-medium text-gray-700 dark:text-gray-200">Memory Threshold</span>
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
        <div className="bg-white rounded-xl border border-gray-200 p-6 mb-4 dark:bg-gray-800">
          <h2 className="font-semibold text-gray-900 mb-1 dark:text-gray-200">⚙️ Advanced — Optimizer Settings</h2>
          <p className="text-xs text-gray-400 mb-6">Controls what the Optimize button targets when run from the dashboard</p>

          <div className="mb-6">
            <div className="flex items-center gap-2 mb-3">
              <div className="w-2 h-2 rounded-full bg-red-500" />
              <span className="text-sm font-medium text-gray-700 dark:text-gray-200">CPU Hog Threshold</span>
              <span className="ml-auto text-sm font-semibold text-red-600 bg-red-50 px-2 py-0.5 rounded">{cleanerCPU}%</span>
            </div>
            <input type="range" min={50} max={100} value={cleanerCPU} onChange={e => setCleanerCPU(Number(e.target.value))} className="w-full" />
            <p className="text-xs text-gray-400 mt-1">Kill processes consuming more than {cleanerCPU}% CPU</p>
          </div>

          <div className="mb-6">
            <div className="flex items-center gap-2 mb-3">
              <div className="w-2 h-2 rounded-full bg-orange-500" />
              <span className="text-sm font-medium text-gray-700 dark:text-gray-200">Memory Hog Threshold</span>
              <span className="ml-auto text-sm font-semibold text-orange-600 bg-orange-50 px-2 py-0.5 rounded">{cleanerMem} MB</span>
            </div>
            <input type="range" min={100} max={2000} step={50} value={cleanerMem} onChange={e => setCleanerMem(Number(e.target.value))} className="w-full" />
            <p className="text-xs text-gray-400 mt-1">Kill processes using more than {cleanerMem} MB of memory</p>
          </div>

          <div className="mb-2">
            <div className="flex items-center gap-2 mb-3">
              <div className="w-2 h-2 rounded-full bg-yellow-500" />
              <span className="text-sm font-medium text-gray-700 dark:text-gray-200">Duplicate Process Threshold</span>
              <span className="ml-auto text-sm font-semibold text-yellow-600 bg-yellow-50 px-2 py-0.5 rounded">{cleanerDupes} instances</span>
            </div>
            <input type="range" min={2} max={10} value={cleanerDupes} onChange={e => setCleanerDupes(Number(e.target.value))} className="w-full" />
            <p className="text-xs text-gray-400 mt-1">Flag processes with more than {cleanerDupes} running instances</p>
          </div>

          <div className="mt-4 p-3 bg-yellow-50 border border-yellow-200 rounded-lg">
            <p className="text-xs text-yellow-700">⚠️ Lowering the duplicate threshold may kill legitimate processes. Browsers and system services commonly run multiple instances.</p>
          </div>
        </div>

        {/* Ignore List */}
        <div className="bg-white rounded-xl border border-gray-200 p-6 mb-8 dark:bg-gray-800">
          <h2 className="font-semibold text-gray-900 mb-1 dark:text-gray-200">🛡️ Optimizer Ignore List</h2>
          <p className="text-xs text-gray-400 mb-4">Processes on this list will never be touched by the Optimizer</p>

          <div className="flex gap-2 mb-4">
            <input
              type="text"
              value={newIgnore}
              onChange={e => setNewIgnore(e.target.value)}
              onKeyDown={e => e.key === 'Enter' && handleAddIgnore()}
              placeholder="e.g. code, node, chrome"
              className="flex-1 border border-gray-300 dark:text-gray-200 rounded-lg px-4 py-2 text-sm focus:outline-none focus:border-blue-500"
            />
            <button
              onClick={handleAddIgnore}
              className="bg-black text-white text-sm px-4 py-2 rounded-lg hover:opacity-80 transition-opacity hover:cursor-pointer"
            >
              Add
            </button>
          </div>

          {ignoreList.length === 0 ? (
            <p className="text-xs text-gray-400">No processes ignored yet</p>
          ) : (
            <div className="space-y-2">
              {ignoreList.map(p => (
                <div key={p.id} className="flex items-center justify-between px-3 py-2 bg-gray-50 rounded-lg border border-gray-100 dark:bg-gray-500">
                  <span className="text-sm text-gray-700 dark:text-gray-100">{p.process_name}</span>
                  <button
                    onClick={() => handleRemoveIgnore(p.id)}
                    className="text-xs text-red-500 hover:text-red-700 hover:cursor-pointer"
                  >
                    Remove
                  </button>
                </div>
              ))}
            </div>
          )}
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