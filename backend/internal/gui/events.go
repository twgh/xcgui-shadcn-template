package gui

import (
	"github.com/twgh/xcgui/app"
	"github.com/twgh/xcgui/common"
	"github.com/twgh/xcgui/edge"
	"github.com/twgh/xcgui/wapi"
	"github.com/twgh/xcgui/xc"
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

// 菜单项ID
const (
	menuItemSetting = iota + 1

	menuItemExit = 99999
)

// 显示托盘菜单
//
// darkMode: 是否为深色模式. 不填时使用当前后端同步的深色模式状态.
func (m *MainWindow) showTrayMenu(darkMode ...bool) int {
	mode := m.darkMode
	if len(darkMode) > 0 {
		mode = darkMode[0]
	}
	// 创建菜单. Choose 类似于三元运算符, 根据模式选择函数
	menu := common.Choose(mode, NewMenuExDark, NewMenuEx)(m.w.Handle)

	// 菜单项选择事件
	menu.AddEvent_Menu_Select(m.w.Handle, func(hWindow int, nID int32, pbHandled *bool) int {
		switch nID {
		case menuItemSetting: // 设置
			m.activateWindow()
			m.wv.Eval("alert('设置页面')")

		case menuItemExit: // 退出
			m.w.DestroyWindow() // 销毁窗口
			app.PostQuitMessage(0)
		}
		return 0
	}, false) // 这里填 false, 不然每次显示菜单都会加一个回调函数

	// 菜单_置左侧宽度
	menu.SetLeftWidth(menu.GetLeftWidth() + 16)

	// 一级菜单
	menu.AddItem(menuItemSetting, "设置", 0, xcc.Menu_Item_Flag_Normal)

	// 分隔栏
	menu.AddSeparator()

	menu.AddItem(menuItemExit, "退出", 0, xcc.Menu_Item_Flag_Normal)

	// 获取鼠标光标的位置
	var pt wapi.POINT
	wapi.GetCursorPos(&pt)
	// 弹出菜单
	menu.Popup(xc.XWnd_GetHWND(m.w.Handle), pt.X, pt.Y, 0, xcc.Menu_Popup_Position_Left_Top)
	return 0
}
