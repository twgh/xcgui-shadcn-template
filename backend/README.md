# 后端开发指南

## 修改应用名称

在项目中搜索 `xcgui-shadcn-ui` 并全部替换为你的应用名称

## 修改应用图标

替换 `backend\winres\icon.ico` 文件后, 执行 `scripts\build_backend.bat` 可看到图标已更新

## 修改窗口默认尺寸

修改 `NewMainWindow` 函数中的 `origWidth` 和 `origHeight`

## 绑定 Go 函数到前端调用

在 `backend\internal\gui\bridge.go` 中添加绑定函数, 然后在 `frontend\src\lib\bridge.ts` 中添加相关定义

## 添加托盘菜单

在 `backend\internal\gui\tray_menu.go` 中添加菜单项ID, 使用 `menu.AddItem` 添加菜单项, 在 `onMenuSelect` 函数中添加菜单项选中事件