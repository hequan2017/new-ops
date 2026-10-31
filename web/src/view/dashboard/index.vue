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
              {{ today }} · 资产 / 终端 / 作业 / 流水线 / 容器 / K8s / 监控 / 工单全插件已交付
            </p>
          </div>
          <div class="flex flex-wrap items-center gap-2">
            <el-button type="primary" class="!ml-0" @click="openRepo">
              <el-icon class="mr-1"><Link /></el-icon>
              项目仓库
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

      <el-card shadow="never">
        <template #header>
          <span class="font-semibold">平台概览</span>
        </template>
        <div class="grid grid-cols-1 gap-3 sm:grid-cols-3">
          <el-skeleton :loading="statLoading" animated :rows="2">
            <template #default>
              <div class="rounded-lg border border-slate-200 p-4 dark:border-slate-700">
                <p class="text-xs text-slate-500 dark:text-slate-400">主机资产</p>
                <p class="mt-1 text-2xl font-bold text-slate-900 dark:text-slate-100">
                  {{ stats.hosts ?? '—' }}
                </p>
              </div>
            </template>
          </el-skeleton>
          <el-skeleton :loading="statLoading" animated :rows="2">
            <template #default>
              <div class="rounded-lg border border-slate-200 p-4 dark:border-slate-700">
                <p class="text-xs text-slate-500 dark:text-slate-400">Agent 实例（在线/总）</p>
                <p class="mt-1 text-2xl font-bold text-slate-900 dark:text-slate-100">
                  {{ stats.agents ?? '—' }}
                </p>
              </div>
            </template>
          </el-skeleton>
          <el-skeleton :loading="statLoading" animated :rows="2">
            <template #default>
              <div class="rounded-lg border border-slate-200 p-4 dark:border-slate-700">
                <p class="text-xs text-slate-500 dark:text-slate-400">监控告警规则</p>
                <p class="mt-1 text-2xl font-bold text-slate-900 dark:text-slate-100">
                  {{ stats.alertRules ?? '—' }}
                </p>
              </div>
            </template>
          </el-skeleton>
        </div>
      </el-card>

      <el-alert
        type="success"
        :closable="false"
        show-icon
        title="✅ 核心插件已全部交付"
        description="资产管理、终端与文件、批量作业、流水线、容器、K8s（含 Helm）、监控告警、工单、数据库工单、Agent 通道等插件均已上线；GPU 算力与 goInception 审核引擎部分能力待对应环境联调。"
      />
    </div>
  </div>
</template>

<script setup>
  import { computed, ref, onMounted } from 'vue'
  import { Link, CircleCheckFilled } from '@element-plus/icons-vue'
  import { getAssetHostList } from '@/plugin/asset/api/assetHost'
  import { getAgentInstanceList } from '@/plugin/agent/api/agentInstance'
  import { getAlertRules } from '@/plugin/monitor/api/alert'

  const repoUrl = 'https://github.com/hequan2017/new-ops'

  // 平台概览统计（现有接口轻量拉取；骨架屏防白卡，单项失败不影响其余卡片）
  const statLoading = ref(true)
  const stats = ref({ hosts: null, agents: null, alertRules: null })

  const loadStats = async () => {
    statLoading.value = true
    try {
      const [hosts, agents, rules] = await Promise.allSettled([
        getAssetHostList({ page: 1, pageSize: 1 }),
        getAgentInstanceList(),
        getAlertRules()
      ])
      if (hosts.status === 'fulfilled' && hosts.value.code === 0) {
        stats.value.hosts = hosts.value.data.total ?? null
      }
      if (agents.status === 'fulfilled' && agents.value.code === 0) {
        const list = agents.value.data || []
        const online = list.filter((a) => a.status === '在线').length
        stats.value.agents = `${online}/${list.length}`
      }
      if (rules.status === 'fulfilled' && rules.value.code === 0) {
        stats.value.alertRules = (rules.value.data || []).length
      }
    } finally {
      statLoading.value = false
    }
  }

  onMounted(loadStats)

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
    window.open(repoUrl, '_blank')
  }

  defineOptions({
    name: 'Dashboard'
  })
</script>

<style lang="scss" scoped></style>
