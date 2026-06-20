<template>
  <v-dialog transition="dialog-bottom-transition" width="400">
    <v-card class="panel-modal pa-4" id="qrcode-modal" style="border-radius: 12px; max-height: 90vh; display: flex; flex-direction: column;">
      <v-card-title class="px-2 pb-2">
        <v-row align="center">
          <v-col class="text-h6 font-weight-bold text-grey-lighten-3 d-flex align-center">
            <v-icon color="cyan" class="mr-2">mdi-qrcode</v-icon>
            订阅与直连
          </v-col>
          <v-spacer></v-spacer>
          <v-col cols="auto">
            <v-btn icon="mdi-close" variant="text" size="small" color="grey" @click="$emit('close')"></v-btn>
          </v-col>
        </v-row>
      </v-card-title>
      <v-divider class="mb-4" style="opacity: 0.1;"></v-divider>
      
      <v-card-text style="overflow-y: auto; padding: 0" class="flex-grow-1">
        <v-tabs
          v-model="tab"
          density="comfortable"
          style="margin-bottom: 16px;"
        >
          <v-tab value="sub" class="text-none font-weight-bold">客户端订阅</v-tab>
          <v-tab value="link" class="text-none font-weight-bold">节点直连</v-tab>
        </v-tabs>
        
        <v-window v-model="tab" class="pt-2">
          <!-- 唯一订阅配置二维码 -->
          <v-window-item value="sub">
            <div class="d-flex flex-column" style="gap: 24px; padding: 10px 0;">
              <div class="flat-card pa-4 d-flex flex-column align-center" style="border-radius: 8px;">
                <span class="text-caption font-weight-bold text-grey-lighten-2 mb-3">自适应唯一订阅 (支持 Clash/Sing-box)</span>
                
                <div v-if="clientSub" class="qrcode-wrapper pa-3 bg-white rounded-lg cursor-pointer" @click="copyToClipboard(clientSub)">
                  <QrcodeVue :value="clientSub" :size="size" :margin="1" />
                </div>
                <div v-else class="text-grey text-caption py-6">无有效订阅 Token</div>

                <span v-if="clientSub" class="text-caption text-grey mt-2 text-center select-text w-100 text-truncate px-2">{{ clientSub }}</span>
                <span v-if="clientSub" class="text-caption text-primary mt-1">点击二维码复制链接</span>
              </div>

              <!-- 一键导入协议按钮 -->
              <div v-if="clientSub" class="d-flex flex-column" style="gap: 8px;">
                <v-btn
                  class="tech-blue-btn text-none w-100"
                  prepend-icon="mdi-rocket-launch-outline"
                  @click="copyToClipboard(singbox)"
                >
                  复制 Sing-box 导入协议
                </v-btn>
              </div>
            </div>
          </v-window-item>

          <!-- 单节点直连分享二维码 -->
          <v-window-item value="link">
            <div class="d-flex flex-column" style="gap: 24px; padding: 10px 0;">
              <div 
                v-for="l in clientLinks" 
                :key="l.uri" 
                class="flat-card pa-4 d-flex flex-column align-center" 
                style="border-radius: 8px;"
              >
                <span class="text-caption font-weight-bold text-grey-lighten-2 mb-3">{{ l.remark ?? l.type }}</span>
                <div class="qrcode-wrapper pa-3 bg-white rounded-lg cursor-pointer" @click="copyToClipboard(l.uri)">
                  <QrcodeVue :value="l.uri" :size="size" :margin="1" />
                </div>
                <span class="text-caption text-grey mt-2 text-truncate w-100 text-center select-text" style="max-width: 250px;">{{ l.uri }}</span>
              </div>

              <!-- 空状态提示 -->
              <div v-if="clientLinks.length === 0" class="text-center py-8 px-4 text-grey">
                <v-icon icon="mdi-link-off" size="40" class="mb-2" />
                <div class="text-body-2 font-weight-bold">暂无节点直连链接</div>
                <div class="text-caption mt-1">请通过【客户端订阅】选项卡获取您的统一订阅链接。</div>
              </div>
            </div>
          </v-window-item>
        </v-window>
      </v-card-text>
    </v-card>
  </v-dialog>
</template>

<script lang="ts">
import QrcodeVue from 'qrcode.vue'
import Clipboard from 'clipboard'
import { push } from 'notivue'

interface Link {
  type: string
  uri: string
  remark?: string
}

export default {
  props: ['visible', 'token', 'client', 'settings'],
  data() {
    return {
      tab: 'sub',
    }
  },
  methods: {
    copyToClipboard(txt: string) {
      if (!txt) return
      const hiddenButton = document.createElement('button')
      hiddenButton.className = 'clipboard-btn'
      document.body.appendChild(hiddenButton)

      const clipboard = new Clipboard('.clipboard-btn', {
        text: () => txt,
        container: document.getElementById('qrcode-modal') ?? undefined
      });

      clipboard.on('success', () => {
        clipboard.destroy()
        push.success({
          message: "已成功复制到剪贴板",
          duration: 3000,
        })
      })

      clipboard.on('error', () => {
        clipboard.destroy()
        push.error({
          message: "复制到剪贴板失败",
          duration: 3000,
        })
      })

      hiddenButton.click()
      document.body.removeChild(hiddenButton)
    },
    getSubscriptionUrl(token: string) {
      const config = this.$props.settings || {}
      if (config.subURI) {
        let uri = config.subURI.trim()
        if (!/^https?:\/\//i.test(uri)) {
          uri = 'http://' + uri
        }
        if (!uri.endsWith('/')) {
          uri += '/'
        }
        return `${uri}sub/${token}`
      }

      const protocol = config.subCertFile ? 'https:' : window.location.protocol
      const host = config.subDomain || window.location.hostname
      const port = config.subPort ? `:${config.subPort}` : ':2096'
      const path = config.subPath || '/sub/'
      return `${protocol}//${host}${port}${path}${token}`
    }
  },
  computed: {
    clientSub() {
      return this.$props.token ? this.getSubscriptionUrl(this.$props.token) : ''
    },
    singbox() {
      const url = this.clientSub
      if (!url) return ''
      return 'sing-box://import-remote-profile?url=' + encodeURIComponent(url) + '#' + (this.$props.client?.name || 'sub')
    },
    clientLinks(): Link[] {
      const rawLinks = this.$props.client?.links
      if (!rawLinks) return []
      try {
        if (typeof rawLinks === 'string') {
          return JSON.parse(rawLinks) || []
        }
        return rawLinks || []
      } catch {
        return []
      }
    },
    size() {
      if (window.innerWidth > 380) return 260
      if (window.innerWidth > 330) return 230
      return 200
    }
  },
  watch: {
    visible(v) {
      if (v) {
        this.tab = 'sub'
      }
    },
  },
  components: { QrcodeVue }
}
</script>
