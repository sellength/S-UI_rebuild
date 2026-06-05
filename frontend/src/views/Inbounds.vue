<template>
  <InboundVue 
    v-model="modal.visible"
    :visible="modal.visible"
    :index="modal.index"
    :stats="modal.stats"
    :data="modal.data"
    :cData="modal.cData"
    :inTags="inTags"
    :outTags="outTags"
    :tlsConfigs="tlsConfigs"
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
        <div class="text-h6 font-weight-bold d-flex align-center">
          <v-icon color="primary" class="mr-2">mdi-router-wireless</v-icon>
          {{ $t('pages.inbounds') || 'Inbounds' }}
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

  <!-- 节点卡片网格 -->
  <v-row>
    <v-col cols="12" sm="6" md="4" lg="3" v-for="item in filteredInbounds" :key="item.tag">
      <v-card class="panel-card d-flex flex-column justify-space-between h-100 pa-4">
        
        <!-- 卡片头部：Tag 与 协议 Badge -->
        <div>
          <div class="d-flex align-center justify-space-between mb-3">
            <span class="text-subtitle-1 font-weight-bold text-truncate" style="max-width: 70%;" :title="item.tag">
              {{ item.tag }}
            </span>
            <span class="text-caption px-2 py-0.5 rounded font-weight-bold" :style="{ backgroundColor: getProtocolBg(item.type), color: '#ffffff' }">
              {{ item.type.toUpperCase() }}
            </span>
          </div>

          <v-divider class="mb-4" style="opacity: 0.1;"></v-divider>

          <!-- 卡片内容体：各种高对比度参数展示 -->
          <div class="d-flex flex-column" style="gap: 8px;">
            <div class="d-flex justify-space-between align-center">
              <span class="text-caption text-grey">{{ $t('in.addr') }}</span>
              <span class="text-body-2 text-truncate" style="max-width: 60%; font-family: monospace; opacity: 0.85;">{{ item.listen || '0.0.0.0' }}</span>
            </div>
            
            <div class="d-flex justify-space-between align-center">
              <span class="text-caption text-grey">{{ $t('in.port') }}</span>
              <span class="text-body-2 font-weight-bold text-cyan" style="font-family: monospace;">{{ item.listen_port }}</span>
            </div>

            <div class="d-flex justify-space-between align-center">
              <span class="text-caption text-grey">{{ $t('objects.tls') }}</span>
              <span class="text-body-2 d-flex align-center">
                <v-icon 
                  size="14" 
                  :color="Object.hasOwn(item,'tls') && (item as any).tls?.enabled ? 'success' : 'grey'" 
                  class="mr-1"
                >
                  {{ Object.hasOwn(item,'tls') && (item as any).tls?.enabled ? 'mdi-shield-check' : 'mdi-shield-off' }}
                </v-icon>
                <span :class="Object.hasOwn(item,'tls') && (item as any).tls?.enabled ? 'text-success font-weight-bold' : 'text-grey'">
                  {{ Object.hasOwn(item,'tls') ? $t((item as any).tls?.enabled ? 'enable' : 'disable') : '-'  }}
                </span>
              </span>
            </div>

            <div class="d-flex justify-space-between align-center">
              <span class="text-caption text-grey">{{ $t('pages.clients') }}</span>
              <span class="text-body-2 opacity-85">
                <v-tooltip activator="parent" location="bottom" v-if="Object.hasOwn(item,'users')">
                  <span v-for="u in findInbounsUsers(item)" :key="u">{{ u }}<br /></span>
                </v-tooltip>
                <v-icon size="14" class="mr-1" color="grey">mdi-account-group</v-icon>
                {{ Array.isArray((item as any).users) ? (item as any).users.length : '-' }}
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

        <!-- 卡片底部操作按钮 -->
        <div>
          <v-divider class="my-4" style="opacity: 0.1;"></v-divider>
          <div class="d-flex justify-space-between align-center">
            <div class="d-flex" style="gap: 4px;">
              <v-btn icon="mdi-file-edit-outline" variant="text" size="small" color="primary" @click="showModal(getRealIndex(item))">
                <v-icon size="18" />
                <v-tooltip activator="parent" location="top" :text="$t('actions.edit')"></v-tooltip>
              </v-btn>
              <v-btn icon="mdi-chart-line" variant="text" size="small" color="cyan" @click="showStats(item.tag)" v-if="v2rayStats.inbounds.includes(item.tag)">
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
            <div class="text-subtitle-1 font-weight-bold mb-2 d-flex align-center">
              <v-icon color="error" class="mr-2">mdi-alert-circle-outline</v-icon>
              {{ $t('actions.del') }}
            </div>
            <v-divider class="mb-4" style="opacity: 0.1;"></v-divider>
            <div class="text-body-2 mb-6" style="opacity: 0.8;">
              {{ $t('confirm') }}
            </div>
            <div class="d-flex justify-end" style="gap: 12px;">
              <v-btn class="tech-grey-btn px-4" size="small" @click="delOverlay[getRealIndex(item)] = false">{{ $t('no') }}</v-btn>
              <v-btn color="error" class="px-4 font-weight-bold" size="small" @click="delInbound(getRealIndex(item))">{{ $t('yes') }}</v-btn>
            </div>
          </v-card>
        </v-overlay>

      </v-card>      
    </v-col>
  </v-row>
</template>

<script lang="ts" setup>
import Data from '@/store/modules/data'
import InboundVue from '@/layouts/modals/Inbound.vue'
import Stats from '@/layouts/modals/Stats.vue'
import { Config, V2rayApiStats } from '@/types/config'
import { computed, ref } from 'vue'
import { InTypes, Inbound, InboundWithUser, ShadowTLS, VLESS } from '@/types/inbounds'
import { Client } from '@/types/clients'
import { Link, LinkUtil } from '@/plugins/link'
import { i18n } from '@/locales'
import { push } from 'notivue'
import { fillData } from '@/plugins/outJson'

const appConfig = computed((): Config => {
  return <Config> Data().config
})

const inbounds = computed((): Inbound[] => {
  return <Inbound[]> appConfig.value.inbounds
})

const tlsConfigs = computed((): any[] => {
  return <any[]> Data().tlsConfigs
})

const inData = computed((): any[] => {
  return <any[]> Data().inData
})

const inTags = computed((): string[] => {
  return inbounds.value?.map(i => i.tag)
})

const outTags = computed((): string[] => {
  return appConfig.value.outbounds?.map(i => i.tag)
})

const clients = computed((): Client[] => {
  return <Client[]> Data().clients
})

const onlines = computed(() => {
  return Data().onlines.inbound ? inbounds.value.map(i => Data().onlines.inbound.includes(i.tag)) : []
})

const v2rayStats = computed((): V2rayApiStats => {
  return <V2rayApiStats> appConfig.value.experimental?.v2ray_api.stats
})

const modal = ref({
  visible: false,
  index: -1,
  data: '',
  cData: '',
  stats: false,
})

let delOverlay = ref(new Array<boolean>)

const searchQuery = ref('')

const filteredInbounds = computed(() => {
  if (!searchQuery.value) return inbounds.value
  const q = searchQuery.value.toLowerCase()
  return inbounds.value.filter(i => 
    i.tag.toLowerCase().includes(q) || 
    i.type.toLowerCase().includes(q) ||
    String(i.listen_port).includes(q)
  )
})

const getRealIndex = (item: Inbound) => {
  return inbounds.value.findIndex(i => i.tag === item.tag)
}

const getProtocolBg = (type: string) => {
  const t = type.toLowerCase()
  if (t.includes('vless') || t.includes('vmess')) return '#06b6d4'
  if (t.includes('trojan')) return '#8b5cf6'
  if (t.includes('shadow')) return '#f97316'
  if (t.includes('hysteria') || t.includes('tuic')) return '#3c6efa'
  return '#6b7280'
}

const showModal = (index: number) => {
  modal.value.index = index
  if (index == -1){
    modal.value.data = ''
    modal.value.cData = ''
    modal.value.stats = false
  } else {
    modal.value.data = JSON.stringify(inbounds.value[index])
    modal.value.stats = v2rayStats.value.inbounds.includes(inbounds.value[index].tag)
    const inDataIndex = inData.value.findIndex(d => d.tag == inbounds.value[index].tag)
    modal.value.cData = inDataIndex == -1 ? '' : JSON.stringify(inData.value[inDataIndex])
  }
  modal.value.visible = true
}
const closeModal = () => {
  modal.value.visible = false
}
const saveModal = (data:Inbound, stats: boolean, tls_id: number, cData: any) => {
  // Check duplicate tag
  const oldTag = modal.value.index != -1 ? inbounds.value[modal.value.index].tag : null
  if (data.tag != oldTag && inTags.value.includes(data.tag)) {
    push.error({
      message: i18n.global.t('error.dplData') + ': ' + i18n.global.t('objects.tag')
    })
    return
  }
  if (cData.id != -1) {
    cData.tag = data.tag
    fillData(cData.outJson, data,tls_id>0 ? tlsConfigs.value.findLast(t => t.id == tls_id).client : {})
  }

  // New or Edit
  if (modal.value.index == -1) {
    inbounds.value.push(data)
    if (stats && data.tag.length>0) {
      v2rayStats.value.inbounds.push(data.tag)
    }
    if (cData.id != -1){
      inData.value.push(cData)
    }
  } else {
    const oldTag = inbounds.value[modal.value.index].tag
    const sIndex = v2rayStats.value.inbounds.findIndex(i => i == data.tag) // Find if new tag exists

    // Update tls preset
    const oldTlsConfigIndex = tlsConfigs?.value.findIndex(t => t.inbounds?.includes(oldTag))
    if (oldTlsConfigIndex != -1){
      tlsConfigs.value[oldTlsConfigIndex].inbounds = tlsConfigs?.value[oldTlsConfigIndex].inbounds.filter((i:string) => i != oldTag)
    }

    if (oldTag != data.tag) {
      v2rayStats.value.inbounds = v2rayStats.value.inbounds.filter(item => item != oldTag)
      changeClientInboundsTag(oldTag,data.tag)
    }

    if (stats) {
      // Add if dos not exist
      if (data.tag.length>0 && sIndex == -1) v2rayStats.value.inbounds.push(data.tag)
    } else {
      // Delete if exists
      if (sIndex != -1) v2rayStats.value.inbounds.splice(sIndex,1)
    }

    inbounds.value[modal.value.index] = data
    const inDataIndex = inData.value.findIndex(indata => indata.tag == oldTag)
    if (cData.id != -1) {
      if (inDataIndex == -1){
        inData.value.push(cData)
      } else {
        inData.value[inDataIndex] = cData
      }
    } else if (inDataIndex != -1) {
      Data().delInData(inData.value[inDataIndex].id)
      inData.value.splice(inDataIndex,1)
    }
  }
  // Update tls preset
  if (tls_id>0) {
    tlsConfigs.value.findLast(t => t.id == tls_id).inbounds.push(data.tag)
    tlsConfigs.value.sort()
  }

  if (Object.hasOwn(data,'users')) {
    // Set users
    data = buildInboundsUsers(data)
    // Update links
    updateLinks(data)
  }
  modal.value.visible = false
}
const updateLinks = (i: any) => {
  if(i.users && i.users.length>0){
    i.users.forEach((u:any) => {
      const client = clients.value.find(c => u.username? c.name == u.username : c.name == u.name)
      if (client){
        const clientInbounds = <Inbound[]>inbounds.value.filter(inb => client?.inbounds.includes(inb.tag))
        const newLinks = <Link[]>[]
        clientInbounds.forEach(i =>{
          const tlsClient = tlsConfigs?.value.findLast((t:any) => t.inbounds.includes(i.tag))?.client?? {}
          const cData = <any>Data().inData?.findLast((d:any) => d.tag == i.tag)
          const addrs = cData ? <any[]>cData.addrs : []
          const uris = LinkUtil.linkGenerator(client.name,i, tlsClient, addrs)
          if (uris.length>0){
            uris.forEach(uri => {
              newLinks.push(<Link>{ type: 'local', remark: i.tag, uri: uri })
            })
          }
        })
        let links = client.links && client.links.length>0? client.links : <Link[]>[]
        links = [...newLinks, ...links.filter(l => l.type != 'local')]

        client.links = links
      }
    })
  }
}
const delInbound = (index: number) => {
  const inb = inbounds.value[index]
  inbounds.value.splice(index,1)
  const tag = inb.tag

  if (Object.hasOwn(inb,'users')) {
    const inbU = <InboundWithUser>inb
    if (inbU.users && inbU.users.length>0){
      inbU.users.forEach((u:any) => {
        const c_index = clients.value.findIndex(c => u.username? u.username == c.name : u.name == c.name)
        if (c_index != -1) {
          clients.value[c_index].inbounds = clients.value[c_index].inbounds.filter((x:string) => x!=tag)
          clients.value[c_index].links = clients.value[c_index].links.filter((x:any) => x.remark!=tag)
        }
      })
    }
  }

  // Delete binded tls if exists
  if (Object.hasOwn(inb,'tls')) {
    const oldTlsConfigIndex = tlsConfigs?.value.findIndex(t => t.inbounds?.includes(inb.tag))
    if (oldTlsConfigIndex != -1){
      tlsConfigs.value[oldTlsConfigIndex].inbounds = tlsConfigs?.value[oldTlsConfigIndex].inbounds.filter((i:string) => i != inb.tag)
    }
  }

  // Delete stats if exists and will be orphaned
  const tagCounts = inbounds.value.filter(i => i.tag == inb.tag).length
  const sIndex = v2rayStats.value.inbounds.findIndex(i => i == inb.tag)
  if (tagCounts == 1 && sIndex != -1){
    v2rayStats.value.inbounds.splice(sIndex,1)
  }
  if (index < Data().oldData.config.inbounds.length){
    Data().delInbound(index)
  }
  delOverlay.value[index] = false
}
const buildInboundsUsers = (inbound:any):Inbound => {
    const users = <any>[]
    const inboundClients = clients.value.filter(c => c.enable && c.inbounds.includes(inbound.tag))
    inboundClients.forEach(c => {
      // Remove flow in non tls VLESS
      if (inbound.type == InTypes.VLESS) {
        const vlessInbound = <VLESS>inbound
        if (!vlessInbound.tls?.enabled || vlessInbound.transport?.type) delete(c.config?.vless?.flow)
      }
      users.push(c.config[inbound.type])
    })
    inbound.users = users

    // Exceptions for Naive and ShadowTLSv3
    if (users.length == 0){
      if (inbound.type == InTypes.Naive){
        inbound.users = <any>[{}]
      } else {
        if (inbound.type == InTypes.ShadowTLS){
          const ssTls = <ShadowTLS>inbound
          if (ssTls.version == 3) inbound.users = <any>[{}]
        }
      }
    }

    return <Inbound>inbound
}
const changeClientInboundsTag = (oldtag: string, newTag:string) => {
  clients.value.forEach((c, c_index) => {
    const inbound_index = c.inbounds.findIndex(i => i == oldtag)
    if (inbound_index != -1) {
      c.inbounds[inbound_index] = newTag
      clients.value[c_index].inbounds = c.inbounds
    }
  })
}
const findInbounsUsers = (inbound: InboundWithUser): string[] => {
  if (inbound.users === null || !Array.isArray(inbound.users) || inbound.users.length == 0) return []

  const users = inbound.users.map(user => 'username' in user ? user.username : user.name)
  return users
}

const stats = ref({
  visible: false,
  resource: 'inbound',
  tag: '',
})

const showStats = (tag: string) => {
  stats.value.tag = tag
  stats.value.visible = true
}
const closeStats = () => {
  stats.value.visible = false
}
</script>
