import { useCallback, useEffect, useState } from "react"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Separator } from "@/components/ui/separator"
import { bridge } from "@/lib/bridge"
import WindowDrag from "@/lib/window-drag"
import WindowResize from "@/lib/window-resize"
import {
  MinusIcon,
  XIcon,
  SquareIcon,
  CopyIcon,
  MonitorIcon,
  WrenchIcon,
  LayersIcon,
  ZapIcon,
  PaletteIcon,
  BoxIcon,
} from "lucide-react"

const features = [
  { icon: MonitorIcon, title: "透明窗口", desc: "圆角 + 阴影 + 透明背景，原生桌面体验" },
  { icon: ZapIcon, title: "Go ↔ JS Bridge", desc: "前后端双向通信，前端直接调用 Go 函数" },
  { icon: LayersIcon, title: "shadcn/ui 组件库", desc: "丰富的 UI 组件，开箱即用" },
  { icon: PaletteIcon, title: "Tailwind CSS v4", desc: "原子化 CSS，快速高效地构建界面" },
  { icon: WrenchIcon, title: "一键开发", desc: "单条命令同时启动前后端开发模式" },
  { icon: BoxIcon, title: "单文件发布", desc: "编译为单个 exe，内嵌所有前端资源" },
]

export function App() {
  const [version, setVersion] = useState("")
  const [isMaximized, setIsMaximized] = useState(false)

  useEffect(() => {
    // 启用窗口拖动（排除按钮、输入框等交互元素）
    const cleanupDrag = WindowDrag.enable(
      "#app-content",
      ".titlebar, button, a, input, select, textarea"
    )

    // 启用窗口边框拖动调整大小
    const cleanupResize = WindowResize.enable(400)

    // 获取版本号
    bridge.getVersion().then((v) => {
      if (v) setVersion(v)
    })

    // 通知后端前端已就绪
    bridge.frontendReady()

    return () => {
      cleanupDrag()
      cleanupResize()
    }
  }, [])

  // 切换最大化时同步 body class
  useEffect(() => {
    if (isMaximized) {
      document.body.classList.add("window-maximized")
    } else {
      document.body.classList.remove("window-maximized")
    }
    return () => {
      document.body.classList.remove("window-maximized")
    }
  }, [isMaximized])

  const handleToggleMaximize = useCallback(() => {
    setIsMaximized((prev) => !prev)
    bridge.toggleWindowMaximize()
  }, [])

  return (
    <div id="shadow-container">
      {/* 边框调整热区 - 放在 content 外面，避免被 overflow:hidden 裁剪 */}
      <div className="resize-edge top" data-edge="top" />
      <div className="resize-edge right" data-edge="right" />
      <div className="resize-edge bottom" data-edge="bottom" />
      <div className="resize-edge left" data-edge="left" />
      <div className="resize-edge top-right" data-corner="top-right" />
      <div className="resize-edge bottom-right" data-corner="bottom-right" />
      <div className="resize-edge bottom-left" data-corner="bottom-left" />
      <div className="resize-edge top-left" data-corner="top-left" />

      <div id="content">
        <div id="app-content" className="flex flex-1 flex-col min-h-0 select-none">
          {/* 标题栏 */}
          <div className="titlebar flex h-9 shrink-0 items-center justify-between border-b bg-muted/30 px-2">
            <span className="px-2 text-xs text-muted-foreground">xcgui-shadcn-ui</span>
            <div className="titlebar-controls flex items-center">
              <Button
                variant="ghost"
                size="icon"
                className="size-8 rounded-none"
                onClick={() => bridge.minimizeWindow()}
              >
                <MinusIcon data-icon="inline-start" />
              </Button>
              <Button
                variant="ghost"
                size="icon"
                className="size-8 rounded-none"
                onClick={handleToggleMaximize}
              >
                {isMaximized ? (
                  <CopyIcon data-icon="inline-start" className="size-3" />
                ) : (
                  <SquareIcon data-icon="inline-start" className="size-3" />
                )}
              </Button>
              <Button
                variant="ghost"
                size="icon"
                className="size-8 rounded-none hover:bg-destructive hover:text-destructive-foreground"
                onClick={() => bridge.closeWindow()}
              >
                <XIcon data-icon="inline-start" />
              </Button>
            </div>
          </div>

          {/* 主内容 */}
          <div className="flex flex-1 flex-col items-center justify-center p-6">
            <div className="flex w-full flex-col gap-6">
              {/* 欢迎区域 */}
              <div className="text-center">
                <h1 className="font-heading text-2xl font-semibold">
                  xcgui-shadcn-ui
                </h1>
                <p className="mt-1 text-sm text-muted-foreground">
                  基于 Go + XCGUI + React + shadcn/ui 的桌面应用开发模版
                </p>
                {version && (
                  <Badge variant="secondary" className="mt-2">
                    v{version}
                  </Badge>
                )}
              </div>

              <Separator />

              {/* 技术栈 */}
              <Card>
                <CardHeader>
                  <CardTitle className="text-base">技术栈</CardTitle>
                  <CardDescription>
                    项目使用的主要技术
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="grid grid-cols-2 gap-3 sm:grid-cols-4">
                    {[
                      { label: "后端", value: "Go + XCGUI", sub: "WebView2" },
                      { label: "前端", value: "React 19", sub: "TypeScript + Vite" },
                      { label: "UI 框架", value: "shadcn/ui", sub: "Tailwind CSS v4" },
                      { label: "打包", value: "单 exe 文件", sub: "内嵌前端资源" },
                    ].map(({ label, value, sub }) => (
                      <div
                        key={label}
                        className="rounded-lg border bg-muted/30 p-3"
                      >
                        <div className="text-xs text-muted-foreground">
                          {label}
                        </div>
                        <div className="mt-1 text-sm font-medium">{value}</div>
                        <div className="text-xs text-muted-foreground">
                          {sub}
                        </div>
                      </div>
                    ))}
                  </div>
                </CardContent>
              </Card>

              {/* 特性 */}
              <Card>
                <CardHeader>
                  <CardTitle className="text-base">特性</CardTitle>
                  <CardDescription>
                    模版已内置的功能
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
                    {features.map(({ icon: Icon, title, desc }) => (
                      <div key={title} className="flex items-start gap-3">
                        <div className="flex size-8 shrink-0 items-center justify-center rounded-lg bg-primary/10">
                          <Icon className="size-4 text-primary" />
                        </div>
                        <div className="min-w-0 flex-1">
                          <div className="text-sm font-medium">{title}</div>
                          <div className="text-xs text-muted-foreground">
                            {desc}
                          </div>
                        </div>
                      </div>
                    ))}
                  </div>
                </CardContent>
              </Card>

              {/* 底部提示 */}
              <div className="text-center">
                <p className="text-xs text-muted-foreground">
                  通过标题栏或空白区域拖动窗口 · 按{" "}
                  <kbd data-slot="kbd">d</kbd>{" "}
                  切换深色/浅色模式
                </p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

export default App
