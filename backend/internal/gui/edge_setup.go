package gui

import (
	"embed"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"xcgui-shadcn-ui/internal/g"

	"github.com/twgh/xcgui/edge"
	"github.com/twgh/xcgui/wapi"
)

//go:embed dist/**
var embedAssets embed.FS

const hostName = "app.example"

// getHost 返回主机名
func (m *MainWindow) getHost() string {
	if g.IsDebug() {
		return "http://localhost:5173"
	}
	return edge.JoinUrlHeader(hostName)
}

// setupDevServer 输出 Vite 开发服务器
func (m *MainWindow) setupDevServer() {
	fmt.Println("开发模式: 连接 Vite 开发服务器 http://localhost:5173")
}

// setupEmbedFS 使用嵌入文件系统
func (m *MainWindow) setupEmbedFS() {
	fmt.Println("正式模式: 使用嵌入文件系统")
	err := edge.SetVirtualHostNameToEmbedFSMapping(hostName, embedAssets)
	if err != nil {
		wapi.MessageBoxW(0, "SetVirtualHostNameToEmbedFSMapping 失败: "+err.Error(), "错误", wapi.MB_OK|wapi.MB_IconError)
		os.Exit(5)
	}
	err = m.wv.EnableVirtualHostNameToEmbedFSMapping(true)
	if err != nil {
		wapi.MessageBoxW(0, "EnableVirtualHostNameToEmbedFSMapping 失败: "+err.Error(), "错误", wapi.MB_OK|wapi.MB_IconError)
		os.Exit(6)
	}
}

// saveMemory 节省内存
func (m *MainWindow) saveMemory() {
	// 设置内存使用目标级别，能节省近一半占用内存
	m.wv.WebView2_19 = m.wv.CoreWebView.MustGetICoreWebView2_19()
	if m.wv.WebView2_19 != nil {
		m.wv.WebView2_19.SetMemoryUsageTargetLevel(edge.COREWEBVIEW2_MEMORY_USAGE_TARGET_LEVEL_LOW)
	}
}

func createEdge() *edge.Edge {
	edg, err := edge.New(edge.Option{
		UserDataFolder: filepath.Join(os.Getenv("APPDATA"), g.AppName),
		EnvOptions: &edge.EnvOptions{
			DisableTrackingPrevention: true,
			ScrollBarStyle:            edge.COREWEBVIEW2_SCROLLBAR_STYLE_FLUENT_OVERLAY,
		},
	})
	if err != nil {
		wapi.MessageBoxW(0, "创建 WebView 环境失败: "+err.Error(), "错误", wapi.MB_OK|wapi.MB_IconError)
		os.Exit(1)
	}
	return edg
}

// checkWebView2 自动检测 WebView2 运行时是否安装, 未安装则从网络下载安装引导程序进行安装
func checkWebView2() {
	// 输出本库使用的 WebView2 版本
	fmt.Println("本库使用的 WebView2 运行时版本号:", edge.GetVersion())

	// 获取本机已安装的 WebView2 运行时版本
	localVersion, err := edge.GetAvailableBrowserVersion()
	if err != nil {
		wapi.MessageBoxW(0, "获取 WebView2 运行时版本号失败: "+err.Error(), "提示", wapi.MB_IconError)
		os.Exit(1)
	}

	// 未安装, 自动运行安装引导程序
	if localVersion == "" {
		go func() {
			wapi.MessageBoxW(0, "首次运行请等待安装必要的 WebView2 运行环境...\n(本弹窗可关闭, 不影响安装运行环境)", "提示", wapi.MB_OK)
		}()

		// 运行 WebView2 运行时的小型安装引导程序
		err := RunWebView2Installer()
		if err != nil {
			wapi.MessageBoxW(0, "安装 WebView2 运行时失败: "+err.Error(), "错误", wapi.MB_OK|wapi.MB_IconError)
			os.Exit(2)
		}

		// 等待安装完成, 等使用 edge.GetAvailableBrowserVersion() 获取到版本号就是安装完成了
		for i := 0; i < 300; i++ {
			time.Sleep(time.Second)
			localVersion, _ = edge.GetAvailableBrowserVersion()
			if localVersion != "" {
				break
			}
		}
		if localVersion == "" {
			wapi.MessageBoxW(0, "WebView2 运行时安装超时, 请检查网络后重新打开程序!", "错误", wapi.MB_OK|wapi.MB_IconError)
			os.Exit(3)
		}
	}

	fmt.Println("本机安装的 WebView2 运行时版本号:", localVersion)

	if ret, _ := edge.CompareBrowserVersions(localVersion, edge.GetVersion()); ret == -1 {
		fmt.Println("本机 WebView2 运行时版本低于本库使用的版本!")
	}
}

// RunWebView2Installer 从网络下载 WebView2 运行时的小型安装引导程序并运行.
//
//   - 使用 edge.WebView2DownloadLink 指向的官方下载地址.
//   - isSilent: 是否静默安装.
func RunWebView2Installer(isSilent ...bool) error {
	// 下载安装引导程序到临时目录
	installerPath, err := DownloadWebView2Installer()
	if err != nil {
		return fmt.Errorf("下载 WebView2 安装引导程序失败: %w", err)
	}

	// 运行安装程序, 等待安装完成
	cmd := exec.Command(installerPath)
	// 是否静默安装
	if len(isSilent) > 0 && isSilent[0] {
		cmd.Args = append(cmd.Args, "/silent", "/install")
	}

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("运行安装程序失败: %w", err)
	}

	return nil
}

// DownloadWebView2Installer 从网络下载 WebView2 运行时的小型安装引导程序到临时目录, 返回其路径.
func DownloadWebView2Installer() (string, error) {
	installerPath := filepath.Join(os.TempDir(), "MicrosoftEdgeWebview2Setup_"+time.Now().Format("20060102")+".exe")

	resp, err := http.Get(edge.WebView2DownloadLink)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("下载失败, HTTP 状态码: %d", resp.StatusCode)
	}

	// 写出文件
	out, err := os.Create(installerPath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", err
	}

	return installerPath, nil
}
