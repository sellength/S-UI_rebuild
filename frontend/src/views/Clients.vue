<template>
  <ClientModal 
    v-model="modal.visible"
    :visible="modal.visible"
    :index="modal.index"
    :data="modal.data"
    :stats="modal.stats"
    :inboundTags="inboundTags"
    @close="closeModal"
    @save="saveModal"
  />
  <QrCode
    v-model="qrcode.visible"
    :visible="qrcode.visible"
    :index="qrcode.index"
    @close="closeQrCode"
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
          <v-icon color="primary" class="mr-2">mdi-account-group</v-icon>
          {{ $t('pages.clients') || 'Users' }}
        </div>
        <div class="d-flex flex-wrap align-center flex-grow-1 flex-sm-grow-0" style="gap: 12px; max-width: 600px; width: 100%;">
          <v-text-field
            v-model="searchQuery"
            prepend-inner-icon="mdi-magnify"
            :placeholder="$t('actions.search') || 'Search...'"
            hide-details
            density="compact"
            variant="outlined"
            class="dark-input flex-grow-1"
            style="min-width: 180px;"
          ></v-text-field>
          <v-select
            v-model="filter"
            :label="$t('filter')"
            :items="filterItems"
            hide-details
            density="compact"
            variant="outlined"
            class="dark-input"
            style="max-width: 160px; min-width: 140px;"
          ></v-select>
          <v-btn class="tech-blue-btn text-none" prepend-icon="mdi-plus" @click="showModal(-1)" style="height: 40px; border-radius: 6px;">
            {{ $t('actions.add') }}
          </v-btn>
        </div>
      </div>
    </v-col>
  </v-row>

  <!-- 工业级高对比度 Data Table -->
  <v-row>
    <v-col cols="12">
      <div class="flat-card pa-4 overflow-x-auto" style="border-radius: 8px;">
        <v-table class="w-100" style="background: transparent;">
          <thead>
            <tr style="border-bottom: 2px solid var(--panel-border-color);">
              <th class="text-left text-grey text-subtitle-2 font-weight-bold py-3" style="background: transparent;">{{ $t('client.name') || 'Username' }}</th>
              <th class="text-left text-grey text-subtitle-2 font-weight-bold py-3" style="background: transparent;">{{ $t('pages.inbounds') || 'Inbounds' }}</th>
              <th class="text-left text-grey text-subtitle-2 font-weight-bold py-3" style="background: transparent;">{{ $t('stats.usage') || 'Usage' }}</th>
              <th class="text-left text-grey text-subtitle-2 font-weight-bold py-3" style="background: transparent;">{{ $t('date.expiry') || 'Expiry' }}</th>
              <th class="text-left text-grey text-subtitle-2 font-weight-bold py-3" style="background: transparent; width: 100px;">{{ $t('status') || 'Status' }}</th>
              <th class="text-right text-grey text-subtitle-2 font-weight-bold py-3" style="background: transparent; width: 240px;">{{ $t('actions') || 'Actions' }}</th>
            </tr>
          </thead>
          <tbody>
            <tr 
              v-for="item in filteredClients" 
              :key="item.id" 
              style="border-bottom: 1px solid var(--panel-border-color); transition: background-color 0.2s;"
            >
              <!-- 1. 用户名 -->
              <td class="py-3" style="background: transparent;">
                <div class="d-flex align-center">
                  <span 
                    v-if="onlines[getRealIndex(item)]"
                    class="d-inline-block rounded-circle bg-success mr-2" 
                    style="width: 8px; height: 8px; box-shadow: 0 0 8px #10b981; flex-shrink: 0; animation: pulse 2s infinite;"
                  ></span>
                  <span 
                    v-else
                    class="d-inline-block rounded-circle bg-grey mr-2" 
                    style="width: 8px; height: 8px; flex-shrink: 0;"
                  ></span>
                  <div>
                    <div class="text-body-1 font-weight-bold">{{ item.name }}</div>
                    <div class="text-caption text-grey">{{ item.desc || '-' }}</div>
                  </div>
                </div>
              </td>
              <!-- 2. 绑定的入站节点 -->
              <td class="py-3" style="background: transparent;">
                <div class="d-flex flex-wrap" style="gap: 4px; max-width: 250px;">
                  <v-tooltip activator="parent" location="bottom" v-if="item.inbounds && item.inbounds.length > 0">
                    <span v-for="i in item.inbounds" :key="i">{{ i }}<br /></span>
                  </v-tooltip>
                  <span 
                    v-for="inb in item.inbounds.slice(0, 3)" 
                    :key="inb"
                    class="text-caption px-2 py-0.5 rounded font-weight-medium"
                    style="background-color: var(--panel-border-color); opacity: 0.9;"
                  >
                    {{ inb }}
                  </span>
                  <span v-if="item.inbounds.length > 3" class="text-caption text-grey">
                    +{{ item.inbounds.length - 3 }}
                  </span>
                  <span v-if="item.inbounds.length === 0" class="text-grey">-</span>
                </div>
              </td>
              <!-- 3. 流量已用/限额 -->
              <td class="py-3" style="background: transparent;">
                <div style="max-width: 180px;">
                  <div class="d-flex justify-space-between text-caption mb-1">
                    <span class="font-weight-medium" style="opacity: 0.9;">
                      {{ HumanReadable.sizeFormat(item.up + item.down) }}
                    </span>
                    <span class="text-grey">
                      / {{ item.volume == 0 ? $t('unlimited') : HumanReadable.sizeFormat(item.volume) }}
                    </span>
                  </div>
                  <!-- 精致微型进度条 -->
                  <v-progress-linear
                    v-if="item.volume > 0"
                    :model-value="((item.up + item.down) / item.volume) * 100"
                    color="primary"
                    height="4"
                    rounded
                    class="w-100"
                  ></v-progress-linear>
                  <v-progress-linear
                    v-else
                    :model-value="100"
                    color="grey"
                    height="4"
                    rounded
                    class="w-100"
                    style="opacity: 0.2;"
                  ></v-progress-linear>
                </div>
              </td>
              <!-- 4. 到期时间 -->
              <td class="py-3" style="background: transparent;">
                  <span class="text-body-2 font-mono opacity-90">
                  {{ item.expiry == 0 ? $t('unlimited') : HumanReadable.remainedDays(item.expiry) ?? $t('date.expired') }}
                </span>
              </td>
              <!-- 5. 启用状态 -->
              <td class="py-3" style="background: transparent;">
                <v-switch 
                  color="cyan"
                  v-model="clients[getRealIndex(item)].enable"
                  @update:model-value="buildInboundsUsers(item.inbounds)"
                  hide-details 
                  density="compact"
                ></v-switch>
              </td>
              <!-- 6. 快捷操作 -->
              <td class="py-3 text-right" style="background: transparent;">
                <div class="d-flex justify-end" style="gap: 4px;">
                  <!-- 复制 -->
                  <v-btn icon="mdi-content-copy" variant="text" size="small" color="grey" @click="copySubLink(item)">
                    <v-icon size="18" />
                    <v-tooltip activator="parent" location="top" text="Copy Link"></v-tooltip>
                  </v-btn>
                  <!-- 二维码 -->
                  <v-btn icon="mdi-qrcode" variant="text" size="small" color="cyan" @click="showQrCode(getRealIndex(item))">
                    <v-icon size="18" />
                    <v-tooltip activator="parent" location="top" text="QR-Code"></v-tooltip>
                  </v-btn>
                  <!-- 统计图表 -->
                  <v-btn icon="mdi-chart-line" variant="text" size="small" color="success" @click="showStats(item.name)" v-if="v2rayStats.users.includes(item.name)">
                    <v-icon size="18" />
                    <v-tooltip activator="parent" location="top" :text="$t('stats.graphTitle')"></v-tooltip>
                  </v-btn>
                  <!-- 编辑 -->
                  <v-btn icon="mdi-account-edit" variant="text" size="small" color="primary" @click="showModal(getRealIndex(item))">
                    <v-icon size="18" />
                    <v-tooltip activator="parent" location="top" :text="$t('actions.edit')"></v-tooltip>
                  </v-btn>
                  <!-- 删除 -->
                  <v-btn icon="mdi-account-minus" variant="text" size="small" color="error" @click="delOverlay[getRealIndex(item)] = true">
                    <v-icon size="18" />
                    <v-tooltip activator="parent" location="top" :text="$t('actions.del')"></v-tooltip>
                  </v-btn>
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
                      <v-btn color="error" class="px-4 font-weight-bold" size="small" @click="delClient(getRealIndex(item))">{{ $t('yes') }}</v-btn>
                    </div>
                  </v-card>
                </v-overlay>

              </td>
            </tr>
          </tbody>
        </v-table>
      </div>
    </v-col>
  </v-row>
</template>

<script lang="ts" setup>
import Data from '@/store/modules/data'
import ClientModal from '@/layouts/modals/Client.vue'
import QrCode from '@/layouts/modals/QrCode.vue'
import Stats from '@/layouts/modals/Stats.vue'
import { Client, createClient } from '@/types/clients'
import { computed, ref } from 'vue'
import { Config, V2rayApiStats } from '@/types/config'
import { InTypes, Inbound,InboundWithUser, ShadowTLS, VLESS } from '@/types/inbounds'
import { Link, LinkUtil } from '@/plugins/link'
import { HumanReadable } from '@/plugins/utils'
import { i18n } from '@/locales'
import { push } from 'notivue'

const clients = computed((): any[] => {
  return Data().clients
})

const onlines = computed(() => {
  return Data().onlines.user ? clients.value.map(c => Data().onlines.user.includes(c.name)) : []
})

const appConfig = computed((): Config => {
  return <Config> Data().config
})

const v2rayStats = computed((): V2rayApiStats => {
  return <V2rayApiStats> appConfig.value.experimental.v2ray_api.stats
})

const inbounds = computed((): Inbound[] => {
  return <Inbound[]> appConfig.value?.inbounds
})

const inboundTags = computed((): string[] => {
  if (!inbounds.value) return []
  return inbounds.value?.filter(i => i.tag != '' && Object.hasOwn(i,'users')).map(i => i.tag)
})

const filter = ref('')
const searchQuery = ref('')

const filterItems = [
  { title: i18n.global.t('none'), value: '' },
  { title: i18n.global.t('disable'), value: 'disable' },
  { title: i18n.global.t('date.expired'), value: 'expired' },
  { title: i18n.global.t('online'), value: 'online' },
]

const checkFilter = (c:any) :boolean => {
  switch (filter.value) {
    case 'disable':
      return !c.enable
    case 'expired':
      return HumanReadable.remainedDays(c.expiry) == null
    case 'online':
      return Data().onlines?.user?.includes(c.name)
    default:
      return true
  }
}

const filteredClients = computed(() => {
  let list = clients.value
  list = list.filter(checkFilter)
  if (searchQuery.value) {
    const q = searchQuery.value.toLowerCase()
    list = list.filter(c => 
      c.name.toLowerCase().includes(q) || 
      c.desc.toLowerCase().includes(q)
    )
  }
  return list
})

const getRealIndex = (item: any) => {
  return clients.value.findIndex(c => c.name === item.name)
}

const modal = ref({
  visible: false,
  index: -1,
  data: '',
  stats: false,
})

const delOverlay = ref(new Array<boolean>(clients.value.length).fill(false))

const copySubLink = (item: any) => {
  if (item.links && item.links.length > 0) {
    const link = item.links[0].uri
    navigator.clipboard.writeText(link).then(() => {
      push.success({ message: i18n.global.t('copy') || 'Copied!' })
    }).catch(() => {
      push.error({ message: 'Copy failed!' })
    })
  } else {
    push.warning({ message: 'No links available' })
  }
}

const showModal = (index: number) => {
  modal.value.index = index
  modal.value.data = index == -1 ? '' : JSON.stringify(clients.value[index])
  modal.value.stats = index == -1 ? false : v2rayStats.value.users.includes(clients.value[index].name)
  modal.value.visible = true
}
const closeModal = () => {
  modal.value.visible = false
}
const saveModal = (data:any, stats:boolean) => {
  // Check duplicate name
  const oldName = modal.value.index != -1 ? clients.value[modal.value.index].name : null
  if (data.name != oldName && clients.value.findIndex(c => c.name == data.name) != -1) {
    push.error({
      message: i18n.global.t('error.dplData') + ': ' + i18n.global.t('client.name')
    })
    return
  }
  if(modal.value.index == -1) {
    clients.value.push(data)
  } else {
    clients.value[modal.value.index] = data
  }

  // Rebuild affected inbounds
  buildInboundsUsers(data.inbounds)

  // Rebuild links
  data.links = updateLinks(data)

  // Set Client Stats
  const sIndex = v2rayStats.value.users.findIndex(i => i == data.name) // Find if new user exists

  if (oldName != data.name) {
    v2rayStats.value.users = v2rayStats.value.users.filter(item => item != oldName)
  }

  if (stats) {
    // Add if dos not exist
    if (data.name.length>0 && sIndex == -1) v2rayStats.value.users.push(data.name)
  } else {
    // Delete if exists
    if (sIndex != -1) v2rayStats.value.users.splice(sIndex,1)
  }

  modal.value.visible = false
}
const buildInboundsUsers = (inboundTags:string[]) => {
    inboundTags.forEach(tag => {
      const inbound_index = inbounds.value.findIndex(i => i.tag == tag)
      if (inbound_index != -1){
        const users = <any>[]
        const newInbound = <InboundWithUser>inbounds.value[inbound_index]
        const inboundClients = clients.value.filter(c => c.enable && c.inbounds.includes(tag))
        inboundClients.forEach(c => {
          // Remove flow in non tls VLESS
          if (newInbound.type == InTypes.VLESS) {
            const vlessInbound = <VLESS>newInbound
            if (!vlessInbound.tls?.enabled || vlessInbound.transport?.type) delete(c.config?.vless?.flow)
          }
          users.push(c.config[newInbound.type])
        })
        newInbound.users = users

        // Exceptions for Naive and ShadowTLSv3
        if (users.length == 0){
          if (newInbound.type == InTypes.Naive) {
            newInbound.users = <any>[{}]
          } else {
            if (newInbound.type == InTypes.ShadowTLS){
              const ssTls = <ShadowTLS>newInbound
              if (ssTls.version == 3) newInbound.users = <any>[{}]
            }
          }
        }

        inbounds.value[inbound_index] = newInbound
      }
    })
}
const updateLinks = (c:Client):Link[] => {
  const clientInbounds = <Inbound[]>inbounds.value.filter(i => c.inbounds.includes(i.tag))
  const newLinks = <Link[]>[]
  clientInbounds.forEach(i =>{
    const tlsConfig = <any>Data().tlsConfigs?.findLast((t:any) => t.inbounds.includes(i.tag))
    const cData = <any>Data().inData?.findLast((d:any) => d.tag == i.tag)
    const addrs = cData ? <any[]>cData.addrs : []
    const uris = LinkUtil.linkGenerator(c.name,i, tlsConfig?.client?? {}, addrs)
    if (uris.length>0){
      uris.forEach(uri => {
        newLinks.push(<Link>{ type: 'local', remark: i.tag, uri: uri })
      })
    }
  })
  let links = c.links && c.links.length>0? c.links : <Link[]>[]
  links = [...newLinks, ...links.filter(l => l.type != 'local')]

  return links
}
const delClient = (clientIndex: number) => {
  const id = clients.value[clientIndex].id
  const oldData = createClient(clients.value[clientIndex])

  // Delete stats if exists and will be orphaned
  const tagCounts = clients.value.filter(i => i.name == oldData.name).length
  const sIndex = v2rayStats.value.users.findIndex(i => i == oldData.name)
  if (tagCounts == 1 && sIndex != -1){
    v2rayStats.value.users.splice(sIndex,1)
  }

  clients.value.splice(clientIndex,1)
  buildInboundsUsers(oldData.inbounds)
  if (id>0) Data().delClient(id)
  delOverlay.value[clientIndex] = false
}

const qrcode = ref({
  visible: false,
  index: 0,
})

const showQrCode = (index: number) => {
  qrcode.value.index = index
  qrcode.value.visible = true
}
const closeQrCode = () => {
  qrcode.value.visible = false
}

const stats = ref({
  visible: false,
  resource: 'user',
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
