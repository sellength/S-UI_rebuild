<template>
  <v-dialog :model-value="visible" @update:model-value="$emit('close')" transition="dialog-bottom-transition" width="400">
    <v-card class="panel-modal pa-4" id="qrcode-modal" style="border-radius: 12px; max-height: 90vh; display: flex; flex-direction: column;">
      <v-card-title class="px-2 pb-2">
        <v-row align="center">
          <v-col class="text-h6 font-weight-bold text-grey-lighten-3 d-flex align-center">
            <v-icon color="cyan" class="mr-2">mdi-qrcode</v-icon>
            订阅二维码
          </v-col>
          <v-spacer></v-spacer>
          <v-col cols="auto">
            <v-btn icon="mdi-close" variant="text" size="small" color="grey" @click="$emit('close')"></v-btn>
          </v-col>
        </v-row>
      </v-card-title>
      <v-divider class="mb-4" style="opacity: 0.1;"></v-divider>
      
      <v-card-text style="overflow-y: auto; padding: 0" class="flex-grow-1">
        <div class="d-flex flex-column" style="gap: 24px; padding: 10px 0;">
          <div class="flat-card pa-4 d-flex flex-column align-center" style="border-radius: 8px;">
            <span class="text-caption font-weight-bold text-grey-lighten-2 mb-3">自适应唯一订阅 (支持 Clash/Sing-box)</span>
            
            <div v-if="clientSub" class="qrcode-wrapper pa-3 bg-white rounded-lg cursor-pointer" @click="copyToClipboard(clientSub)">
              <QrcodeVue :value="clientSub" :size="size" :margin="1" />
            </div>
            <div v-else class="text-grey text-caption py-6">无有效订阅 Token</div>

            <v-textarea
              v-if="clientSub"
              v-model="clientSub"
              readonly
              rows="3"
              variant="outlined"
              density="compact"
              hide-details
              class="w-100 mt-2 text-caption font-mono"
              style="font-size: 11px; word-break: break-all;"
            ></v-textarea>
            <span v-if="clientSub" class="text-caption text-primary mt-2">点击二维码或复制按钮可复制链接</span>
          </div>

          <div v-if="clientSub" class="d-flex flex-column" style="gap: 8px;">
            <v-btn
              class="tech-blue-btn text-none w-100"
              prepend-icon="mdi-content-copy"
              @click="copyToClipboard(clientSub)"
            >
              复制订阅链接
            </v-btn>
          </div>
        </div>
      </v-card-text>
    </v-card>
  </v-dialog>
</template>

<script lang="ts">
import QrcodeVue from 'qrcode.vue'
import Clipboard from 'clipboard'
import { push } from 'notivue'

export default {
  props: ['visible', 'token', 'client', 'settings'],
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
    size() {
      if (window.innerWidth > 380) return 260
      if (window.innerWidth > 330) return 230
      return 200
    }
  },
  components: { QrcodeVue }
}
</script>
