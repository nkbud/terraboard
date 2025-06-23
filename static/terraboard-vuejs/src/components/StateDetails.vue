<template>
  <!-- Resource details view -->
  <div v-if="resource.attributes !== undefined" class="mt-3">
    <h3 class="node-title">
      {{ resource.type }}.{{ resource.name }}{{ resource.index }}
    </h3>
    <div class="panel-group">
      <div class="card">
        <h4 class="card-header">
          Attributes
          <!-- <div class="float-end">
            <label class="form-check-label me-2" for="redactSensitiveToggle">
              Hide sensitive values
            </label>
            <input 
              disabled
              class="form-check-input" 
              type="checkbox" 
              id="redactSensitiveToggle"
              v-model="redactSensitive"
            />
          </div> -->
        </h4>
        <table class="table">
          <thead>
            <th>Attribute</th>
            <th>Value</th>
          </thead>
          <tbody>
            <tr v-for="attr in sortedAttributes" v-bind:key="attr">
              <td class="attr-key">{{ attr.key }}</td>
              <td class="attr-val">{{ displayValue(attr) }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { Options, Vue } from "vue-class-component";

@Options({
  props: {
    resource: {},
  },
  data() {
    return {
      redactSensitive: true, // Default to redacting sensitive values
    };
  },
  computed: {
    sortedAttributes() {
      if (this.resource.attributes !== undefined) {
        return this.resource.attributes.sort((a: any, b: any) => {
          return a.key.localeCompare(b.key);
        });
      }
    },
  },
  methods: {
    displayValue(attr: any): string {
      if (this.redactSensitive && attr.sensitive) {
        if (attr.value == "null") {
          return "(null)";
        }
        return "(" + attr.value.length + ")";
      }
      return attr.value;
    },
  },
})
export default class StateDetails extends Vue {}
</script>

<style lang="scss"></style>
