<template>
  <v-dialog transition="dialog-bottom-transition" width="400">
    <v-card class="panel-modal pa-4" style="border-radius: 12px;">
      <v-card-title class="text-h6 font-weight-bold text-grey-lighten-3 px-2 pb-2">
        {{ $t('admin.changeCred') + " " + user.username }}
      </v-card-title>
      <v-divider class="mb-4" style="opacity: 0.1;"></v-divider>
      <v-card-text class="pa-2">
        <v-row style="row-gap: 16px;">
          <v-col cols="12" class="py-1">
            <v-text-field 
              v-model="newData.oldPass" 
              :label="$t('admin.oldPass')" 
              :rules="passwordRules" 
              type="password" 
              variant="outlined"
              class="dark-input"
              hide-details="auto"
              required
            ></v-text-field>
          </v-col>
          <v-col cols="12" class="py-1">
            <v-text-field 
              v-model="newData.newUsername" 
              :label="$t('login.username')" 
              :rules="usernameRules" 
              variant="outlined"
              class="dark-input"
              hide-details="auto"
              required
            ></v-text-field>
          </v-col>
          <v-col cols="12" class="py-1">
            <v-text-field 
              v-model="newData.newPass" 
              :label="$t('admin.newPass')" 
              :rules="passwordRules" 
              type="password" 
              variant="outlined"
              class="dark-input"
              hide-details="auto"
              required
            ></v-text-field>
          </v-col>
          <v-col cols="12" class="py-1">
            <v-text-field 
              v-model="confirmPass" 
              :label="$t('admin.confirmPass')" 
              :rules="confirmRules" 
              type="password" 
              variant="outlined"
              class="dark-input"
              hide-details="auto"
              required
            ></v-text-field>
          </v-col>
        </v-row>
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
          @click="saveChanges"
        >
          {{ $t('actions.save') }}
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script lang="ts">
import { i18n } from '@/locales'

export default {
  props: ['visible', 'user'],
  data() {
    return {
      newData: {
        id: 0,
        oldPass: "",
        newUsername: "",
        newPass: ""
      },
      confirmPass: "",
      usernameRules: [
        (value: string) => {
          if (value?.length > 0) return true
          return i18n.global.t('login.unRules')
        },
      ],
      passwordRules: [
        (value: string) => {
          if (value?.length > 0) return true
          return i18n.global.t('login.pwRules')
        },
      ],
      confirmRules: [
        (value: string) => {
          if (!value) return "请再次输入新密码"
          if (value !== this.newData.newPass) return "两次输入的密码不一致"
          return true
        }
      ]
    }
  },
  methods: {
    resetData() {
      this.newData.id = this.$props.user.id
      this.newData.oldPass = ""
      this.newData.newUsername = this.$props.user.username
      this.newData.newPass = ""
      this.confirmPass = ""
    },
    closeModal() {
      this.resetData() // reset
      this.$emit('close')
    },
    saveChanges() {
      if (this.newData.oldPass == '' || this.newData.newUsername == '' || this.newData.newPass == '') return
      if (this.confirmPass !== this.newData.newPass) return
      this.$emit('save', this.newData)
    },
  },
  watch: {
    visible(newValue) {
      if (newValue) {
        this.resetData()
      }
    },
  },
}
</script>
