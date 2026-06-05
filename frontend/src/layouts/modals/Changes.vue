<template>
  <v-dialog transition="dialog-bottom-transition" width="90%" max-width="800" :loading="loading">
    <v-card class="panel-modal pa-4" style="border-radius: 12px;">
      <v-card-title class="px-2 pb-2">
        <v-row align="center">
          <v-col class="text-h6 font-weight-bold text-grey-lighten-3 d-flex align-center">
            <v-icon color="cyan" class="mr-2">mdi-history</v-icon>
            {{ $t('admin.changes') }}
          </v-col>
          <v-spacer></v-spacer>
          <v-col cols="auto">
            <v-btn icon="mdi-close" variant="text" size="small" color="grey" @click="$emit('close')"></v-btn>
          </v-col>
        </v-row>
      </v-card-title>
      <v-divider class="mb-4" style="opacity: 0.1;"></v-divider>
      
      <v-card-text class="pa-2">
        <!-- 筛选栏 -->
        <v-row class="mb-4" style="row-gap: 16px;">
          <v-col cols="12" sm="4" md="3" class="py-1">
            <v-select
              hide-details
              :label="$t('admin.actor')"
              :items="actorItems"
              v-model="user"
              @update:model-value="loadData"
              variant="outlined"
              class="dark-input"
              density="compact"
            ></v-select>
          </v-col>
          <v-col cols="12" sm="4" md="3" class="py-1">
            <v-select
              hide-details
              :label="$t('admin.key')"
              :items="keyItems"
              v-model="key"
              @update:model-value="loadData"
              variant="outlined"
              class="dark-input"
              density="compact"
            ></v-select>
          </v-col>
          <v-col cols="6" sm="4" md="3" class="py-1">
            <v-select
              hide-details
              :label="$t('count')"
              :items="[10,20,30,50,100]"
              v-model.number="chngCount"
              @update:model-value="loadData"
              variant="outlined"
              class="dark-input"
              density="compact"
            ></v-select>
          </v-col>
          <v-col cols="auto" class="d-flex align-center justify-center py-1">
            <v-btn
              icon="mdi-refresh"
              class="tech-blue-btn"
              :loading="loading"
              @click="loadData"
              style="width: 40px; height: 40px; border-radius: 6px;"
            ></v-btn>
          </v-col>
        </v-row>

        <!-- 数据表格 -->
        <div class="overflow-x-auto flat-card pa-2" style="border-radius: 8px;">
          <v-data-table
            :headers="changesHeaders"
            :items="changes"
            item-value="id"
            density="compact"
            show-expand
            items-per-page="10"
            style="background: transparent;"
            class="w-100"
          >
            <template v-slot:item.dateTime="{ value }">
              <span class="text-caption font-mono text-grey-lighten-2">
                {{ dateFormatted(value) }}
              </span>
            </template>
            
            <template v-slot:item.action="{ value }">
              <span 
                class="text-caption px-2 py-0.5 rounded font-weight-bold"
                :style="{
                  backgroundColor: value === 'add' ? 'rgba(16, 185, 129, 0.15)' : value === 'del' ? 'rgba(239, 68, 68, 0.15)' : 'rgba(59, 130, 246, 0.15)',
                  color: value === 'add' ? '#10b981' : value === 'del' ? '#ef4444' : '#3b82f6',
                  border: '1px solid ' + (value === 'add' ? 'rgba(16, 185, 129, 0.3)' : value === 'del' ? 'rgba(239, 68, 68, 0.3)' : 'rgba(59, 130, 246, 0.3)')
                }"
              >
                {{ $t('actions.' + value) }}
              </span>
            </template>

            <template v-slot:expanded-row="{ columns, item }">
              <tr>
                <td :colspan="columns.length" class="pa-4" style="background: rgba(0, 0, 0, 0.2);">
                  <div v-if="item.index > 0" class="text-caption text-cyan mb-2 font-mono">
                    Index: {{ item.index }}
                  </div>
                  <div class="pa-3 rounded font-mono text-caption text-grey-lighten-2" style="background: #0f172a; border: 1px solid rgba(255, 255, 255, 0.05); white-space: pre-wrap; max-height: 250px; overflow-y: auto;">
                    {{ item.obj }}
                  </div>
                </td>
              </tr>
            </template>
          </v-data-table>
        </div>
      </v-card-text>
    </v-card>
  </v-dialog>
</template>

<script lang="ts">
import { i18n } from '@/locales'
import HttpUtils from '@/plugins/httputil'

export default {
  props: ['admins', 'actor', 'visible'],
  data() {
    return {
      loading: false,
      changes: <any[]>[],
      user: '',
      key: '',
      chngCount: 10,
      expanded: [],
      changesHeaders: [
        { title: 'ID', key: 'id' },
        { title: i18n.global.t('admin.date') + '-' + i18n.global.t('admin.time'), key: 'dateTime' },
        { title: i18n.global.t('admin.actor'), key: 'Actor' },
        { title: i18n.global.t('admin.key'), key: 'key' },
        { title: i18n.global.t('admin.action'), key: 'action' },
      ],
    }
  },
  methods: {
    async loadData() {
      this.loading = true
      const data = await HttpUtils.get('api/changes',{ a: this.user, k: this.key, c: this.chngCount })
      if (data.success) {
        this.changes = data.obj?? []
        this.loading = false
      }
    },
    dateFormatted(dt: number): string {
      const date = new Date(dt*1000)
      return date.toLocaleString(this.locale)
    },
  },
  computed: {
    locale() {
      const l = i18n.global.locale.value
      switch (l) {
        case "zhHans":
          return "zh-cn"
        case "zhHant":
          return "zh-tw"
        default:
          return l
      }
    },
    actorItems() {
      return [
        { title: this.$t('all'), value: '' },
        { title: 'DepleteJob', value: 'DepleteJob' },
        ...this.$props.admins.map((admin: string) => ({ title: admin, value: admin }))
      ]
    },
    keyItems() {
      return [
        { title: this.$t('all'), value: '' },
        { title: this.$t('objects.inbound') || 'Inbounds', value: 'inData' }, // 对应后端 inData
        { title: this.$t('objects.client') || 'Clients', value: 'clients' },   // 对应后端 clients
        { title: 'TLS', value: 'tls' },
        { title: this.$t('objects.rule') || 'Rules', value: 'config' },       // 对应后端 config
        { title: this.$t('setting.setting') || 'Settings', value: 'settings' } // 对应后端 settings
      ]
    }
  },
  watch: {
    visible(newValue) {
      this.changes = []
      this.user = this.$props.actor
      this.key = ''
      this.chngCount = 10
      if (newValue) {
        this.loadData()
      }
    },
  },
}
</script>
