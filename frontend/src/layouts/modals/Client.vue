<template>
  <v-dialog transition="dialog-bottom-transition" width="800">
    <v-card class="panel-modal pa-4" style="border-radius: 12px; max-height: 90vh; display: flex; flex-direction: column;">
      <v-card-title class="px-2 pb-2">
        <span class="text-h6 font-weight-bold text-grey-lighten-3">
          {{ $t('actions.' + title) + " " + $t('objects.client') }}
        </span>
      </v-card-title>
      <v-divider class="mb-4" style="opacity: 0.1;"></v-divider>
      
      <v-card-text style="padding: 0 20px; overflow-y: auto;" class="flex-grow-1">
        <v-container style="padding: 0;">
          <v-tabs
            v-model="tab"
            align-tabs="start"
            density="comfortable"
            style="margin-bottom: 20px;"
          >
            <v-tab value="t1" class="text-none font-weight-bold">{{ $t('client.basics') }}</v-tab>
            <v-tab value="t2" class="text-none font-weight-bold">{{ $t('client.config') }}</v-tab>
            <v-tab value="t3" class="text-none font-weight-bold">{{ $t('client.links') }}</v-tab>
          </v-tabs>

          <v-window v-model="tab" class="pt-2">
            <!-- 基础设置 -->
            <v-window-item value="t1">
              <v-row class="mb-2">
                <v-col cols="12" sm="6" md="4" class="py-1">
                  <v-switch color="cyan" v-model="client.enable" :label="$t('enable')" hide-details></v-switch>
                </v-col>
              </v-row>
              <v-row style="row-gap: 20px;">
                <v-col cols="12" sm="6" md="4">
                  <v-text-field v-model="client.name" :label="$t('client.name')" hide-details variant="outlined" class="dark-input"></v-text-field>
                </v-col>
                <v-col cols="12" sm="6" md="4">
                  <v-text-field v-model="client.desc" :label="$t('client.desc')" hide-details variant="outlined" class="dark-input"></v-text-field>
                </v-col>
              </v-row>
              <v-row style="row-gap: 20px; margin-top: 10px;">
                <v-col cols="12" sm="6" md="4">
                  <v-text-field v-model.number="Volume" type="number" min="0" :label="$t('stats.volume')" suffix="GiB" hide-details variant="outlined" class="dark-input"></v-text-field>
                </v-col>
                <v-col cols="12" sm="6" md="4">
                  <DatePick :expiry="expDate" @submit="setDate" />
                </v-col>
              </v-row>

              <!-- 流量使用状态 -->
              <v-row v-if="index != -1" class="my-4">
                <v-col cols="12" sm="6" md="5" class="d-flex flex-column justify-center">
                  <div class="d-flex justify-space-between align-center mb-1">
                    <span class="text-caption text-grey-lighten-2">
                      {{ $t('stats.usage') }}: <span class="font-weight-bold text-grey-lighten-4">{{ total }}</span>
                      <sup dir="ltr" v-if="percent>0" class="text-cyan font-weight-bold">({{ percent }}%)</sup>
                    </span>
                    <v-btn density="compact" variant="text" icon="mdi-restore" color="cyan" @click="client.up=0;client.down=0">
                      <v-tooltip activator="parent" location="top">
                        {{ $t('reset') }}
                      </v-tooltip>
                      <v-icon size="16" />
                    </v-btn>
                  </div>
                  <v-progress-linear
                    v-model="percent"
                    :color="percentColor"
                    v-if="client.volume>0"
                    height="6"
                    rounded
                    class="w-100"
                  ></v-progress-linear>
                </v-col>
                <v-col cols="12" sm="6" md="5" class="d-flex align-center font-mono text-caption" style="gap: 12px;">
                  <div class="d-flex align-center">
                    <v-icon icon="mdi-upload-network-outline" color="orange" size="16" class="mr-1" />
                    <span class="text-orange font-weight-bold">{{ up }}</span>
                  </div>
                  <div class="d-flex align-center">
                    <v-icon icon="mdi-download-network-outline" color="success" size="16" class="mr-1" />
                    <span class="text-success font-weight-bold">{{ down }}</span>
                  </div>
                </v-col>
              </v-row>

              <v-row style="margin-top: 10px;">
                <v-col cols="12">
                  <v-combobox
                    v-model="clientInbounds"
                    :items="inboundTags"
                    :label="$t('client.inboundTags')"
                    multiple
                    chips
                    hide-details
                    variant="outlined"
                    class="dark-input"
                  ></v-combobox>
                </v-col>
              </v-row>
              <v-row class="mt-2">
                <v-col cols="auto">
                  <v-switch v-model="clientStats" color="cyan" :label="$t('stats.enable')" hide-details></v-switch>
                </v-col>
              </v-row>
            </v-window-item>

            <!-- 协议配置 -->
            <v-window-item value="t2">
              <div class="d-flex flex-column" style="gap: 20px; max-width: 600px; margin: 0 auto;">
                <v-row v-for="(value, key) in clientConfig" :key="key" align="center" style="row-gap: 12px;">
                  <v-col cols="12" sm="3" class="text-sm-right text-caption text-grey font-weight-bold py-1">
                    {{ String(key).toUpperCase() }}
                  </v-col>
                  <v-col cols="12" sm="9" class="py-1">
                    <v-text-field
                      v-if="value.password != undefined"
                      label="Password"
                      v-model="value.password"
                      hide-details
                      variant="outlined"
                      class="dark-input"
                      density="compact"
                    ></v-text-field>
                    <v-text-field
                      v-if="value.uuid != undefined"
                      label="UUID"
                      v-model="value.uuid"
                      hide-details
                      variant="outlined"
                      class="dark-input"
                      density="compact"
                    ></v-text-field>
                    <v-text-field
                      v-if="value.flow != undefined"
                      label="Flow"
                      v-model="value.flow"
                      hide-details
                      variant="outlined"
                      class="dark-input"
                      density="compact"
                    ></v-text-field>
                    <v-text-field
                      v-if="value.auth_str != undefined"
                      label="Auth String"
                      v-model="value.auth_str"
                      hide-details
                      variant="outlined"
                      class="dark-input"
                      density="compact"
                    ></v-text-field>
                  </v-col>
                </v-row>
              </div>
            </v-window-item>

            <!-- 订阅与外部链接 -->
            <v-window-item value="t3">
              <div class="d-flex flex-column" style="gap: 24px;">
                <!-- 本地链接列表 -->
                <div v-if="links.length > 0">
                  <div class="text-subtitle-2 font-weight-bold text-grey-lighten-3 mb-2">Local URIs</div>
                  <div class="d-flex flex-column" style="gap: 8px;">
                    <div 
                      v-for="(lnk, idx) in links" 
                      :key="idx" 
                      class="pa-3 rounded text-caption font-mono text-grey-lighten-2 d-flex align-center" 
                      style="background: #0f172a; border: 1px solid rgba(255, 255, 255, 0.05); overflow-x: auto; white-space: nowrap;"
                    >
                      <span class="text-cyan font-weight-bold mr-2">#{{ idx + 1 }}</span>
                      {{ lnk.uri }}
                    </div>
                  </div>
                </div>

                <!-- 外部连接设置 -->
                <div>
                  <div class="d-flex align-center justify-space-between mb-3">
                    <span class="text-subtitle-2 font-weight-bold text-grey-lighten-3">{{ $t('client.external') }}</span>
                    <v-btn class="tech-grey-btn text-none" size="small" prepend-icon="mdi-plus" @click="extLinks.push({ type: 'external', uri: ''})">
                      {{ $t('actions.add') }}
                    </v-btn>
                  </div>
                  <div class="d-flex flex-column" style="gap: 12px;">
                    <div v-for="(lnk, idx) in extLinks" :key="idx" class="d-flex align-center" style="gap: 8px;">
                      <v-text-field
                        dir="ltr"
                        hide-details
                        variant="outlined"
                        class="dark-input flex-grow-1"
                        density="compact"
                        :placeholder="'<protocol>://<data>'"
                        v-model="lnk.uri"
                      ></v-text-field>
                      <v-btn icon="mdi-delete-outline" color="error" variant="text" size="small" @click="extLinks.splice(idx,1)"></v-btn>
                    </div>
                  </div>
                </div>

                <!-- 订阅链接设置 -->
                <div>
                  <div class="d-flex align-center justify-space-between mb-3">
                    <span class="text-subtitle-2 font-weight-bold text-grey-lighten-3">{{ $t('client.sub') }}</span>
                    <v-btn class="tech-grey-btn text-none" size="small" prepend-icon="mdi-plus" @click="subLinks.push({ type: 'sub', uri: ''})">
                      {{ $t('actions.add') }}
                    </v-btn>
                  </div>
                  <div class="d-flex flex-column" style="gap: 12px;">
                    <div v-for="(lnk, idx) in subLinks" :key="idx" class="d-flex align-center" style="gap: 8px;">
                      <v-text-field
                        dir="ltr"
                        hide-details
                        variant="outlined"
                        class="dark-input flex-grow-1"
                        density="compact"
                        placeholder="http[s]://<domain>[:]<port>/<path>"
                        v-model="lnk.uri"
                      ></v-text-field>
                      <v-btn icon="mdi-delete-outline" color="error" variant="text" size="small" @click="subLinks.splice(idx,1)"></v-btn>
                    </div>
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
import { Link } from '@/plugins/link'
import { createClient, randomConfigs, updateConfigs } from '@/types/clients'
import DatePick from '@/components/DateTime.vue'
import { HumanReadable } from '@/plugins/utils'

export default {
  props: ['visible', 'data', 'index', 'inboundTags', 'stats'],
  emits: ['close', 'save'],
  data() {
    return {
      client: createClient(),
      title: "add",
      loading: false,
      clientStats: false,
      tab: "t1",
      clientConfig: <any>[],
      links: <Link[]>[],
      extLinks: <Link[]>[],
      subLinks: <Link[]>[],
    }
  },
  methods: {
    updateData() {
      if (this.$props.index != -1) {
        const newData = JSON.parse(this.$props.data)
        this.client = createClient(newData)
        this.title = "edit"
        this.clientConfig = this.client.config
      }
      else {
        this.client = createClient()
        this.title = "add"
        this.clientConfig = randomConfigs('client')
      }
      this.clientStats = this.$props.stats
      this.links = this.client.links.filter(l => l.type == 'local')
      this.extLinks = this.client.links.filter(l => l.type == 'external')
      this.subLinks = this.client.links.filter(l => l.type == 'sub')
      this.tab = "t1"
    },
    closeModal() {
      this.updateData() // reset
      this.$emit('close')
    },
    saveChanges() {
      this.loading = true
      this.client.config = updateConfigs(this.clientConfig, this.client.name)
      this.client.links = [
                        ...this.links,
                        ...this.extLinks.filter(l => l.uri != ''),
                        ...this.subLinks.filter(l => l.uri != '')]
      this.$emit('save', this.client, this.clientStats)
      this.loading = false
    },
    setDate(newDate:number){
      this.client.expiry = newDate
    }
  },
  computed: {
    clientInbounds: {
      get() { return this.client.inbounds.length>0 ? this.client.inbounds.filter(i => this.inboundTags.includes(i)) : [] },
      set(newValue:string[]) { this.client.inbounds = newValue.length == 0 ?  [] : newValue }
    },
    expDate: {
      get() { return this.client.expiry},
      set(v:any) { this.client.expiry = v }
    },
    Volume: {
      get() { return this.client.volume == 0 ? 0 : (this.client.volume / (1024 ** 3)) },
      set(v:number) { this.client.volume = v > 0 ? v*(1024 ** 3) : 0 }
    },
    up() :string { return HumanReadable.sizeFormat(this.client.up) },
    down() :string { return HumanReadable.sizeFormat(this.client.down) },
    total() :string { return HumanReadable.sizeFormat(this.client.down + this.client.up) },
    percent() :number { return this.client.volume>0 ? Math.round((this.client.up + this.client.down) *100 / this.client.volume) : 0 },
    percentColor() :string { return (this.client.up+this.client.down) >= this.client.volume ? 'error' : this.percent>90 ? 'warning' : 'success' },
  },
  watch: {
    visible(newValue) {
      if (newValue) {
        this.updateData()
      }
    },
  },
  components: { DatePick },
}
</script>
