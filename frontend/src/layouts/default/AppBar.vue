<template>
  <v-app-bar :elevation="5">
    <v-icon v-if="isMobile" icon="mdi-menu" @click="$emit('toggleDrawer')" />
    <span v-else style="width: 24px"></span>
    <v-app-bar-title :text="$t(<string>$router.currentRoute.value.name)" class="align-center text-center " />
    
    <!-- Shaking red alarm bell + Save Button -->
    <div v-if="stateChange" class="d-flex align-center mr-2">
      <v-icon color="error" class="alarm-shaking mr-1" size="22">mdi-bell-ring</v-icon>
      <v-btn prepend-icon="mdi-content-save" :text="$t('actions.save')" @click="saveChanges"></v-btn>
    </div>

    <v-switch
      v-model="darkMode"
      @update:modelValue="toggleTheme"
      hide-details
      inset
      true-icon="mdi-moon-waning-crescent"
      false-icon="mdi-white-balance-sunny"
      color="primary"
      class="ml-2 mr-4"
      style="max-width: 54px;"
    ></v-switch>
  </v-app-bar>
</template>

<script lang="ts" setup>
import { computed, ref } from "vue"
import { useTheme } from "vuetify"
import { FindDiff } from "@/plugins/utils"
import Data from "@/store/modules/data"

defineProps(['isMobile'])

const theme = useTheme()
const darkMode = ref(localStorage.getItem('theme') == "dark")

const store = Data()

const toggleTheme = () => {
  theme.global.name.value = darkMode.value ? "dark" : "light"
  localStorage.setItem('theme', theme.global.name.value)
}

const saveChanges = () => {
  store.pushData()
}

const oldData = computed((): any => {
  return {config: store.oldData.config, clients: store.oldData.clients, tls: store.oldData.tlsConfigs, inData: store.oldData.inData}
})

const newData = computed((): any => {
  return {config: store.config, clients: store.clients, tls: store.tlsConfigs, inData: store.inData}
})

const stateChange = computed((): any => {
  return !FindDiff.deepCompare(newData.value,oldData.value)
})
</script>

<style scoped>
@keyframes alarm-wobble {
  0% { transform: rotate(0); }
  15% { transform: rotate(15deg); }
  30% { transform: rotate(-15deg); }
  45% { transform: rotate(10deg); }
  60% { transform: rotate(-10deg); }
  75% { transform: rotate(5deg); }
  90% { transform: rotate(-5deg); }
  100% { transform: rotate(0); }
}

.alarm-shaking {
  animation: alarm-wobble 0.6s infinite;
  display: inline-block;
  transform-origin: top center;
  margin-top: -2px; /* Pull the bell up slightly to align with the button */
}
</style>

