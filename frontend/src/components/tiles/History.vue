<template>
  <div class="panel-chart-wrapper">
    <Line v-if="loaded" :data="chartData" :options="<any>options" />
  </div>
</template>

<script lang="ts">
import { ref } from 'vue'
import { Line } from 'vue-chartjs'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Filler,
} from 'chart.js'
import { HumanReadable } from '@/plugins/utils'

ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Filler
)

ChartJS.defaults.font.family = '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif'

export default {
  components: {
    Line
  },
  props: ['tilesData', 'type'],
  data() {
    return {
      loaded: false,
      labels: new Array(20).fill(''),
      oldValues: <any>{},
      chartData: ref(<any>{})
    }
  },
  computed: {
    options() {
      const isLight = (localStorage.getItem('theme') ?? 'light') === 'light'
      const gridColor = isLight ? '#f1f5f9' : 'rgba(255, 255, 255, 0.05)'
      const textColor = isLight ? '#9ca3af' : '#6b7280'

      const baseOptions = {
        animation: false,
        responsive: true,
        maintainAspectRatio: false,
        elements: {
          point: {
            radius: 0,
            hoverRadius: 4,
            hitRadius: 6
          },
          line: {
            tension: 0.4 // 平滑波形
          }
        },
        interaction: {
          intersect: false,
          mode: 'index',
        },
        plugins: {
          tooltip: {
            enabled: true,
            mode: 'index',
            intersect: false,
            backgroundColor: isLight ? '#ffffff' : '#1f2937',
            titleColor: isLight ? '#1f2937' : '#f3f4f6',
            bodyColor: isLight ? '#4b5563' : '#d1d5db',
            borderColor: isLight ? '#e2e8f0' : '#374151',
            borderWidth: 1,
            padding: 8,
            boxPadding: 4,
            usePointStyle: true
          },
          legend: {
            display: false, // 隐藏多余 Legend，配合 1Panel 顶部徽章
          }
        },
        scales: {
          x: {
            grid: {
              display: false
            },
            ticks: {
              display: false
            }
          },
          y: {
            grid: {
              color: gridColor,
              drawBorder: false
            },
            beginAtZero: true,
            ticks: {
              color: textColor,
              font: {
                size: 10
              },
              maxTicksLimit: 5,
              callback: (label: any) => {
                if (this.$props.type === 'h-net') {
                  return label == 0 ? "0 B" : HumanReadable.sizeFormat(label, 0)
                }
                if (this.$props.type === 'hp-net') {
                  return label == 0 ? "0 P" : HumanReadable.packetFormat(label, 0)
                }
                return Math.round(label) + "%"
              }
            }
          }
        }
      }

      return baseOptions
    }
  },
  methods: {
    updateData1(value1: number) {
      const newData = <number[]>[]
      if (this.chartData.datasets && this.chartData.datasets[0]) {
        newData.push(...this.chartData.datasets[0].data, value1)
      } else {
        newData.push(value1)
      }
      if (newData.length > 20) newData.shift()

      // 1Panel 科技蓝
      this.chartData = {
        labels: this.labels,
        datasets: [
          {
            label: '使用率',
            backgroundColor: 'rgba(0, 82, 217, 0.06)',
            borderColor: '#0052d9',
            borderWidth: 1.5,
            fill: true,
            data: newData
          }
        ],
      }
      this.loaded = true
    },
    updateData2(value1: number, value2: number) {
      const newData1 = <number[]>[]
      const newData2 = <number[]>[]
      if (this.chartData.datasets && this.chartData.datasets[0] && this.chartData.datasets[1]) {
        newData1.push(...this.chartData.datasets[0].data, value1)
        newData2.push(...this.chartData.datasets[1].data, value2)
      } else {
        newData1.push(value1)
        newData2.push(value2)
      }
      if (newData1.length > 20) newData1.shift()
      if (newData2.length > 20) newData2.shift()

      // 上行 (Sent) 蓝色，下行 (Recv) 绿色
      this.chartData = {
        labels: this.labels,
        datasets: [
          {
            label: '上行 (Sent)',
            backgroundColor: 'rgba(0, 82, 217, 0.05)',
            borderColor: '#0052d9',
            borderWidth: 1.5,
            fill: true,
            data: newData1
          },
          {
            label: '下行 (Recv)',
            backgroundColor: 'rgba(43, 182, 115, 0.05)',
            borderColor: '#2bb673',
            borderWidth: 1.5,
            fill: true,
            data: newData2
          }
        ],
      }
      this.loaded = true
    }
  },
  watch: {
    tilesData(v: any) {
      if (!v) return
      switch (this.$props.type) {
        case 'h-cpu':
          if (v.cpu !== undefined) this.updateData1(v.cpu)
          break
        case 'h-mem':
          if (v.mem && v.mem.total) {
            this.updateData1((v.mem.current * 100) / v.mem.total)
          }
          break
        case 'h-net':
          if (v.net) {
            if (this.oldValues.sent !== undefined) {
              const downSpeed = Math.max(0, (v.net.recv - this.oldValues.recv) / 2) // 每两秒
              const upSpeed = Math.max(0, (v.net.sent - this.oldValues.sent) / 2)
              this.updateData2(upSpeed, downSpeed)
            }
            this.oldValues = v.net
          }
          break
        case 'hp-net':
          if (v.net) {
            if (this.oldValues.psent !== undefined) {
              const downSpeed = Math.max(0, (v.net.precv - this.oldValues.precv) / 2)
              const upSpeed = Math.max(0, (v.net.psent - this.oldValues.psent) / 2)
              this.updateData2(upSpeed, downSpeed)
            }
            this.oldValues = v.net
          }
          break
      }
    }
  }
}
</script>

<style scoped>
.panel-chart-wrapper {
  position: relative;
  width: 100%;
  height: 100%;
  min-height: 160px;
}
</style>