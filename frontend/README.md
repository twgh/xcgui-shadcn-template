# React + TypeScript + Vite + shadcn/ui

这是一个基于 React、TypeScript 和 shadcn/ui 的 Vite 项目。

## 添加组件

要在应用中添加组件，请运行以下命令：

```bash
# 使用固定版本
npx shadcn@4.16.0 add button

# 使用最新版本
npx shadcn@latest add button
```

这会将 UI 组件放置在 src/components 目录下。

## 使用组件

要在应用中使用这些组件，可以按如下方式进行导入：

```tsx
import { Button } from "@/components/ui/button"
```

## 更新组件

```bash
# 更新单个组件
npx shadcn@latest add button --overwrite

# 批量更新所有已安装组件
npx shadcn@latest add --all --overwrite
```