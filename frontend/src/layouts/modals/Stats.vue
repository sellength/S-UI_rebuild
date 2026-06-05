<template>
  <v-dialog transition="dialog-bottom-transition" width="800">
    <v-card class="panel-modal pa-4" :loading="loading" style="border-radius: 12px; max-height: 90vh; display: flex; flex-direction: column;">
      <v-card-title class="px-2 pb-2">
        <v-row align="center">
          <v-col class="text-h6 font-weight-bold text-grey-lighten-3 d-flex align-center">
            <v-icon color="cyan" class="mr-2">mdi-chart-timeline-variant-shimmer</v-icon>
            {{ $t('stats.graphTitle') }}
          </v-col>
          <v-spacer></v-spacer>
          <v-col cols="auto">
            <v-btn icon="mdi-close" variant="text" size="small" color="grey" @click="$emit('close')"></v-btn>
          </v-col>
        </v-row>
      </v-card-title>
      <v-divider class="mb-4" style="opacity: 0.1;"></v-divider>
      
      <v-card-text style="padding: 0 20px; overflow-y: auto;" class="flex-grow-1">
        <div class="d-flex align-center justify-center pa-2 mb-4 flat-card" style="border-radius: 6px; font-family: monospace;">
          <span class="text-caption text-grey mr-2">{{ $t('objects.' + resource) }}:</span>
          <span class="text-subtitle-2 font-weight-bold text-cyan">{{ tag }}</span>
        </div>

        <div class="mb-4 d-flex justify-center">
          <v-radio-group v-model="limit" @change="loadData" density="compact" :loading="loading" inline hide-details color="cyan">
            <v-radio v-for="p in periods" :key="p.value" :label="p.title" :value="p.value" class="mr-4 text-caption"></v-radio>
          </v-radio-group>
        </div>

        <v-container id="container" style="height:40vh; position: relative;">
          <v-alert :text="$t('noData')" type="warning" variant="outlined" v-if="alert" style="border-radius: 8px;"></v-alert>
          <Line v-if="loaded" :data="usage" :options="<any>options" />
        </v-container>
      </v-card-text>
    </v-card>
  </v-dialog>
</template>

<script lang="ts">
import { i18n } from '@/locales'
import HttpUtils from '@/plugins/httputil'
import { HumanReadable } from '@/plugins/utils'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  Filler,
} from 'chart.js'
import { ref } from 'vue'
import { Line } from 'vue-chartjs'
ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  Filler
)
ChartJS.defaults.font.family = 'Vazirmatn'
export default {
  components: {
    Line
  },
  props: ['visible','resource','tag'],
  data() {
    return {
      loading: false,
      loaded: false,
      alert: false,
      intervalId: <any>0,
      limit: 1,
      periods: [
        { value: 1, title: i18n.global.n(1) + i18n.global.t('date.h')},
        { value: 6, title: i18n.global.n(6) + i18n.global.t('date.h')},
        { value: 12, title: i18n.global.n(12) + i18n.global.t('date.h')},
        { value: 24, title: i18n.global.n(1) + i18n.global.t('date.d')},
        { value: 48, title: i18n.global.n(2) + i18n.global.t('date.d')},
        { value: 240, title: i18n.global.n(10) + i18n.global.t('date.d')},
        { value: 480, title: i18n.global.n(20) + i18n.global.t('date.d')},
        { value: 720, title: i18n.global.n(30) + i18n.global.t('date.d')},
        { value: 1440, title: i18n.global.n(60) + i18n.global.t('date.d')},
        { value: 2160, title: i18n.global.n(90) + i18n.global.t('date.d')},
      ],
      options: {
        responsive: true,
        maintainAspectRatio: false,
        interaction: {
          intersect: false,
          mode: 'index',
        },
        elements: {
          point: { pointStyle: 'circle', radius: 1, hoverRadius: 5 }
        },
        plugins: {
          legend: {
            labels: {
              color: '#94a3b8',
              font: { size: 11 }
            }
          },
          tooltip: {
            callbacks: {
              text: (ctx:any) => {
                const {axis = 'xy', intersect, mode} = ctx.chart.options.interaction;
                return 'Mode: ' + mode + ', axis: ' + axis + ', intersect: ' + intersect;
              },
              footer: (items:any[]) => {
                return HumanReadable.sizeFormat(items.reduce((acc, c) => acc + c.raw, 0))
              }
            }
          }
        },
        scales: {
          x: {
            grid: {
              color: 'rgba(255, 255, 255, 0.03)',
            },
            ticks: {
              color: '#64748b',
              font: { size: 10 }
            }
          },
          y: {
            grid: {
              color: 'rgba(255, 255, 255, 0.05)',
            },
            beginAtZero: true,
            ticks: {
              color: '#64748b',
              font: { size: 10 },
              callback: function(label:any, index: number) {
                return label == 0 ? 0 : HumanReadable.sizeFormat(label,0)
              },
              count: 10
            }
          }
        }
      },
      usage: ref(<any>{}),
    }
  },
  methods: {
    async loadData() {
      this.loading = true
      const data = await HttpUtils.get('api/stats', { resource: this.resource, tag: this.tag, limit: this.limit })
      if (data.success && data.obj) {
        const obj = <any[]>data.obj
        const l = String(i18n.global.locale) == 'fa' ? "fa-IR" : "en-US"
        const oneStep = this.limit * 3600 * 1000 / 360 // Each 10 sec
        const now = new Date().getTime()
        const steps = <number[]>[]
        for (let i = 360; i >= 0; i--) {
          steps.push(now - (oneStep * i))
        }
        const labels = <string[]>[]
        const uplinkData = <number[]>[]
        const downlinkData = <number[]>[]
        for (let i = 1; i<360; i++) {
          labels.push(this.genLable(steps[i],l))
          let upSum:number
          let downSum:number
          const upTraffics = obj.filter(o => o.direction && o.dateTime*1000 < steps[i] && o.dateTime*1000 > steps[i-1]).map((o:any) => o.traffic)
          upSum = upTraffics.length>0 ? upTraffics.reduce(u => u) : null
          const downTraffics = obj.filter(o => !o.direction && o.dateTime*1000 < steps[i] && o.dateTime*1000 > steps[i-1]).map((o:any) => o.traffic)
          downSum = downTraffics.length>0 ? downTraffics.reduce(d => d) : null
          uplinkData.push(upSum)
          downlinkData.push(downSum)
        }
        this.usage = {
          labels: labels,
          datasets: [
            {
              label: i18n.global.t('stats.upload'),
              backgroundColor: 'rgba(245, 158, 11, 0.15)',
              borderColor: '#f59e0b',
              borderWidth: 1.5,
              fill: true,
              data: uplinkData
            },
            {
              label: i18n.global.t('stats.download'),
              backgroundColor: 'rgba(16, 185, 129, 0.1)',
              borderColor: '#10b981',
              borderWidth: 1.5,
              fill: true,
              data: downlinkData
            }
          ],
        }
        this.loaded = true
        this.alert = false
      } else {
        this.alert = true
        this.loaded = false
      }
      this.loading = false
    },
    genLable(step:number, locale: string) {
      return new Date(step).toLocaleString(locale,{
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
        hour12: false,
      })
    },
  },
  watch: {
    visible(v) {
      if (v) {
        this.limit = 1
        this.loadData()
        this.intervalId = setInterval(() => {
          this.loadData()
        }, 10000)
      } else {
        this.loaded = false
        this.alert = false
        this.usage.labels = []
        if (this.usage.datasets) {
          this.usage.datasets[0].data = []
          this.usage.datasets[1].data = []
        }
        if (this.intervalId && this.intervalId != 0) {
          clearInterval(this.intervalId)
        }
      }
    }
  }
}
</script>
