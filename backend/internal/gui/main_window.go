package gui

import (
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

	darkMode bool // 当前深色模式状态（由前端同步）
}

func NewMainWindow(edg *edge.Edge) *MainWindow {
	m := &MainWindow{
		edg:        edg,
		origWidth:  900,
		origHeight: 710,
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

	if m.w.IsMaxWindow() {
		m.w.ShowWindow(xcc.SW_SHOWMAXIMIZED)
	} else {
		m.w.ShowWindow(xcc.SW_SHOWNORMAL)
	}
	wapi.SetForegroundWindow(m.w.GetHWND())
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
