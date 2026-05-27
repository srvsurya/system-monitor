export default function MetricCard({ title,value,unit,icon}) {
  const getColor = (percent) => {
    if (percent < 50) return 'text-green-500'
    if (percent < 80) return 'text-yellow-500'
    return 'text-red-500'
  }


  return (
    <div className="bg-white rounded-lg p-6 shadow-sm border border-gray-200">
      <div className="flex items-start justify-between mb-4">
        <div>
          <p className="text-sm text-gray-500">{title}</p>
        </div>
        <div className={getColor(value)}>
          {icon}
        </div>
      </div>
      <div className="mb-3">
        <div className="flex justify-between text-sm mb-1">
          <span className="text-gray-600">Usage</span>
          <span className={getColor(value)}>{value}{unit}</span>
        </div>
        <div className="w-full bg-gray-200 rounded-full h-2">
          <div
            className={`h-2 rounded-full transition-all ${
              value < 50 ? 'bg-green-500' : value < 80 ? 'bg-yellow-500' : 'bg-red-500'
            }`}
            style={{ width: `${value}%` }}
          />
        </div>
      </div>
    </div>
  )
}