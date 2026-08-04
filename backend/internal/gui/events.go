package gui

import (
	"github.com/twgh/xcgui/edge"
	"github.com/twgh/xcgui/wapi"
	"github.com/twgh/xcgui/xcc"
)

var firstLoad = true // 第一次加载前端

// frontendReady 前端准备就绪
func (m *MainWindow) frontendReady() {
	if firstLoad {
		firstLoad = false
		// 一些初始化操作
	}
}

// regWebViewEvents 注册 WebView 事件
func (m *MainWindow) regWebViewEvents() {
	// 导航完成事件, 这个是在 frontendReady 之前加载完成的
	m.wv.Event_NavigationCompleted(func(sender *edge.ICoreWebView2, args *edge.ICoreWebView2NavigationCompletedEventArgs) uintptr {
		uri := sender.MustGetSource()
		switch uri {
		case m.getHost() + "/index.html":
			if firstLoad {
				m.w.Show()
			}
		}
		return 0
	})
}

// regXcEvents 注册炫彩事件
func (m *MainWindow) regXcEvents() {
	var wasMinimized = false // 窗口是否最小化

	// 窗口消息过程事件：监听窗口最大化/还原状态变化, 同步给前端
	// （双击标题栏或使用系统快捷键最大化时, 前端 JS 无法感知, 需在此通知）
	m.w.AddEvent_WindProc(func(hWindow int, message uint32, wParam, lParam uintptr, pbHandled *bool) int {
		switch message {
		case wapi.WM_SIZE:
			switch wParam {
			case wapi.SIZE_MINIMIZED: // 窗口最小化
				wasMinimized = true
			case wapi.SIZE_MAXIMIZED: // 窗口最大化
				wasMinimized = false
				m.wv.Eval(`window.__onWindowMaximizeStateChanged && window.__onWindowMaximizeStateChanged(true)`)
			case wapi.SIZE_RESTORED: // 窗口被还原（注意：也包括程序启动时的初始显示）
				if wasMinimized { // 从最小化状态还原
					wasMinimized = false
				} else {
					m.wv.Eval(`window.__onWindowMaximizeStateChanged && window.__onWindowMaximizeStateChanged(false)`)
				}
			}
		}
		return 0
	})

	// 窗口关闭事件
	m.w.AddEvent_Close(func(hWindow int, pbHandled *bool) int {
		*pbHandled = true // 拦截窗口关闭
		m.animateToTray() // 缩放+位移动画后隐藏到托盘
		return 0
	})

	// 托盘图标事件
	m.w.AddEvent_TrayIcon(func(wParam, lParam uintptr, pbHandled *bool) int {
		if int32(wParam) != m.tray.Id { // 不是自定义的托盘图标唯一标识符.
			return 0
		}
		switch xcc.WM_(lParam) {
		case xcc.WM_LBUTTONDOWN: // 鼠标左键按下
			m.activateWindow()
		case xcc.WM_RBUTTONDOWN: // 鼠标右键按下
			m.showTrayMenu()
		}
		return 0
	})
}
