<template>
  <div id="userLayout" class="relative w-full h-full overflow-hidden">
    <!-- 背景：品牌渐变 + 网格 + 光斑 -->
    <div class="absolute inset-0 bg-gradient-to-br from-slate-950 via-indigo-950 to-sky-900">
      <div class="absolute inset-0 ops-login-grid" />
      <div class="absolute -top-32 -left-32 h-96 w-96 rounded-full bg-indigo-600/30 blur-3xl" />
      <div class="absolute -bottom-40 -right-24 h-[26rem] w-[26rem] rounded-full bg-sky-500/25 blur-3xl" />
      <div class="absolute top-1/4 right-1/3 h-64 w-64 rounded-full bg-cyan-400/10 blur-2xl" />
    </div>

    <!-- 居中登录卡片 -->
    <div class="relative z-10 flex h-full w-full items-center justify-center p-4">
      <div
        class="w-full max-w-md rounded-2xl border border-white/10 bg-white/90 shadow-2xl backdrop-blur-xl px-8 py-10 dark:border-slate-700/60 dark:bg-slate-900/80"
      >
          <div>
            <div class="flex items-center justify-center">
              <Logo :size="6" />
            </div>
            <div class="mb-9">
              <p class="text-center text-4xl font-bold text-slate-900 dark:text-white">
                {{ $NEW_OPS.appName }}
              </p>
              <p class="text-center text-sm font-normal text-gray-500 dark:text-gray-400 mt-2.5">
                统一运维开发平台 · Go + Vue
              </p>
            </div>
            <el-form
              ref="loginForm"
              :model="loginFormData"
              :rules="rules"
              :validate-on-rule-change="false"
              @keyup.enter="submitForm"
            >
              <el-form-item prop="username" class="mb-6">
                <el-input
                  v-model="loginFormData.username"
                  size="large"
                  placeholder="请输入用户名"
                  suffix-icon="user"
                />
              </el-form-item>
              <el-form-item prop="password" class="mb-6">
                <el-input
                  v-model="loginFormData.password"
                  show-password
                  size="large"
                  type="password"
                  placeholder="请输入密码"
                />
              </el-form-item>
              <el-form-item
                v-if="loginFormData.openCaptcha"
                prop="captcha"
                class="mb-6"
              >
                <div class="flex w-full justify-between">
                  <el-input
                    v-model="loginFormData.captcha"
                    placeholder="请输入验证码"
                    size="large"
                    class="flex-1 mr-5"
                  />
                  <div class="w-1/3 h-11 bg-[#c3d4f2] rounded">
                    <img
                      v-if="picPath"
                      class="w-full h-full"
                      :src="picPath"
                      alt="请输入验证码"
                      @click="loginVerify()"
                    />
                  </div>
                </div>
              </el-form-item>
              <el-form-item class="mb-6">
                <el-button
                  class="shadow shadow-active h-11 w-full"
                  type="primary"
                  size="large"
                  @click="submitForm"
                  >登 录</el-button
                >
              </el-form-item>
              <el-form-item v-if="isDev" class="mb-6">
                <el-button
                  class="shadow shadow-active h-11 w-full"
                  type="primary"
                  size="large"
                  @click="checkInit"
                  >前往初始化</el-button
                >
              </el-form-item>
            </el-form>
          </div>
      </div>
    </div>
  </div>
</template>

<script setup>
  import { captcha } from '@/api/user'
  import { checkDB } from '@/api/initdb'
  import { reactive, ref } from 'vue'
  import { ElMessage } from 'element-plus'
  import { useRouter } from 'vue-router'
  import { useUserStore } from '@/pinia/modules/user'
  import Logo from '@/components/logo/index.vue'
  import { isDev } from '@/utils/env.js'

  defineOptions({
    name: 'Login'
  })

  const router = useRouter()
  const captchaRequiredLength = ref(6)
  // 验证函数
  const checkUsername = (rule, value, callback) => {
    if (value.length < 5) {
      return callback(new Error('请输入正确的用户名'))
    } else {
      callback()
    }
  }
  const checkPassword = (rule, value, callback) => {
    if (value.length < 6) {
      return callback(new Error('请输入正确的密码'))
    } else {
      callback()
    }
  }
  const checkCaptcha = (rule, value, callback) => {
    if (!loginFormData.openCaptcha) {
      return callback()
    }
    const sanitizedValue = (value || '').replace(/\s+/g, '')
    if (!sanitizedValue) {
      return callback(new Error('请输入验证码'))
    }
    if (!/^\d+$/.test(sanitizedValue)) {
      return callback(new Error('验证码须为数字'))
    }
    if (sanitizedValue.length < captchaRequiredLength.value) {
      return callback(
        new Error(`请输入至少${captchaRequiredLength.value}位数字验证码`)
      )
    }
    if (sanitizedValue !== value) {
      loginFormData.captcha = sanitizedValue
    }
    callback()
  }

  // 获取验证码
  const loginVerify = async () => {
    const ele = await captcha()
    captchaRequiredLength.value = Number(ele.data?.captchaLength) || 0
    picPath.value = ele.data?.picPath
    loginFormData.captchaId = ele.data?.captchaId
    loginFormData.openCaptcha = ele.data?.openCaptcha
  }
  loginVerify()

  // 登录相关操作
  const loginForm = ref(null)
  const picPath = ref('')
  const loginFormData = reactive({
    username: 'admin',
    password: '',
    captcha: '',
    captchaId: '',
    openCaptcha: false
  })
  const rules = reactive({
    username: [{ validator: checkUsername, trigger: 'blur' }],
    password: [{ validator: checkPassword, trigger: 'blur' }],
    captcha: [{ validator: checkCaptcha, trigger: 'blur' }]
  })

  const userStore = useUserStore()
  const login = async () => {
    return await userStore.LoginIn(loginFormData)
  }
  const submitForm = () => {
    loginForm.value.validate(async (v) => {
      if (!v) {
        // 未通过前端静态验证
        ElMessage({
          type: 'error',
          message: '请正确填写登录信息',
          showClose: true
        })
        return false
      }

      // 通过验证，请求登陆
      const flag = await login()

      // 登陆失败，刷新验证码
      if (!flag) {
        await loginVerify()
        return false
      }

      // 登陆成功
      return true
    })
  }

  // 跳转初始化
  const checkInit = async () => {
    const res = await checkDB()
    if (res.code === 0) {
      if (res.data?.needInit) {
        userStore.NeedInit()
        await router.push({ name: 'Init' })
      } else {
        ElMessage({
          type: 'info',
          message: '已配置数据库信息，无法初始化'
        })
      }
    }
  }
</script>

<style scoped>
/* 白泽登录背景：品牌渐变上的科技网格纹理 */
.ops-login-grid {
  background-image:
    linear-gradient(rgba(148, 163, 184, 0.09) 1px, transparent 1px),
    linear-gradient(90deg, rgba(148, 163, 184, 0.09) 1px, transparent 1px);
  background-size: 44px 44px;
  -webkit-mask-image: radial-gradient(ellipse at center, #000 25%, transparent 72%);
  mask-image: radial-gradient(ellipse at center, #000 25%, transparent 72%);
}
</style>
