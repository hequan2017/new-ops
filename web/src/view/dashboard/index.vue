<template>
  <div class="h-full ops-container2 overflow-auto bg-slate-50/60 dark:bg-slate-900">
    <div class="space-y-4 p-4 lg:p-6">
      <section
        class="relative overflow-hidden rounded-xl border border-slate-200/80 bg-white px-5 py-6 shadow-sm dark:border-slate-700 dark:bg-slate-800"
      >
        <div class="relative flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
          <div>
            <p class="text-xs tracking-[0.2em] text-slate-500 dark:text-slate-400">BAIZE · NEW-OPS</p>
            <h1 class="mt-2 text-xl font-semibold text-slate-900 dark:text-slate-100 lg:text-2xl">
              欢迎使用 白泽 · 统一运维开发平台
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
          <span class="font-semibold">关于版权标识与授权</span>
        </template>
        <div class="grid grid-cols-1 gap-3 text-sm lg:grid-cols-3">
          <div class="rounded-lg border border-slate-200 p-3 dark:border-slate-700">
            <div class="flex items-center gap-2 font-semibold text-slate-800 dark:text-slate-100">
              <el-icon class="text-emerald-500"><CircleCheckFilled /></el-icon>
              本项目：白泽（BaiZe）
            </div>
            <div class="mt-2 text-slate-600 dark:text-slate-300">
              以 <b>Apache License 2.0</b> 开源，无商业授权限制，可自由用于个人与商业场景（保留许可证与版权声明即可）。
            </div>
          </div>
          <div class="rounded-lg border border-slate-200 p-3 dark:border-slate-700">
            <div class="flex items-center gap-2 font-semibold text-slate-800 dark:text-slate-100">
              <el-icon class="text-emerald-500"><CircleCheckFilled /></el-icon>
              底座：gin-vue-admin
            </div>
            <div class="mt-2 text-slate-600 dark:text-slate-300">
              基于 gin-vue-admin（Apache License 2.0）构建，归属声明保留于仓库 LICENSE 与 README。
            </div>
          </div>
          <div class="rounded-lg border border-slate-200 p-3 dark:border-slate-700">
            <div class="flex items-center gap-2 font-semibold text-slate-800 dark:text-slate-100">
              <el-icon class="text-emerald-500"><CircleCheckFilled /></el-icon>
              技术支持标识与指纹
            </div>
            <div class="mt-2 text-slate-600 dark:text-slate-300">
              系统底部技术支持标识、全站水印及内置隐蔽指纹均已移除，界面不含任何推广与授权引导内容。
            </div>
          </div>
        </div>
        <div class="mt-4 text-xs text-slate-400 dark:text-slate-500">
          完整许可证文本见仓库根目录 LICENSE 文件；<a
            class="text-blue-500"
            href="https://github.com/flipped-aurora/gin-vue-admin"
            target="_blank"
          >上游项目</a
          >与<a
            class="text-blue-500"
            :href="repoUrl"
            target="_blank"
          >本项目仓库</a
          >。
        </div>
      </el-card>

      <el-alert
        type="warning"
        :closable="false"
        show-icon
        title="🚧 项目正在开发中"
        description="当前为 gin-vue-admin 底座环境。资产管理、终端作业、流水线、容器与 K8s、GPU 算力等业务插件正在按开发计划（docs/DEV_PLAN.md）分阶段交付。"
      />
    </div>
  </div>
</template>

<script setup>
  import { computed } from 'vue'
  import { Link, CircleCheckFilled } from '@element-plus/icons-vue'

  const repoUrl = 'https://github.com/hequan2017/new-ops'

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
    window.open(`${repoUrl}/blob/main/docs/DEV_PLAN.md`, '_blank')
  }

  defineOptions({
    name: 'Dashboard'
  })
</script>

<style lang="scss" scoped></style>
