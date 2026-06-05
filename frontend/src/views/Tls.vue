<template>
  <TlsVue 
    v-model="modal.visible"
    :visible="modal.visible"
    :index="modal.index"
    :data="modal.data"
    @close="closeModal"
    @save="saveModal"
  />

  <!-- 精致顶部 Action Bar -->
  <v-row class="mb-4">
    <v-col cols="12">
      <div class="d-flex flex-column flex-sm-row align-sm-center justify-space-between pa-4 flat-card" style="gap: 16px;">
        <div class="text-h6 font-weight-bold text-grey-lighten-3 d-flex align-center">
          <v-icon color="primary" class="mr-2">mdi-shield-lock-outline</v-icon>
          {{ $t('pages.tls') || 'TLS Settings' }}
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

  <!-- TLS 卡片网格 -->
  <v-row>
    <v-col cols="12" sm="6" md="4" lg="3" v-for="item in filteredTlsConfigs" :key="item.id">
      <v-card class="panel-card d-flex flex-column justify-space-between h-100 pa-4">
        <div>
          <!-- 卡片头部：ID 与名称 -->
          <div class="d-flex align-center justify-space-between mb-1">
            <span class="text-subtitle-1 font-weight-bold text-grey-lighten-4 truncate" style="max-width: 80%;" :title="item.name">
              {{ (item.id ? item.id + '. ' : '* ') + item.name }}
            </span>
            <v-icon color="cyan" size="20">mdi-certificate-outline</v-icon>
          </div>
          <div class="text-caption text-grey text-truncate mb-3" style="font-family: monospace;">
            {{ item.server?.server_name?.length > 0 ? item.server.server_name : '-' }}
          </div>

          <v-divider class="mb-4" style="opacity: 0.1;"></v-divider>

          <!-- TLS 参数信息 -->
          <div class="d-flex flex-column mb-4" style="gap: 8px;">
            <div class="d-flex justify-space-between align-center">
              <span class="text-caption text-grey">{{ $t('pages.inbounds') }}</span>
              <span class="text-body-2 font-weight-bold text-cyan">
                <v-tooltip activator="parent" dir="ltr" location="bottom" v-if="item.inbounds?.length > 0">
                  <span v-for="i in item.inbounds" :key="i">{{ i }}<br /></span>
                </v-tooltip>
                {{ item.inbounds?.length || 0 }} inbounds
              </span>
            </div>

            <div class="d-flex justify-space-between align-center">
              <span class="text-caption text-grey">ACME</span>
              <span class="text-body-2 d-flex align-center">
                <v-icon 
                  size="14" 
                  :color="item.server?.acme !== undefined ? 'success' : 'grey'" 
                  class="mr-1"
                >
                  {{ item.server?.acme !== undefined ? 'mdi-checkbox-marked-outline' : 'mdi-checkbox-blank-outline' }}
                </v-icon>
                <span :class="item.server?.acme !== undefined ? 'text-success font-weight-bold' : 'text-grey'">
                  {{ $t(item.server?.acme == undefined ? 'no' : 'yes') }}
                </span>
              </span>
            </div>

            <div class="d-flex justify-space-between align-center">
              <span class="text-caption text-grey">ECH</span>
              <span class="text-body-2 d-flex align-center">
                <v-icon 
                  size="14" 
                  :color="item.server?.ech !== undefined ? 'success' : 'grey'" 
                  class="mr-1"
                >
                  {{ item.server?.ech !== undefined ? 'mdi-checkbox-marked-outline' : 'mdi-checkbox-blank-outline' }}
                </v-icon>
                <span :class="item.server?.ech !== undefined ? 'text-success font-weight-bold' : 'text-grey'">
                  {{ $t(item.server?.ech == undefined ? 'no' : 'yes') }}
                </span>
              </span>
            </div>

            <div class="d-flex justify-space-between align-center">
              <span class="text-caption text-grey">Reality</span>
              <span class="text-body-2 d-flex align-center">
                <v-icon 
                  size="14" 
                  :color="item.server?.reality !== undefined ? 'success' : 'grey'" 
                  class="mr-1"
                >
                  {{ item.server?.reality !== undefined ? 'mdi-checkbox-marked-outline' : 'mdi-checkbox-blank-outline' }}
                </v-icon>
                <span :class="item.server?.reality !== undefined ? 'text-success font-weight-bold' : 'text-grey'">
                  {{ $t(item.server?.reality == undefined ? 'no' : 'yes') }}
                </span>
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
              <v-btn icon="mdi-content-duplicate" variant="text" size="small" color="cyan" @click="clone(getRealIndex(item))">
                <v-icon size="18" />
                <v-tooltip activator="parent" location="top" :text="$t('actions.clone')"></v-tooltip>
              </v-btn>
            </div>

            <v-btn 
              v-if="item.inbounds?.length == 0" 
              icon="mdi-delete-outline" 
              variant="text" 
              size="small" 
              color="error" 
              @click="delOverlay[getRealIndex(item)] = true"
            >
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
              <v-btn color="error" class="px-4 font-weight-bold" size="small" @click="delTls(getRealIndex(item))">{{ $t('yes') }}</v-btn>
            </div>
          </v-card>
        </v-overlay>

      </v-card>
    </v-col>
  </v-row>
</template>

<script lang="ts" setup>
import TlsVue from '@/layouts/modals/Tls.vue'
import Data from '@/store/modules/data'
import { computed, ref } from 'vue'
import { Config } from '@/types/config'
import { Inbound } from '@/types/inbounds'
import { Client } from '@/types/clients'
import { Link, LinkUtil } from '@/plugins/link'

const tlsConfigs = computed((): any[] => {
  return Data().tlsConfigs
})

const inbounds = computed((): any[] => {
  return <any[]>(<Config>Data().config)?.inbounds
})

const clients = computed((): any[] => {
  return <Client[]>Data().clients
})

const modal = ref({
  visible: false,
  index: -1,
  data: '',
})

const delOverlay = ref(new Array<boolean>(tlsConfigs.value.length).fill(false))
const searchQuery = ref('')

const filteredTlsConfigs = computed(() => {
  if (!searchQuery.value) return tlsConfigs.value
  const q = searchQuery.value.toLowerCase()
  return tlsConfigs.value.filter(t => 
    t.name.toLowerCase().includes(q) || 
    (t.server && t.server.server_name && t.server.server_name.toLowerCase().includes(q))
  )
})

const getRealIndex = (item: any) => {
  return tlsConfigs.value.findIndex(t => t.id === item.id && t.name === item.name)
}

const showModal = (index: number) => {
  modal.value.index = index
  modal.value.data = index == -1 ? '{}' : JSON.stringify(tlsConfigs.value[index])
  modal.value.visible = true
}
const clone = (index: number) => {
  let data = JSON.parse(JSON.stringify(tlsConfigs.value[index]))
  data.id = 0
  data.inbounds = []
  while (tlsConfigs.value.findIndex(t => t.name == data.name) != -1){
    data.name += '-copy'
  }
  saveModal(data)
}
const closeModal = () => {
  modal.value.visible = false
}
const saveModal = (data:any) => {
  // New or Edit
  if (modal.value.index == -1) {
    tlsConfigs.value.push(data)
  } else {
    tlsConfigs.value[modal.value.index] = data
    inbounds?.value.filter(i => tlsConfigs.value[modal.value.index].inbounds.includes(i.tag)).forEach(i =>{
      if (i.tls != undefined) i.tls = data.server
      updateLinks(i,data.client)
    })
  }
  modal.value.visible = false
}

const delTls = (index: number) => {
  if (index < Data().oldData.tlsConfigs.length){
    Data().delTls(tlsConfigs.value[index].id)
  }
  tlsConfigs.value.splice(index,1)
  delOverlay.value[index] = false
}

const updateLinks = (i:any,tlsClient:any) => {
  if(i.users && i.users.length>0){
    i.users.forEach((u:any) => {
      const client = clients.value.find(c => u.username? c.name == u.username : c.name == u.name)
      if (client){
        const clientInbounds = <Inbound[]>inbounds.value.filter(inb => client?.inbounds.includes(inb.tag))
        const newLinks = <Link[]>[]
        clientInbounds.forEach(i =>{
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
        links = [...newLinks, ...links.filter((l:Link) => l.type != 'local')]

        client.links = links
      }
    })
  }
}
</script>
