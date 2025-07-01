<template>
<div class="row justify-content-around">
    <div class="overview-chart col-6 col-md-3 col-xxl-4 text-center" style="min-width: 100px; max-width: 300px;">
        <canvas id="chart-pie-resource-types-v2" class="chart mb-2"></canvas>
        <h5>Resource types{{ selectedPrefix ? ' (' + selectedPrefix + ')' : '' }}</h5>
    </div>
    <div class="overview-chart col-6 col-md-3 col-xxl-4 text-center" style="min-width: 100px; max-width: 300px;">
        <canvas id="chart-pie-terraform-versions-v2" class="chart mb-2"></canvas>
        <h5>Terraform versions{{ selectedPrefix ? ' (' + selectedPrefix + ')' : '' }}</h5>
    </div>
    <div class="overview-chart col-6 col-md-3 col-xxl-4 text-center" style="min-width: 100px; max-width: 300px;">
        <canvas id="chart-pie-ls-v2" class="chart mb-2"></canvas>
        <h5>States locked{{ selectedPrefix ? ' (' + selectedPrefix + ')' : '' }}</h5>
    </div>
</div>
</template>

<script lang="ts">
import { Options, Vue } from 'vue-class-component';
import { Chart, ChartItem, PieController, ArcElement, Tooltip } from 'chart.js'
import axios from "axios"
import router from "../router";
import apiCache from '@/services/ApiCache'

Chart.register( PieController, ArcElement, Tooltip )

const chartOptionsVersions = 
{
  onClick: undefined,
  responsive: true,
  plugins: {
    legend: {
      display: false,
    },
    tooltip: {
      display: true,
    },
  } 
}
const chartOptionsResTypes = 
{
  onClick: undefined,
  responsive: true,
  plugins: {
    legend: {
      display: false,
    },
    tooltip: {
      display: true,
    },
  } 
}
const chartOptionsLocked = 
{
  responsive: true,
  plugins: {
    legend: {
      display: false,
    },
    tooltip: {
      display: true,
    },
  } 
}

@Options({
  props: ['selectedPrefix'],
  data() {
    return {
      locks: {},
      statesTotal: 0,
      allStates: [] as any[],
      resourceTypesChart: null as Chart | null,
      versionsChart: null as Chart | null,
      locksChart: null as Chart | null,
      pieResourceTypes: {
        labels: [[], [], [], [], [], [], ["Total"]],
        data: [0, 0, 0, 0, 0, 0, 0],
        options: chartOptionsResTypes,
      },
      pieTfVersions: {
        labels: [[], [], [], [], [], [], ["Total"]],
        data: [0, 0, 0, 0, 0, 0, 0],
        options: chartOptionsVersions,
      },
      pieLockedStates: {
        labels: ["Locked", "Unlocked"],
        data: [0, 0],
        options: chartOptionsLocked,
      },
    };
  },
  watch: {
    selectedPrefix: {
      handler: function() {
        this.updateChartsForSelectedPrefix();
      },
    }
  },
  methods: {
    isLocked(path: string): boolean {
      if (path in this.locks) {
          return true;
      }
      return false;
    },
    
    getFilteredStates(): any[] {
      if (!this.selectedPrefix) {
        return this.allStates;
      }
      
      return this.allStates.filter((state: any) => state.path.startsWith(this.selectedPrefix));
    },
    
    updateChartsForSelectedPrefix(): void {
      const filteredStates = this.getFilteredStates();
      
      // Update resource types chart
      this.fetchResourceTypesForStates(filteredStates);
      
      // Update terraform versions chart  
      this.fetchVersionsForStates(filteredStates);
      
      // Update locks chart
      this.updateLocksChart(filteredStates);
    },
    
    fetchResourceTypesForStates(states: any[]): void {
      // Get unique lineage values from filtered states
      const lineages = [...new Set(states.map(state => state.lineage_value))];
      
      if (lineages.length === 0) {
        this.updateResourceTypesChart([]);
        return;
      }
      
      // Build query parameter for filtered lineages
      const lineageQuery = lineages.map(l => `lineage=${encodeURIComponent(l)}`).join('&');
      const url = `/api/resource/types/count?${lineageQuery}`;
      const cacheKey = `resource-types-filtered-${lineages.join(',')}`;
      
      const cachedData = apiCache.get(cacheKey);
      if (cachedData) {
        this.updateResourceTypesChart(cachedData);
        return;
      }
      
      axios.get(url)
        .then((response) => {
          apiCache.set(cacheKey, response.data);
          this.updateResourceTypesChart(response.data);
        })
        .catch((err) => {
          console.log("Resource types fetch error:", err);
          // Fallback: use all data if filtered query fails
          if (this.selectedPrefix) {
            this.fetchResourceTypes();
          }
        });
    },
    
    fetchVersionsForStates(states: any[]): void {
      // Calculate version counts from filtered states
      const versionCounts = new Map<string, number>();
      
      states.forEach(state => {
        const version = state.terraform_version;
        versionCounts.set(version, (versionCounts.get(version) || 0) + 1);
      });
      
      const versionData = Array.from(versionCounts.entries()).map(([version, count]) => ({
        name: version,
        count: count.toString()
      }));
      
      versionData.sort((a, b) => a.name.localeCompare(b.name));
      
      this.updateVersionsChart(versionData);
    },
    
    updateLocksChart(states: any[]): void {
      let lockedCount = 0;
      states.forEach(state => {
        if (this.isLocked(state.path)) {
          lockedCount++;
        }
      });
      
      this.pieLockedStates.data[0] = lockedCount;
      this.pieLockedStates.data[1] = states.length - lockedCount;
      
      if (this.locksChart) {
        this.locksChart.data.datasets[0].data = [...this.pieLockedStates.data];
        this.locksChart.update();
      }
    },
    
    updateResourceTypesChart(data: any[]): void {
      // Reset arrays
      this.pieResourceTypes.labels = [[], [], [], [], [], [], ["Total"]];
      this.pieResourceTypes.data = [0, 0, 0, 0, 0, 0, 0];
      
      data.forEach((value: any, i: number) => {
        if(i < 6) {
            this.pieResourceTypes.labels[i] = value.name;
            this.pieResourceTypes.data[i] = parseInt(value.count, 10);
        } else {
            this.pieResourceTypes.labels[6].push(value.name+": "+value.count);
            this.pieResourceTypes.data[6] += parseInt(value.count, 10);
        }
      });
      
      if (this.resourceTypesChart) {
        this.resourceTypesChart.data.labels = [...this.pieResourceTypes.labels];
        this.resourceTypesChart.data.datasets[0].data = [...this.pieResourceTypes.data];
        this.resourceTypesChart.update();
      }
    },
    
    updateVersionsChart(data: any[]): void {
      // Reset arrays
      this.pieTfVersions.labels = [[], [], [], [], [], [], ["Total"]];
      this.pieTfVersions.data = [0, 0, 0, 0, 0, 0, 0];
      
      data.forEach((value: any, i: number) => {
        if(i < 6) {
            this.pieTfVersions.labels[i] = [value.name];
            this.pieTfVersions.data[i] = parseInt(value.count, 10);
        } else {
            this.pieTfVersions.labels[6].push(value.name+": "+value.count);
            this.pieTfVersions.data[6] += parseInt(value.count, 10);
        }
      });
      
      if (this.versionsChart) {
        this.versionsChart.data.labels = [...this.pieTfVersions.labels];
        this.versionsChart.data.datasets[0].data = [...this.pieTfVersions.data];
        this.versionsChart.update();
      }
    },
    
    searchType(evt: any, element: any) {
      let valueIndex = element[0].index;
      router.push({name: "Search", query: { type: this.pieResourceTypes.labels[valueIndex] }});
    },
    
    searchVersion(evt: any, element: any) {
      let valueIndex = element[0].index;
      router.push({name: "Search", query: { tf_version: this.pieTfVersions.labels[valueIndex] }});
    },
    
    fetchResourceTypes(): void {
      const cacheKey = 'api-resource-types-count';
      const cachedData = apiCache.get(cacheKey);
      
      if (cachedData) {
        this.updateResourceTypesChart(cachedData);
        this.pieResourceTypes.options.onClick = this.searchType;
        this.createResourceTypesChart();
        return;
      }
      
      const url = `/api/resource/types/count`;
      axios.get(url)
        .then((response) => {
          apiCache.set(cacheKey, response.data);
          this.updateResourceTypesChart(response.data);
          
          this.pieResourceTypes.options.onClick = this.searchType;
          this.createResourceTypesChart();
        })
        .catch(function (err) {
          if (err.response) {
            console.log("Server Error:", err)
          } else if (err.request) {
            console.log("Network Error:", err)
          } else {
            console.log("Client Error:", err)
          }
        })
        .then(function () {
          // always executed
        });
    },

    createResourceTypesChart(): void {
      const ctx = document.getElementById('chart-pie-resource-types-v2') as ChartItem;
      if (ctx) {
        this.resourceTypesChart = new Chart(ctx, {
            type: 'pie',
            data: {
                labels: this.pieResourceTypes.labels,
                datasets: [{
                    label: 'States Resources Type',
                    data: this.pieResourceTypes.data,
                    backgroundColor: [
                      '#4dc9f6',
                      '#f67019',
                      '#f53794',
                      '#537bc4',
                      '#acc236',
                      '#166a8f',
                      '#00a950',
                    ],
                    hoverOffset: 4
                }]
            },
            options: this.pieResourceTypes.options
        });
      }
    },
    
    fetchVersions(): void {
      const cacheKey = 'api-terraform-versions-count';
      const cachedData = apiCache.get(cacheKey);
      
      if (cachedData) {
        this.updateVersionsChart(cachedData);
        this.pieTfVersions.options.onClick = this.searchVersion;
        this.createVersionsChart();
        return;
      }
      
      const url = `/api/lineages/tfversion/count?orderBy=version`;
      axios.get(url)
        .then((response) => {
          apiCache.set(cacheKey, response.data);
          this.updateVersionsChart(response.data);

          this.pieTfVersions.options.onClick = this.searchVersion;
          this.createVersionsChart();
        })
        .catch(function (err) {
          if (err.response) {
            console.log("Server Error:", err)
          } else if (err.request) {
            console.log("Network Error:", err)
          } else {
            console.log("Client Error:", err)
          }
        })
        .then(function () {
          // always executed
        });
    },

    createVersionsChart(): void {
      const ctx = document.getElementById('chart-pie-terraform-versions-v2') as ChartItem;
      if (ctx) {
        this.versionsChart = new Chart(ctx, {
            type: 'pie',
            data: {
                labels: this.pieTfVersions.labels,
                datasets: [{
                    label: 'States Versions',
                    data: this.pieTfVersions.data,
                    backgroundColor: [
                      '#4dc9f6',
                      '#f67019',
                      '#f53794',
                      '#537bc4',
                      '#acc236',
                      '#166a8f',
                      '#00a950',
                    ],
                    hoverOffset: 4
                }]
            },
            options: this.pieTfVersions.options
        });
      }
    },
    
    fetchLocks(): void {
      const cacheKey = 'api-locks';
      const cachedData = apiCache.get(cacheKey);
      
      if (cachedData) {
        this.locks = cachedData;
        this.createLocksChart();
        return;
      }
      
      const url = `/api/locks`;
      axios.get(url)
        .then((response) => {
          this.locks = response.data;
          apiCache.set(cacheKey, response.data);
          this.createLocksChart();
        })
        .catch(function (err) {
          if (err.response) {
            console.log("Server Error:", err)
          } else if (err.request) {
            console.log("Network Error:", err)
          } else {
            console.log("Client Error:", err)
          }
        })
        .then(function () {
          // always executed
        });
    },

    createLocksChart(): void {
      const ctx = document.getElementById('chart-pie-ls-v2') as ChartItem;
      if (ctx) {
        this.locksChart = new Chart(ctx, {
            type: 'pie',
            data: {
                labels: this.pieLockedStates.labels,
                datasets: [{
                    label: 'States Locks Status',
                    data: this.pieLockedStates.data,
                    backgroundColor: [
                      '#f67019',
                      '#4dc9f6',
                    ],
                    hoverOffset: 4
                }]
            },
            options: this.pieLockedStates.options
        });
      }
    },
    
    fetchAllStates(): void {
      // Try to get states from cache first (shared with StatesListV2)
      const cacheKey = 'api-lineages-stats-all';
      const cachedData: any = apiCache.get(cacheKey);
      
      if (cachedData && cachedData.data) {
        this.allStates = cachedData.data.states || [];
        this.statesTotal = cachedData.data.total || 0;
        this.fetchLocks();
        return;
      }
      
      // Fallback: make API call if no cached data
      const url = `/api/lineages/stats?limit=10000`;
      axios.get(url)
        .then((response) => {
          apiCache.set(cacheKey, response);
          this.allStates = response.data.states || [];
          this.statesTotal = response.data.total || 0;
          this.fetchLocks();
        })
        .catch(function (err) {
          if (err.response) {
            console.log("Server Error:", err)
          } else if (err.request) {
            console.log("Network Error:", err)
          } else {
            console.log("Client Error:", err)
          }
        });
    },
  },
  created() {
    this.fetchResourceTypes();
    this.fetchVersions();
    this.fetchAllStates();
  },
})
export default class ChartsV2 extends Vue {}
</script>

<style scoped lang="scss">

</style>