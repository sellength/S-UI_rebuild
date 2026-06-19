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
