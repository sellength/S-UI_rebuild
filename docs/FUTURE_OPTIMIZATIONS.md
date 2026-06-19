# Future Optimizations & Code Cleanups

This document serves as a guide for removing the legacy single-node `Sing-Box Core` dashboard card and its associated single-node error toast notifications, transitioning the panel to a pure distributed Master-Node architecture.

---

## 1. Remove "Sing-Box Core" Status Card on Dashboard

### Frontend View Component
* **File**: `frontend/src/components/Main.vue` ([Main.vue](file:///Users/junzhuang/Antigravity/s-ui2/frontend/src/components/Main.vue))
* **Target to Edit**: The `Sing-Box Core` section in the right sidebar.
* **Details**:
  * Locate the template code from around line 189 to line 232 (`<!-- Sing-Box Core 服务 -->` block).
  * Remove the entire HTML block containing the service metrics and buttons (Start, Stop, Restart, logs).
  * Cleanup the corresponding methods in `<script lang="ts" setup>` around the end of the file (e.g. `restartSingbox`, `stopSingbox` handlers).

---

## 2. Eliminate Single-Node Sing-Box Error Toasts

### Backend API Handler
* **File**: `backend/api/api.go` ([api.go](file:///Users/junzhuang/Antigravity/s-ui2/backend/api/api.go))
* **Target to Edit**: The `loadData` function around lines 452 to 458.
* **Details**:
  * Current logic reads:
    ```go
    sysInfo := a.ServerService.GetSingboxInfo()
    if sysInfo["running"] == false {
        logs := a.ServerService.GetLogs("sing-box", "1", "debug")
        if len(logs) > 0 {
            data["lastLog"] = logs[0]
        }
    }
    ```
  * Action: Remove this entire block. Since the Master panel does not run its own proxy core in a distributed setup, it shouldn't check if the local process is running, nor should it fetch logs to return as `lastLog`.

### Frontend Store Action
* **File**: `frontend/src/store/modules/data.ts` ([data.ts](file:///Users/junzhuang/Antigravity/s-ui2/frontend/src/store/modules/data.ts))
* **Target to Edit**: Inside the `loadData()` action around lines 31 to 37.
* **Details**:
  * Current logic reads:
    ```typescript
    if (msg.obj.lastLog) {
      push.error({
        title: i18n.global.t('error.core'),
        duration: 5000,
        message: msg.obj.lastLog
      })
    }
    ```
  * Action: Remove or comment out this block. It stops the frontend from popping up the `Sing-Box Error` notification.

---

## 3. Localization Cleanups (Optional)
* **Files**: i18n files under `frontend/src/locales/` (e.g. `en.ts`, `zhcn.ts`)
* **Details**: Clean up unused keys under `error.core` and `main.info.sbd`.

---

## 4. Remove "Local Proxy Account" Option in Client Modal (Completed)

> [!NOTE]
> **已完成**：此优化内容已于 2026-06-19 彻底实装完成。
> 1. 前端 `Client.vue`、`Clients.vue`、`QrCode.vue` 等页面的“账号类型”切换、本地入站绑定输入框以及相关的可用范围列均已精简并移除，且默认并强制采用 `cluster` 集群模式。
> 2. 后端 `client_scope.go` 中的验证逻辑已强制对所有客户端开放集群权限，并阻断本地代理，原先的单元测试用例也已相应地更新完毕。

