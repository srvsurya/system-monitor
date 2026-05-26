export default function MetricCard({ title, value, unit, color = 'blue' }) {
  const colors = {
    blue: 'border-blue-500 text-blue-400',
    green: 'border-green-500 text-green-400',
    yellow: 'border-yellow-500 text-yellow-400',
    red: 'border-red-500 text-red-400',
  }

  return (
    <div className={`bg-gray-900 border-l-4 ${colors[color]} rounded-xl p-6`}>
      <p className="text-gray-400 text-sm mb-1">{title}</p>
      <p className="text-white text-3xl font-bold">
        {value ?? '—'}
        <span className="text-gray-400 text-base ml-1">{unit}</span>
      </p>
    </div>
  )
}