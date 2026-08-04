package gui

import (
	"xcgui-shadcn-ui/internal/g"

	"github.com/twgh/xcgui/xc"
	"github.com/twgh/xcgui/xcc"
)

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
	// 窗口是否最大化
	m.wv.Bind("api.isMaxWindow", func() bool {
		return m.w.IsMaxWindow()
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

	// ===== 主题 =====
	// 同步深色模式状态（前端主题变化时调用, 用于托盘菜单样式）
	m.wv.Bind("api.setDarkMode", func(dark bool) {
		m.darkMode = dark
	})
}
