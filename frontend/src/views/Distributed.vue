<template>
  <v-container fluid class="pa-0">
    <v-row class="mb-4">
      <v-col cols="12">
        <div class="d-flex flex-column flex-md-row align-md-center justify-space-between pa-4 flat-card" style="gap: 16px;">
          <div class="text-h6 font-weight-bold text-grey-lighten-3 d-flex align-center">
            <v-icon color="primary" class="mr-2">mdi-lan-connect</v-icon>
            {{ pageTitle }}
          </div>
          <div class="d-flex align-center" style="gap: 8px;">
            <v-btn
              v-if="route.path === '/distributed' && tab === 'nodes'"
              class="tech-blue-btn text-none"
              prepend-icon="mdi-map-marker-path"
              @click="guideDialog = true"
            >
              配置向导
              <v-chip size="x-small" color="white" variant="tonal" class="ml-2">{{ completedSetupSteps }}/{{ setupSteps.length }}</v-chip>
            </v-btn>
            <v-btn class="tech-grey-btn text-none" prepend-icon="mdi-refresh" @click="loadAll" :loading="loading">
              刷新
            </v-btn>
          </div>
        </div>
      </v-col>
    </v-row>

    <v-tabs v-if="showInlineTabs" v-model="tab" class="mb-4" density="comfortable">
      <v-tab v-for="item in visibleTabs" :key="item.value" :value="item.value">{{ item.title }}</v-tab>
    </v-tabs>

    <v-window v-model="tab">
      <v-window-item value="nodes">
        <v-row>
          <v-col cols="12">
            <div class="cluster-stats">
              <div v-for="stat in clusterStats" :key="stat.label" class="cluster-stat">
                <v-icon :icon="stat.icon" color="primary" size="22" />
                <div>
                  <div class="cluster-stat-value">{{ stat.value }}</div>
                  <div class="cluster-stat-label">{{ stat.label }}</div>
                </div>
              </div>
            </div>
          </v-col>

          <v-col cols="12" v-if="completedSetupSteps < setupSteps.length">
            <v-card class="guide-prompt-card pa-4">
              <div class="d-flex flex-column flex-md-row align-md-center justify-space-between" style="gap: 12px;">
                <div class="d-flex align-center" style="gap: 12px;">
                  <v-icon icon="mdi-map-marker-path" color="primary" size="26" />
                  <div>
                    <div class="card-section-title">集群配置还差 {{ setupSteps.length - completedSetupSteps }} 步</div>
                    <div class="text-caption text-grey mt-1">{{ nextSetupStep?.title }}：{{ nextSetupStep?.caption }}</div>
                  </div>
                </div>
                <v-btn class="tech-blue-btn text-none" prepend-icon="mdi-playlist-check" @click="guideDialog = true">继续向导</v-btn>
              </div>
            </v-card>
          </v-col>

          <v-col cols="12">
            <v-card class="panel-card pa-4">
              <div class="d-flex flex-column flex-md-row align-md-center justify-space-between mb-4" style="gap: 12px;">
                <div>
                  <div class="card-section-title">节点</div>
                  <div class="text-caption text-grey mt-1">远端 Agent 节点和可发布配置</div>
                </div>
                <v-btn class="tech-blue-btn text-none" prepend-icon="mdi-plus" @click="openNodeDrawer()">
                  新增节点
                </v-btn>
              </div>

              <v-alert v-if="nodes.length === 0" type="info" variant="tonal" density="compact" class="mb-4">
                先创建第一个远端节点，再进入证书和服务入口。
              </v-alert>

              <div class="cluster-node-table">
                <div class="cluster-node-table-head">
                  <div>节点</div>
                  <div class="text-center">节点状态</div>
                  <div class="text-center">节点配置</div>
                  <div class="text-center">最后心跳</div>
                  <div class="text-right">操作</div>
                </div>

                <div v-for="node in nodes" :key="node.id" class="cluster-node-row">
                  <div class="cluster-node-cell cluster-node-identity">
                    <v-icon icon="mdi-server-network" color="primary" size="22" />
                    <div class="cluster-node-name-wrap">
                      <div class="cluster-node-name">{{ node.name }}</div>
                      <div class="cluster-node-meta">{{ node.code }} · {{ node.region || '未填写地区' }}</div>
                      <div class="cluster-node-host">{{ node.publicHost || '未填写域名' }}</div>
                    </div>
                  </div>

                  <div class="cluster-node-cell">
                    <div class="node-status-inline">
                      <span class="status-pill" :class="`status-pill--${agentStatusMeta(node).tone}`">
                        <span class="status-dot" :class="`status-dot--${agentStatusMeta(node).tone}`" />
                        {{ agentStatusMeta(node).label }}
                      </span>
                      <span class="status-pill" :class="`status-pill--${singboxStatusMeta(node).tone}`">
                        <span class="status-dot" :class="`status-dot--${singboxStatusMeta(node).tone}`" />
                        {{ singboxStatusMeta(node).label }}
                      </span>
                    </div>
                  </div>

                  <div class="cluster-node-cell">
                    <div v-if="nodeSyncProgress[node.id] !== undefined" class="d-flex flex-column align-center" style="width: 120px;">
                      <div class="text-caption text-primary mb-1 font-weight-bold d-flex align-center">
                        <v-progress-circular indeterminate size="12" width="2" class="mr-1" color="primary" />
                        同步中 {{ nodeSyncProgress[node.id] }}%
                      </div>
                      <v-progress-linear
                        :model-value="nodeSyncProgress[node.id]"
                        color="primary"
                        height="6"
                        rounded
                        striped
                        active
                      />
                    </div>
                    <div v-else class="cluster-node-config-wrap">
                      <span class="config-version-pill" :class="nodeConfigClass(node)">
                        <v-icon 
                          v-if="nodeConfigClass(node) === 'config-version-pill--syncing'" 
                          icon="mdi-sync" 
                          class="mr-1 rotate-anim" 
                          size="12" 
                        />
                        {{ nodeConfigLabel(node) }}
                      </span>
                      <span class="cluster-node-submeta">{{ nodeInboundCount(node.id) }} 个服务入口</span>
                    </div>
                  </div>

                  <div class="cluster-node-cell cluster-node-heartbeat">
                    {{ lastSeenText(node.lastSeenAt) }}
                  </div>

                  <div class="cluster-node-cell cluster-node-actions">
                    <v-btn
                      size="small"
                      variant="tonal"
                      color="primary"
                      class="text-none tech-action-btn"
                      prepend-icon="mdi-pencil-outline"
                      @click="openNodeDrawer(node)"
                    >
                      编辑
                    </v-btn>
                    <v-btn
                      class="tech-blue-btn text-none tech-action-btn"
                      size="small"
                      prepend-icon="mdi-rocket-launch-outline"
                      :loading="publishingNodeIds.includes(node.id)"
                      :disabled="publishingNodeIds.includes(node.id)"
                      @click="publishNode(node.id)"
                    >
                      下发配置
                    </v-btn>
                    <v-btn
                      size="small"
                      variant="tonal"
                      color="error"
                      class="text-none tech-action-btn"
                      prepend-icon="mdi-delete-outline"
                      @click="deleteNode(node.id)"
                    >
                      删除
                    </v-btn>
                  </div>
                </div>
              </div>
            </v-card>
          </v-col>
        </v-row>
      </v-window-item>

      <v-window-item value="certificates">
        <v-row>
          <v-col cols="12">
            <div class="certificate-stats">
              <div class="certificate-stat">
                <v-icon icon="mdi-shield-check-outline" color="primary" size="28" />
                <div>
                  <div class="certificate-stat-label">证书总数</div>
                  <div class="certificate-stat-value">{{ certificates.length }}</div>
                  <div class="certificate-stat-hint">已签发 {{ issuedCertificates.length }}</div>
                </div>
              </div>
              <div class="certificate-stat">
                <v-icon icon="mdi-calendar-alert-outline" color="warning" size="28" />
                <div>
                  <div class="certificate-stat-label">即将过期</div>
                  <div class="certificate-stat-value">{{ expiringCertificateCount }}</div>
                  <div class="certificate-stat-hint">30 天内到期</div>
                </div>
              </div>
              <div class="certificate-stat">
                <v-icon icon="mdi-cloud-lock-outline" color="primary" size="28" />
                <div>
                  <div class="certificate-stat-label">DNS Providers</div>
                  <div class="certificate-stat-value">{{ dnsProviders.length }}</div>
                  <div class="certificate-stat-hint">已配置</div>
                </div>
              </div>
              <div class="certificate-stat">
                <v-icon icon="mdi-clock-check-outline" color="success" size="28" />
                <div>
                  <div class="certificate-stat-label">自动续期</div>
                  <div class="certificate-stat-value certificate-stat-success">{{ autoRenewCertificateCount > 0 ? '已开启' : '未开启' }}</div>
                  <div class="certificate-stat-hint">{{ autoRenewCertificateCount }} 张证书启用</div>
                </div>
              </div>
            </div>
          </v-col>
          <v-col cols="12">
            <v-card class="panel-card pa-4 mb-4">
              <div class="d-flex flex-column flex-md-row align-md-center justify-space-between mb-4" style="gap: 12px;">
                <div>
                  <div class="card-section-title">DNS Provider 列表</div>
                  <div class="text-caption text-grey mt-1">管理用于 DNS API 自动申请证书的凭据。</div>
                </div>
                <v-btn class="tech-blue-btn text-none" prepend-icon="mdi-plus" @click="openDNSProviderDialog">添加 Provider</v-btn>
              </div>
              <v-table density="compact">
                <thead>
                  <tr>
                    <th>名称</th>
                    <th>类型</th>
                    <th>状态</th>
                    <th class="text-right">操作</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="provider in dnsProviders" :key="provider.id">
                    <td>{{ provider.name }}</td>
                    <td>
                      <div class="dns-provider-option dns-provider-option-inline">
                        <img
                          v-if="providerIcon(provider.type)"
                          class="dns-provider-icon"
                          :src="providerIcon(provider.type)"
                          :alt="providerTypeName(provider.type)"
                        />
                        <v-icon v-else icon="mdi-server-network" class="dns-provider-icon dns-provider-icon-fallback" />
                        <span class="dns-provider-name">{{ providerTypeName(provider.type) }}</span>
                      </div>
                    </td>
                    <td>
                      <v-chip size="small" :color="provider.enable ? 'success' : 'grey'" variant="tonal">
                        {{ provider.enable ? '启用' : '禁用' }}
                      </v-chip>
                    </td>
                    <td class="text-right">
                      <v-btn icon="mdi-pencil" size="small" variant="text" @click="editDNSProvider(provider)" />
                    </td>
                  </tr>
                  <tr v-if="dnsProviders.length === 0">
                    <td colspan="4" class="text-center text-grey py-6">还没有 DNS Provider，点击右上角添加。</td>
                  </tr>
                </tbody>
              </v-table>
            </v-card>
            <v-card class="panel-card pa-4">
              <div class="d-flex flex-column flex-md-row align-md-center justify-space-between mb-4" style="gap: 12px;">
                <div>
                  <div class="card-section-title">TLS 证书列表</div>
                  <div class="text-caption text-grey mt-1">管理可下发到集群节点的 TLS 证书；Reality 密钥不属于这里。</div>
                </div>
                <v-btn class="tech-blue-btn text-none" prepend-icon="mdi-plus" @click="openCertificateDialog">申请证书</v-btn>
              </div>
              <v-table density="compact">
                <thead>
                  <tr>
                    <th>名称</th>
                    <th>域名</th>
                    <th>申请方式</th>
                    <th>DNS Provider</th>
                    <th>状态</th>
                    <th>到期日期</th>
                    <th class="text-right">操作</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="cert in certificates" :key="cert.id">
                    <td>{{ cert.name }}</td>
                    <td>{{ formatDomains(cert.domains) }}</td>
                    <td>{{ certSourceTitle(cert.source) }}</td>
                    <td>{{ cert.source === 'acme-dns01' ? dnsProviderName(cert.dnsProviderId) : '-' }}</td>
                    <td>
                      <v-chip size="small" :color="cert.activeVersionId ? 'success' : 'warning'" variant="tonal">
                        {{ cert.activeVersionId ? '已签发' : '未签发' }}
                      </v-chip>
                    </td>
                    <td>{{ cert.notAfter ? formatTimestamp(cert.notAfter) : '-' }}</td>
                    <td class="text-right">
                      <v-btn icon="mdi-pencil" size="small" variant="text" @click="editCertificate(cert)" />
                      <v-btn
                        v-if="cert.source === 'manual'"
                        class="tech-grey-btn text-none"
                        size="small"
                        disabled
                      >
                        旧手动证书
                      </v-btn>
                      <v-btn
                        v-else
                        class="tech-blue-btn text-none"
                        size="small"
                        prepend-icon="mdi-certificate-outline"
                        :loading="issuingCertificateId === cert.id"
                        @click="issueCertificate(cert.id)"
                      >
                        申请/续签
                      </v-btn>
                    </td>
                  </tr>
                  <tr v-if="certificates.length === 0">
                    <td colspan="7" class="text-center text-grey py-6">还没有证书，点击右上角申请。</td>
                  </tr>
                </tbody>
              </v-table>
            </v-card>
          </v-col>
        </v-row>
      </v-window-item>

      <v-window-item value="inbounds">
        <v-row>
          <v-col cols="12">
            <v-card class="panel-card pa-4">
              <div class="d-flex flex-column flex-md-row align-md-center justify-space-between mb-4" style="gap: 12px;">
                <div>
                  <div class="card-section-title">服务入口列表</div>
                  <div class="text-caption text-grey mt-1">每一行是一个节点上的协议入口；入口参数、策略配置和最终配置预览分开管理。</div>
                </div>
                <div class="d-flex align-center inbound-toolbar-actions">
                  <v-btn class="cluster-create-btn text-none" prepend-icon="mdi-plus" @click="openNewInbound">新建协议入口</v-btn>
                  <v-btn
                    class="cluster-delete-btn text-none"
                    prepend-icon="mdi-delete-outline"
                    :disabled="selectedInboundIds.length === 0"
                    @click="deleteSelectedInbounds"
                  >
                    删除选中
                  </v-btn>
                </div>
              </div>
              <v-table density="compact">
                <thead>
                  <tr>
                    <th style="width: 48px;">
                      <v-checkbox-btn
                        density="compact"
                        :model-value="allInboundsSelected"
                        :indeterminate="someInboundsSelected"
                        @update:model-value="toggleAllInbounds"
                      />
                    </th>
                    <th style="width: 64px;">序号</th>
                    <th>节点</th>
                    <th>协议</th>
                    <th>域名</th>
                    <th>端口</th>
                    <th>用户</th>
                    <th>状态</th>
                    <th>配置</th>
                    <th class="text-right">操作</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="(inbound, index) in distributedInbounds" :key="inbound.id">
                    <td>
                      <v-checkbox-btn v-model="selectedInboundIds" density="compact" :value="inbound.id" />
                    </td>
                    <td>{{ index + 1 }}</td>
                    <td>{{ nodeName(inbound.nodeId) }}</td>
                    <td>{{ inbound.protocol }}</td>
                    <td>{{ inbound.publicHost || nodeHost(inbound.nodeId) }}</td>
                    <td>{{ inbound.listenPort }}</td>
                    <td>
                      <v-btn
                        variant="text"
                        size="small"
                        color="primary"
                        class="text-none font-weight-bold px-1"
                        style="min-width: unset; text-decoration: underline;"
                        @click="openInboundUsersDialog(inbound)"
                      >
                        {{ inboundUsers.filter(u => u.inboundId === inbound.id).length }}
                      </v-btn>
                    </td>
                    <td>
                      <v-chip size="x-small" :color="inboundHealth(inbound).color" variant="tonal">
                        {{ inboundHealth(inbound).label }}
                      </v-chip>
                      <v-tooltip activator="parent" location="top" :text="inboundHealth(inbound).detail" />
                    </td>
                    <td>
                      <div class="d-flex align-center" style="gap: 6px;">
                        <v-chip size="x-small" :color="inbound.renderedConfigJson ? 'success' : 'warning'" variant="tonal">
                          {{ inbound.renderedConfigJson ? '已生成' : '未生成' }}
                        </v-chip>
                        <v-chip v-if="hasPolicyOverrides(inbound)" size="x-small" color="deep-purple" variant="tonal">
                          有策略
                        </v-chip>
                      </div>
                    </td>
                    <td class="text-right">
                      <div class="inbound-action-group">
                        <v-btn size="small" variant="text" class="inbound-action-btn text-none" @click="editInbound(inbound)">
                          <v-icon icon="mdi-pencil" size="16" class="mr-1" />
                          协议编辑
                          <v-tooltip activator="parent" location="top" text="协议编辑：修改节点、协议、域名、端口、证书和入站高级 JSON" />
                        </v-btn>
                        <v-btn size="small" variant="text" color="deep-purple" class="inbound-action-btn text-none" @click="openInboundPolicySettings(inbound)">
                          <v-icon icon="mdi-tune-variant" size="16" class="mr-1" />
                          配置设置
                          <v-tooltip activator="parent" location="top" text="配置设置：编辑该节点的 dns / outbounds / route 策略覆盖；Preview 阶段每个节点只允许一份" />
                        </v-btn>
                        <v-btn size="small" variant="text" color="teal" class="inbound-action-btn text-none" @click="openNodeConfig(inbound.nodeId)">
                          <v-icon icon="mdi-file-code-outline" size="16" class="mr-1" />
                          预览
                          <v-tooltip activator="parent" location="top" text="预览最终节点 config.json：包含全局配置和该节点所有服务入口" />
                        </v-btn>
                        <v-btn
                          size="small"
                          variant="text"
                          color="primary"
                          class="inbound-action-btn text-none"
                          :loading="renderingInboundId === inbound.id"
                          @click="renderInbound(inbound.id)"
                        >
                          <v-icon icon="mdi-refresh" size="16" class="mr-1" />
                          生成配置
                          <v-tooltip activator="parent" location="top" text="生成配置：按表单和高级 JSON 生成该服务入口的最终配置" />
                        </v-btn>
                      </div>
                    </td>
                  </tr>
                </tbody>
              </v-table>
            </v-card>
          </v-col>
        </v-row>
      </v-window-item>

      <v-window-item value="subscriptions">
        <v-card class="panel-card pa-4">
          <div class="d-flex align-center justify-space-between mb-4">
            <div class="card-section-title">订阅 Token</div>
            <div class="d-flex align-center" style="gap: 8px;">
              <v-select v-model="subscriptionClientId" :items="clientOptions" label="用户" density="compact" variant="outlined" hide-details style="width: 220px;" />
              <v-btn class="tech-blue-btn" @click="createSubscription">生成订阅</v-btn>
            </div>
          </div>
          <v-alert v-if="lastToken" type="success" variant="tonal" class="mb-4">
            <div class="d-flex align-center justify-space-between">
              <div>
                <strong>新订阅生成成功！</strong>
                <div class="text-caption text-error font-weight-bold mt-1">⚠️ 注意：由于安全策略该链接仅在生成时显示一次，请妥善保存！</div>
              </div>
            </div>
            <v-divider class="my-2" style="opacity: 0.1;" />
            <div class="d-flex align-center justify-space-between mt-2" style="gap: 16px;">
              <div class="text-truncate flex-grow-1 select-text" style="font-family: monospace;">{{ distributedSubUrl(lastToken) }}</div>
              <div class="d-flex" style="gap: 8px;">
                <v-btn size="small" class="tech-blue-btn text-none py-1" prepend-icon="mdi-content-copy" @click="copyText(distributedSubUrl(lastToken))">复制链接</v-btn>
                <v-btn size="small" variant="outlined" color="cyan" class="text-none py-1" prepend-icon="mdi-qrcode" @click="showQrCode(lastToken, getClientById(subscriptionClientId))">二维码</v-btn>
              </div>
            </div>
          </v-alert>
          <v-table density="compact">
            <thead>
              <tr>
                <th>用户</th>
                <th>Token Hash</th>
                <th>最近使用</th>
                <th>说明</th>
                <th class="text-right">操作</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="sub in subscriptions" :key="sub.id">
                <td>{{ clientName(sub.clientId) }}</td>
                <td class="text-truncate" style="max-width: 420px; font-family: monospace;">{{ sub.tokenHash }}</td>
                <td>{{ sub.lastUsedAt ? formatTimestamp(sub.lastUsedAt) : '-' }}</td>
                <td>{{ sub.token ? '自适应订阅 (支持 Clash/Sing-box/Base64)' : '历史旧版订阅' }}</td>
                <td class="text-right">
                  <!-- 复制完整订阅链接 -->
                  <v-btn
                    v-if="sub.token"
                    icon
                    size="small"
                    variant="text"
                    color="primary"
                    class="mr-2"
                    @click="copyText(distributedSubUrl(sub.token))"
                  >
                    <v-icon icon="mdi-content-copy" size="18" />
                    <v-tooltip activator="parent" location="top">复制链接</v-tooltip>
                  </v-btn>
                  <!-- 二维码 -->
                  <v-btn
                    v-if="sub.token"
                    icon
                    size="small"
                    variant="text"
                    color="cyan"
                    class="mr-2"
                    @click="showQrCode(sub.token, getClientById(sub.clientId))"
                  >
                    <v-icon icon="mdi-qrcode" size="18" />
                    <v-tooltip activator="parent" location="top">二维码</v-tooltip>
                  </v-btn>
                  <!-- 删除 -->
                  <v-btn
                    icon
                    size="small"
                    variant="text"
                    color="error"
                    @click="deleteSubscription(sub.id)"
                  >
                    <v-icon icon="mdi-trash-can-outline" size="18" />
                    <v-tooltip activator="parent" location="top">删除</v-tooltip>
                  </v-btn>
                </td>
              </tr>
            </tbody>
          </v-table>
        </v-card>
      </v-window-item>

      <v-window-item value="versions">
        <v-row>
          <v-col cols="12">
            <v-card class="panel-card pa-4">
              <div class="mb-4">
                <div class="card-section-title">全局默认配置</div>
                <div class="text-caption text-grey mt-1">
                  系统内置的基础运行配置，仅用于查看。服务入口和入口级配置会在发布节点时自动融合。
                </div>
              </div>
              <div class="template-editor-shell">
                <div class="template-editor-toolbar">
                  <v-chip size="small" color="primary" variant="tonal">config.json</v-chip>
                  <span>只读：系统默认骨架。协议差异请在服务入口的配置文件设置里调整。</span>
                </div>
                <div class="template-editor">
                  <div class="template-line-numbers" :style="{ transform: `translateY(-${templateEditorScrollTop}px)` }">
                    <span v-for="line in templateLineNumbers" :key="line">{{ line }}</span>
                  </div>
                  <div class="template-code-wrap">
                    <div class="template-code-view" @scroll="syncTemplateEditorScroll">
                      <pre class="template-highlight template-highlight-view" v-html="highlightedNodeTemplate" />
                    </div>
                  </div>
                </div>
              </div>
            </v-card>
            <v-card class="panel-card pa-4 mt-4">
              <div class="card-section-title mb-4">配置版本</div>
              <v-table density="compact">
                <thead>
                  <tr>
                    <th>节点</th>
                    <th>版本</th>
                    <th>状态</th>
                    <th>SHA256</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="version in configVersions" :key="version.id">
                    <td>{{ nodeName(version.nodeId) }}</td>
                    <td>{{ version.version }}</td>
                    <td>{{ version.status }}</td>
                    <td class="text-truncate" style="max-width: 160px; font-family: monospace;">{{ version.sha256 }}</td>
                  </tr>
                </tbody>
              </v-table>
            </v-card>
          </v-col>
        </v-row>
      </v-window-item>
    </v-window>

    <v-dialog v-model="guideDialog" width="620">
      <v-card class="panel-modal pa-4">
        <div class="d-flex align-center justify-space-between mb-4">
          <div>
            <div class="text-h6 font-weight-bold text-grey-lighten-3">集群配置向导</div>
            <div class="text-caption text-grey mt-1">按顺序完成远端节点、证书、入口、用户和配置发布。</div>
          </div>
          <v-btn icon="mdi-close" variant="text" @click="guideDialog = false" />
        </div>
        <div class="guide-progress mb-4">
          <div class="guide-progress-bar" :style="{ width: `${Math.round((completedSetupSteps / setupSteps.length) * 100)}%` }"></div>
        </div>
        <div class="setup-list mt-2">
          <button
            v-for="(step, index) in setupSteps"
            :key="step.key"
            class="setup-step guide-step-item"
            :class="{ 
              'is-active': step.key === nextSetupStep?.key && !step.done,
              'is-done': step.done,
              'is-pending': step.key !== nextSetupStep?.key && !step.done 
            }"
            type="button"
            @click="goSetupStep(step); guideDialog = false"
          >
            <!-- 步骤序号与垂直流水线 -->
            <div class="step-connector-wrapper">
              <div class="step-badge">
                <v-icon v-if="step.done" icon="mdi-check" color="success" size="14" />
                <span v-else>{{ index + 1 }}</span>
              </div>
              <div class="step-connector-line"></div>
            </div>
            
            <!-- 步骤主要内容 -->
            <div class="step-body ml-2">
              <div class="d-flex align-center">
                <v-icon :icon="step.icon" class="mr-2" size="18" />
                <span class="text-body-2 font-weight-medium text-grey-lighten-2">{{ step.title }}</span>
                <!-- 活动步骤指示小红点/呼吸灯 -->
                <span v-if="step.key === nextSetupStep?.key && !step.done" class="pulse-dot ml-2"></span>
              </div>
              <div class="text-caption text-grey mt-1">{{ step.caption }}</div>
            </div>

            <!-- 右侧操作状态药丸 -->
            <div class="step-action-badge">
              <v-chip v-if="step.done" size="x-small" color="success" variant="tonal" class="text-none">已完成</v-chip>
              <v-chip v-else-if="step.key === nextSetupStep?.key" size="x-small" color="primary" variant="flat" class="text-none animate-pulse-btn">立即开始</v-chip>
              <v-chip v-else size="x-small" color="grey" variant="tonal" class="text-none">等待中</v-chip>
            </div>
          </button>
        </div>
      </v-card>
    </v-dialog>

    <v-navigation-drawer
      v-model="nodeDrawer"
      temporary
      location="right"
      width="460"
      class="node-editor-drawer"
    >
      <div class="pa-4">
        <div class="d-flex align-center justify-space-between mb-4">
          <div class="text-subtitle-1 font-weight-bold text-grey-lighten-2">
            {{ nodeForm.id ? '编辑节点' : '新增节点' }}
          </div>
          <v-btn icon="mdi-close" size="small" variant="text" @click="nodeDrawer = false" />
        </div>
        <v-text-field v-model="nodeForm.name" label="名称" placeholder="US-01" density="compact" variant="outlined" hide-details class="mb-3" />
        <v-text-field v-model="nodeForm.code" label="节点代号" placeholder="us-01" density="compact" variant="outlined" hide-details class="mb-3" />
        <v-text-field v-model="nodeForm.region" label="地区" placeholder="us / sg" density="compact" variant="outlined" hide-details class="mb-3" />
        <v-text-field v-model="nodeForm.provider" label="服务商" placeholder="aws" density="compact" variant="outlined" hide-details class="mb-3" />
        <v-text-field v-model="nodeForm.publicHost" label="公开域名" placeholder="us.example.com" density="compact" variant="outlined" hide-details class="mb-3" />
        <v-text-field v-model="nodeForm.publicIp" label="公网 IP" density="compact" variant="outlined" hide-details class="mb-3" />
        <v-switch v-model="nodeForm.enable" color="primary" label="启用" hide-details class="mb-3" />
        <div class="d-flex justify-end mt-6" style="gap: 8px;">
          <v-btn class="tech-grey-btn" @click="resetNode(); nodeDrawer = false">取消</v-btn>
          <v-btn class="tech-blue-btn" :disabled="!canSaveNode" @click="saveNode">保存节点</v-btn>
        </div>
      </div>
    </v-navigation-drawer>

    <v-dialog v-model="inboundDialog" width="760" persistent>
      <v-card class="panel-modal pa-4">
        <div class="d-flex align-center justify-space-between mb-4">
          <div>
            <div class="text-h6 font-weight-bold text-grey-lighten-3">{{ inboundForm.id ? '协议编辑' : '新建协议入口' }}</div>
            <div class="text-caption text-grey mt-1">修改节点、协议、域名、端口、证书和协议高级 JSON。</div>
          </div>
          <v-btn icon="mdi-close" variant="text" @click="cancelInboundEdit" />
        </div>

        <v-row>
          <v-col cols="12" md="6">
            <v-select v-model="inboundForm.nodeId" :items="nodeOptions" label="节点" density="compact" variant="outlined" hide-details class="mb-3" />
          </v-col>
          <v-col cols="12" md="6">
            <v-select v-model="inboundForm.protocol" :items="inboundProtocolOptions" label="协议类型" density="compact" variant="outlined" hide-details class="mb-3" />
          </v-col>
          <v-col cols="12" md="6">
            <v-text-field v-model="inboundForm.publicHost" label="公开域名" density="compact" variant="outlined" hide-details class="mb-3" />
          </v-col>
          <v-col cols="12" md="3">
            <v-text-field v-model="inboundForm.listen" label="监听地址" density="compact" variant="outlined" hide-details class="mb-3" />
          </v-col>
          <v-col cols="12" md="3">
            <v-text-field v-model.number="inboundForm.listenPort" label="端口" type="number" density="compact" variant="outlined" hide-details class="mb-3" />
          </v-col>
          <v-col v-if="protocolNeedsCertificate" cols="12">
            <v-select
              v-model="inboundForm.certificateId"
              :items="certificateOptions"
              label="证书"
              density="compact"
              variant="outlined"
              hide-details
              class="mb-3"
            />
          </v-col>
          <v-col cols="12">
            <v-alert v-if="inboundForm.protocol !== 'anytls'" type="info" variant="tonal" density="compact" class="mb-3">
              当前协议先使用高级 JSON 生成配置。节点、监听地址和端口会作为默认值补齐；协议专属字段请在 JSON 中填写。
            </v-alert>
            <v-textarea
              v-model="inboundForm.advancedText"
              label="协议高级 JSON"
              rows="10"
              density="compact"
              variant="outlined"
              hide-details
              class="json-editor"
            />
          </v-col>
          <v-col cols="12">
            <v-switch v-model="inboundForm.enable" color="primary" label="启用" hide-details />
          </v-col>
        </v-row>

        <div class="d-flex justify-end mt-4" style="gap: 8px;">
          <v-btn class="tech-grey-btn" @click="cancelInboundEdit">取消</v-btn>
          <v-btn class="tech-blue-btn" :disabled="!canSaveInbound" @click="saveDistributedInbound">{{ inboundForm.id ? '保存修改' : '创建入口' }}</v-btn>
        </div>
      </v-card>
    </v-dialog>

    <v-dialog v-model="policyDialog.visible" width="860" persistent>
      <v-card class="panel-modal pa-4">
        <div class="d-flex align-center justify-space-between mb-4">
          <div>
            <div class="text-h6 font-weight-bold text-grey-lighten-3">配置文件设置</div>
            <div class="text-caption text-grey mt-1">{{ policyDialog.title }}</div>
          </div>
          <v-btn icon="mdi-close" variant="text" @click="closePolicyDialog" />
        </div>

        <v-alert type="info" variant="tonal" density="compact" class="mb-4">
          这里保存当前节点的策略覆盖，例如让这个节点走 IPv6 优先出口。发布节点时会与全局配置融合；inbounds 仍由服务入口自动生成。Preview 阶段每个节点只允许一份策略覆盖。
        </v-alert>
        <v-alert v-if="policyDialog.error" type="error" variant="tonal" density="compact" class="mb-4">
          {{ policyDialog.error }}
        </v-alert>

        <div class="template-editor-shell policy-editor-shell">
          <div class="template-editor-toolbar">
            <v-chip size="small" color="primary" variant="tonal">entry-policy.json</v-chip>
            <span>节点级覆盖：只允许配置 dns / outbounds / route / experimental；inbounds 由服务入口自动生成。</span>
          </div>
          <div class="template-editor policy-editor">
            <div class="template-line-numbers" :style="{ transform: `translateY(-${policyDialog.scrollTop}px)` }">
              <span v-for="line in policyLineNumbers" :key="line">{{ line }}</span>
            </div>
            <div class="template-code-wrap">
              <pre
                class="template-highlight"
                :style="{ transform: `translate(${-policyDialog.scrollLeft}px, ${-policyDialog.scrollTop}px)` }"
                v-html="highlightedPolicyConfig"
              />
              <textarea
                v-model="policyDialog.configText"
                class="template-code-input policy-code-input"
                spellcheck="false"
                @scroll="syncPolicyEditorScroll"
              />
            </div>
          </div>
        </div>

        <div class="d-flex justify-end mt-4" style="gap: 8px;">
          <v-btn class="tech-blue-btn" :disabled="!canSubmitPolicy" @click="saveInboundPolicySettings">提交</v-btn>
          <v-btn class="tech-grey-btn" @click="closePolicyDialog">关闭</v-btn>
        </div>
      </v-card>
    </v-dialog>

    <v-dialog v-model="policyCloseConfirm" width="420" persistent>
      <v-card class="panel-modal pa-4">
        <div class="text-h6 font-weight-bold text-grey-lighten-3 mb-2">配置文件有未提交变动</div>
        <div class="text-body-2 text-grey mb-5">
          你可以先提交保存，也可以不保存直接关闭，或回到编辑器继续调整。
        </div>
        <div class="d-flex justify-end" style="gap: 8px;">
          <v-btn class="tech-blue-btn" :disabled="!canSubmitPolicy" @click="saveInboundPolicySettings">保存并关闭</v-btn>
          <v-btn class="tech-grey-btn" @click="discardPolicyChanges">不保存</v-btn>
          <v-btn class="tech-grey-btn" @click="policyCloseConfirm = false">继续编辑</v-btn>
        </div>
      </v-card>
    </v-dialog>

    <v-dialog v-model="certificateDialog" width="720">
      <v-card class="panel-modal pa-4">
        <div class="d-flex align-center justify-space-between mb-4">
          <div>
            <div class="text-h6 font-weight-bold text-grey-lighten-3">{{ certForm.id ? '编辑证书申请' : '申请证书' }}</div>
            <div class="text-caption text-grey mt-1">这里申请的是可下发到 AnyTLS、Hysteria2 等协议使用的 TLS 证书。</div>
          </div>
          <v-btn icon="mdi-close" variant="text" @click="certificateDialog = false" />
        </div>

        <div class="certificate-source-grid mb-4">
          <button
            v-for="source in certSourceOptions"
            :key="source.value"
            type="button"
            class="certificate-source-option"
            :class="{ active: certForm.source === source.value }"
            @click="certForm.source = source.value"
          >
            <v-icon :icon="certificateSourceIcon(String(source.value))" size="22" />
            <span>{{ source.title }}</span>
          </button>
        </div>

        <v-row>
          <v-col cols="12">
            <v-alert density="compact" type="info" variant="tonal">
              域名只在这里填写；DNS Provider 只保存账号 API 凭据，同一个 Provider 可以用于同账号下多个域名。Reality 使用的是独立密钥，不在证书中心管理。
            </v-alert>
          </v-col>
          <v-col cols="12" md="6">
            <v-text-field v-model="certForm.name" label="名称" density="compact" variant="outlined" hide-details />
          </v-col>
          <v-col v-if="certForm.source !== 'manual'" cols="12" md="6">
            <v-select v-model="certForm.ca" :items="caOptions" label="CA" density="compact" variant="outlined" hide-details />
          </v-col>
          <v-col cols="12">
            <v-text-field v-model="certForm.domainsText" label="域名，逗号分隔" placeholder="*.example.com, example.com" density="compact" variant="outlined" hide-details />
          </v-col>
          <v-col v-if="certForm.source === 'acme-dns01'" cols="12">
            <v-select v-model="certForm.dnsProviderId" :items="dnsProviderOptions" label="DNS Provider" density="compact" variant="outlined" hide-details />
          </v-col>
          <template v-if="certForm.source === 'acme.sh'">
            <v-col cols="12">
              <v-alert density="compact" type="warning" variant="tonal">
                acme.sh 托管是高级模式：系统会调用容器内的 acme.sh。DNS API 填 dns_cf、dns_ali、dns_aws、dns_dp 等脚本名称；对应凭据需要在运行环境中可用。
              </v-alert>
            </v-col>
            <v-col cols="12" md="6">
              <v-text-field v-model="certForm.acmeHome" label="acme.sh 目录" placeholder="/root/.acme.sh" density="compact" variant="outlined" hide-details />
            </v-col>
            <v-col cols="12" md="6">
              <v-text-field v-model="certForm.acmeDnsApi" label="DNS API" placeholder="dns_cf" density="compact" variant="outlined" hide-details />
            </v-col>
          </template>
        </v-row>

        <div class="certificate-switch-row mt-4">
          <v-switch v-model="certForm.autoRenew" color="primary" label="自动续期" hide-details />
          <v-switch v-model="certForm.enable" color="primary" label="启用" hide-details />
        </div>

        <div class="d-flex justify-end mt-5" style="gap: 8px;">
          <v-btn class="tech-grey-btn" @click="certificateDialog = false">取消</v-btn>
          <v-btn class="tech-blue-btn" :disabled="!canSaveCertificate" @click="saveCertificate">保存申请</v-btn>
        </div>
      </v-card>
    </v-dialog>

    <v-dialog v-model="dnsProviderDialog" width="620">
      <v-card class="panel-modal pa-4">
        <div class="d-flex align-center justify-space-between mb-4">
          <div>
            <div class="text-h6 font-weight-bold text-grey-lighten-3">{{ dnsForm.id ? '编辑 DNS Provider' : '添加 DNS Provider' }}</div>
            <div class="text-caption text-grey mt-1">保存 DNS 服务商 API 凭据，用于自动申请和续签证书。</div>
          </div>
          <v-btn icon="mdi-close" variant="text" @click="dnsProviderDialog = false" />
        </div>
        <v-text-field v-model="dnsForm.name" label="名称" density="compact" variant="outlined" hide-details class="mb-3" />
        <div class="dns-provider-select-wrap mb-3">
          <div class="dns-provider-select-label">类型</div>
          <v-menu v-model="dnsProviderTypeMenu" :close-on-content-click="true" location="bottom start">
            <template #activator="{ props }">
              <button v-bind="props" type="button" class="dns-provider-select">
                <span class="dns-provider-option dns-provider-option-selection">
                  <img
                    v-if="providerIcon(dnsForm.type)"
                    class="dns-provider-icon"
                    :src="providerIcon(dnsForm.type)"
                    :alt="providerTypeName(dnsForm.type)"
                  />
                  <v-icon v-else icon="mdi-server-network" class="dns-provider-icon dns-provider-icon-fallback" />
                  <span class="dns-provider-name">{{ providerTypeName(dnsForm.type) }}</span>
                </span>
                <v-icon :icon="dnsProviderTypeMenu ? 'mdi-menu-up' : 'mdi-menu-down'" color="#6b7280" />
              </button>
            </template>
            <div class="dns-provider-menu">
              <button
                v-for="item in dnsTypeOptions"
                :key="item.value"
                type="button"
                class="dns-provider-option dns-provider-menu-item"
                :class="{ active: dnsForm.type === item.value }"
                @click="selectDNSProviderType(String(item.value))"
              >
                <img
                  v-if="providerIcon(String(item.value))"
                  class="dns-provider-icon"
                  :src="providerIcon(String(item.value))"
                  :alt="item.title"
                />
                <v-icon v-else icon="mdi-server-network" class="dns-provider-icon dns-provider-icon-fallback" />
                <span class="dns-provider-name">{{ item.title }}</span>
              </button>
            </div>
          </v-menu>
        </div>
        <div class="dns-credential-guide mb-3">
          <div class="d-flex align-start justify-space-between" style="gap: 12px;">
            <div>
              <div class="dns-credential-title">{{ selectedDNSCredentialGuide.title }}</div>
              <div class="dns-credential-desc">{{ selectedDNSCredentialGuide.description }}</div>
            </div>
            <v-btn
              v-if="selectedDNSCredentialGuide.template"
              class="tech-grey-btn text-none"
              size="small"
              prepend-icon="mdi-code-json"
              @click="applyDNSCredentialTemplate"
            >
              套用模板
            </v-btn>
          </div>
          <div v-if="selectedDNSCredentialGuide.fields.length" class="dns-credential-fields">
            <div v-for="field in selectedDNSCredentialGuide.fields" :key="field.key" class="dns-credential-field">
              <code>{{ field.key }}</code>
              <span>{{ field.help }}</span>
            </div>
          </div>
        </div>
        <v-textarea
          v-model="dnsForm.configText"
          label="API 凭据 JSON"
          rows="7"
          density="compact"
          variant="outlined"
          :error-messages="dnsProviderConfigValidation.error"
          class="mb-3 json-editor"
        />
        <v-switch v-model="dnsForm.enable" color="primary" label="启用" hide-details />
        <div class="d-flex justify-end mt-5" style="gap: 8px;">
          <v-btn class="tech-grey-btn" @click="dnsProviderDialog = false">取消</v-btn>
          <v-btn class="tech-blue-btn" :disabled="!canSaveDNSProvider" @click="saveDNSProvider">保存 Provider</v-btn>
        </div>
      </v-card>
    </v-dialog>

    <v-dialog v-model="certVersionDialog" width="680">
      <v-card class="panel-modal pa-4">
        <div class="d-flex align-center justify-space-between mb-4">
          <div>
            <div class="text-h6 font-weight-bold text-grey-lighten-3">手动证书版本</div>
            <div class="text-caption text-grey mt-1">导入已有 fullchain.pem 和 privkey.pem。</div>
          </div>
          <v-btn icon="mdi-close" variant="text" @click="certVersionDialog = false" />
        </div>
        <v-select v-model="certVersionForm.certificateId" :items="certificateOptions" label="证书" density="compact" variant="outlined" hide-details class="mb-3" />
        <v-text-field v-model="certVersionForm.fingerprint" label="指纹" density="compact" variant="outlined" hide-details class="mb-3" />
        <v-textarea v-model="certVersionForm.fullchainPemEncrypted" label="fullchain.pem" rows="5" density="compact" variant="outlined" hide-details class="mb-3" />
        <v-textarea v-model="certVersionForm.privateKeyPemEncrypted" label="privkey.pem" rows="5" density="compact" variant="outlined" hide-details />
        <div class="d-flex justify-end mt-5" style="gap: 8px;">
          <v-btn class="tech-grey-btn" @click="certVersionDialog = false">取消</v-btn>
          <v-btn class="tech-blue-btn" @click="saveCertificateVersion">保存版本</v-btn>
        </div>
      </v-card>
    </v-dialog>

    <v-dialog v-model="renderedConfigDialog.visible" width="820">
      <v-card class="panel-modal pa-4">
        <div class="d-flex align-center justify-space-between mb-3">
          <div>
            <div class="text-h6 font-weight-bold text-grey-lighten-3">{{ renderedConfigDialog.heading }}</div>
            <div class="text-caption text-grey mt-1">{{ renderedConfigDialog.title }}</div>
          </div>
          <v-btn icon="mdi-close" variant="text" @click="renderedConfigDialog.visible = false" />
        </div>
        <pre class="rendered-config-preview" v-html="highlightJSON(renderedConfigDialog.content)" />
      </v-card>
    </v-dialog>

    <!-- 服务入口用户绑定管理 Dialog -->
    <v-dialog v-model="inboundUsersDialog.visible" width="700">
      <v-card class="panel-modal pa-4">
        <div class="d-flex align-center justify-space-between mb-4">
          <div>
            <div class="text-h6 font-weight-bold text-grey-lighten-3">管理授权用户</div>
            <div class="text-caption text-grey mt-1">
              服务入口: {{ inboundUsersDialog.inbound ? `${nodeName(inboundUsersDialog.inbound.nodeId)}:${inboundUsersDialog.inbound.listenPort}` : '-' }}
            </div>
          </div>
          <v-btn icon="mdi-close" variant="text" @click="inboundUsersDialog.visible = false" />
        </div>

        <v-divider class="mb-4" />

        <!-- 绑定用户列表 -->
        <div class="card-section-title mb-2">已绑定用户列表</div>
        <v-table density="compact" class="mb-4">
          <thead>
            <tr>
              <th>用户名</th>
              <th>关联统一用户</th>
              <th>密码</th>
              <th class="text-right">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="user in currentInboundUsers" :key="user.id">
              <td>{{ user.name }}</td>
              <td>{{ clientName(user.clientId) }}</td>
              <td>
                <span class="mr-2">{{ showUserPasswords[user.id] ? user.password : '••••••••' }}</span>
                <v-btn
                  :icon="showUserPasswords[user.id] ? 'mdi-eye-off' : 'mdi-eye'"
                  variant="text"
                  size="x-small"
                  density="comfortable"
                  @click="togglePasswordVisibility(user.id)"
                />
                <v-btn
                  icon="mdi-content-copy"
                  variant="text"
                  size="x-small"
                  density="comfortable"
                  @click="copyText(user.password)"
                />
              </td>
              <td class="text-right">
                <v-btn
                  icon="mdi-trash-can-outline"
                  size="small"
                  variant="text"
                  color="error"
                  @click="deleteInboundUser(user.id)"
                />
              </td>
            </tr>
            <tr v-if="currentInboundUsers.length === 0">
              <td colspan="4" class="text-center text-grey py-4">暂无授权用户。请在下方录入新增。</td>
            </tr>
          </tbody>
        </v-table>

        <v-divider class="my-4" />

        <!-- 新增绑定折叠表单 -->
        <v-expansion-panels v-model="inboundUsersDialog.panel">
          <v-expansion-panel value="add">
            <v-expansion-panel-title class="font-weight-bold text-grey-lighten-2 py-2">
              <v-icon icon="mdi-plus" class="mr-2" color="primary" />
              新增用户绑定
            </v-expansion-panel-title>
            <v-expansion-panel-text>
              <v-row class="mt-2">
                <v-col cols="12" md="6" class="py-1">
                  <v-select
                    v-model="inboundUserForm.clientId"
                    :items="clientOptions"
                    label="选择关联用户"
                    density="compact"
                    variant="outlined"
                    hide-details
                  />
                </v-col>
                <v-col cols="12" md="6" class="py-1">
                  <v-text-field
                    v-model="inboundUserForm.name"
                    label="认证名称"
                    density="compact"
                    variant="outlined"
                    hide-details
                  />
                </v-col>
                <v-col cols="12" class="py-1">
                  <v-text-field
                    v-model="inboundUserForm.password"
                    label="密码"
                    density="compact"
                    variant="outlined"
                    hide-details
                  />
                </v-col>
              </v-row>
              <div class="d-flex justify-end mt-4">
                <v-btn
                  class="tech-blue-btn text-none"
                  :disabled="!canSaveInboundUser"
                  @click="saveInboundUser"
                >
                  保存并绑定
                </v-btn>
              </div>
            </v-expansion-panel-text>
          </v-expansion-panel>
        </v-expansion-panels>
      </v-card>
    </v-dialog>

    <QrCode
      v-model="qrDialog.visible"
      :visible="qrDialog.visible"
      :token="qrDialog.token"
      :client="qrDialog.client"
      :settings="settings"
      @close="closeQrCode"
    />
  </v-container>
</template>

<script setup lang="ts">
import alidnsIcon from '@/assets/dns-providers/alidns.svg'
import awsRoute53Icon from '@/assets/dns-providers/aws-route53.svg'
import cloudflareIcon from '@/assets/dns-providers/cloudflare.svg'
import tencentDnspodIcon from '@/assets/dns-providers/tencent-dnspod.svg'
import HttpUtils from '@/plugins/httputil'
import Data from '@/store/modules/data'
import QrCode from '@/layouts/modals/QrCode.vue'
import { push } from 'notivue'
import { computed, onMounted, onUnmounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

type SelectItem = { title: string, value: number | string }

const route = useRoute()
const router = useRouter()
const tab = ref(initialTab())
const loading = ref(false)
const nodeDrawer = ref(false)
const inboundDialog = ref(false)
const certificateDialog = ref(false)
const dnsProviderDialog = ref(false)
const dnsProviderTypeMenu = ref(false)
const certVersionDialog = ref(false)
const policyCloseConfirm = ref(false)
const guideDialog = ref(false)

const nodes = ref<any[]>([])
const dnsProviders = ref<any[]>([])
const certificates = ref<any[]>([])
const distributedInbounds = ref<any[]>([])
const inboundUsers = ref<any[]>([])
const configVersions = ref<any[]>([])
const subscriptions = ref<any[]>([])
const lastToken = ref('')
const subscriptionClientId = ref<number | null>(null)
const issuingCertificateId = ref(0)
const renderingInboundId = ref(0)
const currentUnix = ref(Math.floor(Date.now() / 1000))
const publishingNodeIds = ref<number[]>([])
const nodeSyncProgress = ref<Record<number, number>>({})
const templateError = ref('')
const templateEditorScrollTop = ref(0)
const templateEditorScrollLeft = ref(0)
const selectedProtocolTemplate = ref('anytls-inbound')
const selectedInboundIds = ref<number[]>([])
const renderedConfigDialog = reactive({ visible: false, heading: '服务入口 JSON', title: '', content: '' })
const policyDialog = reactive({
  visible: false,
  title: '',
  inbound: null as any,
  configText: '{}',
  originalText: '{}',
  scrollTop: 0,
  scrollLeft: 0,
  error: '',
})
let healthTimer: ReturnType<typeof setInterval> | null = null

const DNS_CREDENTIAL_KEEP_ENCRYPTED = '__SUI_KEEP_ENCRYPTED__'
const dnsTypes = ['cloudflare', 'route53', 'aliyun', 'dnspod']
const dnsTypeOptions = [
  { title: 'Cloudflare', value: 'cloudflare' },
  { title: 'AWS', value: 'route53' },
  { title: 'AliCloud', value: 'aliyun' },
  { title: 'Tencent DNSPod', value: 'dnspod' },
]
const dnsCredentialGuides: Record<string, {
  title: string
  description: string
  fields: Array<{ key: string; help: string }>
  template: Record<string, string> | null
}> = {
  cloudflare: {
    title: 'Cloudflare API Token',
    description: '用于 acme.sh 的 dns_cf。这里只保存账号 API 凭据，域名在“申请证书”里填写；DNS-01 只会临时写 TXT 记录，不要求开启橙云代理。',
    fields: [
      { key: 'CF_Token', help: '必填，填入你的 Cloudflare API Token。' },
    ],
    template: {
      CF_Token: '在这里粘贴你的 Cloudflare API Token',
    },
  },
  aliyun: {
    title: 'AliCloud AccessKey',
    description: '用于 acme.sh 的 dns_ali。这里只保存账号 API 凭据，域名在“申请证书”里填写；请使用有 DNS 解析记录管理权限的 RAM AccessKey。',
    fields: [
      { key: 'Ali_Key', help: '必填，阿里云 AccessKey ID。' },
      { key: 'Ali_Secret', help: '必填，阿里云 AccessKey Secret。' },
    ],
    template: {
      Ali_Key: '在这里粘贴阿里云 AccessKey ID',
      Ali_Secret: '在这里粘贴阿里云 AccessKey Secret',
    },
  },
  route53: {
    title: 'AWS Access Key',
    description: '用于 acme.sh 的 dns_aws。这里只保存账号 API 凭据，域名在“申请证书”里填写；请使用有 DNS 解析记录修改权限的 IAM Access Key。',
    fields: [
      { key: 'AWS_ACCESS_KEY_ID', help: '必填，AWS Access Key ID。' },
      { key: 'AWS_SECRET_ACCESS_KEY', help: '必填，AWS Secret Access Key。' },
    ],
    template: {
      AWS_ACCESS_KEY_ID: '在这里粘贴 AWS Access Key ID',
      AWS_SECRET_ACCESS_KEY: '在这里粘贴 AWS Secret Access Key',
    },
  },
  dnspod: {
    title: 'Tencent DNSPod API Key',
    description: '用于 acme.sh 的 dns_dp。这里只保存账号 API 凭据，域名在“申请证书”里填写；请填写 DNSPod 控制台生成的 API ID 和 API Key。',
    fields: [
      { key: 'DP_Id', help: '必填，DNSPod API ID。' },
      { key: 'DP_Key', help: '必填，DNSPod API Key。' },
    ],
    template: {
      DP_Id: '在这里粘贴 DNSPod API ID',
      DP_Key: '在这里粘贴 DNSPod API Key',
    },
  },
}
const certSourceOptions: SelectItem[] = [
  { title: 'DNS API 自动申请', value: 'acme-dns01' },
  { title: 'acme.sh 托管', value: 'acme.sh' },
]
const caOptions = ['letsencrypt', 'zerossl']
const inboundProtocolOptions = [
  { title: 'AnyTLS', value: 'anytls' },
  { title: 'Hysteria2', value: 'hysteria2' },
  { title: 'Hysteria', value: 'hysteria' },
  { title: 'TUIC', value: 'tuic' },
  { title: 'Trojan', value: 'trojan' },
  { title: 'VLESS', value: 'vless' },
  { title: 'VMess', value: 'vmess' },
  { title: 'Naive', value: 'naive' },
  { title: 'Shadowsocks', value: 'shadowsocks' },
  { title: 'ShadowTLS', value: 'shadowtls' },
]
const protocolTemplates = [
  {
    title: 'AnyTLS 入站',
    value: 'anytls-inbound',
    description: '用于远端节点服务入口。域名、证书路径、端口和用户会在发布节点配置时自动替换。',
    content: stringifyJSON({
      type: 'anytls',
      tag: 'anytls-us-01-443',
      listen: '::',
      listen_port: 443,
      tls: {
        enabled: true,
        server_name: 'us.example.com',
        certificate_path: '/usr/local/s-ui-agent/certs/us.example.com/fullchain.pem',
        key_path: '/usr/local/s-ui-agent/certs/us.example.com/privkey.pem',
      },
      users: [
        {
          name: 'alice',
          password: 'change-me',
        },
      ],
    }),
  },
  {
    title: 'AnyTLS 出站',
    value: 'anytls-outbound',
    description: '用于订阅侧客户端配置。server/server_port/password/server_name 来自已授权的集群服务入口。',
    content: stringifyJSON({
      type: 'anytls',
      tag: 'us-alice',
      server: 'us.example.com',
      server_port: 443,
      password: 'change-me',
      tls: {
        enabled: true,
        server_name: 'us.example.com',
      },
    }),
  },
  {
    title: 'Hysteria2 入站',
    value: 'hysteria2-inbound',
    description: '用于 UDP 高性能入口。第一版可复制到服务入口的高级 JSON 中，再按实际域名、证书路径和用户调整。',
    content: stringifyJSON(defaultAdvancedInbound('hysteria2')),
  },
  {
    title: 'Hysteria 入站',
    value: 'hysteria-inbound',
    description: '用于旧版 Hysteria 入站。适合迁移已有 sing-box 配置时作为起点。',
    content: stringifyJSON(defaultAdvancedInbound('hysteria')),
  },
  {
    title: 'TUIC 入站',
    value: 'tuic-inbound',
    description: '用于 TUIC 服务入口。uuid/password 等字段需要按用户体系继续细化。',
    content: stringifyJSON(defaultAdvancedInbound('tuic')),
  },
  {
    title: 'Trojan 入站',
    value: 'trojan-inbound',
    description: '用于 Trojan TLS 入站。可作为高级 JSON 的起点。',
    content: stringifyJSON(defaultAdvancedInbound('trojan')),
  },
  {
    title: 'VLESS 入站',
    value: 'vless-inbound',
    description: '用于 VLESS 入站。Reality 是 VLESS/Trojan 的 TLS 安全模式，不作为单独协议；可在高级 JSON 中配置。',
    content: stringifyJSON(defaultAdvancedInbound('vless')),
  },
  {
    title: 'VLESS + Reality 入站',
    value: 'vless-reality-inbound',
    description: 'Reality 不作为单独协议；选择 VLESS 后，把这个模板放到高级 JSON 里即可。',
    content: stringifyJSON(defaultVLESSRealityInbound()),
  },
  {
    title: 'VMess 入站',
    value: 'vmess-inbound',
    description: '用于 VMess 入站。适合作为从本地节点迁移到集群节点的模板。',
    content: stringifyJSON(defaultAdvancedInbound('vmess')),
  },
  {
    title: 'Naive 入站',
    value: 'naive-inbound',
    description: '用于 NaiveProxy 入站。需要 TLS 证书和用户名密码。',
    content: stringifyJSON(defaultAdvancedInbound('naive')),
  },
  {
    title: 'Shadowsocks 入站',
    value: 'shadowsocks-inbound',
    description: '用于 Shadowsocks 入站。2022 方法的密码格式上线前还需要按 sing-box 校验。',
    content: stringifyJSON(defaultAdvancedInbound('shadowsocks')),
  },
  {
    title: 'ShadowTLS 入站',
    value: 'shadowtls-inbound',
    description: '用于 ShadowTLS 入站。通常通过握手目标伪装，不依赖证书中心下发证书。',
    content: stringifyJSON(defaultAdvancedInbound('shadowtls')),
  },
  {
    title: 'IPv6 优先 DNS/路由骨架',
    value: 'prefer-ipv6-base',
    description: '适合作为全局默认配置的 dns/outbounds/route 参考，和内置默认值保持一致。',
    content: stringifyJSON({
      dns: defaultFullNodeTemplate().dns,
    outbounds: defaultFullNodeTemplate().outbounds,
    route: defaultFullNodeTemplate().route,
    experimental: defaultFullNodeTemplate().experimental,
    }),
  },
]
const clusterTabs = [
  { title: '概览', value: 'nodes' },
  { title: '证书中心', value: 'certificates' },
  { title: '服务入口', value: 'inbounds' },
  { title: '全局配置', value: 'versions' },
]
const visibleTabs = computed(() => {
  if (route.path === '/subscriptions') return [{ title: '订阅', value: 'subscriptions' }]
  return clusterTabs
})
const showInlineTabs = computed(() => route.path === '/distributed' && visibleTabs.value.length > 1)
const pageTitle = computed(() => {
  if (route.path === '/subscriptions') return '订阅'
  return '集群管理'
})

const nodeForm = reactive<any>(newNode())
const dnsForm = reactive<any>(newDNSProvider())
const certForm = reactive<any>(newCertificate())
const certVersionForm = reactive<any>(newCertificateVersion())
const inboundForm = reactive<any>(newInbound())
const inboundUserForm = reactive<any>(newInboundUser())
const originalInboundForm = ref<any | null>(null)
const nodeTemplateForm = reactive<any>(defaultNodeTemplateForm())
const settings = ref<any>({})

const qrDialog = ref({
  visible: false,
  token: '',
  client: null as any
})

const showQrCode = (token: string, client: any) => {
  qrDialog.value.token = token
  qrDialog.value.client = client
  qrDialog.value.visible = true
}

const closeQrCode = () => {
  qrDialog.value.visible = false
}

const getClientById = (clientId: number | null) => {
  if (clientId === null) return null
  return clients.value.find(c => c.id === clientId) || null
}

const inboundUsersDialog = reactive<any>({
  visible: false,
  inbound: null,
  panel: []
})
const showUserPasswords = ref<Record<number, boolean>>({})

const currentInboundUsers = computed(() => {
  if (!inboundUsersDialog.inbound) return []
  return inboundUsers.value.filter(u => u.inboundId === inboundUsersDialog.inbound.id)
})

const clients = computed((): any[] => Data().clients || [])
const nodeOptions = computed<SelectItem[]>(() => nodes.value.map(n => ({ title: `${n.name} (${n.code})`, value: n.id })))
const dnsProviderOptions = computed<SelectItem[]>(() => dnsProviders.value.map(p => ({ title: `${p.name} (${p.type})`, value: p.id })))
const certificateOptions = computed<SelectItem[]>(() => certificates.value.map(c => ({ title: c.name, value: c.id })))
const inboundOptions = computed<SelectItem[]>(() => distributedInbounds.value.map(i => ({ title: `${nodeName(i.nodeId)}:${i.listenPort}`, value: i.id })))
const clusterClients = computed(() => clients.value.filter(allowsClusterAccess))
const clientOptions = computed<SelectItem[]>(() => clusterClients.value.map(c => ({ title: c.name, value: c.id })))
const selectedDNSCredentialGuide = computed(() => dnsCredentialGuides[dnsForm.type] || dnsCredentialGuides.cloudflare)
const dnsProviderConfigValidation = computed(() => validateDNSProviderConfigText(dnsForm.configText, dnsForm.type))
const canSaveDNSProvider = computed(() => {
  return dnsForm.name.trim().length > 0 && dnsForm.type && dnsProviderConfigValidation.value.valid
})
const issuedCertificates = computed(() => certificates.value.filter(c => !!c.activeVersionId))
const autoRenewCertificateCount = computed(() => certificates.value.filter(c => Boolean(c.autoRenew)).length)
const expiringCertificateCount = computed(() => certificates.value.filter(c => {
  const value = c.notAfter
  if (!value) return false
  const expiresAt = value * 1000
  const diff = expiresAt - Date.now()
  return diff > 0 && diff <= 10 * 24 * 60 * 60 * 1000
}).length)
const renderedInboundCount = computed(() => distributedInbounds.value.filter(i => !!i.renderedConfigJson).length)
const publishedNodeCount = computed(() => new Set(configVersions.value.map(v => v.nodeId)).size)
const agentReadyCount = computed(() => nodes.value.filter(n => nodeHealth(n).agentOnline).length)
const healthyNodeCount = computed(() => nodes.value.filter(n => nodeHealth(n).state === 'healthy').length)
const clusterStats = computed(() => [
  { label: '节点', value: nodes.value.length, icon: 'mdi-server-network' },
  { label: 'Agent 在线', value: agentReadyCount.value, icon: 'mdi-lan-check' },
  { label: '健康节点', value: healthyNodeCount.value, icon: 'mdi-heart-pulse' },
  { label: '已发布配置', value: publishedNodeCount.value, icon: 'mdi-rocket-launch-outline' },
])
const protocolTemplateOptions = computed(() => protocolTemplates.map(item => ({ title: item.title, value: item.value })))
const activeProtocolTemplate = computed(() => {
  return protocolTemplates.find(item => item.value === selectedProtocolTemplate.value) || protocolTemplates[0]
})
const templateLineNumbers = computed(() => {
  const count = Math.max(1, nodeTemplateForm.config.split('\n').length)
  return Array.from({ length: count }, (_, index) => index + 1)
})
const highlightedNodeTemplate = computed(() => highlightJSON(nodeTemplateForm.config))
const policyLineNumbers = computed(() => {
  const count = Math.max(1, policyDialog.configText.split('\n').length)
  return Array.from({ length: count }, (_, index) => index + 1)
})
const highlightedPolicyConfig = computed(() => highlightJSON(policyDialog.configText))
const policyHasChanges = computed(() => normalizeJSONText(policyDialog.configText) !== normalizeJSONText(policyDialog.originalText))
const policyValidation = computed(() => validatePolicyConfigText(policyDialog.configText))
const canSubmitPolicy = computed(() => policyHasChanges.value && policyValidation.value.valid)
const setupSteps = computed(() => [
  {
    key: 'node',
    title: '创建远端节点',
    caption: nodes.value.length > 0 ? `${nodes.value.length} 个节点已创建` : '填写节点名称、代号、地区和域名',
    icon: 'mdi-server-plus',
    done: nodes.value.length > 0,
    tab: 'nodes',
    action: 'node',
  },
  {
    key: 'certificate',
    title: '准备证书',
    caption: issuedCertificates.value.length > 0 ? `${issuedCertificates.value.length} 张证书可用` : '申请证书后绑定到服务入口',
    icon: 'mdi-certificate-outline',
    done: issuedCertificates.value.length > 0,
    tab: 'certificates',
  },
  {
    key: 'inbound',
    title: '创建服务入口',
    caption: distributedInbounds.value.length > 0 ? `${distributedInbounds.value.length} 个入口已创建` : '为节点选择协议、域名、端口和证书',
    icon: 'mdi-router-wireless',
    done: distributedInbounds.value.length > 0,
    tab: 'inbounds',
  },
  {
    key: 'user',
    title: '绑定用户',
    caption: inboundUsers.value.length > 0 ? `${inboundUsers.value.length} 个授权已保存` : '把统一用户授权到服务入口',
    icon: 'mdi-account-key-outline',
    done: inboundUsers.value.length > 0,
    tab: 'inbounds',
  },
  {
    key: 'publish',
    title: '发布节点配置',
    caption: publishedNodeCount.value > 0 ? `${publishedNodeCount.value} 个节点已有版本` : '生成并发布可被 Agent 拉取的配置',
    icon: 'mdi-rocket-launch-outline',
    done: publishedNodeCount.value > 0,
    tab: 'versions',
  },
])
const completedSetupSteps = computed(() => setupSteps.value.filter(step => step.done).length)
const nextSetupStep = computed(() => setupSteps.value.find(step => !step.done) || setupSteps.value[setupSteps.value.length - 1])
const canSaveNode = computed(() => nodeForm.name.trim().length > 0 && nodeForm.code.trim().length > 0)
const canSaveInbound = computed(() => {
  if (!inboundForm.nodeId || !inboundForm.listenPort || !inboundForm.protocol) return false
  if (protocolNeedsCertificate.value && !inboundForm.certificateId) return false
  if (String(inboundForm.publicHost || '').trim().length === 0) return false
  if (String(inboundForm.listen || '').trim().length === 0) return false
  if (!validateAdvancedInboundJSON(false)) return false
  if (inboundForm.id && originalInboundForm.value && !inboundChanged()) return false
  return true
})
const protocolNeedsCertificate = computed(() => {
  return ['anytls', 'hysteria', 'hysteria2', 'tuic', 'trojan', 'vless', 'vmess', 'naive'].includes(inboundForm.protocol)
})
const canSaveCertificate = computed(() => {
  if (certForm.name.trim().length === 0) return false
  if (certForm.domainsText.split(',').map((d: string) => d.trim()).filter(Boolean).length === 0) return false
  if (certForm.source === 'acme-dns01' && !certForm.dnsProviderId) return false
  if (certForm.source === 'acme.sh' && (!String(certForm.acmeHome || '').trim() || !String(certForm.acmeDnsApi || '').trim())) return false
  return true
})
const canSaveInboundUser = computed(() => {
  return !!inboundUserForm.inboundId && inboundUserForm.name.trim().length > 0 && inboundUserForm.password.trim().length > 0
})
const canSaveNodeTemplates = computed(() => validateNodeTemplates(false))
const allInboundsSelected = computed(() => {
  return distributedInbounds.value.length > 0 && selectedInboundIds.value.length === distributedInbounds.value.length
})
const someInboundsSelected = computed(() => {
  return selectedInboundIds.value.length > 0 && selectedInboundIds.value.length < distributedInbounds.value.length
})

onMounted(() => {
  loadAll()
  healthTimer = setInterval(() => {
    currentUnix.value = Math.floor(Date.now() / 1000)
    loadNodes()
  }, 15000)
})

onUnmounted(() => {
  if (healthTimer) clearInterval(healthTimer)
})

watch(() => [route.path, route.query.tab], () => {
  const nextTab = initialTab()
  if (tab.value !== nextTab) tab.value = nextTab
})

watch(tab, value => {
  if (route.path !== '/distributed') return
  if (route.query.tab === value) return
  router.replace({ query: { ...route.query, tab: value } })
})

watch(distributedInbounds, () => {
  if (!inboundUserForm.inboundId && distributedInbounds.value.length === 1) {
    inboundUserForm.inboundId = distributedInbounds.value[0].id
  }
}, { immediate: true })

watch(() => inboundUserForm.clientId, () => {
  const client = clients.value.find(c => String(c.id) === String(inboundUserForm.clientId))
  if (client && !inboundUserForm.name) {
    inboundUserForm.name = client.name
  }
})

watch(() => inboundForm.protocol, protocol => {
  const current = parseJSON(inboundForm.advancedText, null)
  if (protocol === 'anytls') {
    if (current?.type && current.type !== 'anytls') inboundForm.advancedText = '{}'
    return
  }
  if (!current || current.type !== protocol) {
    inboundForm.advancedText = stringifyJSON(defaultAdvancedInbound(protocol))
  }
})

watch(() => dnsForm.type, (type, oldType) => {
  if (!dnsProviderDialog.value || dnsForm.id) return
  if (shouldReplaceDNSCredentialTemplate(dnsForm.configText, oldType)) {
    dnsForm.configText = providerCredentialTemplateText(type)
  }
})

watch(() => policyDialog.configText, () => {
  if (!policyDialog.visible) return
  policyDialog.error = policyValidation.value.valid ? '' : policyValidation.value.error
})

async function loadAll() {
  loading.value = true
  await Data().loadData()
  await Promise.all([
    loadNodes(),
    loadDNSProviders(),
    loadCertificates(),
    loadDistributedInbounds(),
    loadInboundUsers(),
    loadConfigVersions(),
    loadSubscriptions(),
    loadNodeTemplates(),
    loadSettings(),
  ])
  loading.value = false
}

async function loadSettings() {
  const msg = await HttpUtils.get('api/setting')
  if (msg.success) settings.value = msg.obj || {}
}

async function loadNodes() {
  const msg = await HttpUtils.get('api/nodes')
  if (msg.success) nodes.value = msg.obj || []
}

async function loadDNSProviders() {
  const msg = await HttpUtils.get('api/dnsProviders')
  if (msg.success) dnsProviders.value = msg.obj || []
}

async function loadCertificates() {
  const msg = await HttpUtils.get('api/certificates')
  if (msg.success) certificates.value = msg.obj || []
}

async function loadDistributedInbounds() {
  const msg = await HttpUtils.get('api/distributedInbounds')
  if (msg.success) {
    distributedInbounds.value = msg.obj || []
    const validIds = new Set(distributedInbounds.value.map(inbound => Number(inbound.id)))
    selectedInboundIds.value = selectedInboundIds.value.filter(id => validIds.has(Number(id)))
  }
}

async function loadInboundUsers() {
  const msg = await HttpUtils.get('api/inboundUsers')
  if (msg.success) inboundUsers.value = msg.obj || []
}

async function loadConfigVersions() {
  const msg = await HttpUtils.get('api/configVersions')
  if (msg.success) configVersions.value = msg.obj || []
}

async function loadSubscriptions() {
  const msg = await HttpUtils.get('api/subscriptions')
  if (msg.success) subscriptions.value = msg.obj || []
}

async function loadNodeTemplates() {
  const msg = await HttpUtils.get('api/nodeConfigTemplates')
  if (!msg.success || !msg.obj) return
  nodeTemplateForm.config = stringifyJSON(composeNodeConfigTemplate(msg.obj))
}

async function saveNode() {
  if (!canSaveNode.value) return
  const msg = await HttpUtils.post('api/saveNode', toForm(nodeForm))
  if (!msg.success) return
  resetNode()
  nodeDrawer.value = false
  await loadNodes()
}

async function saveDNSProvider() {
  if (!canSaveDNSProvider.value) return
  const payload = { ...dnsForm, config: dnsProviderConfigValidation.value.value }
  await HttpUtils.post('api/saveDNSProvider', toForm(payload))
  resetDNSProvider()
  dnsProviderDialog.value = false
  await loadDNSProviders()
}

async function saveCertificate() {
  const domains = certForm.domainsText.split(',').map((d: string) => d.trim()).filter(Boolean)
  const payload = {
    ...certForm,
    wildcard: domains.some((domain: string) => domain.startsWith('*.')),
    domains: JSON.stringify(domains),
    dnsProviderId: certForm.source === 'acme-dns01' ? certForm.dnsProviderId : 0,
    config: JSON.stringify(certificateConfig()),
  }
  await HttpUtils.post('api/saveCertificate', toForm(payload))
  resetCertificate()
  certificateDialog.value = false
  await loadCertificates()
}

async function saveCertificateVersion() {
  await HttpUtils.post('api/saveCertificateVersion', toForm(certVersionForm))
  certVersionDialog.value = false
  await loadCertificates()
}

async function issueCertificate(id: number) {
  issuingCertificateId.value = id
  try {
    await HttpUtils.post('api/issueCertificate', toForm({ id }))
    await loadCertificates()
  } finally {
    issuingCertificateId.value = 0
  }
}

async function saveDistributedInbound() {
  if (!canSaveInbound.value) return
  const payload = {
    ...inboundForm,
    advancedOverridesJson: inboundForm.protocol === 'anytls' ? '{}' : compactJSONString(inboundForm.advancedText),
    policyOverridesJson: inboundForm.policyOverridesJson || '{}',
  }
  const msg = await HttpUtils.post('api/saveDistributedInbound', toForm(payload))
  if (!msg.success) return
  resetInbound()
  inboundDialog.value = false
  await loadDistributedInbounds()
}

async function deleteSelectedInbounds() {
  if (selectedInboundIds.value.length === 0) return
  if (!window.confirm(`确定删除选中的 ${selectedInboundIds.value.length} 个服务入口吗？`)) return
  for (const id of selectedInboundIds.value) {
    await HttpUtils.post('api/deleteDistributedInbound', toForm({ id }))
  }
  selectedInboundIds.value = []
  await loadDistributedInbounds()
}

async function deleteNode(id: number) {
  if (!window.confirm("确定删除该节点吗？此操作将彻底删除该节点及其所有的服务入口，且不可恢复！")) return
  const msg = await HttpUtils.post('api/deleteNode', toForm({ id }))
  if (msg.success) {
    await loadAll()
  }
}

async function deleteSubscription(id: number) {
  if (!window.confirm("确定删除该订阅 Token 吗？删除后此 Token 对应的所有客户端将无法再次拉取配置！")) return
  const msg = await HttpUtils.post(`api/deleteSubscription?id=${id}`, null)
  if (msg.success) {
    await loadSubscriptions()
  }
}

async function saveInboundUser() {
  const msg = await HttpUtils.post('api/saveInboundUser', toForm(inboundUserForm))
  if (!msg.success) return
  const currentInboundId = inboundUserForm.inboundId
  Object.assign(inboundUserForm, newInboundUser())
  inboundUserForm.inboundId = currentInboundId
  inboundUsersDialog.panel = []
  await loadInboundUsers()
}

function openInboundUsersDialog(inbound: any) {
  inboundUsersDialog.inbound = inbound
  inboundUsersDialog.panel = []
  Object.assign(inboundUserForm, newInboundUser())
  inboundUserForm.inboundId = inbound.id
  inboundUsersDialog.visible = true
}

function togglePasswordVisibility(userId: number) {
  showUserPasswords.value[userId] = !showUserPasswords.value[userId]
}

async function deleteInboundUser(userId: number) {
  if (!window.confirm("确定删除该用户的绑定授权吗？")) return
  const msg = await HttpUtils.post(`api/deleteInboundUser?id=${userId}`, null)
  if (msg.success) {
    await loadInboundUsers()
  }
}

function copyText(text: string) {
  navigator.clipboard.writeText(text).then(() => {
    push.success({ message: "复制成功" })
  }).catch(() => {
    push.error({ message: "复制失败" })
  })
}

async function renderInbound(id: number) {
  renderingInboundId.value = id
  try {
    const msg = await HttpUtils.post('api/renderDistributedAnyTLSInbound', toForm({ id }))
    if (msg.success) await loadDistributedInbounds()
  } finally {
    renderingInboundId.value = 0
  }
}

function openRenderedConfig(inbound: any) {
  renderedConfigDialog.heading = '入口 JSON 预览'
  renderedConfigDialog.title = `${nodeName(inbound.nodeId)} · ${inbound.protocol} · ${inbound.publicHost || nodeHost(inbound.nodeId)}:${inbound.listenPort}`
  renderedConfigDialog.content = stringifyJSON(parseJSONValue(inbound.renderedConfigJson, { message: '尚未生成，请先点击生成配置。' }))
  renderedConfigDialog.visible = true
}

function openNodeConfig(nodeId: number) {
  const version = latestVersion(nodeId)
  renderedConfigDialog.heading = '最终配置预览'
  renderedConfigDialog.title = version
    ? `${nodeName(nodeId)} · v${version.version} · ${version.status}`
    : `${nodeName(nodeId)} · 尚未发布`
  renderedConfigDialog.content = stringifyJSON(parseJSONValue(version?.contentJson, {
    message: '这个节点还没有发布配置。请先在概览里点击发布配置，或生成服务入口后再发布节点。',
  }))
  renderedConfigDialog.visible = true
}

async function publishNode(nodeId: number) {
  if (publishingNodeIds.value.includes(nodeId)) return

  publishingNodeIds.value.push(nodeId)
  nodeSyncProgress.value[nodeId] = 10

  const progressTimer = setInterval(() => {
    if (nodeSyncProgress.value[nodeId] !== undefined && nodeSyncProgress.value[nodeId] < 90) {
      nodeSyncProgress.value[nodeId] += Math.floor(Math.random() * 12) + 5
      if (nodeSyncProgress.value[nodeId] > 90) {
        nodeSyncProgress.value[nodeId] = 90
      }
    }
  }, 350)

  try {
    await HttpUtils.post('api/publishNodeConfig', toForm({ nodeId }))
    await loadConfigVersions()

    let checkCount = 0
    const checkSync = async () => {
      await loadNodes()
      const node = nodes.value.find(n => n.id === nodeId)
      if (node) {
        const published = String(node.publishedSha256 || '').trim()
        const applied = String(node.appliedSha256 || '').trim()

        if ((published === applied && published !== '') || checkCount > 15) {
          clearInterval(progressTimer)
          nodeSyncProgress.value[nodeId] = 100
          setTimeout(() => {
            publishingNodeIds.value = publishingNodeIds.value.filter(id => id !== nodeId)
            delete nodeSyncProgress.value[nodeId]
          }, 800)
          return
        }
      }
      checkCount++
      setTimeout(checkSync, 1000)
    }

    setTimeout(checkSync, 1000)
  } catch (err) {
    clearInterval(progressTimer)
    publishingNodeIds.value = publishingNodeIds.value.filter(id => id !== nodeId)
    delete nodeSyncProgress.value[nodeId]
  }
}

async function createSubscription() {
  if (!subscriptionClientId.value) return
  const msg = await HttpUtils.post('api/createSubscription', toForm({ clientId: subscriptionClientId.value }))
  if (msg.success) lastToken.value = msg.obj?.token || ''
  await loadSubscriptions()
}

async function saveNodeTemplates() {
  const parsed = parseNodeConfigTemplate(true)
  if (!parsed) return
  const msg = await HttpUtils.post('api/saveNodeConfigTemplates', toForm({
    log: JSON.stringify(parsed.log),
    dns: JSON.stringify(parsed.dns),
    outbounds: JSON.stringify(parsed.outbounds),
    route: JSON.stringify(parsed.route),
    experimental: JSON.stringify(parsed.experimental || {}),
  }))
  if (msg.success) await loadNodeTemplates()
}

function formatNodeTemplates() {
  const parsed = parseNodeConfigTemplate(true)
  if (!parsed) return
  nodeTemplateForm.config = stringifyJSON(composeNodeConfigTemplate(parsed))
}

function resetNodeTemplates() {
  Object.assign(nodeTemplateForm, defaultNodeTemplateForm())
  templateError.value = ''
}

function copyProtocolTemplate() {
  navigator.clipboard?.writeText(activeProtocolTemplate.value.content)
}

function openNodeDrawer(node?: any) {
  if (node) editNode(node)
  else resetNode()
  nodeDrawer.value = true
}

function editNode(node: any) {
  Object.assign(nodeForm, { ...newNode(), ...node })
}

function editDNSProvider(provider: any) {
  Object.assign(dnsForm, { ...newDNSProvider(), ...provider, configText: JSON.stringify(provider.config || {}, null, 2) })
  dnsProviderDialog.value = true
}

function editCertificate(cert: any) {
  const config = parseJSONValue(cert.config, {})
  Object.assign(certForm, {
    ...newCertificate(),
    ...cert,
    domainsText: formatDomains(cert.domains),
    ca: config.ca || 'letsencrypt',
    acmeHome: config.acmeHome || '/root/.acme.sh',
    acmeDnsApi: config.acmeDnsApi || '',
  })
  certificateDialog.value = true
}

function prepareCertVersion(cert: any) {
  Object.assign(certVersionForm, newCertificateVersion(), { certificateId: cert.id })
  certVersionDialog.value = true
}

function openCertificateDialog() {
  resetCertificate()
  certificateDialog.value = true
}

function openDNSProviderDialog() {
  resetDNSProvider()
  dnsProviderDialog.value = true
}

function editInbound(inbound: any) {
  Object.assign(inboundForm, {
    ...newInbound(),
    ...inbound,
    advancedText: stringifyJSON(parseJSONValue(inbound.advancedOverridesJson, defaultAdvancedInbound(inbound.protocol || 'anytls'))),
  })
  originalInboundForm.value = normalizedInbound(inboundForm)
  inboundDialog.value = true
}

function openNewInbound() {
  resetInbound()
  inboundDialog.value = true
}

function openInboundPolicySettings(inbound: any) {
  const version = latestVersion(inbound.nodeId)
  const config = parseJSONValue(version?.contentJson, {})
  const policy = parseJSONValue(inbound.policyOverridesJson, {})
  const templateConfig = parseJSONValue(nodeTemplateForm.config, {})
  const policyConfig = {
    dns: policy.dns || config.dns || templateConfig.dns || defaultFullNodeTemplate().dns,
    outbounds: policy.outbounds || config.outbounds || templateConfig.outbounds || defaultFullNodeTemplate().outbounds,
    route: policy.route || config.route || templateConfig.route || defaultFullNodeTemplate().route,
    experimental: policy.experimental || config.experimental || templateConfig.experimental || {},
  }
  const policyText = stringifyJSON(policyConfig)
  policyDialog.title = `${nodeName(inbound.nodeId)} · ${inbound.protocol} · ${inbound.publicHost || nodeHost(inbound.nodeId)}:${inbound.listenPort}`
  policyDialog.inbound = inbound
  policyDialog.configText = policyText
  policyDialog.originalText = policyText
  policyDialog.scrollTop = 0
  policyDialog.scrollLeft = 0
  policyDialog.error = ''
  policyDialog.visible = true
}

async function saveInboundPolicySettings() {
  if (!policyDialog.inbound) return
  const policy = parsePolicyConfigText(true)
  if (!policy) return
  const inbound = policyDialog.inbound
  const payload = {
    ...inbound,
    policyOverridesJson: JSON.stringify(policy),
    advancedOverridesJson: inbound.advancedOverridesJson || '{}',
    formValuesJson: inbound.formValuesJson || '{}',
  }
  const msg = await HttpUtils.post('api/saveDistributedInbound', toForm(payload))
  if (!msg.success) return
  policyCloseConfirm.value = false
  policyDialog.visible = false
  policyDialog.inbound = null
  await loadDistributedInbounds()
}

function closePolicyDialog() {
  if (policyHasChanges.value) {
    policyCloseConfirm.value = true
    return
  }
  policyDialog.visible = false
  policyDialog.inbound = null
  policyDialog.error = ''
}

function discardPolicyChanges() {
  policyCloseConfirm.value = false
  policyDialog.visible = false
  policyDialog.inbound = null
  policyDialog.error = ''
}

function cancelInboundEdit() {
  inboundDialog.value = false
  if (inboundForm.id) resetInbound()
}

function toggleAllInbounds(value: boolean) {
  selectedInboundIds.value = value ? distributedInbounds.value.map(inbound => inbound.id) : []
}

function resetNode() {
  Object.assign(nodeForm, newNode())
}

function resetDNSProvider() {
  Object.assign(dnsForm, newDNSProvider())
}

function resetCertificate() {
  Object.assign(certForm, newCertificate())
}

function resetInbound() {
  Object.assign(inboundForm, newInbound())
  originalInboundForm.value = null
}

function newNode() {
  return { id: 0, enable: true, name: '', code: '', region: '', provider: '', publicHost: '', publicIp: '' }
}

function newDNSProvider() {
  return { id: 0, enable: true, name: '', type: 'cloudflare', configText: providerCredentialTemplateText('cloudflare') }
}

function newCertificate() {
  return { id: 0, enable: true, name: '', source: 'acme-dns01', domainsText: '', wildcard: false, autoRenew: true, dnsProviderId: 0, ca: 'letsencrypt', acmeHome: '/root/.acme.sh', acmeDnsApi: '' }
}

function newCertificateVersion() {
  return { id: 0, certificateId: 0, fingerprint: '', fullchainPemEncrypted: '', privateKeyPemEncrypted: '' }
}

function newInbound() {
  return { id: 0, enable: true, nodeId: 0, protocol: 'anytls', publicHost: '', listen: '::', listenPort: 443, certificateId: 0, advancedText: '{}', policyOverridesJson: '{}' }
}

function defaultNodeTemplateForm() {
  return { config: stringifyJSON(defaultFullNodeTemplate()) }
}

function defaultFullNodeTemplate() {
  return {
    _comment: '这是所有集群节点默认继承的基础配置；inbounds 不在这里编辑，会在发布节点时由服务入口自动生成。',
    log: {
      level: 'info',
    },
    dns: {
      servers: [
        { type: 'tls', server: '2001:4860:4860::8888', tag: 'google-dns-v6' },
        { type: 'tls', server: '8.8.8.8', tag: 'google-dns-v4' },
      ],
      final: 'google-dns-v4',
    },
    outbounds: [
      {
        type: 'direct',
        tag: 'direct',
        domain_resolver: {
          server: 'google-dns-v4',
          strategy: 'prefer_ipv4',
        },
      },
    ],
    route: {
      default_domain_resolver: {
        server: 'google-dns-v4',
        strategy: 'prefer_ipv4',
      },
      rules: [
        {
          action: 'sniff',
          timeout: '1s',
        },
      ],
      final: 'direct',
    },
    experimental: {},
  }
}

function normalizedInbound(input: any) {
  return {
    id: Number(input.id || 0),
    enable: !!input.enable,
    nodeId: Number(input.nodeId || 0),
    protocol: String(input.protocol || '').trim(),
    publicHost: String(input.publicHost || '').trim(),
    listen: String(input.listen || '').trim(),
    listenPort: Number(input.listenPort || 0),
    certificateId: Number(input.certificateId || 0),
    advancedText: input.protocol === 'anytls' ? '' : compactJSONString(input.advancedText || '{}'),
  }
}

function inboundChanged() {
  return JSON.stringify(normalizedInbound(inboundForm)) !== JSON.stringify(originalInboundForm.value)
}

function newInboundUser() {
  return { id: 0, inboundId: 0, clientId: 0, name: '', password: randomPassword() }
}

function toForm(input: Record<string, any>) {
  const data = new URLSearchParams()
  Object.entries(input).forEach(([key, value]) => {
    if (value === undefined || value === null) return
    if (typeof value === 'object') data.append(key, JSON.stringify(value))
    else data.append(key, String(value))
  })
  return data
}

function parseJSON(value: string, fallback: any) {
  try {
    return JSON.parse(value)
  } catch {
    return fallback
  }
}

function stringifyJSON(value: any) {
  return JSON.stringify(value, null, 2)
}

function validateNodeTemplates(showError: boolean) {
  return !!parseNodeConfigTemplate(showError)
}

function validateAdvancedInboundJSON(showError: boolean) {
  try {
    const config = JSON.parse(inboundForm.advancedText || '{}')
    if (Array.isArray(config) || config === null || typeof config !== 'object') throw new Error('高级入站 JSON 必须是对象')
    if (config.type && config.type !== inboundForm.protocol) throw new Error(`高级入站 JSON 的 type 必须是 ${inboundForm.protocol}`)
    return true
  } catch (error: any) {
    if (showError) templateError.value = error?.message || String(error)
    return false
  }
}

function compactJSONString(value: string) {
  const parsed = JSON.parse(value || '{}')
  if (parsed === null || (typeof parsed === 'object' && !Array.isArray(parsed) && Object.keys(parsed).length === 0)) return '{}'
  return JSON.stringify(parsed)
}

function parsePolicyConfigText(showError: boolean) {
  const result = validatePolicyConfigText(policyDialog.configText)
  if (!result.valid) {
    if (showError) policyDialog.error = result.error
    return null
  }
  policyDialog.error = ''
  return result.value
}

function validatePolicyConfigText(value: string) {
  try {
    const config = JSON.parse(value || '{}')
    if (Array.isArray(config) || config === null || typeof config !== 'object') throw new Error('配置文件必须是 JSON 对象')
    if ('inbounds' in config) throw new Error('inbounds 由服务入口自动生成，这里不能手动配置')
    const dns = config.dns || {}
    const outbounds = config.outbounds || []
    const route = config.route || {}
    const experimental = config.experimental || {}
    if (Array.isArray(dns) || dns === null || typeof dns !== 'object') throw new Error('DNS JSON 必须是对象')
    if (!Array.isArray(outbounds)) throw new Error('Outbounds JSON 必须是数组')
    if (Array.isArray(route) || route === null || typeof route !== 'object') throw new Error('Route JSON 必须是对象')
    if (Array.isArray(experimental) || experimental === null || typeof experimental !== 'object') throw new Error('Experimental JSON 必须是对象')
    return { valid: true, value: { dns, outbounds, route, experimental }, error: '' }
  } catch (error: any) {
    return { valid: false, value: null, error: error?.message || String(error) }
  }
}

function validateDNSProviderConfigText(value: string, providerType = '') {
  try {
    const config = JSON.parse(value || '{}')
    if (Array.isArray(config) || config === null || typeof config !== 'object') throw new Error('API 凭据必须是 JSON 对象')
    validateDNSProviderCredentials(providerType, config)
    return { valid: true, value: config, error: '' }
  } catch (error: any) {
    return { valid: false, value: {}, error: error?.message || String(error) }
  }
}

function validateDNSProviderCredentials(providerType: string, config: Record<string, any>) {
  const requireOne = (keys: string[], label: string) => {
    const value = keys.map(key => config[key]).find(isRealCredentialValue)
    if (!value) throw new Error(`${label} 不能为空，请把模板提示替换成真实凭据`)
  }
  switch (providerType) {
    case 'cloudflare':
      requireOne(['CF_Token', 'apiToken', 'token'], 'CF_Token')
      break
    case 'aliyun':
      requireOne(['Ali_Key', 'aliKey', 'accessKeyId', 'accessKeyID'], 'Ali_Key')
      requireOne(['Ali_Secret', 'aliSecret', 'accessKeySecret'], 'Ali_Secret')
      break
    case 'route53':
      requireOne(['AWS_ACCESS_KEY_ID', 'awsAccessKeyId', 'accessKeyId'], 'AWS_ACCESS_KEY_ID')
      requireOne(['AWS_SECRET_ACCESS_KEY', 'awsSecretAccessKey', 'secretAccessKey'], 'AWS_SECRET_ACCESS_KEY')
      break
    case 'dnspod':
      requireOne(['DP_Id', 'dpId', 'id'], 'DP_Id')
      requireOne(['DP_Key', 'dpKey', 'key'], 'DP_Key')
      break
  }
}

function isRealCredentialValue(value: any) {
  if (typeof value !== 'string') return false
  const trimmed = value.trim()
  if (!trimmed) return false
  if (trimmed === DNS_CREDENTIAL_KEEP_ENCRYPTED) return true
  return !/^(在这里粘贴|替换为|填入|your |example|test$|test-|change-me)/i.test(trimmed)
}

function normalizeJSONText(value: string) {
  try {
    return JSON.stringify(JSON.parse(value || '{}'))
  } catch {
    return value.trim()
  }
}

function parseNodeConfigTemplate(showError: boolean) {
  try {
    const config = JSON.parse(nodeTemplateForm.config)
    if (Array.isArray(config) || config === null || typeof config !== 'object') throw new Error('config.json 必须是 JSON 对象')
    const log = config.log || { level: 'info' }
    const dns = config.dns
    const outbounds = config.outbounds
    const route = config.route
    if (Array.isArray(log) || log === null || typeof log !== 'object') throw new Error('log 必须是 JSON 对象')
    if (Array.isArray(dns) || dns === null || typeof dns !== 'object') throw new Error('DNS 必须是 JSON 对象')
    if (!Array.isArray(outbounds)) throw new Error('Outbounds 必须是 JSON 数组')
    if (Array.isArray(route) || route === null || typeof route !== 'object') throw new Error('Route 必须是 JSON 对象')
    templateError.value = ''
    const experimental = config.experimental || {}
    if (experimental !== undefined && (Array.isArray(experimental) || experimental === null || typeof experimental !== 'object')) throw new Error('experimental 必须是 JSON 对象')
    return { log, dns, outbounds, route, experimental }
  } catch (error: any) {
    if (showError) templateError.value = error?.message || String(error)
    return null
  }
}

function defaultAdvancedInbound(protocol: string) {
  const base = {
    type: protocol,
    tag: `${protocol}-example`,
    listen: '::',
    listen_port: ['hysteria', 'hysteria2'].includes(protocol) ? 8443 : 443,
  }
  switch (protocol) {
    case 'hysteria':
      return {
        ...base,
        tls: {
          enabled: true,
          server_name: 'hysteria.example.com',
          certificate_path: '/usr/local/s-ui-agent/certs/hysteria.example.com/fullchain.pem',
          key_path: '/usr/local/s-ui-agent/certs/hysteria.example.com/privkey.pem',
        },
        up_mbps: 100,
        down_mbps: 100,
        users: [{ name: 'alice', auth_str: 'change-me' }],
      }
    case 'hysteria2':
      return {
        ...base,
        tls: {
          enabled: true,
          server_name: 'hy2.example.com',
          certificate_path: '/usr/local/s-ui-agent/certs/hy2.example.com/fullchain.pem',
          key_path: '/usr/local/s-ui-agent/certs/hy2.example.com/privkey.pem',
        },
        users: [{ name: 'alice', password: 'change-me' }],
        masquerade: 'https://example.com',
      }
    case 'tuic':
      return {
        ...base,
        tls: {
          enabled: true,
          server_name: 'tuic.example.com',
          certificate_path: '/usr/local/s-ui-agent/certs/tuic.example.com/fullchain.pem',
          key_path: '/usr/local/s-ui-agent/certs/tuic.example.com/privkey.pem',
        },
        users: [{ name: 'alice', uuid: '00000000-0000-0000-0000-000000000000', password: 'change-me' }],
      }
    case 'trojan':
      return {
        ...base,
        tls: {
          enabled: true,
          server_name: 'trojan.example.com',
          certificate_path: '/usr/local/s-ui-agent/certs/trojan.example.com/fullchain.pem',
          key_path: '/usr/local/s-ui-agent/certs/trojan.example.com/privkey.pem',
        },
        users: [{ name: 'alice', password: 'change-me' }],
      }
    case 'vless':
      return {
        ...base,
        tls: {
          enabled: true,
          server_name: 'vless.example.com',
          certificate_path: '/usr/local/s-ui-agent/certs/vless.example.com/fullchain.pem',
          key_path: '/usr/local/s-ui-agent/certs/vless.example.com/privkey.pem',
        },
        users: [{ name: 'alice', uuid: '00000000-0000-0000-0000-000000000000' }],
      }
    case 'vmess':
      return {
        ...base,
        tls: {
          enabled: true,
          server_name: 'vmess.example.com',
          certificate_path: '/usr/local/s-ui-agent/certs/vmess.example.com/fullchain.pem',
          key_path: '/usr/local/s-ui-agent/certs/vmess.example.com/privkey.pem',
        },
        users: [{ name: 'alice', uuid: '00000000-0000-0000-0000-000000000000', alterId: 0 }],
        transport: { type: 'tcp' },
      }
    case 'naive':
      return {
        ...base,
        tls: {
          enabled: true,
          server_name: 'naive.example.com',
          certificate_path: '/usr/local/s-ui-agent/certs/naive.example.com/fullchain.pem',
          key_path: '/usr/local/s-ui-agent/certs/naive.example.com/privkey.pem',
        },
        users: [{ username: 'alice', password: 'change-me' }],
      }
    case 'shadowsocks':
      return {
        ...base,
        listen_port: 8388,
        method: '2022-blake3-aes-128-gcm',
        password: 'change-me',
      }
    case 'shadowtls':
      return {
        ...base,
        listen_port: 8443,
        version: 3,
        users: [{ name: 'alice', password: 'change-me' }],
        handshake: {
          server: 'www.microsoft.com',
          server_port: 443,
        },
        strict_mode: true,
      }
    default:
      return base
  }
}

function defaultVLESSRealityInbound() {
  return {
    type: 'vless',
    tag: 'vless-reality-example',
    listen: '::',
    listen_port: 443,
    users: [
      {
        name: 'alice',
        uuid: '00000000-0000-0000-0000-000000000000',
      },
    ],
    tls: {
      enabled: true,
      server_name: 'www.microsoft.com',
      reality: {
        enabled: true,
        handshake: {
          server: 'www.microsoft.com',
          server_port: 443,
        },
        private_key: 'server-private-key',
        short_id: [''],
      },
    },
  }
}

function composeNodeConfigTemplate(input: any) {
  const config: Record<string, any> = {
    _comment: '这是所有集群节点默认继承的基础配置；inbounds 不在这里编辑，会在发布节点时由服务入口自动生成。',
    log: parseJSONValue(input?.log, { level: 'info' }),
    dns: parseJSONValue(input?.dns, defaultFullNodeTemplate().dns),
    outbounds: parseJSONValue(input?.outbounds, defaultFullNodeTemplate().outbounds),
    route: parseJSONValue(input?.route, defaultFullNodeTemplate().route),
  }
  config.experimental = parseJSONValue(input?.experimental, {})
  return config
}

function syncTemplateEditorScroll(event: Event) {
  const target = event.target as HTMLElement
  templateEditorScrollTop.value = target.scrollTop
  templateEditorScrollLeft.value = target.scrollLeft
}

function syncPolicyEditorScroll(event: Event) {
  const target = event.target as HTMLTextAreaElement
  policyDialog.scrollTop = target.scrollTop
  policyDialog.scrollLeft = target.scrollLeft
}

function highlightJSON(value: string) {
  const escaped = escapeHTML(value)
  return escaped.replace(/("(?:\\.|[^"\\])*")(\s*:)?|(-?\d+(?:\.\d+)?(?:[eE][+-]?\d+)?)\b|\b(true|false|null)\b/g, (match, str, colon, num, keyword) => {
    if (str) {
      return `<span class="${colon ? 'json-key' : 'json-string'}">${str}</span>${colon || ''}`
    }
    if (num) return `<span class="json-number">${num}</span>`
    if (keyword) return `<span class="json-keyword">${keyword}</span>`
    return match
  })
}

function escapeHTML(value: string) {
  return value
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
}

function parseJSONValue(value: any, fallback: any) {
  if (!value) return fallback
  if (typeof value === 'object') return value
  return parseJSON(value, fallback)
}

function hasPolicyOverrides(inbound: any) {
  const policy = parseJSONValue(inbound.policyOverridesJson, {})
  return !!policy && typeof policy === 'object' && !Array.isArray(policy) && Object.keys(policy).length > 0
}

function certificateConfig() {
  if (certForm.source === 'acme-dns01') {
    return {
      ca: certForm.ca,
      challenge: 'dns-01',
      dnsProviderId: certForm.dnsProviderId,
    }
  }
  if (certForm.source === 'acme.sh') {
    return {
      ca: certForm.ca,
      acmeHome: certForm.acmeHome,
      acmeDnsApi: certForm.acmeDnsApi,
    }
  }
  return {}
}

function nodeName(id: number) {
  return nodes.value.find(n => n.id === id)?.name || '-'
}

function nodeHost(id: number) {
  return nodes.value.find(n => n.id === id)?.publicHost || ''
}

function nodeInboundCount(id: number) {
  return distributedInbounds.value.filter(inbound => Number(inbound.nodeId) === Number(id)).length
}

function goSetupStep(step: any) {
  if (step.path) {
    router.push(step.path)
    return
  }
  tab.value = step.tab
  if (step.action === 'node' && !step.done) openNodeDrawer()
}

function initialTab() {
  if (route.path === '/subscriptions') return 'subscriptions'
  return String(route.query.tab || 'nodes')
}

function dnsProviderName(id: number) {
  const provider = dnsProviders.value.find(p => p.id === id)
  return provider ? `${provider.name} (${providerTypeName(provider.type)})` : '未选择'
}

function nodeHealth(node: any) {
  if (!node?.enable) {
    return { state: 'disabled', label: '已禁用', color: 'grey', detail: '节点已禁用', agentOnline: false }
  }

  const agentStatus = String(node.agentStatus || '').toLowerCase()
  const singboxStatus = String(node.singboxStatus || '').toLowerCase()
  const lastSeenAt = Number(node.lastSeenAt || 0)
  const age = lastSeenAt > 0 ? currentUnix.value - lastSeenAt : Number.POSITIVE_INFINITY
  const agentOnline = Boolean(agentStatus) && agentStatus !== 'offline' && lastSeenAt > 0 && age <= 120

  if (!agentStatus || lastSeenAt <= 0) {
    return { state: 'unknown', label: '未接入', color: 'grey', detail: 'Agent 尚未注册或没有心跳记录', agentOnline: false }
  }
  if (agentStatus === 'offline' || age > 300) {
    return { state: 'offline', label: '离线', color: 'error', detail: `最后心跳 ${lastSeenText(lastSeenAt)}`, agentOnline: false }
  }
  if (age > 120) {
    return { state: 'timeout', label: '心跳超时', color: 'warning', detail: `最后心跳 ${lastSeenText(lastSeenAt)}`, agentOnline: false }
  }
  if (singboxStatus && !['running', 'online', 'ok', 'healthy'].includes(singboxStatus)) {
    return { state: 'singbox_error', label: 'Sing-box 异常', color: 'error', detail: `Sing-box 状态：${node.singboxStatus}`, agentOnline }
  }
  if (!latestVersion(node.id)) {
    return { state: 'no_config', label: '未发布配置', color: 'warning', detail: '节点在线，但还没有发布配置版本', agentOnline }
  }
  return { state: 'healthy', label: '正常', color: 'success', detail: 'Agent 在线，Sing-box 状态正常，已有配置版本', agentOnline }
}

function inboundHealth(inbound: any) {
  if (!inbound?.enable) {
    return { label: '已禁用', color: 'grey', detail: '服务入口已禁用' }
  }
  const node = nodes.value.find(item => Number(item.id) === Number(inbound.nodeId))
  if (!node) {
    return { label: '节点缺失', color: 'error', detail: '找不到该服务入口所属节点' }
  }
  const health = nodeHealth(node)
  if (!['healthy', 'no_config'].includes(health.state)) {
    return { label: health.label, color: health.color, detail: `节点状态：${health.detail}` }
  }
  if (!inbound.renderedConfigJson) {
    return { label: '未生成配置', color: 'warning', detail: '请先生成该服务入口配置' }
  }
  if (protocolNeedsCertificateByName(inbound.protocol)) {
    const certificate = certificates.value.find(item => Number(item.id) === Number(inbound.certificateId))
    if (!certificate) {
      return { label: '未选证书', color: 'warning', detail: '该协议需要证书，但当前没有绑定证书' }
    }
    if (!certificate.activeVersionId) {
      return { label: '证书未签发', color: 'warning', detail: '证书还没有可下发的签发版本' }
    }
  }
  if (!latestVersion(inbound.nodeId)) {
    return { label: '待发布', color: 'warning', detail: '服务入口已生成，但节点配置还没有发布' }
  }
  return { label: '基础正常', color: 'success', detail: '节点在线、配置已生成、证书可用。协议握手探测会在下一阶段加入。' }
}

function protocolNeedsCertificateByName(protocol: string) {
  return ['anytls', 'hysteria', 'hysteria2', 'tuic', 'trojan', 'vless', 'vmess', 'naive'].includes(String(protocol || '').toLowerCase())
}

function agentStatusLabel(node: any) {
  const status = String(node.agentStatus || '').trim()
  if (!status) return '未接入'
  if (nodeHealth(node).state === 'timeout') return '心跳超时'
  if (nodeHealth(node).state === 'offline') return '离线'
  if (status === 'registered') return '已注册'
  if (status === 'online') return '在线'
  return status
}

function agentStatusMeta(node: any) {
  const health = nodeHealth(node)
  if (!node?.enable) {
    return { label: 'Agent 已禁用', tone: 'muted' }
  }
  if (health.state === 'timeout') {
    return { label: 'Agent 心跳超时', tone: 'warning' }
  }
  if (health.state === 'offline') {
    return { label: 'Agent 离线', tone: 'danger' }
  }
  if (!node?.lastSeenAt || !String(node.agentStatus || '').trim()) {
    return { label: 'Agent 未接入', tone: 'warning' }
  }
  return { label: `Agent ${agentStatusLabel(node)}`, tone: 'success' }
}

function singboxStatusLabel(node: any) {
  const status = String(node.singboxStatus || '').trim()
  if (!status) return '未上报'
  if (status === 'running') return '运行中'
  if (status === 'stopped') return '已停止'
  if (status === 'failed') return '异常'
  return status
}

function singboxStatusMeta(node: any) {
  const status = String(node.singboxStatus || '').trim().toLowerCase()
  const version = String(node.singboxVersion || '').trim()
  if (!node?.enable) {
    return { label: 'sing-box 已禁用', tone: 'muted' }
  }
  if (!status || status === 'unknown') {
    return { label: 'sing-box 未上报', tone: 'muted' }
  }
  if (['running', 'online', 'ok', 'healthy'].includes(status)) {
    return { label: `sing-box 运行中${version && version !== 'unknown' ? ` · ${version}` : ''}`, tone: 'success' }
  }
  if (status === 'stopped') {
    return { label: 'sing-box 已停止', tone: 'warning' }
  }
  if (['failed', 'error', 'crashed'].includes(status)) {
    return { label: 'sing-box 异常', tone: 'danger' }
  }
  return { label: `sing-box ${singboxStatusLabel(node)}`, tone: 'muted' }
}

function nodeConfigLabel(node: any) {
  const draft = String(node.draftSha256 || '').trim()
  const published = String(node.publishedSha256 || '').trim()
  const applied = String(node.appliedSha256 || '').trim()

  if (!published || draft !== published) {
    return '配置待下发'
  }
  if (published !== applied) {
    return '正在同步中...'
  }
  return '同步成功'
}

function nodeConfigClass(node: any) {
  const draft = String(node.draftSha256 || '').trim()
  const published = String(node.publishedSha256 || '').trim()
  const applied = String(node.appliedSha256 || '').trim()

  if (!published || draft !== published) {
    return 'config-version-pill--pending'
  }
  if (published !== applied) {
    return 'config-version-pill--syncing'
  }
  return 'config-version-pill--success'
}

function lastSeenText(value: any) {
  const lastSeenAt = Number(value || 0)
  if (!lastSeenAt) return '从未上报'
  const diff = Math.max(0, currentUnix.value - lastSeenAt)
  if (diff < 60) return `${diff} 秒前`
  if (diff < 3600) return `${Math.floor(diff / 60)} 分钟前`
  if (diff < 86400) return `${Math.floor(diff / 3600)} 小时前`
  return `${Math.floor(diff / 86400)} 天前`
}

function providerTypeName(type: string) {
  return dnsTypeOptions.find(item => item.value === type)?.title || type || '-'
}

function providerIcon(type: string) {
  const icons: Record<string, string> = {
    cloudflare: cloudflareIcon,
    route53: awsRoute53Icon,
    aliyun: alidnsIcon,
    dnspod: tencentDnspodIcon,
  }
  return icons[type] || ''
}

function selectDNSProviderType(type: string) {
  const oldType = dnsForm.type
  const shouldReplace = shouldReplaceDNSCredentialTemplate(dnsForm.configText, oldType)
  dnsForm.type = type
  if (shouldReplace) dnsForm.configText = providerCredentialTemplateText(type)
  dnsProviderTypeMenu.value = false
}

function providerCredentialTemplateText(type: string) {
  const template = dnsCredentialGuides[type]?.template
  return template ? stringifyJSON(template) : '{}'
}

function shouldReplaceDNSCredentialTemplate(text: string, oldType?: string) {
  const normalized = normalizeJSONText(text || '{}')
  if (normalized === '{}') return true
  if (oldType && normalized === normalizeJSONText(providerCredentialTemplateText(oldType))) return true
  return Object.keys(dnsCredentialGuides).some(type => {
    const template = dnsCredentialGuides[type]?.template
    return template && normalized === normalizeJSONText(stringifyJSON(template))
  })
}

function applyDNSCredentialTemplate() {
  dnsForm.configText = providerCredentialTemplateText(dnsForm.type)
}

function clientName(id: number) {
  return clients.value.find(c => c.id === id)?.name || '-'
}

function allowsClusterAccess(client: any) {
  return true
}

function distributedSubUrl(token: string) {
  if (settings.value.subURI) {
    let uri = settings.value.subURI.trim()
    if (!/^https?:\/\//i.test(uri)) {
      uri = 'http://' + uri
    }
    if (!uri.endsWith('/')) {
      uri += '/'
    }
    return `${uri}sub/${token}`
  }

  const protocol = settings.value.subCertFile ? 'https:' : window.location.protocol
  const host = settings.value.subDomain || window.location.hostname
  const port = settings.value.subPort ? `:${settings.value.subPort}` : ':2096'
  const path = settings.value.subPath || '/sub/'
  return `${protocol}//${host}${port}${path}${token}`
}

function latestVersion(nodeId: number) {
  return configVersions.value.find(v => v.nodeId === nodeId)
}

function formatDomains(value: any) {
  if (!value) return ''
  if (Array.isArray(value)) return value.join(', ')
  if (typeof value === 'string') {
    try {
      const parsed = JSON.parse(value)
      return Array.isArray(parsed) ? parsed.join(', ') : value
    } catch {
      return value
    }
  }
  return ''
}

function formatTimestamp(value: number) {
  if (!value) return '-'
  const date = new Date(value * 1000)
  const y = date.getFullYear()
  const m = String(date.getMonth() + 1).padStart(2, '0')
  const d = String(date.getDate()).padStart(2, '0')
  return `${y}-${m}-${d}`
}

function certSourceTitle(source: string) {
  if (source === 'manual') return '手动上传（旧）'
  return certSourceOptions.find(item => item.value === source)?.title || source || '-'
}

function certificateSourceIcon(source: string) {
  switch (source) {
    case 'acme-dns01':
      return 'mdi-cloud-key-outline'
    case 'acme.sh':
      return 'mdi-console'
    case 'manual':
      return 'mdi-file-upload-outline'
    default:
      return 'mdi-certificate-outline'
  }
}

function randomPassword() {
  const data = new Uint8Array(16)
  crypto.getRandomValues(data)
  return Array.from(data).map(v => v.toString(16).padStart(2, '0')).join('')
}
</script>

<style scoped>
.certificate-stats {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 12px;
  margin-bottom: 16px;
}

.certificate-stat {
  display: grid;
  grid-template-columns: 42px 1fr;
  align-items: center;
  gap: 12px;
  min-height: 92px;
  padding: 16px;
  border: 1px solid var(--panel-card-border);
  border-radius: 8px;
  background: var(--panel-soft-bg);
}

.certificate-stat :deep(.v-icon) {
  padding: 10px;
  border-radius: 8px;
  background: var(--panel-strong-bg);
}

.certificate-stat-label {
  color: var(--panel-muted-text);
  font-size: 12px;
}

.certificate-stat-value {
  margin-top: 4px;
  color: var(--panel-strong-text);
  font-size: 22px;
  font-weight: 800;
  line-height: 1.15;
}

.certificate-stat-success {
  color: #16a34a;
  font-size: 18px;
}

.certificate-stat-hint {
  margin-top: 5px;
  color: var(--panel-muted-text);
  font-size: 12px;
}

.dns-provider-select-wrap {
  position: relative;
}

.dns-provider-select-label {
  position: absolute;
  top: -8px;
  left: 12px;
  z-index: 1;
  padding: 0 4px;
  background: var(--panel-modal-bg);
  color: var(--panel-muted-text);
  font-size: 12px;
  line-height: 1;
}

.dns-provider-select {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  height: 56px;
  padding: 0 14px 0 16px;
  border: 1px solid var(--panel-input-border);
  border-radius: 8px;
  background: var(--panel-input-bg);
  color: var(--panel-strong-text);
  cursor: pointer;
  transition: border-color 0.16s ease, box-shadow 0.16s ease;
}

.dns-provider-select:hover,
.dns-provider-select[aria-expanded="true"] {
  border-color: #2563eb;
  box-shadow: 0 0 0 1px #2563eb;
}

.dns-provider-option {
  display: flex;
  align-items: center;
  gap: 12px;
  height: 56px;
  padding: 0 16px;
  min-width: 0;
}

.dns-provider-option-inline,
.dns-provider-option-selection {
  height: auto;
  padding: 0;
}

.dns-provider-menu {
  overflow: hidden;
  min-width: 360px;
  border: 1px solid #d8e0ea;
  border-radius: 8px;
  background: #fff;
  box-shadow: 0 14px 28px rgba(15, 23, 42, 0.18);
}

.dns-provider-menu-item {
  width: 100%;
  border: 0;
  background: #fff;
  text-align: left;
  cursor: pointer;
}

.dns-provider-menu-item:hover,
.dns-provider-menu-item.active {
  background: #f3f4f6;
}

.dns-provider-icon {
  display: block;
  width: 28px;
  height: 28px;
  flex: 0 0 28px;
  object-fit: contain;
}

.dns-provider-icon-fallback {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  color: #64748b;
}

.dns-provider-name {
  color: #1f2937;
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
  font-size: 16px;
  line-height: 1;
  min-width: 0;
}

.dns-credential-guide {
  padding: 12px;
  border: 1px solid var(--panel-card-border);
  border-radius: 8px;
  background: var(--panel-soft-bg);
}

.dns-credential-title {
  color: var(--panel-strong-text);
  font-size: 14px;
  font-weight: 800;
}

.dns-credential-desc {
  margin-top: 4px;
  color: var(--panel-muted-text);
  font-size: 12px;
  line-height: 1.6;
}

.dns-credential-fields {
  display: grid;
  gap: 8px;
  margin-top: 12px;
}

.dns-credential-field {
  display: grid;
  grid-template-columns: minmax(150px, 190px) 1fr;
  gap: 10px;
  align-items: start;
  color: var(--panel-muted-text);
  font-size: 12px;
}

.dns-credential-field code {
  width: fit-content;
  padding: 3px 6px;
  border-radius: 6px;
  background: rgba(37, 99, 235, 0.08);
  color: #2563eb;
  font-size: 12px;
}

.certificate-source-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 10px;
}

.certificate-source-option {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  min-height: 48px;
  border: 1px solid var(--panel-card-border);
  border-radius: 8px;
  background: var(--panel-soft-bg);
  color: var(--panel-strong-text);
  font-weight: 700;
  transition: border-color 0.16s ease, background 0.16s ease, color 0.16s ease;
}

.certificate-source-option:hover,
.certificate-source-option.active {
  border-color: rgba(59, 130, 246, 0.72);
  background: rgba(59, 130, 246, 0.12);
  color: #2563eb;
}

.certificate-switch-row {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
}

.cluster-stats {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 12px;
  margin-bottom: 16px;
}

.cluster-stat {
  display: grid;
  grid-template-columns: 34px 1fr;
  align-items: center;
  gap: 12px;
  min-height: 82px;
  padding: 16px;
  border: 1px solid var(--panel-card-border);
  border-radius: 8px;
  background: var(--panel-soft-bg);
}

.cluster-stat-value {
  color: var(--panel-strong-text);
  font-size: 24px;
  font-weight: 700;
  line-height: 1.1;
}

.cluster-stat-label {
  margin-top: 4px;
  color: var(--panel-muted-text);
  font-size: 12px;
}

.guide-prompt-card {
  border: 1px solid rgba(59, 130, 246, 0.28);
  border-radius: 8px;
  background: linear-gradient(135deg, rgba(59, 130, 246, 0.12), var(--panel-strong-bg));
}

.guide-progress {
  height: 8px;
  overflow: hidden;
  border-radius: 999px;
  background: rgba(var(--v-border-color), 0.16);
}

.guide-progress-bar {
  height: 100%;
  border-radius: inherit;
  background: linear-gradient(90deg, rgb(var(--v-theme-primary)), rgba(0, 210, 255, 0.82));
  transition: width 0.2s ease;
}

.setup-list {
  display: grid;
  gap: 10px;
}

/* 时间轴 stepper 样式 */
.guide-step-item {
  position: relative;
  display: grid;
  grid-template-columns: 32px 1fr auto;
  align-items: center;
  gap: 12px;
  width: 100%;
  padding: 16px;
  border: 1px solid var(--panel-card-border);
  border-radius: 12px;
  background: var(--panel-strong-bg);
  text-align: left;
  transition: all 0.24s cubic-bezier(0.4, 0, 0.2, 1);
}

.guide-step-item:hover {
  border-color: rgba(0, 210, 255, 0.48);
  background: rgba(0, 210, 255, 0.04);
  transform: translateY(-1px);
}

/* 激活的当前步骤高亮效果 */
.guide-step-item.is-active {
  border-color: rgb(var(--v-theme-primary));
  background: linear-gradient(135deg, rgba(var(--v-theme-primary), 0.15), var(--panel-strong-bg));
  box-shadow: 0 0 16px rgba(var(--v-theme-primary), 0.18);
}

.guide-step-item.is-active .text-grey-lighten-2 {
  color: rgb(var(--v-theme-primary)) !important;
  font-weight: 600 !important;
}

/* 时间轴线条和徽章 */
.step-connector-wrapper {
  display: flex;
  flex-direction: column;
  align-items: center;
  height: 100%;
  position: relative;
}

.step-badge {
  width: 24px;
  height: 24px;
  border-radius: 50%;
  background: rgba(var(--v-border-color), 0.12);
  border: 1px solid rgba(var(--v-border-color), 0.24);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 11px;
  font-weight: 700;
  color: var(--panel-muted-text);
  z-index: 2;
  transition: all 0.2s ease;
}

.is-done .step-badge {
  background: rgba(var(--v-theme-success), 0.12);
  border-color: rgb(var(--v-theme-success));
  color: rgb(var(--v-theme-success));
}

.is-active .step-badge {
  background: rgb(var(--v-theme-primary));
  border-color: rgb(var(--v-theme-primary));
  color: white;
}

.step-connector-line {
  position: absolute;
  top: 24px;
  bottom: -28px;
  width: 2px;
  background: rgba(var(--v-border-color), 0.08);
  z-index: 1;
}

.guide-step-item:last-child .step-connector-line {
  display: none;
}

/* 呼吸小红点 */
.pulse-dot {
  width: 6px;
  height: 6px;
  background-color: rgb(var(--v-theme-primary));
  border-radius: 50%;
  display: inline-block;
  box-shadow: 0 0 0 0 rgba(var(--v-theme-primary), 0.7);
  animation: pulse-animation 1.6s infinite cubic-bezier(0.66, 0, 0, 1);
}

@keyframes pulse-animation {
  0% {
    box-shadow: 0 0 0 0 rgba(var(--v-theme-primary), 0.7);
  }
  70% {
    box-shadow: 0 0 0 6px rgba(var(--v-theme-primary), 0);
  }
  100% {
    box-shadow: 0 0 0 0 rgba(var(--v-theme-primary), 0);
  }
}

.animate-pulse-btn {
  animation: pulse-button 2s infinite ease-in-out;
}

@keyframes pulse-button {
  0%, 100% {
    opacity: 0.9;
    transform: scale(1);
  }
  50% {
    opacity: 1;
    transform: scale(1.03);
  }
}


.cluster-node-table {
  display: grid;
  gap: 8px;
}

.cluster-node-table-head,
.cluster-node-row {
  display: grid;
  grid-template-columns: minmax(180px, 1fr) minmax(280px, 1.5fr) minmax(120px, 0.7fr) minmax(100px, 0.6fr) minmax(240px, 1.2fr);
  align-items: center;
  gap: 14px;
}

.cluster-node-table-head {
  padding: 0 12px 8px;
  border-bottom: 1px solid var(--panel-card-border);
  color: var(--panel-muted-text);
  font-size: 13px;
  font-weight: 700;
}

.cluster-node-row {
  min-height: 74px;
  padding: 12px;
  border: 1px solid var(--panel-card-border);
  border-radius: 8px;
  background: var(--panel-strong-bg);
}

.cluster-node-cell {
  min-width: 0;
}

.cluster-node-identity,
.cluster-node-actions,
.node-status-inline {
  display: flex;
  align-items: center;
}

.cluster-node-identity {
  gap: 10px;
}

.cluster-node-name-wrap {
  min-width: 0;
}

.cluster-node-name {
  overflow: hidden;
  color: var(--panel-title-text);
  font-size: 15px;
  font-weight: 800;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.cluster-node-meta {
  overflow: hidden;
  margin-top: 3px;
  color: var(--panel-muted-text);
  font-size: 12px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.cluster-node-host {
  overflow: hidden;
  margin-top: 5px;
  color: var(--panel-strong-text);
  font-size: 13px;
  font-weight: 700;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.node-status-inline {
  flex-wrap: wrap;
  justify-content: center;
  gap: 8px;
}

.status-pill,
.config-version-pill {
  display: inline-flex;
  align-items: center;
  max-width: 100%;
  min-height: 30px;
  border-radius: 999px;
  font-size: 13px;
  font-weight: 700;
  line-height: 1;
  white-space: nowrap;
}

.status-pill {
  gap: 7px;
  padding: 0 11px;
}

.status-dot {
  width: 9px;
  height: 9px;
  flex: 0 0 9px;
  border-radius: 999px;
}

.status-pill--success {
  background: rgba(34, 197, 94, 0.12);
  color: #16a34a;
}

.status-pill--warning {
  background: rgba(245, 158, 11, 0.14);
  color: #d97706;
}

.status-pill--danger {
  background: rgba(239, 68, 68, 0.13);
  color: #dc2626;
}

.status-pill--muted {
  background: rgba(100, 116, 139, 0.13);
  color: var(--panel-muted-text);
}

.status-dot--success {
  background: #16a34a;
  animation: status-dot-pulse 1.6s ease-out infinite;
}

.status-dot--warning {
  background: #d97706;
}

.status-dot--danger {
  background: #dc2626;
  animation: status-dot-pulse 1.6s ease-out infinite;
}

.status-dot--muted {
  background: #94a3b8;
}

.config-version-pill {
  padding: 0 10px;
  background: rgba(37, 99, 235, 0.10);
  color: #2563eb;
}

.config-version-pill--pending {
  background: rgba(100, 116, 139, 0.12);
  color: var(--panel-muted-text);
}

.config-version-pill--success {
  background: rgba(22, 163, 74, 0.10);
  color: #16a34a;
}

.cluster-node-config-wrap {
  display: inline-flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  width: 100%;
}

.cluster-node-submeta {
  color: var(--panel-muted-text);
  font-size: 12px;
  font-weight: 600;
}

.cluster-node-heartbeat {
  color: var(--panel-strong-text);
  font-size: 13px;
  font-weight: 700;
  text-align: center;
}

.cluster-node-actions {
  justify-content: flex-end;
  gap: 6px;
}

@keyframes status-dot-pulse {
  0% {
    opacity: 1;
  }
  50% {
    opacity: 0.42;
  }
  100% {
    opacity: 1;
  }
}

@media (prefers-reduced-motion: reduce) {
  .status-dot--success,
  .status-dot--danger {
    animation: none;
  }
}

@media (max-width: 1100px) {
  .cluster-node-table-head {
    display: none;
  }

  .cluster-node-row {
    grid-template-columns: 1fr;
    align-items: start;
  }

  .cluster-node-actions {
    justify-content: flex-start;
  }
}

.node-editor-drawer {
  background: var(--panel-card-bg);
  border-left: 1px solid var(--panel-card-border);
}

.inbound-action-group {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  min-width: max-content;
  padding: 3px;
  border: 1px solid var(--panel-card-border);
  border-radius: 8px;
  background: var(--panel-soft-bg);
}

.inbound-action-btn {
  min-width: 58px;
  height: 30px;
  border-radius: 6px;
  font-size: 12px;
  transition: background 0.18s ease, color 0.18s ease;
}

.inbound-action-btn:hover {
  background: var(--panel-soft-bg-hover);
}

.inbound-toolbar-actions {
  gap: 10px;
}

.cluster-create-btn {
  min-height: 38px;
  border-radius: 6px;
  background: linear-gradient(135deg, #11a66a, #16c784);
  color: #ffffff;
  font-weight: 700;
  box-shadow: 0 8px 18px rgba(17, 166, 106, 0.22);
}

.cluster-delete-btn {
  min-height: 38px;
  border: 1px solid rgba(239, 68, 68, 0.55);
  border-radius: 6px;
  background: rgba(239, 68, 68, 0.12);
  color: #ef4444;
  font-weight: 700;
}

.cluster-delete-btn:disabled {
  border-color: var(--panel-card-border);
  background: var(--panel-soft-bg);
  color: var(--panel-muted-text);
  opacity: 0.65;
}

.advanced-json-panel {
  border: 1px solid var(--panel-card-border);
  border-radius: 8px;
  background: var(--panel-soft-bg);
}

.template-editor-shell {
  overflow: hidden;
  border: 1px solid var(--panel-card-border);
  border-radius: 8px;
  background: var(--panel-strong-bg);
}

.template-editor-toolbar {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 12px;
  border-bottom: 1px solid var(--panel-card-border);
  color: var(--panel-muted-text);
  font-size: 12px;
}

.template-editor {
  display: grid;
  grid-template-columns: 52px 1fr;
  height: 520px;
  background: var(--panel-card-bg);
}

.template-line-numbers {
  padding: 14px 10px 14px 0;
  border-right: 1px solid var(--panel-card-border);
  color: var(--panel-muted-text);
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace;
  font-size: 13px;
  line-height: 1.55;
  text-align: right;
  user-select: none;
}

.template-line-numbers span {
  display: block;
  height: 20.15px;
}

.template-code-wrap {
  position: relative;
  overflow: hidden;
}

.template-highlight,
.template-code-input {
  box-sizing: border-box;
  width: 100%;
  min-height: 100%;
  margin: 0;
  padding: 14px 16px;
  border: 0;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace;
  font-size: 13px;
  line-height: 1.55;
  tab-size: 2;
  white-space: pre;
}

.template-highlight {
  position: absolute;
  top: 0;
  left: 0;
  min-width: 100%;
  min-height: max-content;
  overflow: visible;
  color: var(--panel-strong-text);
  pointer-events: none;
}

.template-code-input {
  position: relative;
  z-index: 1;
  height: 520px;
  overflow: auto;
  resize: none;
  outline: none;
  background: transparent;
  caret-color: var(--panel-strong-text);
  color: transparent;
}

.template-code-view {
  height: 520px;
  overflow: auto;
  overscroll-behavior: contain;
}

.template-highlight-view {
  position: static;
  display: block;
  width: max-content;
  min-width: 100%;
}

.template-code-input::selection {
  background: rgba(59, 130, 246, 0.28);
}

.policy-editor {
  height: min(54vh, 430px);
  min-height: 320px;
}

.policy-editor .template-line-numbers,
.policy-editor .template-code-wrap {
  min-height: 0;
  height: 100%;
}

.policy-editor .template-highlight {
  min-height: auto;
}

.policy-code-input {
  height: 100%;
  min-height: 0;
  overflow: auto;
  overscroll-behavior: contain;
}

.json-editor :deep(textarea) {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace;
  font-size: 13px;
  line-height: 1.55;
  tab-size: 2;
}

.rendered-config-preview,
.protocol-template-preview {
  overflow: auto;
  max-height: 360px;
  margin: 0;
  padding: 12px;
  border: 1px solid var(--panel-card-border);
  border-radius: 8px;
  background: var(--panel-strong-bg);
  color: var(--panel-strong-text);
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace;
  font-size: 12px;
  line-height: 1.55;
  white-space: pre;
}

.rendered-config-preview {
  max-height: 620px;
}

.template-highlight :deep(.json-key),
.rendered-config-preview :deep(.json-key),
.protocol-template-preview :deep(.json-key) {
  color: #60a5fa;
}

.template-highlight :deep(.json-string),
.rendered-config-preview :deep(.json-string),
.protocol-template-preview :deep(.json-string) {
  color: #22c55e;
}

.template-highlight :deep(.json-number),
.rendered-config-preview :deep(.json-number),
.protocol-template-preview :deep(.json-number) {
  color: #f59e0b;
}

.template-highlight :deep(.json-keyword),
.rendered-config-preview :deep(.json-keyword),
.protocol-template-preview :deep(.json-keyword) {
  color: #c084fc;
}

@media (max-width: 960px) {
  .cluster-stats,
  .certificate-stats,
  .node-meta-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .certificate-source-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 600px) {
  .cluster-stats,
  .certificate-stats,
  .node-meta-grid {
    grid-template-columns: 1fr;
  }

  .node-card-main {
    align-items: flex-start;
    flex-direction: column;
  }
}

/* 微交互操作按钮动效 */
.tech-action-btn {
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1) !important;
}

.tech-action-btn:hover {
  transform: translateY(-1px) scale(1.02);
  box-shadow: 0 4px 10px rgba(37, 99, 235, 0.15);
}

.tech-action-btn:active {
  transform: translateY(0px) scale(0.96);
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.2);
}

/* 同步中流光动画 */
.config-version-pill--syncing {
  background: rgba(37, 99, 235, 0.12) !important;
  color: #2563eb !important;
  position: relative;
  overflow: hidden;
}

.config-version-pill--syncing::after {
  content: "";
  position: absolute;
  top: 0;
  right: 0;
  bottom: 0;
  left: 0;
  background: linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.25), transparent);
  transform: translateX(-100%);
  animation: shimmer 1.5s infinite;
}

/* 慢速旋转 */
.rotate-anim {
  animation: rotate 2s linear infinite;
}

@keyframes rotate {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

@keyframes shimmer {
  100% {
    transform: translateX(100%);
  }
}
</style>
