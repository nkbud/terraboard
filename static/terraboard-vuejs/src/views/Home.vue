<template>
  <Charts :selectedPrefix="selectedPrefix"/>
  <hr>
  <StatesList @prefix-selected="onPrefixSelected"/>
</template>

<script lang="ts">
import { Options, Vue } from 'vue-class-component';
import Charts from '@/components/Charts.vue'; // @ is an alias to /src
import StatesList from '@/components/StatesList.vue'; // @ is an alias to /src

@Options({
  title: 'Overview',
  components: {
    Charts,
    StatesList,
  },
  data() {
    return {
      selectedPrefix: null,
    };
  },
  methods: {
    onPrefixSelected(prefix: string | null) {
      this.selectedPrefix = prefix;
    },
  },
  emits: ["refresh"],
  created() {
    this.updateTitle();
  },
})
export default class Home extends Vue {}
</script>