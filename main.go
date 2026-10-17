//go:build windows

// Dust Front RTS Demo 汉化补丁安装工具
package main

import (
	_ "embed"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/lxn/win"
	"golang.org/x/sys/windows"
)

// 内嵌的汉化补丁（已打好的 resources.assets）
//
//go:embed assets/resources.assets
var patchData []byte

const (
	appTitle      = "Dust Front RTS Demo 汉化补丁"
	appVersion    = "v1.1.0"
	gameDirName   = "Dust Front RTS Demo"
	dataDirName   = "Dust Front RTS_Data"
	assetFileName = "resources.assets"
	cacheDirName  = "DustFrontRTS_Hanhua"
)

var (
	user32DLL       = windows.NewLazySystemDLL("user32.dll")
	procMessageBeep = user32DLL.NewProc("MessageBeep")
)

// ---------------------------------------------------------------- 小工具

func playWarningSound() {
	const mbIconWarning = 0x00000030
	procMessageBeep.Call(uintptr(mbIconWarning))
}

func fileExists(p string) bool {
	fi, err := os.Stat(p)
	return err == nil && !fi.IsDir()
}

func exeDir() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	if resolved, err := filepath.EvalSymlinks(exe); err == nil {
		exe = resolved
	}
	return filepath.Dir(exe), nil
}

// cacheRoot 返回 %LocalAppData%\DustFrontRTS_Hanhua
func cacheRoot() (string, error) {
	base, err := os.UserCacheDir()
	if err != nil || base == "" {
		return "", fmt.Errorf("无法定位 AppData 缓存目录")
	}
	return filepath.Join(base, cacheDirName), nil
}

func backupPath(root string) string {
	return filepath.Join(root, "backup", assetFileName)
}

func tempPath(root string) string {
	return filepath.Join(root, "temp", assetFileName)
}

func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	if err := out.Sync(); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}

func writeTempPatch(p string) error {
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	return os.WriteFile(p, patchData, 0o644)
}

// ---------------------------------------------------------------- 目录检测

// findByGameDir 检查程序是否被放置在 “Dust Front RTS Demo” 目录内
func findByGameDir() (string, bool) {
	dir, err := exeDir()
	if err != nil {
		return "", false
	}
	return dir, strings.EqualFold(strings.TrimSpace(filepath.Base(dir)), gameDirName)
}

func placementMessage(action string) string {
	return "请将程序放置在游戏目录内！\n" +
		"游戏目录位置：在Steam库中选择" + gameDirName + "，\n" +
		"点击右侧齿轮图标>选择“管理”>选择“浏览本地文件”，\n" +
		"然后把本程序放到游戏目录内再进行" + action + "。"
}

// requireGameDir 做目录检测，不合格时弹出提示
func requireGameDir(owner walk.Form, action string) (string, bool) {
	dir, ok := findByGameDir()
	if !ok {
		walk.MsgBox(owner, "提示", placementMessage(action),
			walk.MsgBoxIconWarning|walk.MsgBoxOK)
		return "", false
	}
	return dir, true
}

func errorBox(owner walk.Form, msg string) {
	walk.MsgBox(owner, "错误", msg, walk.MsgBoxIconError|walk.MsgBoxOK)
}

// ---------------------------------------------------------------- 业务逻辑

func doInstall(owner walk.Form) {
	dir, ok := requireGameDir(owner, "安装")
	if !ok {
		return
	}

	asset := filepath.Join(dir, dataDirName, assetFileName)
	if !fileExists(asset) {
		errorBox(owner, "在游戏目录内未找到：\n"+asset+"\n\n请确认游戏文件完整。")
		return
	}

	root, err := cacheRoot()
	if err != nil {
		errorBox(owner, err.Error())
		return
	}
	backup := backupPath(root)

	// 首次安装时备份原始文件；已有备份则保留（避免用汉化版覆盖掉真正的原始文件）
	if !fileExists(backup) {
		if fi, statErr := os.Stat(asset); statErr == nil && fi.Size() == int64(len(patchData)) {
			walk.MsgBox(owner, "提示",
				"没有找到原始文件备份，而当前游戏文件已经是汉化版本，\n"+
					"无法自动备份原始文件。\n\n"+
					"建议先在 Steam 中对该游戏执行“验证游戏文件完整性”\n"+
					"还原成原版后，再使用本工具安装。",
				walk.MsgBoxIconWarning|walk.MsgBoxOK)
			return
		}
		if err := copyFile(asset, backup); err != nil {
			errorBox(owner, "备份原始文件失败：\n"+err.Error())
			return
		}
	}

	temp := tempPath(root)
	if err := writeTempPatch(temp); err != nil {
		errorBox(owner, "释放汉化补丁失败：\n"+err.Error())
		return
	}

	if err := copyFile(temp, asset); err != nil {
		errorBox(owner, "写入游戏目录失败：\n"+err.Error()+
			"\n\n请确认游戏已经关闭后重试。")
		return
	}

	if fi, statErr := os.Stat(asset); statErr != nil || fi.Size() != int64(len(patchData)) {
		errorBox(owner, "写入后的文件大小校验不通过，安装可能未成功。")
		return
	}

	walk.MsgBox(owner, "安装完成",
		"汉化补丁安装成功！\n\n启动游戏即可看到简体中文。",
		walk.MsgBoxIconInformation|walk.MsgBoxOK)
}

func doRestore(owner walk.Form) {
	dir, ok := requireGameDir(owner, "还原")
	if !ok {
		return
	}

	asset := filepath.Join(dir, dataDirName, assetFileName)

	root, err := cacheRoot()
	if err != nil {
		errorBox(owner, err.Error())
		return
	}
	backup := backupPath(root)

	if !fileExists(backup) {
		walk.MsgBox(owner, "提示",
			"没有找到原始文件备份，无法还原。\n\n"+
				"本工具只会在安装汉化补丁时备份原始文件，\n"+
				"如果从未安装过，则不存在可还原的备份。",
			walk.MsgBoxIconWarning|walk.MsgBoxOK)
		return
	}

	if err := copyFile(backup, asset); err != nil {
		errorBox(owner, "还原失败：\n"+err.Error()+
			"\n\n请确认游戏已经关闭后重试。")
		return
	}

	walk.MsgBox(owner, "还原完成",
		"已还原为原始游戏文件。",
		walk.MsgBoxIconInformation|walk.MsgBoxOK)
}

func showClearCacheDialog(owner *walk.MainWindow) {
	playWarningSound()

	const warnText = "警告！\n" +
		"这可能会删除本工具备份的原始游戏文件，\n" +
		"无法还原！"

	var dlg *walk.Dialog
	var okBtn, cancelBtn *walk.PushButton
	var warnBox *walk.CustomWidget

	cmd, err := Dialog{
		AssignTo:      &dlg,
		Title:         "警告",
		FixedSize:     true,
		MinSize:       Size{Width: 430, Height: 210},
		DefaultButton: &okBtn,
		CancelButton:  &cancelBtn,
		Layout:        VBox{},
		Children: []Widget{
			// 用自绘控件显示红色警告文字，
			// walk 的 Label.TextColor 在这里不生效。
			CustomWidget{
				AssignTo:         &warnBox,
				MinSize:          Size{Height: 80},
				ClearsBackground: true,
				Paint: func(canvas *walk.Canvas, updateBounds walk.Rectangle) error {
					return canvas.DrawText(warnText, warnBox.Font(), walk.RGB(200, 0, 0),
						warnBox.ClientBounds(),
						walk.TextLeft|walk.TextTop|walk.TextWordbreak)
				},
			},
			VSpacer{},
			Composite{
				Layout: HBox{MarginsZero: true},
				Children: []Widget{
					HSpacer{},
					PushButton{
						AssignTo:  &okBtn,
						Text:      "确认",
						MinSize:   Size{Width: 92, Height: 32},
						MaxSize:   Size{Width: 92, Height: 32},
						OnClicked: func() { dlg.Accept() },
					},
					PushButton{
						AssignTo:  &cancelBtn,
						Text:      "取消",
						MinSize:   Size{Width: 92, Height: 32},
						MaxSize:   Size{Width: 92, Height: 32},
						OnClicked: func() { dlg.Cancel() },
					},
				},
			},
		},
	}.Run(owner)
	if err != nil {
		errorBox(owner, "打开确认窗口失败：\n"+err.Error())
		return
	}

	if cmd != walk.DlgCmdOK {
		return
	}

	root, err := cacheRoot()
	if err != nil {
		errorBox(owner, err.Error())
		return
	}
	if err := os.RemoveAll(root); err != nil {
		errorBox(owner, "删除缓存失败：\n"+err.Error())
		return
	}
	walk.MsgBox(owner, "完成", "缓存已删除。", walk.MsgBoxIconInformation|walk.MsgBoxOK)
}

// ---------------------------------------------------------------- 界面

// centerOnScreen 在屏幕上居中窗口。
// 注意：walk 的 Bounds() 返回的是 DPI 缩放后的逻辑坐标，
// 而 GetSystemMetrics 返回物理像素，二者混用会算错，
// 因此这里全程使用 Win32 的物理坐标。
func centerOnScreen(mw *walk.MainWindow) {
	var r win.RECT
	if !win.GetWindowRect(mw.Handle(), &r) {
		return
	}
	w := r.Right - r.Left
	h := r.Bottom - r.Top
	sw := win.GetSystemMetrics(win.SM_CXSCREEN)
	sh := win.GetSystemMetrics(win.SM_CYSCREEN)

	x := (sw - w) / 2
	y := (sh - h) / 2
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	win.SetWindowPos(mw.Handle(), 0, x, y, 0, 0,
		win.SWP_NOSIZE|win.SWP_NOZORDER|win.SWP_NOACTIVATE)
}

func main() {
	var mw *walk.MainWindow

	btnSize := Size{Width: 230, Height: 48}

	if err := (MainWindow{
		AssignTo: &mw,
		Title:    appTitle,
		Size:     Size{Width: 420, Height: 268},
		MinSize:  Size{Width: 420, Height: 268},
		Layout:   VBox{},
		Children: []Widget{
			VSpacer{},
			Composite{
				Layout: HBox{MarginsZero: true},
				Children: []Widget{
					HSpacer{},
					Composite{
						Layout: VBox{MarginsZero: true, Spacing: 14},
						Children: []Widget{
							PushButton{
								Text:      "安装汉化补丁",
								MinSize:   btnSize,
								MaxSize:   btnSize,
								OnClicked: func() { doInstall(mw) },
							},
							PushButton{
								Text:      "还原游戏文件",
								MinSize:   btnSize,
								MaxSize:   btnSize,
								OnClicked: func() { doRestore(mw) },
							},
						},
					},
					HSpacer{},
				},
			},
			VSpacer{},
			Composite{
				Layout: HBox{MarginsZero: true},
				Children: []Widget{
					Label{Text: appVersion, TextColor: walk.RGB(128, 128, 128)},
					HSpacer{},
					LinkLabel{
						// 注意：walk 的 LinkLabelLink.Id() 取自 <a id="...">，
						// URL() 才取自 href，两个都写上以免判断落空。
						Text: `<a id="clearcache" href="clearcache">删除缓存</a>`,
						OnLinkActivated: func(link *walk.LinkLabelLink) {
							if link.Id() == "clearcache" || link.URL() == "clearcache" {
								showClearCacheDialog(mw)
							}
						},
					},
				},
			},
		},
	}).Create(); err != nil {
		walk.MsgBox(nil, "错误", "创建窗口失败：\n"+err.Error(), walk.MsgBoxIconError|walk.MsgBoxOK)
		return
	}

	centerOnScreen(mw)
	mw.Run()
}
