## xcgui-shadcn-ui

基于 Go + XCGUI (WebView2) + React + shadcn/ui 的项目模版, 方便快速开发项目。

## 技术栈

- **前端**: React 19 + TypeScript + Vite + shadcn/ui (base) + Tailwind CSS v4
- **后端**: Go + XCGUI (炫彩界面库) + WebView2

## 项目结构

```
├── frontend/          # React 前端
│   ├── src/
│   │   ├── App.tsx              # 主界面
│   │   ├── index.css            # css样式
│   │   ├── lib/bridge.ts        # Go <-> JS 通信桥接
│   │   └── components/ui/       # shadcn 组件
│   └── package.json
├── backend/           # Go 后端
│   ├── main.go                  # 程序入口
│   ├── go.mod
│   └── winres/                  # 图标, 版本信息, 程序清单
│   └── internal/                # 内部包
│       ├── gui/                 # 程序 GUI 界面
│       └── g/                   # 全局变量
├── scripts/           # 脚本目录
│   ├── build_all.bat           # 编译前端和后端正式版
│   ├── build_backend.bat       # 编译后端正式版
│   ├── build_frontend.bat      # 编译前端
│   ├── build_run_backend.bat   # 编译后端正式版并运行
│   ├── dev_run_all.bat         # 同时启动前端开发服务器和后端开发版本
│   ├── dev_start_backend.bat   # 启动后端开发版本
│   └── dev_start_frontend.bat  # 启动前端开发服务器
└── README.md
```

## 安装依赖

**首先确保你电脑上有 [Node.js](https://nodejs.org/zh-cn/download)**

然后安装 pnpm

```
npm install -g pnpm
```

**确保电脑上有 go-winres**

```
go install github.com/tc-hib/go-winres@latest
```

**最后在项目根目录中执行**

```bash
cd frontend && pnpm install
cd backend && go mod tidy
```

## 开发模式

```bash
# 进入脚本目录
cd scripts

# 启动前端和后端
dev_run_all.bat
```

## 生产构建

```bash
cd scripts

# 编译前端和后端
build_all.bat
```

编译后得到单个exe，内嵌了所有前端资源。

---

## 前端开发

### 添加组件

要在应用中添加组件，请运行以下命令：

```bash
cd frontend

# 使用固定版本
npx shadcn@4.16.0 add button

# 使用最新版本
npx shadcn@latest add button
```

这会将 UI 组件放置在 `frontend/src/components` 目录下。

### 使用组件

要在应用中使用这些组件，可以按如下方式进行导入：

```tsx
import { Button } from "@/components/ui/button"
```

### 更新组件

```bash
cd frontend

# 更新单个组件
npx shadcn@latest add button --overwrite

# 批量更新所有已安装组件
npx shadcn@latest add --all --overwrite
```