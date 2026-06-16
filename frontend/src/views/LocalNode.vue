<template>
  <v-container fluid class="pa-0">
    <v-row class="mb-4">
      <v-col cols="12">
        <div class="d-flex flex-column flex-md-row align-md-center justify-space-between pa-4 flat-card" style="gap: 16px;">
          <div>
            <div class="text-h6 font-weight-bold text-grey-lighten-3 d-flex align-center">
              <v-icon color="primary" class="mr-2">mdi-server-network</v-icon>
              本地节点搭建
            </div>
            <div class="text-caption text-grey mt-1">当前服务器上的 sing-box 配置与运行入口</div>
          </div>
          <v-btn class="tech-blue-btn text-none" prepend-icon="mdi-router-wireless" to="/inbounds">
            配置本地入站
          </v-btn>
        </div>
      </v-col>
    </v-row>

    <v-row>
      <v-col cols="12" lg="8">
        <v-card class="panel-card pa-4 mb-4">
          <div class="card-section-title mb-4">本地搭建路径</div>
          <div class="local-flow">
            <router-link v-for="item in setupFlow" :key="item.path" :to="item.path" class="local-step">
              <v-icon :icon="item.icon" color="primary" size="24" />
              <div>
                <div class="text-body-1 font-weight-medium text-grey-lighten-2">{{ item.title }}</div>
                <div class="text-caption text-grey mt-1">{{ item.caption }}</div>
              </div>
              <v-icon icon="mdi-chevron-right" color="grey" />
            </router-link>
          </div>
        </v-card>

        <v-card class="panel-card pa-4">
          <div class="card-section-title mb-4">本地资源</div>
          <v-row>
            <v-col v-for="item in resourceCards" :key="item.path" cols="12" md="6">
              <router-link :to="item.path" class="resource-card">
                <div class="d-flex align-center justify-space-between">
                  <v-icon :icon="item.icon" color="cyan" size="22" />
                  <v-icon icon="mdi-arrow-right" color="grey" size="18" />
                </div>
                <div class="text-subtitle-1 font-weight-bold text-grey-lighten-2 mt-4">{{ item.title }}</div>
                <div class="text-caption text-grey mt-1">{{ item.caption }}</div>
              </router-link>
            </v-col>
          </v-row>
        </v-card>
      </v-col>

      <v-col cols="12" lg="4">
        <v-card class="panel-card pa-4 mb-4">
          <div class="card-section-title mb-4">本地模式</div>
          <div class="mode-line">
            <span>运行位置</span>
            <strong>控制端同机</strong>
          </div>
          <div class="mode-line">
            <span>配置来源</span>
            <strong>原版 S-UI</strong>
          </div>
          <div class="mode-line">
            <span>适用场景</span>
            <strong>单机部署</strong>
          </div>
        </v-card>

        <v-card class="panel-card pa-4">
          <div class="card-section-title mb-4">下一步</div>
          <div class="next-action">
            <v-icon icon="mdi-account-group" color="primary" />
            <div>
              <div class="text-body-2 text-grey-lighten-2">统一用户仍在“用户管理”维护</div>
              <div class="text-caption text-grey mt-1">本地订阅和集群订阅后续会收敛到同一个用户入口。</div>
            </div>
          </div>
          <v-btn class="tech-grey-btn text-none mt-4" prepend-icon="mdi-account-multiple" to="/clients" block>
            打开用户管理
          </v-btn>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<script setup lang="ts">
const setupFlow = [
  { title: '基础配置', caption: 'DNS、路由与通用 sing-box 配置', icon: 'mdi-application-cog', path: '/basics' },
  { title: '创建入站', caption: '配置本机监听协议和端口', icon: 'mdi-cloud-download', path: '/inbounds' },
  { title: '配置出站', caption: '维护本机 outbound 和链路选择', icon: 'mdi-cloud-upload', path: '/outbounds' },
  { title: '绑定用户', caption: '为用户生成订阅和可用凭据', icon: 'mdi-account-multiple', path: '/clients' },
]

const resourceCards = [
  { title: 'TLS 设置', caption: '本机证书、Reality、ECH 等配置', icon: 'mdi-certificate', path: '/tls' },
  { title: '路由规则', caption: '本机规则集和流量分流', icon: 'mdi-routes', path: '/rules' },
  { title: '入站管理', caption: '本地监听入口', icon: 'mdi-router-wireless', path: '/inbounds' },
  { title: '出站管理', caption: '本地出口策略', icon: 'mdi-export', path: '/outbounds' },
]
</script>

<style scoped>
.local-flow {
  display: grid;
  gap: 10px;
}

.local-step,
.resource-card {
  border: 1px solid var(--panel-card-border);
  border-radius: 8px;
  background: var(--panel-soft-bg);
  color: inherit;
  text-decoration: none;
  transition: border-color 0.16s ease, background 0.16s ease, transform 0.16s ease;
}

.local-step {
  display: grid;
  grid-template-columns: 34px 1fr 24px;
  align-items: center;
  gap: 12px;
  padding: 14px;
}

.resource-card {
  display: block;
  min-height: 128px;
  padding: 16px;
}

.local-step:hover,
.resource-card:hover {
  border-color: rgba(0, 210, 255, 0.52);
  background: var(--panel-soft-bg-hover);
  transform: translateY(-1px);
}

.mode-line {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  padding: 12px 0;
  border-bottom: 1px solid var(--panel-card-border);
  color: var(--panel-muted-text);
}

.mode-line:last-child {
  border-bottom: 0;
}

.mode-line strong {
  color: var(--panel-strong-text);
  font-weight: 600;
}

.next-action {
  display: grid;
  grid-template-columns: 28px 1fr;
  gap: 10px;
  align-items: start;
}
</style>
