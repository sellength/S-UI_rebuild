<template>
  <div class="d-flex flex-column" style="gap: 24px;">
    <!-- 顶部 Action Bar，包含保存与重启 -->
    <div class="d-flex flex-column flex-sm-row align-sm-center justify-space-between pa-4 flat-card" style="gap: 16px;">
      <div class="text-h6 font-weight-bold d-flex align-center">
        <v-icon color="primary" class="mr-2">mdi-cog-outline</v-icon>
        {{ $t('pages.settings') || 'System Settings' }}
      </div>
      <div class="d-flex align-center" style="gap: 12px;">
        <v-btn 
          color="warning" 
          variant="outlined"
          class="text-none font-weight-bold px-4" 
          @click="restartApp" 
          :loading="loading" 
          :disabled="stateChange"
          style="height: 40px; border-radius: 6px; border: 1px solid rgba(245, 158, 11, 0.4) !important;"
        >
          {{ $t('actions.restartApp') }}
        </v-btn>
        <v-btn 
          class="tech-blue-btn text-none px-6" 
          @click="saveChanges" 
          :loading="loading" 
          :disabled="!stateChange"
          style="height: 40px; border-radius: 6px;"
        >
          {{ $t('actions.save') }}
        </v-btn>
      </div>
    </div>

    <!-- 主 Tab 面板 -->
    <div class="flat-card" style="border-radius: 8px;">
      <v-tabs
        v-model="tab"
        align-tabs="start"
        density="comfortable"
        show-arrows
        style="padding: 0 16px; margin-bottom: 16px;"
      >
        <v-tab value="t1" class="text-none font-weight-bold py-4">{{ $t('setting.interface') }}</v-tab>
        <v-tab value="t2" class="text-none font-weight-bold py-4">{{ $t('setting.sub') }}</v-tab>
        <v-tab value="t3" class="text-none font-weight-bold py-4">{{ $t('setting.jsonSub') }}</v-tab>
        <v-tab value="t4" class="text-none font-weight-bold py-4">Language</v-tab>
      </v-tabs>

      <div class="pa-6">
        <v-window v-model="tab">
          <!-- 面板接口设置 -->
          <v-window-item value="t1">
            <v-row style="row-gap: 20px;" class="pt-4">
              <v-col cols="12" sm="6" md="4">
                <v-text-field v-model="settings.webListen" :label="$t('setting.addr')" hide-details variant="outlined" class="dark-input"></v-text-field>
              </v-col>
              <v-col cols="12" sm="6" md="4">
                <v-text-field v-model.number="webPort" min="1" type="number" :label="$t('setting.port')" hide-details variant="outlined" class="dark-input"></v-text-field>
              </v-col>
              <v-col cols="12" sm="6" md="4">
                <v-text-field v-model="settings.webPath" :label="$t('setting.webPath')" hide-details variant="outlined" class="dark-input"></v-text-field>
              </v-col>
              <v-col cols="12" sm="6" md="4">
                <v-text-field v-model="settings.webDomain" :label="$t('setting.domain')" hide-details variant="outlined" class="dark-input"></v-text-field>
              </v-col>
              <v-col cols="12" sm="6" md="4">
                <v-text-field v-model="settings.webKeyFile" :label="$t('setting.sslKey')" hide-details variant="outlined" class="dark-input"></v-text-field>
              </v-col>
              <v-col cols="12" sm="6" md="4">
                <v-text-field v-model="settings.webCertFile" :label="$t('setting.sslCert')" hide-details variant="outlined" class="dark-input"></v-text-field>
              </v-col>
              <v-col cols="12" sm="6" md="4">
                <v-text-field v-model="settings.webURI" :label="$t('setting.webUri')" hide-details variant="outlined" class="dark-input"></v-text-field>
              </v-col>
              <v-col cols="12" sm="6" md="4">
                <v-text-field
                  type="number"
                  v-model.number="sessionMaxAge"
                  min="0"
                  :label="$t('setting.sessionAge')"
                  :suffix="$t('date.m')"
                  hide-details
                  variant="outlined"
                  class="dark-input"
                ></v-text-field>
              </v-col>
              <v-col cols="12" sm="6" md="4">
                <v-text-field
                  type="number"
                  v-model.number="trafficAge"
                  min="0"
                  :label="$t('setting.trafficAge')"
                  :suffix="$t('date.d')"
                  hide-details
                  variant="outlined"
                  class="dark-input"
                ></v-text-field>
              </v-col>
              <v-col cols="12" sm="6" md="4">
                <v-text-field v-model="settings.timeLocation" :label="$t('setting.timeLoc')" hide-details variant="outlined" class="dark-input"></v-text-field>
              </v-col>
            </v-row>
          </v-window-item>

          <!-- 订阅设置 -->
          <v-window-item value="t2">
            <div class="d-flex flex-column pt-4" style="gap: 20px;">
              <v-row>
                <v-col cols="12" sm="6" md="4">
                  <v-switch color="cyan" v-model="subEncode" :label="$t('setting.subEncode')" hide-details />
                </v-col>
                <v-col cols="12" sm="6" md="4">
                  <v-switch color="cyan" v-model="subShowInfo" :label="$t('setting.subInfo')" hide-details />
                </v-col>
              </v-row>
              <v-row style="row-gap: 20px;">
                <v-col cols="12" sm="6" md="4">
                  <v-text-field v-model="settings.subListen" :label="$t('setting.addr')" hide-details variant="outlined" class="dark-input"></v-text-field>
                </v-col>
                <v-col cols="12" sm="6" md="4">
                  <v-text-field
                    type="number"
                    v-model.number="subPort"
                    min="1"
                    :label="$t('setting.port')"
                    hide-details
                    variant="outlined"
                    class="dark-input"
                  ></v-text-field>
                </v-col>
              </v-row>
              <v-row style="row-gap: 20px;">
                <v-col cols="12" sm="6" md="4">
                  <v-text-field v-model="settings.subKeyFile" :label="$t('setting.sslKey')" hide-details variant="outlined" class="dark-input"></v-text-field>
                </v-col>
                <v-col cols="12" sm="6" md="4">
                  <v-text-field v-model="settings.subCertFile" :label="$t('setting.sslCert')" hide-details variant="outlined" class="dark-input"></v-text-field>
                </v-col>
              </v-row>
              <v-row style="row-gap: 20px;">
                <v-col cols="12" sm="6" md="4">
                  <v-text-field v-model="settings.subDomain" :label="$t('setting.domain')" hide-details variant="outlined" class="dark-input"></v-text-field>
                </v-col>
                <v-col cols="12" sm="6" md="4">
                  <v-text-field v-model="settings.subPath" :label="$t('setting.path')" hide-details variant="outlined" class="dark-input"></v-text-field>
                </v-col>
              </v-row>
              <v-row style="row-gap: 20px;">
                <v-col cols="12" sm="6" md="4">
                  <v-text-field
                    type="number"
                    v-model.number="subUpdates"
                    min="0"
                    :label="$t('setting.update')"
                    hide-details
                    variant="outlined"
                    class="dark-input"
                  ></v-text-field>
                </v-col>
                <v-col cols="12" sm="6" md="4">
                  <v-text-field v-model="settings.subURI" :label="$t('setting.subUri')" hide-details variant="outlined" class="dark-input"></v-text-field>
                </v-col>
              </v-row>
            </div>
          </v-window-item>

          <!-- JSON 订阅设置 -->
          <v-window-item value="t3">
            <SubJsonExtVue :settings="settings" />
          </v-window-item>

          <!-- 语言设置 -->
          <v-window-item value="t4">
            <v-row class="pt-4">
              <v-col cols="12" sm="6" md="4">
                <v-select
                  hide-details
                  label="Language"
                  :items="languages"
                  v-model="$i18n.locale"
                  @update:modelValue="changeLocale"
                  variant="outlined"
                  class="dark-input"
                ></v-select>
                <div class="text-caption text-grey mt-2 px-1">
                  {{ $t('setting.langTip') }}
                </div>
              </v-col>
            </v-row>
          </v-window-item>
        </v-window>
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { useLocale } from 'vuetify'
import { languages, i18n } from '@/locales'
import { Ref, computed, inject, onMounted, ref } from 'vue'
import HttpUtils from '@/plugins/httputil'
import { push } from 'notivue'
import { FindDiff } from '@/plugins/utils'
import SubJsonExtVue from '@/components/SubJsonExt.vue'
const locale = useLocale()
const tab = ref('t1')
const loading:Ref = inject('loading')?? ref(false)
const oldSettings = ref({})

const settings = ref({
	webListen: '',
	webDomain: '',
	webPort: '2095',
	webCertFile: '',
	webKeyFile: '',
  webPath: '/app/',
  webURI: '',
	sessionMaxAge: '0',
  trafficAge: '30',
	timeLocation: 'Asia/Tehran',
  subListen: '',
	subPort: '2096',
	subPath: '/sub/',
	subDomain: '',
	subCertFile: '',
	subKeyFile: '',
	subUpdates: '12',
	subEncode: 'true',
	subShowInfo: 'false',
	subURI: '',
  subJsonExt: '',
})

onMounted(async () => {loadData()})

const changeLocale = (l: any) => {
  locale.current.value = l ?? 'en'
  localStorage.setItem('locale', locale.current.value)
  push.success({
    message: i18n.global.t('setting.langSaved')
  })
}

const loadData = async () => {
  loading.value = true
  const msg = await HttpUtils.get('api/setting')
  loading.value = false
  if (msg.success) {
    settings.value = msg.obj
    oldSettings.value = { ...msg.obj }
  }
}

const saveChanges = async () => {
  loading.value = true
  const diff = {
    settings: JSON.stringify(FindDiff.Settings(settings.value,oldSettings.value)),
  }
  const msg = await HttpUtils.post('api/save', diff)
  if (msg.success) {
    loadData()
  }
  loading.value = false
}

const sleep = (ms: number) => new Promise(resolve => setTimeout(resolve, ms))

const restartApp = async () => {
  loading.value = true
  const msg = await HttpUtils.post('api/restartApp',{})
  if (msg.success) {
    let url = settings.value.webURI
    if (url !== '') {
      const isTLS = settings.value.webCertFile !== '' || settings.value.webKeyFile !== ''
      url = buildURL(settings.value.webDomain,settings.value.webPort.toString(),isTLS, settings.value.webPath)
    }
    await sleep(3000)
    window.location.replace(url)
  }
  loading.value = false
}

const buildURL = (host: string, port: string, isTLS: boolean, path: string) => {
  if (!host || host.length == 0) host = window.location.hostname
  if (!port || port.length == 0) port = window.location.port

  const protocol = isTLS ? 'https:' : 'http:'

  if (port === '' || (isTLS && port === '443') || (!isTLS && port === '80')) {
      port = ''
  } else {
      port = `:${port}`
  }

  return `${protocol}//${host}${port}${path}settings`
}

const subEncode = computed({
  get: () => { return settings.value.subEncode == 'true' },
  set: (v:boolean) => { settings.value.subEncode = v ? 'true' : 'false' }
})

const subShowInfo = computed({
  get: () => { return settings.value.subShowInfo == 'true' },
  set: (v:boolean) => { settings.value.subShowInfo = v ? 'true' : 'false' }
})

const webPort = computed({
  get: () => { return settings.value.webPort.length>0 ? parseInt(settings.value.webPort) : 2095 },
  set: (v:number) => { settings.value.webPort = v>0 ? v.toString() : '2095' }
})

const sessionMaxAge = computed({
  get: () => { return settings.value.sessionMaxAge.length>0 ? parseInt(settings.value.sessionMaxAge) : 0 },
  set: (v:number) => { settings.value.sessionMaxAge = v>0 ? v.toString() : '0' }
})

const trafficAge = computed({
  get: () => { return settings.value.trafficAge.length>0 ? parseInt(settings.value.trafficAge) : 0 },
  set: (v:number) => { settings.value.trafficAge = v>0 ? v.toString() : '0' }
})

const subPort = computed({
  get: () => { return settings.value.subPort.length>0 ? parseInt(settings.value.subPort) : 2096 },
  set: (v:number) => { settings.value.subPort = v>0 ? v.toString() : '2096' }
})

const subUpdates = computed({
  get: () => { return settings.value.subUpdates.length>0 ? parseInt(settings.value.subUpdates) : 12 },
  set: (v:number) => { settings.value.subUpdates = v>0 ? v.toString() : '12' }
})

const stateChange = computed(() => {
  return !FindDiff.deepCompare(settings.value,oldSettings.value)
})
</script>
