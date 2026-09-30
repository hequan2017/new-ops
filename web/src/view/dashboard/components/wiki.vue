<template>
  <div class="grid grid-cols-2 gap-2">
    <a
      v-for="item in wikis"
      :key="item.url"
      :href="item.url"
      class="text-sm text-black/70 dark:text-white/70 no-underline hover:text-[var(--el-color-primary)] dark:hover:text-white"
      target="_blank"
    >
      {{ item.title }}
    </a>
  </div>
</template>

<script setup>
  const wikis = [
    {
      title: 'Vue3',
      url: 'https://v3.cn.vuejs.org/guide/introduction.html'
    },
    {
      title: 'GIN 文档',
      url: 'https://gin-gonic.com/'
    },
    {
      title: 'GVA 文档',
      url: 'https://www.new-ops.com/'
    },
    {
      title: '插件市场',
      url: 'https://plugin.new-ops.com/'
    },
    {
      title: 'github 仓库',
      url: 'https://github.com/hequan2017/new-ops'
    }
  ]
</script>

<style scoped lang="scss"></style>
