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
          </thead>
          <tbody>
              <tr v-for="node in displayNodes" v-bind:key="node.path" 
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
              </tr>
          </tbody>
      </table>
    </div>
</div>
</template>

<script lang="ts">
import { Options, Vue } from 'vue-class-component';
import axios from "axios"
import apiCache from '@/services/ApiCache'

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
      
      // Get top-level nodes (level 0) and sort by resource count descending
      const topLevelNodes = Array.from(this.prefixTree.values()).filter((node: any) => node.level === 0) as PrefixNode[];
      topLevelNodes.sort((a: PrefixNode, b: PrefixNode) => b.resourceCount - a.resourceCount);
      
      topLevelNodes.forEach(node => {
        this.addNodeToDisplay(node);
      });
    },
    
    addNodeToDisplay(node: PrefixNode): void {
      this.displayNodes.push(node);
      
      if (node.expanded && node.hasChildren) {
        const children = Array.from(node.children.values());
        children.sort((a, b) => b.resourceCount - a.resourceCount); // Sort children by resource count descending too
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
export default class StatesList extends Vue {}
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