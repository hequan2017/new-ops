<template>
  <el-dialog
    v-model="dialogVisible"
    width="30%"
    class="overlay"
    :show-close="false"
  >
    <template #header>
      <input
        v-model="searchInput"
        class="quick-input"
        placeholder="请输入你需要快捷到达的功能（↑↓ 选择，回车跳转）"
        @keydown.down.prevent="moveActive(1)"
        @keydown.up.prevent="moveActive(-1)"
        @keydown.enter.prevent="activeItem && activeItem.func()"
      />
    </template>

    <div v-for="(option, index) in options" :key="index">
      <div v-if="option.children.length" class="quick-title">
        {{ option.label }}
      </div>
      <div
        v-for="(item, key) in option.children"
        :key="index + '-' + key"
        class="quick-item"
        :class="{ 'quick-item--active': item === activeItem }"
        @click="item.func"
        @mouseenter="setActive(item)"
      >
        {{ item.label }}
      </div>
    </div>

    <template #footer>
      <span class="dialog-footer">
        <el-button @click="close">关闭</el-button>
      </span>
    </template>
  </el-dialog>
</template>

<script setup>
  import { computed, nextTick, reactive, ref, watch } from 'vue'
  import { useRouter } from 'vue-router'
  import { useRouterStore } from '@/pinia/modules/router'
  import { useAppStore, useUserStore } from '@/pinia'
  defineOptions({
    name: 'CommandMenu'
  })
  const appStore = useAppStore()
  const userStore = useUserStore()

  const router = useRouter()
  const route = useRouter()
  const routerStore = useRouterStore()
  const dialogVisible = ref(false)
  const searchInput = ref('')
  const options = reactive([])

  // 键盘导航：扁平化候选项，↑↓ 环形移动、回车执行、hover 同步高亮
  const flatItems = computed(() => options.flatMap((g) => g.children))
  const activeIndex = ref(0)
  const activeItem = computed(() => flatItems.value[activeIndex.value] || null)
  const setActive = (item) => {
    const i = flatItems.value.indexOf(item)
    if (i > -1) activeIndex.value = i
  }
  const moveActive = (delta) => {
    const n = flatItems.value.length
    if (!n) return
    activeIndex.value = (activeIndex.value + delta + n) % n
    scrollActiveIntoView()
  }
  const scrollActiveIntoView = () => {
    nextTick(() => {
      document
        .querySelector('.overlay .quick-item--active')
        ?.scrollIntoView({ block: 'nearest' })
    })
  }
  const deepMenus = (menus) => {
    const arr = []
    menus?.forEach((menu) => {
      if (menu.children && menu.children.length > 0) {
        arr.push(...deepMenus(menu.children))
      } else {
        if (
          menu.meta.title &&
          menu.meta.title.indexOf(searchInput.value) > -1
        ) {
          arr.push({
            label: menu.meta.title,
            func: () => changeRouter(menu)
          })
        }
      }
    })
    return arr
  }

  const addQuickMenu = () => {
    const option = {
      label: '跳转',
      children: []
    }
    const menus = deepMenus(routerStore.asyncRouters[0]?.children || [])
    option.children.push(...menus)
    options.push(option)
  }

  const addQuickOption = () => {
    const option = {
      label: '操作',
      children: []
    }
    const quickArr = [
      {
        label: '亮色主题',
        func: () => changeMode(false)
      },
      {
        label: '暗色主题',
        func: () => changeMode(true)
      },
      {
        label: '退出登录',
        func: () => userStore.LoginOut()
      }
    ]
    option.children.push(
      ...quickArr.filter((item) => item.label.indexOf(searchInput.value) > -1)
    )
    options.push(option)
  }

  addQuickMenu()
  addQuickOption()

  const open = () => {
    dialogVisible.value = true
  }

  const changeRouter = (e) => {
    const index = e.name
    const query = {}
    const params = {}
    routerStore.routeMap[index]?.parameters &&
      routerStore.routeMap[index]?.parameters.forEach((item) => {
        if (item.type === 'query') {
          query[item.key] = item.value
        } else {
          params[item.key] = item.value
        }
      })
    if (index === route.name) return
    if (e.name.indexOf('http://') > -1 || e.name.indexOf('https://') > -1) {
      window.open(e.name)
    } else {
      router.push({ name: index, query, params })
    }
    dialogVisible.value = false
  }

  const changeMode = (darkMode) => {
    appStore.toggleTheme(darkMode)
  }

  const close = () => {
    dialogVisible.value = false
  }

  defineExpose({ open })

  watch(searchInput, () => {
    options.length = 0
    addQuickMenu()
    addQuickOption()
    activeIndex.value = 0
  })
</script>

<style lang="scss">
  .overlay {
    border-radius: 4px;
    .el-dialog__header {
      padding: 0 !important;
      margin-right: 0 !important;
    }
    .el-dialog__body {
      padding: 12px !important;
      height: 50vh;
      overflow: auto !important;
    }
    .quick-title {
      margin-top: 8px;
      font-size: 12px;
      font-weight: 600;
      color: #666;
    }
    .quick-input {
      @apply bg-gray-50 dark:bg-gray-800;
      color: #666;
      border-radius: 4px 4px 0 0;
      border: none;
      padding: 12px 16px;
      box-sizing: border-box;
      width: 100%;
      font-size: 16px;
      border-bottom: 1px solid #ddd;
    }
    .quick-item {
      font-size: 14px;
      padding: 8px;
      margin: 4px 0;
      &:hover {
        @apply bg-gray-200 dark:bg-slate-500;
        cursor: pointer;
        border-radius: 4px;
      }
    }
    .quick-item--active {
      @apply bg-gray-200 dark:bg-slate-500;
      border-radius: 4px;
    }
  }
</style>
