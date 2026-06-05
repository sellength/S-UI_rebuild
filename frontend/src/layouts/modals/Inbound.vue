<template>
  <v-dialog transition="dialog-bottom-transition" width="800">
    <v-card class="panel-modal pa-4" style="border-radius: 12px; max-height: 90vh; display: flex; flex-direction: column;">
      <v-card-title class="px-2 pb-2">
        <span class="text-h6 font-weight-bold text-grey-lighten-3">
          {{ $t('actions.' + title) + " " + $t('objects.inbound') }}
        </span>
      </v-card-title>
      <v-divider class="mb-4" style="opacity: 0.1;"></v-divider>
      
      <v-card-text style="padding: 0 20px; overflow-y: auto;" class="flex-grow-1">
        <v-container style="padding: 0;">
          <v-row style="row-gap: 16px;" class="mb-4">
            <v-col cols="12" sm="6" md="4" class="py-1">
              <v-select
                hide-details
                :label="$t('type')"
                :items="Object.keys(inTypes).map((key,index) => ({title: key, value: Object.values(inTypes)[index]}))"
                v-model="inbound.type"
                @update:modelValue="changeType"
                variant="outlined"
                class="dark-input"
                density="comfortable"
              ></v-select>
            </v-col>
            <v-col cols="12" sm="6" md="4" class="py-1">
              <v-text-field v-model="inbound.tag" :label="$t('objects.tag')" hide-details variant="outlined" class="dark-input" density="comfortable"></v-text-field>
            </v-col>
          </v-row>

          <v-tabs
            v-if="HasInData.includes(inbound.type)"
            v-model="side"
            density="comfortable"
            style="margin-bottom: 16px;"
          >
            <v-tab value="s" class="text-none font-weight-bold">{{ $t('in.sSide') }}</v-tab>
            <v-tab value="c" class="text-none font-weight-bold">{{ $t('in.cSide') }}</v-tab>
          </v-tabs>

          <v-window v-model="side" class="pt-2">
            <!-- 服务端设置 -->
            <v-window-item value="s">
              <div class="d-flex flex-column" style="gap: 16px;">
                <Listen :inbound="inbound" :inTags="inTags" />
                <Direct v-if="inbound.type == inTypes.Direct" direction="in" :data="inbound" />
                <Shadowsocks v-if="inbound.type == inTypes.Shadowsocks" direction="in" :data="inbound" />
                <Hysteria v-if="inbound.type == inTypes.Hysteria" direction="in" :data="inbound" />
                <Hysteria2 v-if="inbound.type == inTypes.Hysteria2" direction="in" :data="inbound" />
                <Naive v-if="inbound.type == inTypes.Naive" :inbound="inbound" />
                <ShadowTls v-if="inbound.type == inTypes.ShadowTLS" direction="in" :data="inbound" :outTags="outTags" />
                <Tuic v-if="inbound.type == inTypes.TUIC" direction="in" :data="inbound" />
                <TProxy v-if="inbound.type == inTypes.TProxy" :inbound="inbound" />
                <AnyTls v-if="inbound.type == inTypes.AnyTLS" :data="inbound" />
                <Transport v-if="Object.hasOwn(inbound,'transport')" :data="inbound" />
                <Users v-if="HasOptionalUser.includes(inbound.type)" :inbound="inbound" />
                <InTls v-if="Object.hasOwn(inbound,'tls')"  :inbound="inbound" :tlsConfigs="tlsConfigs" :tls_id="tls_id" />
                <Multiplex v-if="Object.hasOwn(inbound,'multiplex')" direction="in" :data="inbound" />
                <v-card class="pb-4">
                  <v-row>
                    <v-col cols="12" sm="6" md="4">
                      <v-switch v-model="inboundStats" color="cyan" :label="$t('stats.enable')" hide-details></v-switch>
                    </v-col>
                  </v-row>
                </v-card>
              </div>
            </v-window-item>

            <!-- 客户端设置 -->
            <v-window-item value="c">
              <div class="d-flex flex-column" style="gap: 16px;">
                <OutJsonVue :inData="inData" :type="inbound.type" />
                
                <div class="flat-card pa-4 mt-2" style="border-radius: 8px;">
                  <div class="text-subtitle-2 text-grey-lighten-2 d-flex align-center justify-space-between mb-2">
                    <span>{{ $t('in.multiDomain') }}</span>
                    <v-btn icon="mdi-plus" variant="text" size="small" color="cyan" @click="add_addr"></v-btn>
                  </div>
                  <v-divider class="mb-4" style="opacity: 0.1;"></v-divider>
                  <div 
                    v-for="(addr, idx) in inData.addrs" 
                    :key="idx" 
                    class="mb-4 pa-4 rounded" 
                    style="background: rgba(255,255,255,0.01); border: 1px solid rgba(255, 255, 255, 0.05); border-radius: 8px;"
                  >
                    <div class="d-flex align-center justify-space-between mb-2">
                      <span class="text-caption text-grey font-weight-bold">{{ $t('in.addr') }} #{{ (idx+1) }}</span>
                      <v-btn icon="mdi-delete-outline" color="error" variant="text" size="small" @click="inData.addrs.splice(idx,1)"></v-btn>
                    </div>
                    <AddrVue :addr="addr" :hasTls="Object.hasOwn(inbound,'tls')" />
                  </div>
                </div>
              </div>
            </v-window-item>
          </v-window>
        </v-container>
      </v-card-text>
      
      <v-card-actions class="px-2 pt-4">
        <v-spacer></v-spacer>
        <v-btn
          class="tech-grey-btn px-4"
          size="comfortable"
          @click="closeModal"
        >
          {{ $t('actions.close') }}
        </v-btn>
        <v-btn
          class="tech-blue-btn px-4"
          size="comfortable"
          :loading="loading"
          @click="saveChanges"
        >
          {{ $t('actions.save') }}
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script lang="ts">
import { InTypes, createInbound } from '@/types/inbounds'
import { Addr, InData } from '@/plugins/inData'
import RandomUtil from '@/plugins/randomUtil'

import Listen from '@/components/Listen.vue'
import Direct from '@/components/protocols/Direct.vue'
import Users from '@/components/Users.vue'
import Shadowsocks from '@/components/protocols/Shadowsocks.vue'
import Hysteria from '@/components/protocols/Hysteria.vue'
import Hysteria2 from '@/components/protocols/Hysteria2.vue'
import Naive from '@/components/protocols/Naive.vue'
import ShadowTls from '@/components/protocols/ShadowTls.vue'
import Tuic from '@/components/protocols/Tuic.vue'
import InTls from '@/components/tls/InTLS.vue'
import TProxy from '@/components/protocols/TProxy.vue'
import Multiplex from '@/components/Multiplex.vue'
import Transport from '@/components/Transport.vue'
import AddrVue from '@/components/Addr.vue'
import OutJsonVue from '@/components/OutJson.vue'
import AnyTls from '@/components/protocols/AnyTls.vue'
export default {
  props: ['visible', 'data', 'cData', 'index', 'stats', 'inTags', 'outTags', 'tlsConfigs'],
  emits: ['close', 'save'],
  data() {
    return {
      inbound: createInbound("direct",{ "tag": "" }),
      inData: <InData>{},
      title: "add",
      loading: false,
      side: "s",
      inTypes: InTypes,
      inboundStats: false,
      tls_id: { value: 0 },
      HasOptionalUser: [InTypes.Mixed,InTypes.SOCKS,InTypes.HTTP,InTypes.Shadowsocks],
      HasInData: [
        InTypes.SOCKS,
        InTypes.HTTP,
        InTypes.Shadowsocks,
        InTypes.VMess,
        InTypes.ShadowTLS,
        InTypes.Trojan,
        InTypes.Hysteria,
        InTypes.VLESS,
        InTypes.TUIC,
        InTypes.Hysteria2,
        InTypes.Naive,
        InTypes.AnyTLS,
      ]
    }
  },
  methods: {
    updateData() {
      if (this.$props.index != -1) {
        const newData = JSON.parse(this.$props.data)
        this.inbound = createInbound(newData.type, newData)
        this.tls_id.value = this.$props.tlsConfigs?.findLast((t:any) => t.inbounds?.includes(this.inbound.tag))?.id?? 0
        if (this.HasInData.includes(this.inbound.type)){
          this.inData = this.$props.cData?.length> 0 ? <InData>JSON.parse(this.$props.cData) : <InData>{id: 0, tag: this.inbound.tag, addrs: [], outJson: {}}
        } else {
          this.inData = <InData>{id: -1}
        }
        this.title = "edit"
      }
      else {
        const port = RandomUtil.randomIntRange(10000, 60000)
        this.inbound = createInbound("direct",{ tag: "direct-"+port ,listen: "::", listen_port: port })
        this.tls_id.value = 0
        if (this.HasInData.includes(this.inbound.type)){
          this.inData = <InData>{id: 0, tag: this.inbound.tag, addrs: [], outJson: {}}
        } else {
          this.inData = <InData>{id: -1}
        }
        this.title = "add"
      }
      this.inboundStats = this.$props.stats
      this.side = "s"
    },
    changeType() {
      // Tag change only in add outbound
      const tag = this.$props.index != -1 ? this.inbound.tag : this.inbound.type + "-" + this.inbound.listen_port
      // Use previous data
      const prevConfig = { tag: tag ,listen: this.inbound.listen, listen_port: this.inbound.listen_port }
      this.inbound = createInbound(this.inbound.type, prevConfig)
      if (this.HasInData.includes(this.inbound.type)){
        this.inData = <InData>{id: 0, tag: this.inbound.tag, addrs: [], outJson: {}}
      } else {
        this.inData = <InData>{id: -1}
      }
      this.side = "s"
    },
    add_addr() {
      this.inData.addrs.push(<Addr>{ server: location.hostname, server_port: this.inbound.listen_port })
    },
    closeModal() {
      this.updateData() // reset
      this.$emit('close')
    },
    saveChanges() {
      this.loading = true
      this.$emit('save', this.inbound, this.inboundStats, this.tls_id.value, this.inData)
      this.loading = false
    },
  },
  watch: {
    visible(newValue) {
      if (newValue) {
        this.updateData()
      }
    },
  },
  components: {
    Listen, InTls, Hysteria2, Naive, Direct, Shadowsocks,
    Users, Hysteria, ShadowTls, TProxy, Multiplex, Tuic, Transport,
    AddrVue, OutJsonVue, AnyTls
  }
}
</script>
