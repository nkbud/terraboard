<template>
  <!-- Outputs view -->
  <div v-if="module.outputs !== undefined" class="mt-3">
    <h3 class="node-title">
      Outputs for {{module.path ? module.path : "root"}}
    </h3>
    <div class="panel-group">
      <table class="table">
        <thead>
          <th>Name</th>
          <th>Value</th>
        </thead>
        <tbody>
          <tr v-for="out in sortedOutputs" v-bind:key="out">
            <td class="attr-key">{{ out.name }}</td>
            <td class="attr-val">{{ displayValue(out) }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script lang="ts">
import { Options, Vue } from "vue-class-component";

@Options({
  props: {
    module: {},
  },
  data() {
    return {
      redactSensitive: true,
    };
  },
  methods: {
    displayValue(out: any): string {
      if (this.redactSensitive && out.sensitive) {
        if (out.value == "null") {
          return "(null)";
        }
        return "(" + out.value.length + ")";
      }
      return out.value;
    },
  },
  computed: {
    sortedOutputs() {
      if (this.module.outputs !== undefined) {
        return this.module.outputs.sort((a: any, b: any) => {
          return a.name.localeCompare(b.name);
        });
      }
    },
  },
})
export default class StateOutputs extends Vue {}
</script>

<style lang="scss"></style>
