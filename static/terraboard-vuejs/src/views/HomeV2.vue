<template>
  <ChartsV2 :selectedPrefix="selectedPrefix"/>
  <hr>
  <StatesListV2 @prefix-selected="onPrefixSelected"/>
</template>

<script lang="ts">
import { Options, Vue } from 'vue-class-component';
import ChartsV2 from '@/components/ChartsV2.vue'; // @ is an alias to /src
import StatesListV2 from '@/components/StatesListV2.vue'; // @ is an alias to /src

@Options({
  title: 'OverviewV2',
  components: {
    ChartsV2,
    StatesListV2,
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
export default class HomeV2 extends Vue {}
</script>