<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const auth = useAuthStore()
const code = ref('')
const loading = ref(false)

async function submit() {
  if (!code.value.trim()) {
    ElMessage.warning('请输入进门密码')
    return
  }
  loading.value = true
  try {
    await auth.verifyGate(code.value.trim())
    ElMessage.success('验证通过')
    await router.replace('/login')
  } catch (e: unknown) {
    ElMessage.error(e instanceof Error ? e.message : '验证失败')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="auth-page">
    <el-card class="auth-card" shadow="hover">
      <h2>进门验证</h2>
      <p class="hint">请输入小团体共享的进门密码</p>
      <el-form @submit.prevent="submit">
        <el-form-item>
          <el-input
            v-model="code"
            type="text"
            placeholder="进门密码（支持中英文与符号）"
            autocomplete="off"
            @keyup.enter="submit"
          />
        </el-form-item>
        <el-button type="primary" :loading="loading" style="width: 100%" @click="submit">
          进入
        </el-button>
      </el-form>
    </el-card>
  </div>
</template>

<style scoped>
.auth-page {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 60vh;
  padding: 24px;
}
.auth-card {
  width: 100%;
  max-width: 400px;
}
.auth-card h2 {
  margin: 0 0 8px;
  font-size: 1.25rem;
}
.hint {
  margin: 0 0 20px;
  color: #909399;
  font-size: 0.9rem;
}
</style>
