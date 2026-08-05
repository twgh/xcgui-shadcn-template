package gui

import (
	"github.com/twgh/xcgui/app"
	"github.com/twgh/xcgui/common"
	"github.com/twgh/xcgui/wapi"
	"github.com/twgh/xcgui/xc"
	"github.com/twgh/xcgui/xcc"
)

// 菜单项ID

const (
	menuItemSetting = iota + 1 // 设置

	menuItemExit = 99999 // 退出
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
	menu.AddEvent_Menu_Select(m.w.Handle, m.onMenuSelect, false) // 这里填 false, 不然每次显示菜单都会加一个回调函数

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

// 菜单项选择事件
func (m *MainWindow) onMenuSelect(hWindow int, nID int32, pbHandled *bool) int {
	switch nID {
	case menuItemSetting: // 设置
		m.activateWindow()
		m.wv.Eval("alert('设置页面')")

	case menuItemExit: // 退出
		m.w.DestroyWindow() // 销毁窗口
		app.PostQuitMessage(0)
	}

	return 0
}
