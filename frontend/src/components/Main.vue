<template>
  <LogVue
    v-model="logModal.visible"
    :visible="logModal.visible"
    :logType="logModal.logType"
    @close="closeLogs"
  />

  <v-container fluid class="panel-layout-container py-6 px-6">
    <v-row>
      <!-- 左侧大主栏 -->
      <v-col cols="12" md="9" class="d-flex flex-column pr-md-4">
        
        <!-- 一、概览大卡片 (Overview) -->
        <v-card class="panel-card px-6 py-5 mb-5" variant="flat">
          <div class="card-section-title mb-4">概览</div>
          <v-row class="text-center justify-space-around align-center">
            <v-col cols="6" sm="3" class="stat-col">
              <div class="stat-num">{{ inboundCount }}</div>
              <div class="stat-label">节点数量</div>
            </v-col>
            <v-col cols="6" sm="3" class="stat-col">
              <div class="stat-num">{{ clientCount }}</div>
              <div class="stat-label">用户总数</div>
            </v-col>
            <v-col cols="6" sm="3" class="stat-col">
              <div class="stat-num color-online">{{ onlineCount }}</div>
              <div class="stat-label">在线用户</div>
            </v-col>
            <v-col cols="6" sm="3" class="stat-col no-border">
              <div class="stat-num">{{ ruleCount }}</div>
              <div class="stat-label">路由规则</div>
            </v-col>
          </v-row>
        </v-card>

        <!-- 二、状态大卡片 (Status Circular Gauges) -->
        <v-card class="panel-card px-6 py-5 mb-5" variant="flat">
          <div class="card-section-title mb-3">状态</div>
          <v-row class="align-center justify-space-around">
            <v-col cols="6" sm="3" class="d-flex justify-center">
              <Gauge :tilesData="tilesData" type="g-load" />
            </v-col>
            <v-col cols="6" sm="3" class="d-flex justify-center">
              <Gauge :tilesData="tilesData" type="g-cpu" />
            </v-col>
            <v-col cols="6" sm="3" class="d-flex justify-center">
              <Gauge :tilesData="tilesData" type="g-mem" />
            </v-col>
            <v-col cols="6" sm="3" class="d-flex justify-center">
              <Gauge :tilesData="tilesData" type="g-disk" />
            </v-col>
          </v-row>
        </v-card>

        <!-- 三、监控大卡片 (Monitor Area Chart) -->
        <v-card class="panel-card px-6 py-5" variant="flat">
          <v-row class="align-center justify-space-between mb-4">
            <v-col cols="auto">
              <div class="card-section-title">监控</div>
            </v-col>
            <v-col cols="auto">
              <!-- 折线类型切换按钮组 -->
              <v-btn-toggle
                v-model="chartToggle"
                mandatory
                density="comfortable"
                variant="outlined"
                style="border-radius: 6px; border: 1px solid rgba(255,255,255,0.08);"
              >
                <v-btn value="net" size="small" class="px-4 text-none">流量</v-btn>
                <v-btn value="pnet" size="small" class="px-4 text-none">封包</v-btn>
              </v-btn-toggle>
            </v-col>
          </v-row>

          <!-- 流量统计徽章 (Up, Down, Sent, Recv) -->
          <v-row class="mb-5 gap-y-3">
            <v-col cols="6" sm="3">
              <div class="monitor-badge upload">
                <span class="badge-label">上行速度</span>
                <span class="badge-value">{{ currentUploadSpeed }}</span>
              </div>
            </v-col>
            <v-col cols="6" sm="3">
              <div class="monitor-badge download">
                <span class="badge-label">下行速度</span>
                <span class="badge-value">{{ currentDownloadSpeed }}</span>
              </div>
            </v-col>
            <v-col cols="6" sm="3">
              <div class="monitor-badge sent">
                <span class="badge-label">总发送</span>
                <span class="badge-value">{{ totalSent }}</span>
              </div>
            </v-col>
            <v-col cols="6" sm="3">
              <div class="monitor-badge recv">
                <span class="badge-label">总接收</span>
                <span class="badge-value">{{ totalRecv }}</span>
              </div>
            </v-col>
          </v-row>

          <!-- Chart.js 面积图 -->
          <div class="chart-container-wrapper mt-2">
            <History :tilesData="tilesData" :type="activeChartType" />
          </div>
        </v-card>

      </v-col>

      <!-- 右侧侧边栏 -->
      <v-col cols="12" md="3" class="d-flex flex-column pl-md-2 mt-5 mt-md-0">
        
        <!-- 一、系统信息卡片 -->
        <v-card class="panel-card px-6 py-5 mb-5" variant="flat">
          <div class="card-section-title mb-4">系统信息</div>
          <div class="info-list-wrapper">
            <div class="info-item">
              <span class="label">主机名称</span>
              <span class="value font-weight-bold">{{ tilesData.sys?.hostName || '-' }}</span>
            </div>
            <div class="info-item">
              <span class="label">内核处理器</span>
              <span class="value select-text">
                <v-tooltip activator="parent" location="top">
                  CPU: {{ tilesData.sys?.cpuType || '-' }}
                </v-tooltip>
                {{ tilesData.sys?.cpuCount || 1 }} 核
              </span>
            </div>
            <div class="info-item">
              <span class="label">局域网地址 (LAN)</span>
              <span class="value select-text font-weight-bold">
                {{ [...(tilesData.sys?.ipv4 || []), ...(tilesData.sys?.ipv6 || [])].join(', ') || '-' }}
              </span>
            </div>
            <div class="info-item">
              <span class="label">公网 IPv4 (WAN)</span>
              <span class="value select-text font-weight-bold text-primary">
                {{ tilesData.sys?.wanIpv4 || 'N/A' }}
              </span>
            </div>
            <div class="info-item">
              <span class="label">公网 IPv6 (WAN)</span>
              <span class="value select-text font-weight-bold text-success">
                {{ tilesData.sys?.wanIpv6 || 'N/A' }}
              </span>
            </div>
            <div class="info-item">
              <span class="label">S-UI 版本</span>
              <span class="value font-weight-bold text-primary">v{{ tilesData.sys?.appVersion || '1.0.0' }}</span>
            </div>
            <div class="info-item">
              <span class="label">运行时间</span>
              <span class="value select-text">{{ formattedUptime }}</span>
            </div>
          </div>
        </v-card>

        <!-- 二、应用与进程状态卡片 -->
        <v-card class="panel-card px-6 py-5" variant="flat">
          <div class="card-section-title mb-4">服务状态</div>
          
          <!-- S-UI Panel 服务 -->
          <div class="service-item mb-5 pb-5">
            <div class="d-flex align-center justify-space-between mb-3">
              <span class="service-name font-weight-bold">S-UI Panel</span>
              <v-chip density="compact" color="success" variant="flat" size="small" class="px-2">运行中</v-chip>
            </div>
            <div class="service-metrics">
              <div class="metric">
                <span class="label">内存占用</span>
                <span class="value">{{ formattedAppMem }}</span>
              </div>
              <div class="metric">
                <span class="label">协程数量</span>
                <span class="value">{{ tilesData.sys?.appThreads || '-' }}</span>
              </div>
            </div>
            <div class="service-actions mt-3 d-flex justify-end">
              <v-btn variant="text" size="small" density="comfortable" color="primary" class="px-2 font-weight-bold" @click="openLogs('s-ui')">
                <v-icon icon="mdi-list-box-outline" class="mr-1" />查看日志
              </v-btn>
            </div>
          </div>

          <!-- Sing-Box Core 服务 -->
          <div class="service-item">
            <div class="d-flex align-center justify-space-between mb-3">
              <span class="service-name font-weight-bold">Sing-Box Core</span>
              <v-chip density="compact" :color="tilesData.sbd?.running ? 'success' : 'error'" variant="flat" size="small" class="px-2">
                {{ tilesData.sbd?.running ? '运行中' : '已停止' }}
              </v-chip>
            </div>
            <template v-if="tilesData.sbd?.running">
              <div class="service-metrics">
                <div class="metric">
                  <span class="label">内存占用</span>
                  <span class="value">{{ formattedSbdMem }}</span>
                </div>
                <div class="metric">
                  <span class="label">协程数量</span>
                  <span class="value">{{ tilesData.sbd?.stats?.NumGoroutine || '-' }}</span>
                </div>
                <div class="metric">
                  <span class="label">运行时间</span>
                  <span class="value select-text">{{ formattedSbdUptime }}</span>
                </div>
              </div>
            </template>
            <div class="service-actions mt-3 d-flex align-center justify-end" style="gap: 8px;">
              <!-- 启动 / 停止 状态切换按钮 -->
              <v-btn v-if="tilesData.sbd?.running" variant="text" size="small" density="comfortable" color="error" class="px-2 font-weight-bold" @click="stopSingbox" :loading="loadingSingbox" :disabled="loadingSingbox">
                <v-icon icon="mdi-stop" class="mr-1" />停止
              </v-btn>
              <v-btn v-else variant="text" size="small" density="comfortable" color="success" class="px-2 font-weight-bold" @click="restartSingbox" :loading="loadingSingbox" :disabled="loadingSingbox">
                <v-icon icon="mdi-play" class="mr-1" />启动
              </v-btn>

              <!-- 重启按钮 始终常驻 -->
              <v-btn variant="text" size="small" density="comfortable" color="warning" class="px-2 font-weight-bold" @click="restartSingbox" :loading="loadingSingbox" :disabled="loadingSingbox">
                <v-icon icon="mdi-refresh" class="mr-1" />重启
              </v-btn>

              <!-- 查看日志 始终常驻 -->
              <v-btn variant="text" size="small" density="comfortable" color="primary" class="px-2 font-weight-bold" @click="openLogs('sing-box')">
                <v-icon icon="mdi-list-box-outline" class="mr-1" />查看日志
              </v-btn>
            </div>
          </div>

        </v-card>

      </v-col>
    </v-row>
  </v-container>
</template>

<script lang="ts" setup>
import HttpUtils from '@/plugins/httputil'
import { HumanReadable } from '@/plugins/utils'
import Data from '@/store/modules/data'
import Gauge from '@/components/tiles/Gauge.vue'
import History from '@/components/tiles/History.vue'
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import LogVue from '@/layouts/modals/Logs.vue'

// 监控图表类型：'net' (流量), 'pnet' (封包)
const chartToggle = ref('net')
const activeChartType = computed(() => {
  return chartToggle.value === 'net' ? 'h-net' : 'hp-net'
})

// 业务概览数据绑定
const inboundCount = computed(() => Data().config?.inbounds?.length || 0)
const clientCount = computed(() => Data().clients?.length || 0)
const onlineCount = computed(() => Data().onlines?.user?.length || 0)
const ruleCount = computed(() => Data().config?.route?.rules?.length || 0)

// 实时的系统 API 轮询数据
const loadingSingbox = ref(false)
const tilesData = ref<any>({})
const oldNet = ref<any>(null)

// 速度与流量统计缓存
const currentUploadSpeed = ref('0 B/s')
const currentDownloadSpeed = ref('0 B/s')
const totalSent = ref('0 B')
const totalRecv = ref('0 B')

// Uptime 计算自适应
const formattedUptime = computed(() => {
  return tilesData.value.uptime ? HumanReadable.formatSecond(tilesData.value.uptime) : '-'
})
const formattedSbdUptime = computed(() => {
  return tilesData.value.sbd?.stats?.Uptime ? HumanReadable.formatSecond(tilesData.value.sbd?.stats?.Uptime) : '-'
})

// 内存计算自适应
const formattedAppMem = computed(() => {
  return tilesData.value.sys?.appMem ? HumanReadable.sizeFormat(tilesData.value.sys?.appMem) : '-'
})
const formattedSbdMem = computed(() => {
  return tilesData.value.sbd?.stats?.Alloc ? HumanReadable.sizeFormat(tilesData.value.sbd?.stats?.Alloc) : '-'
})

// 监听状态数据以动态更新流量统计和上/下行速度
watch(tilesData, (newVal) => {
  if (!newVal || !newVal.net) return

  // 计算总量
  totalSent.value = HumanReadable.sizeFormat(newVal.net.sent || 0)
  totalRecv.value = HumanReadable.sizeFormat(newVal.net.recv || 0)

  // 计算秒级速度
  if (oldNet.value && oldNet.value.sent !== undefined) {
    const up = Math.max(0, (newVal.net.sent - oldNet.value.sent) / 2) // s-ui 轮询为每 2 秒一次，除以 2 获得真正的秒速
    const down = Math.max(0, (newVal.net.recv - oldNet.value.recv) / 2)
    
    currentUploadSpeed.value = `${HumanReadable.sizeFormat(up)}/s`
    currentDownloadSpeed.value = `${HumanReadable.sizeFormat(down)}/s`
  }
  oldNet.value = newVal.net
}, { deep: true })

// 核心数据定时拉取
const reloadData = async () => {
  const data = await HttpUtils.get('api/status', { r: 'cpu,mem,net,sys,sbd,disk' })
  if (data.success && data.obj) {
    tilesData.value = data.obj
  }
}

let intervalId: any = null

const startTimer = () => {
  intervalId = setInterval(() => {
    reloadData()
  }, 2000)
}

const stopTimer = () => {
  if (intervalId) {
    clearInterval(intervalId)
    intervalId = null
  }
}

onMounted(async () => {
  // 必须确保 loadData 被调用以得到节点与用户数量
  await Data().loadData()
  reloadData()
  startTimer()
})

onBeforeUnmount(() => {
  stopTimer()
})

// 日志确认窗口状态管理
const logModal = ref({
  visible: false,
  logType: 's-ui'
})

const openLogs = (logType: string) => {
  logModal.value.logType = logType
  logModal.value.visible = true
}

const closeLogs = () => {
  logModal.value.logType = 's-ui'
  logModal.value.visible = false
}

const restartSingbox = async () => {
  loadingSingbox.value = true
  try {
    const msg = await HttpUtils.post('api/restartSingbox', {})
    if (msg.success) {
      await reloadData()
    }
  } catch (err) {
    console.error('Failed to restart Sing-Box:', err)
  } finally {
    loadingSingbox.value = false
  }
}

const stopSingbox = async () => {
  loadingSingbox.value = true
  try {
    const msg = await HttpUtils.post('api/stopSingbox', {})
    if (msg.success) {
      await reloadData()
    }
  } catch (err) {
    console.error('Failed to stop Sing-Box:', err)
  } finally {
    loadingSingbox.value = false
  }
}
</script>

<style scoped>
.panel-layout-container {
  max-width: 100%;
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
}

/* 一致的标题样式 */
.card-section-title {
  font-size: 15px;
  font-weight: 700;
  letter-spacing: 0.2px;
}

/* 概览统计数值与副标题 */
.stat-col {
  border-right: 1px solid var(--panel-border-color);
  padding: 12px 0;
}
.stat-col.no-border {
  border-right: none;
}
.stat-num {
  font-size: 26px;
  font-weight: 700;
  line-height: 1.2;
}
.stat-label {
  font-size: 12px;
  margin-top: 6px;
  opacity: 0.6;
}

/* 监控指标徽章样式 (1Panel 精准复刻) */
.monitor-badge {
  display: flex;
  flex-direction: column;
  padding: 10px 16px;
  border-radius: 8px;
  border: 1px solid var(--panel-border-color);
  transition: all 0.2s ease;
}
.monitor-badge .badge-label {
  font-size: 11px;
  opacity: 0.6;
  margin-bottom: 4px;
}
.monitor-badge .badge-value {
  font-size: 14px;
  font-weight: 700;
}

/* 系统信息键值对列表 */
.info-list-wrapper {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.info-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 12px;
  padding-bottom: 10px;
  border-bottom: 1px dashed var(--panel-border-color);
}
.info-item:last-child {
  border-bottom: none;
  padding-bottom: 0;
}
.info-item .label {
  opacity: 0.6;
}
.info-item .value {
  text-align: right;
  max-width: 65%;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* IP Chip 图标 */
.ip-chip-primary, .ip-chip-success {
  display: inline-block;
  font-size: 10px;
  padding: 1px 5px;
  border-radius: 4px;
  border: 1px solid currentColor;
}
.ip-chip-primary {
  color: var(--gauge-progress-color);
}
.ip-chip-success {
  color: rgb(var(--v-theme-success));
}

/* 服务监控列表 */
.service-item {
  border-bottom: 1px solid var(--panel-border-color);
}
.service-item:last-child {
  border-bottom: none;
}
.service-metrics {
  display: flex;
  flex-direction: column;
  gap: 8px;
  font-size: 12px;
}
.service-metrics .metric {
  display: flex;
  justify-content: space-between;
}
.service-metrics .metric .label {
  opacity: 0.6;
}

/* 图表外框自适应高度 */
.chart-container-wrapper {
  position: relative;
  width: 100%;
  height: 200px;
}

/* 按钮切换状态美化 */
.btn-toggle-active {
  background-color: var(--gauge-progress-color) !important;
  color: #ffffff !important;
  border-color: var(--gauge-progress-color) !important;
}

/* 主题颜色自适应 */
.v-theme--light {
  .card-section-title { color: #1f2937; }
  .stat-num { color: #1f2937; }
  .color-online { color: #0052d9 !important; }
  .stat-label { color: #4b5563; }
  .monitor-badge {
    background-color: #f8fafc;
    .badge-label { color: #4b5563; }
    .badge-value { color: #1f2937; }
  }
  .info-item {
    .label { color: #4b5563; }
    .value { color: #1f2937; }
  }
  .service-name { color: #1f2937; }
  .service-metrics .metric {
    .label { color: #4b5563; }
    .value { color: #1f2937; }
  }
}

.v-theme--dark {
  .card-section-title { color: #f3f4f6; }
  .stat-num { color: #f3f4f6; }
  .color-online { color: #3b82f6 !important; }
  .stat-label { color: #9ca3af; }
  .monitor-badge {
    background-color: #111827;
    .badge-label { color: #9ca3af; }
    .badge-value { color: #f3f4f6; }
  }
  .info-item {
    .label { color: #9ca3af; }
    .value { color: #f3f4f6; }
  }
  .service-name { color: #f3f4f6; }
  .service-metrics .metric {
    .label { color: #9ca3af; }
    .value { color: #f3f4f6; }
  }
}
</style>
