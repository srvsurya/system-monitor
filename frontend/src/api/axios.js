import axios from 'axios'

const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL || '/', // for dev - 8080. for prod (if it exists), a prod env variable that has my ec2 instance public ipv4 address
})

api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

export default api