<template>
<div id="results" class="row">
    <div class="table-responsive">
      <table class="table table-border table-striped">
          <thead>
              <th></th> <!-- expandable -->
              <th>
                  Path
              </th>
              <th>
                  Resources
              </th>
              <th>
                  Activity
              </th>
          </thead>
          <tbody>
              <tr v-for="(node, index) in displayNodes" v-bind:key="node.path" 
                  :class="{ 'selected-row': selectedPrefix === node.path }"
                  @click="selectPrefix(node.path)">
                  <td class="align-middle">
                    <span v-if="node.hasChildren" 
                          @click.stop="toggleExpansion(node.path)"
                          class="fas"
                          :class="node.expanded ? 'fa-caret-down' : 'fa-caret-right'"
                          role="button"
                          style="cursor: pointer;"></span>
                  </td>
                  <td class="align-middle" :style="`padding-left: ${node.level * 20 + 10}px`">
                    <span v-if="node.states.length === 1">
                      <router-link :to="`/lineage/${node.states[0].lineage_value}`">{{node.displayPath}}</router-link>
                    </span>
                    <span v-else>{{node.displayPath}}</span>
                  </td>
                  <td class="align-middle">{{node.resourceCount}}</td>
                  <td class="text-center align-middle p-0">
                      <canvas v-bind:id="'spark-'+index" width="200" height="70" style="max-width: 200px; max-height: 70px;">
                        {{getAggregatedActivity(index, node, 'spark-'+index)}}
                      </canvas>
                  </td>
              </tr>
          </tbody>
      </table>
    </div>
</div>
</template>

<script lang="ts">
import { Options, Vue } from 'vue-class-component';
import { Chart, ChartItem, CategoryScale, PointElement,
LineController, LineElement, LinearScale, Tooltip } from 'chart.js'
import axios from "axios"
import apiCache from '@/services/ApiCache'

Chart.register( CategoryScale, LineElement, LineController, LinearScale, PointElement, Tooltip )

interface StateStat {
  path: string;
  lineage_value: string;
  terraform_version: string;
  serial: number;
  version_id: string;
  last_modified: string;
  resource_count: number;
}

interface PrefixNode {
  path: string;
  displayPath: string;
  level: number;
  hasChildren: boolean;
  expanded: boolean;
  resourceCount: number;
  states: StateStat[];
  children: Map<string, PrefixNode>;
}

@Options({
  data() {
    return {
      locksStatus: {},
      results: {},
      prefixTree: new Map<string, PrefixNode>(),
      displayNodes: [] as PrefixNode[],
      selectedPrefix: null as string | null,
      allStates: [] as StateStat[],
    }
  },
  emits: ['prefix-selected'],
  methods: {
    fetchLocks(): void {
      const cacheKey = 'api-locks';
      const cachedData = apiCache.get(cacheKey);
      
      if (cachedData) {
        this.locksStatus = cachedData;
        return;
      }
      
      const url = `/api/locks`;
      axios.get(url)
        .then((response) => {
          this.locksStatus = response.data;
          apiCache.set(cacheKey, response.data);
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
    
    extractPrefixPath(fullPath: string): string[] {
      // Extract meaningful prefix paths from S3 paths
      // e.g., "env/production/terraform.tfstate" -> ["env/", "env/production/"]
      const parts = fullPath.split('/');
      const prefixes: string[] = [];
      
      // Only create prefixes if there are at least 2 parts (directory + file)
      if (parts.length > 1) {
        for (let i = 1; i < parts.length; i++) {
          if (parts[i-1]) { // Skip empty parts
            const prefix = parts.slice(0, i).join('/') + '/';
            prefixes.push(prefix);
          }
        }
      }
      
      return prefixes;
    },
    
    buildPrefixTree(states: StateStat[]): void {
      this.prefixTree.clear();
      
      states.forEach(state => {
        const prefixes = this.extractPrefixPath(state.path);
        
        prefixes.forEach((prefix: string, index: number) => {
          if (!this.prefixTree.has(prefix)) {
            // Create display path by taking the last meaningful part
            const pathParts = prefix.replace(/\/$/, '').split('/');
            const displayPath = pathParts.length > 0 ? pathParts[pathParts.length - 1] + '/' : prefix;
            
            this.prefixTree.set(prefix, {
              path: prefix,
              displayPath: displayPath,
              level: index,
              hasChildren: false,
              expanded: false,
              resourceCount: 0,
              states: [],
              children: new Map()
            });
          }
          
          const node = this.prefixTree.get(prefix)!;
          if (!node.states.find((s: StateStat) => s.path === state.path)) {
            node.states.push(state);
            node.resourceCount += state.resource_count;
          }
          
          // Mark parent as having children
          if (index > 0) {
            const parentPrefix = prefixes[index - 1];
            const parentNode = this.prefixTree.get(parentPrefix);
            if (parentNode) {
              parentNode.hasChildren = true;
              parentNode.children.set(prefix, node);
            }
          }
        });
      });
      
      this.updateDisplayNodes();
    },
    
    updateDisplayNodes(): void {
      this.displayNodes = [];
      
      // Get top-level nodes (level 0)
      const topLevelNodes = Array.from(this.prefixTree.values()).filter((node: any) => node.level === 0) as PrefixNode[];
      topLevelNodes.sort((a: PrefixNode, b: PrefixNode) => a.path.localeCompare(b.path));
      
      topLevelNodes.forEach(node => {
        this.addNodeToDisplay(node);
      });
    },
    
    addNodeToDisplay(node: PrefixNode): void {
      this.displayNodes.push(node);
      
      if (node.expanded && node.hasChildren) {
        const children = Array.from(node.children.values());
        children.sort((a, b) => a.path.localeCompare(b.path));
        children.forEach(child => {
          this.addNodeToDisplay(child);
        });
      }
    },
    
    toggleExpansion(path: string): void {
      const node = this.prefixTree.get(path);
      if (node) {
        node.expanded = !node.expanded;
        this.updateDisplayNodes();
      }
    },
    
    selectPrefix(path: string): void {
      if (this.selectedPrefix === path) {
        this.selectedPrefix = null;
      } else {
        this.selectedPrefix = path;
      }
      this.$emit('prefix-selected', this.selectedPrefix);
    },
    
    getAggregatedActivity(idx: number, node: PrefixNode, elementId: string): void {
      // Check if we already have a cached aggregated activity for this node
      const cacheKey = `activity-${node.path}`;
      const cachedActivity: any = apiCache.get(cacheKey);
      
      if (cachedActivity && cachedActivity.labels && cachedActivity.data) {
        this.createSparkChart(elementId, cachedActivity.labels, cachedActivity.data);
        return;
      }
      
      // Aggregate activity data from all states under this prefix
      const activityPromises: Promise<any>[] = [];
      
      node.states.forEach(state => {
        const stateCacheKey = `activity-state-${state.lineage_value}`;
        const cachedStateActivity = apiCache.get(stateCacheKey);
        
        if (cachedStateActivity) {
          activityPromises.push(Promise.resolve(cachedStateActivity));
        } else {
          const promise = axios.get(`/api/lineages/${state.lineage_value}/activity`)
            .then(response => {
              apiCache.set(stateCacheKey, response.data);
              return response.data;
            })
            .catch(err => {
              console.log("Activity fetch error:", err);
              return [];
            });
          activityPromises.push(promise);
        }
      });
      
      Promise.all(activityPromises).then(results => {
        const aggregatedData = new Map<string, number>();
        
        results.forEach(stateActivity => {
          if (Array.isArray(stateActivity)) {
            stateActivity.forEach((activity: any) => {
              const date = this.formatDate(activity.last_modified);
              const count = aggregatedData.get(date) || 0;
              aggregatedData.set(date, count + (activity.resource_count || 0));
            });
          }
        });
        
        // Sort dates and prepare chart data
        const sortedDates = Array.from(aggregatedData.keys()).sort((a, b) => {
          return new Date(a).getTime() - new Date(b).getTime();
        });
        
        const labels = sortedDates;
        const data = sortedDates.map(date => aggregatedData.get(date)!.toString());
        
        // Cache the aggregated result
        const activityResult = { labels, data };
        apiCache.set(cacheKey, activityResult);
        
        this.createSparkChart(elementId, labels, data);
      });
    },
    
    formatDate(date: string): string {
        // Create a more consistent date format for aggregation
        const dateObj = new Date(date);
        return dateObj.toISOString().split('T')[0]; // YYYY-MM-DD format
    },
    
    createSparkChart(id: string, labels: string[], data: string[]): void {
      const ctx = document.getElementById(id) as ChartItem;
      if (!ctx) return;
      
      const sparkchart = new Chart(ctx, {
        type: 'line',
        data: {
          labels: labels,
          datasets: [
            {
              data: data
            }
          ]
        },
        options: {
          responsive: true,
          elements: {
            line: {
              borderColor: '#4dc9f6',
              borderWidth: 1
            },
            point: {
              radius: 1
            }
          },
          scales: {
            yAxes:
              {
                display: true,
                ticks: {
                  stepSize: 1
                }
              },
            xAxes:
              {
                display: false
              }
          },
          plugins: {
            legend: {
              display: false
            },
            tooltip: {
              enabled: true
            },
          }
        }
      });
    },
    
    updatePager(response: any): void {
      this.results = response.data;
      this.allStates = response.data.states || [];
    },
    
    fetchStats(): void {
      const cacheKey = 'api-lineages-stats-all';
      const cachedData = apiCache.get(cacheKey);
      
      if (cachedData) {
        this.updatePager(cachedData);
        this.buildPrefixTree(this.allStates);
        return;
      }
      
      const url = `/api/lineages/stats`; // Fetch all states without pagination
      axios.get(url)
        .then((response) => {
          apiCache.set(cacheKey, response);
          this.updatePager(response);
          this.buildPrefixTree(this.allStates);
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
    }
  },
  created() {
    this.fetchLocks();
    this.fetchStats();
  },
})
export default class StatesListV2 extends Vue {}
</script>

<style scoped lang="scss">
.selected-row {
  background-color: #d9edf7 !important;
  color: #337ab7 !important;
}

.selected-row:hover {
  background-color: #c4e3f3 !important;
}

tbody tr {
  cursor: pointer;
}

tbody tr:hover {
  background-color: #f5f5f5;
}
</style>