<template>
  <AdminModal 
    v-model="editModal.visible"
    :visible="editModal.visible"
    :user="editModal.user"
    @close="closeEditModal"
    @save="saveEditModal"
  />
  <ChngModal
    v-model="changesModal.visible"
    :visible="changesModal.visible"
    :admins="users.map((u:any) => u.username)"
    :actor="changesModal.actor"
    @close="closeChangesModal"
  />

  <!-- 精致顶部 Action Bar -->
  <v-row class="mb-4">
    <v-col cols="12">
      <div class="d-flex flex-column flex-sm-row align-sm-center justify-space-between pa-4 flat-card" style="gap: 16px;">
        <div class="text-h6 font-weight-bold text-grey-lighten-3 d-flex align-center">
          <v-icon color="primary" class="mr-2">mdi-shield-account-outline</v-icon>
          {{ $t('pages.admins') || 'Administrators' }}
        </div>
        <div>
          <v-btn class="tech-blue-btn text-none" prepend-icon="mdi-history" @click="showChangesModal('')" style="height: 40px; border-radius: 6px;">
            {{ $t('admin.changes') }}
          </v-btn>
        </div>
      </div>
    </v-col>
  </v-row>

  <!-- 管理员卡片网格 -->
  <v-row>
    <v-col cols="12" sm="6" md="4" lg="3" v-for="(item, index) in <any[]>users" :key="item.id">
      <v-card class="panel-card d-flex flex-column justify-space-between h-100 pa-4" style="min-height: 220px;">
        <div>
          <!-- 卡片头部：管理员名称 -->
          <div class="d-flex align-center justify-space-between mb-3">
            <span class="text-subtitle-1 font-weight-bold text-high-emphasis truncate" style="max-width: 70%;" :title="item.username">
              {{ item.username }}
            </span>
            <v-icon color="cyan" size="20">mdi-shield-check-outline</v-icon>
          </div>

          <v-divider class="mb-3" style="opacity: 0.08;"></v-divider>

          <!-- 登录详情：横向三列规整工业布局，彻底根除空旷感 -->
          <div class="d-flex flex-column mb-1">
            <div class="text-caption text-grey font-weight-bold mb-2">{{ $t('admin.lastLogin') }}</div>
            
            <v-row class="ma-0 text-center rounded pa-2" style="background: rgba(var(--v-theme-surface-variant), 0.04); border: 1px solid var(--panel-border-color);">
              <v-col cols="4" class="pa-1 d-flex flex-column justify-center" style="border-right: 1px dashed var(--panel-border-color);">
                <span class="text-grey mb-1" style="font-size: 0.725rem !important;">{{ $t('admin.date') }}</span>
                <span class="text-body-2 font-weight-bold text-high-emphasis font-mono" style="font-size: 0.775rem !important; line-height: 1.2;">{{ item.loginDate }}</span>
              </v-col>
              <v-col cols="4" class="pa-1 d-flex flex-column justify-center" style="border-right: 1px dashed var(--panel-border-color);">
                <span class="text-grey mb-1" style="font-size: 0.725rem !important;">{{ $t('admin.time') }}</span>
                <span class="text-body-2 font-weight-bold text-high-emphasis font-mono" style="font-size: 0.775rem !important; line-height: 1.2;">{{ item.loginTime }}</span>
              </v-col>
              <v-col cols="4" class="pa-1 d-flex flex-column justify-center">
                <span class="text-grey mb-1" style="font-size: 0.725rem !important;">IP</span>
                <span class="text-body-2 font-weight-bold text-cyan font-mono" style="font-size: 0.775rem !important; line-height: 1.2; overflow-x: auto;">{{ item.ip }}</span>
              </v-col>
            </v-row>
          </div>
        </div>

        <!-- 底部快捷操作：高品质扁平控制按钮 -->
        <div>
          <v-divider class="my-3" style="opacity: 0.08;"></v-divider>
          <div class="d-flex justify-space-between align-center" style="gap: 10px;">
            <v-btn 
              variant="outlined" 
              size="small" 
              color="primary" 
              class="flex-grow-1 text-none font-weight-bold" 
              prepend-icon="mdi-account-key-outline" 
              @click="showEditModal(item)"
              style="border-radius: 6px; text-transform: none; border-color: rgba(var(--v-theme-primary), 0.2); font-size: 0.8rem !important;"
            >
              {{ $t('actions.edit') }}
            </v-btn>
            <v-btn 
              variant="outlined" 
              size="small" 
              color="cyan" 
              class="flex-grow-1 text-none font-weight-bold" 
              prepend-icon="mdi-list-box-outline" 
              @click="showChangesModal(item.username)"
              style="border-radius: 6px; text-transform: none; border-color: rgba(6, 182, 212, 0.2); font-size: 0.8rem !important;"
            >
              {{ $t('admin.changes') }}
            </v-btn>
          </div>
        </div>

      </v-card>
    </v-col>
  </v-row>
</template>

<script lang="ts" setup>
import AdminModal from '@/layouts/modals/Admin.vue'
import ChngModal from '@/layouts/modals/Changes.vue'
import { i18n } from '@/locales'
import HttpUtils from '@/plugins/httputil'
import { Ref, ref, inject, onMounted } from 'vue'

const loading:Ref = inject('loading')?? ref(false)

const users = ref(<any[]>[])

onMounted(async () => {loadData()})

const loadData = async () => {
  loading.value = true
  const msg = await HttpUtils.get('api/users')
  loading.value = false
  if (msg.success) {
    users.value = [] // 确保每次加载时清空，防止重复追加
    msg.obj.forEach((u:any) => {
      const lastLogin = u.lastLogin.split(' ')
      const localLastLogin = lastLogin.length > 2 ? dateFormatted(Date.parse(lastLogin[0] + ' ' + lastLogin[1])) : '- -'
      const loginDateTime = localLastLogin.split(' ')
      users.value.push({
        id: u.id,
        username: u.username,
        loginDate: loginDateTime[0] ? loginDateTime[0].replace(',', '') : '-',
        loginTime: loginDateTime[1] ? loginDateTime[1].replace(',', '') : '-',
        ip: lastLogin[2]?? '-',
      })
    })
  }
}

const dateFormatted = (dt: number): string => {
  const locale = i18n.global.locale.value.replace('zh', 'zh-')
  const date = new Date(dt)
  return date.toLocaleString(locale)
}

const editModal = ref({
  visible: false,
  user: {},
})

const showEditModal = (user: any) => {
  editModal.value.user = user
  editModal.value.visible = true
}
const closeEditModal = () => {
  editModal.value.visible = false
  editModal.value.user = {}
}
const saveEditModal = async (data:any) => {
  loading.value=true
  const response = await HttpUtils.post('api/changePass',data)
  if(response.success){
    setTimeout(() => {
      loading.value=false
      editModal.value.visible = false
    }, 500)
  } else {
    loading.value=false
  }
}

const changesModal = ref({
  visible: false,
  actor: '',
})
const showChangesModal = (actor: string) => {
  changesModal.value.actor = actor
  changesModal.value.visible = true
}
const closeChangesModal = () => {
  changesModal.value.visible = false
  changesModal.value.actor = ''
}
</script>
