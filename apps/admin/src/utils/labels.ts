export const roleLabel: Record<string, string> = {
  admin: '超级管理员',
  manager: '店长',
  staff: '普通员工',
}

export const auditStatusLabel: Record<string, string> = {
  pending: '待审核',
  approved: '已通过',
  rejected: '已拒绝',
  待审核: '待审核',
  已通过: '已通过',
  已拒绝: '已拒绝',
}

export function labelRole(v: string) {
  return roleLabel[v] || v
}

export function labelAudit(v: string) {
  return auditStatusLabel[v] || v
}

export function isPendingAudit(v: string) {
  return v === 'pending' || v === '待审核'
}
