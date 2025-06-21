<template>
  <div class="callback-container">
    <div class="callback-card">
      <div class="card">
        <div class="card-body text-center">
          <div v-if="loading">
            <i class="fas fa-spinner fa-spin fa-3x text-primary mb-3"></i>
            <h4>Processing Authentication...</h4>
            <p class="text-muted">Please wait while we complete your login.</p>
          </div>
          
          <div v-else-if="error" class="text-danger">
            <i class="fas fa-exclamation-triangle fa-3x mb-3"></i>
            <h4>Authentication Failed</h4>
            <p>{{ error }}</p>
            <button @click="retryLogin" class="btn btn-primary">
              <i class="fas fa-redo me-2"></i>
              Try Again
            </button>
          </div>
          
          <div v-else class="text-success">
            <i class="fas fa-check-circle fa-3x mb-3"></i>
            <h4>Authentication Successful</h4>
            <p>Redirecting you to Terraboard...</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { Options, Vue } from 'vue-class-component';
import router from '../router';

@Options({
  data() {
    return {
      loading: true,
      error: null
    };
  },
  async mounted() {
    await this.handleCallback();
  },
  methods: {
    async handleCallback() {
      try {
        // The backend handles the actual callback processing
        // We just need to check if the user is now authenticated
        await this.checkAuthStatus();
      } catch (err) {
        console.error('Callback error:', err);
        this.error = 'Authentication failed. Please try again.';
        this.loading = false;
      }
    },
    
    async checkAuthStatus() {
      // Give the backend a moment to process the callback
      await new Promise(resolve => setTimeout(resolve, 1000));
      
      try {
        const response = await fetch('/api/user');
        const user = await response.json();
        
        if (user.authenticated) {
          // Authentication successful, redirect to home
          setTimeout(() => {
            router.push('/');
          }, 1500);
        } else {
          this.error = 'Authentication verification failed.';
          this.loading = false;
        }
      } catch (err) {
        this.error = 'Failed to verify authentication status.';
        this.loading = false;
      }
    },
    
    retryLogin() {
      window.location.href = '/auth/login';
    }
  }
})
export default class Callback extends Vue {}
</script>

<style scoped>
.callback-container {
  min-height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
}

.callback-card {
  width: 100%;
  max-width: 400px;
}

.card {
  border: none;
  border-radius: 15px;
  box-shadow: 0 15px 35px rgba(0, 0, 0, 0.1);
  overflow: hidden;
}

.card-body {
  padding: 50px 40px;
}

.btn-primary {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border: none;
  border-radius: 10px;
  padding: 10px 20px;
  font-weight: 600;
}
</style>