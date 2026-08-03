package gui

import (
	"fmt"
	"os"
	"time"

	"xcgui-shadcn-ui/internal/g"

	"github.com/twgh/xcgui/common"
	"github.com/twgh/xcgui/ease"
	"github.com/twgh/xcgui/edge"
	"github.com/twgh/xcgui/wapi"
	"github.com/twgh/xcgui/wapi/wutil"
	"github.com/twgh/xcgui/window"
	"github.com/twgh/xcgui/xc"
	"github.com/twgh/xcgui/xcc"
)

type MainWindow struct {
	edg  *edge.Edge
	w    *window.Window
	wv   *edge.WebView
	tray *window.TrayIcon // 托盘图标

	origWidth  int32    // 窗口原始宽度
	origHeight int32    // 窗口原始高度
	lastPos    xc.POINT // 窗口上一次的位置, -999 代表无效位置

}

func NewMainWindow(edg *edge.Edge) *MainWindow {
	m := &MainWindow{
		edg:        edg,
		origWidth:  900,
		origHeight: 700,
		lastPos:    xc.POINT{X: -999, Y: -999},
	}

	var err error
	m.w, m.wv, err = m.edg.NewWebViewWithWindow(
		edge.WithXmlWindowTitle(g.AppName),
		edge.WithXmlWindowClassName(g.AppName),
		edge.WithXmlWindowSize(m.origWidth, m.origHeight),
		edge.WithFillParent(true),
		edge.WithDebug(g.IsDebug()),
		edge.WithDefaultContextMenus(g.IsDebug()),
		edge.WithBrowserAcceleratorKeys(g.IsDebug()),
		edge.WithStatusBar(false),
		edge.WithZoomControl(false),
		edge.WithAutoFocus(true),
		edge.WithAppDrag(true),
		edge.WithDefaultBackgroundColor(edge.NewColor(0, 0, 0, 0)),
	)
	if err != nil {
		wapi.MessageBoxW(0, "创建 WebView 失败: "+err.Error(), "错误", wapi.MB_OK|wapi.MB_IconError)
		os.Exit(1)
	}

	// 设置为透明窗口
	m.w.SetTransparentType(xcc.Window_Transparent_Shaped)
	// 设置窗口最小尺寸
	m.w.SetMinimumSize(400, 400)
	// 设置图标
	m.setIcon()

	// 注册炫彩事件
	m.regXcEvents()

	// 正式版嵌入文件, 调试版使用 dev server
	if !g.IsDebug() {
		m.setupEmbedFS()
	} else {
		m.setupDevServer()
	}

	// 节省 WebView 内存
	m.saveMemory()
	// 注册 WebView 事件
	m.regWebViewEvents()
	// 绑定函数
	m.bindFunctions()
	// 加载首页
	m.wv.Navigate(m.getHost() + "/index.html")
	return m
}

// setIcon 设置图标
func (m *MainWindow) setIcon() {
	// 从资源中加载程序图标
	hMod := wapi.GetModuleHandleW("")
	hIconApp := wapi.LoadImageW(hMod, common.StrPtr("APPICON"), wapi.IMAGE_ICON, 0, 0, wapi.LR_SHARED|wapi.LR_DEFAULTSIZE)

	// 设置任务栏预览窗口左上角的图标, 使用24x24尺寸, 也会影响任务管理器里的图标
	hIcon24 := wapi.LoadImageW(hMod, common.StrPtr("APPICON"), wapi.IMAGE_ICON, 24, 24, wapi.LR_SHARED)
	m.w.SetSmallIcon(hIcon24)
	// 设置大图标, 会影响任务栏图标, Alt+Tab 窗口图标
	m.w.SetBigIcon(hIconApp)

	// 创建托盘图标
	m.tray = m.w.CreateTrayIcon(hIconApp, g.AppName)
	// 显示托盘图标
	m.tray.Show()
}

// regXcEvents 注册炫彩事件
func (m *MainWindow) regXcEvents() {
	// 窗口消息过程事件：监听窗口最大化/还原状态变化, 同步给前端
	// （双击标题栏或使用系统快捷键最大化时, 前端 JS 无法感知, 需在此通知）
	var wasMinimized = false // 窗口是否最小化
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

// frontendReady 前端准备就绪
func (m *MainWindow) frontendReady() {
	if firstLoad {
		firstLoad = false
		// 一些初始化操作
	}
}

var firstLoad = true // 第一次加载前端

// regWebViewEvents 注册 WebView 事件
func (m *MainWindow) regWebViewEvents() {
	// 导航完成事件, 这个是在 frontendReady 之前加载完成的
	m.wv.Event_NavigationCompleted(func(sender *edge.ICoreWebView2, args *edge.ICoreWebView2NavigationCompletedEventArgs) uintptr {
		uri := sender.MustGetSource()
		fmt.Println("导航完成:", uri)
		if uri == m.getHost()+"/index.html" {
			if firstLoad {
				m.w.Show()
			}
		}
		return 0
	})
}

// bindFunctions 绑定函数
func (m *MainWindow) bindFunctions() {
	// ===== 窗口控制 =====
	// 最小化窗口
	m.wv.Bind("api.minimizeWindow", func() {
		m.w.ShowWindow(xcc.SW_MINIMIZE)
	})
	// 切换窗口最大化
	m.wv.Bind("api.toggleWindowMaximize", func() {
		m.w.MaxWindow(!m.w.IsMaxWindow())
	})
	// 关闭窗口
	m.wv.Bind("api.closeWindow", func() {
		m.w.CloseWindow()
	})
	// 移动窗口
	m.wv.Bind("api.moveWindow", func(x, y int32) {
		m.w.SetPosition(m.w.DpiConv(x), m.w.DpiConv(y))
	})
	// 设置窗口尺寸
	m.wv.Bind("api.setWindowSize", func(width, height int32) {
		m.w.SetSize(width, height)
	})
	// 获取窗口矩形
	m.wv.Bind("api.getWindowRect", func() xc.RECT {
		return m.w.GetRectEx()
	})

	// ===== 系统 =====
	// 前端准备就绪
	m.wv.Bind("api.frontendReady", func() {
		m.frontendReady()
	})

	// 获取版本号
	m.wv.Bind("api.getVersion", func() string {
		return g.Version
	})
}

// activateWindow 激活窗口到前台
func (m *MainWindow) activateWindow() {
	m.wv.Show() // 显示 WebView
	if m.lastPos.X != -999 && m.lastPos.Y != -999 {
		// 恢复窗口原始大小和位置（动画过程中可能被缩放过）
		m.w.SetRect(&xc.RECT{Left: m.lastPos.X, Top: m.lastPos.Y, Right: m.lastPos.X + m.origWidth, Bottom: m.lastPos.Y + m.origHeight})
		// 设置过位置后, 将上次位置重置为无效值
		m.lastPos.X = -999
		m.lastPos.Y = -999
	}

	m.w.ShowWindow(xcc.SW_SHOWNORMAL)
}

// hideWindow 隐藏窗口和 WebView
//   - 加一个 WebView 的隐藏是因为这样能让它在后台自动变成效率模式
func (m *MainWindow) hideWindow() {
	m.wv.Show(false)
	m.w.Show(false)
}

// animateToTray 缩放+位移动画：窗口等比例缩小并向屏幕右下角托盘区域平移，结束后隐藏窗口。
func (m *MainWindow) animateToTray() {
	// 获取窗口大小和位置
	rc := m.w.GetRectEx()
	m.lastPos.X = rc.Left
	m.lastPos.Y = rc.Top
	winW := rc.Right - rc.Left
	winH := rc.Bottom - rc.Top
	// 窗口中心坐标
	winCX := (rc.Left + rc.Right) / 2
	winCY := (rc.Top + rc.Bottom) / 2

	dpi := m.w.GetDPI()
	// 屏幕右下角（托盘位置）, 这个获取的屏幕大小是物理坐标, 比如2560*1600,
	// 要转换成逻辑坐标, 比如在系统 150% 缩放下计算出来是 1707*1067
	targetCX := wapi.MulDiv(wutil.GetScreenWidth(), 96, dpi)
	targetCY := wapi.MulDiv(wutil.GetScreenHeight(), 96, dpi)
	// 缩放到 80% 屏幕宽度位置
	targetCX = int32(0.8 * float32(targetCX))

	// 缓动动画，每步 10ms
	const steps = 15
	for t := 0; t < steps; t++ {
		v := ease.Quad(float32(t)/float32(steps), xcc.Ease_Type_InOut)

		// 缩放：1.0 → 0.01（避免缩到 0 导致窗口消失闪烁）
		scale := float32(1.0 - v*0.99)

		// 新窗口大小
		newW := int32(float32(winW) * scale)
		newH := int32(float32(winH) * scale)

		// 新窗口中心：从原位置插值到屏幕右下角
		newCX := int32(float32(winCX) + v*float32(targetCX-winCX))
		newCY := int32(float32(winCY) + v*float32(targetCY-winCY))

		// 逆推左上角坐标
		newLeft := newCX - newW/2
		newTop := newCY - newH/2

		rect := xc.RECT{
			Left:   newLeft,
			Top:    newTop,
			Right:  newLeft + newW,
			Bottom: newTop + newH,
		}
		m.w.SetRect(&rect).Redraw(true)
		time.Sleep(time.Millisecond * 10)
	}

	// 动画结束后隐藏窗口
	m.hideWindow()
}
