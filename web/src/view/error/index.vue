<template>
  <div>
    <div class="w-full h-screen bg-gray-50 dark:bg-slate-900 flex items-center justify-center">
      <div class="flex flex-col items-center text-2xl gap-4">
        <img class="w-1/3" src="../../assets/404.png" />
        <p class="text-xl font-semibold text-gray-800 dark:text-gray-100">404 · 页面不存在或暂无访问权限</p>
        <p class="text-sm text-gray-500 dark:text-gray-400 max-w-md text-center leading-6">
          常见原因：链接已失效，或该路由尚未分配给当前角色（管理员可在「角色管理 → 菜单权限」中分配）
        </p>
        <div class="flex items-center gap-3 mt-2">
          <el-button type="primary" @click="toDashboard">返回首页</el-button>
          <el-button v-if="canGoBack" @click="goBack">返回上一页</el-button>
          <el-button link type="danger" @click="reLogin">重新登录</el-button>
        </div>
        <p class="text-xs text-gray-400 mt-2">
          项目地址：<a
            href="https://github.com/hequan2017/new-ops"
            target="_blank"
            class="text-blue-500"
            >github.com/hequan2017/new-ops</a
          >
        </p>
      </div>
    </div>
  </div>
</template>

<script setup>
  import { useUserStore } from '@/pinia/modules/user'
  import { useRouter } from 'vue-router'
  import { emitter } from '@/utils/bus'
  import { ElMessageBox } from 'element-plus'

  defineOptions({
    name: 'Error'
  })

  const userStore = useUserStore()
  const router = useRouter()
  const canGoBack = !!history.state?.back
  const goBack = () => router.back()

  const toDashboard = () => {
    try {
      router.push({ name: userStore.userInfo.authority.defaultRouter })
    } catch (error) {
        emitter.emit('show-error', {
        code: '401',
        message: "检测到其他用户修改了路由权限，请重新登录",
        fn: () => {
          userStore.ClearStorage()
          router.push({ name: 'Login', replace: true })
        }
      })
    }
  }

  // 重新登录：清本地登录态回登录页（误触防护：二次确认）
  const reLogin = () => {
    ElMessageBox.confirm('将清除本地登录态并返回登录页，确定继续？', '重新登录', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
      .then(() => {
        userStore.ClearStorage()
        router.push({ name: 'Login', replace: true })
      })
      .catch(() => {})
  }
</script>
