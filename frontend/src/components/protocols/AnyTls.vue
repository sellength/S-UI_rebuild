<template>
  <v-card subtitle="AnyTLS">
    <v-row>
      <v-col cols="12" sm="6" md="4">
        <v-text-field v-model="data.password" :label="$t('types.pw')" hide-details variant="outlined" class="dark-input" density="comfortable"></v-text-field>
      </v-col>
    </v-row>
  </v-card>
</template>

<script lang="ts">
export default {
  props: ['data'],
  data() {
    return {}
  }
}
</script>
