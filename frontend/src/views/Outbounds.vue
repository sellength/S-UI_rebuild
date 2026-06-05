<template>
  <OutboundVue 
    v-model="modal.visible"
    :visible="modal.visible"
    :id="modal.id"
    :stats="modal.stats"
    :data="modal.data"
    :tags="outboundTags"
    @close="closeModal"
    @save="saveModal"
  />
  <Stats
    v-model="stats.visible"
    :visible="stats.visible"
    :resource="stats.resource"
    :tag="stats.tag"
    @close="closeStats"
  />

  <!-- 精致顶部 Action Bar -->
  <v-row class="mb-4">
    <v-col cols="12">
      <div class="d-flex flex-column flex-sm-row align-sm-center justify-space-between pa-4 flat-card" style="gap: 16px;">
        <div class="text-h6 font-weight-bold text-grey-lighten-3 d-flex align-center">
          <v-icon color="primary" class="mr-2">mdi-export</v-icon>
          {{ $t('pages.outbounds') || 'Outbounds' }}
        </div>
        <div class="d-flex align-center flex-grow-1 flex-sm-grow-0" style="gap: 12px; max-width: 450px; width: 100%;">
          <v-text-field
            v-model="searchQuery"
            prepend-inner-icon="mdi-magnify"
            :placeholder="$t('actions.search') || 'Search...'"
            hide-details
            density="compact"
            variant="outlined"
            class="dark-input flex-grow-1"
          ></v-text-field>
          <v-btn class="tech-blue-btn text-none" prepend-icon="mdi-plus" @click="showModal(-1)" style="height: 40px; border-radius: 6px;">
            {{ $t('actions.add') }}
          </v-btn>
        </div>
      </div>
    </v-col>
  </v-row>

  <!-- 出站卡片网格 -->
  <v-row>
    <v-col cols="12" sm="6" md="4" lg="3" v-for="item in filteredOutbounds" :key="item.tag">
      <v-card class="panel-card d-flex flex-column justify-space-between h-100 pa-4">
        
        <!-- 卡片头部：Tag 与 协议 Badge -->
        <div>
          <div class="d-flex align-center justify-space-between mb-3">
            <span class="text-subtitle-1 font-weight-bold text-grey-lighten-4 text-truncate" style="max-width: 70%;" :title="item.tag">
              {{ item.tag }}
            </span>
            <span class="text-caption px-2 py-0.5 rounded font-weight-bold" :style="{ backgroundColor: getProtocolBg(item.type), color: '#ffffff' }">
              {{ item.type.toUpperCase() }}
            </span>
          </div>

          <v-divider class="mb-4" style="opacity: 0.1;"></v-divider>

          <!-- 卡片内容体 -->
          <div class="d-flex flex-column" style="gap: 8px;">
            <div class="d-flex justify-space-between align-center">
              <span class="text-caption text-grey">{{ $t('in.addr') }}</span>
              <span class="text-body-2 text-grey-lighten-2 text-truncate" style="max-width: 60%; font-family: monospace;">{{ item.server || '-' }}</span>
            </div>
            
            <div class="d-flex justify-space-between align-center">
              <span class="text-caption text-grey">{{ $t('in.port') }}</span>
              <span class="text-body-2 font-weight-bold text-cyan" style="font-family: monospace;">{{ item.server_port || '-' }}</span>
            </div>

            <div class="d-flex justify-space-between align-center">
              <span class="text-caption text-grey">{{ $t('objects.tls') }}</span>
              <span class="text-body-2 d-flex align-center">
                <v-icon 
                  size="14" 
                  :color="Object.hasOwn(item,'tls') && item.tls?.enabled ? 'success' : 'grey'" 
                  class="mr-1"
                >
                  {{ Object.hasOwn(item,'tls') && item.tls?.enabled ? 'mdi-shield-check' : 'mdi-shield-off' }}
                </v-icon>
                <span :class="Object.hasOwn(item,'tls') && item.tls?.enabled ? 'text-success font-weight-bold' : 'text-grey'">
                  {{ Object.hasOwn(item,'tls') ? $t(item.tls?.enabled ? 'enable' : 'disable') : '-'  }}
                </span>
              </span>
            </div>

            <div class="d-flex justify-space-between align-center">
              <span class="text-caption text-grey">{{ $t('online') }}</span>
              <span class="d-flex align-center">
                <template v-if="onlines[getRealIndex(item)]">
                  <span class="d-inline-block rounded-circle bg-success mr-2" style="width: 8px; height: 8px; box-shadow: 0 0 8px #10b981; animation: pulse 2s infinite;"></span>
                  <span class="text-caption text-success font-weight-bold">{{ $t('online') }}</span>
                </template>
                <template v-else>
                  <span class="text-body-2 text-grey">-</span>
                </template>
              </span>
            </div>
          </div>
        </div>

        <!-- 卡片底部操作 -->
        <div>
          <v-divider class="my-4" style="opacity: 0.1;"></v-divider>
          <div class="d-flex justify-space-between align-center">
            <div class="d-flex" style="gap: 4px;">
              <v-btn icon="mdi-file-edit-outline" variant="text" size="small" color="primary" @click="showModal(getRealIndex(item))">
                <v-icon size="18" />
                <v-tooltip activator="parent" location="top" :text="$t('actions.edit')"></v-tooltip>
              </v-btn>
              <v-btn icon="mdi-chart-line" variant="text" size="small" color="cyan" @click="showStats(item.tag)" v-if="v2rayStats.outbounds.includes(item.tag)">
                <v-icon size="18" />
                <v-tooltip activator="parent" location="top" :text="$t('stats.graphTitle')"></v-tooltip>
              </v-btn>
            </div>

            <v-btn icon="mdi-delete-outline" variant="text" size="small" color="error" @click="delOverlay[getRealIndex(item)] = true">
              <v-icon size="18" />
              <v-tooltip activator="parent" location="top" :text="$t('actions.del')"></v-tooltip>
            </v-btn>
          </div>
        </div>

        <!-- 删除二次确认的覆盖模态框 -->
        <v-overlay
          v-model="delOverlay[getRealIndex(item)]"
          contained
          class="align-center justify-center"
        >
          <v-card class="panel-modal pa-4" style="max-width: 320px; border-radius: 10px;">
            <div class="text-subtitle-1 font-weight-bold text-grey-lighten-3 mb-2 d-flex align-center">
              <v-icon color="error" class="mr-2">mdi-alert-circle-outline</v-icon>
              {{ $t('actions.del') }}
            </div>
            <v-divider class="mb-4" style="opacity: 0.1;"></v-divider>
            <div class="text-body-2 text-grey-lighten-1 mb-6">
              {{ $t('confirm') }}
            </div>
            <div class="d-flex justify-end" style="gap: 12px;">
              <v-btn class="tech-grey-btn px-4" size="small" @click="delOverlay[getRealIndex(item)] = false">{{ $t('no') }}</v-btn>
              <v-btn color="error" class="px-4 font-weight-bold" size="small" @click="delOutbound(getRealIndex(item))">{{ $t('yes') }}</v-btn>
            </div>
          </v-card>
        </v-overlay>

      </v-card>      
    </v-col>
  </v-row>
</template>

<script lang="ts" setup>
import Data from '@/store/modules/data'
import OutboundVue from '@/layouts/modals/Outbound.vue'
import Stats from '@/layouts/modals/Stats.vue'
import { Config, V2rayApiStats } from '@/types/config'
import { Outbound } from '@/types/outbounds'
import { computed, ref } from 'vue'
import { i18n } from '@/locales'
import { push } from 'notivue'

const appConfig = computed((): Config => {
  return <Config> Data().config
})

const outbounds = computed((): Outbound[] => {
  return <Outbound[]> appConfig.value.outbounds
})

const outboundTags = computed((): string[] => {
  return outbounds.value?.map((o:Outbound) => o.tag)
})

const onlines = computed(() => {
  return Data().onlines.outbound ? outbounds.value.map(i => Data().onlines.outbound.includes(i.tag)) : []
})

const v2rayStats = computed((): V2rayApiStats => {
  return <V2rayApiStats> appConfig.value.experimental?.v2ray_api.stats
})

const modal = ref({
  visible: false,
  id: -1,
  data: '',
  stats: false,
})

let delOverlay = ref(new Array<boolean>)
const searchQuery = ref('')

const filteredOutbounds = computed(() => {
  if (!searchQuery.value) return outbounds.value
  const q = searchQuery.value.toLowerCase()
  return outbounds.value.filter(o => 
    o.tag.toLowerCase().includes(q) || 
    o.type.toLowerCase().includes(q) ||
    (o.server && o.server.toLowerCase().includes(q))
  )
})

const getRealIndex = (item: Outbound) => {
  return outbounds.value.findIndex(o => o.tag === item.tag)
}

const getProtocolBg = (type: string) => {
  const t = type.toLowerCase()
  if (t.includes('vless') || t.includes('vmess')) return '#06b6d4'
  if (t.includes('trojan')) return '#8b5cf6'
  if (t.includes('shadow')) return '#f97316'
  if (t.includes('hysteria') || t.includes('tuic') || t.includes('wireguard')) return '#3c6efa'
  return '#6b7280'
}

const showModal = (id: number) => {
  modal.value.id = id
  modal.value.data = id == -1 ? '' : JSON.stringify(outbounds.value[id])
  modal.value.stats = id == -1 ? false : v2rayStats.value.outbounds.includes(outbounds.value[id].tag)
  modal.value.visible = true
}

const closeModal = () => {
  modal.value.visible = false
}
const saveModal = (data:Outbound, stats: boolean) => {
  // Check duplicate tag
  const oldTag = modal.value.id != -1 ? outbounds.value[modal.value.id].tag : null
  if (data.tag != oldTag && outboundTags.value.includes(data.tag)) {
    push.error({
      message: i18n.global.t('error.dplData') + ': ' + i18n.global.t('objects.tag')
    })
    return
  }
  // New or Edit
  if (modal.value.id == -1) {
    outbounds.value.push(data)
    if (stats && data.tag.length>0) {
      v2rayStats.value.outbounds.push(data.tag)
    }
  } else {
    const sIndex = v2rayStats.value.outbounds.findIndex(i => i == data.tag) // Find if new tag exists

    if (stats) {
      // Add if dos not exist
      if (data.tag.length>0 && sIndex == -1) v2rayStats.value.outbounds.push(data.tag)
    } else {
      // Delete if exists
      if (sIndex != -1) v2rayStats.value.outbounds.splice(sIndex,1)
    }

    outbounds.value[modal.value.id] = data
  }
  modal.value.visible = false
}

const stats = ref({
  visible: false,
  resource: 'outbound',
  tag: '',
})

const delOutbound = (index: number) => {
  const inb = outbounds.value[index]
  outbounds.value.splice(index,1)
  const tag = inb.tag

  // Delete stats if exists and will be orphaned
  const tagCounts = outbounds.value.filter(i => i.tag == inb.tag).length
  const sIndex = v2rayStats.value.outbounds.findIndex(i => i == inb.tag)
  if (tagCounts == 1 && sIndex != -1){
    v2rayStats.value.outbounds.splice(sIndex,1)
  }
  if (index < Data().oldData.config.outbounds.length){
    Data().delOutbound(index)
  }
  delOverlay.value[index] = false
}

const showStats = (tag: string) => {
  stats.value.tag = tag
  stats.value.visible = true
}
const closeStats = () => {
  stats.value.visible = false
}
</script>
