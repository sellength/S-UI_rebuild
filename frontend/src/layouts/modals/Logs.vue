<template>
  <v-dialog transition="dialog-bottom-transition" width="90%" max-width="1200" :loading="loading">
    <v-card class="panel-modal pa-4" style="border-radius: 12px;">
      <v-card-title class="px-2 pb-2">
        <v-row align="center">
          <v-col class="text-h6 font-weight-bold text-grey-lighten-3 d-flex align-center">
            <v-icon color="cyan" class="mr-2">mdi-text-box-search-outline</v-icon>
            {{ $t('basic.log.title') + " - " + (logType == 's-ui'? "S-UI" : "Sing-Box") }}
          </v-col>
          <v-spacer></v-spacer>
          <v-col cols="auto">
            <v-btn icon="mdi-close" variant="text" size="small" color="grey" @click="$emit('close')"></v-btn>
          </v-col>
        </v-row>
      </v-card-title>
      <v-divider class="mb-4" style="opacity: 0.1;"></v-divider>
      
      <v-card-text class="pa-2">
        <v-row align="center" style="row-gap: 16px;">
          <v-col cols="12" sm="5" class="py-1">
            <v-select
              hide-details
              :label="$t('basic.log.level')"
              :items="logLevels"
              v-model="logLevel"
              @update:model-value="loadData"
              variant="outlined"
              class="dark-input"
              density="compact"
            ></v-select>
          </v-col>
          <v-col cols="12" sm="5" class="py-1">
            <v-select
              hide-details
              :label="$t('count')"
              :items="[10,20,30,50,100]"
              v-model.number="logCount"
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
        
        <!-- 极客终端风格黑色卡片框 -->
        <div 
          class="pa-4 rounded mt-4 font-mono text-caption text-grey-lighten-2 overflow-y-auto" 
          style="background: #090d16; border: 1px solid rgba(255, 255, 255, 0.05); max-height: 450px; line-height: 1.6; white-space: pre-wrap; word-break: break-all;"
          dir="ltr"
          v-html="lines.join('<br />')"
        ></div>
      </v-card-text>
    </v-card>
  </v-dialog>
</template>

<script lang="ts">
import HttpUtils from '@/plugins/httputil';

export default {
  props: ['logType', 'visible'],
  data() {
    return {
      loading: false,
      lines: [],
      logLevel: 'info',
      logLevels: [
        { title: 'DEBUG', value: 'debug' },
        { title: 'INFO', value: 'info' },
        { title: 'WARNING', value: 'warning' },
        { title: 'ERROR', value: 'err' },
      ],
      logCount: 10,
    }
  },
  methods: {
    async loadData() {
      this.loading = true
      const data = await HttpUtils.get('api/logs',{ s: this.$props.logType, c: this.logCount, l: this.logLevel })
      if (data.success) {
        this.lines = data.obj?? []
        this.loading = false
      }
    }
  },
  watch: {
    visible(newValue) {
      this.lines = []
      this.logLevel = 'info'
      this.logCount = 10
      if (newValue) {
        this.loadData()
      }
    },
  },
}
</script>
