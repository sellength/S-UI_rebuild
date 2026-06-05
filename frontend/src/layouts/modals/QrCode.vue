<template>
  <v-dialog transition="dialog-bottom-transition" width="400">
    <v-card class="panel-modal pa-4" id="qrcode-modal" style="border-radius: 12px; max-height: 90vh; display: flex; flex-direction: column;">
      <v-card-title class="px-2 pb-2">
        <v-row align="center">
          <v-col class="text-h6 font-weight-bold text-grey-lighten-3 d-flex align-center">
            <v-icon color="cyan" class="mr-2">mdi-qrcode</v-icon>
            QrCode
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
          <v-tab value="sub" class="text-none font-weight-bold">{{ $t('setting.sub') }}</v-tab>
          <v-tab value="link" class="text-none font-weight-bold">{{ $t('client.links') }}</v-tab>
        </v-tabs>
        
        <v-window v-model="tab" class="pt-2">
          <!-- 订阅配置二维码 -->
          <v-window-item value="sub">
            <div class="d-flex flex-column" style="gap: 24px; padding: 10px 0;">
              <!-- Sub QR -->
              <div class="flat-card pa-4 d-flex flex-column align-center" style="border-radius: 8px;">
                <span class="text-caption font-weight-bold text-grey-lighten-2 mb-3">{{ $t('setting.sub') }}</span>
                <div class="qrcode-wrapper pa-3 bg-white rounded-lg cursor-pointer" @click="copyToClipboard(clientSub)">
                  <QrcodeVue :value="clientSub" :size="size" :margin="1" />
                </div>
                <span class="text-caption text-grey mt-2">Click to copy URL</span>
              </div>

              <!-- JSON Sub QR -->
              <div class="flat-card pa-4 d-flex flex-column align-center" style="border-radius: 8px;">
                <span class="text-caption font-weight-bold text-grey-lighten-2 mb-3">{{ $t('setting.jsonSub') }}</span>
                <div class="qrcode-wrapper pa-3 bg-white rounded-lg cursor-pointer" @click="copyToClipboard(clientSub + '?format=json')">
                  <QrcodeVue :value="clientSub + '?format=json'" :size="size" :margin="1" />
                </div>
                <span class="text-caption text-grey mt-2">Click to copy URL</span>
              </div>

              <!-- SING-BOX QR -->
              <div class="flat-card pa-4 d-flex flex-column align-center" style="border-radius: 8px;">
                <span class="text-caption font-weight-bold text-grey-lighten-2 mb-3">SING-BOX</span>
                <div class="qrcode-wrapper pa-3 bg-white rounded-lg cursor-pointer" @click="copyToClipboard(singbox)">
                  <QrcodeVue :value="singbox" :size="size" :margin="1" />
                </div>
                <span class="text-caption text-grey mt-2">Click to copy import scheme</span>
              </div>
            </div>
          </v-window-item>

          <!-- 局部链接二维码 -->
          <v-window-item value="link">
            <div class="d-flex flex-column" style="gap: 24px; padding: 10px 0;">
              <div 
                v-for="l in clientLinks" 
                :key="l.uri" 
                class="flat-card pa-4 d-flex flex-column align-center" 
                style="border-radius: 8px;"
              >
                <span class="text-caption font-weight-bold text-grey-lighten-2 mb-3">{{ l.remark ?? $t('client.' + l.type) }}</span>
                <div class="qrcode-wrapper pa-3 bg-white rounded-lg cursor-pointer" @click="copyToClipboard(l.uri)">
                  <QrcodeVue :value="l.uri" :size="size" :margin="1" />
                </div>
                <span class="text-caption text-grey mt-2 text-truncate w-100 text-center" style="max-width: 250px;">{{ l.uri }}</span>
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
import Data from '@/store/modules/data'
import Clipboard from 'clipboard'
import { i18n } from '@/locales'
import { push } from 'notivue'

export default {
  props: ['index', 'visible'],
  data() {
    return {
      tab: 'sub',
    }
  },
  methods: {
    copyToClipboard(txt:string) {
      const hiddenButton = document.createElement('button')
      hiddenButton.className = 'clipboard-btn'
      document.body.appendChild(hiddenButton)

      const clipboard = new Clipboard('.clipboard-btn', {
        text: () => txt,
        container: document.getElementById('qrcode-modal')?? undefined
      });

      clipboard.on('success', () => {
        clipboard.destroy()
        push.success({
          message: i18n.global.t('success') + ': ' + i18n.global.t('copyToClipboard'),
          duration: 5000,
        })
      })

      clipboard.on('error', () => {
        clipboard.destroy()
        push.error({
          message: i18n.global.t('failed') + ': ' + i18n.global.t('copyToClipboard'),
          duration: 5000,
        })
      })

      hiddenButton.click()
      document.body.removeChild(hiddenButton)
    }
  },
  computed: {
    clients() { return Data().clients },
    client() {
      if ( typeof this.$props.index != 'number' ) return <any>{}
      return this.clients[this.$props.index]
    },
    clientSub() {
      return Data().subURI + this.client.name
    },
    singbox() {
      const url = Data().subURI + this.client.name + '?format=json'
      return 'sing-box://import-remote-profile?url=' +  encodeURIComponent(url) + '#' + this.client.name
    },
    clientLinks() {
      return this.client.links?? []
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
