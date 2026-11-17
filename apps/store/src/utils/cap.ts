import { Capacitor } from '@capacitor/core'

export function isNativeApp() {
  return Capacitor.isNativePlatform()
}

/** 浏览器：非 localhost 的 http 不是安全上下文，Geolocation 会被直接拒绝 */
export function isInsecureHttpOrigin() {
  if (typeof location === 'undefined') return false
  if (location.protocol !== 'http:') return false
  const host = location.hostname
  return host !== 'localhost' && host !== '127.0.0.1'
}

function mapGeoError(raw: unknown): string {
  const msg = String((raw as any)?.message || raw || '')
  const lower = msg.toLowerCase()
  if (lower.includes('denied') || lower.includes('permission') || lower.includes('未授权')) {
    if (isInsecureHttpOrigin() && !isNativeApp()) {
      return '浏览器禁止：请用 http://localhost:5174 调试定位，或安装 App（局域网 IP 的 HTTP 无法定位）'
    }
    return '未授权定位，请在系统设置中允许位置权限（建议选「精确位置」）'
  }
  if (
    lower.includes('obtain location in time') ||
    lower.includes('timeout') ||
    lower.includes('timed out') ||
    lower.includes('超时')
  ) {
    return '定位超时：请打开手机「位置信息」，稍后点「重新定位」'
  }
  if (lower.includes('unavailable') || lower.includes('location services') || lower.includes('disabled')) {
    return '定位服务不可用，请打开手机「位置信息/GPS」'
  }
  return msg || '定位失败，请重试'
}

function withTimeout<T>(promise: Promise<T>, ms: number, message: string): Promise<T> {
  let timer: ReturnType<typeof setTimeout> | undefined
  const timeout = new Promise<T>((_, reject) => {
    timer = setTimeout(() => reject(new Error(message)), ms)
  })
  return Promise.race([promise, timeout]).finally(() => {
    if (timer) clearTimeout(timer)
  }) as Promise<T>
}

function readBrowserGeo(options: {
  enableHighAccuracy: boolean
  timeout: number
  maximumAge: number
}): Promise<{ lat: number; lng: number; accuracy?: number }> {
  return new Promise((resolve, reject) => {
    if (!navigator.geolocation) {
      reject(new Error('当前环境不支持定位'))
      return
    }
    navigator.geolocation.getCurrentPosition(
      (pos) => {
        resolve({
          lat: pos.coords.latitude,
          lng: pos.coords.longitude,
          accuracy: pos.coords.accuracy,
        })
      },
      (err) => {
        let msg = '定位失败，请允许位置权限后重试'
        if (err.code === 1) msg = '未授权定位，请在系统/浏览器中允许位置权限'
        if (err.code === 2) msg = '暂时获取不到位置，请到空旷处或检查 GPS'
        if (err.code === 3) msg = '定位超时，请重试'
        reject(new Error(msg))
      },
      {
        enableHighAccuracy: options.enableHighAccuracy,
        timeout: options.timeout,
        maximumAge: options.maximumAge,
      },
    )
  })
}

/**
 * App 内定位策略：
 * - Capacitor Geolocation 插件依赖 Google Play 服务，国内手机常会超时
 * - App WebView 源是 http://localhost（安全上下文），可用 navigator.geolocation
 * - Capacitor Bridge 会在 WebView 申请定位时自动弹出系统权限框
 */
export async function getGeoPosition(options?: {
  enableHighAccuracy?: boolean
  timeout?: number
  maximumAge?: number
}): Promise<{ lat: number; lng: number; accuracy?: number }> {
  const enableHighAccuracy = options?.enableHighAccuracy ?? false
  const timeout = options?.timeout ?? 10000
  const maximumAge = options?.maximumAge ?? 60000

  // H5 用局域网 IP 的 http：浏览器安全策略直接禁定位
  if (!isNativeApp() && isInsecureHttpOrigin()) {
    throw new Error(
      '浏览器禁止：http://192.168.x.x 无法定位。请用 http://localhost:5174 调试，或安装 App',
    )
  }

  try {
    // App / localhost H5：都走 WebView/浏览器定位（与 localhost 秒出一致）
    return await withTimeout(
      readBrowserGeo({
        // App 内先不要高精度，避免干等 GPS
        enableHighAccuracy: isNativeApp() ? false : enableHighAccuracy,
        timeout,
        maximumAge,
      }),
      timeout + 2000,
      '定位超时：请打开手机「位置信息」，稍后点「重新定位」',
    )
  } catch (e) {
    // 仅 App：WebView 失败时再试一次原生插件（有 GMS 的机器可能可用）
    if (isNativeApp()) {
      try {
        const { Geolocation } = await import('@capacitor/geolocation')
        const pos = await withTimeout(
          Geolocation.getCurrentPosition({
            enableHighAccuracy: false,
            timeout: 8000,
            maximumAge,
          }),
          10000,
          '定位超时',
        )
        return {
          lat: pos.coords.latitude,
          lng: pos.coords.longitude,
          accuracy: pos.coords.accuracy,
        }
      } catch {
        // fall through
      }
    }
    throw new Error(mapGeoError(e))
  }
}
