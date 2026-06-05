<script lang="ts" setup>
import { HumanReadable } from '@/plugins/utils';
import { computed } from 'vue';

const props = defineProps({
  tilesData: {
    type: Object,
    default: () => ({})
  },
  type: String
})

// 计算每个仪表盘的具体数值和文本
const data = computed(() => {
  const d = props.tilesData
  if (!d || (!d.mem && !d.cpu)) {
    return { percent: 0, title: '', subtitle: '', color: 'primary' }
  }

  switch (props.type) {
    case 'g-load': {
      const cpuVal = d.cpu ?? 0
      let statusText = '运行流畅'
      if (cpuVal > 80) statusText = '高负荷'
      else if (cpuVal > 50) statusText = '运行一般'
      return {
        percent: Math.ceil(cpuVal),
        title: '负荷',
        subtitle: statusText,
        color: cpuVal > 80 ? 'error' : cpuVal > 50 ? 'warning' : 'primary'
      }
    }
    case 'g-cpu': {
      const cpuVal = d.cpu ?? 0
      const cores = d.sys?.cpuCount ?? 1
      const activeCores = ((cpuVal * cores) / 100).toFixed(2)
      return {
        percent: Math.ceil(cpuVal),
        title: 'CPU',
        subtitle: `( ${activeCores} / ${cores} ) 核`,
        color: cpuVal > 80 ? 'error' : cpuVal > 50 ? 'warning' : 'primary'
      }
    }
    case 'g-mem': {
      if (!d.mem) return { percent: 0, title: '内存', subtitle: '-', color: 'primary' }
      const current = d.mem.current ?? 0
      const total = d.mem.total ?? 1
      const pct = Math.ceil((current * 100) / total)
      return {
        percent: pct,
        title: '内存',
        subtitle: `${HumanReadable.sizeFormat(current, 1)} / ${HumanReadable.sizeFormat(total, 1)}`,
        color: pct > 85 ? 'error' : pct > 70 ? 'warning' : 'primary'
      }
    }
    case 'g-disk': {
      if (!d.disk) return { percent: 0, title: '磁盘', subtitle: '-', color: 'primary' }
      const current = d.disk.current ?? 0
      const total = d.disk.total ?? 1
      const pct = Math.ceil((current * 100) / total)
      return {
        percent: pct,
        title: '/',
        subtitle: `${HumanReadable.sizeFormat(current, 1)} / ${HumanReadable.sizeFormat(total, 1)}`,
        color: pct > 85 ? 'error' : pct > 70 ? 'warning' : 'primary'
      }
    }
  }
  return { percent: 0, title: '', subtitle: '', color: 'primary' }
})

// 圆形 SVG 的周长：r = 40，C = 2 * PI * r = 251.327
const strokeDasharray = 251.327
const strokeDashoffset = computed(() => {
  const pct = data.value.percent
  const validPct = Math.min(Math.max(pct, 0), 100)
  return strokeDasharray - (validPct / 100) * strokeDasharray
})
</script>

<template>
  <div class="panel-gauge-container">
    <div class="svg-wrapper">
      <svg class="circular-svg" viewBox="0 0 100 100">
        <!-- 轨道环 -->
        <circle
          cx="50"
          cy="50"
          r="40"
          class="track-circle"
        />
        <!-- 进度环 -->
        <circle
          cx="50"
          cy="50"
          r="40"
          class="progress-circle"
          :class="`color-${data.color}`"
          :stroke-dasharray="strokeDasharray"
          :stroke-dashoffset="strokeDashoffset"
        />
      </svg>
      <!-- 中间数值 -->
      <div class="inner-value">
        <span class="num">{{ data.percent }}</span>
        <span class="pct">%</span>
      </div>
    </div>
    <!-- 标题与副标题 -->
    <div class="info-wrapper">
      <div class="gauge-title">{{ data.title }}</div>
      <div class="gauge-subtitle">{{ data.subtitle }}</div>
    </div>
  </div>
</template>

<style scoped>
.panel-gauge-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  width: 100%;
  padding: 8px 0;
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
}

.svg-wrapper {
  position: relative;
  width: 90px;
  height: 90px;
}

.circular-svg {
  width: 100%;
  height: 100%;
  transform: rotate(-90deg);
}

.track-circle {
  fill: none;
  stroke: var(--gauge-track-color);
  stroke-width: 5.5;
}

.progress-circle {
  fill: none;
  stroke-width: 5.5;
  stroke-linecap: round;
  transition: stroke-dashoffset 0.6s cubic-bezier(0.4, 0, 0.2, 1);
}

/* 主题颜色绑定 */
.progress-circle.color-primary {
  stroke: var(--gauge-progress-color);
}
.progress-circle.color-warning {
  stroke: rgb(var(--v-theme-warning));
}
.progress-circle.color-error {
  stroke: rgb(var(--v-theme-error));
}

.inner-value {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-direction: row;
}

.inner-value .num {
  font-size: 18px;
  font-weight: 700;
  color: inherit;
}

.inner-value .pct {
  font-size: 10px;
  font-weight: 500;
  margin-left: 1px;
  opacity: 0.8;
}

.info-wrapper {
  margin-top: 8px;
  text-align: center;
  width: 100%;
}

.gauge-title {
  font-size: 13px;
  font-weight: 600;
  opacity: 0.9;
  margin-bottom: 2px;
}

.gauge-subtitle {
  font-size: 11px;
  opacity: 0.65;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  padding: 0 4px;
}

/* 浅色与深色主题下的文字主色适配 */
.v-theme--light {
  .gauge-title { color: #1f2937; }
  .inner-value { color: #1f2937; }
  .gauge-subtitle { color: #4b5563; }
}
.v-theme--dark {
  .gauge-title { color: #f3f4f6; }
  .inner-value { color: #f3f4f6; }
  .gauge-subtitle { color: #9ca3af; }
}
</style>
