<template>
  <div class="h-full ops-container2 overflow-auto bg-slate-50/60 dark:bg-slate-900">
    <div class="space-y-4 p-4 lg:p-6">
      <section
        class="relative overflow-hidden rounded-xl border border-slate-200/80 bg-white px-5 py-6 shadow-sm dark:border-slate-700 dark:bg-slate-800"
      >
        <div class="relative flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
          <div>
            <p class="text-xs tracking-[0.2em] text-slate-500 dark:text-slate-400">NEW-OPS</p>
            <h1 class="mt-2 text-xl font-semibold text-slate-900 dark:text-slate-100 lg:text-2xl">
              欢迎使用 new-ops · 统一运维开发平台
            </h1>
            <p class="mt-2 text-sm text-slate-600 dark:text-slate-300">
              {{ today }} · 项目正在开发中，业务功能按里程碑逐步交付
            </p>
          </div>
          <div class="flex flex-wrap items-center gap-2">
            <el-button type="primary" class="!ml-0" @click="openRepo">
              <el-icon class="mr-1"><Link /></el-icon>
              开发计划
            </el-button>
          </div>
        </div>
      </section>

      <el-card shadow="never">
        <template #header>
          <span class="font-semibold">当前状态</span>
        </template>
        <el-alert
          type="warning"
          :closable="false"
          show-icon
          title="🚧 项目正在开发中"
          description="当前为 gin-vue-admin 底座环境。资产管理、终端作业、流水线、容器与 K8s、GPU 算力等业务插件正在按开发计划（docs/DEV_PLAN.md）分阶段交付，完成后将在本页面汇总常用入口。"
        />
        <div class="mt-4 grid grid-cols-1 gap-3 text-sm text-slate-600 sm:grid-cols-3 dark:text-slate-300">
          <div class="rounded-lg border border-slate-200 p-3 dark:border-slate-700">
            <div class="font-semibold text-slate-800 dark:text-slate-100">基座就绪</div>
            <div class="mt-1">RBAC 权限、用户角色菜单、代码生成器、插件机制、操作日志</div>
          </div>
          <div class="rounded-lg border border-slate-200 p-3 dark:border-slate-700">
            <div class="font-semibold text-slate-800 dark:text-slate-100">开发中</div>
            <div class="mt-1">M1 资产中心（主机资产、凭据保险库、SSH 采集、云同步）</div>
          </div>
          <div class="rounded-lg border border-slate-200 p-3 dark:border-slate-700">
            <div class="font-semibold text-slate-800 dark:text-slate-100">规划中</div>
            <div class="mt-1">终端作业、流水线发布、Docker/K8s、GPU 算力、监控告警、AI 运维</div>
          </div>
        </div>
      </el-card>
    </div>
  </div>
</template>

<script setup>
  import { computed } from 'vue'
  import { Link } from '@element-plus/icons-vue'

  const REPO_URL = 'https://github.com/hequan2017/new-ops'

  const today = computed(() => {
    try {
      return new Date().toLocaleDateString('zh-CN', {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit'
      })
    } catch (e) {
      return new Date().toISOString().slice(0, 10)
    }
  })

  const openRepo = () => {
    window.open(`${REPO_URL}/blob/main/docs/DEV_PLAN.md`, '_blank')
  }

  defineOptions({
    name: 'Dashboard'
  })
</script>

<style lang="scss" scoped></style>
