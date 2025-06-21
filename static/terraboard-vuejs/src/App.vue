<template>
  <div id="app">
    <div v-if="loading" class="loading-container">
      <div class="text-center">
        <i class="fas fa-spinner fa-spin fa-3x text-primary mb-3"></i>
        <h4>Loading Terraboard...</h4>
      </div>
    </div>
    
    <div v-else-if="!isAuthenticated && isOIDCEnabled && !isCallbackRoute" class="login-required">
      <Login />
    </div>
    
    <div v-else>
      <Navbar v-if="isAuthenticated" />
      <div id="maincont" class="container" :class="{ 'no-auth': !isAuthenticated }">
        <router-view @refresh="this.refresh()" :key="this.hash"/>
      </div>
      <AppFooter v-if="isAuthenticated" />
    </div>
  </div>
</template>

<script lang="ts">
import { Options, Vue } from 'vue-class-component';
import Navbar from '@/components/Navbar.vue';
import AppFooter from '@/components/Footer.vue';
import Login from '@/components/Login.vue';
import axios from 'axios';

@Options({
  components: {
    Navbar,
    AppFooter,
    Login,
  },
  data() {
    return {
      hash: 1,
      loading: true,
      isAuthenticated: false,
      isOIDCEnabled: false,
      user: null
    }
  },
  computed: {
    isCallbackRoute() {
      return this.$route.path === '/callback';
    }
  },
  async created() {
    await this.checkAuthentication();
  },
  methods: {
    refresh() {
      this.hash++;
    },
    
    async checkAuthentication() {
      try {
        const response = await axios.get('/api/user');
        const userData = response.data;
        
        this.user = userData;
        this.isAuthenticated = userData.authenticated || false;
        this.isOIDCEnabled = userData.is_oidc || false;
        
        // If not authenticated and OIDC is enabled, but we have X-Forwarded headers,
        // then we're using proxy auth and should be authenticated
        if (!this.isAuthenticated && !this.isOIDCEnabled && userData.name) {
          this.isAuthenticated = true;
        }
        
      } catch (error) {
        console.error('Failed to check authentication:', error);
        this.isAuthenticated = false;
      } finally {
        this.loading = false;
      }
    }
  },
  watch: {
    '$route'() {
      // Re-check authentication when route changes (useful for callback)
      if (this.isCallbackRoute) {
        setTimeout(() => {
          this.checkAuthentication();
        }, 2000);
      }
    }
  }
})
export default class App extends Vue {}
</script>

<style lang="scss">
a {
  text-decoration: none !important;
}

.loading-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f8f9fa;
}

.login-required {
  min-height: 100vh;
}

#maincont.no-auth {
  margin-top: 0;
  padding-top: 20px;
}

@media (max-width:767px) {
  .navbar-toggle {
    float: left;
  }
  .navbar-right {
    position: absolute;
    top: 0;
    right: 20px;
  }
  .breadcrumb {
    display: none;
  }
  h2.node-title {
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
}

@media (min-width: 992px) {
  body, html, #mainrow, #leftcol {
    margin: 0;
    height: 100%;
  }
  .container {
    max-width: 75vw;
  }
  #maincont, #leftcol .panel-group {
    overflow: hidden;
  }
  #maincont {
    margin-bottom: 50px;
  }
  #nodes, #node {
    overflow-y: auto;
  }
  #nodes {
    max-height: calc(100% - 200px);
  }
  #node {
    height: 100%;
  }
  #states-select {
    margin: 10px;
  }
  #states-select a {
    color: black;
  }
}
</style>
