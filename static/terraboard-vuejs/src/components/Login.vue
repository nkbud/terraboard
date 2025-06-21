<template>
  <div class="login-container">
    <div class="login-card">
      <div class="card">
        <div class="card-header">
          <h3 class="text-center">
            <i class="fas fa-lock"></i> Login Required
          </h3>
        </div>
        <div class="card-body">
          <div class="text-center mb-4">
            <img src="/favicon.ico" alt="Terraboard" width="64" height="64" />
            <h4 class="mt-2">Terraboard</h4>
            <p class="text-muted">Please sign in to continue</p>
          </div>
          
          <div v-if="error" class="alert alert-danger" role="alert">
            <i class="fas fa-exclamation-triangle"></i>
            {{ error }}
          </div>
          
          <div class="d-grid">
            <button 
              @click="login" 
              :disabled="loading"
              class="btn btn-primary btn-lg"
            >
              <i v-if="loading" class="fas fa-spinner fa-spin me-2"></i>
              <i v-else class="fas fa-sign-in-alt me-2"></i>
              {{ loading ? 'Signing in...' : 'Sign in with OIDC' }}
            </button>
          </div>
          
          <div class="text-center mt-3">
            <small class="text-muted">
              You will be redirected to your identity provider
            </small>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { Options, Vue } from 'vue-class-component';

@Options({
  data() {
    return {
      loading: false,
      error: null
    };
  },
  methods: {
    async login() {
      this.loading = true;
      this.error = null;
      
      try {
        // Redirect to OIDC login endpoint
        window.location.href = '/auth/login';
      } catch (err) {
        this.error = 'Failed to initialize login. Please try again.';
        this.loading = false;
      }
    }
  }
})
export default class Login extends Vue {}
</script>

<style scoped>
.login-container {
  min-height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
}

.login-card {
  width: 100%;
  max-width: 400px;
}

.card {
  border: none;
  border-radius: 15px;
  box-shadow: 0 15px 35px rgba(0, 0, 0, 0.1);
  overflow: hidden;
}

.card-header {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  border: none;
  padding: 30px;
}

.card-body {
  padding: 40px;
}

.btn-primary {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border: none;
  border-radius: 10px;
  padding: 15px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 1px;
  transition: all 0.3s ease;
}

.btn-primary:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 8px 25px rgba(102, 126, 234, 0.4);
}

.btn-primary:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.alert {
  border-radius: 10px;
  border: none;
}
</style>