<template>
  <RuleVue
    v-model="ruleModal.visible"
    :visible="ruleModal.visible"
    :index="ruleModal.index"
    :data="ruleModal.data"
    :clients="clients"
    :inTags="inboundTags"
    :outTags="outboundTags"
    :rsTags="rulesetTags"
    @close="closeRuleModal"
    @save="saveRuleModal"
  />
  <RulesetVue
    v-model="rulesetModal.visible"
    :visible="rulesetModal.visible"
    :index="rulesetModal.index"
    :data="rulesetModal.data"
    :outTags="outboundTags"
    @close="closeRulesetModal"
    @save="saveRulesetModal"
  />

  <!-- 精致顶部 Action Bar -->
  <v-row class="mb-4">
    <v-col cols="12">
      <div class="d-flex flex-column flex-sm-row align-sm-center justify-space-between pa-4 flat-card" style="gap: 16px;">
        <div class="text-h6 font-weight-bold d-flex align-center">
          <v-icon color="primary" class="mr-2">mdi-routes</v-icon>
          {{ $t('pages.rules') || 'Routing & Rules' }}
        </div>
        <div class="d-flex align-center" style="gap: 12px;">
          <v-btn class="tech-grey-btn text-none" prepend-icon="mdi-playlist-plus" @click="showRulesetModal(-1)" style="height: 40px; border-radius: 6px;">
            {{ $t('ruleset.add') }}
          </v-btn>
          <v-btn class="tech-blue-btn text-none" prepend-icon="mdi-plus" @click="showRuleModal(-1)" style="height: 40px; border-radius: 6px;">
            {{ $t('rule.add') }}
          </v-btn>
        </div>
      </div>
    </v-col>
  </v-row>

  <!-- 1. 规则集管理 Section -->
  <v-row class="mb-6">
    <v-col cols="12" class="d-flex align-center mb-1">
      <v-icon color="cyan" class="mr-2" size="18">mdi-folder-zip-outline</v-icon>
      <span class="text-subtitle-1 font-weight-bold">{{ $t('rule.ruleset') }}</span>
    </v-col>
    
    <v-col cols="12" sm="6" md="4" lg="3" v-for="(item, index) in <any[]>rulesets" :key="item.tag">
      <v-card class="panel-card d-flex flex-column justify-space-between h-100 pa-4">
        <div>
          <!-- 卡片头部：索引与类型 -->
          <div class="d-flex align-center justify-space-between mb-3">
            <span class="text-body-2 font-weight-bold">
              #{{ index + 1 }} {{ $t('ruleset.' + item.type) }}
            </span>
            <span class="text-caption px-2 py-0.5 rounded font-weight-bold bg-cyan-darken-3 text-cyan-lighten-4">
              {{ (item.format || '').toUpperCase() }}
            </span>
          </div>

          <v-divider class="mb-4" style="opacity: 0.1;"></v-divider>

          <!-- 卡片内容 -->
          <div class="d-flex flex-column mb-4" style="gap: 8px;">
            <div class="d-flex justify-space-between align-center">
              <span class="text-caption text-grey">{{ $t('objects.tag') }}</span>
              <span class="text-body-2 text-cyan font-weight-medium font-mono truncate" style="max-width: 60%;">{{ item.tag }}</span>
            </div>
            
            <div class="d-flex justify-space-between align-center">
              <span class="text-caption text-grey">{{ $t('ruleset.format') }}</span>
              <span class="text-body-2 font-mono" style="opacity: 0.85;">{{ item.format }}</span>
            </div>

            <div class="d-flex justify-space-between align-center">
              <span class="text-caption text-grey">{{ $t('actions.update') }}</span>
              <span class="text-body-2" style="opacity: 0.85;">{{ item.update_interval ?? '-' }}</span>
            </div>
          </div>
        </div>

        <!-- 卡片操作 -->
        <div>
          <v-divider class="my-4" style="opacity: 0.1;"></v-divider>
          <div class="d-flex justify-space-between align-center">
            <v-btn icon="mdi-file-edit-outline" variant="text" size="small" color="primary" @click="showRulesetModal(index)">
              <v-icon size="18" />
              <v-tooltip activator="parent" location="top" :text="$t('actions.edit')"></v-tooltip>
            </v-btn>
            <v-btn icon="mdi-delete-outline" variant="text" size="small" color="error" @click="delRulesetOverlay[index] = true">
              <v-icon size="18" />
              <v-tooltip activator="parent" location="top" :text="$t('actions.del')"></v-tooltip>
            </v-btn>
          </div>
        </div>

        <!-- 规则集删除二次确认 -->
        <v-overlay
          v-model="delRulesetOverlay[index]"
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
              <v-btn class="tech-grey-btn px-4" size="small" @click="delRulesetOverlay[index] = false">{{ $t('no') }}</v-btn>
              <v-btn color="error" class="px-4 font-weight-bold" size="small" @click="delRuleset(index)">{{ $t('yes') }}</v-btn>
            </div>
          </v-card>
        </v-overlay>

      </v-card>
    </v-col>
  </v-row>

  <!-- 2. 路由规则链条 Section -->
  <v-row>
    <v-col cols="12" class="d-flex align-center mb-1">
      <v-icon color="primary" class="mr-2" size="18">mdi-swap-horizontal</v-icon>
      <span class="text-subtitle-1 font-weight-bold">{{ $t('pages.rules') }}</span>
    </v-col>

    <v-col cols="12" sm="6" md="4" lg="3" v-for="(item, index) in <any[]>rules" :key="index">
      <v-card class="panel-card d-flex flex-column justify-space-between h-100 pa-4">
        <div>
          <!-- 卡片头部：索引与逻辑 -->
          <div class="d-flex align-center justify-space-between mb-3">
            <span class="text-body-2 font-weight-bold">
              Rule #{{ index + 1 }}
            </span>
            <span class="text-caption text-grey">
              {{ item.type != undefined ? $t('rule.logical') + ' (' + item.mode + ')' : $t('rule.simple') }}
            </span>
          </div>

          <v-divider class="mb-4" style="opacity: 0.1;"></v-divider>

          <!-- 卡片内容 -->
          <div class="d-flex flex-column mb-4" style="gap: 8px;">
            <div class="d-flex justify-space-between align-center">
              <span class="text-caption text-grey">{{ $t('objects.outbound') }}</span>
              <span 
                class="text-caption px-2 py-0.5 rounded font-weight-bold"
                :style="{ 
                  backgroundColor: getOutboundChipColor(item.outbound || item.action || 'direct').bg, 
                  color: getOutboundChipColor(item.outbound || item.action || 'direct').text, 
                  border: '1px solid ' + getOutboundChipColor(item.outbound || item.action || 'direct').border 
                }"
              >
                {{ (item.outbound || item.action || 'direct').toUpperCase() }}
              </span>
            </div>
            
            <div class="d-flex justify-space-between align-center">
              <span class="text-caption text-grey">{{ $t('pages.rules') }}</span>
              <span class="text-body-2 font-weight-medium" style="opacity: 0.85;">
                {{ item.rules ? item.rules.length : Object.keys(item).filter(r => !['rule_set_ipcidr_match_source','invert','outbound'].includes(r)).length }} items
              </span>
            </div>

            <div class="d-flex justify-space-between align-center">
              <span class="text-caption text-grey">{{ $t('rule.invert') }}</span>
              <span class="text-body-2 d-flex align-center">
                <v-icon 
                  size="14" 
                  :color="(item.invert?? false) ? 'warning' : 'grey'" 
                  class="mr-1"
                >
                  {{ (item.invert?? false) ? 'mdi-checkbox-marked-outline' : 'mdi-checkbox-blank-outline' }}
                </v-icon>
                <span :class="(item.invert?? false) ? 'text-warning font-weight-bold' : 'text-grey'">
                  {{ $t((item.invert?? false) ? 'yes' : 'no') }}
                </span>
              </span>
            </div>
          </div>
        </div>

        <!-- 卡片操作 -->
        <div>
          <v-divider class="my-4" style="opacity: 0.1;"></v-divider>
          <div class="d-flex justify-space-between align-center">
            <v-btn icon="mdi-file-edit-outline" variant="text" size="small" color="primary" @click="showRuleModal(index)">
              <v-icon size="18" />
              <v-tooltip activator="parent" location="top" :text="$t('actions.edit')"></v-tooltip>
            </v-btn>
            <v-btn icon="mdi-delete-outline" variant="text" size="small" color="error" @click="delRuleOverlay[index] = true">
              <v-icon size="18" />
              <v-tooltip activator="parent" location="top" :text="$t('actions.del')"></v-tooltip>
            </v-btn>
          </div>
        </div>

        <!-- 规则删除二次确认 -->
        <v-overlay
          v-model="delRuleOverlay[index]"
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
              <v-btn class="tech-grey-btn px-4" size="small" @click="delRuleOverlay[index] = false">{{ $t('no') }}</v-btn>
              <v-btn color="error" class="px-4 font-weight-bold" size="small" @click="delRule(index)">{{ $t('yes') }}</v-btn>
            </div>
          </v-card>
        </v-overlay>

      </v-card>
    </v-col>
  </v-row>
</template>

<script lang="ts" setup>
import Data from '@/store/modules/data'
import { computed, ref } from 'vue'
import RuleVue from '@/layouts/modals/Rule.vue'
import RulesetVue from '@/layouts/modals/Ruleset.vue'
import { Config } from '@/types/config'
import { logicalRule, ruleset } from '@/types/rules'

const appConfig = computed((): Config => {
  return <Config> Data().config
})

const clients = computed((): string[] => {
  return Data().clients.map((c:any) => c.name)
})

const route = computed((): any => {
  return appConfig.value.route
})

const rules = computed((): any[] => {
  const data = route.value
  if (!data){
    return []
  }
  if (!('rules' in data) || !Array.isArray(data.rules)) {
    data.rules = []
  }
  return data.rules
})

const rulesets = computed((): any[] => {
  const data = route.value
  if (!data){
    return []
  }
  if (!('rule_set' in data) || !Array.isArray(data.rule_set)) {
    data.rule_set = []
  }
  return data.rule_set
})

const rulesetTags = computed((): any[] => {
  return rulesets.value.map((rs:any) => rs.tag)
})

const outboundTags = computed((): string[] => {
  return appConfig.value.outbounds?.map((o:any) => o.tag)
})

const inboundTags = computed((): string[] => {
  return appConfig.value.inbounds?.map((i:any) => i.tag)
})

let delRuleOverlay = ref(new Array<boolean>)
let delRulesetOverlay = ref(new Array<boolean>)

const getOutboundChipColor = (outbound: string) => {
  const o = outbound?.toLowerCase() || ''
  if (o.includes('direct')) {
    return { bg: 'rgba(6, 182, 212, 0.15)', text: '#06b6d4', border: 'rgba(6, 182, 212, 0.3)' } // Cyan for Direct
  }
  if (o.includes('block') || o.includes('reject')) {
    return { bg: 'rgba(239, 68, 68, 0.15)', text: '#ef4444', border: 'rgba(239, 68, 68, 0.3)' } // Red for Block
  }
  return { bg: 'rgba(139, 92, 246, 0.15)', text: '#8b5cf6', border: 'rgba(139, 92, 246, 0.3)' } // Purple for Proxy
}

const ruleModal = ref({
  visible: false,
  index: -1,
  data: '',
})

const showRuleModal = (index: number) => {
  ruleModal.value.index = index
  ruleModal.value.data = index == -1 ? '' : JSON.stringify(rules.value[index])
  ruleModal.value.visible = true
}

const closeRuleModal = () => {
  ruleModal.value.visible = false
}

const saveRuleModal = (data:logicalRule) => {
  // Logical or simple
  const ruleData = data.type == 'logical' ? data : data.rules[0]

  // New or Edit
  if (ruleModal.value.index == -1) {
    rules.value.push(ruleData)
  } else {
    rules.value[ruleModal.value.index] = ruleData
  }
  ruleModal.value.visible = false
}

const delRule = (index: number) => {
  rules.value.splice(index,1)
  delRuleOverlay.value[index] = false
}

const rulesetModal = ref({
  visible: false,
  index: -1,
  data: '',
})

const showRulesetModal = (index: number) => {
  rulesetModal.value.index = index
  rulesetModal.value.data = index == -1 ? '' : JSON.stringify(rulesets.value[index])
  rulesetModal.value.visible = true
}

const closeRulesetModal = () => {
  rulesetModal.value.visible = false
}

const saveRulesetModal = (data:ruleset) => {
  // New or Edit
  if (rulesetModal.value.index == -1) {
    rulesets.value.push(data)
  } else {
    rulesets.value[rulesetModal.value.index] = data
  }
  rulesetModal.value.visible = false
}

const delRuleset = (index: number) => {
  rulesets.value.splice(index,1)
  delRulesetOverlay.value[index] = false
}
</script>
