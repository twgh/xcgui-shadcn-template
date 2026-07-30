/**
 * 窗口拖动工具 - 通过 JS 调用后端 API 实现窗口拖动.
 * 不依赖 CSS 的 app-region: drag，避免右键弹出系统菜单的问题.
 * 内部使用了 api.moveWindow 来移动窗口, 需在 Go 后端绑定.
 */

class WindowDrag {
  /**
   * 为指定元素启用窗口拖动功能
   * @param element - 要启用拖动的元素或选择器
   * @param exclude - 要排除的元素或选择器（点击时不触发拖动）
   * @returns 取消绑定函数
   */
  static enable(
    element: string | HTMLElement,
    exclude?: string | HTMLElement | Array<string | HTMLElement>
  ): () => void {
    const targetEl = typeof element === 'string'
      ? document.querySelector(element) as HTMLElement
      : element

    if (!targetEl) {
      console.error('[WindowDrag] 目标元素不存在:', element)
      return () => {}
    }

    const moveFn = (x: number, y: number) => {
      window?.api?.moveWindow(Math.round(x), Math.round(y))
    }

    // 处理排除列表：区分 CSS 选择器（动态匹配）和 HTMLElement（静态引用）
    const staticEls: HTMLElement[] = []
    const dynamicSelectors: string[] = []

    if (exclude) {
      const items = Array.isArray(exclude) ? exclude : [exclude]
      for (const item of items) {
        if (typeof item === 'string') {
          // CSS 选择器：使用 matches/closest 动态匹配而不是 querySelector
          dynamicSelectors.push(item)
        } else {
          staticEls.push(item)
        }
      }
    }

    const isExcluded = (el: EventTarget | null): boolean => {
      if (!(el instanceof HTMLElement)) return false
      // 静态元素（预计算的 HTMLElement 引用）
      if (staticEls.some(excludeEl => excludeEl === el || excludeEl.contains(el))) return true
      // 动态 CSS 选择器（在点击时实时匹配，支持选择器返回多个元素）
      if (dynamicSelectors.length > 0) {
        return dynamicSelectors.some(sel => el.matches(sel) || el.closest(sel) !== null)
      }
      return false
    }

    let isDragging = false
    let offsetX = 0
    let offsetY = 0

    // 添加 user-select: none 防止拖动时选中文本
    targetEl.style.userSelect = 'none'

    const onMouseDown = (e: MouseEvent) => {
      if (e.button !== 0) return // 只响应左键
      if (isExcluded(e.target)) return

      isDragging = true
      offsetX = e.clientX
      offsetY = e.clientY
    }

    const onMouseMove = (e: MouseEvent) => {
      if (!isDragging) return

      const newX = e.screenX - offsetX
      const newY = e.screenY - offsetY
      moveFn(newX, newY)
    }

    const onMouseUp = () => {
      isDragging = false
    }

    targetEl.addEventListener('mousedown', onMouseDown)
    document.addEventListener('mousemove', onMouseMove)
    document.addEventListener('mouseup', onMouseUp)

    // 返回取消绑定函数
    return () => {
      targetEl.removeEventListener('mousedown', onMouseDown)
      document.removeEventListener('mousemove', onMouseMove)
      document.removeEventListener('mouseup', onMouseUp)
    }
  }
}

// 挂载到 window 上供全局访问
;(window as unknown as Record<string, unknown>).WindowDrag = WindowDrag

export default WindowDrag
