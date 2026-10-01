# Dust Front RTS Demo 汉化补丁

用于游戏 `Dust Front RTS Demo` 的汉化补丁安装工具。

## 使用方法

1. 从 [Releases](https://github.com/RainNight0/Dust-Front-RTS_Chinese/releases) 下载最新的 zip 压缩包并解压。
2. 将 `Dust Front RTS Demo汉化补丁.exe` 复制到游戏目录内。

   > 游戏目录位置：在 Steam 库中选择 `Dust Front RTS Demo`，
   > 点击右侧齿轮图标 > 选择“管理” > 选择“浏览本地文件”。

3. 在游戏目录内运行该程序，点击 **安装汉化补丁**。
4. 需要恢复原版时，再次运行该程序，点击 **还原游戏文件**。

> 安装前请先关闭游戏。原始文件会自动备份到 `%LocalAppData%\DustFrontRTS_Hanhua`，
> 点击程序右下角的“删除缓存”会一并删除该备份，删除后无法再还原。

## 自己构建

将仓库拉取到本地，在项目目录中使用下方命令运行 PowerShell 脚本：

```powershell
powershell -ExecutionPolicy Bypass -File build.ps1
```

构建产物为项目根目录下的 `Dust Front RTS Demo汉化补丁.exe`。

也可以指定输出路径：

```powershell
powershell -ExecutionPolicy Bypass -File build.ps1 -OutFile "dist\patch.exe"
```

### 构建依赖

- Go（版本见 [`go.mod`](go.mod)）
- 首次构建会自动执行 `go install github.com/akavel/rsrc@latest`

`rsrc` 会用 [`app.manifest`](app.manifest) 生成 `rsrc.syso`，为程序启用
Common Controls v6，缺少它 walk 界面会在启动时报 `TTM_ADDTOOL failed`。
`rsrc.syso` 属于构建产物，已加入 `.gitignore`，请勿直接 `go build` 而不经过 `build.ps1`。

## 说明

- 补丁内容内嵌在程序中（`assets/resources.assets`），安装时会写入
  `<游戏目录>\Dust Front RTS_Data\resources.assets`。
- 本仓库仅包含汉化补丁工具与补丁数据，游戏本体版权归原作者所有。

## 鸣谢

本项目由 [`爱吃白饭的蓝色大肥鱼`](https://platform.deepseek.com/) 提供技术支持

## 许可证

[MIT](LICENSE)
