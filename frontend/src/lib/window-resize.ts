/**
 * 窗口边框拖动调整大小工具 - 通过 JS 调用后端 API 调整窗口大小.
 * 需在 Go 后端绑定: api.getWindowRect / api.setWindowSize / api.moveWindow
 */

interface Rect {
  Left: number
  Top: number
  Right: number
  Bottom: number
}

class WindowResize {
  /**
   * 为所有 .resize-edge 热区启用边框拖动调整窗口大小
   * @param minSize - 窗口最小尺寸
   * @returns 取消绑定函数
   */
  static enable(minSize = 400): () => void {
    let isResizing = false
    let resizeEdge: string | undefined
    let resizeCorner: string | undefined
    let startX = 0
    let startY = 0
    let startWidth = 0
    let startHeight = 0
    let startLeft = 0
    let startTop = 0

    const edges = document.querySelectorAll<HTMLElement>('.resize-edge')

    const onMouseDown = async (e: MouseEvent) => {
      if (e.button !== 0) return // 只响应左键
      const edge = e.currentTarget as HTMLElement
      e.preventDefault() // 防止选中文本

      isResizing = true
      resizeEdge = edge.dataset.edge
      resizeCorner = edge.dataset.corner
      startX = e.screenX
      startY = e.screenY

      // 获取窗口矩形
      const rc = (await window?.api?.getWindowRect()) as Rect
      if (!rc) return
      startLeft = rc.Left
      startTop = rc.Top
      startWidth = rc.Right - rc.Left
      startHeight = rc.Bottom - rc.Top
    }

    const onMouseMove = async (e: MouseEvent) => {
      if (!isResizing) return

      const deltaX = e.screenX - startX
      const deltaY = e.screenY - startY

      let newWidth = startWidth
      let newHeight = startHeight

      // 根据边缘或角落计算新尺寸
      if (resizeEdge === 'right' || resizeCorner === 'top-right' || resizeCorner === 'bottom-right') {
        newWidth = startWidth + deltaX
      }
      if (resizeEdge === 'bottom' || resizeCorner === 'bottom-right' || resizeCorner === 'bottom-left') {
        newHeight = startHeight + deltaY
      }
      if (resizeEdge === 'left' || resizeCorner === 'top-left' || resizeCorner === 'bottom-left') {
        newWidth = startWidth - deltaX
      }
      if (resizeEdge === 'top' || resizeCorner === 'top-left' || resizeCorner === 'top-right') {
        newHeight = startHeight - deltaY
      }

      // 确保最小窗口尺寸
      const limitedWidth = Math.max(newWidth, minSize)
      const limitedHeight = Math.max(newHeight, minSize)

      // 计算新的窗口位置
      let newX: number | null = null
      let newY: number | null = null

      // 处理左侧边缘（包括角落）
      if (resizeEdge === 'left' || resizeCorner === 'top-left' || resizeCorner === 'bottom-left') {
        if (newWidth > minSize) {
          newX = startLeft + deltaX
        }
      }

      // 处理上侧边缘（包括角落）
      if (resizeEdge === 'top' || resizeCorner === 'top-left' || resizeCorner === 'top-right') {
        if (newHeight > minSize) {
          newY = startTop + deltaY
        }
      }

      // 设置窗口尺寸
      await window?.api?.setWindowSize(limitedWidth, limitedHeight)

      // 设置窗口位置
      if (newX !== null || newY !== null) {
        const finalX = newX !== null ? newX : startLeft
        const finalY = newY !== null ? newY : startTop
        await window?.api?.moveWindow(finalX, finalY)
      }
    }

    const onMouseUp = () => {
      isResizing = false
      resizeEdge = undefined
      resizeCorner = undefined
    }

    edges.forEach((edge) => {
      edge.addEventListener('mousedown', onMouseDown)
    })
    document.addEventListener('mousemove', onMouseMove)
    document.addEventListener('mouseup', onMouseUp)

    // 返回取消绑定函数
    return () => {
      edges.forEach((edge) => {
        edge.removeEventListener('mousedown', onMouseDown)
      })
      document.removeEventListener('mousemove', onMouseMove)
      document.removeEventListener('mouseup', onMouseUp)
    }
  }
}

export default WindowResize
