<template>
  <v-navigation-drawer
    v-model="showDrawer"
    :temporary="isMobile"
    :expand-on-hover="!isMobile && !lockedOpen"
    :rail="!isMobile && !lockedOpen"
    :permanent="!isMobile"
    :width="272"
    :rail-width="64"
    :class="['app-dock', { 'drawer-collapsed': !isMobile && !lockedOpen && !dockHover }]"
    @mouseenter="dockHover = true"
    @mouseleave="dockHover = false"
    @click="isMobile ? $emit('toggleDrawer') : null"
  >
    <v-list-item
      height="63"
      prepend-avatar="@/assets/logo.svg"
      title="S-UI"
    >
      <template v-slot:append v-if="!isMobile">
        <v-btn
          :icon="lockedOpen ? 'mdi-pin' : 'mdi-pin-outline'"
          size="small"
          variant="text"
          class="dock-pin"
          @click.stop="lockedOpen = !lockedOpen"
        />
      </template>
      <template v-slot:append v-if="isMobile">
        <v-icon icon="mdi-close" />
      </template>
    </v-list-item>

    <v-divider></v-divider>

    <v-list density="compact" nav class="dock-list" v-model:opened="opened">
      <template v-for="item in menu" :key="item.title">
        <v-list-group
          v-if="item.children"
          :value="item.title"
          fluid
        >
          <template v-slot:activator="{ props: groupProps }">
            <v-list-item
              v-bind="groupProps"
              :active="isGroupActive(item)"
            >
              <template v-slot:prepend>
                <v-icon :icon="item.icon"></v-icon>
              </template>
              <v-list-item-title v-text="$t(item.title)"></v-list-item-title>
            </v-list-item>
          </template>
          <v-list-item
            v-for="child in item.children"
            :key="child.title"
            link
            :to="child.to"
            :active="isItemActive(child)"
            class="drawer-child"
          >
            <template v-slot:prepend>
              <v-icon :icon="child.icon"></v-icon>
            </template>
            <v-list-item-title v-text="$t(child.title)"></v-list-item-title>
          </v-list-item>
        </v-list-group>

        <v-list-item
          v-else
          link
          :to="item.to"
          :active="isItemActive(item)"
        >
        <template v-slot:prepend>
          <v-icon :icon="item.icon"></v-icon>
        </template>
        <v-list-item-title v-text="$t(item.title)"></v-list-item-title>
        </v-list-item>
      </template>
    </v-list>
    <template v-slot:append>
      <v-list-item class="dock-logout" prepend-icon="mdi-logout" :title="$t('menu.logout')" @click="Logout"></v-list-item>
    </template>
  </v-navigation-drawer>
</template>

<script lang="ts" setup>
import { computed, ref, watch } from 'vue'
import router from '@/router'
import { logout } from '@/plugins/httputil'

const props = defineProps(['isMobile','displayDrawer'])
const lockedOpen = ref(false)
const dockHover = ref(false)

const showDrawer = computed((): boolean => {
  return props.displayDrawer
})

type MenuItem = {
  title: string
  icon: string
  path?: string
  to?: string | Record<string, any>
  match?: string[]
  children?: MenuItem[]
}

const menu: MenuItem[] = [
  { title: 'pages.home', icon: 'mdi-home', to: '/', match: ['/'] },
  {
    title: 'pages.distributed',
    icon: 'mdi-lan-connect',
    match: ['/distributed'],
    children: [
      { title: 'pages.clusterOverview', icon: 'mdi-view-dashboard-outline', to: { path: '/distributed', query: { tab: 'nodes' } }, match: ['/distributed:nodes'] },
      { title: 'pages.certificateCenter', icon: 'mdi-certificate-outline', to: { path: '/distributed', query: { tab: 'certificates' } }, match: ['/distributed:certificates'] },
      { title: 'pages.serviceEntrances', icon: 'mdi-router-wireless', to: { path: '/distributed', query: { tab: 'inbounds' } }, match: ['/distributed:inbounds'] },
      { title: 'pages.configTemplates', icon: 'mdi-code-json', to: { path: '/distributed', query: { tab: 'versions' } }, match: ['/distributed:versions'] },
    ],
  },
  {
    title: 'menu.userAndSubscription',
    icon: 'mdi-account-group',
    match: ['/clients', '/subscriptions'],
    children: [
      { title: 'pages.clients', icon: 'mdi-account-multiple', to: '/clients', match: ['/clients'] },
      { title: 'pages.subscriptions', icon: 'mdi-link-variant', to: '/subscriptions', match: ['/subscriptions'] },
    ],
  },
  {
    title: 'menu.system',
    icon: 'mdi-cog-outline',
    match: ['/admins', '/settings'],
    children: [
      { title: 'pages.admins', icon: 'mdi-account-tie', to: '/admins', match: ['/admins'] },
      { title: 'pages.settings', icon: 'mdi-cog', to: '/settings', match: ['/settings'] },
    ],
  },
]

menu.forEach((item: any) => {
  if (item.path && !item.to) item.to = item.path
})

const routeKey = computed(() => {
  const route = router.currentRoute.value
  if (route.path === '/distributed') {
    return `/distributed:${route.query.tab || 'nodes'}`
  }
  return route.path
})

const opened = ref<string[]>([])

const isExpanded = computed(() => {
  return props.isMobile || lockedOpen.value || dockHover.value
})

watch(routeKey, () => {
  if (isExpanded.value) {
    const activeGroup = menu.find(item => item.children && isGroupActive(item))
    if (activeGroup && !opened.value.includes(activeGroup.title)) {
      opened.value = [activeGroup.title]
    }
  }
}, { immediate: true })

watch(isExpanded, (expanded) => {
  if (!expanded) {
    opened.value = []
  } else {
    const activeGroup = menu.find(item => item.children && isGroupActive(item))
    if (activeGroup && !opened.value.includes(activeGroup.title)) {
      opened.value = [activeGroup.title]
    }
  }
})

const isItemActive = (item: MenuItem) => {
  return !!item.match?.includes(routeKey.value)
}

const isGroupActive = (item: MenuItem) => {
  const route = router.currentRoute.value
  return !!item.match?.includes(route.path)
}

const Logout = async () => {
  logout()
}
</script>

<style scoped>
.app-dock {
  border-right: 1px solid rgba(var(--v-border-color), 0.16);
}

.dock-list {
  padding: 10px 8px;
}

.dock-list :deep(.v-list-item) {
  min-height: 44px;
  border-radius: 8px;
  margin-bottom: 4px;
}

.dock-list :deep(.v-list-item:hover) {
  background: rgba(var(--v-theme-primary), 0.08);
}

.dock-list :deep(.v-list-item--active) {
  background: rgba(var(--v-theme-primary), 0.12);
  color: rgb(var(--v-theme-primary));
}

.drawer-child {
  margin: 0 0 4px 18px;
  padding-inline-start: 14px !important;
  border-left: 0;
  background: transparent;
}

.drawer-child :deep(.v-list-item__prepend) {
  opacity: 0.78;
}

.drawer-child.v-list-item--active {
  background: rgba(var(--v-theme-primary), 0.12);
}

.drawer-child.v-list-item--active :deep(.v-list-item__prepend) {
  opacity: 1;
}

.dock-pin {
  opacity: 0.7;
}

.dock-logout {
  margin: 8px;
  border-radius: 8px;
}

/* 当导航栏处于折叠状态时，隐藏所有子菜单项以防止排版错乱 */
.app-dock.drawer-collapsed :deep(.v-list-group__items),
.app-dock.drawer-collapsed :deep(.drawer-child) {
  display: none !important;
}
</style>
