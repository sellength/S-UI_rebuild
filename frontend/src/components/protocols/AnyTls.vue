<template>
  <v-card subtitle="AnyTLS">
    <v-row class="ma-0">
      <!-- 密码仅在具有 password 字段时（如客户端/出站）才渲染展示 -->
      <v-col v-if="Object.hasOwn(data, 'password')" cols="12" sm="6" md="4" class="px-1 py-2">
        <v-text-field
          v-model="data.password"
          :label="$t('types.pw')"
          hide-details
          variant="outlined"
          class="dark-input"
          density="comfortable"
        ></v-text-field>
      </v-col>

      <!-- AnyTLS 混淆填充方案编辑框 -->
      <v-col cols="12" class="px-1 py-2">
        <v-textarea
          v-model="paddingSchemeText"
          :label="$t('types.anytls.paddingScheme')"
          rows="7"
          variant="outlined"
          class="dark-input"
          density="comfortable"
          hide-details
          placeholder="stop=8&#10;0=30-30&#10;1=100-400"
        ></v-textarea>
      </v-col>
    </v-row>
  </v-card>
</template>

<script lang="ts">
export default {
  props: ['data'],
  computed: {
    paddingSchemeText: {
      get() {
        if (Array.isArray(this.data.padding_scheme)) {
          return this.data.padding_scheme.join('\n');
        }
        return '';
      },
      set(val: string) {
        if (!val) {
          this.data.padding_scheme = [];
          return;
        }
        this.data.padding_scheme = val
          .split('\n')
          .map((line: string) => line.trim())
          .filter((line: string) => line.length > 0);
      }
    }
  }
}
</script>
