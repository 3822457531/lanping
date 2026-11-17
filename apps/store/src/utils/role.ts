import { useSession } from '@/stores/session'

export function isManager() {
  return useSession().user?.role === 'manager'
}

export function isStaff() {
  return useSession().user?.role === 'staff'
}
