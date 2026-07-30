// Bridge 通信层：前端与 Go 后端之间的 JS Bridge

declare global {
  interface Window {
    api: {
      // 窗口控制
      minimizeWindow: () => Promise<void>
      closeWindow: () => Promise<void>
      moveWindow: (x: number, y: number) => Promise<void>
      // 系统
      getVersion: () => Promise<string>
      frontendReady: () => Promise<void>
    }
  }
}

export function isBridgeAvailable(): boolean {
  return typeof window.api !== "undefined"
}

async function callApi<T>(method: string, ...args: unknown[]): Promise<T | null> {
  if (!isBridgeAvailable()) {
    console.warn(`[Bridge] api."${method}" not available (dev mode)`)
    return null
  }
  try {
    const fn = (window.api as unknown as Record<string, (...args: unknown[]) => Promise<T>>)[method]
    if (!fn) {
      console.warn(`[Bridge] api."${method}" not found`)
      return null
    }
    return await fn(...args)
  } catch (err) {
    console.error(`[Bridge] Error calling api."${method}":`, err)
    return null
  }
}

export const bridge = {
  minimizeWindow: () => callApi<void>("minimizeWindow"),
  closeWindow: () => callApi<void>("closeWindow"),
  moveWindow: (x: number, y: number) => callApi<void>("moveWindow", x, y),
  getVersion: () => callApi<string>("getVersion"),
  frontendReady: () => callApi<void>("frontendReady"),
}
