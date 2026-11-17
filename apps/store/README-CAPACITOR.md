# 兰瓶门店 · Capacitor APK

## 已完成

- Capacitor 8 + Android 工程（`android/`）
- 定位插件（App 内用原生定位，H5 仍走浏览器）
- 明文 HTTP（方便连电脑局域网 API）
- 命令行可直接出 debug APK（无需 Android Studio）

## 当前 debug 包

`apps/store/store-debug.apk`（约 5MB）

或：`android/app/build/outputs/apk/debug/app-debug.apk`

## 真机联调前必改

1. 编辑 `.env.capacitor`，把 `VITE_API_BASE` 改成你电脑当前局域网 IP：

```env
VITE_API_BASE=http://192.168.x.x:8080
```

2. 手机和电脑同一 Wi-Fi；电脑防火墙放行 8080；API 已启动。

3. 改 IP 后重新打包：

```bash
cd apps/store
npm run cap:sync
cd android
# JAVA_HOME 指向 JDK 21
.\gradlew.bat assembleDebug
```

## 出 APK（命令行）

本仓库已内置 SDK：`.tools/android-sdk`（`android/local.properties` 已指向它）。

需要：**JDK 21**（Capacitor 插件要求）。

```powershell
$env:JAVA_HOME = "C:\Program Files\Microsoft\jdk-21.0.13.21-hotspot"
$env:ANDROID_HOME = "d:\workSpace\lan\.tools\android-sdk"
cd apps\store
npm run cap:sync
cd android
.\gradlew.bat assembleDebug
```

产物：`android/app/build/outputs/apk/debug/app-debug.apk`

## 常用命令

| 命令 | 作用 |
|------|------|
| `npm run build:cap` | 仅打包 Web 资源 |
| `npm run cap:sync` | 打包并同步到 Android |
| `npm run cap:open` | 用 Android Studio 打开工程（可选） |
