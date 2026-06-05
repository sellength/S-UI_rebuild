<template>
  <v-dialog transition="dialog-bottom-transition" width="800">
    <v-card class="panel-modal pa-4" style="border-radius: 12px; max-height: 90vh; display: flex; flex-direction: column;">
      <v-card-title class="px-2 pb-2">
        <span class="text-h6 font-weight-bold text-grey-lighten-3">
          {{ $t('actions.' + title) + " " + $t('objects.rule') }}
        </span>
      </v-card-title>
      <v-divider class="mb-4" style="opacity: 0.1;"></v-divider>
      
      <v-card-text style="padding: 0 20px; overflow-y: auto;" class="flex-grow-1">
        <v-container style="padding: 0;">
          <v-row align="center" class="mb-4">
            <v-col cols="12" sm="6" class="py-1">
              <v-switch color="cyan" v-model="logical" :label="$t('rule.logical')" hide-details></v-switch>
            </v-col>
            <v-spacer></v-spacer>
            <v-col cols="auto" v-if="logical" class="py-1">
              <v-btn class="tech-blue-btn text-none" size="small" prepend-icon="mdi-plus" @click="ruleData.rules.push({})">
                {{ $t('actions.add') + " " + $t('objects.rule') }}
              </v-btn>
            </v-col>
          </v-row>

          <!-- 逻辑多规则嵌套列表 -->
          <template v-if="ruleData.type == 'logical'">
            <div 
              v-for="(r, index) in ruleData.rules" 
              :key="index" 
              class="flat-card pa-4 mb-4" 
              style="border-radius: 8px;"
            >
              <div class="text-subtitle-2 text-grey-lighten-2 d-flex align-center justify-space-between mb-2">
                <span>{{ $t('objects.rule') + ' ' + (index+1) }}</span>
                <v-btn 
                  v-if="ruleData.rules.length > 1" 
                  icon="mdi-delete-outline" 
                  color="error" 
                  variant="text" 
                  size="small" 
                  @click="ruleData.rules.splice(index,1)"
                ></v-btn>
              </div>
              <v-divider class="mb-4" style="opacity: 0.1;"></v-divider>
              <RuleOptions
                :rule="r"
                :clients="clients"
                :inTags="inTags"
                :rsTags="rsTags" 
              />
            </div>
          </template>
          
          <template v-else>
            <div class="flat-card pa-4 mb-4" style="border-radius: 8px;">
              <RuleOptions
                :rule="ruleData.rules[0]"
                :clients="clients"
                :inTags="inTags"
                :rsTags="rsTags" 
              />
            </div>
          </template>

          <v-row style="row-gap: 16px; margin-top: 10px;">
            <v-col cols="12" sm="4" class="py-1">
              <v-combobox
                v-model="ruleData.outbound"
                :items="outTags"
                :label="$t('objects.outbound')"
                hide-details
                variant="outlined"
                class="dark-input"
                density="comfortable"
              ></v-combobox>
            </v-col>
            <v-col cols="12" sm="4" v-if="logical" class="py-1">
              <v-combobox
                v-model="ruleData.mode"
                :items="['and', 'or']"
                :label="$t('rule.mode')"
                hide-details
                variant="outlined"
                class="dark-input"
                density="comfortable"
              ></v-combobox>
            </v-col>
            <v-col cols="12" sm="4" class="py-1 d-flex align-center">
              <v-switch color="cyan" v-model="ruleData.invert" :label="$t('rule.invert')" hide-details></v-switch>
            </v-col>
          </v-row>
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
import { logicalRule, rule } from '@/types/rules'
import RuleOptions from '@/components/Rule.vue'
export default {
  props: ['visible', 'data', 'index', 'clients', 'inTags', 'outTags', 'rsTags'],
  emits: ['close', 'save'],
  data() {
    return {
      title: 'add',
      loading: false,
      ruleData: <logicalRule>{
        type: 'logical',
        mode: 'and',
        rules: <rule[]>[{}],
        invert: false,
        outbound: 'direct',
      }
    }
  },
  methods: {
    updateData() {
      if (this.$props.index != -1) {
        const newData = JSON.parse(this.$props.data)
        if (newData.type) {
          this.ruleData = newData
        } else {
          this.ruleData = <logicalRule>{
            type: 'simple',
            mode: 'and',
            rules: <rule[]>[{...newData}],
            invert: newData.invert,
            outbound: newData.outbound,
          }
        }
        this.title = 'edit'
      }
      else {
        this.ruleData = <logicalRule>{
            type: 'simple',
            mode: 'and',
            rules: <rule[]>[{}],
            invert: false,
            outbound: this.$props.outTags[0]?? 'direct',
          }
        this.title = 'add'
      }
    },
    closeModal() {
      this.updateData() // reset
      this.$emit('close')
    },
    saveChanges() {
      this.loading = true
      if (this.ruleData.type == 'simple'){
        this.ruleData.rules[0].outbound = this.ruleData.outbound
        this.ruleData.rules[0].invert = this.ruleData.invert
      }
      this.$emit('save', this.ruleData)
      this.loading = false
    },
    deleteRule(index:number) {
      this.ruleData.rules.splice(index,1)
    }
  },
  computed: {
    logical: {
      get() { return this.ruleData.type == 'logical' },
      set(v:boolean) {
        if (v) {
          this.ruleData.type = 'logical'
          this.ruleData.outbound = this.ruleData.rules[0].outbound?? this.$props.outTags[0]
          delete this.ruleData.rules[0].outbound
        } else {
          this.ruleData.type = 'simple'
          this.ruleData.rules[0].outbound = this.ruleData.outbound
        }
      }
    }
  },
  watch: {
    visible(newValue) {
      if (newValue) {
        this.updateData()
      }
    },
  },
  components: { RuleOptions }
}
</script>
