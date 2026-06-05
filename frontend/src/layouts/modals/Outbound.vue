<template>
  <v-dialog transition="dialog-bottom-transition" width="800">
    <v-card class="panel-modal pa-4" style="border-radius: 12px; max-height: 90vh; display: flex; flex-direction: column;">
      <v-card-title class="px-2 pb-2">
        <span class="text-h6 font-weight-bold text-grey-lighten-3">
          {{ $t('actions.' + title) + " " + $t('objects.outbound') }}
        </span>
      </v-card-title>
      <v-divider class="mb-4" style="opacity: 0.1;"></v-divider>
      
      <v-card-text style="padding: 0 20px; overflow-y: auto;" class="flex-grow-1">
        <v-container style="padding: 0;">
          <v-tabs
            v-model="tab"
            density="comfortable"
            style="margin-bottom: 16px;"
          >
            <v-tab value="t1" class="text-none font-weight-bold">{{ $t('client.basics') }}</v-tab>
            <v-tab value="t2" class="text-none font-weight-bold">{{ $t('client.external') }}</v-tab>
          </v-tabs>

          <v-window v-model="tab" class="pt-2">
            <!-- 基础设置 -->
            <v-window-item value="t1">
              <v-row style="row-gap: 16px;" class="mb-4">
                <v-col cols="12" sm="6" md="4" class="py-1">
                  <v-select
                    hide-details
                    :label="$t('type')"
                    :items="Object.keys(outTypes).map((key,index) => ({title: key, value: Object.values(outTypes)[index]}))"
                    v-model="outbound.type"
                    @update:modelValue="changeType"
                    variant="outlined"
                    class="dark-input"
                    density="comfortable"
                  ></v-select>
                </v-col>
                <v-col cols="12" sm="6" md="4" class="py-1">
                  <v-text-field v-model="outbound.tag" :label="$t('objects.tag')" hide-details variant="outlined" class="dark-input" density="comfortable"></v-text-field>
                </v-col>
              </v-row>
              
              <v-row v-if="!NoServer.includes(outbound.type)" style="row-gap: 16px;" class="mb-4">
                <v-col cols="12" sm="6" md="4" class="py-1">
                  <v-text-field
                    :label="$t('out.addr')"
                    hide-details
                    v-model="outbound.server"
                    variant="outlined"
                    class="dark-input"
                    density="comfortable"
                  ></v-text-field>
                </v-col>
                <v-col cols="12" sm="6" md="4" class="py-1">
                  <v-text-field
                    :label="$t('out.port')"
                    type="number"
                    min="0"
                    hide-details
                    v-model.number="outbound.server_port"
                    variant="outlined"
                    class="dark-input"
                    density="comfortable"
                  ></v-text-field>
                </v-col>
              </v-row>

              <div class="d-flex flex-column" style="gap: 16px;">
                <Direct v-if="outbound.type == outTypes.Direct" direction="out" :data="outbound" />
                <Socks v-if="outbound.type == outTypes.SOCKS" :data="outbound" />
                <Http v-if="outbound.type == outTypes.HTTP" :data="outbound" />
                <Shadowsocks v-if="outbound.type == outTypes.Shadowsocks" direction="out" :data="outbound" />
                <Vmess v-if="outbound.type == outTypes.VMess" :data="outbound" />
                <Trojan v-if="outbound.type == outTypes.Trojan" :data="outbound" />
                <Wireguard v-if="outbound.type == outTypes.Wireguard" :data="outbound" />
                <Hysteria v-if="outbound.type == outTypes.Hysteria" direction="out" :data="outbound" />
                <ShadowTls v-if="outbound.type == outTypes.ShadowTLS" :data="outbound" />
                <Vless v-if="outbound.type == outTypes.VLESS" :data="outbound" />
                <Tuic v-if="outbound.type == outTypes.TUIC" direction="out" :data="outbound" />
                <Hysteria2 v-if="outbound.type == outTypes.Hysteria2" direction="out" :data="outbound" />
                <AnyTls v-if="outbound.type == outTypes.AnyTLS" :data="outbound" />
                <Tor v-if="outbound.type == outTypes.Tor" :data="outbound" />
                <Ssh v-if="outbound.type == outTypes.SSH" :data="outbound" />
                <Selector v-if="outbound.type == outTypes.Selector" :data="outbound" :tags="tags" />
                <UrlTest v-if="outbound.type == outTypes.URLTest" :data="outbound" :tags="tags" />

                <Transport v-if="Object.hasOwn(outbound,'transport')" :data="outbound" />
                <OutTLS v-if="Object.hasOwn(outbound,'tls')" :outbound="outbound" />
                <Multiplex v-if="Object.hasOwn(outbound,'multiplex')" direction="out" :data="outbound" />
                <Dial v-if="!NoDial.includes(outbound.type)" :dial="outbound" :outTags="tags" />
                <v-card class="pb-4">
                  <v-row>
                    <v-col cols="12" sm="6" md="4">
                      <v-switch v-model="outboundStats" color="cyan" :label="$t('stats.enable')" hide-details></v-switch>
                    </v-col>
                  </v-row>
                </v-card>
              </div>
            </v-window-item>

            <!-- 转换订阅或分享链接 -->
            <v-window-item value="t2">
              <v-row style="row-gap: 20px;">
                <v-col cols="12">
                  <v-text-field v-model="link" :label="$t('client.external')" hide-details variant="outlined" class="dark-input"></v-text-field>
                </v-col>
                <v-col cols="12" class="d-flex justify-center">
                  <v-btn class="tech-blue-btn px-6" :loading="loading" @click="linkConvert">{{ $t('submit') }}</v-btn>
                </v-col>
              </v-row>
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
import { OutTypes, createOutbound } from '@/types/outbounds'
import RandomUtil from '@/plugins/randomUtil'
import Dial from '@/components/Dial.vue'
import Multiplex from '@/components/Multiplex.vue'
import Transport from '@/components/Transport.vue'
import OutTLS from '@/components/tls/OutTLS.vue'
import Direct from '@/components/protocols/Direct.vue'
import Socks from '@/components/protocols/Socks.vue'
import Http from '@/components/protocols/Http.vue'
import Shadowsocks from '@/components/protocols/Shadowsocks.vue'
import Vmess from '@/components/protocols/Vmess.vue'
import Trojan from '@/components/protocols/Trojan.vue'
import Wireguard from '@/components/protocols/Wireguard.vue'
import Hysteria from '@/components/protocols/Hysteria.vue'
import ShadowTls from '@/components/protocols/OutShadowTls.vue'
import Vless from '@/components/protocols/Vless.vue'
import Tuic from '@/components/protocols/Tuic.vue'
import Hysteria2 from '@/components/protocols/Hysteria2.vue'
import AnyTls from '@/components/protocols/AnyTls.vue'
import Tor from '@/components/protocols/Tor.vue'
import Ssh from '@/components/protocols/Ssh.vue'
import Selector from '@/components/protocols/Selector.vue'
import UrlTest from '@/components/protocols/UrlTest.vue'
import HttpUtils from '@/plugins/httputil'
export default {
  props: ['visible', 'data', 'id', 'stats', 'tags'],
  emits: ['close', 'save'],
  data() {
    return {
      outbound: createOutbound("direct",{ "tag": "" }),
      title: "add",
      tab: "t1",
      link: "",
      loading: false,
      outTypes: OutTypes,
      outboundStats: false,
      NoDial: [OutTypes.Block, OutTypes.DNS, OutTypes.Selector, OutTypes.URLTest],
      NoServer: [OutTypes.Direct, OutTypes.Block, OutTypes.DNS, OutTypes.Selector, OutTypes.URLTest, OutTypes.Tor],
    }
  },
  methods: {
    updateData() {
      if (this.$props.id != -1) {
        const newData = JSON.parse(this.$props.data)
        this.outbound = createOutbound(newData.type, newData)
        this.title = "edit"
      }
      else {
        this.outbound = createOutbound("direct",{ tag: "direct-" + RandomUtil.randomSeq(3) })
        this.title = "add"
      }
      this.tab = "t1"
      this.outboundStats = this.$props.stats
    },
    changeType() {
      // Tag change only in add outbound
      const tag = this.$props.id != -1 ? this.outbound.tag : this.outbound.type + "-" + RandomUtil.randomSeq(3)
      // Use previous data
      const prevConfig = { tag: tag ,listen: this.outbound.listen, listen_port: this.outbound.listen_port }
      this.outbound = createOutbound(this.outbound.type, prevConfig)
    },
    closeModal() {
      this.updateData() // reset
      this.$emit('close')
    },
    saveChanges() {
      this.loading = true
      this.$emit('save', this.outbound, this.outboundStats)
      this.loading = false
    },
    async linkConvert() {
      if (this.link.length>0){
        this.loading = true
        const msg = await HttpUtils.post('api/linkConvert', { link: this.link })
        this.loading = false
        if (msg.success) {
          this.outbound = msg.obj
          this.tab = "t1"
          this.link = ""
        }
      }
    }
  },
  watch: {
    visible(newValue) {
      if (newValue) {
        this.updateData()
      }
    },
  },
  components: { Dial, Multiplex, Transport, OutTLS,
    Direct, Socks, Http, Shadowsocks, Vmess, Trojan,
    Wireguard, Hysteria, ShadowTls, Vless, Tuic,
    Hysteria2, AnyTls, Tor, Ssh, Selector, UrlTest }
}
</script>
